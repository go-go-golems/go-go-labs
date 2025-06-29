package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
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

	log.Debug().
		Str("host", config.Host).
		Int("port", config.Port).
		Dur("read_header_timeout", 30*time.Second).
		Dur("write_timeout", 30*time.Second).
		Dur("idle_timeout", 120*time.Second).
		Msg("Creating HTTP server")

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
	log.Debug().
		Bool("auth_validator_set", validator != nil).
		Msg("Auth validator configured")
}

// SetMCPServer configures the MCP server instance
func (h *HTTPServer) SetMCPServer(mcpServer types.MCPServer) {
	h.mcpServer = mcpServer
	log.Debug().
		Bool("mcp_server_set", mcpServer != nil).
		Msg("MCP server configured")
}

// SetTransport configures the SSE transport layer
func (h *HTTPServer) SetTransport(transport types.Transport) {
	h.transport = transport
	// Wire up transport dependencies
	if h.authValidator != nil {
		transport.SetAuthValidator(h.authValidator)
		log.Debug().Msg("Transport wired with auth validator")
	}
	if h.mcpServer != nil {
		transport.SetMCPServer(h.mcpServer)
		log.Debug().Msg("Transport wired with MCP server")
	}
	log.Debug().
		Bool("transport_set", transport != nil).
		Bool("auth_validator_available", h.authValidator != nil).
		Bool("mcp_server_available", h.mcpServer != nil).
		Msg("Transport configured")
}

// SetDiscoveryService configures the well-known endpoints service
func (h *HTTPServer) SetDiscoveryService(discovery types.DiscoveryService) {
	h.discovery = discovery
	log.Debug().
		Bool("discovery_service_set", discovery != nil).
		Msg("Discovery service configured")
}

// setupRoutes configures all HTTP routes and middleware
func (h *HTTPServer) setupRoutes() {
	setupStart := time.Now()
	log.Debug().Msg("Setting up HTTP routes and middleware")

	// Health check endpoint - no auth required
	h.router.HandleFunc("/health", h.handleHealth).Methods("GET")
	h.router.HandleFunc("/health/ready", h.handleReady).Methods("GET")
	h.router.HandleFunc("/health/live", h.handleLive).Methods("GET")
	log.Debug().
		Strs("health_endpoints", []string{"/health", "/health/ready", "/health/live"}).
		Msg("Health check endpoints registered")

	// Well-known discovery endpoints - no auth required
	if h.discovery != nil {
		// Register discovery routes (/.well-known/*)
		h.router.HandleFunc("/.well-known/ai-plugin.json", h.discovery.GetPluginManifestHandler()).Methods("GET")
		h.router.HandleFunc("/.well-known/oauth-authorization-server", h.discovery.GetOAuthConfigHandler()).Methods("GET")
		log.Debug().
			Strs("discovery_endpoints", []string{"/.well-known/ai-plugin.json", "/.well-known/oauth-authorization-server"}).
			Msg("Discovery endpoints registered")
	} else {
		log.Debug().Msg("Discovery service not configured, skipping well-known endpoints")
	}

	// MCP SSE endpoint - requires OAuth authentication
	if h.transport != nil {
		// Apply auth middleware to SSE endpoint
		sseRouter := h.router.PathPrefix("/sse").Subrouter()
		log.Debug().Msg("Setting up SSE subrouter with middleware chain")

		sseRouter.Use(h.requestIDMiddleware)
		log.Debug().Msg("Added request ID middleware to SSE router")

		sseRouter.Use(h.enhancedLoggingMiddleware)
		log.Debug().Msg("Added enhanced logging middleware to SSE router")

		sseRouter.Use(h.authMiddleware)
		log.Debug().Msg("Added auth middleware to SSE router")

		sseRouter.Use(h.corsMiddleware)
		log.Debug().Msg("Added CORS middleware to SSE router")

		sseRouter.PathPrefix("/").HandlerFunc(h.transport.ServeHTTP)
		log.Debug().
			Str("sse_path_prefix", "/sse").
			Msg("SSE endpoint configured with full middleware chain")
	} else {
		log.Debug().Msg("Transport not configured, skipping SSE endpoints")
	}

	// Add global middleware
	h.router.Use(h.requestIDMiddleware)
	log.Debug().Msg("Added global request ID middleware")

	h.router.Use(h.enhancedLoggingMiddleware)
	log.Debug().Msg("Added global enhanced logging middleware")

	h.router.Use(h.securityHeadersMiddleware)
	log.Debug().Msg("Added global security headers middleware")

	h.router.Use(h.recoveryMiddleware)
	log.Debug().Msg("Added global recovery middleware")

	setupDuration := time.Since(setupStart)
	log.Debug().
		Dur("setup_duration", setupDuration).
		Msg("Route and middleware setup completed")
}

