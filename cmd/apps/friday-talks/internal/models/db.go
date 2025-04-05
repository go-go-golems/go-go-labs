package models

import (
	"database/sql"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// SetupDB initializes a database connection and performs migrations
func SetupDB(dbPath string, migrations fs.FS) (*sql.DB, error) {
	// Connect to SQLite
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, errors.Wrap(err, "error opening database")
	}

	// Ensure the database is accessible
	if err := db.Ping(); err != nil {
		return nil, errors.Wrap(err, "error connecting to database")
	}

	// Run migrations
	if err := runMigrations(db, migrations); err != nil {
		return nil, errors.Wrap(err, "error running migrations")
	}

	return db, nil
}

// runMigrations applies database migrations
func runMigrations(db *sql.DB, migrationsFS fs.FS) error {
	// Read migration files
	migrations, err := loadMigrations(migrationsFS)
	if err != nil {
		return errors.Wrap(err, "failed to load migrations")
	}

	// Apply migrations using raw sql execution since migration package
	// doesn't support embedding directly
	for _, migrationPath := range migrations {
		content, err := fs.ReadFile(migrationsFS, migrationPath)
		if err != nil {
			return errors.Wrapf(err, "failed to read migration file: %s", migrationPath)
		}

		// Split multiple statements and execute them
		for _, statement := range strings.Split(string(content), ";") {
			statement = strings.TrimSpace(statement)
			if statement == "" {
				continue
			}

			_, err = db.Exec(statement)
			if err != nil {
				return errors.Wrapf(err, "failed to execute migration: %s", migrationPath)
			}
		}

		fmt.Printf("Applied migration: %s\n", migrationPath)
	}

	return nil
}

// loadMigrations reads migration files from an embedded filesystem
func loadMigrations(migrationsFS fs.FS) ([]string, error) {
	var migrations []string

	// Walk through the migrations filesystem
	err := fs.WalkDir(migrationsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Only include SQL files that don't have "_down" suffix
		ext := filepath.Ext(path)
		if ext == ".sql" && !strings.Contains(path, "_down") {
			migrations = append(migrations, path)
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "error walking migrations directory")
	}

	// Sort migrations by filename
	sort.Strings(migrations)

	return migrations, nil
}

// NewRepositories creates new repository instances using the provided database
func NewRepositories(db *sql.DB) (*Repositories, error) {
	return &Repositories{
		User:       NewSQLiteUserRepository(db),
		Talk:       NewSQLiteTalkRepository(db),
		Vote:       NewSQLiteVoteRepository(db),
		Attendance: NewSQLiteAttendanceRepository(db),
		Resource:   NewSQLiteResourceRepository(db),
	}, nil
}

// Repositories holds all the repository instances
type Repositories struct {
	User       UserRepository
	Talk       TalkRepository
	Vote       VoteRepository
	Attendance AttendanceRepository
	Resource   ResourceRepository
}
