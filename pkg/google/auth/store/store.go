package store

import (
	"context"
	"errors"

	"golang.org/x/oauth2"
)

// ErrTokenNotFound is returned when no token exists in the store
var ErrTokenNotFound = errors.New("token not found")

// TokenStore defines the interface for storing and retrieving OAuth2 tokens
type TokenStore interface {
	// Save stores the OAuth2 token
	// Context can be used for timeouts, cancellation, or external storage systems
	Save(ctx context.Context, token *oauth2.Token) error

	// Load retrieves a previously stored OAuth2 token
	// Returns ErrTokenNotFound if no token exists
	Load(ctx context.Context) (*oauth2.Token, error)

	// Clear removes the stored token
	// Should not return an error if no token exists
	Clear(ctx context.Context) error
}
