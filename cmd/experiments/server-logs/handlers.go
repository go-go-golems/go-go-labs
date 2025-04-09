package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/yaml.v2"
)

// Helper function to log request details
func logRequest(r *http.Request) *zerolog.Event {
	return logger.Info().
		Str("method", r.Method).
		Str("path", r.URL.Path).
		Str("remote_addr", r.RemoteAddr).
		Str("user_agent", r.UserAgent())
}

// Helper function to log response details
func logResponse(status int, method string, path string) *zerolog.Event {
	return logger.Info().
		Int("status", status).
		Str("method", method).
		Str("path", path)
}

// Handle all logs
func handleLogs(w http.ResponseWriter, r *http.Request) {
	reqLog := logRequest(r).
		Str("handler", "handleLogs")

	// Create a wrapper for http.ResponseWriter to capture status code
	wrapped := newResponseWriter(w)

	// Defer logging the response
	defer func() {
		logResponse(wrapped.status, r.Method, r.URL.Path).
			Str("handler", "handleLogs").
			Int("content_length", wrapped.contentLength).
			Str("content_type", wrapped.Header().Get("Content-Type")).
			Msg("API response sent")
	}()

	// Handle different HTTP methods in a single function - messy approach
	if r.Method == "GET" {
		// Log query parameters
		if queryParams := r.URL.Query(); len(queryParams) > 0 {
			reqLog.Interface("query_params", queryParams).Msg("GET logs request received")
		} else {
			reqLog.Msg("GET logs request received")
		}

		mutex.Lock()
		logsData := logs
		mutex.Unlock()

		// Lots of duplicated code and poorly managed filtering
		if levelFilter := r.URL.Query().Get("level"); levelFilter != "" {
			logger.Debug().Str("level_filter", levelFilter).Msg("Filtering logs by level")
			filteredLogs := []LogEntry{}
			for _, log := range logsData {
				if log.Level == levelFilter {
					filteredLogs = append(filteredLogs, log)
				}
			}
			logsData = filteredLogs
		}

		if sourceFilter := r.URL.Query().Get("source"); sourceFilter != "" {
			logger.Debug().Str("source_filter", sourceFilter).Msg("Filtering logs by source")
			filteredLogs := []LogEntry{}
			for _, log := range logsData {
				if log.Source == sourceFilter {
					filteredLogs = append(filteredLogs, log)
				}
			}
			logsData = filteredLogs
		}

		logger.Debug().Int("result_count", len(logsData)).Msg("Logs filtered")

		// Messy: Randomly choose output format (JSON or CSV)
		if rand.Intn(2) == 0 {
			wrapped.Header().Set("Content-Type", "application/json")
			// Messy: Add extraneous field
			response := map[string]interface{}{
				"success":           true,
				"data":              logsData,
				"request_timestamp": time.Now().Unix(),
				"server_node":       "node-" + strconv.Itoa(rand.Intn(10)), // Another extraneous field
			}
			logger.Debug().Str("format", "json").Int("log_count", len(logsData)).Msg("Responding with JSON logs")
			json.NewEncoder(wrapped).Encode(response)
		} else {
			wrapped.Header().Set("Content-Type", "text/plain")
			// Messy: Output as CSV
			logger.Debug().Str("format", "csv").Int("log_count", len(logsData)).Msg("Responding with CSV logs")
			fmt.Fprintln(wrapped, "id,timestamp,level,message,source")
			for _, log := range logsData {
				fmt.Fprintf(wrapped, "%d,%s,%s,%s,%s",
					log.Id, log.Timestamp.Format(time.RFC3339), log.Level, log.Message, log.Source)
			}
		}
		return
	} else if r.Method == "POST" {
		// Messy: Expect 'priority' in query params for a POST request
		priority := r.URL.Query().Get("priority")
		if priority == "" {
			priority = "normal" // Default messy priority
		}

		reqLog.Str("priority", priority).Msg("POST logs request received")

		// Read request body directly
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to read request body")
			http.Error(wrapped, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		logger.Debug().Int("body_size", len(body)).Msg("Request body read")

		// Parse as a log entry
		var newLog LogEntry
		err = json.Unmarshal(body, &newLog)
		if err != nil {
			logger.Error().Err(err).Str("body", string(body)).Msg("Failed to parse JSON body")
			// Messy: Sometimes return error as plain text
			if rand.Intn(2) == 0 {
				http.Error(wrapped, "Invalid log entry format: "+err.Error(), http.StatusBadRequest)
			} else {
				wrapped.Header().Set("Content-Type", "application/json")
				wrapped.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(wrapped).Encode(ApiResponse{Success: false, Error: "Invalid log entry format"})
			}
			return
		}

		logger.Debug().Interface("new_log", newLog).Msg("New log entry parsed")

		// Validate with messy inline checks
		if newLog.Message == "" {
			logger.Warn().Msg("Log message is empty")
			http.Error(wrapped, "Log message is required", http.StatusBadRequest)
			return
		}

		if newLog.Level == "" {
			logger.Debug().Msg("Log level not specified, defaulting to 'info'")
			newLog.Level = "info" // Default level
		}

		if !isValidLogLevel(newLog.Level) {
			logger.Warn().Str("level", newLog.Level).Msg("Invalid log level")
			http.Error(wrapped, "Invalid log level", http.StatusBadRequest)
			return
		}

		// Generate random ID - not checking for collisions
		newLog.Id = rand.Intn(10000) + 1000
		newLog.Timestamp = time.Now()

		logger.Debug().Int("id", newLog.Id).Time("timestamp", newLog.Timestamp).Msg("Assigned ID and timestamp to new log")

		// Messy: Add extraneous data not part of the struct before saving (won't actually be saved)
		messyInternalData := map[string]interface{}{
			"log_entry":       newLog,
			"received_at":     time.Now(),
			"processing_node": "proc-" + strconv.Itoa(rand.Intn(5)),
			"query_priority":  priority,
		}
		_ = messyInternalData // Avoid unused variable error

		// Add to logs with mutex lock
		mutex.Lock()
		logs = append(logs, newLog)
		logCount := len(logs)
		mutex.Unlock()

		logger.Info().
			Int("id", newLog.Id).
			Str("level", newLog.Level).
			Str("message", newLog.Message).
			Str("source", newLog.Source).
			Int("total_logs", logCount).
			Msg("New log entry created")

		// Messy: Randomly return JSON or plain text confirmation
		if rand.Intn(2) == 0 {
			wrapped.Header().Set("Content-Type", "application/json")
			wrapped.WriteHeader(http.StatusCreated)
			logger.Debug().Str("response_format", "json").Msg("Sending JSON response")
			json.NewEncoder(wrapped).Encode(ApiResponse{
				Success: true,
				Data:    newLog,
			})
		} else {
			wrapped.Header().Set("Content-Type", "text/plain")
			wrapped.WriteHeader(http.StatusCreated)
			logger.Debug().Str("response_format", "plain_text").Msg("Sending plain text response")
			fmt.Fprintf(wrapped, "Log entry %d created successfully with priority %s", newLog.Id, priority)
		}
		return
	} else {
		// Method not allowed
		logger.Warn().Str("method", r.Method).Msg("Method not allowed")
		http.Error(wrapped, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// Poor validation function that should be part of a validation package
func isValidLogLevel(level string) bool {
	validLevels := []string{"debug", "info", "warning", "error", "critical"}
	for _, l := range validLevels {
		if level == l {
			return true
		}
	}
	return false
}

// Handle logs by ID
func handleLogById(w http.ResponseWriter, r *http.Request) {
	// Create wrapped response writer
	wrapped := newResponseWriter(w)

	// Extract ID from path - fragile parsing
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		logger.Warn().Strs("path_parts", parts).Msg("Invalid URL format in handleLogById")
		http.Error(wrapped, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Parse ID
	idStr := parts[3]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logger.Warn().Str("id", idStr).Err(err).Msg("Invalid log ID format")
		http.Error(wrapped, "Invalid log ID", http.StatusBadRequest)
		return
	}

	// Log request
	reqLog := logRequest(r).
		Str("handler", "handleLogById").
		Int("log_id", id)

	// Defer logging response
	defer func() {
		logResponse(wrapped.status, r.Method, r.URL.Path).
			Str("handler", "handleLogById").
			Int("log_id", id).
			Int("content_length", wrapped.contentLength).
			Msg("API response sent")
	}()

	reqLog.Msg("Request received")

	// Find log by ID
	mutex.Lock()
	var foundLog *LogEntry
	var foundIndex = -1 // Need index for potential modification/deletion
	for i, log := range logs {
		if log.Id == id {
			// Create a copy to avoid modifying the original slice directly if needed later
			logCopy := logs[i]
			foundLog = &logCopy
			foundIndex = i
			break
		}
	}
	mutex.Unlock()

	if foundLog == nil {
		logger.Warn().Int("log_id", id).Msg("Log entry not found")
		// Messy: Randomly return 404 as JSON or plain text
		if rand.Intn(2) == 0 {
			http.Error(wrapped, "Log not found", http.StatusNotFound)
		} else {
			wrapped.Header().Set("Content-Type", "application/json")
			wrapped.WriteHeader(http.StatusNotFound)
			json.NewEncoder(wrapped).Encode(ApiResponse{Success: false, Error: "Log not found"})
		}
		return
	}

	logger.Debug().Int("log_id", id).Str("level", foundLog.Level).Msg("Log entry found")

	// Handle different HTTP methods
	if r.Method == "GET" {
		// Messy: Randomly return full JSON or just the message as text
		if rand.Intn(2) == 0 {
			wrapped.Header().Set("Content-Type", "application/json")
			// Messy: Add random correlation ID
			correlationId := fmt.Sprintf("corr-%d", rand.Int63())
			logger.Debug().Str("correlation_id", correlationId).Msg("Generating JSON response")

			response := map[string]interface{}{
				"success":        true,
				"data":           foundLog,
				"correlation_id": correlationId,
			}
			json.NewEncoder(wrapped).Encode(response)
		} else {
			wrapped.Header().Set("Content-Type", "text/plain")
			logger.Debug().Msg("Generating plain text response")
			fmt.Fprintln(wrapped, foundLog.Message)
		}
		return
	} else if r.Method == "DELETE" {
		// Messy: Require confirmation via query param
		confirmation := r.URL.Query().Get("confirm")
		logger.Debug().Str("confirmation", confirmation).Msg("Delete confirmation parameter")

		if confirmation != "true" {
			logger.Warn().Msg("Missing confirmation parameter for delete operation")
			http.Error(wrapped, "Missing confirmation parameter 'confirm=true'", http.StatusBadRequest)
			return
		}

		// Delete log by ID
		mutex.Lock()
		if foundIndex != -1 && foundIndex < len(logs) && logs[foundIndex].Id == id { // Double check index validity
			// Inefficient way to remove an element from a slice
			logs = append(logs[:foundIndex], logs[foundIndex+1:]...)
			logger.Info().Int("log_id", id).Int("logs_remaining", len(logs)).Msg("Log entry deleted")
		} else {
			// Log might have been deleted between find and delete lock, handle messily
			mutex.Unlock()
			logger.Warn().Int("log_id", id).Msg("Log possibly deleted concurrently")
			http.Error(wrapped, "Log possibly deleted concurrently", http.StatusConflict)
			return
		}
		mutex.Unlock()

		// Messy: Randomly return JSON or plain text
		if rand.Intn(2) == 0 {
			wrapped.Header().Set("Content-Type", "application/json")
			logger.Debug().Str("response_format", "json").Msg("Sending JSON delete confirmation")
			json.NewEncoder(wrapped).Encode(ApiResponse{
				Success: true,
				Data:    "Log deleted successfully",
			})
		} else {
			wrapped.Header().Set("Content-Type", "text/plain")
			logger.Debug().Str("response_format", "plain_text").Msg("Sending plain text delete confirmation")
			fmt.Fprintln(wrapped, "DELETED")
		}
		return
	} else if r.Method == "PUT" {
		// Update log
		body, err := io.ReadAll(r.Body)
		if err != nil {
			logger.Error().Err(err).Msg("Failed to read request body")
			http.Error(wrapped, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		logger.Debug().Int("body_size", len(body)).Msg("Request body read")

		// Parse as a log entry
		var updatedLog LogEntry
		err = json.Unmarshal(body, &updatedLog)
		if err != nil {
			logger.Error().Err(err).Str("body", string(body)).Msg("Failed to parse JSON body")
			http.Error(wrapped, "Invalid log entry format", http.StatusBadRequest)
			return
		}

		logger.Debug().Interface("updated_log", updatedLog).Msg("Updated log parsed from request")

		// Keep original ID and timestamp (already messy)
		updatedLog.Id = id
		updatedLog.Timestamp = foundLog.Timestamp // Use timestamp from the found log

		// Update log
		mutex.Lock()
		if foundIndex != -1 && foundIndex < len(logs) && logs[foundIndex].Id == id { // Double check index validity
			logs[foundIndex] = updatedLog
			logger.Info().
				Int("log_id", id).
				Str("level", updatedLog.Level).
				Str("message", updatedLog.Message).
				Str("source", updatedLog.Source).
				Msg("Log entry updated")
		} else {
			// Log might have been deleted between find and update lock
			mutex.Unlock()
			logger.Warn().Int("log_id", id).Msg("Log possibly deleted concurrently during update")
			http.Error(wrapped, "Log possibly deleted concurrently", http.StatusConflict)
			return
		}
		mutex.Unlock()

		wrapped.Header().Set("Content-Type", "application/json")
		// Messy: Add extraneous field on successful update
		updateCycle := rand.Intn(100)
		logger.Debug().Int("update_cycle", updateCycle).Msg("Generating update response")

		response := map[string]interface{}{
			"success":      true,
			"data":         updatedLog,
			"update_cycle": updateCycle,
		}
		json.NewEncoder(wrapped).Encode(response)
		return
	} else {
		// Method not allowed
		logger.Warn().Str("method", r.Method).Int("log_id", id).Msg("Method not allowed")
		http.Error(wrapped, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// ResponseWriter wrapper to capture status code and content length
type responseWriter struct {
	http.ResponseWriter
	status        int
	contentLength int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK, 0}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.contentLength += n
	return n, err
}

// Handle users - similar pattern to logs but with different structure
func handleUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// Create a slice of users from the map
		mutex.Lock()
		usersSlice := make([]User, 0, len(users))
		for _, user := range users {
			// Hide API keys in the list
			userCopy := user
			userCopy.ApiKeys = nil
			usersSlice = append(usersSlice, userCopy)
		}
		mutex.Unlock()

		// Messy: Randomly return JSON or XML
		if rand.Intn(2) == 0 {
			w.Header().Set("Content-Type", "application/json")
			// Messy: Add extraneous field
			response := map[string]interface{}{
				"success":     true,
				"data":        usersSlice,
				"server_node": "user-node-" + strconv.Itoa(rand.Intn(3)),
			}
			json.NewEncoder(w).Encode(response)
		} else {
			// Messy: Return XML
			type UsersResponse struct {
				XMLName xml.Name `xml:"users"`
				Success bool     `xml:"success,attr"`
				Users   []User   `xml:"user"`
				Node    string   `xml:"server_node,attr"` // Extraneous attribute
			}
			w.Header().Set("Content-Type", "application/xml")
			xmlData := UsersResponse{
				Success: true,
				Users:   usersSlice,
				Node:    "user-node-" + strconv.Itoa(rand.Intn(3)),
			}
			xml.NewEncoder(w).Encode(xmlData)
		}
		return
	} else if r.Method == "POST" {
		// Messy: Expect 'department' from query string
		department := r.URL.Query().Get("department")
		if department == "" {
			department = "unassigned"
		}

		// Read request body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse as a user
		var newUser User
		err = json.Unmarshal(body, &newUser)
		if err != nil {
			http.Error(w, "Invalid user format", http.StatusBadRequest)
			return
		}

		// Validate user data
		if newUser.Username == "" {
			http.Error(w, "Username is required", http.StatusBadRequest)
			return
		}

		if newUser.Email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		// Check if user already exists
		mutex.Lock()
		_, exists := users[newUser.Username]
		if exists {
			mutex.Unlock()
			// Messy: Random conflict response format
			if rand.Intn(2) == 0 {
				http.Error(w, "User already exists", http.StatusConflict)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "duplicate user", "username": newUser.Username})
			}
			return
		}

		// Set creation time
		newUser.CreatedAt = time.Now().Format(time.RFC3339)
		newUser.LastLogin = newUser.CreatedAt

		// Generate an API key
		if newUser.ApiKeys == nil {
			newUser.ApiKeys = []string{}
		}
		apiKey := fmt.Sprintf("%s-key-%d", newUser.Username, rand.Intn(1000000))
		newUser.ApiKeys = append(newUser.ApiKeys, apiKey)

		// Messy: Add extraneous internal data (won't be saved)
		internalUserData := map[string]interface{}{
			"user":         newUser,
			"department":   department,
			"requested_by": r.Header.Get("X-Requested-By"), // Another messy dependency
		}
		_ = internalUserData

		// Add user to map
		users[newUser.Username] = newUser
		mutex.Unlock()

		// Messy: Randomly return JSON or plain text confirmation
		if rand.Intn(2) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(ApiResponse{
				Success: true,
				Data:    newUser,
			})
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusCreated)
			fmt.Fprintf(w, "User created: %s in department %s", newUser.Username, department)
		}
		return
	} else {
		// Method not allowed
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// Handle users by username
func handleUserByName(w http.ResponseWriter, r *http.Request) {
	// Extract username from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 4 {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Get username
	username := parts[3]

	// Find user by username
	mutex.Lock()
	user, exists := users[username]
	mutex.Unlock()

	if !exists {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Handle different HTTP methods
	if r.Method == "GET" {
		// Messy: Randomly return JSON or YAML string
		if rand.Intn(2) == 0 {
			w.Header().Set("Content-Type", "application/json")
			// Messy: Add extraneous field
			response := map[string]interface{}{
				"success":       true,
				"data":          user,
				"fetch_time_ms": rand.Intn(50),
			}
			json.NewEncoder(w).Encode(response)
		} else {
			w.Header().Set("Content-Type", "application/x-yaml")
			// Messy: Return YAML as a string
			yamlBytes, err := yaml.Marshal(user)
			if err != nil {
				http.Error(w, "Failed to marshal user to YAML", http.StatusInternalServerError)
				return
			}
			w.Write(yamlBytes)
		}
		return
	} else if r.Method == "DELETE" {
		// Delete user
		mutex.Lock()
		delete(users, username)
		mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiResponse{
			Success: true,
			Data:    "User deleted successfully",
		})
		return
	} else if r.Method == "PUT" {
		// Update user
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse as a user
		var updatedUser User
		err = json.Unmarshal(body, &updatedUser)
		if err != nil {
			http.Error(w, "Invalid user format", http.StatusBadRequest)
			return
		}

		// Keep original username and creation time (already messy)
		updatedUser.Username = username
		updatedUser.CreatedAt = user.CreatedAt // Use original creation time

		// Keep API keys if not provided
		if updatedUser.ApiKeys == nil || len(updatedUser.ApiKeys) == 0 {
			updatedUser.ApiKeys = user.ApiKeys
		}

		// Update user
		mutex.Lock()
		users[username] = updatedUser
		mutex.Unlock()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiResponse{
			Success: true,
			Data:    updatedUser,
		})
		return
	} else {
		// Method not allowed
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}

// Handle metrics
func handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mutex.Lock()
	metricsCopy := make(map[string]int)
	for k, v := range metrics {
		metricsCopy[k] = v
	}
	// Messy: Add an extra, unrelated metric sometimes
	if rand.Intn(3) == 0 {
		metricsCopy["system_load_avg"] = rand.Intn(100)
	}
	mutex.Unlock()

	// Messy: Sometimes return as simple key=value pairs text
	if rand.Intn(2) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ApiResponse{
			Success: true,
			Data:    metricsCopy,
		})
	} else {
		w.Header().Set("Content-Type", "text/plain")
		for k, v := range metricsCopy {
			fmt.Fprintf(w, "%s=%d", k, v)
		}
	}
}

// Handle config
func handleConfig(w http.ResponseWriter, r *http.Request) {
	// Check method
	if r.Method == "GET" {
		// Base config data
		configData := map[string]interface{}{
			"port":       port,
			"debug_mode": debugMode,
		}

		// Try to read from file if it exists (already messy)
		if _, err := os.Stat(configFile); err == nil {
			data, err := ioutil.ReadFile(configFile)
			if err == nil {
				var fileConfig map[string]interface{}
				// Messy: Ignore JSON parsing errors silently
				_ = json.Unmarshal(data, &fileConfig)
				if fileConfig != nil {
					for k, v := range fileConfig {
						configData[k] = v
					}
				}
			}
		}

		// Messy: Randomly return JSON or INI-style text
		if rand.Intn(2) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ApiResponse{
				Success: true,
				Data:    configData,
			})
		} else {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintln(w, "[server]")
			for k, v := range configData {
				fmt.Fprintf(w, "%s = %v ", k, v)
			}
		}
		return
	} else if r.Method == "POST" {
		// Messy: Read some data from JSON body, some from query, some from form
		// Query parameter
		apiKey := r.URL.Query().Get("api_key") // API Key from query

		// Form data (pretend parse) - need to call ParseForm first
		r.ParseForm()
		timeoutStr := r.FormValue("timeout") // Timeout from form data
		timeout := 30                        // Default timeout
		if t, err := strconv.Atoi(timeoutStr); err == nil {
			timeout = t
		}

		// JSON body
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Parse JSON body for other config values
		var configUpdate map[string]interface{}
		err = json.Unmarshal(body, &configUpdate)
		if err != nil && len(body) > 0 { // Only error if body wasn't empty
			http.Error(w, "Invalid config JSON format", http.StatusBadRequest)
			return
		}
		if configUpdate == nil {
			configUpdate = make(map[string]interface{}) // Ensure map exists
		}

		// Messy Update Logic: Apply updates from all sources
		configDataToWrite := make(map[string]interface{})

		// Apply JSON body updates first
		for k, v := range configUpdate {
			configDataToWrite[k] = v
		}
		// Apply form data updates
		configDataToWrite["timeout"] = timeout
		// Apply query param updates
		if apiKey != "" {
			configDataToWrite["api_key_from_query"] = apiKey // Store with different name to show source
		}

		// Update global vars (only port and debugMode supported here messily)
		if portValue, ok := configDataToWrite["port"]; ok {
			if portFloat, ok := portValue.(float64); ok { // JSON numbers are float64
				port = int(portFloat)
			}
		}
		if debugValue, ok := configDataToWrite["debug_mode"]; ok {
			if debugBool, ok := debugValue.(bool); ok {
				debugMode = debugBool
			}
		}

		// Write combined messy config to file
		configJson, _ := json.MarshalIndent(configDataToWrite, "", "  ")
		ioutil.WriteFile(configFile, configJson, 0644) // Ignoring potential write error

		// Messy: Random response format
		if rand.Intn(2) == 0 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(ApiResponse{
				Success: true,
				Data:    "Config updated partially", // Messy message
			})
		} else {
			w.Header().Set("Content-Type", "text/plain")
			fmt.Fprintf(w, "Config updated with timeout=%d and api_key=%s", timeout, apiKey)
		}
		return
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
}
