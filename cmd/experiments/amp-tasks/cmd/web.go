package cmd

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

//go:embed static/*
var staticFiles embed.FS

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Start web server for live dashboard and reports",
	Long: `Start a web server that provides live visualization of project data.
	
The web interface includes:
- Live dashboard with project overview
- Detailed project reports
- Agent and task management views
- Auto-refreshing data

Examples:
  amp-tasks web                    # Start server on port 8080
  amp-tasks web --port 3000       # Start on custom port
  amp-tasks web --host 0.0.0.0    # Bind to all interfaces`,
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetInt("port")
		host, _ := cmd.Flags().GetString("host")
		autoRefresh, _ := cmd.Flags().GetBool("auto-refresh")
		refreshInterval, _ := cmd.Flags().GetInt("refresh-interval")

		return startWebServer(host, port, autoRefresh, refreshInterval)
	},
}

func startWebServer(host string, port int, autoRefresh bool, refreshInterval int) error {
	dbPath, _ := rootCmd.PersistentFlags().GetString("db")
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Initialize task manager
	tm, err := NewTaskManager(dbPath, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize task manager: %w", err)
	}
	defer tm.Close()

	// Create web handler
	handler := &WebHandler{
		tm:              tm,
		logger:          logger,
		autoRefresh:     autoRefresh,
		refreshInterval: refreshInterval,
	}

	// Setup routes
	mux := http.NewServeMux()
	
	// Static files
	mux.Handle("/static/", http.FileServer(http.FS(staticFiles)))
	
	// Main routes
	mux.HandleFunc("/", handler.dashboardHandler)
	mux.HandleFunc("/dashboard", handler.dashboardHandler)
	mux.HandleFunc("/report", handler.reportHandler)
	mux.HandleFunc("/api/dashboard", handler.apiDashboardHandler)
	mux.HandleFunc("/api/report", handler.apiReportHandler)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Setup graceful shutdown
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in goroutine
	go func() {
		logger.Info().
			Str("host", host).
			Int("port", port).
			Bool("auto_refresh", autoRefresh).
			Int("refresh_interval", refreshInterval).
			Msg("Starting web server")
		
		fmt.Printf("🌐 Web server starting at http://%s:%d\n", host, port)
		fmt.Printf("📊 Dashboard: http://%s:%d/dashboard\n", host, port)
		fmt.Printf("📋 Reports: http://%s:%d/report\n", host, port)
		fmt.Printf("Press Ctrl+C to stop\n\n")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	logger.Info().Msg("Shutdown signal received")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
		return err
	}

	logger.Info().Msg("Server exited gracefully")
	return nil
}

type WebHandler struct {
	tm              *TaskManager
	logger          zerolog.Logger
	autoRefresh     bool
	refreshInterval int
}

func (h *WebHandler) dashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get current project
	project, err := h.tm.GetDefaultProject()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get project: %v", err), http.StatusInternalServerError)
		return
	}

	// Gather dashboard data
	data, err := gatherDashboardData(h.tm, project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to gather dashboard data: %v", err), http.StatusInternalServerError)
		return
	}

	// Render dashboard template
	component := DashboardPage(data, h.autoRefresh, h.refreshInterval)
	component.Render(r.Context(), w)
}

func (h *WebHandler) reportHandler(w http.ResponseWriter, r *http.Request) {
	verbose := r.URL.Query().Get("verbose") == "true"
	summary := r.URL.Query().Get("summary") == "true"

	// Generate report
	report, err := generateReport(h.tm, verbose, summary)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate report: %v", err), http.StatusInternalServerError)
		return
	}

	// Render report template
	component := ReportPage(report, h.autoRefresh, h.refreshInterval)
	component.Render(r.Context(), w)
}

func (h *WebHandler) apiDashboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// Get current project
	project, err := h.tm.GetDefaultProject()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get project: %v", err), http.StatusInternalServerError)
		return
	}

	// Gather dashboard data
	data, err := gatherDashboardData(h.tm, project)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to gather dashboard data: %v", err), http.StatusInternalServerError)
		return
	}

	// Render partial dashboard content
	component := DashboardContent(data)
	component.Render(r.Context(), w)
}

func (h *WebHandler) apiReportHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	verbose := r.URL.Query().Get("verbose") == "true"
	summary := r.URL.Query().Get("summary") == "true"

	// Generate report
	report, err := generateReport(h.tm, verbose, summary)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate report: %v", err), http.StatusInternalServerError)
		return
	}

	// Render partial report content
	component := ReportContent(report)
	component.Render(r.Context(), w)
}

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().IntP("port", "p", 8080, "Port to bind the web server")
	webCmd.Flags().StringP("host", "H", "localhost", "Host to bind the web server")
	webCmd.Flags().Bool("auto-refresh", true, "Enable auto-refresh of data")
	webCmd.Flags().Int("refresh-interval", 30, "Auto-refresh interval in seconds")
}
