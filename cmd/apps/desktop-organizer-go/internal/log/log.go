package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
)

// InitLogger initializes the zerolog logger with console and optional file outputs.
// It takes the desired log level as a string and the path for an optional debug log file.
// Returns a context with the configured logger attached.
func InitLogger(logLevelStr string, debugLogPath string) (context.Context, error) {
	// Setup console writer
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	// Determine log level from string
	logLevel, err := zerolog.ParseLevel(strings.ToLower(logLevelStr))
	if err != nil {
		// Use the global logger for this initial warning, as the context logger isn't fully set up yet
		zlog.Warn().Str("level", logLevelStr).Msg("Invalid log level provided, defaulting to info")
		logLevel = zerolog.InfoLevel
	}

	var writers []io.Writer
	writers = append(writers, consoleWriter)

	// Setup file logger if path provided
	var fileWriter io.Writer
	if debugLogPath != "" {
		logFile, err := os.Create(debugLogPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create debug log file: %w", err)
		}

		// File always gets debug level logs in JSON format
		fileWriter = zerolog.New(logFile).With().
			Timestamp().
			Caller().
			Logger().Level(zerolog.DebugLevel)

		writers = append(writers, fileWriter)
	}

	// Create multi-writer
	// The console writer within multiWriter will respect the main logLevel.
	// The fileWriter (if present) always logs at Debug level.
	multiWriter := zerolog.MultiLevelWriter(writers...)

	// Set global logger
	logger := zerolog.New(multiWriter).With().
		Timestamp().
		Logger().Level(logLevel)

	zlog.Logger = logger

	// Create context with logger
	ctx := logger.WithContext(context.Background())

	return ctx, nil
}

// FromCtx retrieves the logger from context.
// If none exists, returns the global logger.
func FromCtx(ctx context.Context) *zerolog.Logger {
	logger := zerolog.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
		return &zlog.Logger
	}
	return logger
}

// Debug is a shortcut for FromCtx(ctx).Debug()
func Debug(ctx context.Context) *zerolog.Event {
	return FromCtx(ctx).Debug()
}

// Info is a shortcut for FromCtx(ctx).Info()
func Info(ctx context.Context) *zerolog.Event {
	return FromCtx(ctx).Info()
}

// Error is a shortcut for FromCtx(ctx).Error()
func Error(ctx context.Context) *zerolog.Event {
	return FromCtx(ctx).Error()
}

// Warn is a shortcut for FromCtx(ctx).Warn()
func Warn(ctx context.Context) *zerolog.Event {
	return FromCtx(ctx).Warn()
}

// Fatal is a shortcut for FromCtx(ctx).Fatal()
func Fatal(ctx context.Context) *zerolog.Event {
	return FromCtx(ctx).Fatal()
}
