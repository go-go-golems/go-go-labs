package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// TalkStatus represents the status of a talk
type TalkStatus string

// Talk status constants
const (
	TalkStatusProposed  TalkStatus = "proposed"
	TalkStatusScheduled TalkStatus = "scheduled"
	TalkStatusCompleted TalkStatus = "completed"
	TalkStatusCanceled  TalkStatus = "canceled"
)

// Talk represents a talk in the system
type Talk struct {
	ID             int        `json:"id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	SpeakerID      int        `json:"speaker_id"`
	Speaker        *User      `json:"speaker,omitempty"`
	ScheduledDate  *time.Time `json:"scheduled_date"`
	PreferredDates []string   `json:"preferred_dates"`
	Status         TalkStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// TalkRepository defines the interface for talk data operations
type TalkRepository interface {
	FindByID(ctx context.Context, id int) (*Talk, error)
	Create(ctx context.Context, talk *Talk) error
	Update(ctx context.Context, talk *Talk) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]*Talk, error)
	ListByStatus(ctx context.Context, status TalkStatus) ([]*Talk, error)
	ListBySpeaker(ctx context.Context, speakerID int) ([]*Talk, error)
	FindProposedWithPreferredDate(ctx context.Context, date time.Time) ([]*Talk, error)
}

// SQLiteTalkRepository implements TalkRepository for SQLite
type SQLiteTalkRepository struct {
	db *sql.DB
}

// NewSQLiteTalkRepository creates a new SQLiteTalkRepository
func NewSQLiteTalkRepository(db *sql.DB) *SQLiteTalkRepository {
	return &SQLiteTalkRepository{db: db}
}

// FindByID finds a talk by ID
func (r *SQLiteTalkRepository) FindByID(ctx context.Context, id int) (*Talk, error) {
	query := `
		SELECT t.id, t.title, t.description, t.speaker_id, t.scheduled_date, 
		       t.preferred_dates, t.status, t.created_at, t.updated_at,
		       u.id, u.name, u.email, u.created_at, u.updated_at
		FROM talks t
		JOIN users u ON t.speaker_id = u.id
		WHERE t.id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	talk := &Talk{}
	speaker := &User{}
	var preferredDatesJSON string
	var scheduledDateSQL sql.NullTime

	err := row.Scan(
		&talk.ID,
		&talk.Title,
		&talk.Description,
		&talk.SpeakerID,
		&scheduledDateSQL,
		&preferredDatesJSON,
		&talk.Status,
		&talk.CreatedAt,
		&talk.UpdatedAt,
		&speaker.ID,
		&speaker.Name,
		&speaker.Email,
		&speaker.CreatedAt,
		&speaker.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("talk not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "error querying talk")
	}

	// Parse preferred dates JSON
	if err := json.Unmarshal([]byte(preferredDatesJSON), &talk.PreferredDates); err != nil {
		return nil, errors.Wrap(err, "error parsing preferred dates")
	}

	// Handle nullable scheduled date
	if scheduledDateSQL.Valid {
		talk.ScheduledDate = &scheduledDateSQL.Time
	}

	talk.Speaker = speaker

	return talk, nil
}

// Create creates a new talk
func (r *SQLiteTalkRepository) Create(ctx context.Context, talk *Talk) error {
	query := `
		INSERT INTO talks (title, description, speaker_id, scheduled_date, 
		                 preferred_dates, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	// Convert preferred dates to JSON
	preferredDatesJSON, err := json.Marshal(talk.PreferredDates)
	if err != nil {
		return errors.Wrap(err, "error serializing preferred dates")
	}

	result, err := r.db.ExecContext(
		ctx,
		query,
		talk.Title,
		talk.Description,
		talk.SpeakerID,
		talk.ScheduledDate,
		preferredDatesJSON,
		talk.Status,
		now,
		now,
	)

	if err != nil {
		return errors.Wrap(err, "error creating talk")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "error getting last insert id")
	}

	talk.ID = int(id)
	talk.CreatedAt = now
	talk.UpdatedAt = now

	return nil
}

// Update updates an existing talk
func (r *SQLiteTalkRepository) Update(ctx context.Context, talk *Talk) error {
	query := `
		UPDATE talks 
		SET title = ?, description = ?, speaker_id = ?, scheduled_date = ?,
		    preferred_dates = ?, status = ?, updated_at = ? 
		WHERE id = ?
	`

	now := time.Now()

	// Convert preferred dates to JSON
	preferredDatesJSON, err := json.Marshal(talk.PreferredDates)
	if err != nil {
		return errors.Wrap(err, "error serializing preferred dates")
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		talk.Title,
		talk.Description,
		talk.SpeakerID,
		talk.ScheduledDate,
		preferredDatesJSON,
		talk.Status,
		now,
		talk.ID,
	)

	if err != nil {
		return errors.Wrap(err, "error updating talk")
	}

	talk.UpdatedAt = now

	return nil
}

// Delete deletes a talk by ID
func (r *SQLiteTalkRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM talks WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)

	if err != nil {
		return errors.Wrap(err, "error deleting talk")
	}

	return nil
}

// scanTalk scans a row into a Talk struct
func scanTalk(row *sql.Row) (*Talk, error) {
	talk := &Talk{}
	var preferredDatesJSON string
	var scheduledDateSQL sql.NullTime

	err := row.Scan(
		&talk.ID,
		&talk.Title,
		&talk.Description,
		&talk.SpeakerID,
		&scheduledDateSQL,
		&preferredDatesJSON,
		&talk.Status,
		&talk.CreatedAt,
		&talk.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse preferred dates JSON
	if err := json.Unmarshal([]byte(preferredDatesJSON), &talk.PreferredDates); err != nil {
		return nil, errors.Wrap(err, "error parsing preferred dates")
	}

	// Handle nullable scheduled date
	if scheduledDateSQL.Valid {
		talk.ScheduledDate = &scheduledDateSQL.Time
	}

	return talk, nil
}

// scanTalkRows scans SQL rows into Talk structs
func scanTalkRows(rows *sql.Rows) ([]*Talk, error) {
	var talks []*Talk

	for rows.Next() {
		talk := &Talk{}
		var preferredDatesJSON string
		var scheduledDateSQL sql.NullTime

		err := rows.Scan(
			&talk.ID,
			&talk.Title,
			&talk.Description,
			&talk.SpeakerID,
			&scheduledDateSQL,
			&preferredDatesJSON,
			&talk.Status,
			&talk.CreatedAt,
			&talk.UpdatedAt,
		)

		if err != nil {
			return nil, errors.Wrap(err, "error scanning talk row")
		}

		// Parse preferred dates JSON
		if err := json.Unmarshal([]byte(preferredDatesJSON), &talk.PreferredDates); err != nil {
			return nil, errors.Wrap(err, "error parsing preferred dates")
		}

		// Handle nullable scheduled date
		if scheduledDateSQL.Valid {
			talk.ScheduledDate = &scheduledDateSQL.Time
		}

		talks = append(talks, talk)
	}

	return talks, nil
}

// List returns all talks
func (r *SQLiteTalkRepository) List(ctx context.Context) ([]*Talk, error) {
	query := `
		SELECT id, title, description, speaker_id, scheduled_date, 
		       preferred_dates, status, created_at, updated_at
		FROM talks
		ORDER BY COALESCE(scheduled_date, created_at) DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "error querying talks")
	}
	defer rows.Close()

	talks, err := scanTalkRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating talk rows")
	}

	return talks, nil
}

