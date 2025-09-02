package auth

import (
	"fmt"
	"os"

	"context"

	clay_sql "github.com/go-go-golems/clay/pkg/sql"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/go-go-labs/pkg/google/auth/store"
	"github.com/pkg/errors"
)

// GetOAuthTokenStoreLayers returns all layers required for OAuth token store configuration
func GetOAuthTokenStoreLayers() (*layers.ParameterLayers, error) {
	authLayer, err := NewAuthParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create auth parameter layer: %w", err)
	}

	dbAuthLayer, err := NewDBAuthParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create database auth parameter layer: %w", err)
	}

	sqlConnectionLayer, err := clay_sql.NewSqlConnectionParameterLayer(layers.WithPrefix("db-"))
	if err != nil {
		return nil, fmt.Errorf("could not create SQL connection layer: %w", err)
	}

	dbtLayer, err := clay_sql.NewDbtParameterLayer()
	if err != nil {
		return nil, fmt.Errorf("could not create DBT layer: %w", err)
	}

	return layers.NewParameterLayers(
		layers.WithLayers(
			authLayer,
			dbAuthLayer,
			sqlConnectionLayer,
			dbtLayer,
		),
	), nil
}

// CreateTokenStoreFromLayers creates a token store from parsed layers
func CreateTokenStoreFromLayers(parsedLayers *layers.ParsedLayers) (store.TokenStore, error) {
	// Get auth settings
	s := &AuthSettings{}
	if err := parsedLayers.InitializeStruct(AuthSlug, s); err != nil {
		return nil, fmt.Errorf("failed to initialize auth settings: %w", err)
	}

	// Get database settings if needed
	var dbs *DBAuthSettings
	if s.TokenStoreType == "database" {
		dbs = &DBAuthSettings{}
		if err := parsedLayers.InitializeStruct(DBAuthSlug, dbs); err != nil {
			return nil, fmt.Errorf("failed to initialize database settings: %w", err)
		}

		// Get database connection from SQL layers
		db, err := clay_sql.OpenDatabaseFromSqlConnectionLayer(
			context.Background(),
			parsedLayers,
			clay_sql.SqlConnectionSlug,
			clay_sql.DbtSlug,
		)
		if err != nil {
			return nil, errors.Wrap(err, "could not open database connection")
		}

		// Create token store with database connection
		tokenStore, err := CreateTokenStoreFromSettings(s, dbs, db)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to create token store: %w", err)
		}
		return tokenStore, nil
	}

	// Create file token store
	return store.NewFileTokenStore(s.TokenStorePath, os.FileMode(s.TokenStorePerms)), nil
}

// CreateAuthenticatorFromLayers creates an authenticator from parsed layers
func CreateAuthenticatorFromLayers(parsedLayers *layers.ParsedLayers) (*Authenticator, error) {
	// Get auth settings
	s := &AuthSettings{}
	if err := parsedLayers.InitializeStruct(AuthSlug, s); err != nil {
		return nil, fmt.Errorf("failed to initialize auth settings: %w", err)
	}

	// Create authenticator options
	opts, err := CreateOptionsFromSettings(s)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator options: %w", err)
	}

	// Create token store
	tokenStore, err := CreateTokenStoreFromLayers(parsedLayers)
	if err != nil {
		return nil, fmt.Errorf("failed to create token store: %w", err)
	}

	// Add token store to options
	opts = append(opts, WithTokenStore(tokenStore))

	// Create authenticator
	return NewAuthenticator(opts...)
}
