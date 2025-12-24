package animals

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

type Animal struct {
	ID        int64
	Name      string
	CreatedAt time.Time
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all animals ordered by name.
func (r *Repository) List(ctx context.Context) ([]Animal, error) {
	query := `SELECT id, name, created_at FROM animals ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query animals")
	}
	defer rows.Close()

	var animals []Animal
	for rows.Next() {
		var a Animal
		var createdAtStr string
		if err := rows.Scan(&a.ID, &a.Name, &createdAtStr); err != nil {
			return nil, errors.Wrap(err, "failed to scan animal")
		}
		createdAt, err := time.Parse(time.RFC3339Nano, createdAtStr)
		if err != nil {
			// Fallback to RFC3339 if nanoseconds aren't present
			createdAt, err = time.Parse(time.RFC3339, createdAtStr)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse created_at: %q", createdAtStr)
			}
		}
		a.CreatedAt = createdAt
		animals = append(animals, a)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating rows")
	}

	return animals, nil
}

// Clear deletes all animals.
func (r *Repository) Clear(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM animals`)
	if err != nil {
		return errors.Wrap(err, "failed to clear animals")
	}
	return nil
}

// InsertMode determines how to handle existing animals.
type InsertMode string

const (
	InsertModeReplace InsertMode = "replace" // Delete all, then insert
	InsertModeAppend  InsertMode = "append"  // Insert with IGNORE
)

// InsertMany inserts multiple animal names.
// If mode is Replace, it clears the table first.
// Duplicates are ignored (unique constraint).
func (r *Repository) InsertMany(ctx context.Context, names []string, mode InsertMode) (int, error) {
	if len(names) == 0 {
		return 0, nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, errors.Wrap(err, "failed to begin transaction")
	}
	defer tx.Rollback()

	if mode == InsertModeReplace {
		if _, err := tx.ExecContext(ctx, `DELETE FROM animals`); err != nil {
			return 0, errors.Wrap(err, "failed to clear animals")
		}
	}

	stmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO animals (name) VALUES (?)`)
	if err != nil {
		return 0, errors.Wrap(err, "failed to prepare insert statement")
	}
	defer stmt.Close()

	inserted := 0
	for _, name := range names {
		result, err := stmt.ExecContext(ctx, name)
		if err != nil {
			return 0, errors.Wrapf(err, "failed to insert animal %q", name)
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get rows affected")
		}
		if rowsAffected > 0 {
			inserted++
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, errors.Wrap(err, "failed to commit transaction")
	}

	return inserted, nil
}

