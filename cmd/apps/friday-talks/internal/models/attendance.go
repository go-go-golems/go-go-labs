package models

import (
	"context"
	"database/sql"
	"time"

	"github.com/pkg/errors"
)

// AttendanceStatus represents the status of a user's attendance at a talk
type AttendanceStatus string

// Attendance status constants
const (
	AttendanceStatusConfirmed AttendanceStatus = "confirmed"
	AttendanceStatusAttended  AttendanceStatus = "attended"
	AttendanceStatusDeclined  AttendanceStatus = "declined"
	AttendanceStatusNoShow    AttendanceStatus = "no-show"
)

// Attendance represents a user's attendance at a talk
type Attendance struct {
	TalkID    int              `json:"talk_id"`
	UserID    int              `json:"user_id"`
	Status    AttendanceStatus `json:"status"`
	Feedback  string           `json:"feedback,omitempty"`
	User      *User            `json:"user,omitempty"`
	Talk      *Talk            `json:"talk,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

// AttendanceRepository defines the interface for attendance data operations
type AttendanceRepository interface {
	FindByIDs(ctx context.Context, talkID, userID int) (*Attendance, error)
	Create(ctx context.Context, attendance *Attendance) error
	Update(ctx context.Context, attendance *Attendance) error
	Delete(ctx context.Context, talkID, userID int) error
	ListByTalk(ctx context.Context, talkID int) ([]*Attendance, error)
	ListByUser(ctx context.Context, userID int) ([]*Attendance, error)
	ListByStatus(ctx context.Context, status AttendanceStatus) ([]*Attendance, error)
	GetTalkAttendanceCount(ctx context.Context, talkID int) (int, error)
}

// SQLiteAttendanceRepository implements AttendanceRepository for SQLite
type SQLiteAttendanceRepository struct {
	db *sql.DB
}

// NewSQLiteAttendanceRepository creates a new SQLiteAttendanceRepository
func NewSQLiteAttendanceRepository(db *sql.DB) *SQLiteAttendanceRepository {
	return &SQLiteAttendanceRepository{db: db}
}

// FindByIDs finds an attendance record by talk ID and user ID
func (r *SQLiteAttendanceRepository) FindByIDs(ctx context.Context, talkID, userID int) (*Attendance, error) {
	query := `
		SELECT talk_id, user_id, status, feedback, created_at, updated_at
		FROM attendance
		WHERE talk_id = ? AND user_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, talkID, userID)

	attendance := &Attendance{}
	var feedbackSQL sql.NullString

	err := row.Scan(
		&attendance.TalkID,
		&attendance.UserID,
		&attendance.Status,
		&feedbackSQL,
		&attendance.CreatedAt,
		&attendance.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("attendance record not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "error querying attendance")
	}

	// Handle nullable feedback
	if feedbackSQL.Valid {
		attendance.Feedback = feedbackSQL.String
	}

	return attendance, nil
}