// ListByStatus returns all talks with a given status
func (r *SQLiteTalkRepository) ListByStatus(ctx context.Context, status TalkStatus) ([]*Talk, error) {
	query := `
		SELECT id, title, description, speaker_id, scheduled_date, 
		       preferred_dates, status, created_at, updated_at
		FROM talks
		WHERE status = ?
		ORDER BY COALESCE(scheduled_date, created_at) DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, errors.Wrap(err, "error querying talks by status")
	}
	defer rows.Close()

	talks, err := scanTalkRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating talk rows")
	}

	return talks, nil
}

// ListBySpeaker returns all talks by a given speaker
func (r *SQLiteTalkRepository) ListBySpeaker(ctx context.Context, speakerID int) ([]*Talk, error) {
	query := `
		SELECT id, title, description, speaker_id, scheduled_date, 
		       preferred_dates, status, created_at, updated_at
		FROM talks
		WHERE speaker_id = ?
		ORDER BY COALESCE(scheduled_date, created_at) DESC
	`

	rows, err := r.db.QueryContext(ctx, query, speakerID)
	if err != nil {
		return nil, errors.Wrap(err, "error querying talks by speaker")
	}
	defer rows.Close()

	talks, err := scanTalkRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating talk rows")
	}

	return talks, nil
}

// FindProposedWithPreferredDate finds proposed talks with a preferred date
func (r *SQLiteTalkRepository) FindProposedWithPreferredDate(ctx context.Context, date time.Time) ([]*Talk, error) {
	// We need to use JSON functions to search within the preferred_dates JSON array
	// SQLite doesn't have built-in JSON functions, so we'll use LIKE for a simple implementation
	// A better implementation would use a proper JSON query in PostgreSQL or MySQL

	dateStr := date.Format("2006-01-02")
	query := `
		SELECT id, title, description, speaker_id, scheduled_date, 
		       preferred_dates, status, created_at, updated_at
		FROM talks
		WHERE status = ? AND preferred_dates LIKE ?
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, TalkStatusProposed, "%"+dateStr+"%")
	if err != nil {
		return nil, errors.Wrap(err, "error querying proposed talks with preferred date")
	}
	defer rows.Close()

	var candidates []*Talk

	talks, err := scanTalkRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating talk rows")
	}

	// Filter talks to ensure the date is actually in the preferred dates
	// This is needed because our LIKE query is imprecise
	for _, talk := range talks {
		for _, preferred := range talk.PreferredDates {
			if preferred == dateStr {
				candidates = append(candidates, talk)
				break
			}
		}
	}

	return candidates, nil
}

// Ensure SQLiteTalkRepository implements TalkRepository
var _ TalkRepository = &SQLiteTalkRepository{}
