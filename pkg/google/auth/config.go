package auth

import (
	"os"
	"time"

	"github.com/go-go-golems/go-go-labs/pkg/google/auth/server"
	"github.com/go-go-golems/go-go-labs/pkg/google/auth/store"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Option is a function that configures the authenticator
type Option func(*config) error

// config holds the authenticator configuration
type config struct {
	oauthConfig *oauth2.Config

	tokenStore  store.TokenStore
	serverMode  server.ServerMode
	scopes      []string
	timeout     time.Duration
	logger      *zerolog.Logger
	successPage []byte
	errorPage   []byte
}

// WithCredentialsFile sets the OAuth2 credentials from a file
func WithCredentialsFile(path string) Option {
	return func(c *config) error {
		b, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrap(err, "unable to read credentials file")
		}

		cfg, err := google.ConfigFromJSON(b, c.scopes...)
		if err != nil {
			return errors.Wrap(err, "unable to parse credentials file")
		}
:
		c.oauthConfig = cfg
		return nil
	}
}

// WithCredentialsJSON sets the OAuth2 credentials from JSON bytes
func WithCredentialsJSON(json []byte) Option {
	return func(c *config) error {
		cfg, err := google.ConfigFromJSON(json, c.scopes...)
		if err != nil {
			return errors.Wrap(err, "unable to parse credentials JSON")
		}

		c.oauthConfig = cfg
		return nil
	}
}

// WithOAuthConfig sets a pre-configured OAuth2 config
func WithOAuthConfig(cfg *oauth2.Config) Option {
	return func(c *config) error {
		if cfg == nil {
			return errors.New("oauth config cannot be nil")
		}
		c.oauthConfig = cfg
		return nil
	}
}

// WithTokenStore sets the token storage implementation
func WithTokenStore(store store.TokenStore) Option {
	return func(c *config) error {
		if store == nil {
			return errors.New("token store cannot be nil")
		}
		c.tokenStore = store
		return nil
	}
}

// WithServerMode sets the server mode implementation
func WithServerMode(mode server.ServerMode) Option {
	return func(c *config) error {
		if mode == nil {
			return errors.New("server mode cannot be nil")
		}
		c.serverMode = mode
		return nil
	}
}

// WithScopes sets the OAuth2 scopes
func WithScopes(scopes ...string) Option {
	return func(c *config) error {
		if len(scopes) == 0 {
			return errors.New("at least one scope must be provided")
		}
		c.scopes = scopes

		// If we already have an OAuth config, append the scopes
		if c.oauthConfig != nil {
			c.oauthConfig.Scopes = append(c.oauthConfig.Scopes, scopes...)
		}
		return nil
	}
}

// WithTimeout sets the maximum duration to wait for authentication
func WithTimeout(duration time.Duration) Option {
	return func(c *config) error {
		if duration <= 0 {
			return errors.New("timeout must be positive")
		}
		c.timeout = duration
		return nil
	}
}

// WithLogger sets a custom logger
func WithLogger(logger *zerolog.Logger) Option {
	return func(c *config) error {
		if logger == nil {
			return errors.New("logger cannot be nil")
		}
		c.logger = logger
		return nil
	}
}

// WithSuccessPage sets a custom success page
func WithSuccessPage(page []byte) Option {
	return func(c *config) error {
		if len(page) == 0 {
			return errors.New("success page cannot be empty")
		}
		c.successPage = page
		return nil
	}
}

// WithErrorPage sets a custom error page
func WithErrorPage(page []byte) Option {
	return func(c *config) error {
		if len(page) == 0 {
			return errors.New("error page cannot be empty")
		}
		c.errorPage = page
		return nil
	}
}

// defaultConfig returns the default configuration
func defaultConfig() *config {
	return &config{
		timeout:     5 * time.Minute,
		serverMode:  server.NewStandaloneServer(8080, "/callback"),
		successPage: []byte("Authentication successful! You can close this window."),
		errorPage:   []byte("Authentication failed. Please try again."),
	}
}

// AuthResult contains the authentication result
type AuthResult struct {
	Client *oauth2.Config
	Token  *oauth2.Token
}
