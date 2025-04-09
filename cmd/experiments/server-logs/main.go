package main

import (
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Global variables everywhere
var (
	logs        = []LogEntry{}
	metrics     = map[string]int{}
	users       = map[string]User{}
	configFile  = "config.json"
	port        = 8080
	debugMode   = true
	mutex       = &sync.Mutex{}
	tmpl        *template.Template
	sessionKeys = map[string]string{}
	// Add a global logger
	logger zerolog.Logger
)

// Types with loose structure
type LogEntry struct {
	Id        int         `json:"id"`
	Timestamp time.Time   `json:"timestamp"`
	Level     string      `json:"level"`
	Message   string      `json:"message"`
	Source    string      `json:"source,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type User struct {
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	CreatedAt string   `json:"created_at"`
	LastLogin string   `json:"last_login"`
	Roles     []string `json:"roles"`
	ApiKeys   []string `json:"api_keys,omitempty"`
}

type ApiResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Initialize with hardcoded data
func init() {
	rand.Seed(time.Now().UnixNano())

	// Check for DEBUG environment variable
	if os.Getenv("DEBUG") != "" {
		debugMode = true
	}

	// Setup zerolog
	setupLogger()

	logger.Info().Msg("Initializing server")

	// Add some sample logs
	logs = append(logs, LogEntry{
		Id:        1,
		Timestamp: time.Now().Add(-10 * time.Hour),
		Level:     "info",
		Message:   "Server started",
		Source:    "system",
	})

	logs = append(logs, LogEntry{
		Id:        2,
		Timestamp: time.Now().Add(-5 * time.Hour),
		Level:     "error",
		Message:   "Database connection failed",
		Source:    "database",
		Data:      map[string]interface{}{"error": "connection timeout", "retries": 3},
	})

	logs = append(logs, LogEntry{
		Id:        3,
		Timestamp: time.Now().Add(-2 * time.Hour),
		Level:     "warning",
		Message:   "High memory usage detected",
		Source:    "monitoring",
		Data:      map[string]interface{}{"memory_usage": "85%", "threshold": "80%"},
	})

	// Add some sample users
	users["admin"] = User{
		Username:  "admin",
		Email:     "admin@example.com",
		CreatedAt: "2023-01-01T00:00:00Z",
		LastLogin: "2023-06-15T08:30:45Z",
		Roles:     []string{"admin", "user"},
		ApiKeys:   []string{"admin-key-123456"},
	}

	users["user1"] = User{
		Username:  "user1",
		Email:     "user1@example.com",
		CreatedAt: "2023-02-15T00:00:00Z",
		LastLogin: "2023-06-14T15:22:10Z",
		Roles:     []string{"user"},
	}

	// Initialize some metrics
	metrics["requests"] = 0
	metrics["errors"] = 0
	metrics["active_users"] = 0

	// Parse templates - inconsistent pattern for template handling
	tmpl = template.Must(template.New("index").Parse(indexTemplate))
}

// Setup zerolog logger
func setupLogger() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if debugMode {
		// Pretty logging for development
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		logger = zerolog.New(consoleWriter).With().Timestamp().Caller().Logger()

		// Set global log level to debug in debug mode
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// JSON logging for production
		logger = zerolog.New(os.Stdout).With().Timestamp().Logger()

		// Set global log level to info in production mode
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

func main() {
	// Check command line arguments in a messy way
	if len(os.Args) > 1 {
		p, err := strconv.Atoi(os.Args[1])
		if err == nil {
			port = p
		} else {
			logger.Warn().Str("input", os.Args[1]).Msg("Invalid port number, using default")
			fmt.Println("Invalid port number, using default:", port)
		}
	}

	// Initialize router without using a proper router package
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/api/logs", handleLogs)
	http.HandleFunc("/api/logs/", handleLogById)
	http.HandleFunc("/api/users", handleUsers)
	http.HandleFunc("/api/users/", handleUserByName)
	http.HandleFunc("/api/metrics", handleMetrics)
	http.HandleFunc("/api/config", handleConfig)
	http.HandleFunc("/dashboard", handleDashboard)

	// Serve static files
	staticDir := "./static"
	fs := http.FileServer(http.Dir(staticDir))

	// Logging wrapper for static file server
	loggedStaticHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a wrapper for http.ResponseWriter to capture status code
		wrapped := newResponseWriter(w)

		// Start timing the request
		startTime := time.Now()

		// Log the static file request
		staticPath := strings.TrimPrefix(r.URL.Path, "/static/")
		logger.Debug().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("static_path", staticPath).
			Str("remote_addr", r.RemoteAddr).
			Msg("Static file request received")

		// Serve the file
		fs.ServeHTTP(wrapped, r)

		// Calculate request duration
		duration := time.Since(startTime)

		// Log the completed request
		logger.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("static_path", staticPath).
			Int("status", wrapped.status).
			Int("content_length", wrapped.contentLength).
			Dur("duration_ms", duration).
			Msg("Static file served")
	})

	http.Handle("/static/", http.StripPrefix("/static/", loggedStaticHandler))

	// Middleware implemented in a messy way
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Create a request ID for tracking
		requestID := fmt.Sprintf("req-%d", rand.Int63())

		// Create wrapped response writer
		wrapped := newResponseWriter(w)

		// Start timing the request
		startTime := time.Now()

		// Create request-scoped logger with request ID
		reqLogger := logger.With().
			Str("request_id", requestID).
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Logger()

		reqLogger.Info().Msg("API request received")

		// Log request headers at debug level
		if debugMode {
			headerMap := make(map[string]string)
			for key, values := range r.Header {
				headerMap[key] = strings.Join(values, ", ")
			}
			reqLogger.Debug().Interface("headers", headerMap).Msg("Request headers")
		}

		// Ugly auth check
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			reqLogger.Warn().Str("auth_header", authHeader).Msg("Unauthorized request")
			http.Error(wrapped, "Unauthorized", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		reqLogger.Debug().Str("token", token).Msg("Auth token extracted")

		// Increment request count
		mutex.Lock()
		metrics["requests"]++
		reqCount := metrics["requests"]
		mutex.Unlock()

		reqLogger.Debug().Int("total_requests", reqCount).Msg("Request counter incremented")

		// Continue with the actual handler
		// Find the registered handler for the specific API path
		// This is still messy as it relies on DefaultServeMux internals
		handler, pattern := http.DefaultServeMux.Handler(r)
		reqLogger.Debug().Str("matched_pattern", pattern).Msg("Matched route pattern")

		// Only call the handler if it's not the catch-all /api/ itself
		if pattern != "/api/" {
			// Call the handler
			handler.ServeHTTP(wrapped, r)

			// Calculate request duration
			duration := time.Since(startTime)

			// Log the completed request
			reqLogger.Info().
				Int("status", wrapped.status).
				Dur("duration_ms", duration).
				Int("content_length", wrapped.contentLength).
				Str("pattern", pattern).
				Msg("API request completed")
		} else {
			// Handle cases where no more specific /api/ handler matches
			reqLogger.Warn().Msg("No specific handler found for API path")
			http.NotFound(wrapped, r)
		}
	})

	// Start the server
	logger.Info().Int("port", port).Msg("Server starting")
	fmt.Printf("Server starting on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Root handler
func handleRoot(w http.ResponseWriter, r *http.Request) {
	// Create wrapped response writer
	wrapped := newResponseWriter(w)

	// Log request
	reqLog := logRequest(r).
		Str("handler", "handleRoot")

	// Defer logging response
	defer func() {
		logResponse(wrapped.status, r.Method, r.URL.Path).
			Str("handler", "handleRoot").
			Int("content_length", wrapped.contentLength).
			Msg("Response sent")
	}()

	reqLog.Msg("Request received")

	if r.URL.Path != "/" {
		logger.Warn().Str("path", r.URL.Path).Msg("Path not found")
		http.NotFound(wrapped, r)
		return
	}

	// Just redirect to dashboard
	logger.Debug().Msg("Redirecting to dashboard")
	http.Redirect(wrapped, r, "/dashboard", http.StatusSeeOther)
}

// Dashboard handler - loads and executes template file
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Create wrapped response writer
	wrapped := newResponseWriter(w)

	// Log request
	reqLog := logRequest(r).
		Str("handler", "handleDashboard")

	// Defer logging response
	defer func() {
		logResponse(wrapped.status, r.Method, r.URL.Path).
			Str("handler", "handleDashboard").
			Int("content_length", wrapped.contentLength).
			Msg("Dashboard response sent")
	}()

	reqLog.Msg("Request received")

	// Load template on each request (messy and inefficient)
	tmplPath := filepath.Join("static", "index.html")
	logger.Debug().Str("template_path", tmplPath).Msg("Loading template")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		logger.Error().Err(err).Str("template_path", tmplPath).Msg("Failed to load template")
		http.Error(wrapped, "Failed to load template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	mutex.Lock()
	// Directly modifying global state
	metrics["active_users"]++
	logCount := len(logs)
	userCount := len(users)
	metricsCopy := make(map[string]int)
	for k, v := range metrics {
		metricsCopy[k] = v
	}
	mutex.Unlock()

	logger.Debug().
		Int("log_count", logCount).
		Int("user_count", userCount).
		Int("active_users", metricsCopy["active_users"]).
		Int("metric_count", len(metricsCopy)).
		Msg("Prepared dashboard data")

	data := map[string]interface{}{
		"Title":      "Log Server Dashboard",
		"LogCount":   logCount,
		"UserCount":  userCount,
		"Metrics":    metricsCopy, // Use the copy
		"RecentLogs": getRecentLogs(5),
		"ServerTime": time.Now().Format(time.RFC1123),
	}

	err = tmpl.Execute(wrapped, data)
	if err != nil {
		logger.Error().Err(err).Msg("Template execution error")
		http.Error(wrapped, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	logger.Info().Msg("Dashboard rendered successfully")
}

// Get recent logs with duplicated code
func getRecentLogs(count int) []LogEntry {
	logger.Debug().Int("requested_count", count).Msg("Getting recent logs")

	mutex.Lock()
	defer mutex.Unlock()

	if count > len(logs) {
		logger.Debug().Int("requested_count", count).Int("available_count", len(logs)).Msg("Requested more logs than available")
		count = len(logs)
	}

	result := make([]LogEntry, count)
	// Get the last 'count' logs (assuming they are added chronologically)
	startIndex := len(logs) - count
	if startIndex < 0 {
		startIndex = 0
	}
	copy(result, logs[startIndex:])

	// Reverse the slice so newest is first
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	logger.Debug().
		Int("start_index", startIndex).
		Int("result_count", len(result)).
		Msg("Recent logs retrieved")

	return result
}
