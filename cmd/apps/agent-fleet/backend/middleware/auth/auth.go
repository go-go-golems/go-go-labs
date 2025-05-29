package auth

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/models"
)

// For simplicity, we'll use a hardcoded token. In production, this should be
// a proper JWT validation or database lookup
const validToken = "fleet-agent-token-123"

// BearerTokenMiddleware validates Bearer token authentication
func BearerTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "MISSING_AUTH_TOKEN", "Authorization header is required")
			return
		}

		// Check Bearer prefix
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeErrorResponse(w, http.StatusUnauthorized, "INVALID_AUTH_FORMAT", "Authorization header must start with 'Bearer '")
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			writeErrorResponse(w, http.StatusUnauthorized, "EMPTY_TOKEN", "Bearer token cannot be empty")
			return
		}

		// Validate token (in production, this would be proper validation)
		if token != validToken {
			writeErrorResponse(w, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid bearer token")
			return
		}

		log.Debug().Str("token", token).Msg("Valid bearer token")

		// Token is valid, continue to next handler
		next.ServeHTTP(w, r)
	})
}

// writeErrorResponse writes a standardized error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResp := models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
		},
	}

	// Simple JSON encoding to avoid import cycles
	json := `{"error":{"code":"` + errorResp.Error.Code + `","message":"` + errorResp.Error.Message + `"}}`
	w.Write([]byte(json))
}
