package main

import (
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/go-go-labs/cmd/apps/feedboii/cmds"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "feedboii",
	Short: "Downloads and outputs an RSS feed as structured data.",
}

func main() {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	log.Logger = log.Output(output)

	jsonCmd, err := cmds.NewFeedCommand()
	cobra.CheckErr(err)
	command, err := cli.BuildCobraCommandFromGlazeCommand(jsonCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(command)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Error executing command")
	}
}
