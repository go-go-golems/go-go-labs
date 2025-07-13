package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/auth"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/config"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/discovery"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/mcp"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/middleware"
)

var (
	logLevel string
	envFile  string
)

func main() {
	startTime := time.Now()
	pid := os.Getpid()

	// Initialize basic logging first
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().
		Int("pid", pid).
		Str("go_version", runtime.Version()).
		Int("num_cpu", runtime.NumCPU()).
		Str("arch", runtime.GOARCH).
		Str("os", runtime.GOOS).
		Msg("Application starting")

	rootCmd := &cobra.Command{
		Use:   "remote-chatgpt-connector",
		Short: "Self-contained MCP connector with OIDC authentication",
		Long: `A self-contained Model Context Protocol (MCP) server that implements its own 
OIDC Authorization Server with dynamic client registration. Allows ChatGPT and other
MCP clients to register dynamically and perform authenticated searches/fetches.

Features:
- Built-in OIDC Authorization Server using Fosite
- Dynamic Client Registration (RFC 7591)
- OAuth 2.0 Authorization Code flow with PKCE
- Bearer token validation for all MCP endpoints
- Search and fetch capabilities with demo data
- SSE transport for real-time communication`,
		RunE: runServer,
	}

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&envFile, "env-file", ".env", "Environment file path")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().
			Err(err).
			Dur("startup_duration", time.Since(startTime)).
			Int("pid", pid).
			Msg("Application failed to start")
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	startTime := time.Now()
	ctx := cmd.Context()

	// Load environment file if it exists
	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			log.Warn().
				Err(err).
				Str("env_file", envFile).
				Msg("Could not load environment file, using system environment")
		} else {
			log.Info().
				Str("env_file", envFile).
				Msg("Environment file loaded successfully")
		}
	}

	// Configure logging based on level
	setupLogging(logLevel)

	logger := log.With().
		Str("component", "main").
		Int("pid", os.Getpid()).
		Logger()

	logger.Info().
		Str("log_level", logLevel).
		Msg("Remote ChatGPT MCP Connector with OIDC starting")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Error().
			Err(err).
			Dur("startup_duration", time.Since(startTime)).
			Msg("Failed to load configuration")
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	logger.Info().
		Interface("config", map[string]interface{}{
			"host":   cfg.Host,
			"port":   cfg.Port,
			"issuer": cfg.OIDCIssuer,
		}).
		Msg("Configuration loaded successfully")

	// Initialize OIDC provider
	logger.Info().Msg("Initializing OIDC Authorization Server")
	oidcConfig := &auth.OIDCConfig{
		Issuer:          cfg.OIDCIssuer,
		Port:            cfg.Port,
		DefaultUser:     cfg.DefaultUser,
		DefaultPassword: cfg.DefaultPassword,
	}

	oidcProvider, err := auth.NewOIDCProvider(oidcConfig)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("startup_duration", time.Since(startTime)).
			Msg("Failed to initialize OIDC provider")
		return fmt.Errorf("failed to initialize OIDC provider: %w", err)
	}

	logger.Info().Msg("OIDC Authorization Server initialized successfully")

	// Initialize authentication middleware
	authMiddleware := middleware.NewAuthMiddleware(oidcProvider)

	// Initialize discovery service
	logger.Info().Msg("Initializing discovery service")
	manifestService := discovery.NewManifestService(*cfg)

	// Initialize MCP wrapper
	logger.Info().Msg("Initializing MCP server")
	mcpWrapper, err := mcp.NewMCPWrapper(
		logger.With().Str("component", "mcp").Logger(),
		*cfg,
	)
	if err != nil {
		logger.Error().
			Err(err).
			Dur("startup_duration", time.Since(startTime)).
			Msg("Failed to initialize MCP wrapper")
		return fmt.Errorf("failed to initialize MCP wrapper: %w", err)
	}

	// Register handlers
	logger.Info().Msg("Registering MCP handlers")
	if err := mcpWrapper.RegisterSearch(mcp.DemoSearchHandler); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to register search handler")
		return fmt.Errorf("failed to register search handler: %w", err)
	}

	if err := mcpWrapper.RegisterFetch(mcp.DemoFetchHandler); err != nil {
		logger.Error().
			Err(err).
			Msg("Failed to register fetch handler")
		return fmt.Errorf("failed to register fetch handler: %w", err)
	}

	// Set authentication middleware
	mcpWrapper.SetAuthMiddleware(authMiddleware)

	// Note: We don't call mcpWrapper.Start() here because it would start its own HTTP server
	// Instead, we'll just get the HTTP handler and integrate it into our main router

	// Create HTTP router
	router := mux.NewRouter()

	// OIDC endpoints
	router.HandleFunc("/register", oidcProvider.DynamicClientRegistration).Methods("POST")
	router.HandleFunc("/authorize", oidcProvider.AuthorizeHandler).Methods("GET", "POST")
	router.HandleFunc("/token", oidcProvider.TokenHandler).Methods("POST")

	// Discovery endpoints
	router.HandleFunc("/.well-known/ai-plugin.json", manifestService.ServeManifest).Methods("GET")
	router.HandleFunc("/.well-known/oauth-authorization-server", manifestService.ServeOAuthConfig).Methods("GET")
	router.HandleFunc("/.well-known/jwks.json", manifestService.ServeJWKS).Methods("GET")
	router.HandleFunc("/userinfo", manifestService.ServeUserInfo).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"remote-chatgpt-connector"}`))
	}).Methods("GET")

	// MCP endpoints (protected by authentication middleware)
	mcpHandler := mcpWrapper.GetHTTPHandler()
	router.PathPrefix("/").Handler(mcpHandler)

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start HTTP server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info().
			Str("address", server.Addr).
			Msg("Starting HTTP server")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	logger.Info().
		Str("address", server.Addr).
		Str("issuer", cfg.OIDCIssuer).
		Dur("startup_duration", time.Since(startTime)).
		Msg("Server started successfully")

	// Log available endpoints
	logger.Info().
		Strs("endpoints", []string{
			"POST /register - Dynamic client registration",
			"GET /authorize - OAuth authorization endpoint",
			"POST /token - Token exchange endpoint",
			"GET /.well-known/ai-plugin.json - Plugin manifest",
			"GET /.well-known/oauth-authorization-server - OIDC discovery",
			"GET /.well-known/jwks.json - JSON Web Key Set",
			"GET /userinfo - User information endpoint",
			"GET /health - Health check",
			"GET /sse - MCP Server-Sent Events (protected)",
			"POST /messages - MCP messages (protected)",
		}).
		Msg("Available endpoints")

	// Wait for interrupt signal or server error
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error().
			Err(err).
			Msg("Server error occurred")
		return fmt.Errorf("server error: %w", err)

	case sig := <-interrupt:
		logger.Info().
			Str("signal", sig.String()).
			Msg("Interrupt signal received, shutting down gracefully")

		// Create shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Shutdown HTTP server
		logger.Info().Msg("Shutting down HTTP server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error().
				Err(err).
				Msg("HTTP server shutdown error")
		}

		// Stop MCP server
		logger.Info().Msg("Shutting down MCP server")
		if err := mcpWrapper.Stop(); err != nil {
			logger.Error().
				Err(err).
				Msg("MCP server shutdown error")
		}

		logger.Info().
			Dur("total_runtime", time.Since(startTime)).
			Msg("Shutdown completed successfully")
	}

	return nil
}

func setupLogging(level string) {
	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Warn().
			Str("provided_level", level).
			Str("default_level", "info").
			Msg("Invalid log level provided, using default")
	}

	// Add caller information in debug mode
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		log.Logger = log.With().Caller().Logger()
	}

	log.Info().
		Str("log_level", level).
		Str("effective_level", zerolog.GlobalLevel().String()).
		Bool("caller_enabled", zerolog.GlobalLevel() == zerolog.DebugLevel).
		Msg("Logging configured")
}
