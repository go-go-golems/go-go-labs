package modem

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	_ "github.com/mattn/go-sqlite3"
)

// Database represents the SQLite database for storing modem history
type Database struct {
	db *sql.DB
}

// NewDatabase creates a new database connection and initializes tables
func NewDatabase() (*Database, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	database := &Database{db: db}
	if err := database.initTables(); err != nil {
		db.Close()
		return nil, err
	}

	log.Debug().Str("path", dbPath).Msg("Database initialized")
	return database, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// getDatabasePath returns the path to the database file
func getDatabasePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get user home directory")
	}
	
	configDir := filepath.Join(homeDir, ".config", "poll-modem")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", errors.Wrap(err, "failed to create config directory")
	}
	
	return filepath.Join(configDir, "history.db"), nil
}

// initTables creates the necessary tables if they don't exist
func (d *Database) initTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS sessions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS cable_modem_info (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			hw_version TEXT,
			vendor TEXT,
			boot_version TEXT,
			core_version TEXT,
			model TEXT,
			product_type TEXT,
			flash_part TEXT,
			download_version TEXT,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS downstream_channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			channel_id TEXT NOT NULL,
			lock_status TEXT,
			frequency TEXT,
			snr TEXT,
			power_level TEXT,
			modulation TEXT,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS upstream_channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			channel_id TEXT NOT NULL,
			lock_status TEXT,
			frequency TEXT,
			symbol_rate TEXT,
			power_level TEXT,
			modulation TEXT,
			channel_type TEXT,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
		`CREATE TABLE IF NOT EXISTS error_channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			channel_id TEXT NOT NULL,
			unerrored_codewords TEXT,
			correctable_codewords TEXT,
			uncorrectable_codewords TEXT,
			FOREIGN KEY (session_id) REFERENCES sessions(id)
		)`,
	}

	for _, query := range queries {
		if _, err := d.db.Exec(query); err != nil {
			return errors.Wrapf(err, "failed to execute query: %s", query)
		}
	}

	return nil
}

// StartSession creates a new session and returns its ID
func (d *Database) StartSession() (int64, error) {
	result, err := d.db.Exec("INSERT INTO sessions (start_time) VALUES (?)", time.Now())
	if err != nil {
		return 0, errors.Wrap(err, "failed to start session")
	}

	sessionID, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "failed to get session ID")
	}

	log.Debug().Int64("session_id", sessionID).Msg("Started new session")
	return sessionID, nil
}

// EndSession marks a session as ended
func (d *Database) EndSession(sessionID int64) error {
	_, err := d.db.Exec("UPDATE sessions SET end_time = ? WHERE id = ?", time.Now(), sessionID)
	if err != nil {
		return errors.Wrap(err, "failed to end session")
	}

	log.Debug().Int64("session_id", sessionID).Msg("Ended session")
	return nil
}

