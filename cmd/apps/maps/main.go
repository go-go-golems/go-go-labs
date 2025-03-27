package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/apps/maps/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/apps/maps/doc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	// Create root command
	rootCmd, err := cmds.NewRootCommand()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating root command: %v\n", err)
		os.Exit(1)
	}

	// Initialize help system
	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Register embedded documentation
	err = helpSystem.LoadSectionsFromFS(doc.Files, ".")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading documentation: %v\n", err)
		os.Exit(1)
	}

	// initializing as mento-service to get all the environment variables
	err = clay.InitViper("maps", rootCmd)
	cobra.CheckErr(err)

	fmt.Println("Starting maps")
	log.Info().Msg("Starting maps")
	log.Debug().Msg("Starting maps")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Setup context with cancellation
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	// Execute command
	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
