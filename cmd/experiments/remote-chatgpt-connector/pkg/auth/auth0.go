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

// Auth0AuthValidator implements Auth0 OAuth token validation
type Auth0AuthValidator struct {
	issuer       string
	audience     string
	clientID     string
	clientSecret string
	httpClient   *http.Client
	logger       zerolog.Logger
}

// NewAuth0AuthValidator creates a new Auth0 auth validator
func NewAuth0AuthValidator(issuer, audience, clientID, clientSecret string, logger zerolog.Logger) *Auth0AuthValidator {
	validator := &Auth0AuthValidator{
		issuer:       strings.TrimSuffix(issuer, "/"),
		audience:     audience,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.With().Str("component", "auth0_auth").Logger(),
	}

	// Log validator creation with security-aware details
	validator.logger.Info().
		Str("issuer", validator.issuer).
		Str("audience", validator.audience).
		Bool("client_id_set", clientID != "").
		Bool("client_secret_set", clientSecret != "").
		Msg("Auth0 auth validator created")

	return validator
}

// ValidateToken validates an Auth0 OAuth token
func (v *Auth0AuthValidator) ValidateToken(ctx context.Context, token string) (*types.UserInfo, error) {
	start := time.Now()
	requestID := fmt.Sprintf("validate-%d", time.Now().UnixNano())

	logger := v.logger.With().
		Str("operation", "validate_token").
		Str("request_id", requestID).
		Str("token_prefix", token[:min(8, len(token))]).
		Logger()

	logger.Debug().Msg("starting Auth0 token validation")

	if token == "" {
		logger.Warn().
			Dur("validation_duration", time.Since(start)).
			Msg("empty token provided")
		return nil, fmt.Errorf("empty token")
	}

	// Call Auth0 userinfo endpoint to validate token and get user info
	userinfoURL := fmt.Sprintf("%s/userinfo", v.issuer)
	
	req, err := http.NewRequestWithContext(ctx, "GET", userinfoURL, nil)
	if err != nil {
		logger.Error().
			Err(err).
			Str("userinfo_url", userinfoURL).
			Dur("validation_duration", time.Since(start)).
			Msg("failed to create userinfo request")
		return nil, fmt.Errorf("failed to create userinfo request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", "go-mcp-connector/1.0")

	logger.Debug().
		Str("userinfo_url", userinfoURL).
		Str("method", "GET").
		Msg("making userinfo API call")

	apiCallStart := time.Now()
	resp, err := v.httpClient.Do(req)
	apiCallDuration := time.Since(apiCallStart)

	if err != nil {
		logger.Error().
			Err(err).
			Str("userinfo_url", userinfoURL).
			Dur("api_call_duration", apiCallDuration).
			Dur("total_validation_duration", time.Since(start)).
			Msg("Auth0 userinfo API call failed")
		return nil, fmt.Errorf("Auth0 userinfo API call failed: %w", err)
	}
	defer resp.Body.Close()

	logger.Debug().
		Int("status_code", resp.StatusCode).
		Dur("api_call_duration", apiCallDuration).
		Int64("content_length", resp.ContentLength).
		Str("content_type", resp.Header.Get("Content-Type")).
		Msg("received userinfo API response")

	if resp.StatusCode != http.StatusOK {
		logger.Warn().
			Int("status_code", resp.StatusCode).
			Str("status", resp.Status).
			Dur("api_call_duration", apiCallDuration).
			Dur("total_validation_duration", time.Since(start)).
			Msg("Auth0 userinfo API returned non-200 status - token likely invalid")
		
		return nil, fmt.Errorf("invalid token: Auth0 returned status %d", resp.StatusCode)
	}

	var userInfo struct {
		Sub           string `json:"sub"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Picture       string `json:"picture"`
		Nickname      string `json:"nickname"`
	}

	parseStart := time.Now()
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		parseDuration := time.Since(parseStart)
		logger.Error().
			Err(err).
			Dur("parse_duration", parseDuration).
			Dur("total_validation_duration", time.Since(start)).
			Msg("failed to parse Auth0 userinfo response")
		return nil, fmt.Errorf("failed to parse userinfo response: %w", err)
	}
	parseDuration := time.Since(parseStart)

	totalDuration := time.Since(start)
	
	result := &types.UserInfo{
		ID:       userInfo.Sub,
		Login:    userInfo.Nickname,
		Email:    userInfo.Email,
		Verified: userInfo.EmailVerified,
	}

	if result.Login == "" {
		result.Login = userInfo.Name
	}

	logger.Info().
		Str("user_id", result.ID).
		Str("user_login", result.Login).
		Str("user_email", result.Email).
		Bool("email_verified", result.Verified).
		Dur("api_call_duration", apiCallDuration).
		Dur("parse_duration", parseDuration).
		Dur("total_validation_duration", totalDuration).
		Msg("Auth0 token validation successful")

	return result, nil
}

// GetAuthEndpoints returns Auth0 OAuth endpoints
func (v *Auth0AuthValidator) GetAuthEndpoints() types.AuthEndpoints {
	endpoints := types.AuthEndpoints{
		AuthorizeURL: fmt.Sprintf("%s/authorize", v.issuer),
		TokenURL:     fmt.Sprintf("%s/oauth/token", v.issuer),
		Scopes:       []string{"openid", "profile", "email"},
	}

	v.logger.Debug().
		Interface("auth_endpoints", endpoints).
		Msg("Auth0 auth endpoints generated")

	return endpoints
}
