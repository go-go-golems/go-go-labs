package main

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/api"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/cache"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/database"
	ebpfops "github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/ebpf"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/export"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/filecache"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/formatter"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/models"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/processor"
	"github.com/go-go-golems/go-go-labs/cmd/experiments/sniff-writes/pkg/web"
)

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -target amd64 -cflags "-I/usr/include/x86_64-linux-gnu -mllvm -bpf-stack-size=8192" sniffwrites sniff_writes.c

//go:embed static/style.css
var styleCSS []byte

//go:embed static/app.js
var appJS []byte

var config models.Config
var sqliteDB *database.SQLiteDB
var webHub *web.WebHub
var pathCache *cache.PathCache
var fileCache *filecache.FileCache

// Query-specific flags
var queryFlags struct {
	startTime string
	endTime   string
	filename  string
	pid       uint32
	exportFmt string
	limit     int
	offset    int
}

// Server-specific flags
var serverFlags struct {
	port int
}

func initSQLite() error {
	var err error
	sqliteDB, err = database.NewSQLiteDB(config.SqliteDB)
	return err
}

func closeSQLite() {
	if sqliteDB != nil {
		sqliteDB.Close()
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	index().Render(r.Context(), w)
}

func startWebServer() {
	webHub = web.StartServer(config.WebPort, styleCSS, appJS, handleIndex, sqliteDB, &config)
}

func main() {
	// Initialize logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if err := rootCmd.Execute(); err != nil {
		zlog.Fatal().Err(err).Msg("Error executing command")
	}
}

var rootCmd = &cobra.Command{
	Use:   "sniff-writes",
	Short: "Monitor file operations using eBPF",
	Long: `sniff-writes monitors file operations on specified directories using eBPF 
tracepoints. By default, monitors open, write, and close operations (read operations 
are excluded by default as they can be very noisy). Only shows regular files and 
excludes pipes, sockets, and virtual filesystems. Supports capturing full write/read 
content with automatic chunking for large I/O operations, and tracks file offsets.

Examples:
  # Monitor default directory with plain output (open, write, close only)
  sudo sniff-writes monitor

  # Monitor specific directory with JSON output for 30 seconds
  sudo sniff-writes monitor -d /var/log -f json -t 30s

  # Monitor only write operations with table format
  sudo sniff-writes monitor -o write -f table --show-fd
  
  # Include read operations (can be noisy)
  sudo sniff-writes monitor -o open,read,write,close

  # Show all file types including pipes and sockets
  sudo sniff-writes monitor --show-all-files

  # Capture write content (first 4096 bytes)
  sudo sniff-writes monitor --capture-content

  # Capture more content (first 8192 bytes with chunking) 
  sudo sniff-writes monitor --capture-content --content-size 8192
  
  # Show colored diffs when writes occur (requires content capture)
  sudo sniff-writes monitor --capture-content --show-diffs
  
  # Show diffs without colors (for scripting/logging)
  sudo sniff-writes monitor --capture-content --show-diffs --no-color
  
  # Show compact diffs (only 1 line of context around changes)
  sudo sniff-writes monitor --capture-content --show-diffs --diff-context 1
  
  # Show only changed lines (no context)
  sudo sniff-writes monitor --capture-content --show-diffs --diff-context 0
  
  # Show read/write sizes and include lseek operations
  sudo sniff-writes monitor --show-sizes -o open,read,write,close,lseek

  # Filter by process name and file patterns
  sudo sniff-writes monitor -p nginx --glob "*.log" --glob "!*.tmp"

  # Filter by process glob patterns (include nginx*, exclude systemd*)
  sudo sniff-writes monitor --process-glob "nginx*" --process-glob "!systemd*"
  
  # Mix positive and negative patterns
  sudo sniff-writes monitor --glob "*.log" --glob "*.txt" --glob "!*debug*"

  # Log events to SQLite database
  sudo sniff-writes monitor --sqlite /tmp/file_events.db

  # Combined filtering with database logging
  sudo sniff-writes monitor --glob "*.go" --process-glob "*server" --sqlite /tmp/go_files.db -v

  # Enable real-time web UI
  sudo sniff-writes monitor --web --web-port 8080

  # Web UI with filtering and database logging
  sudo sniff-writes monitor --web --glob "*.log" --sqlite /tmp/web_events.db`,
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "Start monitoring file operations",
	Long:  "Start monitoring file operations on the specified directory using eBPF tracepoints.",
	RunE:  runMonitor,
}

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Query past events from database",
	Long:  "Query and search past file operation events stored in the database with filtering and pagination.",
	RunE:  runQuery,
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start REST API server",
	Long:  "Start a REST API server to query events via HTTP endpoints.",
	RunE:  runServer,
}

