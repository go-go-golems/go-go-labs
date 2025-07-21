package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
- Hierarchical task organization (parent-child relationships)
- DAG-based task dependencies 
- Agent assignment with UUIDs
- Multiple output formats (table, JSON, YAML, CSV)
- Intelligent task scheduling based on dependencies`,
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
