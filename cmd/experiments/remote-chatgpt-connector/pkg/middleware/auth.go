package middleware

import (
	"context"
	"net/http"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/auth"
	"github.com/ory/fosite/handler/openid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AuthMiddleware provides bearer token validation for protected endpoints
type AuthMiddleware struct {
	oidcProvider *auth.OIDCProvider
	logger       zerolog.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(oidcProvider *auth.OIDCProvider) *AuthMiddleware {
	return &AuthMiddleware{
		oidcProvider: oidcProvider,
		logger:       log.With().Str("component", "auth-middleware").Logger(),
	}
}

// RequireAuth is a middleware that validates bearer tokens
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := m.logger.With().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Logger()

		// Extract bearer token
		authHeader := r.Header.Get("Authorization")
		token := auth.ExtractBearerToken(authHeader)

		if token == "" {
			logger.Warn().Msg("missing bearer token")
			w.Header().Set("WWW-Authenticate", `Bearer realm="MCP", error="invalid_request"`)
			http.Error(w, "Missing access token", http.StatusUnauthorized)
			return
		}

		logger.Debug().Str("token_prefix", token[:min(10, len(token))]).Msg("validating access token")

		// Validate token using OIDC provider
		ctx := r.Context()
		requester, err := m.oidcProvider.IntrospectToken(ctx, token)
		if err != nil {
			logger.Warn().Err(err).Msg("invalid or expired token")
			w.Header().Set("WWW-Authenticate", `Bearer realm="MCP", error="invalid_token"`)
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Get session for user info
		session := requester.GetSession()
		if session == nil {
			logger.Error().Msg("token validation succeeded but no session found")
			http.Error(w, "Invalid token session", http.StatusUnauthorized)
			return
		}

		// Extract subject/username from session
		var username string
		if oidcSession, ok := session.(*openid.DefaultSession); ok {
			username = oidcSession.Username
		}
		if username == "" {
			username = "unknown"
		}

		logger.Info().
			Str("username", username).
			Str("client_id", requester.GetClient().GetID()).
			Strs("scopes", requester.GetGrantedScopes()).
			Msg("access token validated successfully")

		// Add user context to request
		ctx = context.WithValue(ctx, "user", username)
		ctx = context.WithValue(ctx, "client_id", requester.GetClient().GetID())
		ctx = context.WithValue(ctx, "scopes", requester.GetGrantedScopes())

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireScope is a middleware that checks for specific OAuth scopes
func (m *AuthMiddleware) RequireScope(scope string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return m.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
			logger := m.logger.With().
				Str("required_scope", scope).
				Str("path", r.URL.Path).
				Logger()

			// Get scopes from context (set by RequireAuth)
			scopesValue := r.Context().Value("scopes")
			if scopesValue == nil {
				logger.Error().Msg("scopes not found in context")
				http.Error(w, "Invalid authorization", http.StatusForbidden)
				return
			}

			scopes, ok := scopesValue.([]string)
			if !ok {
				logger.Error().Msg("scopes context value is not string slice")
				http.Error(w, "Invalid authorization", http.StatusForbidden)
				return
			}

			// Check if required scope is present
			hasScope := false
			for _, s := range scopes {
				if s == scope {
					hasScope = true
					break
				}
			}

			if !hasScope {
				logger.Warn().
					Strs("granted_scopes", scopes).
					Msg("required scope not granted")
				w.Header().Set("WWW-Authenticate", `Bearer realm="MCP", error="insufficient_scope"`)
				http.Error(w, "Insufficient scope", http.StatusForbidden)
				return
			}

			logger.Debug().Msg("scope requirement satisfied")
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts the authenticated username from request context
func GetUserFromContext(ctx context.Context) string {
	if user := ctx.Value("user"); user != nil {
		if username, ok := user.(string); ok {
			return username
		}
	}
	return ""
}

// GetClientIDFromContext extracts the client ID from request context
func GetClientIDFromContext(ctx context.Context) string {
	if clientID := ctx.Value("client_id"); clientID != nil {
		if id, ok := clientID.(string); ok {
			return id
		}
	}
	return ""
}

// GetScopesFromContext extracts the granted scopes from request context
func GetScopesFromContext(ctx context.Context) []string {
	if scopes := ctx.Value("scopes"); scopes != nil {
		if scopeSlice, ok := scopes.([]string); ok {
			return scopeSlice
		}
	}
	return nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