func init() {
	// Add subcommands
	rootCmd.AddCommand(monitorCmd)
	rootCmd.AddCommand(queryCmd)
	rootCmd.AddCommand(serverCmd)

	// Add global logging flags
	rootCmd.PersistentFlags().Bool("debug-logging", false, "Enable debug logging")
	rootCmd.PersistentFlags().String("log-level", "info", "Set log level (debug, info, warn, error)")

	// Set up pre-run hook to configure logging
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		debugLogging, _ := cmd.Flags().GetBool("debug-logging")
		logLevel, _ := cmd.Flags().GetString("log-level")

		if debugLogging {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		} else {
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				return fmt.Errorf("invalid log level: %s", logLevel)
			}
			zerolog.SetGlobalLevel(level)
		}

		zlog.Debug().Msg("Debug logging enabled")
		return nil
	}

	// Add flags to monitor command
	monitorCmd.Flags().StringVarP(&config.Directory, "directory", "d", "cmd/n8n-cli", "Directory to monitor")
	monitorCmd.Flags().StringVarP(&config.OutputFormat, "format", "f", "plain", "Output format: plain, json, table")
	monitorCmd.Flags().StringSliceVarP(&config.Operations, "operations", "o", []string{"open", "write", "close"}, "Operations to monitor (add 'read' for read operations, 'lseek' for seek operations)")
	monitorCmd.Flags().StringVarP(&config.ProcessFilter, "process", "p", "", "Filter by process name (substring match)")
	monitorCmd.Flags().DurationVarP(&config.Duration, "duration", "t", 0, "Duration to run (0 = infinite)")
	monitorCmd.Flags().BoolVarP(&config.Verbose, "verbose", "v", false, "Verbose output")
	monitorCmd.Flags().BoolVar(&config.ShowFd, "show-fd", false, "Show file descriptor numbers")
	monitorCmd.Flags().BoolVar(&config.ShowSizes, "show-sizes", false, "Show read/write sizes in output")
	monitorCmd.Flags().StringVar(&config.OutputFile, "output", "", "Output to file instead of stdout")
	monitorCmd.Flags().BoolVar(&config.Debug, "debug", false, "Debug mode - show all events regardless of filters")
	monitorCmd.Flags().BoolVar(&config.ShowAllFiles, "show-all-files", false, "Show all file types including pipes, sockets, etc. (default: only regular files)")
	monitorCmd.Flags().BoolVar(&config.CaptureContent, "capture-content", false, "Capture write content (warning: may impact performance)")
	monitorCmd.Flags().IntVar(&config.ContentSize, "content-size", 4096, "Maximum bytes of write content to capture (default: 4096)")
	monitorCmd.Flags().BoolVar(&config.ShowDiffs, "show-diffs", false, "Show diffs for write operations (requires --capture-content)")
	monitorCmd.Flags().StringVar(&config.DiffFormat, "diff-format", "unified", "Diff format: unified, pretty (default: unified)")
	monitorCmd.Flags().BoolVar(&config.NoColor, "no-color", false, "Disable colored output for diffs")
	monitorCmd.Flags().IntVar(&config.DiffContextLines, "diff-context", 3, "Number of context lines to show around changes in diffs (default: 3)")
	monitorCmd.Flags().StringSliceVar(&config.GlobPatterns, "glob", []string{}, "Include/exclude files with glob patterns (e.g., '*.go', '!*.tmp')")
	monitorCmd.Flags().StringSliceVar(&config.GlobExclude, "glob-exclude", []string{}, "Exclude files matching these glob patterns (legacy, use --glob '!pattern')")
	monitorCmd.Flags().StringSliceVar(&config.ProcessGlob, "process-glob", []string{}, "Include/exclude processes with glob patterns (e.g., 'nginx*', '!systemd*')")
	monitorCmd.Flags().StringSliceVar(&config.ProcessGlobExclude, "process-glob-exclude", []string{}, "Exclude processes matching these glob patterns (legacy, use --process-glob '!pattern')")
	monitorCmd.Flags().StringVar(&config.SqliteDB, "sqlite", "", "Log events to SQLite database (specify database file path)")
	monitorCmd.Flags().BoolVar(&config.WebUI, "web", false, "Enable real-time web UI")
	monitorCmd.Flags().IntVar(&config.WebPort, "web-port", 8080, "Web UI port (default: 8080)")

	// Add flags to query command
	queryCmd.Flags().StringVar(&config.SqliteDB, "sqlite", "", "SQLite database file path (required)")
	queryCmd.Flags().StringVar(&config.ProcessFilter, "process", "", "Filter by process name (substring match)")
	queryCmd.Flags().StringSliceVarP(&config.Operations, "operations", "o", []string{}, "Filter by operations: open, read, write, close")
	queryCmd.Flags().StringVar(&config.OutputFormat, "format", "table", "Output format: plain, json, table")
	queryCmd.Flags().StringVar(&config.OutputFile, "output", "", "Output to file instead of stdout")
	queryCmd.Flags().IntVar(&queryFlags.limit, "limit", 100, "Maximum number of events to return")
	queryCmd.Flags().IntVar(&queryFlags.offset, "offset", 0, "Number of events to skip (for pagination)")
	queryCmd.Flags().StringVar(&queryFlags.startTime, "start-time", "", "Start time filter (RFC3339 format)")
	queryCmd.Flags().StringVar(&queryFlags.endTime, "end-time", "", "End time filter (RFC3339 format)")
	queryCmd.Flags().StringVar(&queryFlags.filename, "filename", "", "Filter by filename pattern")
	queryCmd.Flags().Uint32Var(&queryFlags.pid, "pid", 0, "Filter by process ID")
	queryCmd.Flags().StringVar(&queryFlags.exportFmt, "export", "", "Export format: json, csv, markdown")
	queryCmd.MarkFlagRequired("sqlite")

	// Add flags to server command
	serverCmd.Flags().StringVar(&config.SqliteDB, "sqlite", "", "SQLite database file path (required)")
	serverCmd.Flags().IntVar(&serverFlags.port, "port", 8080, "API server port (default: 8080)")
	serverCmd.MarkFlagRequired("sqlite")
}

