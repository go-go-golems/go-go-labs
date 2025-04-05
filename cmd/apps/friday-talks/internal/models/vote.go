package models

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// Vote represents a user's vote on a talk
type Vote struct {
	UserID        int             `json:"user_id"`
	TalkID        int             `json:"talk_id"`
	InterestLevel int             `json:"interest_level"` // 1-5
	Availability  map[string]bool `json:"availability"`   // Map of date strings to availability
	User          *User           `json:"user,omitempty"`
	Talk          *Talk           `json:"talk,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

// VoteRepository defines the interface for vote data operations
type VoteRepository interface {
	FindByIDs(ctx context.Context, userID, talkID int) (*Vote, error)
	Create(ctx context.Context, vote *Vote) error
	Update(ctx context.Context, vote *Vote) error
	Delete(ctx context.Context, userID, talkID int) error
	ListByTalk(ctx context.Context, talkID int) ([]*Vote, error)
	ListByUser(ctx context.Context, userID int) ([]*Vote, error)
	GetTalkInterestCount(ctx context.Context, talkID int) (int, error)
	GetAvailabilityForDate(ctx context.Context, talkID int, date time.Time) (map[int]bool, error)
}

// SQLiteVoteRepository implements VoteRepository for SQLite
type SQLiteVoteRepository struct {
	db *sql.DB
}

// NewSQLiteVoteRepository creates a new SQLiteVoteRepository
func NewSQLiteVoteRepository(db *sql.DB) *SQLiteVoteRepository {
	return &SQLiteVoteRepository{db: db}
}

// FindByIDs finds a vote by user ID and talk ID
func (r *SQLiteVoteRepository) FindByIDs(ctx context.Context, userID, talkID int) (*Vote, error) {
	query := `
		SELECT user_id, talk_id, interest_level, availability, created_at, updated_at
		FROM votes
		WHERE user_id = ? AND talk_id = ?
	`

	row := r.db.QueryRowContext(ctx, query, userID, talkID)

	vote := &Vote{}
	var availabilityJSON string

	err := row.Scan(
		&vote.UserID,
		&vote.TalkID,
		&vote.InterestLevel,
		&availabilityJSON,
		&vote.CreatedAt,
		&vote.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("vote not found")
	} else if err != nil {
		return nil, errors.Wrap(err, "error querying vote")
	}

	// Parse availability JSON
	if err := json.Unmarshal([]byte(availabilityJSON), &vote.Availability); err != nil {
		return nil, errors.Wrap(err, "error parsing availability")
	}

	return vote, nil
}

// Create creates a new vote
func (r *SQLiteVoteRepository) Create(ctx context.Context, vote *Vote) error {
	query := `
		INSERT INTO votes (user_id, talk_id, interest_level, availability, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	// Convert availability to JSON
	availabilityJSON, err := json.Marshal(vote.Availability)
	if err != nil {
		return errors.Wrap(err, "error serializing availability")
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		vote.UserID,
		vote.TalkID,
		vote.InterestLevel,
		availabilityJSON,
		now,
		now,
	)

	if err != nil {
		return errors.Wrap(err, "error creating vote")
	}

	vote.CreatedAt = now
	vote.UpdatedAt = now

	return nil
}

// Update updates an existing vote
func (r *SQLiteVoteRepository) Update(ctx context.Context, vote *Vote) error {
	query := `
		UPDATE votes 
		SET interest_level = ?, availability = ?, updated_at = ? 
		WHERE user_id = ? AND talk_id = ?
	`

	now := time.Now()

	// Convert availability to JSON
	availabilityJSON, err := json.Marshal(vote.Availability)
	if err != nil {
		return errors.Wrap(err, "error serializing availability")
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		vote.InterestLevel,
		availabilityJSON,
		now,
		vote.UserID,
		vote.TalkID,
	)

	if err != nil {
		return errors.Wrap(err, "error updating vote")
	}

	vote.UpdatedAt = now

	return nil
}

// Delete deletes a vote
func (r *SQLiteVoteRepository) Delete(ctx context.Context, userID, talkID int) error {
	query := `DELETE FROM votes WHERE user_id = ? AND talk_id = ?`

	_, err := r.db.ExecContext(ctx, query, userID, talkID)

	if err != nil {
		return errors.Wrap(err, "error deleting vote")
	}

	return nil
}

// scanVote scans a row into a Vote struct
func scanVote(row *sql.Row) (*Vote, error) {
	vote := &Vote{}
	var availabilityJSON string

	err := row.Scan(
		&vote.UserID,
		&vote.TalkID,
		&vote.InterestLevel,
		&availabilityJSON,
		&vote.CreatedAt,
		&vote.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse availability JSON
	if err := json.Unmarshal([]byte(availabilityJSON), &vote.Availability); err != nil {
		return nil, errors.Wrap(err, "error parsing availability")
	}

	return vote, nil
}

// scanVoteRows scans SQL rows into Vote structs
func scanVoteRows(rows *sql.Rows) ([]*Vote, error) {
	var votes []*Vote

	for rows.Next() {
		vote := &Vote{}
		var availabilityJSON string

		err := rows.Scan(
			&vote.UserID,
			&vote.TalkID,
			&vote.InterestLevel,
			&availabilityJSON,
			&vote.CreatedAt,
			&vote.UpdatedAt,
		)

		if err != nil {
			return nil, errors.Wrap(err, "error scanning vote row")
		}

		// Parse availability JSON
		if err := json.Unmarshal([]byte(availabilityJSON), &vote.Availability); err != nil {
			return nil, errors.Wrap(err, "error parsing availability")
		}

		votes = append(votes, vote)
	}

	return votes, nil
}

// ListByTalk returns all votes for a talk
func (r *SQLiteVoteRepository) ListByTalk(ctx context.Context, talkID int) ([]*Vote, error) {
	query := `
		SELECT user_id, talk_id, interest_level, availability, created_at, updated_at
		FROM votes
		WHERE talk_id = ?
		ORDER BY interest_level DESC, created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, talkID)
	if err != nil {
		return nil, errors.Wrap(err, "error querying votes")
	}
	defer rows.Close()

	votes, err := scanVoteRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating vote rows")
	}

	return votes, nil
}

// ListByUser returns all votes by a user
func (r *SQLiteVoteRepository) ListByUser(ctx context.Context, userID int) ([]*Vote, error) {
	query := `
		SELECT user_id, talk_id, interest_level, availability, created_at, updated_at
		FROM votes
		WHERE user_id = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Wrap(err, "error querying votes")
	}
	defer rows.Close()

	votes, err := scanVoteRows(rows)
	if err != nil {
		return nil, err
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error iterating vote rows")
	}

	return votes, nil
}

// GetTalkInterestCount returns the total number of users interested in a talk (interest level >= 3)
func (r *SQLiteVoteRepository) GetTalkInterestCount(ctx context.Context, talkID int) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM votes
		WHERE talk_id = ? AND interest_level >= 3
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, talkID).Scan(&count)

	if err != nil {
		return 0, errors.Wrap(err, "error counting interested users")
	}

	return count, nil
}

// GetAvailabilityForDate returns a map of user IDs to their availability for a specific date
func (r *SQLiteVoteRepository) GetAvailabilityForDate(ctx context.Context, talkID int, date time.Time) (map[int]bool, error) {
	dateStr := date.Format("2006-01-02")

	// Get all votes for the talk
	votes, err := r.ListByTalk(ctx, talkID)
	if err != nil {
		return nil, err
	}

	// Create a map of user IDs to availability for the date
	userAvailability := make(map[int]bool)
	for _, vote := range votes {
		if available, exists := vote.Availability[dateStr]; exists {
			userAvailability[vote.UserID] = available
		}
	}

	return userAvailability, nil
}

// Ensure SQLiteVoteRepository implements VoteRepository
var _ VoteRepository = &SQLiteVoteRepository{}
