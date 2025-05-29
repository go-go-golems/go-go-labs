package cmd

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/apps/poll-modem/internal/tui"
)

var (
	url          string
	pollInterval time.Duration
	username     string
	password     string

	rootCmd = &cobra.Command{
		Use:   "poll-modem",
		Short: "Cable modem monitoring TUI",
		Long: `A Terminal User Interface (TUI) application that polls a cable modem's 
network setup page and displays the channel information in a nice table format.

The application continuously polls the modem endpoint and displays:
- Cable modem hardware information
- Downstream channel details (frequency, SNR, power levels, etc.)
- Upstream channel details
- Error codeword statistics

Use tab/shift+tab to navigate between different views.

If the modem requires authentication, provide username and password flags.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return logging.InitLoggerFromViper()
		},
		RunE: runTUI,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	err := clay.InitViper("poll-modem", rootCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Viper")
	}

	// Add application-specific flags
	rootCmd.PersistentFlags().StringVarP(&url, "url", "u", "http://192.168.0.1", "Modem base URL (e.g., http://192.168.0.1)")
	rootCmd.PersistentFlags().DurationVarP(&pollInterval, "interval", "i", 30*time.Second, "Poll interval (e.g., 30s, 1m, 5m)")
	rootCmd.PersistentFlags().StringVarP(&username, "username", "n", "", "Modem username for authentication")
	rootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "Modem password for authentication")
}

func runTUI(cmd *cobra.Command, args []string) error {
	logging.InitLoggerFromViper()
	log.Info().Msg("Running TUI")

	// Use the provided URL as base URL
	baseURL := url
	if baseURL == "" {
		baseURL = "http://192.168.0.1"
	}

	app := tui.NewApp(baseURL, pollInterval, username, password)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()

	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}
