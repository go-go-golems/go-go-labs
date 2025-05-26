package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	if dbPath == "" {
		return &SQLiteDB{}, nil
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Create the events table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS file_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		pid INTEGER NOT NULL,
		process TEXT NOT NULL,
		operation TEXT NOT NULL,
		filename TEXT,
		fd INTEGER,
		write_size INTEGER,
		content TEXT,
		truncated BOOLEAN DEFAULT FALSE
	);
	
	CREATE INDEX IF NOT EXISTS idx_timestamp ON file_events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_pid ON file_events(pid);
	CREATE INDEX IF NOT EXISTS idx_process ON file_events(process);
	CREATE INDEX IF NOT EXISTS idx_operation ON file_events(operation);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return &SQLiteDB{db: db}, nil
}

func (s *SQLiteDB) Close() {
	if s.db != nil {
		s.db.Close()
	}
}

func (s *SQLiteDB) LogEvent(event models.EventOutput) error {
	if s.db == nil {
		return nil
	}

	insertSQL := `
	INSERT INTO file_events (timestamp, pid, process, operation, filename, fd, write_size, content, truncated)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	var fd *int32
	if event.Fd != 0 {
		fd = &event.Fd
	}

	var writeSize *uint64
	if event.WriteSize > 0 {
		writeSize = &event.WriteSize
	}

	var content *string
	if event.Content != "" {
		content = &event.Content
	}

	_, err := s.db.Exec(insertSQL,
		event.Timestamp,
		event.Pid,
		event.Process,
		event.Operation,
		event.Filename,
		fd,
		writeSize,
		content,
		event.Truncated)

	return err
}

// QueryFilter represents search criteria for querying events
type QueryFilter struct {
	StartTime       *time.Time
	EndTime         *time.Time
	ProcessFilter   string
	OperationFilter []string
	FilenamePattern string
	PID             *uint32
	Limit           int
	Offset          int
}

// QueryEvents retrieves events from the database based on filters
func (s *SQLiteDB) QueryEvents(filter QueryFilter) ([]models.EventOutput, error) {
	if s.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Build the query
	whereConditions := []string{}
	args := []interface{}{}

	if filter.StartTime != nil {
		whereConditions = append(whereConditions, "datetime(timestamp) >= datetime(?)")
		args = append(args, filter.StartTime.Format(time.RFC3339))
	}

	if filter.EndTime != nil {
		whereConditions = append(whereConditions, "datetime(timestamp) <= datetime(?)")
		args = append(args, filter.EndTime.Format(time.RFC3339))
	}

	if filter.ProcessFilter != "" {
		whereConditions = append(whereConditions, "process LIKE ?")
		args = append(args, "%"+filter.ProcessFilter+"%")
	}

	if len(filter.OperationFilter) > 0 {
		placeholders := strings.Repeat("?,", len(filter.OperationFilter)-1) + "?"
		whereConditions = append(whereConditions, "operation IN ("+placeholders+")")
		for _, op := range filter.OperationFilter {
			args = append(args, op)
		}
	}

	if filter.FilenamePattern != "" {
		whereConditions = append(whereConditions, "filename LIKE ?")
		args = append(args, "%"+filter.FilenamePattern+"%")
	}

	if filter.PID != nil {
		whereConditions = append(whereConditions, "pid = ?")
		args = append(args, *filter.PID)
	}

	query := "SELECT timestamp, pid, process, operation, filename, fd, write_size, content, truncated FROM file_events"
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}
	query += " ORDER BY datetime(timestamp) DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []models.EventOutput
	for rows.Next() {
		var event models.EventOutput
		var fd sql.NullInt32
		var writeSize sql.NullInt64
		var content sql.NullString

		err := rows.Scan(
			&event.Timestamp,
			&event.Pid,
			&event.Process,
			&event.Operation,
			&event.Filename,
			&fd,
			&writeSize,
			&content,
			&event.Truncated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if fd.Valid {
			event.Fd = fd.Int32
		}
		if writeSize.Valid {
			event.WriteSize = uint64(writeSize.Int64)
		}
		if content.Valid {
			event.Content = content.String
		}

		events = append(events, event)
	}

	return events, nil
}

// CountEvents returns the total number of events matching the filter
func (s *SQLiteDB) CountEvents(filter QueryFilter) (int, error) {
	if s.db == nil {
		return 0, fmt.Errorf("database not initialized")
	}

	whereConditions := []string{}
	args := []interface{}{}

	if filter.StartTime != nil {
		whereConditions = append(whereConditions, "datetime(timestamp) >= datetime(?)")
		args = append(args, filter.StartTime.Format(time.RFC3339))
	}

	if filter.EndTime != nil {
		whereConditions = append(whereConditions, "datetime(timestamp) <= datetime(?)")
		args = append(args, filter.EndTime.Format(time.RFC3339))
	}

	if filter.ProcessFilter != "" {
		whereConditions = append(whereConditions, "process LIKE ?")
		args = append(args, "%"+filter.ProcessFilter+"%")
	}

	if len(filter.OperationFilter) > 0 {
		placeholders := strings.Repeat("?,", len(filter.OperationFilter)-1) + "?"
		whereConditions = append(whereConditions, "operation IN ("+placeholders+")")
		for _, op := range filter.OperationFilter {
			args = append(args, op)
		}
	}

	if filter.FilenamePattern != "" {
		whereConditions = append(whereConditions, "filename LIKE ?")
		args = append(args, "%"+filter.FilenamePattern+"%")
	}

	if filter.PID != nil {
		whereConditions = append(whereConditions, "pid = ?")
		args = append(args, *filter.PID)
	}

	query := "SELECT COUNT(*) FROM file_events"
	if len(whereConditions) > 0 {
		query += " WHERE " + strings.Join(whereConditions, " AND ")
	}

	var count int
	err := s.db.QueryRow(query, args...).Scan(&count)
	return count, err
}