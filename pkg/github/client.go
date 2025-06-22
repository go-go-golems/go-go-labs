package github

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Client represents a GitHub GraphQL client
type Client struct {
	client *graphql.Client
	token  string
}

// NewClient creates a new GitHub GraphQL client
func NewClient() (*Client, error) {
	start := time.Now()
	log.Debug().Msg("Starting GitHub client initialization")

	// Log configuration parameters
	log.Debug().
		Str("endpoint", "https://api.github.com/graphql").
		Msg("Configuring GraphQL endpoint")

	// Token validation with debug logging
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Error().Msg("GITHUB_TOKEN environment variable not set")
		return nil, errors.New("GITHUB_TOKEN environment variable not set")
	}

	// Log token validation (mask the token for security)
	tokenMasked := maskToken(token)
	log.Debug().
		Str("token_masked", tokenMasked).
		Int("token_length", len(token)).
		Msg("Token validation successful")

	// HTTP client setup
	log.Debug().Msg("Setting up GraphQL client")
	client := graphql.NewClient("https://api.github.com/graphql")

	// Log client configuration
	log.Debug().
		Str("client_type", "graphql.Client").
		Msg("GraphQL client configured")

	// Create client instance
	githubClient := &Client{
		client: client,
		token:  token,
	}

	// Log successful initialization with timing
	duration := time.Since(start)
	log.Debug().
		Dur("init_duration", duration).
		Msg("GitHub client initialization completed successfully")

	return githubClient, nil
}

// maskToken masks the token for secure logging
func maskToken(token string) string {
	if len(token) <= 8 {
		return strings.Repeat("*", len(token))
	}
	return token[:4] + strings.Repeat("*", len(token)-8) + token[len(token)-4:]
}

// ExecuteQuery executes a GraphQL query with authentication
func (c *Client) ExecuteQuery(ctx context.Context, query string, variables map[string]interface{}, response interface{}) error {
	start := time.Now()

	// Log query preparation
	log.Debug().
		Str("operation", "execute_query").
		Msg("Starting GraphQL query execution")

	req := graphql.NewRequest(query)

	// Set variables with logging
	log.Debug().
		Int("variable_count", len(variables)).
		Interface("variables", variables).
		Msg("Setting query variables")

	for key, value := range variables {
		req.Var(key, value)
		log.Debug().
			Str("var_key", key).
			Interface("var_value", value).
			Msg("Variable set")
	}

	// Set authorization header with logging
	log.Debug().
		Str("token_masked", maskToken(c.token)).
		Msg("Setting authorization header")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Log the full request details
	log.Debug().
		Str("query", query).
		Interface("variables", variables).
		Str("method", "POST").
		Str("endpoint", "https://api.github.com/graphql").
		Msg("Executing GraphQL query")

	// Execute with performance timing
	execStart := time.Now()
	if err := c.client.Run(ctx, req, response); err != nil {
		execDuration := time.Since(execStart)
		totalDuration := time.Since(start)

		// Log detailed error information
		log.Error().
			Err(err).
			Str("query", query).
			Interface("variables", variables).
			Dur("exec_duration", execDuration).
			Dur("total_duration", totalDuration).
			Msg("GraphQL query failed")

		// Log error recovery attempt
		log.Debug().
			Str("error_type", "graphql_execution").
			Msg("Attempting error recovery")

		return errors.Wrap(err, "GraphQL query failed")
	}

	// Log successful execution with performance metrics
	execDuration := time.Since(execStart)
	totalDuration := time.Since(start)

	log.Debug().
		Interface("response", response).
		Dur("exec_duration", execDuration).
		Dur("total_duration", totalDuration).
		Msg("GraphQL query successful")

	return nil
}

// GetViewer returns the authenticated user information
func (c *Client) GetViewer(ctx context.Context) (*Viewer, error) {
	start := time.Now()

	log.Debug().
		Str("operation", "get_viewer").
		Msg("Starting viewer query")

	query := `
		query {
			viewer {
				login
				name
				email
			}
		}
	`

	log.Debug().
		Str("query_type", "viewer").
		Int("query_length", len(query)).
		Msg("Viewer query prepared")

	var resp struct {
		Viewer Viewer
	}

	if err := c.ExecuteQuery(ctx, query, nil, &resp); err != nil {
		duration := time.Since(start)
		log.Error().
			Err(err).
			Dur("duration", duration).
			Str("operation", "get_viewer").
			Msg("Failed to get viewer information")
		return nil, err
	}

	duration := time.Since(start)
	log.Debug().
		Str("viewer_login", resp.Viewer.Login).
		Str("viewer_name", resp.Viewer.Name).
		Bool("has_email", resp.Viewer.Email != "").
		Dur("duration", duration).
		Msg("Viewer information retrieved successfully")

	return &resp.Viewer, nil
}

// Viewer represents the authenticated user
type Viewer struct {
	Login string `json:"login"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
