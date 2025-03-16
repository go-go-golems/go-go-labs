package main

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type FileState struct {
	Size    int64
	ModTime time.Time
	IsNew   bool
}

var (
	fileStates      = make(map[string]FileState)
	fileStatesMutex sync.RWMutex
	lastChange      = time.Now()
	lastChangeMutex sync.RWMutex
	hasSmallNewFile bool

	// Command line flags
	watchDir          string
	logLevel          string
	smallFileInterval time.Duration
)

func updateLastChange() {
	lastChangeMutex.Lock()
	lastChange = time.Now()
	lastChangeMutex.Unlock()
}

func getLastChange() time.Time {
	lastChangeMutex.RLock()
	defer lastChangeMutex.RUnlock()
	return lastChange
}

func getWaitInterval() time.Duration {
	fileStatesMutex.RLock()
	defer fileStatesMutex.RUnlock()

	// Check if any new file is less than 100 bytes
	hasSmallNewFile = false
	for _, state := range fileStates {
		if state.IsNew && state.Size < 100 {
			hasSmallNewFile = true
			return smallFileInterval
		}
	}
	return 30 * time.Second
}

func checkFiles(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		fileStatesMutex.Lock()
		defer fileStatesMutex.Unlock()

		oldState, exists := fileStates[path]
		currentState := FileState{
			Size:    info.Size(),
			ModTime: info.ModTime(),
			IsNew:   !exists,
		}

		if !exists {
			log.Debug().
				Str("path", path).
				Int64("size", currentState.Size).
				Time("modTime", currentState.ModTime).
				Bool("smallFile", currentState.Size < 100).
				Msg("New file detected")
			updateLastChange()
		} else {
			if oldState.Size != currentState.Size {
				log.Debug().
					Str("path", path).
					Int64("oldSize", oldState.Size).
					Int64("newSize", currentState.Size).
					Time("modTime", currentState.ModTime).
					Bool("wasSmallFile", oldState.Size < 100).
					Bool("isSmallFile", currentState.Size < 100).
					Msg("File size changed")
				updateLastChange()
			} else if oldState.ModTime != currentState.ModTime {
				log.Debug().
					Str("path", path).
					Int64("size", currentState.Size).
					Time("oldModTime", oldState.ModTime).
					Time("newModTime", currentState.ModTime).
					Msg("File modification time changed")
				updateLastChange()
			}
			// Preserve IsNew status if size is still < 100 bytes
			currentState.IsNew = oldState.IsNew && oldState.Size < 100
		}

		fileStates[path] = currentState
		return nil
	})
}

func setupLogging(level string) {
	// Set up pretty console logging
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	// Parse and set log level
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		log.Warn().Msgf("Invalid log level %q, defaulting to debug", level)
		logLevel = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(logLevel)
}

func run() error {
	setupLogging(logLevel)

	log.Info().
		Str("directory", watchDir).
		Dur("normalInterval", 30*time.Second).
		Dur("smallFileInterval", smallFileInterval).
		Msg("Starting directory watch")

	for {
		if err := checkFiles(watchDir); err != nil {
			log.Error().Err(err).Msg("Error checking files")
		}

		waitInterval := getWaitInterval()
		timeSinceLastChange := time.Since(getLastChange())

		if timeSinceLastChange >= waitInterval {
			log.Info().
				Dur("duration", timeSinceLastChange.Round(time.Second)).
				Dur("waitInterval", waitInterval).
				Bool("hadSmallNewFile", hasSmallNewFile).
				Msg("No changes detected")
			return nil
		}

		time.Sleep(1 * time.Second)
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "watch-for-scan",
		Short: "Watch a directory for file changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return run()
		},
	}

	rootCmd.Flags().StringVarP(&watchDir, "dir", "d", ".", "Directory to watch")
	rootCmd.Flags().StringVarP(&logLevel, "log-level", "l", "debug", "Log level (debug, info, warn, error)")
	rootCmd.Flags().DurationVarP(&smallFileInterval, "small-file-interval", "s", 2*time.Minute, "Wait interval when small files (<100 bytes) are present")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}
