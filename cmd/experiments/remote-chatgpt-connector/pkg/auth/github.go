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
	return &GitHubAuthValidator{
		clientID:     clientID,
		clientSecret: clientSecret,
		allowedLogin: allowedLogin,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.With().Str("component", "github_auth").Logger(),
	}
}

// ValidateToken validates a GitHub OAuth token and checks user allowlist
func (g *GitHubAuthValidator) ValidateToken(ctx context.Context, token string) (*types.UserInfo, error) {
	// Remove Bearer prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	g.logger.Debug().Msg("validating github token")

	// Call GitHub /user API to validate token and get user info
	req, err := http.NewRequestWithContext(ctx, "GET", githubAPIBaseURL+"/user", nil)
	if err != nil {
		g.logger.Error().Err(err).Msg("failed to create request")
		return nil, fmt.Errorf("failed to create github request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", githubUserAgent)

	resp, err := g.httpClient.Do(req)
	if err != nil {
		g.logger.Error().Err(err).Msg("github api request failed")
		return nil, fmt.Errorf("github api request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle non-200 responses
	if resp.StatusCode != http.StatusOK {
		g.logger.Warn().
			Int("status_code", resp.StatusCode).
			Str("status", resp.Status).
			Msg("github api returned non-200 status")

		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, fmt.Errorf("invalid or expired github token")
		case http.StatusForbidden:
			return nil, fmt.Errorf("github token lacks required permissions")
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("github api rate limit exceeded")
		default:
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

	if err := json.NewDecoder(resp.Body).Decode(&ghUser); err != nil {
		g.logger.Error().Err(err).Msg("failed to decode github user response")
		return nil, fmt.Errorf("failed to decode github user response: %w", err)
	}

	g.logger.Debug().
		Str("github_login", ghUser.Login).
		Int("github_id", ghUser.ID).
		Msg("got github user info")

	// Check if user is in allowlist
	if ghUser.Login != g.allowedLogin {
		g.logger.Warn().
			Str("github_login", ghUser.Login).
			Str("allowed_login", g.allowedLogin).
			Msg("github user not in allowlist")
		return nil, fmt.Errorf("github user %s not authorized", ghUser.Login)
	}

	g.logger.Info().
		Str("github_login", ghUser.Login).
		Msg("github token validated successfully")

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
	return types.AuthEndpoints{
		AuthorizeURL: "https://github.com/login/oauth/authorize",
		TokenURL:     "https://github.com/login/oauth/access_token",
		Scopes:       []string{"read:user"},
	}
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
	m.logger.Debug().Bool("should_succeed", m.shouldSucceed).Msg("mock validating token")

	if !m.shouldSucceed {
		return nil, fmt.Errorf("mock validation failed")
	}

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
