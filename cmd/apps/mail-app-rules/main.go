package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/middlewares"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/apps/mail-app-rules/commands"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func main() {
	// Setup logging
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	rootCmd := &cobra.Command{
		Use:   "smailnail",
		Short: "Process mail rules on an IMAP server",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			err := clay.InitLogger()
			cobra.CheckErr(err)
		},
	}

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// initializing as snailmail-service to get all the environment variables
	err := clay.InitViper("smailnail", rootCmd)
	cobra.CheckErr(err)
	err = clay.InitLogger()
	cobra.CheckErr(err)

	log.Debug().Msg("Starting smailnail")

	// Configure middlewares for all commands
	middlewaresFunc := func(
		parsedLayers *layers.ParsedLayers,
		cmd *cobra.Command,
		args []string,
	) ([]middlewares.Middleware, error) {
		log.Debug().Msg("Setting up middlewares")
		return []middlewares.Middleware{
			// Command line args (highest priority)
			middlewares.ParseFromCobraCommand(cmd),
			middlewares.GatherArguments(args),

			// Viper config for environment variables
			middlewares.GatherFlagsFromViper(),

			// Defaults (lowest priority)
			middlewares.SetFromDefaults(),
		}, nil
	}

	// Create and add the mail-rules command
	mailRulesCmd, err := commands.NewMailRulesCommand()
	if err != nil {
		fmt.Printf("Error creating mail rules command: %v\n", err)
		os.Exit(1)
	}

	cobraMailRulesCmd, err := cli.BuildCobraCommandFromCommand(mailRulesCmd,
		cli.WithCobraMiddlewaresFunc(middlewaresFunc),
	)
	if err != nil {
		fmt.Printf("Error building Cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraMailRulesCmd)

	// Create and add the fetch-mail command
	fetchMailCmd, err := commands.NewFetchMailCommand()
	if err != nil {
		fmt.Printf("Error creating fetch mail command: %v\n", err)
		os.Exit(1)
	}

	cobraFetchMailCmd, err := cli.BuildCobraCommandFromCommand(fetchMailCmd,
		cli.WithCobraMiddlewaresFunc(middlewaresFunc),
	)
	if err != nil {
		fmt.Printf("Error building Cobra command: %v\n", err)
		os.Exit(1)
	}
	rootCmd.AddCommand(cobraFetchMailCmd)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Setup context with cancellation
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
		os.Exit(1)
	}
}
