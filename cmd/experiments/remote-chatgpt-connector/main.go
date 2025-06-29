package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/auth"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/config"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/discovery"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/mcp"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/server"
)

var (
	logLevel string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "remote-chatgpt-connector",
		Short: "Remote MCP connector for ChatGPT with GitHub OAuth",
		Long: `A remote Model Context Protocol (MCP) server that allows ChatGPT to connect 
via GitHub OAuth and perform secure searches and data fetches.`,
		RunE: runServer,
	}

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", 
		"Log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	// Configure logging
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	log.Info().Str("level", level.String()).Msg("Starting remote ChatGPT MCP connector")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("allowed_login", cfg.AllowedLogin).
		Msg("Configuration loaded")

	// Create components
	authValidator := auth.NewGitHubAuthValidator(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.AllowedLogin, log.Logger)
	discoveryService := discovery.NewService(cfg)
	mcpWrapper, err := mcp.NewMCPWrapper(log.Logger, *cfg)
	if err != nil {
		return fmt.Errorf("failed to create MCP wrapper: %w", err)
	}
	httpServer := server.NewHTTPServer(cfg)

	// Configure MCP with demo handlers
	if err := mcpWrapper.RegisterSearch(mcp.DemoSearchHandler); err != nil {
		return fmt.Errorf("failed to register search handler: %w", err)
	}
	if err := mcpWrapper.RegisterFetch(mcp.DemoFetchHandler); err != nil {
		return fmt.Errorf("failed to register fetch handler: %w", err)
	}

	// Wire components together
	httpServer.SetAuthValidator(authValidator)
	httpServer.SetMCPServer(mcpWrapper)
	httpServer.SetDiscoveryService(discoveryService)

	// Setup routes
	router := mux.NewRouter()
	
	// Discovery endpoints (no auth required)
	router.HandleFunc("/.well-known/ai-plugin.json", discoveryService.GetPluginManifestHandler()).Methods("GET")
	router.HandleFunc("/.well-known/oauth-authorization-server", discoveryService.GetOAuthConfigHandler()).Methods("GET")
	
	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy","service":"remote-chatgpt-connector"}`)
	}).Methods("GET")

	// MCP SSE endpoint (requires auth) - wire auth manually since httpServer doesn't handle this
	router.Handle("/sse", authMiddleware(authValidator, mcpWrapper.GetHTTPHandler())).Methods("GET", "POST")

	// Create HTTP server
	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("addr", srv.Addr).Msg("Starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
		return err
	}

	log.Info().Msg("Server shutdown complete")
	return nil
}

// authMiddleware creates auth middleware that validates GitHub tokens
func authMiddleware(validator *auth.GitHubAuthValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn().Msg("Missing or invalid Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[7:] // Remove "Bearer " prefix
		userInfo, err := validator.ValidateToken(r.Context(), token)
		if err != nil {
			log.Warn().Err(err).Msg("Token validation failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Debug().Str("user", userInfo.Login).Msg("Token validated successfully")
		next.ServeHTTP(w, r)
	})
}
