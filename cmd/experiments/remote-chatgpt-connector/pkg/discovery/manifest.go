package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Service implements the DiscoveryService interface
type Service struct {
	config *types.Config
}

// NewService creates a new discovery service
func NewService(config *types.Config) *Service {
	log.Debug().
		Interface("config", map[string]interface{}{
			"host": config.Host,
			"port": config.Port,
		}).
		Msg("creating new discovery service")

	return &Service{
		config: config,
	}
}

// PluginManifest represents the ChatGPT plugin manifest structure
type PluginManifest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Version     string      `json:"version"`
	Auth        AuthConfig  `json:"auth"`
	API         APIConfig   `json:"api"`
	Contact     ContactInfo `json:"contact,omitempty"`
	Legal       LegalInfo   `json:"legal,omitempty"`
}

// AuthConfig represents OAuth configuration for ChatGPT
type AuthConfig struct {
	Type             string   `json:"type"`
	AuthorizationURL string   `json:"authorization_url"`
	TokenURL         string   `json:"token_url"`
	Scopes           []string `json:"scopes"`
	ClientID         string   `json:"client_id,omitempty"`
}

// APIConfig represents the API configuration
type APIConfig struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Email string `json:"email,omitempty"`
}

// LegalInfo represents legal information
type LegalInfo struct {
	PrivacyURL string `json:"privacy_url,omitempty"`
	TermsURL   string `json:"terms_url,omitempty"`
}

// OAuthAuthorizationServer represents the OAuth authorization server metadata
type OAuthAuthorizationServer struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	ScopesSupported                   []string `json:"scopes_supported"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
	CodeChallengeMethodsSupported     []string `json:"code_challenge_methods_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"token_endpoint_auth_methods_supported"`
	RegistrationEndpoint              string   `json:"registration_endpoint,omitempty"`
}

// GetPluginManifest returns the ChatGPT plugin manifest
func (s *Service) GetPluginManifest() ([]byte, error) {
	return s.getPluginManifestWithHost("")
}

// getPluginManifestWithHost generates manifest with specific host
func (s *Service) getPluginManifestWithHost(host string) ([]byte, error) {
	start := time.Now()
	logger := log.With().
		Str("component", "discovery").
		Str("operation", "generate_manifest").
		Logger()

	logger.Debug().Msg("starting plugin manifest generation")

	manifest := PluginManifest{
		Name:        "Go SSE MCP Demo",
		Description: "Remote MCP server for ChatGPT integration with GitHub OAuth",
		Version:     "0.1.0",
		Auth: AuthConfig{
			Type:             "oauth",
			AuthorizationURL: "https://github.com/login/oauth/authorize",
			TokenURL:         "https://github.com/login/oauth/access_token",
			Scopes:           []string{"read:user"},
			ClientID:         s.config.GitHubClientID,
		},
		API: APIConfig{
			Type: "mcp", 
			URL:  s.getSSEURL(host),
		},
	}

	logger.Debug().
		Interface("manifest_structure", manifest).
		Msg("manifest structure created")

	// Time JSON serialization
	serializeStart := time.Now()
	data, err := json.MarshalIndent(manifest, "", "  ")
	serializeDuration := time.Since(serializeStart)

	if err != nil {
		logger.Error().
			Err(err).
			Dur("total_duration", time.Since(start)).
			Msg("failed to serialize plugin manifest to JSON")
		return nil, err
	}

	totalDuration := time.Since(start)
	logger.Debug().
		Int("json_size_bytes", len(data)).
		Dur("serialize_duration", serializeDuration).
		Dur("total_generation_duration", totalDuration).
		Str("json_preview", string(data[:min(200, len(data))])+"...").
		Msg("plugin manifest generation completed successfully")

	return data, nil
}