// requestIDMiddleware adds a unique request ID to each request
func (h *HTTPServer) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "requestID", requestID)
		w.Header().Set("X-Request-ID", requestID)

		log.Debug().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Msg("Request ID assigned")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authMiddleware validates OAuth bearer tokens
func (h *HTTPServer) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authStart := time.Now()
		requestID := getRequestID(r.Context())

		logger := log.With().
			Str("request_id", requestID).
			Str("middleware", "auth").
			Logger()

		logger.Debug().
			Str("path", r.URL.Path).
			Str("method", r.Method).
			Str("remote_addr", r.RemoteAddr).
			Msg("Auth middleware processing")

		if h.authValidator == nil {
			logger.Error().Msg("Auth validator not configured")
			http.Error(w, "Server misconfigured", http.StatusInternalServerError)
			return
		}

		// Extract bearer token
		authHeader := r.Header.Get("Authorization")
		logger.Debug().
			Bool("auth_header_present", authHeader != "").
			Msg("Checking authorization header")

		if authHeader == "" {
			logger.Warn().
				Str("path", r.URL.Path).
				Str("user_agent", r.UserAgent()).
				Msg("Missing authorization header")
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn().
				Str("auth_header_format", "invalid").
				Int("auth_header_length", len(authHeader)).
				Msg("Invalid authorization format")
			http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			logger.Warn().Msg("Empty bearer token")
			http.Error(w, "Empty bearer token", http.StatusUnauthorized)
			return
		}

		logger.Debug().
			Int("token_length", len(token)).
			Str("token_prefix", token[:min(8, len(token))]).
			Msg("Validating bearer token")

		// Validate token
		tokenValidationStart := time.Now()
		userInfo, err := h.authValidator.ValidateToken(r.Context(), token)
		tokenValidationDuration := time.Since(tokenValidationStart)

		if err != nil {
			logger.Warn().
				Err(err).
				Str("token_prefix", token[:min(8, len(token))]).
				Dur("validation_duration", tokenValidationDuration).
				Msg("Token validation failed")
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user info to request context
		ctx := context.WithValue(r.Context(), "user", userInfo)
		authDuration := time.Since(authStart)

		logger.Debug().
			Str("user_login", userInfo.Login).
			Str("user_id", userInfo.ID).
			Dur("auth_duration", authDuration).
			Dur("token_validation_duration", tokenValidationDuration).
			Msg("User authenticated successfully")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// corsMiddleware adds CORS headers for SSE connections
func (h *HTTPServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r.Context())

		logger := log.With().
			Str("request_id", requestID).
			Str("middleware", "cors").
			Logger()

		corsHeaders := map[string]string{
			"Access-Control-Allow-Origin":      "*", // Configure appropriately for production
			"Access-Control-Allow-Methods":     "GET, POST, OPTIONS",
			"Access-Control-Allow-Headers":     "Authorization, Content-Type, Accept",
			"Access-Control-Allow-Credentials": "true",
		}

		// Set CORS headers for SSE
		for header, value := range corsHeaders {
			w.Header().Set(header, value)
		}

		logger.Debug().
			Str("origin", r.Header.Get("Origin")).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Interface("cors_headers", corsHeaders).
			Msg("CORS headers applied")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			logger.Debug().
				Str("preflight_origin", r.Header.Get("Origin")).
				Str("preflight_method", r.Header.Get("Access-Control-Request-Method")).
				Str("preflight_headers", r.Header.Get("Access-Control-Request-Headers")).
				Msg("Handling CORS preflight request")
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// securityHeadersMiddleware adds security headers
func (h *HTTPServer) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := getRequestID(r.Context())

		logger := log.With().
			Str("request_id", requestID).
			Str("middleware", "security").
			Logger()

		securityHeaders := map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
			"X-XSS-Protection":       "1; mode=block",
			"Referrer-Policy":        "strict-origin-when-cross-origin",
		}

		for header, value := range securityHeaders {
			w.Header().Set(header, value)
		}

		// Only add HSTS in production with HTTPS
		if r.TLS != nil {
			hstsHeader := "max-age=31536000; includeSubDomains"
			w.Header().Set("Strict-Transport-Security", hstsHeader)
			securityHeaders["Strict-Transport-Security"] = hstsHeader
			logger.Debug().Msg("Added HSTS header for HTTPS connection")
		}

		logger.Debug().
			Bool("tls_enabled", r.TLS != nil).
			Str("scheme", getScheme(r)).
			Interface("security_headers", securityHeaders).
			Msg("Security headers applied")

		next.ServeHTTP(w, r)
	})
}