func runMonitor(cmd *cobra.Command, args []string) error {
	// Check for root privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("this program requires root privileges to load eBPF programs")
	}

	// Validate content size parameter
	if config.ContentSize < 1 || config.ContentSize > 131072 { // Allow up to 128KB (32 chunks * 4096)
		return fmt.Errorf("content-size must be between 1 and 131072 bytes (128KB)")
	}

	// Validate diff options
	if config.ShowDiffs && !config.CaptureContent {
		return fmt.Errorf("--show-diffs requires --capture-content to be enabled")
	}
	
	// Validate diff context lines
	if config.DiffContextLines < 0 {
		return fmt.Errorf("--diff-context must be >= 0")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Set up duration-based cancellation if specified
	if config.Duration > 0 {
		durationCtx, durationCancel := context.WithTimeout(ctx, config.Duration)
		defer durationCancel()
		ctx = durationCtx
	}

	// Initialize path cache and file cache
	pathCache = cache.New()
	fileCache = filecache.New()

	// Initialize SQLite database if specified
	if err := initSQLite(); err != nil {
		return fmt.Errorf("failed to initialize SQLite: %w", err)
	}
	defer closeSQLite()

	// Start web server if enabled
	if config.WebUI {
		startWebServer()
	}

	// Remove memory limit for eBPF
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("failed to remove memlock limit: %w", err)
	}

	// Load pre-compiled programs and maps into the kernel
	spec, err := loadSniffwrites()
	if err != nil {
		return fmt.Errorf("failed to load eBPF spec: %w", err)
	}

	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("failed to create eBPF collection: %w", err)
	}
	defer coll.Close()

	// Attach tracepoints based on selected operations
	links, err := ebpfops.AttachTracepoints(coll, &config)
	if err != nil {
		return fmt.Errorf("failed to attach tracepoints: %w", err)
	}
	defer func() {
		for _, l := range links {
			l.Close()
		}
	}()

	// Configure content capture in eBPF
	if err := ebpfops.ConfigureContentCapture(coll, &config); err != nil {
		return fmt.Errorf("failed to configure content capture: %w", err)
	}

	// Open ring buffer reader
	rd, err := ringbuf.NewReader(coll.Maps["events"])
	if err != nil {
		return fmt.Errorf("failed to create ring buffer reader: %w", err)
	}
	defer rd.Close()

	if config.Verbose {
		fmt.Printf("Monitoring %s operations on directory '%s'...\n",
			strings.Join(config.Operations, ", "), config.Directory)
		if config.ProcessFilter != "" {
			fmt.Printf("Filtering processes containing: %s\n", config.ProcessFilter)
		}
		if config.Duration > 0 {
			fmt.Printf("Running for: %s\n", config.Duration)
		}
		if config.CaptureContent {
			fmt.Printf("Content capture enabled (max %d bytes per chunk)\n", config.ContentSize)
		}
	} else {
		// Always show which operations are being monitored for clarity
		fmt.Printf("Monitoring operations: %s\n", strings.Join(config.Operations, ", "))
		if config.CaptureContent {
			fmt.Printf("Content capture enabled\n")
		}
	}

	// Set up output writer
	outputWriter := os.Stdout
	if config.OutputFile != "" {
		file, err := os.Create(config.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		outputWriter = file
	}

	// Handle context cancellation
	go func() {
		<-ctx.Done()
		rd.Close()
	}()

	// Process events
	return processEvents(ctx, rd, outputWriter)
}

