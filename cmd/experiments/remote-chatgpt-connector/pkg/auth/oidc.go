package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"


	mystorage "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/storage"
)

// OIDCProvider manages OAuth2/OIDC flows using Fosite
type OIDCProvider struct {
	provider fosite.OAuth2Provider
	store    *storage.MemoryStore
	mystore  *mystorage.MemoryStore
	config   *OIDCConfig
	logger   zerolog.Logger
}

// OIDCConfig holds configuration for the OIDC provider
type OIDCConfig struct {
	Issuer         string
	Port           int
	SigningKey     *rsa.PrivateKey
	DefaultUser    string
	DefaultPassword string
}

// NewOIDCProvider creates a new OIDC provider with Fosite
func NewOIDCProvider(config *OIDCConfig) (*OIDCProvider, error) {
	logger := log.With().Str("component", "oidc-provider").Logger()

	// Create Fosite's built-in memory storage
	store := storage.NewMemoryStore()
	
	// Create our custom storage for user management and client registration
	mystore := mystorage.NewMemoryStore()

	// Generate RSA key if not provided
	if config.SigningKey == nil {
		logger.Info().Msg("generating RSA signing key")
		key, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return nil, errors.Wrap(err, "failed to generate RSA key")
		}
		config.SigningKey = key
	}

	// Create Fosite configuration
	fositeConfig := &fosite.Config{
		GlobalSecret:                   []byte("global-secret-for-hmac-tokens-32bytes!!"),
		AccessTokenLifespan:            time.Hour,
		RefreshTokenLifespan:           time.Hour * 24,
		AuthorizeCodeLifespan:          time.Minute * 10,
		IDTokenLifespan:                time.Hour,
		IDTokenIssuer:                  config.Issuer,
		HashCost:                       12, // bcrypt cost
		SendDebugMessagesToClients:     true, // Enable for development
		EnforcePKCE:                    true, // Require PKCE
		EnforcePKCEForPublicClients:    true,
	}

	// Create HMAC strategy for access/refresh tokens
	hmacStrategy := compose.NewOAuth2HMACStrategy(fositeConfig)

	// Create OpenID Connect strategy
	oidcStrategy := compose.NewOpenIDConnectStrategy(func(ctx context.Context) (interface{}, error) {
		return config.SigningKey, nil
	}, fositeConfig)

	// Compose the provider with required handlers
	provider := compose.Compose(
		fositeConfig,
		store,
		&compose.CommonStrategy{
			CoreStrategy:               hmacStrategy,
			OpenIDConnectTokenStrategy: oidcStrategy,
		},
		// Core OAuth2 flows
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2TokenIntrospectionFactory, 
		compose.OAuth2TokenRevocationFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OAuth2PKCEFactory,
		// OIDC support
		compose.OpenIDConnectExplicitFactory,
		compose.OpenIDConnectRefreshFactory,
	)

	return &OIDCProvider{
		provider: provider,
		store:    store,
		mystore:  mystore,
		config:   config,
		logger:   logger,
	}, nil
}

