package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/database"
	"github.com/go-go-golems/go-go-labs/cmd/apps/lumon-stream/backend/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

var (
	port      int
	dbPath    string
	debugMode bool
)

func init() {
	flag.IntVar(&port, "port", 8080, "Port to run the server on")
	flag.StringVar(&dbPath, "db", "./lumonstream.db", "Path to SQLite database file")
	flag.BoolVar(&debugMode, "debug", false, "Run in debug mode")
}

func main() {
	flag.Parse()

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if _, err := os.Stat(dbDir); os.IsNotExist(err) {
		if err := os.MkdirAll(dbDir, 0755); err != nil {
			log.Fatalf("Failed to create database directory: %v", err)
		}
	}

	// Initialize database
	if err := database.InitDB(dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Create router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/stream-info", handlers.GetStreamInfo).Methods("GET")
	api.HandleFunc("/stream-info", handlers.UpdateStreamInfo).Methods("POST")
	api.HandleFunc("/steps", handlers.AddStep).Methods("POST")
	api.HandleFunc("/steps/status", handlers.UpdateStepStatus).Methods("POST")

	// Serve static files in production mode
	if !debugMode {
		// When built with -tags embed, this will use the embedded files
		// Otherwise, this function is a no-op in debug mode
		SetupStaticFiles(r)
	}

	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // In production, you should restrict this
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})

	// Create server
	handler := c.Handler(r)

	// Start server
	serverAddr := fmt.Sprintf("0.0.0.0:%d", port)
	log.Printf("Starting server in %s mode on %s", map[bool]string{true: "debug", false: "production"}[debugMode], serverAddr)
	log.Printf("Database path: %s", dbPath)
	log.Fatal(http.ListenAndServe(serverAddr, handler))
}
