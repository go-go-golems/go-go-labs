package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func newSNSCommand() *cobra.Command {
	snsDebugCmd := &cobra.Command{
		Use:   "sns",
		Short: "Debug SNS topic configuration",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging SNS topics\n")
			fmt.Printf("Main Topic: %s\n", resources.SNSTopic)
			fmt.Printf("Notification Topic: %s\n", resources.NotificationTopic)

			// Get topic attributes
			err = runAWSCommand("sns", "get-topic-attributes",
				"--topic-arn", resources.SNSTopic)
			if err != nil {
				log.Printf("Failed to get main topic attributes: %v", err)
			}

			err = runAWSCommand("sns", "get-topic-attributes",
				"--topic-arn", resources.NotificationTopic)
			if err != nil {
				log.Printf("Failed to get notification topic attributes: %v", err)
			}

			// List subscriptions
			err = runAWSCommand("sns", "list-subscriptions-by-topic",
				"--topic-arn", resources.SNSTopic)
			if err != nil {
				log.Printf("Failed to list main topic subscriptions: %v", err)
			}

			err = runAWSCommand("sns", "list-subscriptions-by-topic",
				"--topic-arn", resources.NotificationTopic)
			if err != nil {
				log.Printf("Failed to list notification topic subscriptions: %v", err)
			}
		},
	}

	return snsDebugCmd
}
