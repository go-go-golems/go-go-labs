package db

import (
	"database/sql"
	_ "modernc.org/sqlite"

	"github.com/pkg/errors"
)

// Migrate creates the database schema if it doesn't exist.
func Migrate(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS animals (
  id          INTEGER PRIMARY KEY AUTOINCREMENT,
  name        TEXT NOT NULL,
  created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_animals_name_unique ON animals(name);
`

	if _, err := db.Exec(schema); err != nil {
		return errors.Wrap(err, "failed to create schema")
	}

	return nil
}

// OpenDB opens a SQLite database and runs migrations.
func OpenDB(dbPath string) (*sql.DB, error) {
	// Use modernc.org/sqlite (pure Go, no CGO)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open database")
	}

	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "failed to ping database")
	}

	if err := Migrate(db); err != nil {
		return nil, errors.Wrap(err, "failed to migrate database")
	}

	return db, nil
}

