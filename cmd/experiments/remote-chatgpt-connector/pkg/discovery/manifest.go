package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// ManifestService provides ChatGPT plugin discovery endpoints for OIDC
type ManifestService struct {
	config types.Config
	logger zerolog.Logger
}

// NewManifestService creates a new manifest service
func NewManifestService(config types.Config) *ManifestService {
	return &ManifestService{
		config: config,
		logger: log.With().Str("component", "manifest").Logger(),
	}
}

// ServeManifest serves the .well-known/ai-plugin.json endpoint
func (m *ManifestService) ServeManifest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Get the base URL (protocol + host)
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)
	
	manifest := map[string]interface{}{
		"name":        "Go OIDC MCP Connector", 
		"description": "Self-contained MCP server with OIDC authentication and dynamic client registration",
		"version":     "0.2.0",
		"auth": map[string]interface{}{
			"type":               "oauth",
			"authorization_url":  baseURL + "/authorize",
			"token_url":          baseURL + "/token",
			"scopes":             []string{"openid", "offline_access"},
			// Note: We don't include client_id here since we support dynamic registration
		},
		"api": map[string]interface{}{
			"type": "mcp",
			"url":  baseURL + "/sse",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	
	if err := json.NewEncoder(w).Encode(manifest); err != nil {
		m.logger.Error().
			Err(err).
			Str("host", r.Host).
			Dur("duration", time.Since(start)).
			Msg("failed to encode plugin manifest")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	m.logger.Info().
		Str("user_agent", r.UserAgent()).
		Str("host", r.Host).
		Int("status_code", http.StatusOK).
		Dur("duration", time.Since(start)).
		Msg("plugin manifest served successfully")
}

// ServeOAuthConfig serves the .well-known/oauth-authorization-server endpoint for OIDC discovery
func (m *ManifestService) ServeOAuthConfig(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Get the base URL (protocol + host)
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, r.Host)
	
	oauthConfig := map[string]interface{}{
		"issuer":                                baseURL,
		"authorization_endpoint":                baseURL + "/authorize",
		"token_endpoint":                        baseURL + "/token",
		"registration_endpoint":                 baseURL + "/register",
		"userinfo_endpoint":                     baseURL + "/userinfo",
		"jwks_uri":                             baseURL + "/.well-known/jwks.json",
		"scopes_supported":                     []string{"openid", "offline_access"},
		"response_types_supported":             []string{"code"},
		"response_modes_supported":             []string{"query", "fragment"},
		"grant_types_supported":                []string{"authorization_code", "refresh_token"},
		"subject_types_supported":              []string{"public"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_post", "client_secret_basic", "none"},
		"code_challenge_methods_supported":     []string{"S256", "plain"},
		"claims_supported":                     []string{"sub", "iss", "aud", "exp", "iat", "auth_time"},
		"claim_types_supported":                []string{"normal"},
		"request_parameter_supported":          false,
		"request_uri_parameter_supported":      false,
		"require_request_uri_registration":    false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	
	if err := json.NewEncoder(w).Encode(oauthConfig); err != nil {
		m.logger.Error().
			Err(err).
			Str("host", r.Host).
			Dur("duration", time.Since(start)).
			Msg("failed to encode OAuth config")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	m.logger.Info().
		Str("user_agent", r.UserAgent()).
		Str("host", r.Host).
		Int("status_code", http.StatusOK).
		Dur("duration", time.Since(start)).
		Msg("OIDC discovery metadata served successfully")
}

// ServeJWKS serves the JSON Web Key Set for JWT verification
func (m *ManifestService) ServeJWKS(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// For now, return empty JWKS since we're using opaque tokens primarily
	// In production, this would contain the public keys for JWT verification
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			// Empty for now - would contain RSA public key for ID token verification
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours
	
	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		m.logger.Error().
			Err(err).
			Str("host", r.Host).
			Dur("duration", time.Since(start)).
			Msg("failed to encode JWKS")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	m.logger.Info().
		Str("user_agent", r.UserAgent()).
		Str("host", r.Host).
		Int("status_code", http.StatusOK).
		Dur("duration", time.Since(start)).
		Msg("JWKS served successfully")
}

// ServeUserInfo serves basic user information (for OIDC compliance)
func (m *ManifestService) ServeUserInfo(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	
	// Extract bearer token (this would normally be validated)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.Header().Set("WWW-Authenticate", `Bearer realm="MCP"`)
		http.Error(w, "Missing access token", http.StatusUnauthorized)
		return
	}

	// For now, return basic user info
	// In production, validate token and return actual user claims
	userInfo := map[string]interface{}{
		"sub":  "wesen-user-id",
		"name": "wesen",
		"preferred_username": "wesen",
	}

	w.Header().Set("Content-Type", "application/json")
	
	if err := json.NewEncoder(w).Encode(userInfo); err != nil {
		m.logger.Error().
			Err(err).
			Str("host", r.Host).
			Dur("duration", time.Since(start)).
			Msg("failed to encode user info")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	m.logger.Info().
		Str("user_agent", r.UserAgent()).
		Str("host", r.Host).
		Int("status_code", http.StatusOK).
		Dur("duration", time.Since(start)).
		Msg("user info served successfully")
}