// GetOAuthConfig returns the OAuth authorization server configuration
func (s *Service) GetOAuthConfig() ([]byte, error) {
	start := time.Now()
	logger := log.With().
		Str("component", "discovery").
		Str("operation", "generate_oauth_config").
		Logger()

	logger.Debug().Msg("starting OAuth authorization server config generation")

	config := OAuthAuthorizationServer{
		Issuer:                "https://github.com",
		AuthorizationEndpoint: "https://github.com/login/oauth/authorize",
		TokenEndpoint:         "https://github.com/login/oauth/access_token",
		ScopesSupported: []string{
			"read:user",
			"user:email",
			"public_repo",
			"repo",
		},
		ResponseTypesSupported: []string{
			"code",
		},
		GrantTypesSupported: []string{
			"authorization_code",
		},
		CodeChallengeMethodsSupported: []string{
			"S256",
			"plain",
		},
		TokenEndpointAuthMethodsSupported: []string{
			"client_secret_post",
		},
	}

	logger.Debug().
		Interface("oauth_config_structure", config).
		Int("scopes_count", len(config.ScopesSupported)).
		Int("response_types_count", len(config.ResponseTypesSupported)).
		Int("grant_types_count", len(config.GrantTypesSupported)).
		Int("challenge_methods_count", len(config.CodeChallengeMethodsSupported)).
		Int("auth_methods_count", len(config.TokenEndpointAuthMethodsSupported)).
		Msg("OAuth config structure created")

	// Time JSON serialization
	serializeStart := time.Now()
	data, err := json.MarshalIndent(config, "", "  ")
	serializeDuration := time.Since(serializeStart)

	if err != nil {
		logger.Error().
			Err(err).
			Dur("total_duration", time.Since(start)).
			Msg("failed to serialize OAuth config to JSON")
		return nil, err
	}

	totalDuration := time.Since(start)
	logger.Debug().
		Int("json_size_bytes", len(data)).
		Dur("serialize_duration", serializeDuration).
		Dur("total_generation_duration", totalDuration).
		Str("json_preview", string(data[:min(200, len(data))])+"...").
		Msg("OAuth config generation completed successfully")

	return data, nil
}

// GetPluginManifestHandler returns the plugin manifest handler
func (s *Service) GetPluginManifestHandler() http.HandlerFunc {
	return s.handlePluginManifest
}

// GetOAuthConfigHandler returns the OAuth config handler
func (s *Service) GetOAuthConfigHandler() http.HandlerFunc {
	return s.handleOAuthConfig
}

// isChatGPTUserAgent detects if the request is coming from ChatGPT
func isChatGPTUserAgent(userAgent string) bool {
	userAgent = strings.ToLower(userAgent)
	chatGPTIndicators := []string{
		"chatgpt",
		"openai",
		"gpt-",
		"mozilla/5.0 applewebkit/537.36 (khtml, like gecko) chatgpt",
	}

	for _, indicator := range chatGPTIndicators {
		if strings.Contains(userAgent, indicator) {
			return true
		}
	}
	return false
}

// logRequestHeaders logs all request headers for debugging
func logRequestHeaders(logger zerolog.Logger, headers http.Header) {
	headerMap := make(map[string]string)
	for key, values := range headers {
		if len(values) > 0 {
			headerMap[key] = values[0]
		}
	}
	logger.Debug().
		Interface("all_headers", headerMap).
		Msg("request headers received")
}

