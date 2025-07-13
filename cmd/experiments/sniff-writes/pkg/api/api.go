package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/database"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/export"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
)

// Server represents the API server
type Server struct {
	db     *database.SQLiteDB
	config *models.Config
	mux    *mux.Router
	port   int
}

// NewServer creates a new API server
func NewServer(db *database.SQLiteDB, config *models.Config, port int) *Server {
	server := &Server{
		db:     db,
		config: config,
		mux:    mux.NewRouter(),
		port:   port,
	}
	server.setupRoutes()
	return server
}

// EventsResponse represents the response for events endpoint
type EventsResponse struct {
	Events     []models.EventOutput `json:"events"`
	Total      int                  `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// StatsResponse represents the response for stats endpoint
type StatsResponse struct {
	TotalEvents     int            `json:"total_events"`
	OperationCounts map[string]int `json:"operation_counts"`
	ProcessCounts   map[string]int `json:"process_counts"`
	DateRange       *DateRange     `json:"date_range,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error      string      `json:"error"`
	Message    string      `json:"message"`
	StatusCode int         `json:"status_code"`
	Details    interface{} `json:"details,omitempty"`
}

func (s *Server) setupRoutes() {
	// Add logging middleware and panic recovery
	s.mux.Use(s.loggingMiddleware)
	s.mux.Use(s.panicRecoveryMiddleware)

	s.mux.HandleFunc("/events", s.HandleEvents).Methods("GET")
	s.mux.HandleFunc("/events/export", s.HandleExport).Methods("GET")
	s.mux.HandleFunc("/stats", s.HandleStats).Methods("GET")
	s.mux.HandleFunc("/health", s.HandleHealth).Methods("GET")
	s.mux.HandleFunc("/api/status", s.HandleStatus).Methods("GET")
}

// loggingMiddleware logs all HTTP requests and responses
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Debug().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Interface("headers", r.Header).
			Msg("HTTP request started")

		// Wrap the ResponseWriter to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: 200}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Info().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Str("remote_addr", r.RemoteAddr).
			Int("status_code", wrapped.statusCode).
			Dur("duration", duration).
			Msg("HTTP request completed")
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// panicRecoveryMiddleware recovers from panics and logs them
func (s *Server) panicRecoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Error().
					Interface("panic", err).
					Str("method", r.Method).
					Str("url", r.URL.String()).
					Str("remote_addr", r.RemoteAddr).
					Msg("HTTP handler panicked")

				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Start starts the API server
func (s *Server) Start() error {
	fmt.Printf("Starting API server on port %d\n", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux)
}

// sendErrorResponse sends a structured JSON error response
func (s *Server) sendErrorResponse(w http.ResponseWriter, statusCode int, errorType string, message string, details interface{}) {
	errorResp := ErrorResponse{
		Error:      errorType,
		Message:    message,
		StatusCode: statusCode,
		Details:    details,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.Error().Err(err).Msg("Failed to encode error response")
	}
}

// checkDatabaseAvailable checks if database is available and sends error if not
func (s *Server) checkDatabaseAvailable(w http.ResponseWriter, r *http.Request) bool {
	if s.db == nil {
		log.Error().
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Database not available for request")

		s.sendErrorResponse(w, http.StatusServiceUnavailable,
			"database_unavailable",
			"Database is not available. Please ensure the server was started with a valid database configuration.",
			map[string]string{
				"suggestion": "The server needs to be started with the --sqlite flag pointing to a valid database file",
			})
		return false
	}
	return true
}

