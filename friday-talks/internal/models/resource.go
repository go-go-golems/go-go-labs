package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// ResourceType represents the type of a resource
type ResourceType string

// Resource type constants
const (
	ResourceTypeSlides  ResourceType = "slides"
	ResourceTypeVideo   ResourceType = "video"
	ResourceTypeCode    ResourceType = "code"
	ResourceTypeArticle ResourceType = "article"
	ResourceTypeOther   ResourceType = "other"
)

// Resource represents a resource attached to a talk
type Resource struct {
	ID        int          `json:"id"`
	TalkID    int          `json:"talk_id"`
	Title     string       `json:"title"`
	URL       string       `json:"url"`
	Type      ResourceType `json:"type"`
	Talk      *Talk        `json:"talk,omitempty"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// ResourceRepository defines the interface for resource data operations
type ResourceRepository interface {
	FindByID(ctx context.Context, id int) (*Resource, error)
	Create(ctx context.Context, resource *Resource) error
	Update(ctx context.Context, resource *Resource) error
	Delete(ctx context.Context, id int) error
	ListByTalk(ctx context.Context, talkID int) ([]*Resource, error)
	ListByType(ctx context.Context, resourceType ResourceType) ([]*Resource, error)
}

// SQLiteResourceRepository implements ResourceRepository for SQLite
type SQLiteResourceRepository struct {
	db *sql.DB
}

// NewSQLiteResourceRepository creates a new SQLiteResourceRepository
func NewSQLiteResourceRepository(db *sql.DB) *SQLiteResourceRepository {
	return &SQLiteResourceRepository{db: db}
}

// FindByID finds a resource by ID
func (r *SQLiteResourceRepository) FindByID(ctx context.Context, id int) (*Resource, error) {
	query := `
		SELECT id, talk_id, title, url, type, created_at, updated_at
		FROM resources
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	resource := &Resource{}

	err := row.Scan(
		&resource.ID,
		&resource.TalkID,
		&resource.Title,
		&resource.URL,
		&resource.Type,
		&resource.CreatedAt,
		&resource.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("resource not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "error querying resource")
	}

	return resource, nil
}

// Create creates a new resource
func (r *SQLiteResourceRepository) Create(ctx context.Context, resource *Resource) error {
	query := `
		INSERT INTO resources (talk_id, title, url, type, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	result, err := r.db.ExecContext(
		ctx,
		query,
		resource.TalkID,
		resource.Title,
		resource.URL,
		resource.Type,
		now,
		now,
	)

	if err != nil {
		return errors.Wrap(err, "error creating resource")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "error getting last insert id")
	}

	resource.ID = int(id)
	resource.CreatedAt = now
	resource.UpdatedAt = now

	return nil
}

// Update updates an existing resource
func (r *SQLiteResourceRepository) Update(ctx context.Context, resource *Resource) error {
	query := `
		UPDATE resources 
		SET talk_id = ?, title = ?, url = ?, type = ?, updated_at = ? 
		WHERE id = ?
	`

	now := time.Now()

	_, err := r.db.ExecContext(
		ctx,
		query,
		resource.TalkID,
		resource.Title,
		resource.URL,
		resource.Type,
		now,
		resource.ID,
	)

	if err != nil {
		return errors.Wrap(err, "error updating resource")
	}

	resource.UpdatedAt = now

	return nil
}

// Delete deletes a resource by ID
func (r *SQLiteResourceRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM resources WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return errors.Wrap(err, "error deleting resource")
	}

	return nil
}

// scanResource scans a row into a Resource struct
func scanResource(row *sql.Row) (*Resource, error) {
	resource := &Resource{}

	err := row.Scan(
		&resource.ID,
		&resource.TalkID,
		&resource.Title,
		&resource.URL,
		&resource.Type,
		&resource.CreatedAt,
		&resource.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return resource, nil
}

// scanResourceRows scans SQL rows into Resource structs
func scanResourceRows(rows *sql.Rows) ([]*Resource, error) {
	var resources []*Resource

	for rows.Next() {
		resource := &Resource{}

		err := rows.Scan(
			&resource.ID,
			&resource.TalkID,
			&resource.Title,
			&resource.URL,
			&resource.Type,
			&resource.CreatedAt,
			&resource.UpdatedAt,
		)

		if err != nil {
			return nil, errors.Wrap(err, "error scanning resource row")
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

// ListByTalk returns all resources for a talk
func (r *SQLiteResourceRepository) ListByTalk(ctx context.Context, talkID int) ([]*Resource, error) {
	query := `
		SELECT id, talk_id, title, url, type, created_at, updated_at
		FROM resources
		WHERE talk_id = ?
		ORDER BY type, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, talkID)
	if err != nil {
		return nil, errors.Wrap(err, "error querying resources")
	}
	defer rows.Close()

	resources, err := scanResourceRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating resource rows")
	}

	return resources, nil
}

// ListByType returns all resources of a given type
func (r *SQLiteResourceRepository) ListByType(ctx context.Context, resourceType ResourceType) ([]*Resource, error) {
	query := `
		SELECT id, talk_id, title, url, type, created_at, updated_at
		FROM resources
		WHERE type = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, resourceType)
	if err != nil {
		return nil, errors.Wrap(err, "error querying resources by type")
	}
	defer rows.Close()

	resources, err := scanResourceRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating resource rows")
	}

	return resources, nil
}

// Ensure SQLiteResourceRepository implements ResourceRepository
var _ ResourceRepository = &SQLiteResourceRepository{}
