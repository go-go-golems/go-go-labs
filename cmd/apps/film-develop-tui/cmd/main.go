package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg"
)

var (
	logLevel string
	version  = "dev"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "film-develop-tui",
		Short: "A TUI for calculating B&W film development parameters",
		Long: `A minimal terminal user interface for calculating B&W film development parameters 
using ILFOSOL 3, ILFOSTOP, and Sprint Fixer.

Features:
- Film type and EI rating selection
- Tank size calculation based on roll count and format
- Chemical dilution calculations
- Fixer capacity tracking
- Development time display`,
		Version: version,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup logging
			setupLogging()

			log.Info().Msg("Starting film development TUI")

			// Create and run the application
			model := pkg.NewModel()

			p := tea.NewProgram(
				model,
				tea.WithAltScreen(),
				tea.WithMouseCellMotion(),
			)

			if _, err := p.Run(); err != nil {
				log.Error().Err(err).Msg("Failed to run application")
				return fmt.Errorf("failed to run application: %w", err)
			}

			return nil
		},
	}

	// Add flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")

	// Execute the command
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Failed to execute command")
		os.Exit(1)
	}
}

func setupLogging() {
	// Parse log level
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// Set global log level
	zerolog.SetGlobalLevel(level)

	// Configure pretty logging for development
	if level <= zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
}
