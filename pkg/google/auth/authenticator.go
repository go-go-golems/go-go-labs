package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-labs/pkg/google/auth/server"
	"github.com/go-go-golems/go-go-labs/pkg/google/auth/store"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/api/calendar/v3"
)

const (
	tokenFile      = "token.json"
	credentialsDir = ".gcal"
)

var (
	scopes = []string{
		calendar.CalendarReadonlyScope,
		calendar.CalendarEventsScope,
	}
)

// NewGoogleClient creates a new authenticated Google Calendar client
func NewGoogleClient(ctx context.Context, opts ...Option) (*http.Client, error) {
	// Create authenticator with options
	auth, err := NewAuthenticator(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Run authentication flow
	result, err := auth.Authenticate(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	// Return the HTTP client
	return result.Client.Client(ctx, result.Token), nil
}

// RemoveToken removes the stored token file
func RemoveToken() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	tokenStore := store.NewFileTokenStore(
		filepath.Join(home, credentialsDir, tokenFile),
		0600,
	)

	if err := tokenStore.Clear(context.Background()); err != nil {
		return fmt.Errorf("failed to clear token: %w", err)
	}

	return nil
}

// CheckAuthStatus checks if a valid token exists
func CheckAuthStatus() (bool, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return false, err
	}

	tokenStore := store.NewFileTokenStore(
		filepath.Join(home, credentialsDir, tokenFile),
		0600,
	)

	token, err := tokenStore.Load(context.Background())
	if err != nil {
		if errors.Is(err, store.ErrTokenNotFound) {
			return false, nil
		}
		return false, err
	}

	return token.Valid(), nil
}

// Authenticator handles OAuth2 authentication flow
type Authenticator struct {
	config     *config
	oauthConf  *oauth2.Config
	logger     zerolog.Logger
	state      string
	resultChan chan *oauth2.Token
	errChan    chan error
}

// NewAuthenticator creates a new authenticator with the given options
func NewAuthenticator(opts ...Option) (*Authenticator, error) {
	cfg := defaultConfig()

	// Apply all options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Set default logger if not provided
	logger := cfg.logger
	if logger == nil {
		l := log.With().Str("component", "oauth2-authenticator").Logger()
		logger = &l
	}

	// Verify we have credentials
	if cfg.oauthConfig == nil {
		return nil, errors.New("no credentials provided: must provide credentials file, JSON, or OAuth2 config")
	}

	auth := &Authenticator{
		config:     cfg,
		oauthConf:  cfg.oauthConfig,
		logger:     *logger,
		resultChan: make(chan *oauth2.Token, 1),
		errChan:    make(chan error, 1),
	}

	// Generate secure state parameter
	state, err := auth.generateState()
	if err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	auth.state = state

	return auth, nil
}

// Authenticate starts the OAuth2 flow and returns the result
func (a *Authenticator) Authenticate(ctx context.Context) (*AuthResult, error) {
	// Check for existing token
	if a.config.tokenStore != nil {
		token, err := a.config.tokenStore.Load(ctx)
		if err == nil && token.Valid() {
			return &AuthResult{
				Client: a.oauthConf,
				Token:  token,
			}, nil
		}
		if err != nil && !errors.Is(err, store.ErrTokenNotFound) {
			return nil, fmt.Errorf("failed to load token: %w", err)
		}
	}

	// Set callback URL from server mode
	a.oauthConf.RedirectURL = a.config.serverMode.GetCallbackURL()

	// Create errgroup with cancellation
	g, gctx := errgroup.WithContext(ctx)

	// Create a channel to signal callback completion
	callbackDone := make(chan struct{})

	// Launch the server in its own goroutine
	g.Go(func() error {
		return a.config.serverMode.Setup(gctx, a.createCallbackHandler(callbackDone))
	})

	// Generate authorization URL
	authURL := a.oauthConf.AuthCodeURL(a.state, oauth2.AccessTypeOffline)
	a.logger.Info().Str("url", authURL).Msg("Please visit this URL to authorize the application")

	// Launch the token receiver in its own goroutine
	var result *AuthResult
	g.Go(func() error {
		select {
		case <-gctx.Done():
			return gctx.Err()
		case err := <-a.errChan:
			return err
		case token := <-a.resultChan:
			// Save token if store is configured
			if a.config.tokenStore != nil {
				if err := a.config.tokenStore.Save(gctx, token); err != nil {
					return fmt.Errorf("failed to save token: %w", err)
				}
			}
			result = &AuthResult{
				Client: a.oauthConf,
				Token:  token,
			}
			return nil
		}
	})

	// Launch cleanup goroutine
	g.Go(func() error {
		select {
		case <-gctx.Done():
			// Context was cancelled
			if err := a.config.serverMode.Cleanup(context.Background()); err != nil {
				a.logger.Warn().Err(err).Msg("Error during server cleanup")
			}
			return gctx.Err()
		case <-callbackDone:
			// Normal completion
			return a.config.serverMode.Cleanup(context.Background())
		}
	})

	// Wait for all goroutines to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return result, nil
}

