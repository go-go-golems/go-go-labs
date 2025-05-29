package cmd

import (
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/cmd/apps/poll-modem/internal/tui"
)

var (
	url          string
	pollInterval time.Duration

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

Use tab/shift+tab to navigate between different views.`,
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
	// Initialize logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Add logging layer to root command
	err := logging.AddLoggingLayerToRootCommand(rootCmd, "poll-modem")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to add logging layer")
	}

	// Add application-specific flags
	rootCmd.PersistentFlags().StringVarP(&url, "url", "u", "http://192.168.0.1/network_setup.jst", "Modem URL to poll")
	rootCmd.PersistentFlags().DurationVarP(&pollInterval, "interval", "i", 30*time.Second, "Poll interval (e.g., 30s, 1m, 5m)")
}

func runTUI(cmd *cobra.Command, args []string) error {
	// For TUI applications, we want to disable logging to avoid interfering with the display
	// unless debug logging is specifically enabled
	logLevel, _ := cmd.Flags().GetString("log-level")
	if logLevel != "debug" && logLevel != "trace" {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	}

	app := tui.NewApp(url, pollInterval)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err := p.Run()

	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
} 