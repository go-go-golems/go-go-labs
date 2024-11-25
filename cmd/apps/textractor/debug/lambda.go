package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func newLambdaCommand() *cobra.Command {
	lambdaDebugCmd := &cobra.Command{
		Use:   "lambda",
		Short: "Debug Lambda functions",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging Lambda functions\n")
			fmt.Printf("Document Processor: %s\n", resources.DocumentProcessorName)
			fmt.Printf("Completion Processor: %s\n", resources.CompletionProcessorName)

			// Get function configurations
			err = runAWSCommand("lambda", "get-function",
				"--function-name", resources.DocumentProcessorName)
			if err != nil {
				log.Printf("Failed to get document processor config: %v", err)
			}

			err = runAWSCommand("lambda", "get-function",
				"--function-name", resources.CompletionProcessorName)
			if err != nil {
				log.Printf("Failed to get completion processor config: %v", err)
			}
		},
	}

	// Add logs subcommand
	lambdaDebugCmd.AddCommand(&cobra.Command{
		Use:   "logs [processor]",
		Short: "View Lambda function logs",
		Long:  "View Lambda function logs. Processor can be 'document' or 'completion'",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			var logGroup string
			switch args[0] {
			case "document":
				logGroup = resources.DocumentProcessorLogGroup
			case "completion":
				logGroup = resources.CompletionProcessorLogGroup
			default:
				log.Fatalf("Invalid processor type. Must be 'document' or 'completion'")
			}

			err = runAWSCommand("logs", "tail", logGroup)
			if err != nil {
				log.Fatalf("Failed to get logs: %v", err)
			}
		},
	})

	return lambdaDebugCmd
}