func processEvents(ctx context.Context, rd *ringbuf.Reader, outputWriter *os.File) error {
	var event models.Event

	// Print table header if using table format
	if config.OutputFormat == "table" {
		formatter.PrintTableHeader(outputWriter, &config)
	}

	for {
		record, err := rd.Read()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			zlog.Debug().Err(err).Msg("Error reading from perf event reader")
			continue
		}

		// Ring buffer doesn't have LostSamples like perf events
		// Ring buffer handles overflow differently

		if err := processor.ParseEvent(record.RawSample, &event); err != nil {
			zlog.Debug().Err(err).Msg("Failed to parse event")
			continue
		}

		// Resolve the path before filtering
		resolvedPath := cache.ResolvePath(&event, pathCache)

		if config.Debug || processor.ShouldProcessEvent(&event, resolvedPath, &config) {
			if config.Debug {
				comm := cString(event.Comm[:])
				fmt.Printf("[DEBUG] PID=%d FD=%d COMM=%s FILE=%s HASH=%d TYPE=%d\n",
					event.Pid, event.Fd, comm, resolvedPath, event.PathHash, event.Type)
			}
			if !config.Debug || processor.ShouldProcessEvent(&event, resolvedPath, &config) {
				outputEvent(&event, resolvedPath, outputWriter)
			}
		}
	}
}

func outputEvent(event *models.Event, resolvedPath string, writer *os.File) {
	eventOutput := formatter.CreateEventOutput(event, resolvedPath, &config)

	// Handle content caching and diff generation
	if config.CaptureContent && event.ContentLen > 0 {
		content := cString(event.Content[:event.ContentLen])
		contentBytes := []byte(content)

		switch event.Type {
		case 1: // read
			// Store read content for future diff comparison
			fileCache.StoreReadContent(event.Pid, event.Fd, event.PathHash, contentBytes, event.FileOffset)
		case 2: // write
			// Generate diff if we have cached read content and diffs are enabled
			if config.ShowDiffs {
				var diffText string
				var hasDiff bool

				if config.DiffFormat == "pretty" {
					diffText, hasDiff = fileCache.GenerateDiff(event.Pid, event.Fd, event.PathHash, event.FileOffset, contentBytes)
				} else {
					// Use elided diff if context lines is configured
					if config.DiffContextLines >= 0 {
						diffText, hasDiff = fileCache.GenerateElidedUnifiedDiff(event.Pid, event.Fd, event.PathHash, event.FileOffset, contentBytes, resolvedPath, config.DiffContextLines)
					} else {
						diffText, hasDiff = fileCache.GenerateUnifiedDiff(event.Pid, event.Fd, event.PathHash, event.FileOffset, contentBytes, resolvedPath)
					}
				}

				if hasDiff {
					// For web UI, format the diff with HTML
					if config.WebUI {
						eventOutput.Diff = formatter.FormatDiffForWeb(diffText)
					} else {
						eventOutput.Diff = diffText
					}
				}
			}

			// Update cache with new written content
			fileCache.UpdateWriteContent(event.Pid, event.Fd, event.PathHash, contentBytes, event.FileOffset)
		}
	}

	// Log to SQLite if configured
	if err := sqliteDB.LogEvent(eventOutput); err != nil {
		zlog.Error().Err(err).Msg("Failed to log event to SQLite")
	}

	// Broadcast to WebSocket clients if web UI is enabled
	if config.WebUI && webHub != nil {
		webHub.Broadcast(eventOutput)
	}

	switch config.OutputFormat {
	case "json":
		formatter.OutputJSON(eventOutput, writer, &config)
	case "table":
		formatter.OutputTable(eventOutput, writer, &config)
	default:
		formatter.OutputPlain(eventOutput, writer, &config)
	}
}