// createCallbackHandler creates the HTTP handler for the OAuth2 callback
func (a *Authenticator) createCallbackHandler(done chan<- struct{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer close(done)

		// Verify state parameter
		if r.URL.Query().Get("state") != a.state {
			a.errChan <- errors.New("invalid state parameter")
			_, _ = w.Write(a.config.errorPage)
			return
		}

		// Get authorization code
		code := r.URL.Query().Get("code")
		if code == "" {
			a.errChan <- errors.New("no code parameter")
			_, _ = w.Write(a.config.errorPage)
			return
		}

		// Exchange code for token
		token, err := a.oauthConf.Exchange(r.Context(), code)
		if err != nil {
			a.errChan <- fmt.Errorf("failed to exchange token: %w", err)
			_, _ = w.Write(a.config.errorPage)
			return
		}

		a.resultChan <- token
		_, _ = w.Write(a.config.successPage)
	}
}

// generateState generates a random state parameter
func (a *Authenticator) generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Clear removes any stored token
func (a *Authenticator) Clear(ctx context.Context) error {
	if a.config.tokenStore != nil {
		return a.config.tokenStore.Clear(ctx)
	}
	return nil
}

// CreateOptionsFromSettings creates authenticator options from AuthSettings
func CreateOptionsFromSettings(s *AuthSettings) ([]Option, error) {
	var opts []Option

	// Expand ~ to home directory if present in credentials file path
	if strings.HasPrefix(s.CredentialsFile, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		s.CredentialsFile = filepath.Join(home, s.CredentialsFile[1:])
	}

	// Add credentials file option
	opts = append(opts, WithCredentialsFile(s.CredentialsFile))

	// Add server mode option
	opts = append(opts, WithServerMode(server.NewStandaloneServer(s.ServerPort, s.CallbackPath)))

	// Add timeout option
	opts = append(opts, WithTimeout(time.Duration(s.Timeout)*time.Minute))

	// Add scopes
	opts = append(opts, WithScopes(scopes...))

	return opts, nil
}

// CreateTokenStoreFromSettings creates a token store from settings
func CreateTokenStoreFromSettings(s *AuthSettings, dbs *DBAuthSettings, db interface{}) (store.TokenStore, error) {
	switch s.TokenStoreType {
	case "file":
		return store.NewFileTokenStore(s.TokenStorePath, os.FileMode(s.TokenStorePerms)), nil
	case "database":
		if dbs == nil {
			return nil, fmt.Errorf("database settings required for database token store")
		}
		if db == nil {
			return nil, fmt.Errorf("database connection required for database token store")
		}
		sqlDB, ok := db.(*sql.DB)
		if !ok {
			sqlxDB, ok := db.(*sqlx.DB)
			if !ok {
				return nil, fmt.Errorf("database connection must be *sql.DB or *sqlx.DB")
			}
			sqlDB = sqlxDB.DB
		}
		return store.NewMentoDatabaseTokenStore(sqlDB,
			store.WithUserID(dbs.UserID),
			store.WithProvider(dbs.Provider),
			store.WithScopes(dbs.Scopes),
			store.WithTeamID(dbs.TeamID),
			store.WithUserAppID(dbs.UserAppID),
			store.WithAppID(dbs.AppID),
		), nil
	default:
		return nil, fmt.Errorf("unsupported token store type: %s", s.TokenStoreType)
	}
}
