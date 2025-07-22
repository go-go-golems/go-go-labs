package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	dbPath   string
	logLevel string
	logger   zerolog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "amp-tasks",
	Short: "Agent task coordination system",
	Long: `A hierarchical task management system for coding agents with DAG dependencies.
	
Features:
- Project management with guidelines and context
- Hierarchical task organization (parent-child relationships)
- DAG-based task dependencies 
- Agent types and smart assignment
- Multiple output formats (table, JSON, YAML, CSV)
- Intelligent task scheduling based on dependencies

Quick Start:
- amp-tasks demo                    # Set up sample data
- amp-tasks docs quick-start        # Essential commands
- amp-tasks docs agent-guide        # Work reference
- amp-tasks docs workflow           # Detailed workflow`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Setup logging
		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid log level %q: %v\n", logLevel, err)
			os.Exit(1)
		}
		zerolog.SetGlobalLevel(level)
		logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

		// Make logger globally available
		log.Logger = logger
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "tasks.db", "Path to SQLite database")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
}
