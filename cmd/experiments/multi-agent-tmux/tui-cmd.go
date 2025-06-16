package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// createTUICommand creates the TUI subcommand
func createTUICommand() *cobra.Command {
	var socketPath string

	tuiCmd := &cobra.Command{
		Use:   "tui",
		Short: "Start TUI visualizer for agent output",
		Long: `Start a terminal user interface (TUI) that connects to a Unix socket
to display real-time agent output with styled formatting.`,
		Run: func(cmd *cobra.Command, args []string) {
			if socketPath == "" {
				fmt.Fprintf(os.Stderr, "Error: socket path is required\n")
				os.Exit(1)
			}

			if err := RunTUI(socketPath); err != nil {
				fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	tuiCmd.Flags().StringVar(&socketPath, "socket", "", "Unix socket path to connect to (required)")
	tuiCmd.MarkFlagRequired("socket")

	return tuiCmd
}
