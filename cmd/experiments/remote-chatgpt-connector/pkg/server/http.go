package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
)

// HTTPServer provides the HTTP server component with SSE support for MCP
type HTTPServer struct {
	config        *types.Config
	router        *mux.Router
	server        *http.Server
	authValidator types.AuthValidator
	mcpServer     types.MCPServer
	transport     types.Transport
	discovery     types.DiscoveryService
}

// NewHTTPServer creates a new HTTP server with gorilla/mux routing
func NewHTTPServer(config *types.Config) *HTTPServer {
	router := mux.NewRouter()

	return &HTTPServer{
		config: config,
		router: router,
		server: &http.Server{
			Addr:              fmt.Sprintf("%s:%d", config.Host, config.Port),
			Handler:           router,
			ReadHeaderTimeout: 30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
	}
}

// SetAuthValidator configures the OAuth validation middleware
func (h *HTTPServer) SetAuthValidator(validator types.AuthValidator) {
	h.authValidator = validator
	log.Debug().Msg("Auth validator configured")
}

// SetMCPServer configures the MCP server instance
func (h *HTTPServer) SetMCPServer(mcpServer types.MCPServer) {
	h.mcpServer = mcpServer
	log.Debug().Msg("MCP server configured")
}

// SetTransport configures the SSE transport layer
func (h *HTTPServer) SetTransport(transport types.Transport) {
	h.transport = transport
	// Wire up transport dependencies
	if h.authValidator != nil {
		transport.SetAuthValidator(h.authValidator)
	}
	if h.mcpServer != nil {
		transport.SetMCPServer(h.mcpServer)
	}
	log.Debug().Msg("Transport configured")
}

// SetDiscoveryService configures the well-known endpoints service
func (h *HTTPServer) SetDiscoveryService(discovery types.DiscoveryService) {
	h.discovery = discovery
	log.Debug().Msg("Discovery service configured")
}

// setupRoutes configures all HTTP routes and middleware
func (h *HTTPServer) setupRoutes() {
	// Health check endpoint - no auth required
	h.router.HandleFunc("/health", h.handleHealth).Methods("GET")
	h.router.HandleFunc("/health/ready", h.handleReady).Methods("GET")
	h.router.HandleFunc("/health/live", h.handleLive).Methods("GET")

	// Well-known discovery endpoints - no auth required
	if h.discovery != nil {
		// Register discovery routes (/.well-known/*)
		h.router.HandleFunc("/.well-known/ai-plugin.json", h.discovery.GetPluginManifestHandler()).Methods("GET")
		h.router.HandleFunc("/.well-known/oauth-authorization-server", h.discovery.GetOAuthConfigHandler()).Methods("GET")
	}

	// MCP SSE endpoint - requires OAuth authentication
	if h.transport != nil {
		// Apply auth middleware to SSE endpoint
		sseRouter := h.router.PathPrefix("/sse").Subrouter()
		sseRouter.Use(h.authMiddleware)
		sseRouter.Use(h.corsMiddleware)
		sseRouter.Use(h.loggingMiddleware)
		sseRouter.PathPrefix("/").HandlerFunc(h.transport.ServeHTTP)
		log.Debug().Msg("SSE endpoint configured with auth middleware")
	}

	// Add global middleware
	h.router.Use(h.securityHeadersMiddleware)
	h.router.Use(h.recoveryMiddleware)
}

// authMiddleware validates OAuth bearer tokens
func (h *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.authValidator == nil {
			log.Error().Msg("Auth validator not configured")
			http.Error(w, "Server misconfigured", http.StatusInternalServerError)
			return
		}

		// Extract bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			log.Warn().Str("path", r.URL.Path).Msg("Missing authorization header")
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn().Str("auth_header", authHeader).Msg("Invalid authorization format")
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			log.Warn().Msg("Empty bearer token")
			http.Error(w, "Empty bearer token", http.StatusUnauthorized)
			return
		}

		// Validate token
		userInfo, err := h.authValidator.ValidateToken(r.Context(), token)
		if err != nil {
			log.Warn().Err(err).Str("token_prefix", token[:min(8, len(token))]).Msg("Token validation failed")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user", userInfo)
		log.Debug().Str("user_login", userInfo.Login).Str("user_id", userInfo.ID).Msg("User authenticated")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// corsMiddleware adds CORS headers for SSE connections
func (h *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for SSE
		w.Header().Set("Access-Control-Allow-Origin", "*") // Configure appropriately for production
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// securityHeadersMiddleware adds security headers
func (h *HTTPServer) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Only add HSTS in production with HTTPS
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs HTTP requests
func (h *HTTPServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status
		wrapped := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Log the request
		duration := time.Since(start)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", wrapped.statusCode).
			Dur("duration", duration).
			Str("user_agent", r.UserAgent()).
			Str("remote_addr", r.RemoteAddr).
			Msg("HTTP request")
	})
}

// recoveryMiddleware recovers from panics
func (h *HTTPServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("panic", err).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Msg("Panic recovered in HTTP handler")

				if !headersSent(w) {
					http.Error(w, "Internal server error", http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Health check handlers
func (h *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"ok","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

func (h *HTTPServer) handleReady(w http.ResponseWriter, r *http.Request) {
	// Check if all components are ready
	ready := true
	components := make(map[string]bool)

	components["auth"] = h.authValidator != nil
	components["mcp"] = h.mcpServer != nil
	components["transport"] = h.transport != nil
	components["discovery"] = h.discovery != nil

	for _, status := range components {
		if !status {
			ready = false
			break
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	fmt.Fprintf(w, `{"ready":%t,"components":%v,"timestamp":"%s"}`,
		ready, components, time.Now().UTC().Format(time.RFC3339))
}

func (h *HTTPServer) handleLive(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"alive":true,"timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))
}

// Start begins serving HTTP requests
func (h *HTTPServer) Start(ctx context.Context) error {
	// Setup routes with all configured components
	h.setupRoutes()

	log.Info().
		Str("addr", h.server.Addr).
		Msg("Starting HTTP server")

	// Start server in goroutine
	go func() {
		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("HTTP server failed")
		}
	}()

	return nil
}

// Stop gracefully shuts down the HTTP server
func (h *HTTPServer) Stop(ctx context.Context) error {
	log.Info().Msg("Shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	return h.server.Shutdown(shutdownCtx)
}

// Helper types and functions

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func headersSent(w http.ResponseWriter) bool {
	// This is a simple heuristic - in practice you might need a wrapper
	// to track if headers were sent
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