func cString(b []int8) string {
	n := -1
	for i, v := range b {
		if v == 0 {
			n = i
			break
		}
	}
	if n == -1 {
		n = len(b)
	}
	// Convert []int8 to []byte
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = byte(b[i])
	}
	return string(bytes)
}

func runQuery(cmd *cobra.Command, args []string) error {
	// Initialize SQLite database
	if err := initSQLite(); err != nil {
		return fmt.Errorf("failed to initialize SQLite: %w", err)
	}
	defer closeSQLite()

	// Build query filter
	filter := database.QueryFilter{
		ProcessFilter:   config.ProcessFilter,
		OperationFilter: config.Operations,
		FilenamePattern: queryFlags.filename,
		Limit:           queryFlags.limit,
		Offset:          queryFlags.offset,
	}

	// Parse time filters
	if queryFlags.startTime != "" {
		startTime, err := time.Parse(time.RFC3339, queryFlags.startTime)
		if err != nil {
			return fmt.Errorf("invalid start-time format (use RFC3339): %w", err)
		}
		filter.StartTime = &startTime
	}

	if queryFlags.endTime != "" {
		endTime, err := time.Parse(time.RFC3339, queryFlags.endTime)
		if err != nil {
			return fmt.Errorf("invalid end-time format (use RFC3339): %w", err)
		}
		filter.EndTime = &endTime
	}

	if queryFlags.pid != 0 {
		filter.PID = &queryFlags.pid
	}

	// Query events
	events, err := sqliteDB.QueryEvents(filter)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	// Handle export format
	if queryFlags.exportFmt != "" {
		var exportFormat export.ExportFormat
		switch queryFlags.exportFmt {
		case "json":
			exportFormat = export.FormatJSON
		case "csv":
			exportFormat = export.FormatCSV
		case "markdown":
			exportFormat = export.FormatMarkdown
		default:
			return fmt.Errorf("unsupported export format: %s. Use: json, csv, markdown", queryFlags.exportFmt)
		}

		// Set up output writer
		outputWriter := os.Stdout
		if config.OutputFile != "" {
			file, err := os.Create(config.OutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer file.Close()
			outputWriter = file
		}

		exporter := export.New(outputWriter, exportFormat)
		return exporter.Export(events)
	}

	// Regular output using existing formatter
	outputWriter := os.Stdout
	if config.OutputFile != "" {
		file, err := os.Create(config.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		outputWriter = file
	}

	// Print table header if using table format
	if config.OutputFormat == "table" {
		formatter.PrintTableHeader(outputWriter, &config)
	}

	for _, event := range events {
		switch config.OutputFormat {
		case "json":
			formatter.OutputJSON(event, outputWriter, &config)
		case "table":
			formatter.OutputTable(event, outputWriter, &config)
		default:
			formatter.OutputPlain(event, outputWriter, &config)
		}
	}

	fmt.Printf("\nTotal events found: %d\n", len(events))
	return nil
}

func runServer(cmd *cobra.Command, args []string) error {
	zlog.Info().Msg("Starting sniff-writes REST API server")

	// Initialize SQLite database
	if err := initSQLite(); err != nil {
		zlog.Error().Err(err).Str("database", config.SqliteDB).Msg("Failed to initialize SQLite database")
		return fmt.Errorf("failed to initialize SQLite: %w", err)
	}
	defer closeSQLite()

	zlog.Info().Str("database", config.SqliteDB).Msg("SQLite database initialized successfully")

	// Create and start API server
	server := api.NewServer(sqliteDB, &config, serverFlags.port)
	zlog.Info().Int("port", serverFlags.port).Msg("Starting API server")
	fmt.Printf("Starting API server on port %d\n", serverFlags.port)
	return server.Start()
}
