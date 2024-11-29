package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func addTestCommand(debugCmd *cobra.Command) {
	debugCmd.AddCommand(newTestCommand())
}

func newTestCommand() *cobra.Command {
	testDebugCmd := &cobra.Command{
		Use:   "test",
		Short: "Run end-to-end tests",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Running end-to-end tests\n")
			fmt.Printf("Document Processor: %s\n", resources.DocumentProcessorName)
			fmt.Printf("Completion Processor: %s\n", resources.CompletionProcessorName)

			// Test Lambda functions
			fmt.Printf("\nTesting Lambda functions...\n")
			err = runAWSCommand("lambda", "invoke",
				"--function-name", resources.DocumentProcessorName,
				"--payload", `{"test": true}`,
				"/dev/null")
			if err != nil {
				log.Printf("Failed to test document processor: %v", err)
			}

			err = runAWSCommand("lambda", "invoke",
				"--function-name", resources.CompletionProcessorName,
				"--payload", `{"test": true}`,
				"/dev/null")
			if err != nil {
				log.Printf("Failed to test completion processor: %v", err)
			}

			// Test SQS queues
			fmt.Printf("\nTesting SQS queues...\n")
			testQueue(resources.InputQueue, "Test message for input queue")
			testQueue(resources.CompletionQueue, "Test message for completion queue")
			testQueue(resources.NotificationsQueue, "Test message for notifications queue")

			// Test SNS topics
			fmt.Printf("\nTesting SNS topics...\n")
			testTopic(resources.SNSTopic, "Test message for main topic")
			testTopic(resources.NotificationTopic, "Test message for notification topic")
		},
	}

	return testDebugCmd
}

func testQueue(queueURL, message string) {
	err := runAWSCommand("sqs", "send-message",
		"--queue-url", queueURL,
		"--message-body", message)
	if err != nil {
		log.Printf("Failed to send test message to queue %s: %v", queueURL, err)
		return
	}

	err = runAWSCommand("sqs", "receive-message",
		"--queue-url", queueURL,
		"--wait-time-seconds", "5")
	if err != nil {
		log.Printf("Failed to receive test message from queue %s: %v", queueURL, err)
	}
}

func testTopic(topicARN, message string) {
	err := runAWSCommand("sns", "publish",
		"--topic-arn", topicARN,
		"--message", message)
	if err != nil {
		log.Printf("Failed to publish test message to topic %s: %v", topicARN, err)
	}
}