// StoreModemInfo stores complete modem information in the database
func (d *Database) StoreModemInfo(sessionID int64, info *ModemInfo) error {
	tx, err := d.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	timestamp := info.LastUpdated

	// Store cable modem info
	_, err = tx.Exec(`
		INSERT INTO cable_modem_info 
		(session_id, timestamp, hw_version, vendor, boot_version, core_version, model, product_type, flash_part, download_version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sessionID, timestamp, info.CableModem.HWVersion, info.CableModem.Vendor,
		info.CableModem.BOOTVersion, info.CableModem.CoreVersion, info.CableModem.Model,
		info.CableModem.ProductType, info.CableModem.FlashPart, info.CableModem.DownloadVersion)
	if err != nil {
		return errors.Wrap(err, "failed to store cable modem info")
	}

	// Store downstream channels
	for _, ch := range info.Downstream {
		_, err = tx.Exec(`
			INSERT INTO downstream_channels 
			(session_id, timestamp, channel_id, lock_status, frequency, snr, power_level, modulation)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
			sessionID, timestamp, ch.ChannelID, ch.LockStatus, ch.Frequency,
			ch.SNR, ch.PowerLevel, ch.Modulation)
		if err != nil {
			return errors.Wrap(err, "failed to store downstream channel")
		}
	}

	// Store upstream channels
	for _, ch := range info.Upstream {
		_, err = tx.Exec(`
			INSERT INTO upstream_channels 
			(session_id, timestamp, channel_id, lock_status, frequency, symbol_rate, power_level, modulation, channel_type)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			sessionID, timestamp, ch.ChannelID, ch.LockStatus, ch.Frequency,
			ch.SymbolRate, ch.PowerLevel, ch.Modulation, ch.ChannelType)
		if err != nil {
			return errors.Wrap(err, "failed to store upstream channel")
		}
	}

	// Store error channels
	for _, ch := range info.ErrorCodewords {
		_, err = tx.Exec(`
			INSERT INTO error_channels 
			(session_id, timestamp, channel_id, unerrored_codewords, correctable_codewords, uncorrectable_codewords)
			VALUES (?, ?, ?, ?, ?, ?)`,
			sessionID, timestamp, ch.ChannelID, ch.UnerroredCodewords,
			ch.CorrectableCodewords, ch.UncorrectableCodewords)
		if err != nil {
			return errors.Wrap(err, "failed to store error channel")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	log.Debug().Int64("session_id", sessionID).Time("timestamp", timestamp).Msg("Stored modem info")
	return nil
}

// ExportMode represents different export modes
type ExportMode int

const (
	ExportCurrent ExportMode = iota
	ExportSession
	ExportAll
)

// ExportData exports data based on the specified mode
func (d *Database) ExportData(mode ExportMode, sessionID int64) (*ExportResult, error) {
	var whereClause string
	var args []interface{}

	switch mode {
	case ExportCurrent:
		// Export only the latest entry for the current session
		whereClause = "WHERE session_id = ? ORDER BY timestamp DESC LIMIT 1"
		args = []interface{}{sessionID}
	case ExportSession:
		// Export all data from the current session
		whereClause = "WHERE session_id = ? ORDER BY timestamp"
		args = []interface{}{sessionID}
	case ExportAll:
		// Export all data from all sessions
		whereClause = "ORDER BY timestamp"
		args = []interface{}{}
	}

	result := &ExportResult{}

	// Export downstream channels
	downstream, err := d.exportDownstreamChannels(whereClause, args...)
	if err != nil {
		return nil, err
	}
	result.Downstream = downstream

	// Export upstream channels
	upstream, err := d.exportUpstreamChannels(whereClause, args...)
	if err != nil {
		return nil, err
	}
	result.Upstream = upstream

	// Export error channels
	errors, err := d.exportErrorChannels(whereClause, args...)
	if err != nil {
		return nil, err
	}
	result.Errors = errors

	return result, nil
}

// ExportResult holds exported data
type ExportResult struct {
	Downstream []DownstreamExport
	Upstream   []UpstreamExport
	Errors     []ErrorExport
}

type DownstreamExport struct {
	Timestamp   time.Time
	SessionID   int64
	ChannelID   string
	LockStatus  string
	Frequency   string
	SNR         string
	PowerLevel  string
	Modulation  string
}

type UpstreamExport struct {
	Timestamp   time.Time
	SessionID   int64
	ChannelID   string
	LockStatus  string
	Frequency   string
	SymbolRate  string
	PowerLevel  string
	Modulation  string
	ChannelType string
}

type ErrorExport struct {
	Timestamp              time.Time
	SessionID              int64
	ChannelID              string
	UnerroredCodewords     string
	CorrectableCodewords   string
	UncorrectableCodewords string
}

func (d *Database) exportDownstreamChannels(whereClause string, args ...interface{}) ([]DownstreamExport, error) {
	query := `SELECT session_id, timestamp, channel_id, lock_status, frequency, snr, power_level, modulation 
			  FROM downstream_channels ` + whereClause

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query downstream channels")
	}
	defer rows.Close()

	var results []DownstreamExport
	for rows.Next() {
		var export DownstreamExport
		err := rows.Scan(&export.SessionID, &export.Timestamp, &export.ChannelID,
			&export.LockStatus, &export.Frequency, &export.SNR,
			&export.PowerLevel, &export.Modulation)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan downstream channel")
		}
		results = append(results, export)
	}

	return results, nil
}

func (d *Database) exportUpstreamChannels(whereClause string, args ...interface{}) ([]UpstreamExport, error) {
	query := `SELECT session_id, timestamp, channel_id, lock_status, frequency, symbol_rate, power_level, modulation, channel_type 
			  FROM upstream_channels ` + whereClause

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query upstream channels")
	}
	defer rows.Close()

	var results []UpstreamExport
	for rows.Next() {
		var export UpstreamExport
		err := rows.Scan(&export.SessionID, &export.Timestamp, &export.ChannelID,
			&export.LockStatus, &export.Frequency, &export.SymbolRate,
			&export.PowerLevel, &export.Modulation, &export.ChannelType)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan upstream channel")
		}
		results = append(results, export)
	}

	return results, nil
}

func (d *Database) exportErrorChannels(whereClause string, args ...interface{}) ([]ErrorExport, error) {
	query := `SELECT session_id, timestamp, channel_id, unerrored_codewords, correctable_codewords, uncorrectable_codewords 
			  FROM error_channels ` + whereClause

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query error channels")
	}
	defer rows.Close()

	var results []ErrorExport
	for rows.Next() {
		var export ErrorExport
		err := rows.Scan(&export.SessionID, &export.Timestamp, &export.ChannelID,
			&export.UnerroredCodewords, &export.CorrectableCodewords,
			&export.UncorrectableCodewords)
		if err != nil {
			return nil, errors.Wrap(err, "failed to scan error channel")
		}
		results = append(results, export)
	}

	return results, nil
} 