// enhancedLoggingMiddleware provides comprehensive HTTP request logging
func (h *HTTPServer) enhancedLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := getRequestID(r.Context())

		logger := log.With().
			Str("request_id", requestID).
			Str("middleware", "logging").
			Logger()

		// Log incoming request details
		logger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("raw_query", r.URL.RawQuery).
			Str("user_agent", r.UserAgent()).
			Str("remote_addr", r.RemoteAddr).
			Str("host", r.Host).
			Str("proto", r.Proto).
			Int64("content_length", r.ContentLength).
			Str("content_type", r.Header.Get("Content-Type")).
			Str("accept", r.Header.Get("Accept")).
			Str("referer", r.Header.Get("Referer")).
			Bool("auth_header_present", r.Header.Get("Authorization") != "").
			Interface("request_headers", sanitizeHeaders(r.Header)).
			Msg("HTTP request received")

		// Create a response writer wrapper to capture status and size
		wrapped := &enhancedResponseWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			responseSize:   0,
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate timing
		duration := time.Since(start)

		// Determine log level based on status code
		var logEvent *zerolog.Event
		if wrapped.statusCode >= 500 {
			logEvent = logger.Error()
		} else if wrapped.statusCode >= 400 {
			logEvent = logger.Warn()
		} else {
			logEvent = logger.Info()
		}

		// Log response details
		logEvent.
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("raw_query", r.URL.RawQuery).
			Int("status", wrapped.statusCode).
			Str("status_text", http.StatusText(wrapped.statusCode)).
			Int64("response_size", wrapped.responseSize).
			Dur("duration", duration).
			Str("user_agent", r.UserAgent()).
			Str("remote_addr", r.RemoteAddr).
			Interface("response_headers", sanitizeHeaders(w.Header())).
			Float64("duration_ms", float64(duration.Nanoseconds())/1e6).
			Msg("HTTP request completed")

		// Log performance warnings for slow requests
		if duration > 5*time.Second {
			logger.Warn().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("duration", duration).
				Msg("Slow HTTP request detected")
		}

		// Log SSE-specific details if this is an SSE endpoint
		if strings.HasPrefix(r.URL.Path, "/sse") {
			logger.Debug().
				Str("connection_type", "sse").
				Str("cache_control", w.Header().Get("Cache-Control")).
				Str("connection", w.Header().Get("Connection")).
				Bool("streaming_response", wrapped.statusCode == http.StatusOK).
				Msg("SSE connection details")
		}
	})
}

