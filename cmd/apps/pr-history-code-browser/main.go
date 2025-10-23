package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-go-golems/go-go-labs/cmd/apps/pr-history-code-browser/internal/handlers"
	"github.com/go-go-golems/go-go-labs/cmd/apps/pr-history-code-browser/internal/models"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

//go:embed frontend/dist
var frontendFiles embed.FS

// Config holds application configuration
type Config struct {
	Port   int
	DBPath string
	DevMode bool
}

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	// Default configuration
	config := Config{
		Port:   8080,
		DBPath: "",
		DevMode: false,
	}

	// Define root command
	rootCmd := &cobra.Command{
		Use:   "pr-history-code-browser",
		Short: "PR History and Code Browser - Visualize git history and PR work",
		Long: `A web application that visualizes the git history database and PR work tracking.
Provides an interactive interface to browse commits, files, PRs, and analysis notes.`,
		Run: func(cmd *cobra.Command, args []string) {
			if config.DBPath == "" {
				log.Fatal().Msg("Database path is required (use --db flag)")
			}

			if err := runServer(config); err != nil {
				log.Fatal().Err(err).Msg("Server failed")
			}
		},
	}

	// Add flags
	rootCmd.Flags().IntVarP(&config.Port, "port", "p", config.Port, "Port to listen on")
	rootCmd.Flags().StringVarP(&config.DBPath, "db", "d", config.DBPath, "Path to SQLite database file (required)")
	rootCmd.Flags().BoolVar(&config.DevMode, "dev", config.DevMode, "Enable development mode (allows CORS)")

	// Mark db flag as required
	rootCmd.MarkFlagRequired("db")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Command execution failed")
	}
}

func runServer(config Config) error {
	// Open database
	db, err := models.NewDB(config.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	log.Info().Str("path", config.DBPath).Msg("Connected to database")

	// Create handlers
	handler := handlers.NewHandler(db)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS for development
	if config.DevMode {
		log.Info().Msg("Development mode enabled - CORS allowed from all origins")
		corsHandler := cors.New(cors.Options{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders: []string{"*"},
		})
		r.Use(corsHandler.Handler)
	}

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Statistics
		r.Get("/stats", handler.HandleGetStats)

		// Commits
		r.Get("/commits", handler.HandleListCommits)
		r.Get("/commits/{hash}", handler.HandleGetCommit)

		// PRs
		r.Get("/prs", handler.HandleListPRs)
		r.Get("/prs/{id}", handler.HandleGetPR)

		// Files
		r.Get("/files", handler.HandleListFiles)
		r.Get("/files/{id}/history", handler.HandleGetFileHistory)

		// Analysis Notes
		r.Get("/notes", handler.HandleListAnalysisNotes)
	})

	// Serve frontend
	if config.DevMode {
		// In dev mode, frontend is served separately by Vite
		log.Info().Msg("Frontend should be served by Vite dev server on http://localhost:5173")
	} else {
		// In production, serve embedded frontend files
		frontendFS, err := fs.Sub(frontendFiles, "frontend/dist")
		if err != nil {
			return fmt.Errorf("failed to load frontend files: %w", err)
		}

		// Serve static files
		fileServer := http.FileServer(http.FS(frontendFS))
		r.Handle("/*", fileServer)
	}

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      r,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server in a separate goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info().Int("port", config.Port).Msg("Starting server")
		serverErr <- server.ListenAndServe()
	}()

	// Wait for interrupt signal
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for either an error or an interrupt
	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case <-stop:
		log.Info().Msg("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("error during server shutdown: %w", err)
		}
	}

	return nil
}