func (s *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Str("remote_addr", r.RemoteAddr).
		Str("query_params", r.URL.RawQuery).
		Msg("Handling events request")

	// Check database availability
	if !s.checkDatabaseAvailable(w, r) {
		return
	}

	filter, err := s.parseQueryFilter(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("query_params", r.URL.RawQuery).
			Msg("Failed to parse query filter")
		s.sendErrorResponse(w, http.StatusBadRequest,
			"invalid_parameters",
			"Invalid query parameters provided",
			map[string]interface{}{
				"error_details": err.Error(),
				"query_params":  r.URL.RawQuery,
			})
		return
	}

	log.Debug().Interface("filter", filter).Msg("Parsed query filter")

	// Get total count
	log.Debug().Msg("Starting database count query")
	total, err := s.db.CountEvents(filter)
	if err != nil {
		log.Error().
			Err(err).
			Interface("filter", filter).
			Msg("Failed to count events in database")
		s.sendErrorResponse(w, http.StatusInternalServerError,
			"database_error",
			"Failed to count events from database",
			map[string]interface{}{
				"error_details": err.Error(),
				"operation":     "count_events",
				"filter":        filter,
			})
		return
	}

	log.Debug().Int("total_count", total).Msg("Retrieved total event count")

	// Get events
	log.Debug().Msg("Starting database events query")
	events, err := s.db.QueryEvents(filter)
	if err != nil {
		log.Error().
			Err(err).
			Interface("filter", filter).
			Msg("Failed to query events from database")
		s.sendErrorResponse(w, http.StatusInternalServerError,
			"database_error",
			"Failed to query events from database",
			map[string]interface{}{
				"error_details": err.Error(),
				"operation":     "query_events",
				"filter":        filter,
			})
		return
	}

	log.Debug().Int("events_count", len(events)).Msg("Retrieved events")

	// Calculate pagination info
	page := 1
	if filter.Offset > 0 && filter.Limit > 0 {
		page = (filter.Offset / filter.Limit) + 1
	}

	pageSize := filter.Limit
	if pageSize == 0 {
		pageSize = len(events)
	}

	totalPages := 1
	if filter.Limit > 0 {
		totalPages = (total + filter.Limit - 1) / filter.Limit
	}

	response := EventsResponse{
		Events:     events,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}

	w.Header().Set("Content-Type", "application/json")
	log.Debug().
		Int("page", response.Page).
		Int("page_size", response.PageSize).
		Int("total_pages", response.TotalPages).
		Int("total", response.Total).
		Msg("Sending events response")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleExport(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Str("remote_addr", r.RemoteAddr).
		Str("query_params", r.URL.RawQuery).
		Msg("Handling export request")

	// Check database availability
	if !s.checkDatabaseAvailable(w, r) {
		return
	}

	filter, err := s.parseQueryFilter(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("query_params", r.URL.RawQuery).
			Msg("Failed to parse query filter for export")
		s.sendErrorResponse(w, http.StatusBadRequest,
			"invalid_parameters",
			"Invalid query parameters for export",
			map[string]interface{}{
				"error_details": err.Error(),
				"query_params":  r.URL.RawQuery,
			})
		return
	}

	// Remove pagination for export
	filter.Limit = 0
	filter.Offset = 0

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	log.Debug().Str("export_format", format).Interface("filter", filter).Msg("Preparing export")

	var exportFormat export.ExportFormat
	switch format {
	case "json":
		exportFormat = export.FormatJSON
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", "attachment; filename=events.json")
	case "csv":
		exportFormat = export.FormatCSV
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=events.csv")
	case "markdown":
		exportFormat = export.FormatMarkdown
		w.Header().Set("Content-Type", "text/markdown")
		w.Header().Set("Content-Disposition", "attachment; filename=events.md")
	default:
		http.Error(w, "Unsupported format. Use: json, csv, markdown", http.StatusBadRequest)
		return
	}

	events, err := s.db.QueryEvents(filter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to query events for export")
		http.Error(w, fmt.Sprintf("Failed to query events: %v", err), http.StatusInternalServerError)
		return
	}

	log.Debug().Int("events_count", len(events)).Str("format", format).Msg("Starting export")

	exporter := export.New(w, exportFormat)
	if err := exporter.Export(events); err != nil {
		log.Error().Err(err).Str("format", format).Msg("Failed to export events")
		http.Error(w, fmt.Sprintf("Failed to export events: %v", err), http.StatusInternalServerError)
		return
	}

	log.Debug().Int("events_exported", len(events)).Str("format", format).Msg("Export completed successfully")
}

func (s *Server) HandleStats(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("method", r.Method).
		Str("url", r.URL.String()).
		Str("remote_addr", r.RemoteAddr).
		Msg("Handling stats request")

	// Check database availability
	if !s.checkDatabaseAvailable(w, r) {
		return
	}

	// TODO: Implement stats functionality
	// For now, return basic info
	filter := database.QueryFilter{}
	log.Debug().Msg("Starting database count query for stats")
	total, err := s.db.CountEvents(filter)
	if err != nil {
		log.Error().
			Err(err).
			Interface("filter", filter).
			Msg("Failed to count events for stats")
		s.sendErrorResponse(w, http.StatusInternalServerError,
			"database_error",
			"Failed to count events for statistics",
			map[string]interface{}{
				"error_details": err.Error(),
				"operation":     "count_events_for_stats",
			})
		return
	}

	log.Debug().Int("total_events", total).Msg("Retrieved stats")

	response := StatsResponse{
		TotalEvents:     total,
		OperationCounts: map[string]int{},
		ProcessCounts:   map[string]int{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("method", r.Method).
		Str("remote_addr", r.RemoteAddr).
		Msg("Handling health check request")

	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Debug().Msg("Health check completed")
}

// StatusResponse represents the system status
type StatusResponse struct {
	Status         string   `json:"status"`
	DatabaseStatus string   `json:"database_status"`
	DatabaseError  string   `json:"database_error,omitempty"`
	CanSearch      bool     `json:"can_search"`
	Message        string   `json:"message"`
	Operations     []string `json:"operations"`
	Directory      string   `json:"directory"`
	Timestamp      string   `json:"timestamp"`
}

func (s *Server) HandleStatus(w http.ResponseWriter, r *http.Request) {
	log.Debug().
		Str("method", r.Method).
		Str("remote_addr", r.RemoteAddr).
		Msg("Handling status request")

	response := StatusResponse{
		Status:    "ok",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Include configuration information
	if s.config != nil {
		response.Operations = s.config.Operations
		response.Directory = s.config.Directory
	}

	// Check database status
	if s.db == nil {
		response.DatabaseStatus = "unavailable"
		response.DatabaseError = "Database not initialized"
		response.CanSearch = false
		response.Message = "Database is not available. History search is disabled."
	} else {
		// Try a simple query to test database connectivity
		filter := database.QueryFilter{Limit: 1}
		_, err := s.db.CountEvents(filter)
		if err != nil {
			response.DatabaseStatus = "error"
			response.DatabaseError = err.Error()
			response.CanSearch = false
			response.Message = "Database is connected but has errors. History search may not work properly."
		} else {
			response.DatabaseStatus = "connected"
			response.CanSearch = true
			response.Message = "All systems operational. History search is available."
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	log.Debug().
		Str("database_status", response.DatabaseStatus).
		Bool("can_search", response.CanSearch).
		Msg("Status check completed")
}

func (s *Server) parseQueryFilter(r *http.Request) (database.QueryFilter, error) {
	filter := database.QueryFilter{}

	log.Debug().
		Str("query_string", r.URL.RawQuery).
		Interface("query_params", r.URL.Query()).
		Msg("Parsing query filter")

	// Parse time filters
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		log.Debug().Str("start_time", startStr).Msg("Parsing start_time parameter")
		startTime, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			log.Error().
				Err(err).
				Str("start_time", startStr).
				Msg("Invalid start_time format")
			return filter, fmt.Errorf("invalid start_time format (use RFC3339): %w", err)
		}
		filter.StartTime = &startTime
		log.Debug().Time("parsed_start_time", startTime).Msg("Successfully parsed start_time")
	}

	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		log.Debug().Str("end_time", endStr).Msg("Parsing end_time parameter")
		endTime, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			log.Error().
				Err(err).
				Str("end_time", endStr).
				Msg("Invalid end_time format")
			return filter, fmt.Errorf("invalid end_time format (use RFC3339): %w", err)
		}
		filter.EndTime = &endTime
		log.Debug().Time("parsed_end_time", endTime).Msg("Successfully parsed end_time")
	}

	// Parse other filters
	filter.ProcessFilter = r.URL.Query().Get("process")
	filter.FilenamePattern = r.URL.Query().Get("filename")

	log.Debug().
		Str("process_filter", filter.ProcessFilter).
		Str("filename_pattern", filter.FilenamePattern).
		Msg("Parsed text filters")

	if operations := r.URL.Query().Get("operations"); operations != "" {
		filter.OperationFilter = strings.Split(operations, ",")
		log.Debug().Strs("operations", filter.OperationFilter).Msg("Parsed operations filter")
	}

	if pidStr := r.URL.Query().Get("pid"); pidStr != "" {
		log.Debug().Str("pid_str", pidStr).Msg("Parsing PID parameter")
		pid, err := strconv.ParseUint(pidStr, 10, 32)
		if err != nil {
			log.Error().
				Err(err).
				Str("pid_str", pidStr).
				Msg("Invalid PID format")
			return filter, fmt.Errorf("invalid pid: %w", err)
		}
		pidVal := uint32(pid)
		filter.PID = &pidVal
		log.Debug().Uint32("parsed_pid", pidVal).Msg("Successfully parsed PID")
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		log.Debug().Str("limit_str", limitStr).Msg("Parsing limit parameter")
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			log.Error().
				Err(err).
				Str("limit_str", limitStr).
				Msg("Invalid limit format")
			return filter, fmt.Errorf("invalid limit: %w", err)
		}
		filter.Limit = limit
		log.Debug().Int("parsed_limit", limit).Msg("Successfully parsed limit")
	} else {
		filter.Limit = 100 // Default limit
		log.Debug().Int("default_limit", filter.Limit).Msg("Using default limit")
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		log.Debug().Str("offset_str", offsetStr).Msg("Parsing offset parameter")
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			log.Error().
				Err(err).
				Str("offset_str", offsetStr).
				Msg("Invalid offset format")
			return filter, fmt.Errorf("invalid offset: %w", err)
		}
		filter.Offset = offset
		log.Debug().Int("parsed_offset", offset).Msg("Successfully parsed offset")
	}

	log.Debug().Interface("final_filter", filter).Msg("Query filter parsing completed successfully")
	return filter, nil
}