// recoveryMiddleware recovers from panics
func (h *HTTPServer) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				requestID := getRequestID(r.Context())

				logger := log.With().
					Str("request_id", requestID).
					Str("middleware", "recovery").
					Logger()

				logger.Error().
					Interface("panic", err).
					Str("path", r.URL.Path).
					Str("method", r.Method).
					Str("user_agent", r.UserAgent()).
					Str("remote_addr", r.RemoteAddr).
					Interface("request_headers", sanitizeHeaders(r.Header)).
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
	requestID := getRequestID(r.Context())

	logger := log.With().
		Str("request_id", requestID).
		Str("handler", "health").
		Logger()

	logger.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Msg("Health check endpoint hit")

	response := fmt.Sprintf(`{"status":"ok","timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))

	logger.Debug().
		Int("response_size", len(response)).
		Msg("Health check response sent")
}

func (h *HTTPServer) handleReady(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	logger := log.With().
		Str("request_id", requestID).
		Str("handler", "ready").
		Logger()

	logger.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Msg("Readiness check endpoint hit")

	// Check if all components are ready
	ready := true
	components := make(map[string]bool)

	components["auth"] = h.authValidator != nil
	components["mcp"] = h.mcpServer != nil
	components["transport"] = h.transport != nil
	components["discovery"] = h.discovery != nil

	logger.Debug().
		Interface("component_status", components).
		Msg("Component readiness status checked")

	for _, status := range components {
		if !status {
			ready = false
			break
		}
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	response := fmt.Sprintf(`{"ready":%t,"components":%v,"timestamp":"%s"}`,
		ready, components, time.Now().UTC().Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write([]byte(response))

	logger.Debug().
		Bool("ready", ready).
		Int("status_code", statusCode).
		Int("response_size", len(response)).
		Interface("components", components).
		Msg("Readiness check response sent")
}

func (h *HTTPServer) handleLive(w http.ResponseWriter, r *http.Request) {
	requestID := getRequestID(r.Context())

	logger := log.With().
		Str("request_id", requestID).
		Str("handler", "liveness").
		Logger()

	logger.Debug().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Msg("Liveness check endpoint hit")

	response := fmt.Sprintf(`{"alive":true,"timestamp":"%s"}`, time.Now().UTC().Format(time.RFC3339))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))

	logger.Debug().
		Int("response_size", len(response)).
		Msg("Liveness check response sent")
}

// Start begins serving HTTP requests
func (h *HTTPServer) Start(ctx context.Context) error {
	startTime := time.Now()

	log.Debug().
		Str("addr", h.server.Addr).
		Dur("read_header_timeout", h.server.ReadHeaderTimeout).
		Dur("write_timeout", h.server.WriteTimeout).
		Dur("idle_timeout", h.server.IdleTimeout).
		Msg("Initializing HTTP server startup")

	// Setup routes with all configured components
	h.setupRoutes()

	log.Info().
		Str("addr", h.server.Addr).
		Bool("auth_configured", h.authValidator != nil).
		Bool("mcp_configured", h.mcpServer != nil).
		Bool("transport_configured", h.transport != nil).
		Bool("discovery_configured", h.discovery != nil).
		Msg("Starting HTTP server")

	// Start server in goroutine
	go func() {
		log.Debug().
			Str("addr", h.server.Addr).
			Msg("HTTP server goroutine started, calling ListenAndServe")

		if err := h.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().
				Err(err).
				Str("addr", h.server.Addr).
				Dur("uptime", time.Since(startTime)).
				Msg("HTTP server failed")
		} else {
			log.Debug().
				Str("addr", h.server.Addr).
				Dur("uptime", time.Since(startTime)).
				Msg("HTTP server stopped gracefully")
		}
	}()

	log.Debug().
		Str("addr", h.server.Addr).
		Dur("startup_duration", time.Since(startTime)).
		Msg("HTTP server startup completed")

	return nil
}

// Stop gracefully shuts down the HTTP server
func (h *HTTPServer) Stop(ctx context.Context) error {
	shutdownStart := time.Now()
	shutdownTimeout := 30 * time.Second

	log.Info().
		Str("addr", h.server.Addr).
		Dur("shutdown_timeout", shutdownTimeout).
		Msg("Shutting down HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	err := h.server.Shutdown(shutdownCtx)
	shutdownDuration := time.Since(shutdownStart)

	if err != nil {
		log.Error().
			Err(err).
			Str("addr", h.server.Addr).
			Dur("shutdown_duration", shutdownDuration).
			Dur("shutdown_timeout", shutdownTimeout).
			Msg("HTTP server shutdown failed")
	} else {
		log.Info().
			Str("addr", h.server.Addr).
			Dur("shutdown_duration", shutdownDuration).
			Msg("HTTP server shutdown completed")
	}

	return err
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

type enhancedResponseWrapper struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
}

func (rw *enhancedResponseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *enhancedResponseWrapper) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.responseSize += int64(size)
	return size, err
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

// getRequestID extracts the request ID from context
func getRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("requestID").(string); ok {
		return requestID
	}
	return "unknown"
}

// getScheme determines the scheme (http/https) from the request
func getScheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	// Check for forwarded proto headers (common in reverse proxy setups)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	if proto := r.Header.Get("X-Forwarded-Protocol"); proto != "" {
		return proto
	}
	return "http"
}

// sanitizeHeaders removes sensitive headers for logging
func sanitizeHeaders(headers http.Header) map[string]string {
	sanitized := make(map[string]string)
	for key, values := range headers {
		lowerKey := strings.ToLower(key)

		// Skip sensitive headers
		if lowerKey == "authorization" || lowerKey == "cookie" || lowerKey == "set-cookie" {
			sanitized[key] = "[REDACTED]"
			continue
		}

		// Log first value only for multi-value headers
		if len(values) > 0 {
			sanitized[key] = values[0]
		}
	}
	return sanitized
}