// Create creates a new attendance record
func (r *SQLiteAttendanceRepository) Create(ctx context.Context, attendance *Attendance) error {
	query := `
		INSERT INTO attendance (talk_id, user_id, status, feedback, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	// Use NULL for empty feedback
	var feedbackValue interface{}
	if attendance.Feedback == "" {
		feedbackValue = nil
	} else {
		feedbackValue = attendance.Feedback
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		attendance.TalkID,
		attendance.UserID,
		attendance.Status,
		feedbackValue,
		now,
		now,
	)

	if err != nil {
		return errors.Wrap(err, "error creating attendance record")
	}

	attendance.CreatedAt = now
	attendance.UpdatedAt = now

	return nil
}

// Update updates an existing attendance record
func (r *SQLiteAttendanceRepository) Update(ctx context.Context, attendance *Attendance) error {
	query := `
		UPDATE attendance 
		SET status = ?, feedback = ?, updated_at = ? 
		WHERE talk_id = ? AND user_id = ?
	`

	now := time.Now()

	// Use NULL for empty feedback
	var feedbackValue interface{}
	if attendance.Feedback == "" {
		feedbackValue = nil
	} else {
		feedbackValue = attendance.Feedback
	}

	_, err := r.db.ExecContext(
		ctx,
		query,
		attendance.Status,
		feedbackValue,
		now,
		attendance.TalkID,
		attendance.UserID,
	)

	if err != nil {
		return errors.Wrap(err, "error updating attendance record")
	}

	attendance.UpdatedAt = now

	return nil
}

// Delete deletes an attendance record
func (r *SQLiteAttendanceRepository) Delete(ctx context.Context, talkID, userID int) error {
	query := `DELETE FROM attendance WHERE talk_id = ? AND user_id = ?`

	_, err := r.db.ExecContext(ctx, query, talkID, userID)

	if err != nil {
		return errors.Wrap(err, "error deleting attendance record")
	}

	return nil
}

// scanAttendance scans a row into an Attendance struct
func scanAttendance(row *sql.Row) (*Attendance, error) {
	attendance := &Attendance{}
	var feedbackSQL sql.NullString

	err := row.Scan(
		&attendance.TalkID,
		&attendance.UserID,
		&attendance.Status,
		&feedbackSQL,
		&attendance.CreatedAt,
		&attendance.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Handle nullable feedback
	if feedbackSQL.Valid {
		attendance.Feedback = feedbackSQL.String
	}

	return attendance, nil
}

// scanAttendanceRows scans SQL rows into Attendance structs
func scanAttendanceRows(rows *sql.Rows) ([]*Attendance, error) {
	var attendances []*Attendance

	for rows.Next() {
		attendance := &Attendance{}
		var feedbackSQL sql.NullString

		err := rows.Scan(
			&attendance.TalkID,
			&attendance.UserID,
			&attendance.Status,
			&feedbackSQL,
			&attendance.CreatedAt,
			&attendance.UpdatedAt,
		)

		if err != nil {
			return nil, errors.Wrap(err, "error scanning attendance row")
		}

		// Handle nullable feedback
		if feedbackSQL.Valid {
			attendance.Feedback = feedbackSQL.String
		}

		attendances = append(attendances, attendance)
	}

	return attendances, nil
}

// ListByTalk returns all attendance records for a talk
func (r *SQLiteAttendanceRepository) ListByTalk(ctx context.Context, talkID int) ([]*Attendance, error) {
	query := `
		SELECT talk_id, user_id, status, feedback, created_at, updated_at
		FROM attendance
		WHERE talk_id = ?
		ORDER BY status, created_at
	`

	rows, err := r.db.QueryContext(ctx, query, talkID)
	if err != nil {
		return nil, errors.Wrap(err, "error querying attendance records")
	}
	defer rows.Close()

	attendances, err := scanAttendanceRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating attendance rows")
	}

	return attendances, nil
}

// ListByUser returns all attendance records for a user
func (r *SQLiteAttendanceRepository) ListByUser(ctx context.Context, userID int) ([]*Attendance, error) {
	query := `
		SELECT talk_id, user_id, status, feedback, created_at, updated_at
		FROM attendance
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "error querying attendance records")
	}
	defer rows.Close()

	attendances, err := scanAttendanceRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating attendance rows")
	}

	return attendances, nil
}

// ListByStatus returns all attendance records with a given status
func (r *SQLiteAttendanceRepository) ListByStatus(ctx context.Context, status AttendanceStatus) ([]*Attendance, error) {
	query := `
		SELECT talk_id, user_id, status, feedback, created_at, updated_at
		FROM attendance
		WHERE status = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, errors.Wrap(err, "error querying attendance records")
	}
	defer rows.Close()

	attendances, err := scanAttendanceRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating attendance rows")
	}

	return attendances, nil
}

// GetTalkAttendanceCount returns the total number of confirmed attendees for a talk
func (r *SQLiteAttendanceRepository) GetTalkAttendanceCount(ctx context.Context, talkID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM attendance
		WHERE talk_id = ? AND status IN (?, ?)
	`

	var count int
	err := r.db.QueryRowContext(
		ctx,
		query,
		talkID,
		AttendanceStatusConfirmed,
		AttendanceStatusAttended,
	).Scan(&count)

	if err != nil {
		return 0, errors.Wrap(err, "error counting attendees")
	}

	return count, nil
}

// Ensure SQLiteAttendanceRepository implements AttendanceRepository
var _ AttendanceRepository = &SQLiteAttendanceRepository{}
