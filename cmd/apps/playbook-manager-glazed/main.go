package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/go-go-golems/go-go-labs/pkg/playbook"
)

func main() {
	var (
		logLevel string
		dbPath   string
	)

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
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Database path (default: ~/.playbooks/registry.db)")

	// Initialize help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Initialize storage
	var storage *playbook.Storage
	initStorage := func() {
		if storage != nil {
			return
		}

		if dbPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to get user home directory")
			}
			dbPath = filepath.Join(homeDir, ".playbooks", "registry.db")
		}

		var err error
		storage, err = playbook.NewStorage(dbPath)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to initialize storage")
		}
	}

	// Create commands
	commands := []func() (cmds.Command, error){
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewRegisterCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewCreateCollectionCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewListCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewSearchCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewShowCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewDeployCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewAddToCollectionCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewRemoveCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewSetMetadataCommand(storage)
		},
		func() (cmds.Command, error) {
			initStorage()
			return playbook.NewGetMetadataCommand(storage)
		},
	}

	// Convert commands to Cobra commands and add to root
	for _, cmdFunc := range commands {
		cmd, err := cmdFunc()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating command: %v\n", err)
			os.Exit(1)
		}

		cobraCmd, err := cli.BuildCobraCommandFromCommand(cmd)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error building Cobra command: %v\n", err)
			os.Exit(1)
		}

		rootCmd.AddCommand(cobraCmd)
	}

	// Add cleanup
	cobra.OnFinalize(func() {
		if storage != nil {
			storage.Close()
		}
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
