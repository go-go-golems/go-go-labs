package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

// AuthorizationProxy handles authorization requests for dynamic clients
type AuthorizationProxy struct {
	staticClientID     string
	staticClientSecret string
	authorizationURL   string
	// Store state mappings: dynamic_state -> {original_state, dynamic_client_id, redirect_uri}
	stateMappings map[string]StateMapping
}

// StateMapping maps authorization state between dynamic clients and static Auth0 client
type StateMapping struct {
	OriginalState       string
	DynamicClientID     string
	OriginalRedirectURI string
	CreatedAt           time.Time
}

// NewAuthorizationProxy creates a new authorization proxy
func NewAuthorizationProxy(staticClientID, staticClientSecret, authorizationURL string) *AuthorizationProxy {
	return &AuthorizationProxy{
		staticClientID:     staticClientID,
		staticClientSecret: staticClientSecret,
		authorizationURL:   authorizationURL,
		stateMappings:      make(map[string]StateMapping),
	}
}

// HandleAuthorizationRequest proxies authorization requests from dynamic clients to Auth0
func (p *AuthorizationProxy) HandleAuthorizationRequest(w http.ResponseWriter, r *http.Request) {
	logger := log.With().
		Str("component", "auth_proxy").
		Str("operation", "authorization").
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Logger()

	logger.Info().Msg("received authorization request for dynamic client")

	// Parse query parameters
	query := r.URL.Query()
	dynamicClientID := query.Get("client_id")
	originalState := query.Get("state")
	redirectURI := query.Get("redirect_uri")
	scope := query.Get("scope")
	responseType := query.Get("response_type")
	codeChallenge := query.Get("code_challenge")
	codeChallengeMethod := query.Get("code_challenge_method")

	logger.Debug().
		Str("dynamic_client_id", dynamicClientID).
		Str("original_state", originalState).
		Str("redirect_uri", redirectURI).
		Str("scope", scope).
		Msg("parsed authorization request parameters")

	// Validate the dynamic client ID exists (this should be checked against our dynamic client store)
	if dynamicClientID == "" {
		logger.Warn().Msg("missing client_id in authorization request")
		http.Error(w, "invalid_request: missing client_id", http.StatusBadRequest)
		return
	}

	if !p.isDynamicClient(dynamicClientID) {
		logger.Warn().
			Str("dynamic_client_id", dynamicClientID).
			Msg("unknown dynamic client_id")
		http.Error(w, "invalid_client: unknown client_id", http.StatusBadRequest)
		return
	}

	// Generate a new state for the Auth0 request
	proxyState := p.generateState()

	// Store the state mapping
	p.stateMappings[proxyState] = StateMapping{
		OriginalState:       originalState,
		DynamicClientID:     dynamicClientID,
		OriginalRedirectURI: redirectURI,
		CreatedAt:           time.Now(),
	}

	logger.Debug().
		Str("proxy_state", proxyState).
		Str("original_state", originalState).
		Msg("created state mapping for authorization proxy")

	// Build Auth0 authorization URL with static client_id
	auth0URL, err := url.Parse(p.authorizationURL)
	if err != nil {
		logger.Error().Err(err).Msg("failed to parse authorization URL")
		http.Error(w, "server_error", http.StatusInternalServerError)
		return
	}

	// Set query parameters for Auth0
	params := url.Values{}
	params.Set("response_type", responseType)
	params.Set("client_id", p.staticClientID)
	params.Set("redirect_uri", "https://f.beagle-duck.ts.net/oauth2/callback") // Our callback endpoint
	params.Set("state", proxyState)
	params.Set("scope", scope)
	if codeChallenge != "" {
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", codeChallengeMethod)
	}

	auth0URL.RawQuery = params.Encode()
	finalURL := auth0URL.String()

	logger.Info().
		Str("auth0_redirect_url", finalURL).
		Str("static_client_id", p.staticClientID).
		Msg("redirecting to Auth0 with static client credentials")

	// Redirect to Auth0
	http.Redirect(w, r, finalURL, http.StatusFound)
}

// generateState creates a random state parameter
func (p *AuthorizationProxy) generateState() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "proxy_" + hex.EncodeToString(bytes)
}

// isDynamicClient checks if a client_id is a dynamic client (placeholder for now)
func (p *AuthorizationProxy) isDynamicClient(clientID string) bool {
	// For now, just check if it starts with "mcp_"
	// In a real implementation, this should check against the dynamic client store
	return len(clientID) > 4 && clientID[:4] == "mcp_"
}

// GetStateMapping retrieves the original request details for a proxy state
func (p *AuthorizationProxy) GetStateMapping(proxyState string) (StateMapping, bool) {
	mapping, exists := p.stateMappings[proxyState]
	return mapping, exists
}

// CleanupExpiredStates removes old state mappings (should be called periodically)
func (p *AuthorizationProxy) CleanupExpiredStates() {
	cutoff := time.Now().Add(-10 * time.Minute) // 10 minute expiry
	for state, mapping := range p.stateMappings {
		if mapping.CreatedAt.Before(cutoff) {
			delete(p.stateMappings, state)
		}
	}
}
