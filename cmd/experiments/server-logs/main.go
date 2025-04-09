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

func main() {
	// Check command line arguments in a messy way
	if len(os.Args) > 1 {
		p, err := strconv.Atoi(os.Args[1])
		if err == nil {
			port = p
		} else {
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
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Middleware implemented in a messy way
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Ugly auth check
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Increment request count
		mutex.Lock()
		metrics["requests"]++
		mutex.Unlock()

		// Continue with the actual handler
		// Find the registered handler for the specific API path
		// This is still messy as it relies on DefaultServeMux internals
		handler, pattern := http.DefaultServeMux.Handler(r)
		fmt.Printf("API Request: %s, Pattern: %s\n", r.URL.Path, pattern)
		// Only call the handler if it's not the catch-all /api/ itself
		if pattern != "/api/" {
			handler.ServeHTTP(w, r)
		} else {
			// Handle cases where no more specific /api/ handler matches
			http.NotFound(w, r)
		}
	})

	// Start the server
	fmt.Printf("Server starting on port %d...\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

// Root handler
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Just redirect to dashboard
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// Dashboard handler - loads and executes template file
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Load template on each request (messy and inefficient)
	tmplPath := filepath.Join("static", "index.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Failed to load template: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error parsing template %s: %v", tmplPath, err)
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

	data := map[string]interface{}{
		"Title":      "Log Server Dashboard",
		"LogCount":   logCount,
		"UserCount":  userCount,
		"Metrics":    metricsCopy, // Use the copy
		"RecentLogs": getRecentLogs(5),
		"ServerTime": time.Now().Format(time.RFC1123),
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Template execution error: "+err.Error(), http.StatusInternalServerError)
		log.Printf("Error executing template: %v", err)
		return
	}
}

// Get recent logs with duplicated code
func getRecentLogs(count int) []LogEntry {
	mutex.Lock()
	defer mutex.Unlock()

	if count > len(logs) {
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

	return result
}
