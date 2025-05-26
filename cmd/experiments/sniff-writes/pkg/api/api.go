package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/database"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/export"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
)

// Server represents the API server
type Server struct {
	db   *database.SQLiteDB
	mux  *mux.Router
	port int
}

// NewServer creates a new API server
func NewServer(db *database.SQLiteDB, port int) *Server {
	server := &Server{
		db:   db,
		mux:  mux.NewRouter(),
		port: port,
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
	TotalEvents      int               `json:"total_events"`
	OperationCounts  map[string]int    `json:"operation_counts"`
	ProcessCounts    map[string]int    `json:"process_counts"`
	DateRange        *DateRange        `json:"date_range,omitempty"`
}

// DateRange represents a date range
type DateRange struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

func (s *Server) setupRoutes() {
	s.mux.HandleFunc("/events", s.HandleEvents).Methods("GET")
	s.mux.HandleFunc("/events/export", s.HandleExport).Methods("GET")
	s.mux.HandleFunc("/stats", s.HandleStats).Methods("GET")
	s.mux.HandleFunc("/health", s.HandleHealth).Methods("GET")
}

// Start starts the API server
func (s *Server) Start() error {
	fmt.Printf("Starting API server on port %d\n", s.port)
	return http.ListenAndServe(fmt.Sprintf(":%d", s.port), s.mux)
}

func (s *Server) HandleEvents(w http.ResponseWriter, r *http.Request) {
	filter, err := s.parseQueryFilter(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid query parameters: %v", err), http.StatusBadRequest)
		return
	}

	// Get total count
	total, err := s.db.CountEvents(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count events: %v", err), http.StatusInternalServerError)
		return
	}

	// Get events
	events, err := s.db.QueryEvents(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query events: %v", err), http.StatusInternalServerError)
		return
	}

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
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleExport(w http.ResponseWriter, r *http.Request) {
	filter, err := s.parseQueryFilter(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid query parameters: %v", err), http.StatusBadRequest)
		return
	}

	// Remove pagination for export
	filter.Limit = 0
	filter.Offset = 0

	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

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
		http.Error(w, fmt.Sprintf("Failed to query events: %v", err), http.StatusInternalServerError)
		return
	}

	exporter := export.New(w, exportFormat)
	if err := exporter.Export(events); err != nil {
		http.Error(w, fmt.Sprintf("Failed to export events: %v", err), http.StatusInternalServerError)
		return
	}
}

func (s *Server) HandleStats(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement stats functionality
	// For now, return basic info
	filter := database.QueryFilter{}
	total, err := s.db.CountEvents(filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to count events: %v", err), http.StatusInternalServerError)
		return
	}

	response := StatsResponse{
		TotalEvents: total,
		OperationCounts: map[string]int{},
		ProcessCounts: map[string]int{},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) HandleHealth(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"time":   time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) parseQueryFilter(r *http.Request) (database.QueryFilter, error) {
	filter := database.QueryFilter{}

	// Parse time filters
	if startStr := r.URL.Query().Get("start_time"); startStr != "" {
		startTime, err := time.Parse(time.RFC3339, startStr)
		if err != nil {
			return filter, fmt.Errorf("invalid start_time format (use RFC3339): %w", err)
		}
		filter.StartTime = &startTime
	}

	if endStr := r.URL.Query().Get("end_time"); endStr != "" {
		endTime, err := time.Parse(time.RFC3339, endStr)
		if err != nil {
			return filter, fmt.Errorf("invalid end_time format (use RFC3339): %w", err)
		}
		filter.EndTime = &endTime
	}

	// Parse other filters
	filter.ProcessFilter = r.URL.Query().Get("process")
	filter.FilenamePattern = r.URL.Query().Get("filename")

	if operations := r.URL.Query().Get("operations"); operations != "" {
		filter.OperationFilter = strings.Split(operations, ",")
	}

	if pidStr := r.URL.Query().Get("pid"); pidStr != "" {
		pid, err := strconv.ParseUint(pidStr, 10, 32)
		if err != nil {
			return filter, fmt.Errorf("invalid pid: %w", err)
		}
		pidVal := uint32(pid)
		filter.PID = &pidVal
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return filter, fmt.Errorf("invalid limit: %w", err)
		}
		filter.Limit = limit
	} else {
		filter.Limit = 100 // Default limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return filter, fmt.Errorf("invalid offset: %w", err)
		}
		filter.Offset = offset
	}

	return filter, nil
}