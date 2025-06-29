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
	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/server"
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
		Short: "Remote MCP connector for ChatGPT with GitHub OAuth",
		Long: `A remote Model Context Protocol (MCP) server that allows ChatGPT to connect 
via GitHub OAuth and perform secure searches and data fetches.`,
		RunE: runServer,
	}

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", 
		"Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&envFile, "env-file", ".env", 
		"Path to .env file for loading environment variables")

	log.Debug().
		Str("default_log_level", "info").
		Str("default_env_file", ".env").
		Msg("Command line flags configured")

	if err := rootCmd.Execute(); err != nil {
		log.Error().
			Err(err).
			Dur("startup_duration", time.Since(startTime)).
			Msg("Application failed to start")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	serverStartTime := time.Now()
	
	log.Info().
		Str("phase", "startup").
		Str("step", "begin").
		Str("requested_log_level", logLevel).
		Str("env_file_path", envFile).
		Msg("Server startup initiated")

	// Load .env file if specified and exists
	envLoadStart := time.Now()
	log.Debug().
		Str("phase", "env_loading").
		Str("env_file", envFile).
		Bool("env_file_specified", envFile != "").
		Msg("Checking environment file")

	if envFile != "" {
		if stat, err := os.Stat(envFile); err == nil {
			log.Debug().
				Str("env_file", envFile).
				Int64("file_size", stat.Size()).
				Time("mod_time", stat.ModTime()).
				Msg("Environment file found, loading")

			if err := godotenv.Load(envFile); err != nil {
				log.Error().
					Err(err).
					Str("env_file", envFile).
					Msg("Failed to load environment file")
				return fmt.Errorf("failed to load env file %s: %w", envFile, err)
			}
			
			log.Info().
				Str("env_file", envFile).
				Dur("load_duration", time.Since(envLoadStart)).
				Msg("Environment file loaded successfully")
		} else if envFile != ".env" {
			// Only error if user explicitly specified a file that doesn't exist
			log.Error().
				Str("env_file", envFile).
				Err(err).
				Msg("Explicitly specified environment file not found")
			return fmt.Errorf("env file %s not found", envFile)
		} else {
			log.Debug().
				Str("env_file", envFile).
				Msg("Default .env file not found, continuing without it")
		}
	}

	// Configure logging
	loggingConfigStart := time.Now()
	log.Debug().
		Str("phase", "logging_config").
		Str("requested_level", logLevel).
		Msg("Configuring logging system")

	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		log.Error().
			Err(err).
			Str("invalid_level", logLevel).
			Msg("Invalid log level specified")
		return fmt.Errorf("invalid log level: %w", err)
	}
	
	zerolog.SetGlobalLevel(level)
	
	// Enable caller info and configure console writer
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "3:04PM",
	}).With().Caller().Logger()

	log.Info().
		Str("level", level.String()).
		Str("env_file", envFile).
		Dur("logging_config_duration", time.Since(loggingConfigStart)).
		Msg("Logging system configured")

	// Log environment variables (non-sensitive ones)
	log.Debug().
		Str("phase", "env_audit").
		Str("GITHUB_CLIENT_ID", os.Getenv("GITHUB_CLIENT_ID")).
		Bool("GITHUB_CLIENT_SECRET_set", os.Getenv("GITHUB_CLIENT_SECRET") != "").
		Str("ALLOWED_LOGIN", os.Getenv("ALLOWED_LOGIN")).
		Str("HOST", os.Getenv("HOST")).
		Str("PORT", os.Getenv("PORT")).
		Msg("Environment variables audit")

	log.Info().
		Str("phase", "startup").
		Str("step", "config_load").
		Msg("Starting remote ChatGPT MCP connector")

	// Load configuration
	configLoadStart := time.Now()
	log.Debug().
		Str("phase", "config_load").
		Msg("Loading application configuration")

	cfg, err := config.Load()
	if err != nil {
		log.Error().
			Err(err).
			Dur("config_load_duration", time.Since(configLoadStart)).
			Msg("Failed to load configuration")
		return fmt.Errorf("failed to load config: %w", err)
	}

	log.Debug().
		Str("phase", "config_validation").
		Msg("Validating configuration")

	if err := config.Validate(cfg); err != nil {
		log.Error().
			Err(err).
			Dur("config_load_duration", time.Since(configLoadStart)).
			Msg("Configuration validation failed")
		return fmt.Errorf("invalid config: %w", err)
	}

	log.Info().
		Str("host", cfg.Host).
		Int("port", cfg.Port).
		Str("allowed_login", cfg.AllowedLogin).
		Bool("github_client_id_set", cfg.GitHubClientID != "").
		Bool("github_client_secret_set", cfg.GitHubClientSecret != "").
		Dur("config_load_duration", time.Since(configLoadStart)).
		Msg("Configuration loaded and validated")

	// Create components
	componentInitStart := time.Now()
	log.Info().
		Str("phase", "component_init").
		Str("step", "auth_validator").
		Msg("Initializing GitHub auth validator")

	authValidator := auth.NewGitHubAuthValidator(cfg.GitHubClientID, cfg.GitHubClientSecret, cfg.AllowedLogin, log.Logger)
	log.Debug().
		Str("component", "auth_validator").
		Str("allowed_login", cfg.AllowedLogin).
		Msg("GitHub auth validator created")

	log.Debug().
		Str("phase", "component_init").
		Str("step", "discovery_service").
		Msg("Initializing discovery service")
	discoveryService := discovery.NewService(cfg)

	log.Debug().
		Str("phase", "component_init").
		Str("step", "mcp_wrapper").
		Msg("Initializing MCP wrapper")
	mcpWrapper, err := mcp.NewMCPWrapper(log.Logger, *cfg)
	if err != nil {
		log.Error().
			Err(err).
			Dur("component_init_duration", time.Since(componentInitStart)).
			Msg("Failed to create MCP wrapper")
		return fmt.Errorf("failed to create MCP wrapper: %w", err)
	}

	log.Debug().
		Str("phase", "component_init").
		Str("step", "http_server").
		Msg("Initializing HTTP server component")
	httpServer := server.NewHTTPServer(cfg)

	log.Info().
		Dur("component_init_duration", time.Since(componentInitStart)).
		Msg("All components initialized successfully")

	// Configure MCP with demo handlers
	handlerRegStart := time.Now()
	log.Debug().
		Str("phase", "handler_registration").
		Str("handler", "search").
		Msg("Registering MCP search handler")

	if err := mcpWrapper.RegisterSearch(mcp.DemoSearchHandler); err != nil {
		log.Error().
			Err(err).
			Str("handler", "search").
			Msg("Failed to register search handler")
		return fmt.Errorf("failed to register search handler: %w", err)
	}

	log.Debug().
		Str("phase", "handler_registration").
		Str("handler", "fetch").
		Msg("Registering MCP fetch handler")

	if err := mcpWrapper.RegisterFetch(mcp.DemoFetchHandler); err != nil {
		log.Error().
			Err(err).
			Str("handler", "fetch").
			Msg("Failed to register fetch handler")
		return fmt.Errorf("failed to register fetch handler: %w", err)
	}

	log.Info().
		Dur("handler_registration_duration", time.Since(handlerRegStart)).
		Msg("MCP handlers registered successfully")

	// Wire components together
	wireStart := time.Now()
	log.Debug().
		Str("phase", "component_wiring").
		Msg("Wiring components together")

	httpServer.SetAuthValidator(authValidator)
	log.Debug().Str("wiring", "auth_validator").Msg("Auth validator wired to HTTP server")

	httpServer.SetMCPServer(mcpWrapper)
	log.Debug().Str("wiring", "mcp_server").Msg("MCP server wired to HTTP server")

	httpServer.SetDiscoveryService(discoveryService)
	log.Debug().Str("wiring", "discovery_service").Msg("Discovery service wired to HTTP server")

	log.Info().
		Dur("wiring_duration", time.Since(wireStart)).
		Msg("Component wiring completed")

	// Setup routes
	routeSetupStart := time.Now()
	log.Debug().
		Str("phase", "route_setup").
		Msg("Setting up HTTP routes")

	router := mux.NewRouter()
	
	// Discovery endpoints (no auth required)
	log.Debug().
		Str("route", "/.well-known/ai-plugin.json").
		Str("methods", "GET").
		Bool("auth_required", false).
		Msg("Registering plugin manifest route")
	router.HandleFunc("/.well-known/ai-plugin.json", discoveryService.GetPluginManifestHandler()).Methods("GET")
	
	log.Debug().
		Str("route", "/.well-known/oauth-authorization-server").
		Str("methods", "GET").
		Bool("auth_required", false).
		Msg("Registering OAuth config route")
	router.HandleFunc("/.well-known/oauth-authorization-server", discoveryService.GetOAuthConfigHandler()).Methods("GET")
	
	// Health check endpoint
	log.Debug().
		Str("route", "/health").
		Str("methods", "GET").
		Bool("auth_required", false).
		Msg("Registering health check route")
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().
			Str("endpoint", "/health").
			Str("method", r.Method).
			Str("remote_addr", r.RemoteAddr).
			Msg("Health check requested")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"healthy","service":"remote-chatgpt-connector"}`)
	}).Methods("GET")

	// MCP SSE endpoint (requires auth) - wire auth manually since httpServer doesn't handle this
	log.Debug().
		Str("route", "/sse").
		Str("methods", "GET,POST").
		Bool("auth_required", true).
		Msg("Registering MCP SSE route with auth middleware")
	router.Handle("/sse", authMiddleware(authValidator, mcpWrapper.GetHTTPHandler())).Methods("GET", "POST")

	log.Info().
		Dur("route_setup_duration", time.Since(routeSetupStart)).
		Int("total_routes", 4).
		Msg("HTTP routes configured")

	// Create HTTP server
	serverCreateStart := time.Now()
	serverAddr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	log.Debug().
		Str("phase", "server_creation").
		Str("bind_addr", serverAddr).
		Dur("read_header_timeout", 5*time.Second).
		Dur("write_timeout", 30*time.Second).
		Dur("idle_timeout", 60*time.Second).
		Msg("Creating HTTP server")

	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Info().
		Str("bind_addr", srv.Addr).
		Dur("server_creation_duration", time.Since(serverCreateStart)).
		Msg("HTTP server created")

	// Log memory stats before starting server
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	log.Debug().
		Str("phase", "pre_startup_stats").
		Uint64("alloc_mb", memStats.Alloc/1024/1024).
		Uint64("sys_mb", memStats.Sys/1024/1024).
		Uint32("num_gc", memStats.NumGC).
		Int("num_goroutine", runtime.NumGoroutine()).
		Msg("Memory and runtime stats before server start")

	// Start server in goroutine
	serverStarted := make(chan struct{})
	go func() {
		log.Info().
			Str("phase", "server_start").
			Str("addr", srv.Addr).
			Msg("Starting HTTP server")
		
		close(serverStarted) // Signal that we're about to start listening
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().
				Err(err).
				Str("addr", srv.Addr).
				Msg("HTTP server failed")
		}
	}()

	// Wait for server to start
	<-serverStarted
	
	log.Info().
		Str("addr", srv.Addr).
		Dur("total_startup_duration", time.Since(serverStartTime)).
		Msg("Server startup completed successfully")

	// Setup signal handling
	signalSetupStart := time.Now()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	
	log.Debug().
		Str("phase", "signal_setup").
		Strs("signals", []string{"SIGINT", "SIGTERM"}).
		Dur("signal_setup_duration", time.Since(signalSetupStart)).
		Msg("Signal handling configured")

	// Wait for interrupt signal for graceful shutdown
	log.Info().Msg("Server ready - waiting for shutdown signal...")
	sig := <-quit
	
	log.Info().
		Str("signal", sig.String()).
		Msg("Shutdown signal received")

	// Graceful shutdown with timeout
	shutdownStart := time.Now()
	log.Info().
		Str("phase", "shutdown").
		Dur("timeout", 30*time.Second).
		Msg("Initiating graceful shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Log final memory stats
	runtime.ReadMemStats(&memStats)
	log.Debug().
		Str("phase", "pre_shutdown_stats").
		Uint64("alloc_mb", memStats.Alloc/1024/1024).
		Uint64("sys_mb", memStats.Sys/1024/1024).
		Uint32("num_gc", memStats.NumGC).
		Int("num_goroutine", runtime.NumGoroutine()).
		Msg("Memory and runtime stats before shutdown")

	if err := srv.Shutdown(ctx); err != nil {
		log.Error().
			Err(err).
			Dur("shutdown_attempt_duration", time.Since(shutdownStart)).
			Msg("Server forced to shutdown")
		return err
	}

	log.Info().
		Dur("shutdown_duration", time.Since(shutdownStart)).
		Dur("total_runtime", time.Since(serverStartTime)).
		Msg("Server shutdown completed successfully")
	
	return nil
}

// authMiddleware creates auth middleware that validates GitHub tokens
func authMiddleware(validator *auth.GitHubAuthValidator, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authStart := time.Now()
		
		log.Debug().
			Str("middleware", "auth").
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.Header.Get("User-Agent")).
			Msg("Authentication middleware invoked")

		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			log.Warn().
				Str("remote_addr", r.RemoteAddr).
				Str("path", r.URL.Path).
				Bool("auth_header_present", authHeader != "").
				Dur("auth_duration", time.Since(authStart)).
				Msg("Missing or invalid Authorization header")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := authHeader[7:] // Remove "Bearer " prefix
		tokenHash := fmt.Sprintf("sha256:%x", token[:min(8, len(token))])
		
		log.Debug().
			Str("token_prefix", tokenHash).
			Str("remote_addr", r.RemoteAddr).
			Msg("Validating GitHub token")

		userInfo, err := validator.ValidateToken(r.Context(), token)
		if err != nil {
			log.Warn().
				Err(err).
				Str("remote_addr", r.RemoteAddr).
				Str("path", r.URL.Path).
				Str("token_prefix", tokenHash).
				Dur("auth_duration", time.Since(authStart)).
				Msg("Token validation failed")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		log.Info().
			Str("user", userInfo.Login).
			Str("remote_addr", r.RemoteAddr).
			Str("path", r.URL.Path).
			Dur("auth_duration", time.Since(authStart)).
			Msg("Token validated successfully")
		
		next.ServeHTTP(w, r)
	})
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
