package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
)

const (
	githubAPIBaseURL = "https://api.github.com"
	githubUserAgent  = "go-mcp-connector/1.0"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GitHubAuthValidator implements GitHub OAuth token validation
type GitHubAuthValidator struct {
	clientID     string
	clientSecret string
	allowedLogin string
	httpClient   *http.Client
	logger       zerolog.Logger
}

// NewGitHubAuthValidator creates a new GitHub auth validator
func NewGitHubAuthValidator(clientID, clientSecret, allowedLogin string, logger zerolog.Logger) *GitHubAuthValidator {
	validator := &GitHubAuthValidator{
		clientID:     clientID,
		clientSecret: clientSecret,
		allowedLogin: allowedLogin,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.With().Str("component", "github_auth").Logger(),
	}

	// Log validator creation with security-aware details
	validator.logger.Info().
		Bool("has_client_id", len(clientID) > 0).
		Bool("has_client_secret", len(clientSecret) > 0).
		Str("allowed_login", allowedLogin).
		Dur("http_timeout", 10*time.Second).
		Msg("github auth validator created")

	// Validate configuration
	if len(clientID) == 0 {
		validator.logger.Warn().Msg("github client ID is empty")
	}
	if len(clientSecret) == 0 {
		validator.logger.Warn().Msg("github client secret is empty")
	}
	if len(allowedLogin) == 0 {
		validator.logger.Warn().Msg("no allowed login specified - all users will be rejected")
	}

	return validator
}

// ValidateToken validates a GitHub OAuth token and checks user allowlist
func (g *GitHubAuthValidator) ValidateToken(ctx context.Context, token string) (*types.UserInfo, error) {
	start := time.Now()

	// Remove Bearer prefix if present
	originalToken := token
	token = strings.TrimPrefix(token, "Bearer ")

	// Token format validation and redaction for logging
	tokenPrefix := ""
	if len(token) >= 8 {
		tokenPrefix = token[:8] + "..."
	} else if len(token) > 0 {
		tokenPrefix = token[:min(len(token), 4)] + "..."
	}

	g.logger.Debug().
		Str("token_prefix", tokenPrefix).
		Bool("has_bearer_prefix", originalToken != token).
		Int("token_length", len(token)).
		Msg("starting github token validation")

	// Basic token format validation
	if len(token) == 0 {
		g.logger.Debug().Msg("empty token provided")
		return nil, fmt.Errorf("empty github token")
	}

	if len(token) < 20 {
		g.logger.Debug().
			Int("token_length", len(token)).
			Msg("token appears too short for github format")
	}

	// Create request with timing
	reqStart := time.Now()
	req, err := http.NewRequestWithContext(ctx, "GET", githubAPIBaseURL+"/user", nil)
	if err != nil {
		g.logger.Error().
			Err(err).
			Str("endpoint", githubAPIBaseURL+"/user").
			Dur("req_creation_time", time.Since(reqStart)).
			Msg("failed to create github api request")
		return nil, fmt.Errorf("failed to create github request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", githubUserAgent)

	g.logger.Debug().
		Str("endpoint", githubAPIBaseURL+"/user").
		Str("user_agent", githubUserAgent).
		Str("accept_header", "application/vnd.github+json").
		Msg("making github api request")

	// Make API call with timing
	apiStart := time.Now()
	resp, err := g.httpClient.Do(req)
	apiDuration := time.Since(apiStart)

	if err != nil {
		g.logger.Error().
			Err(err).
			Str("token_prefix", tokenPrefix).
			Dur("api_call_duration", apiDuration).
			Str("endpoint", githubAPIBaseURL+"/user").
			Msg("github api request failed")
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	// Log response details
	g.logger.Debug().
		Int("status_code", resp.StatusCode).
		Str("status", resp.Status).
		Dur("api_call_duration", apiDuration).
		Str("content_type", resp.Header.Get("Content-Type")).
		Str("x_ratelimit_limit", resp.Header.Get("X-RateLimit-Limit")).
		Str("x_ratelimit_remaining", resp.Header.Get("X-RateLimit-Remaining")).
		Str("x_ratelimit_reset", resp.Header.Get("X-RateLimit-Reset")).
		Str("x_github_request_id", resp.Header.Get("X-GitHub-Request-Id")).
		Msg("received github api response")

	// Handle non-200 responses with detailed logging
	if resp.StatusCode != http.StatusOK {
		rateLimitRemaining := resp.Header.Get("X-RateLimit-Remaining")
		rateLimitReset := resp.Header.Get("X-RateLimit-Reset")

		g.logger.Warn().
			Int("status_code", resp.StatusCode).
			Str("status", resp.Status).
			Str("token_prefix", tokenPrefix).
			Dur("api_call_duration", apiDuration).
			Str("rate_limit_remaining", rateLimitRemaining).
			Str("rate_limit_reset", rateLimitReset).
			Str("github_request_id", resp.Header.Get("X-GitHub-Request-Id")).
			Msg("github api returned non-200 status")

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			g.logger.Debug().
				Str("token_prefix", tokenPrefix).
				Msg("token validation failed - unauthorized")
			return nil, fmt.Errorf("invalid or expired github token")
		case http.StatusForbidden:
			g.logger.Debug().
				Str("token_prefix", tokenPrefix).
				Str("rate_limit_remaining", rateLimitRemaining).
				Msg("token validation failed - forbidden or rate limited")
			if rateLimitRemaining == "0" {
				return nil, fmt.Errorf("github api rate limit exceeded")
			}
			return nil, fmt.Errorf("github token lacks required permissions")
		case http.StatusTooManyRequests:
			g.logger.Warn().
				Str("token_prefix", tokenPrefix).
				Str("rate_limit_reset", rateLimitReset).
				Msg("github api rate limit exceeded")
			return nil, fmt.Errorf("github api rate limit exceeded")
		default:
			g.logger.Debug().
				Str("token_prefix", tokenPrefix).
				Msg("github api returned unexpected error status")
			return nil, fmt.Errorf("github api error: %s", resp.Status)
		}
	}

	// Parse GitHub user response
	var ghUser struct {
		ID    int    `json:"id"`
		Login string `json:"login"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}

	parseStart := time.Now()
	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		g.logger.Error().
			Err(err).
			Str("token_prefix", tokenPrefix).
			Dur("parse_duration", time.Since(parseStart)).
			Str("content_type", resp.Header.Get("Content-Type")).
			Msg("failed to decode github user response")
		return nil, fmt.Errorf("failed to decode github user response: %w", err)
	}

	g.logger.Debug().
		Str("github_login", ghUser.Login).
		Int("github_id", ghUser.ID).
		Str("github_name", ghUser.Name).
		Str("github_email", ghUser.Email).
		Str("token_prefix", tokenPrefix).
		Dur("parse_duration", time.Since(parseStart)).
		Msg("successfully parsed github user info")

	// Check if user is in allowlist
	if ghUser.Login != g.allowedLogin {
		g.logger.Warn().
			Str("github_login", ghUser.Login).
			Str("allowed_login", g.allowedLogin).
			Int("github_id", ghUser.ID).
			Str("token_prefix", tokenPrefix).
			Dur("total_validation_time", time.Since(start)).
			Msg("github user not in allowlist - authorization denied")
		return nil, fmt.Errorf("github user %s not authorized", ghUser.Login)
	}

	totalDuration := time.Since(start)
	g.logger.Info().
		Str("github_login", ghUser.Login).
		Int("github_id", ghUser.ID).
		Str("token_prefix", tokenPrefix).
		Dur("total_validation_time", totalDuration).
		Dur("api_call_duration", apiDuration).
		Msg("github token validated successfully")

	// Log performance metrics for monitoring
	g.logger.Debug().
		Dur("total_validation_time", totalDuration).
		Dur("api_call_duration", apiDuration).
		Float64("api_call_pct", float64(apiDuration.Nanoseconds())/float64(totalDuration.Nanoseconds())*100).
		Msg("github auth validation performance metrics")

	// Return UserInfo
	return &types.UserInfo{
		ID:       fmt.Sprintf("%d", ghUser.ID),
		Login:    ghUser.Login,
		Email:    ghUser.Email,
		Verified: true, // GitHub tokens are inherently verified
	}, nil
}

// GetAuthEndpoints returns GitHub OAuth endpoints
func (g *GitHubAuthValidator) GetAuthEndpoints() types.AuthEndpoints {
	endpoints := types.AuthEndpoints{
		AuthorizeURL: "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		Scopes:       []string{"read:user"},
	}

	g.logger.Debug().
		Str("authorize_url", endpoints.AuthorizeURL).
		Str("token_url", endpoints.TokenURL).
		Strs("scopes", endpoints.Scopes).
		Msg("returning github oauth endpoints")

	return endpoints
}

// MockAuthValidator is a simple mock for testing other components
type MockAuthValidator struct {
	shouldSucceed bool
	userInfo      *types.UserInfo
	logger        zerolog.Logger
}

// NewMockAuthValidator creates a new mock auth validator
func NewMockAuthValidator(shouldSucceed bool, logger zerolog.Logger) *MockAuthValidator {
	return &MockAuthValidator{
		shouldSucceed: shouldSucceed,
		userInfo: &types.UserInfo{
			ID:       "mock-123",
			Login:    "mockuser",
			Email:    "mock@example.com",
			Verified: true,
		},
		logger: logger.With().Str("component", "mock_auth").Logger(),
	}
}

// ValidateToken validates a token (mock implementation)
func (m *MockAuthValidator) ValidateToken(ctx context.Context, token string) (*types.UserInfo, error) {
	tokenPrefix := ""
	if len(token) >= 8 {
		tokenPrefix = token[:8] + "..."
	} else if len(token) > 0 {
		tokenPrefix = token[:min(len(token), 4)] + "..."
	}

	m.logger.Debug().
		Bool("should_succeed", m.shouldSucceed).
		Str("token_prefix", tokenPrefix).
		Int("token_length", len(token)).
		Msg("mock validating token")

	if !m.shouldSucceed {
		m.logger.Debug().
			Str("token_prefix", tokenPrefix).
			Msg("mock validation configured to fail")
		return nil, fmt.Errorf("mock validation failed")
	}

	m.logger.Debug().
		Str("mock_user_id", m.userInfo.ID).
		Str("mock_user_login", m.userInfo.Login).
		Msg("mock validation succeeded")

	return m.userInfo, nil
}

// GetAuthEndpoints returns mock OAuth endpoints
func (m *MockAuthValidator) GetAuthEndpoints() types.AuthEndpoints {
	return types.AuthEndpoints{
		AuthorizeURL: "https://mock.example.com/authorize",
		TokenURL:     "https://mock.example.com/token",
		Scopes:       []string{"read:user"},
	}
}

// Ensure interfaces are implemented
var _ types.AuthValidator = (*GitHubAuthValidator)(nil)
var _ types.AuthValidator = (*MockAuthValidator)(nil)
