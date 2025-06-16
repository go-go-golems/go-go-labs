package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/GianlucaP106/gotmux/gotmux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logLevel    string
	sessionName string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "tmux-control",
		Short: "Experiment with tmux control using gotmux",
		Run:   runExperiment,
	}

	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().StringVar(&sessionName, "session", "go-experiment", "Tmux session name")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runExperiment(cmd *cobra.Command, args []string) {
	setupLogging()

	log.Info().Str("session", sessionName).Msg("Starting tmux control experiment")

	ctx := context.Background()
	if err := tmuxExperiment(ctx); err != nil {
		log.Fatal().Err(err).Msg("Tmux experiment failed")
	}

	log.Info().Msg("Tmux experiment completed successfully")
}

func setupLogging() {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func tmuxExperiment(ctx context.Context) error {
	// Connect to default tmux server (or start one)
	log.Debug().Msg("Connecting to tmux server")
	tmux, err := gotmux.DefaultTmux()
	if err != nil {
		return fmt.Errorf("failed to connect to tmux: %w", err)
	}

	// Create a new session
	log.Info().Str("session", sessionName).Msg("Creating new session")
	session, err := tmux.NewSession(&gotmux.SessionOptions{
		Name: sessionName,
	})
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Create a logs window
	log.Debug().Msg("Creating logs window")
	logsWindow, err := session.NewWindow(nil)
	if err != nil {
		return fmt.Errorf("failed to create logs window: %w", err)
	}

	// Get the default pane from logs window
	pane, err := logsWindow.GetPaneByIndex(0)
	if err != nil {
		return fmt.Errorf("failed to get default pane: %w", err)
	}

	// Split the pane
	log.Debug().Msg("Splitting window into panes")
	err = pane.Split()
	if err != nil {
		return fmt.Errorf("failed to split pane: %w", err)
	}

	// Get the new pane (should be index 1)
	rightPane, err := logsWindow.GetPaneByIndex(1)
	if err != nil {
		return fmt.Errorf("failed to get right pane: %w", err)
	}

	// Run tail in the right pane
	log.Debug().Msg("Starting tail command in right pane")
	err = rightPane.SendKeys("echo 'Monitoring logs...' && tail -f /var/log/syslog || echo 'No syslog found, showing dmesg instead' && dmesg -w")
	if err != nil {
		return fmt.Errorf("failed to send keys to right pane: %w", err)
	}

	// Create a development window
	log.Debug().Msg("Creating development window")
	devWindow, err := session.NewWindow(nil)
	if err != nil {
		return fmt.Errorf("failed to create dev window: %w", err)
	}

	// Get the default pane from dev window
	devPane, err := devWindow.GetPaneByIndex(0)
	if err != nil {
		return fmt.Errorf("failed to get dev pane: %w", err)
	}

	// Split dev window vertically
	err = devPane.Split()
	if err != nil {
		return fmt.Errorf("failed to split dev window: %w", err)
	}

	// Get the bottom pane
	bottomPane, err := devWindow.GetPaneByIndex(1)
	if err != nil {
		return fmt.Errorf("failed to get bottom pane: %w", err)
	}

	// Run htop in bottom pane
	log.Debug().Msg("Starting htop in bottom pane")
	err = bottomPane.SendKeys("htop || top")
	if err != nil {
		return fmt.Errorf("failed to send keys to bottom pane: %w", err)
	}

	// Send some commands to the main dev pane
	log.Debug().Msg("Setting up main dev pane")
	err = devPane.SendKeys("echo 'Welcome to the tmux experiment!'")
	if err != nil {
		return fmt.Errorf("failed to send welcome message: %w", err)
	}

	err = devPane.SendKeys("echo 'Session: " + sessionName + "'")
	if err != nil {
		return fmt.Errorf("failed to send session info: %w", err)
	}

	// Wait a moment for things to settle
	time.Sleep(2 * time.Second)

	// List all sessions
	log.Debug().Msg("Listing tmux sessions")
	sessions, err := tmux.ListSessions()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list sessions")
	} else {
		log.Info().Int("count", len(sessions)).Msg("Found tmux sessions")
		for _, s := range sessions {
			log.Info().Str("session", s.Name).Msg("Active session")
		}
	}

	log.Info().Str("session", sessionName).Msg("Tmux experiment setup complete. Use 'tmux attach -t " + sessionName + "' to connect")

	return nil
}
