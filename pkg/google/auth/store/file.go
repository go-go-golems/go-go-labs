package store

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"golang.org/x/oauth2"
)

// FileTokenStore implements TokenStore using a local file
type FileTokenStore struct {
	path string
	perm os.FileMode
}

// NewFileTokenStore creates a new file-based token store
func NewFileTokenStore(path string, perm os.FileMode) *FileTokenStore {
	// Expand ~ to home directory if present
	if path[:1] == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get user home directory")
			// Fall back to original path if home dir cannot be determined
		} else {
			path = filepath.Join(home, path[1:])
		}
	}

	return &FileTokenStore{
		path: path,
		perm: perm,
	}
}

// Save implements TokenStore
func (s *FileTokenStore) Save(ctx context.Context, token *oauth2.Token) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, s.perm); err != nil {
		return errors.Wrap(err, "failed to create token directory")
	}

	// Create or truncate the file
	f, err := os.OpenFile(s.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, s.perm)
	if err != nil {
		return errors.Wrap(err, "failed to create token file")
	}
	defer f.Close()

	log.Debug().Str("path", s.path).Msg("Saving token to file")

	return json.NewEncoder(f).Encode(token)
}

// Load implements TokenStore
func (s *FileTokenStore) Load(ctx context.Context) (*oauth2.Token, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrTokenNotFound
		}
		return nil, errors.Wrap(err, "failed to open token file")
	}
	defer f.Close()

	token := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, errors.Wrap(err, "failed to decode token")
	}

	return token, nil
}

// Clear implements TokenStore
func (s *FileTokenStore) Clear(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := os.Remove(s.path)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "failed to remove token file")
	}

	return nil
}