// handlePluginManifest serves the ChatGPT plugin manifest
func (s *Service) handlePluginManifest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := fmt.Sprintf("manifest-%d", time.Now().UnixNano())
	userAgent := r.Header.Get("User-Agent")
	isChatGPT := isChatGPTUserAgent(userAgent)

	logger := log.With().
		Str("component", "discovery").
		Str("operation", "serve_manifest").
		Str("request_id", requestID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.RawQuery).
		Str("user_agent", userAgent).
		Str("remote_addr", r.RemoteAddr).
		Str("host", r.Host).
		Bool("is_chatgpt", isChatGPT).
		Str("referer", r.Header.Get("Referer")).
		Str("accept", r.Header.Get("Accept")).
		Str("accept_encoding", r.Header.Get("Accept-Encoding")).
		Str("accept_language", r.Header.Get("Accept-Language")).
		Logger()

	logger.Info().Msg("received .well-known plugin manifest request")

	// Log all request headers in debug mode
	logRequestHeaders(logger, r.Header)

	if r.Method != http.MethodGet {
		logger.Warn().
			Str("allowed_method", "GET").
			Int("status_code", http.StatusMethodNotAllowed).
			Msg("method not allowed for plugin manifest endpoint")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handle CORS preflight requests
	if r.Method == http.MethodOptions {
		logger.Debug().Msg("handling CORS preflight request for plugin manifest")
		s.setCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		logger.Debug().
			Dur("request_duration", time.Since(start)).
			Msg("CORS preflight response sent")
		return
	}

	// Generate manifest with timing
	logger.Debug().Msg("generating plugin manifest for request")
	manifest, err := s.getPluginManifestWithHost(s.getHostFromRequest(r))
	if err != nil {
		logger.Error().
			Err(err).
			Dur("request_duration", time.Since(start)).
			Msg("failed to generate plugin manifest for request")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set response headers
	logger.Debug().Msg("setting response headers for plugin manifest")
	s.setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes cache
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Log response headers being set
	responseHeaders := make(map[string]string)
	for key, values := range w.Header() {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}
	logger.Debug().
		Interface("response_headers", responseHeaders).
		Msg("response headers set for plugin manifest")

	// Write response
	bytesWritten, err := w.Write(manifest)
	requestDuration := time.Since(start)

	if err != nil {
		logger.Error().
			Err(err).
			Int("bytes_written", bytesWritten).
			Dur("request_duration", requestDuration).
			Msg("failed to write plugin manifest response")
		return
	}

	logger.Info().
		Int("response_size_bytes", bytesWritten).
		Int("manifest_size_bytes", len(manifest)).
		Dur("request_duration", requestDuration).
		Int("status_code", http.StatusOK).
		Bool("chatgpt_request", isChatGPT).
		Msg("plugin manifest served successfully")

	if isChatGPT {
		logger.Info().Msg("ChatGPT plugin manifest request completed - connector should now be discoverable")
	}
}

// handleOAuthConfig serves the OAuth authorization server configuration
func (s *Service) handleOAuthConfig(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := fmt.Sprintf("oauth-%d", time.Now().UnixNano())
	userAgent := r.Header.Get("User-Agent")
	isChatGPT := isChatGPTUserAgent(userAgent)

	logger := log.With().
		Str("component", "discovery").
		Str("operation", "serve_oauth_config").
		Str("request_id", requestID).
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("query", r.URL.RawQuery).
		Str("user_agent", userAgent).
		Str("remote_addr", r.RemoteAddr).
		Str("host", r.Host).
		Bool("is_chatgpt", isChatGPT).
		Str("referer", r.Header.Get("Referer")).
		Str("accept", r.Header.Get("Accept")).
		Str("accept_encoding", r.Header.Get("Accept-Encoding")).
		Str("accept_language", r.Header.Get("Accept-Language")).
		Logger()

	logger.Info().Msg("received .well-known OAuth authorization server config request")

	// Log all request headers in debug mode
	logRequestHeaders(logger, r.Header)

	if r.Method != http.MethodGet {
		logger.Warn().
			Str("allowed_method", "GET").
			Int("status_code", http.StatusMethodNotAllowed).
			Msg("method not allowed for OAuth config endpoint")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handle CORS preflight requests
	if r.Method == http.MethodOptions {
		logger.Debug().Msg("handling CORS preflight request for OAuth config")
		s.setCORSHeaders(w)
		w.WriteHeader(http.StatusOK)
		logger.Debug().
			Dur("request_duration", time.Since(start)).
			Msg("CORS preflight response sent")
		return
	}

	// Generate OAuth config with timing
	logger.Debug().Msg("generating OAuth config for request")
	config, err := s.GetOAuthConfig()
	if err != nil {
		logger.Error().
			Err(err).
			Dur("request_duration", time.Since(start)).
			Msg("failed to generate OAuth config for request")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set response headers
	logger.Debug().Msg("setting response headers for OAuth config")
	s.setCORSHeaders(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes cache
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Log response headers being set
	responseHeaders := make(map[string]string)
	for key, values := range w.Header() {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}
	logger.Debug().
		Interface("response_headers", responseHeaders).
		Msg("response headers set for OAuth config")

	// Write response
	bytesWritten, err := w.Write(config)
	requestDuration := time.Since(start)

	if err != nil {
		logger.Error().
			Err(err).
			Int("bytes_written", bytesWritten).
			Dur("request_duration", requestDuration).
			Msg("failed to write OAuth config response")
		return
	}

	logger.Info().
		Int("response_size_bytes", bytesWritten).
		Int("config_size_bytes", len(config)).
		Dur("request_duration", requestDuration).
		Int("status_code", http.StatusOK).
		Bool("chatgpt_request", isChatGPT).
		Msg("OAuth config served successfully")

	if isChatGPT {
		logger.Info().Msg("ChatGPT OAuth config request completed - OAuth flow should be available")
	}
}

// setCORSHeaders sets CORS headers for .well-known endpoints
func (s *Service) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetBaseURL returns the base URL for the server
func (s *Service) GetBaseURL() string {
	var baseURL string
	if s.config.Host == "0.0.0.0" {
		baseURL = fmt.Sprintf("http://localhost:%d", s.config.Port)
	} else {
		baseURL = fmt.Sprintf("http://%s:%d", s.config.Host, s.config.Port)
	}

	log.Debug().
		Str("component", "discovery").
		Str("configured_host", s.config.Host).
		Int("configured_port", s.config.Port).
		Str("base_url", baseURL).
		Msg("generated base URL for discovery service")

	return baseURL
}

// GetManifestURL returns the full URL to the plugin manifest
func (s *Service) GetManifestURL() string {
	manifestURL := s.GetBaseURL() + "/.well-known/ai-plugin.json"

	log.Debug().
		Str("component", "discovery").
		Str("manifest_url", manifestURL).
		Str("endpoint_path", "/.well-known/ai-plugin.json").
		Msg("generated plugin manifest URL")

	return manifestURL
}

// GetOAuthConfigURL returns the full URL to the OAuth config
func (s *Service) GetOAuthConfigURL() string {
	oauthURL := s.GetBaseURL() + "/.well-known/oauth-authorization-server"

	log.Debug().
		Str("component", "discovery").
		Str("oauth_config_url", oauthURL).
		Str("endpoint_path", "/.well-known/oauth-authorization-server").
		Msg("generated OAuth config URL")

	return oauthURL
}

// LogDiscoveryInfo logs comprehensive discovery service information
func (s *Service) LogDiscoveryInfo() {
	log.Info().
		Str("component", "discovery").
		Str("base_url", s.GetBaseURL()).
		Str("manifest_url", s.GetManifestURL()).
		Str("oauth_config_url", s.GetOAuthConfigURL()).
		Interface("config", map[string]interface{}{
			"host": s.config.Host,
			"port": s.config.Port,
		}).
		Msg("discovery service URLs configured - ready for ChatGPT integration")
}

// getHostFromRequest extracts the host from the HTTP request
func (s *Service) getHostFromRequest(r *http.Request) string {
	// Check for X-Forwarded-Host header (used by Tailscale Funnel)
	if forwardedHost := r.Header.Get("X-Forwarded-Host"); forwardedHost != "" {
		return forwardedHost
	}
	
	// Fallback to request Host header
	return r.Host
}

// getSSEURL builds the complete SSE endpoint URL
func (s *Service) getSSEURL(host string) string {
	if host == "" {
		// Fallback to config-based URL
		return fmt.Sprintf("http://%s:%d/sse", s.config.Host, s.config.Port)
	}
	return fmt.Sprintf("https://%s/sse", host)
}
