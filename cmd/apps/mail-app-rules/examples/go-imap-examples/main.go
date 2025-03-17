package main

import (
	"fmt"
	"os"
	"time"

	"github.com/emersion/go-message/examples/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Configure zerolog
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}
	log.Logger = zerolog.New(output).With().Timestamp().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	rootCmd := &cobra.Command{
		Use:   "imap-examples",
		Short: "Examples for working with IMAP and email messages",
		Long: `A collection of examples demonstrating how to use go-imap and go-message
to fetch and process email messages over IMAP.`,
	}

	// Add all commands
	rootCmd.AddCommand(cmd.ConnectCmd)
	rootCmd.AddCommand(cmd.FetchMetadataCmd)
	rootCmd.AddCommand(cmd.FetchStructureCmd)
	rootCmd.AddCommand(cmd.FetchContentCmd)
	rootCmd.AddCommand(cmd.FetchPartsCmd)
	rootCmd.AddCommand(cmd.StreamLargeCmd)
	rootCmd.AddCommand(cmd.HandleAlternativesCmd)
	rootCmd.AddCommand(cmd.HandleEmbeddedImagesCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
