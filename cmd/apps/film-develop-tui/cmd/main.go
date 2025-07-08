package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/pkg/ui/model"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "film-develop-tui",
		Short: "A terminal user interface for calculating B&W film development parameters using ILFOSOL 3, ILFOSTOP, and Sprint Fixer.",
		Run: func(cmd *cobra.Command, args []string) {
			// Configure zerolog
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				log.Fatal().Str("level", logLevel).Msg("Invalid log level")
			}
			zerolog.SetGlobalLevel(level)
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

			log.Info().Msg("Starting Film Development TUI")

			m := model.NewAppModel()
			p := tea.NewProgram(m, tea.WithAltScreen())

			if _, err := p.Run(); err != nil {
				fmt.Printf("Error running program: %v\n", err)
				os.Exit(1)
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
		os.Exit(1)
	}
}
