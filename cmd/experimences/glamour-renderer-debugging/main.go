package main

import (
	"fmt"
	"os"
	"runtime"
	time "time"

	"github.com/charmbracelet/glamour"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logger     zerolog.Logger
	logFile    *os.File
	logLevel   string
	renderMd   bool // Flag to control if rendering happens
	initialMsg string
)

// --- Logging Setup (Copied from bubbletea-markdown-test) ---

func setupLogging(level string) error {
	// Create log file
	var err error
	// Log to a different file to avoid conflicts
	logFile, err = os.OpenFile("/tmp/glamour-debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Parse log level
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	consoleWriter := zerolog.ConsoleWriter{Out: logFile, TimeFormat: time.RFC3339}
	logger = zerolog.New(consoleWriter).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = logger

	logger.Info().Str("level", level).Msg("Logging initialized for glamour-debug")
	return nil
}

func logWithCaller(level zerolog.Level, msg string, fields map[string]interface{}) {
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	var event *zerolog.Event
	switch level {
	case zerolog.DebugLevel:
		event = logger.Debug()
	case zerolog.InfoLevel:
		event = logger.Info()
	case zerolog.WarnLevel:
		event = logger.Warn()
	case zerolog.ErrorLevel:
		event = logger.Error()
	default:
		event = logger.Info()
	}

	event.Str("file", file).Int("line", line)
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case bool:
			event.Bool(k, val)
		case time.Duration:
			event.Dur(k, val)
		default:
			event.Interface(k, v)
		}
	}

	event.Msg(msg)
}

// --- Test Application Logic ---

func runTest(cmd *cobra.Command, args []string) error {
	if err := setupLogging(logLevel); err != nil {
		return err
	}
	defer logFile.Close()

	logWithCaller(zerolog.InfoLevel, "Starting glamour renderer creation test", map[string]interface{}{})

	// Time the renderer creation
	startTime := time.Now()
	logWithCaller(zerolog.DebugLevel, "Calling glamour.NewTermRenderer", map[string]interface{}{
		"width": 80, // Fixed width for test
	})

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80), // Use a fixed width
	)

	duration := time.Since(startTime)

	logWithCaller(zerolog.InfoLevel, "glamour.NewTermRenderer call completed", map[string]interface{}{
		"duration_ms": duration.Milliseconds(),
		"duration":    duration,
	})

	if err != nil {
		logWithCaller(zerolog.ErrorLevel, "Failed to create glamour renderer", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	} else {
		logWithCaller(zerolog.InfoLevel, "Glamour renderer created successfully", nil)
	}

	// Optionally render some text if flag is set
	if renderMd {
		textToRender := initialMsg
		if textToRender == "" {
			textToRender = "# Test Header\n\nThis is *test* markdown."
		}
		logWithCaller(zerolog.InfoLevel, "Rendering test markdown", map[string]interface{}{})
		startTimeRender := time.Now()
		renderedContent, renderErr := renderer.Render(textToRender)
		durationRender := time.Since(startTimeRender)

		if renderErr != nil {
			logWithCaller(zerolog.ErrorLevel, "Failed to render markdown", map[string]interface{}{
				"error": renderErr.Error(),
			})
		} else {
			logWithCaller(zerolog.InfoLevel, "Markdown rendered successfully", map[string]interface{}{
				"duration_ms":   durationRender.Milliseconds(),
				"output_length": len(renderedContent),
			})
			// Optionally print to stdout for quick check
			// fmt.Println("--- Rendered Output ---")
			// fmt.Println(renderedContent)
			// fmt.Println("-----------------------")
		}
	}

	logWithCaller(zerolog.InfoLevel, "Glamour renderer test finished", nil)
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "glamour-debug",
		Short: "A minimal app to test glamour renderer creation speed",
		RunE:  runTest,
	}

	// Flags (similar to the main app)
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVar(&renderMd, "render-markdown", false, "Render test markdown after creating renderer")
	rootCmd.Flags().StringVar(&initialMsg, "initial-text", "", "Initial text to render if --render-markdown is true")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
