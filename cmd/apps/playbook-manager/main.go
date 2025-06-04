package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/pkg/playbook"
)

var (
	storage   *playbook.Storage
	logLevel  string
	dbPath    string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "pb",
		Short: "Playbook Manager - manage contextual documents for LLMs",
		Long:  `Playbook Manager is a CLI tool for managing playbooks and collections of contextual documents that can be passed to LLMs.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set up logging
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				level = zerolog.InfoLevel
			}
			zerolog.SetGlobalLevel(level)
			log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

			// Initialize storage
			if dbPath == "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					log.Fatal().Err(err).Msg("Failed to get user home directory")
				}
				dbPath = filepath.Join(homeDir, ".playbooks", "registry.db")
			}

			storage, err = playbook.NewStorage(dbPath)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize storage")
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if storage != nil {
				storage.Close()
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Database path (default: ~/.playbooks/registry.db)")

	// Add commands
	rootCmd.AddCommand(registerCmd())
	rootCmd.AddCommand(createCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(searchCmd())
	rootCmd.AddCommand(showCmd())
	rootCmd.AddCommand(addCmd())
	rootCmd.AddCommand(removeCmd())
	rootCmd.AddCommand(metaCmd())
	rootCmd.AddCommand(deployCmd())
	rootCmd.AddCommand(updateCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
