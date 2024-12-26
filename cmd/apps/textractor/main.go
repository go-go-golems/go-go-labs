package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/cmds"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/cmds/debug"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "textractor",
		Short: "Manage Textractor AWS resources and process PDFs",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logLevel, _ := cmd.Flags().GetString("log-level")
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				log.Warn().Msgf("Invalid log level: %s, defaulting to info", logLevel)
				level = zerolog.InfoLevel
			}
			zerolog.SetGlobalLevel(level)
		},
	}

	// Configure zerolog for console output and show caller information
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.With().Caller().Logger().Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	helpSystem := help.NewHelpSystem()
	helpSystem.SetupCobraRootCommand(rootCmd)

	// Add persistent flags to root command
	rootCmd.PersistentFlags().String("tf-dir", "terraform", "Directory containing Terraform state")
	rootCmd.PersistentFlags().String("config", "", "JSON config file containing resource configuration")
	rootCmd.PersistentFlags().String("log-level", "info", "Set the logging level (debug, info, warn, error, fatal)")

	// Initialize list command with glazed support
	listCmd, err := cmds.NewListCommand()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create list command")
	}
	cobraListCmd, err := cli.BuildCobraCommandFromGlazeCommand(listCmd)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to build cobra list command")
	}
	rootCmd.AddCommand(cobraListCmd)

	// Add other commands
	rootCmd.AddCommand(debug.NewDebugCommand())
	rootCmd.AddCommand(newSaveConfigCommand())
	rootCmd.AddCommand(cmds.NewSubmitCommand())
	rootCmd.AddCommand(cmds.NewStatusCommand())
	rootCmd.AddCommand(cmds.NewFetchCommand())

	addDebugVarCommands(rootCmd, "terraform")

	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func addDebugVarCommands(rootCmd *cobra.Command, tfDir string) {
	debugVarsCmd := &cobra.Command{
		Use:   "debug-vars",
		Short: "Print environment variables for debugging",
		Run: func(cmd *cobra.Command, args []string) {
			stateLoader := pkg.NewStateLoader()
			resources, err := stateLoader.LoadStateFromCommand(cmd)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to load Terraform state")
			}

			// Print in a format suitable for shell script
			fmt.Printf("export BUCKET_NAME=\"%s\"\n", resources.DocumentS3Bucket)
			fmt.Printf("export INPUT_QUEUE_URL=\"%s\"\n", resources.InputQueue)
			fmt.Printf("export COMPLETION_QUEUE_URL=\"%s\"\n", resources.CompletionQueue)
			fmt.Printf("export NOTIFICATIONS_QUEUE_URL=\"%s\"\n", resources.NotificationsQueue)
			fmt.Printf("export SNS_TOPIC_ARN=\"%s\"\n", resources.SNSTopic)
			fmt.Printf("export AWS_REGION=\"%s\"\n", resources.Region)
			fmt.Printf("export JOBS_TABLE=\"%s\"\n", resources.JobsTable)
			fmt.Printf("export DOCUMENT_PROCESSOR_ARN=\"%s\"\n", resources.DocumentProcessorARN)
			fmt.Printf("export COMPLETION_PROCESSOR_ARN=\"%s\"\n", resources.CompletionProcessorARN)
			fmt.Printf("export INPUT_DLQ_URL=\"%s\"\n", resources.InputDLQURL)
			fmt.Printf("export COMPLETION_DLQ_URL=\"%s\"\n", resources.CompletionDLQURL)

			// Print helper message
			fmt.Println("\n# To use these variables, run:")
			fmt.Println("# eval $(textractor debug-vars)")
		},
	}
	rootCmd.AddCommand(debugVarsCmd)
}

// Add this new function
func newSaveConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save-config",
		Short: "Save resource configuration to JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			stateLoader := pkg.NewStateLoader()
			resources, err := stateLoader.LoadStateFromCommand(cmd)
			if err != nil {
				return fmt.Errorf("failed to load terraform state: %w", err)
			}

			output, _ := cmd.Flags().GetString("output")
			if output == "" {
				output = "textractor-config.json"
			}

			data, err := json.MarshalIndent(resources, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			if err := os.WriteFile(output, data, 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("Configuration saved to %s\n", output)
			return nil
		},
	}

	cmd.Flags().StringP("output", "o", "", "Output file for the configuration")

	return cmd
}
