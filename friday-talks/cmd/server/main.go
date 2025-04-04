package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/spf13/cobra"
	"github.com/wesen/friday-talks/internal/auth"
	"github.com/wesen/friday-talks/internal/handlers"
	"github.com/wesen/friday-talks/internal/models"
	"github.com/wesen/friday-talks/internal/services"
)

//go:embed ../../migrations
var migrationsFS embed.FS

//go:embed ../../static
var staticFiles embed.FS

// Config holds the application configuration
type Config struct {
	Port       int
	DBPath     string
	JWTSecret  string
	StaticPath string
}

func main() {
	// Default configuration
	config := Config{
		Port:       8080,
		DBPath:     "friday-talks.db",
		JWTSecret:  "your-secret-key", // In production, use a secure random key
		StaticPath: "static",
	}

	// Define root command
	rootCmd := &cobra.Command{
		Use:   "friday-talks",
		Short: "Friday Talks - A talk scheduling application",
		Run: func(cmd *cobra.Command, args []string) {
			// Create the server
			if err := runServer(config); err != nil {
				log.Fatalf("Error running server: %v", err)
			}
		},
	}

	// Add flags
	rootCmd.Flags().IntVarP(&config.Port, "port", "p", config.Port, "Port to listen on")
	rootCmd.Flags().StringVarP(&config.DBPath, "db", "d", config.DBPath, "Path to SQLite database file")
	rootCmd.Flags().StringVarP(&config.JWTSecret, "jwt-secret", "j", config.JWTSecret, "Secret key for JWT tokens")
	rootCmd.Flags().StringVarP(&config.StaticPath, "static", "s", config.StaticPath, "Path to static files")

	// Execute the root command
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error executing command: %v", err)
	}
}

// runServer initializes and starts the HTTP server
func runServer(config Config) error {
	// Extract migrations FS
	migrations, err := fs.Sub(migrationsFS, "../../migrations")
	if err != nil {
		return fmt.Errorf("error extracting migrations: %w", err)
	}

	// Setup database
	db, err := models.SetupDB(config.DBPath, migrations)
	if err != nil {
		return fmt.Errorf("error setting up database: %w", err)
	}
	defer db.Close()

	// Create repositories
	repos, err := models.NewRepositories(db)
	if err != nil {
		return fmt.Errorf("error creating repositories: %w", err)
	}

	// Create authentication
	authService := auth.NewAuth(config.JWTSecret, repos.User)

	// Create scheduler service
	schedulerService := services.NewSchedulerService(
		repos.Talk,
		repos.Vote,
		repos.Attendance,
	)

	// Create notification config
	notificationConfig := services.NotificationConfig{
		Enabled:    false, // Disable for now, enable in production
		SMTPHost:   "smtp.example.com",
		SMTPPort:   587,
		SMTPUser:   "user@example.com",
		SMTPPass:   "password",
		SenderName: "Friday Talks",
		SenderMail: "talks@example.com",
	}

	// Create notification service
	notificationService := services.NewNotificationService(
		repos.User,
		repos.Talk,
		repos.Attendance,
		notificationConfig,
	)

	// Create handlers
	homeHandler := handlers.NewHomeHandler(repos.Talk)
	authHandler := handlers.NewAuthHandler(repos.User, authService)
	calendarHandler := handlers.NewCalendarHandler(repos.Talk)
	talkHandler := handlers.NewTalkHandler(
		repos.Talk,
		repos.User,
		repos.Vote,
		repos.Attendance,
		repos.Resource,
		schedulerService,
	)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))
	r.Use(authService.AuthMiddleware)

	// Serve static files
	static, err := fs.Sub(staticFiles, "../../static")
	if err != nil {
		return fmt.Errorf("error extracting static files: %w", err)
	}

	fileServer := http.FileServer(http.FS(static))
	r.Handle("/static/*", http.StripPrefix("/static", fileServer))

	// Home routes
	r.Get("/", homeHandler.HandleHome)

	// Auth routes
	r.Get("/login", authHandler.HandleLogin)
	r.Post("/login", authHandler.HandleLogin)
	r.Get("/register", authHandler.HandleRegister)
	r.Post("/register", authHandler.HandleRegister)
	r.Get("/logout", authHandler.HandleLogout)

	// User routes
	r.Route("/profile", func(r chi.Router) {
		r.Use(authService.RequireAuth)
		r.Get("/", authHandler.HandleProfile)
		r.Post("/", authHandler.HandleProfile)
	})

	// Calendar routes
	r.Get("/calendar", calendarHandler.HandleCalendar)

	// Talk routes
	r.Route("/talks", func(r chi.Router) {
		r.Get("/", talkHandler.HandleListTalks)
		r.Get("/{id}", talkHandler.HandleGetTalk)

		// Routes requiring authentication
		r.Group(func(r chi.Router) {
			r.Use(authService.RequireAuth)

			r.Get("/propose", talkHandler.HandleProposeTalk)
			r.Post("/propose", talkHandler.HandleProposeTalk)

			r.Get("/my", talkHandler.HandleMyTalks)

			r.Post("/{id}/vote", talkHandler.HandleVoteOnTalk)
			r.Post("/{id}/attend", talkHandler.HandleManageAttendance)
			r.Post("/{id}/feedback", talkHandler.HandleProvideFeedback)

			r.Get("/{id}/edit", talkHandler.HandleEditTalk)
			r.Post("/{id}/edit", talkHandler.HandleEditTalk)

			r.Post("/{id}/cancel", talkHandler.HandleCancelTalk)
			r.Post("/{id}/complete", talkHandler.HandleCompleteTalk)

			r.Get("/{id}/schedule", talkHandler.HandleScheduleTalk)
			r.Post("/{id}/schedule", talkHandler.HandleScheduleTalk)

			r.Post("/{id}/resources", talkHandler.HandleAddResource)
			r.Post("/{id}/resources/{resourceId}/delete", talkHandler.HandleDeleteResource)
		})
	})

	// Handle 404
	r.NotFound(homeHandler.HandleNotFound)

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
		log.Printf("Starting server on port %d", config.Port)
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
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			return fmt.Errorf("error during server shutdown: %w", err)
		}
	}

	return nil
}
