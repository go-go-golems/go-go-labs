package debug

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

func addSubmitFlowCommand(debugCmd *cobra.Command) {
	debugCmd.AddCommand(&cobra.Command{
		Use:   "submit-flow",
		Short: "Debug submit command flow",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Println("🔍 Debugging submit command flow")

			// Check S3 bucket configuration
			fmt.Printf("\nChecking input S3 bucket configuration:\n")
			err = runAWSCommand("s3api", "get-bucket-versioning",
				"--bucket", resources.DocumentS3Bucket)
			if err != nil {
				log.Printf("Failed to get bucket versioning: %v", err)
			}

			// Check Lambda configurations
			fmt.Printf("\nChecking Lambda configurations:\n")
			for _, function := range []string{resources.DocumentProcessorName, resources.CompletionProcessorName} {
				err = runAWSCommand("lambda", "get-function-configuration",
					"--function-name", function)
				if err != nil {
					log.Printf("Failed to get Lambda configuration for %s: %v", function, err)
				}
			}

			// Check SQS queues
			fmt.Printf("\nChecking SQS queues:\n")
			for _, queue := range []string{resources.InputQueue, resources.CompletionQueue} {
				err = runAWSCommand("sqs", "get-queue-attributes",
					"--queue-url", queue,
					"--attribute-names", "All")
				if err != nil {
					log.Printf("Failed to get queue attributes for %s: %v", queue, err)
				}
			}

			// Monitor Lambda logs for a short period
			fmt.Printf("\nMonitoring Lambda logs for 30 seconds:\n")
			startTime := time.Now().Add(-5 * time.Minute).Unix()
			for _, function := range []string{resources.DocumentProcessorName, resources.CompletionProcessorName} {
				err = runAWSCommand("logs", "filter-log-events",
					"--log-group-name", fmt.Sprintf("/aws/lambda/%s", function),
					"--start-time", fmt.Sprintf("%d", startTime*1000))
				if err != nil {
					log.Printf("Failed to get logs for %s: %v", function, err)
				}
			}
		},
	})
}
