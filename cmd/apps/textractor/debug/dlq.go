package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func addDLQCommand(debugCmd *cobra.Command) {
	debugCmd.AddCommand(&cobra.Command{
		Use:   "dlq [input|completion]",
		Short: "Debug Dead Letter Queues",
		Long: `Debug Dead Letter Queues. Specify which DLQ to debug:
  input       - Input queue DLQ
  completion  - Completion queue DLQ`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			queueURL := resources.InputDLQURL
			if args[0] == "completion" {
				queueURL = resources.CompletionDLQURL
			}

			fmt.Printf("🔍 Debugging %s DLQ\n", args[0])

			// Get queue attributes
			err = runAWSCommand("sqs", "get-queue-attributes",
				"--queue-url", queueURL,
				"--attribute-names", "All")
			if err != nil {
				log.Printf("Failed to get queue attributes: %v", err)
			}

			// Try to receive messages
			fmt.Println("\nChecking for messages in the DLQ:")
			err = runAWSCommand("sqs", "receive-message",
				"--queue-url", queueURL,
				"--max-number-of-messages", "10",
				"--visibility-timeout", "30",
				"--wait-time-seconds", "5",
				"--attribute-names", "All",
				"--message-attribute-names", "All")
			if err != nil {
				log.Printf("Failed to receive messages: %v", err)
			}

			// Get approximate number of messages
			err = runAWSCommand("sqs", "get-queue-attributes",
				"--queue-url", queueURL,
				"--attribute-names", "ApproximateNumberOfMessages",
				"ApproximateNumberOfMessagesNotVisible",
				"ApproximateNumberOfMessagesDelayed")
			if err != nil {
				log.Printf("Failed to get message counts: %v", err)
			}
		},
	})
}
