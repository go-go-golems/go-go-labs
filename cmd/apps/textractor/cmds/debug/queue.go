package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func newQueueCommand() *cobra.Command {
	queueDebugCmd := &cobra.Command{
		Use:   "queue",
		Short: "Debug SQS queues",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging SQS queues\n")
			fmt.Printf("Input Queue: %s\n", resources.InputQueue)
			fmt.Printf("Completion Queue: %s\n", resources.CompletionQueue)
			fmt.Printf("Notifications Queue: %s\n", resources.NotificationsQueue)

			// Get queue attributes
			err = runAWSCommand("sqs", "get-queue-attributes",
				"--queue-url", resources.InputQueue,
				"--attribute-names", "All")
			if err != nil {
				log.Printf("Failed to get input queue attributes: %v", err)
			}

			err = runAWSCommand("sqs", "get-queue-attributes",
				"--queue-url", resources.CompletionQueue,
				"--attribute-names", "All")
			if err != nil {
				log.Printf("Failed to get completion queue attributes: %v", err)
			}

			err = runAWSCommand("sqs", "get-queue-attributes",
				"--queue-url", resources.NotificationsQueue,
				"--attribute-names", "All")
			if err != nil {
				log.Printf("Failed to get notifications queue attributes: %v", err)
			}
		},
	}

	return queueDebugCmd
}