// DynamicClientRegistration handles POST /register for client registration
func (p *OIDCProvider) DynamicClientRegistration(w http.ResponseWriter, r *http.Request) {
	logger := p.logger.With().Str("handler", "register").Logger()
	
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var regReq struct {
		RedirectURIs               []string `json:"redirect_uris"`
		GrantTypes                 []string `json:"grant_types,omitempty"`
		ResponseTypes              []string `json:"response_types,omitempty"`
		ClientName                 string   `json:"client_name,omitempty"`
		TokenEndpointAuthMethod    string   `json:"token_endpoint_auth_method,omitempty"`
		Scope                      string   `json:"scope,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&regReq); err != nil {
		logger.Error().Err(err).Msg("failed to decode registration request")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	logger.Info().
		Strs("redirect_uris", regReq.RedirectURIs).
		Strs("grant_types", regReq.GrantTypes).
		Str("client_name", regReq.ClientName).
		Msg("processing client registration")

	// Validate required fields
	if len(regReq.RedirectURIs) == 0 {
		http.Error(w, "redirect_uris is required", http.StatusBadRequest)
		return
	}

	// Validate redirect URIs
	for _, uri := range regReq.RedirectURIs {
		if _, err := url.Parse(uri); err != nil {
			http.Error(w, fmt.Sprintf("invalid redirect_uri: %s", uri), http.StatusBadRequest)
			return
		}
	}

	// Set defaults
	if len(regReq.GrantTypes) == 0 {
		regReq.GrantTypes = []string{"authorization_code", "refresh_token"}
	}
	if len(regReq.ResponseTypes) == 0 {
		regReq.ResponseTypes = []string{"code"}
	}

	// Determine if public client
	publicClient := regReq.TokenEndpointAuthMethod == "none" || regReq.TokenEndpointAuthMethod == ""
	
	// Generate client credentials
	clientID := p.mystore.GenerateClientID()
	var clientSecret string
	if !publicClient {
		clientSecret = p.mystore.GenerateClientSecret()
	}

	// Parse scopes
	var scopes []string
	if regReq.Scope != "" {
		scopes = strings.Fields(regReq.Scope)
	}

	// Create client
	client := &fosite.DefaultClient{
		ID:            clientID,
		RedirectURIs:  regReq.RedirectURIs,
		ResponseTypes: regReq.ResponseTypes,
		GrantTypes:    regReq.GrantTypes,
		Scopes:        scopes,
		Public:        publicClient,
	}

	if !publicClient {
		// Store the secret directly (Fosite will hash it)
		client.Secret = []byte(clientSecret)
	}

	// Store the client in fosite storage
	p.store.Clients[clientID] = client
	
	// Also store in our custom storage for tracking
	if err := p.mystore.CreateClient(client); err != nil {
		logger.Error().Err(err).Msg("failed to store client in custom storage")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Prepare response
	resp := map[string]interface{}{
		"client_id":         clientID,
		"client_id_issued_at": time.Now().Unix(),
		"redirect_uris":     regReq.RedirectURIs,
		"grant_types":       regReq.GrantTypes,
		"response_types":    regReq.ResponseTypes,
	}

	if !publicClient {
		resp["client_secret"] = clientSecret
		resp["client_secret_expires_at"] = 0 // Never expires
		resp["token_endpoint_auth_method"] = "client_secret_post"
	} else {
		resp["token_endpoint_auth_method"] = "none"
	}

	if regReq.ClientName != "" {
		resp["client_name"] = regReq.ClientName
	}
	if regReq.Scope != "" {
		resp["scope"] = regReq.Scope
	}

	logger.Info().
		Str("client_id", clientID).
		Bool("public_client", publicClient).
		Msg("client registered successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// AuthorizeHandler handles GET/POST /authorize
func (p *OIDCProvider) AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
	logger := p.logger.With().Str("handler", "authorize").Logger()
	ctx := r.Context()

	// Parse the authorize request
	authReq, err := p.provider.NewAuthorizeRequest(ctx, r)
	if err != nil {
		logger.Error().Err(err).Msg("invalid authorize request")
		p.provider.WriteAuthorizeError(ctx, w, authReq, err)
		return
	}

	logger.Info().
		Str("client_id", authReq.GetClient().GetID()).
		Strs("scopes", authReq.GetRequestedScopes()).
		Str("redirect_uri", authReq.GetRedirectURI().String()).
		Msg("processing authorize request")

	// Handle GET vs POST
	if r.Method == http.MethodGet {
		// Show login form
		p.showLoginForm(w, r)
		return
	}

	if r.Method == http.MethodPost {
		// Process login
		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		if username == "" || password == "" {
			logger.Warn().Msg("missing username or password")
			http.Error(w, "Username and password required", http.StatusBadRequest)
			return
		}

		// Authenticate user
		user, err := p.mystore.AuthenticateUser(username, password)
		if err != nil {
			logger.Warn().Str("username", username).Msg("authentication failed")
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		logger.Info().Str("username", username).Msg("user authenticated successfully")

		// Grant all requested scopes (in production, show consent page)
		for _, scope := range authReq.GetRequestedScopes() {
			authReq.GrantScope(scope)
		}

		// Create session
		session := &openid.DefaultSession{
			Claims: &jwt.IDTokenClaims{
				Issuer:    p.config.Issuer,
				Subject:   user.Subject,
				Audience:  []string{authReq.GetClient().GetID()},
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(time.Hour),
				AuthTime:  time.Now(),
			},
			Headers:    &jwt.Headers{},
			ExpiresAt:  map[fosite.TokenType]time.Time{},
			Username:   user.Username,
			Subject:    user.Subject,
		}

		// Create authorize response (generates code)
		response, err := p.provider.NewAuthorizeResponse(ctx, authReq, session)
		if err != nil {
			logger.Error().Err(err).Msg("failed to create authorize response") 
			p.provider.WriteAuthorizeError(ctx, w, authReq, err)
			return
		}

		logger.Info().
			Str("username", username).
			Str("client_id", authReq.GetClient().GetID()).
			Msg("authorization code issued")

		// Redirect back to client
		p.provider.WriteAuthorizeResponse(ctx, w, authReq, response)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// TokenHandler handles POST /token
func (p *OIDCProvider) TokenHandler(w http.ResponseWriter, r *http.Request) {
	logger := p.logger.With().Str("handler", "token").Logger()
	ctx := r.Context()

	logger.Info().
		Str("content_type", r.Header.Get("Content-Type")).
		Str("authorization", r.Header.Get("Authorization")).
		Msg("processing token request")

	// Parse form data
	if err := r.ParseForm(); err != nil {
		logger.Error().Err(err).Msg("failed to parse form")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Create access request
	accessReq, err := p.provider.NewAccessRequest(ctx, r, &openid.DefaultSession{})
	if err != nil {
		logger.Error().Err(err).Msg("invalid access request")
		p.provider.WriteAccessError(ctx, w, accessReq, err)
		return
	}

	logger.Info().
		Strs("grant_types", accessReq.GetGrantTypes()).
		Str("client_id", accessReq.GetClient().GetID()).
		Msg("access request validated")

	// Create access response (issues tokens)
	response, err := p.provider.NewAccessResponse(ctx, accessReq)
	if err != nil {
		logger.Error().Err(err).Msg("failed to create access response")
		p.provider.WriteAccessError(ctx, w, accessReq, err)
		return
	}

	logger.Info().
		Str("client_id", accessReq.GetClient().GetID()).
		Msg("tokens issued successfully")

	// Send token response
	p.provider.WriteAccessResponse(ctx, w, accessReq, response)
}

// IntrospectToken validates an access token and returns the associated request
func (p *OIDCProvider) IntrospectToken(ctx context.Context, token string) (fosite.Requester, error) {
	session := &openid.DefaultSession{}
	tokenUse, requester, err := p.provider.IntrospectToken(ctx, token, fosite.AccessToken, session)
	if err != nil {
		return nil, err
	}
	// Verify it's an access token
	if tokenUse != fosite.AccessToken {
		return nil, fosite.ErrInvalidTokenFormat
	}
	return requester, nil
}

// ExtractBearerToken extracts the bearer token from Authorization header
func ExtractBearerToken(authHeader string) string {
	if authHeader == "" {
		return ""
	}
	
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	
	return parts[1]
}

// showLoginForm displays a simple HTML login form
func (p *OIDCProvider) showLoginForm(w http.ResponseWriter, r *http.Request) {
	// Include all query parameters in form action to maintain state
	queryParams := r.URL.RawQuery
	
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Login - MCP Connector</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 400px; margin: 100px auto; padding: 20px; }
        .form-group { margin-bottom: 15px; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input { width: 100%%; padding: 8px; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
        button { width: 100%%; padding: 10px; background: #007bff; color: white; border: none; border-radius: 4px; cursor: pointer; }
        button:hover { background: #0056b3; }
        .info { background: #f8f9fa; padding: 15px; border-radius: 4px; margin-bottom: 20px; font-size: 14px; }
    </style>
</head>
<body>
    <h2>MCP Connector Login</h2>
    <div class="info">
        <strong>Demo Credentials:</strong><br>
        Username: wesen<br>
        Password: secret
    </div>
    <form method="POST" action="/authorize?%s">
        <div class="form-group">
            <label for="username">Username:</label>
            <input type="text" id="username" name="username" required>
        </div>
        <div class="form-group">
            <label for="password">Password:</label>
            <input type="password" id="password" name="password" required>
        </div>
        <button type="submit">Log In</button>
    </form>
</body>
</html>`, queryParams)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// GetPublicKey returns the public key for JWT verification
func (p *OIDCProvider) GetPublicKey() *rsa.PublicKey {
	return &p.config.SigningKey.PublicKey
}

// GetPublicKeyPEM returns the public key in PEM format for JWKS
func (p *OIDCProvider) GetPublicKeyPEM() (string, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&p.config.SigningKey.PublicKey)
	if err != nil {
		return "", err
	}

	publicKeyPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return string(publicKeyPem), nil
}
