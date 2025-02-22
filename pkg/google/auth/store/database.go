package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// MentoDatabaseTokenStore implements TokenStore using a database
type MentoDatabaseTokenStore struct {
	db        *sql.DB
	userID    int
	provider  string
	scopes    []string
	teamID    string
	userAppID string
	appID     string
}

type MentoDatabaseTokenStoreOption func(*MentoDatabaseTokenStore)

// WithUserID sets the user ID for the token store
func WithUserID(userID int) MentoDatabaseTokenStoreOption {
	return func(s *MentoDatabaseTokenStore) {
		s.userID = userID
	}
}

// WithProvider sets the OAuth provider for the token store
func WithProvider(provider string) MentoDatabaseTokenStoreOption {
	return func(s *MentoDatabaseTokenStore) {
		s.provider = provider
	}
}

// WithScopes sets the OAuth scopes for the token store
func WithScopes(scopes []string) MentoDatabaseTokenStoreOption {
	return func(s *MentoDatabaseTokenStore) {
		s.scopes = scopes
	}
}

// WithTeamID sets the Slack team ID for the token store
func WithTeamID(teamID string) MentoDatabaseTokenStoreOption {
	return func(s *MentoDatabaseTokenStore) {
		s.teamID = teamID
	}
}

// WithUserAppID sets the Slack user ID for the token store
func WithUserAppID(userAppID string) MentoDatabaseTokenStoreOption {
	return func(s *MentoDatabaseTokenStore) {
		s.userAppID = userAppID
	}
}

// WithAppID sets the Slack app installation ID for the token store
func WithAppID(appID string) MentoDatabaseTokenStoreOption {
	return func(s *MentoDatabaseTokenStore) {
		s.appID = appID
	}
}

// NewMentoDatabaseTokenStore creates a new database token store
func NewMentoDatabaseTokenStore(db *sql.DB, options ...MentoDatabaseTokenStoreOption) *MentoDatabaseTokenStore {
	store := &MentoDatabaseTokenStore{
		db: db,
	}

	for _, opt := range options {
		opt(store)
	}

	return store
}

// Save implements TokenStore
func (s *MentoDatabaseTokenStore) Save(ctx context.Context, token *oauth2.Token) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if s.userID == 0 {
		return errors.New("user ID is required")
	}

	if s.provider == "" {
		return errors.New("provider is required")
	}

	scopesJSON, err := json.Marshal(s.scopes)
	if err != nil {
		return errors.Wrap(err, "failed to marshal scopes")
	}

	// First try to update existing token
	query := `
		UPDATE oauth_tokens 
		SET access_token = $1, 
			refresh_token = $2, 
			expires_at = $3, 
			updated_at = NOW(),
			scopes = $4,
			slack_team_id = $5,
			slack_user_id = $6,
			slack_app_installation_id = $7
		WHERE user_id = $8 
		AND provider = $9
		RETURNING id`

	var id int
	err = s.db.QueryRowContext(ctx, query,
		token.AccessToken,
		token.RefreshToken,
		token.Expiry,
		scopesJSON,
		s.teamID,
		s.userAppID,
		s.appID,
		s.userID,
		s.provider,
	).Scan(&id)

	if err == sql.ErrNoRows {
		// If no existing token, insert new one
		query = `
			INSERT INTO oauth_tokens (
				user_id, 
				provider, 
				access_token, 
				refresh_token, 
				expires_at, 
				created_at, 
				updated_at,
				scopes,
				slack_team_id,
				slack_user_id,
				slack_app_installation_id
			) VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), $6, $7, $8, $9)
			RETURNING id`

		err = s.db.QueryRowContext(ctx, query,
			s.userID,
			s.provider,
			token.AccessToken,
			token.RefreshToken,
			token.Expiry,
			scopesJSON,
			s.teamID,
			s.userAppID,
			s.appID,
		).Scan(&id)
	}

	if err != nil {
		return errors.Wrap(err, "failed to save token")
	}

	log.Debug().
		Int("id", id).
		Int("user_id", s.userID).
		Str("provider", s.provider).
		Time("expires_at", token.Expiry).
		Msg("Saved OAuth token to database")

	return nil
}

// Load implements TokenStore
func (s *MentoDatabaseTokenStore) Load(ctx context.Context) (*oauth2.Token, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if s.userID == 0 {
		return nil, errors.New("user ID is required")
	}

	if s.provider == "" {
		return nil, errors.New("provider is required")
	}

	query := `
		SELECT access_token, refresh_token, expires_at, updated_at
		FROM oauth_tokens
		WHERE user_id = $1 
		AND provider = $2`

	var token oauth2.Token
	var updatedAt time.Time

	err := s.db.QueryRowContext(ctx, query, s.userID, s.provider).Scan(
		&token.AccessToken,
		&token.RefreshToken,
		&token.Expiry,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTokenNotFound
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to load token")
	}

	// Set token type to Bearer as it's the most common
	token.TokenType = "Bearer"

	log.Debug().
		Int("user_id", s.userID).
		Str("provider", s.provider).
		Time("expires_at", token.Expiry).
		Time("updated_at", updatedAt).
		Msg("Loaded OAuth token from database")

	return &token, nil
}

// Clear implements TokenStore
func (s *MentoDatabaseTokenStore) Clear(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if s.userID == 0 {
		return errors.New("user ID is required")
	}

	if s.provider == "" {
		return errors.New("provider is required")
	}

	query := `
		DELETE FROM oauth_tokens 
		WHERE user_id = $1 
		AND provider = $2`

	result, err := s.db.ExecContext(ctx, query, s.userID, s.provider)
	if err != nil {
		return errors.Wrap(err, "failed to clear token")
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "failed to get rows affected")
	}

	log.Debug().
		Int("user_id", s.userID).
		Str("provider", s.provider).
		Int64("rows_affected", rowsAffected).
		Msg("Cleared OAuth token from database")

	return nil
}
