package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/animal-website/internal/animals"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/animal-website/internal/db"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/animal-website/internal/httpui"
	"github.com/rs/zerolog"
)

func main() {
	var (
		logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
		listenAddr = flag.String("listen-addr", ":8080", "Address to listen on")
		dbPath     = flag.String("db-path", "./animals.db", "Path to SQLite database file")
	)
	flag.Parse()

	// Setup logging
	level, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		log.Fatalf("Invalid log level: %v", err)
	}
	zerolog.SetGlobalLevel(level)
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().
		Str("listen_addr", *listenAddr).
		Str("db_path", *dbPath).
		Msg("Starting animal website server")

	// Initialize database
	db, err := db.OpenDB(*dbPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to open database")
	}
	defer db.Close()

	// Setup HTTP handlers
	repo := animals.NewRepository(db)
	handlers := httpui.NewHandlers(repo, logger)

	mux := http.NewServeMux()
	handlers.RegisterRoutes(mux)

	server := &http.Server{
		Addr:    *listenAddr,
		Handler: mux,
	}

	// Graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-ctx.Done()
		logger.Info().Msg("Shutting down server...")
		if err := server.Shutdown(context.Background()); err != nil {
			logger.Error().Err(err).Msg("Error shutting down server")
		}
	}()

	logger.Info().Msg("Server starting")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal().Err(err).Msg("Server failed")
	}
}

