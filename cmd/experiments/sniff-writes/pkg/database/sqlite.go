package database

import (
	"database/sql"
	"fmt"

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
