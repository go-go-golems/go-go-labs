package discovery

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
	"github.com/rs/zerolog/log"
)

// Service implements the DiscoveryService interface
type Service struct {
	config *types.Config
}

// NewService creates a new discovery service
func NewService(config *types.Config) *Service {
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
}

// GetPluginManifest returns the ChatGPT plugin manifest
func (s *Service) GetPluginManifest() ([]byte, error) {
	log.Debug().Msg("generating plugin manifest")

	manifest := PluginManifest{
		Name:        "GitHub MCP Connector",
		Description: "Secure MCP server with GitHub OAuth authentication for personal use",
		Version:     "1.0.0",
		Auth: AuthConfig{
			Type:             "oauth",
			AuthorizationURL: "https://github.com/login/oauth/authorize",
			TokenURL:         "https://github.com/login/oauth/access_token",
			Scopes:           []string{"read:user"},
		},
		API: APIConfig{
			Type: "mcp",
			URL:  "/sse",
		},
		Contact: ContactInfo{
			Email: "admin@example.com",
		},
		Legal: LegalInfo{
			PrivacyURL: "https://example.com/privacy",
			TermsURL:   "https://example.com/terms",
		},
	}

	return json.MarshalIndent(manifest, "", "  ")
}

// GetOAuthConfig returns the OAuth authorization server configuration
func (s *Service) GetOAuthConfig() ([]byte, error) {
	log.Debug().Msg("generating oauth authorization server config")

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
			"client_secret_basic",
		},
	}

	return json.MarshalIndent(config, "", "  ")
}

// GetPluginManifestHandler returns the plugin manifest handler
func (s *Service) GetPluginManifestHandler() http.HandlerFunc {
	return s.handlePluginManifest
}

// GetOAuthConfigHandler returns the OAuth config handler
func (s *Service) GetOAuthConfigHandler() http.HandlerFunc {
	return s.handleOAuthConfig
}

// handlePluginManifest serves the ChatGPT plugin manifest
func (s *Service) handlePluginManifest(w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("user_agent", r.Header.Get("User-Agent")).
		Str("remote_addr", r.RemoteAddr).
		Msg("serving plugin manifest")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	manifest, err := s.GetPluginManifest()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate plugin manifest")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if _, err := w.Write(manifest); err != nil {
		log.Error().Err(err).Msg("failed to write plugin manifest response")
	}

	log.Debug().Msg("plugin manifest served successfully")
}

// handleOAuthConfig serves the OAuth authorization server configuration
func (s *Service) handleOAuthConfig(w http.ResponseWriter, r *http.Request) {
	log.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("user_agent", r.Header.Get("User-Agent")).
		Str("remote_addr", r.RemoteAddr).
		Msg("serving oauth config")

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config, err := s.GetOAuthConfig()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate oauth config")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if _, err := w.Write(config); err != nil {
		log.Error().Err(err).Msg("failed to write oauth config response")
	}

	log.Debug().Msg("oauth config served successfully")
}

// GetBaseURL returns the base URL for the server
func (s *Service) GetBaseURL() string {
	if s.config.Host == "0.0.0.0" {
		return fmt.Sprintf("http://localhost:%d", s.config.Port)
	}
	return fmt.Sprintf("http://%s:%d", s.config.Host, s.config.Port)
}

// GetManifestURL returns the full URL to the plugin manifest
func (s *Service) GetManifestURL() string {
	return s.GetBaseURL() + "/.well-known/ai-plugin.json"
}

// GetOAuthConfigURL returns the full URL to the OAuth config
func (s *Service) GetOAuthConfigURL() string {
	return s.GetBaseURL() + "/.well-known/oauth-authorization-server"
}
