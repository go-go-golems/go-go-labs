package github

import (
	"context"
	"os"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Client represents a GitHub GraphQL client
type Client struct {
	client *graphql.Client
	token  string
	logger zerolog.Logger
}

// NewClient creates a new GitHub GraphQL client
func NewClient(logger zerolog.Logger) (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, errors.New("GITHUB_TOKEN environment variable not set")
	}

	client := graphql.NewClient("https://api.github.com/graphql")

	return &Client{
		client: client,
		token:  token,
		logger: logger,
	}, nil
}

// ExecuteQuery executes a GraphQL query with authentication
func (c *Client) ExecuteQuery(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	req := graphql.NewRequest(query)

	// Set variables
	for key, value := range variables {
		req.Var(key, value)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+c.token)

	c.logger.Debug().
		Str("query", query).
		Interface("variables", variables).
		Msg("Executing GraphQL query")

	if err := c.client.Run(ctx, req, response); err != nil {
		c.logger.Error().
			Err(err).
			Str("query", query).
			Interface("variables", variables).
			Msg("GraphQL query failed")
		return errors.Wrap(err, "GraphQL query failed")
	}

	c.logger.Debug().
		Interface("response", response).
		Msg("GraphQL query successful")

	return nil
}

// GetViewer returns the authenticated user information
func (c *Client) GetViewer(ctx context.Context) (*Viewer, error) {
	query := `
		query {
			viewer {
				login
				name
				email
			}
		}
	`

	var resp struct {
		Viewer Viewer
	}

	if err := c.ExecuteQuery(ctx, query, nil, &resp); err != nil {
		return nil, err
	}

	return &resp.Viewer, nil
}

// Viewer represents the authenticated user
type Viewer struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
