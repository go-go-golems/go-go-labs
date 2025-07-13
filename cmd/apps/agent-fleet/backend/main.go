package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/database"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/handlers"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/backend/middleware/auth"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "agent-fleet-backend",
		Short: "Agent Fleet Management Backend Server",
		Long:  `A backend server for managing and monitoring AI coding agent fleets with REST API and real-time updates.`,
		Run:   runServer,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.agent-fleet.yaml)")
	rootCmd.PersistentFlags().StringP("port", "p", "8080", "Port to run the server on")
	rootCmd.PersistentFlags().StringP("host", "H", "localhost", "Host to bind the server to")
	rootCmd.PersistentFlags().StringP("database", "d", "./agent-fleet.db", "SQLite database file path")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (trace, debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("dev", "", false, "Development mode")
	rootCmd.PersistentFlags().BoolP("disable-auth", "", false, "Disable authentication for testing")

	viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("database", rootCmd.PersistentFlags().Lookup("database"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("dev", rootCmd.PersistentFlags().Lookup("dev"))
	viper.BindPFlag("disable-auth", rootCmd.PersistentFlags().Lookup("disable-auth"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".agent-fleet")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("config", viper.ConfigFileUsed()).Msg("Using config file")
	}
}

func setupLogger() {
	level := viper.GetString("log-level")
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		return fmt.Sprintf("%s:%d", short, line)
	}

	if viper.GetBool("dev") {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()
	} else {
		log.Logger = log.With().Caller().Logger()
	}
}

func runServer(cmd *cobra.Command, args []string) {
	setupLogger()

	dbPath := viper.GetString("database")
	db, err := database.Initialize(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer db.Close()

	log.Info().Str("database", dbPath).Msg("Database initialized")

	// Create handlers
	h := handlers.New(db)

	// Setup router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(c.Handler)

	// API routes
	r.Route("/v1", func(r chi.Router) {
		// Authentication middleware for API routes (unless disabled)
		if !viper.GetBool("disable-auth") {
			r.Use(auth.BearerTokenMiddleware)
		}

		// Agents
		r.Route("/agents", func(r chi.Router) {
			r.Get("/", h.ListAgents)
			r.Post("/", h.CreateAgent)
			r.Get("/{agentID}", h.GetAgent)
			r.Patch("/{agentID}", h.UpdateAgent)
			r.Delete("/{agentID}", h.DeleteAgent)

			// Agent events
			r.Get("/{agentID}/events", h.ListAgentEvents)
			r.Post("/{agentID}/events", h.CreateAgentEvent)

			// Agent todos
			r.Get("/{agentID}/todos", h.ListAgentTodos)
			r.Post("/{agentID}/todos", h.CreateAgentTodo)
			r.Patch("/{agentID}/todos/{todoID}", h.UpdateAgentTodo)
			r.Delete("/{agentID}/todos/{todoID}", h.DeleteAgentTodo)

			// Agent commands
			r.Get("/{agentID}/commands", h.ListAgentCommands)
			r.Post("/{agentID}/commands", h.CreateAgentCommand)
			r.Patch("/{agentID}/commands/{commandID}", h.UpdateAgentCommand)
		})

		// Tasks
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", h.ListTasks)
			r.Post("/", h.CreateTask)
			r.Get("/{taskID}", h.GetTask)
			r.Patch("/{taskID}", h.UpdateTask)
			r.Delete("/{taskID}", h.DeleteTask)
		})

		// Fleet operations
		r.Route("/fleet", func(r chi.Router) {
			r.Get("/status", h.GetFleetStatus)
			r.Get("/recent-updates", h.GetRecentUpdates)
		})

		// Server-Sent Events
		r.Get("/stream", h.SSEHandler)
	})

	// Web interface routes (optional)
	r.Get("/", h.IndexPage)
	r.Get("/agents", h.AgentsPage)

	// Static files
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	host := viper.GetString("host")
	port := viper.GetString("port")
	addr := fmt.Sprintf("%s:%s", host, port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info().
			Str("address", addr).
			Bool("auth_disabled", viper.GetBool("disable-auth")).
			Msg("Starting server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exited")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}
