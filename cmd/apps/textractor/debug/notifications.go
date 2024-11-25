package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func addNotificationsCommand(debugCmd *cobra.Command) {
	debugCmd.AddCommand(&cobra.Command{
		Use:   "notifications",
		Short: "Debug notifications queue",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging notifications queue: %s\n", resources.NotificationsQueue)

			// Get queue attributes
			err = runAWSCommand("sqs", "get-queue-attributes",
				"--queue-url", resources.NotificationsQueue,
				"--attribute-names", "All")
			if err != nil {
				log.Printf("Failed to get queue attributes: %v", err)
			}

			// Try to receive messages
			fmt.Println("\nChecking for messages in the queue:")
			err = runAWSCommand("sqs", "receive-message",
				"--queue-url", resources.NotificationsQueue,
				"--max-number-of-messages", "10",
				"--visibility-timeout", "30",
				"--wait-time-seconds", "5")
			if err != nil {
				log.Printf("Failed to receive messages: %v", err)
			}

			// Check queue metrics
			fmt.Println("\nChecking queue metrics:")
			err = runAWSCommand("cloudwatch", "get-metric-statistics",
				"--namespace", "AWS/SQS",
				"--metric-name", "NumberOfMessagesReceived",
				"--dimensions", fmt.Sprintf("Name=QueueName,Value=%s", resources.NotificationsQueue),
				"--start-time", "$(date -u -v-1H +%Y-%m-%dT%H:%M:%SZ)",
				"--end-time", "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
				"--period", "300",
				"--statistics", "Sum")
			if err != nil {
				log.Printf("Failed to get queue metrics: %v", err)
			}
		},
	})
}
