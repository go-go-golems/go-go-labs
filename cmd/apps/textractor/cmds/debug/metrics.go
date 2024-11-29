package debug

import (
	"fmt"
	"log"
	"time"

	"github.com/spf13/cobra"
)

func newMetricsCommand() *cobra.Command {
	metricsDebugCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Debug CloudWatch metrics",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging CloudWatch metrics\n")
			fmt.Printf("Document Processor: %s\n", resources.DocumentProcessorName)
			fmt.Printf("Completion Processor: %s\n", resources.CompletionProcessorName)

			// Get Lambda metrics
			getMetrics(resources.DocumentProcessorName, "AWS/Lambda")
			getMetrics(resources.CompletionProcessorName, "AWS/Lambda")

			// Get SQS metrics
			getQueueMetrics(resources.InputQueue)
			getQueueMetrics(resources.CompletionQueue)
			getQueueMetrics(resources.NotificationsQueue)
		},
	}

	// Add logs subcommand
	metricsDebugCmd.AddCommand(&cobra.Command{
		Use:   "logs [processor]",
		Short: "View CloudWatch log groups",
		Long:  "View CloudWatch log groups. Processor can be 'document', 'completion', or 'cloudtrail'",
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
			case "cloudtrail":
				logGroup = resources.CloudTrailLogGroup
			default:
				log.Fatalf("Invalid processor type. Must be 'document', 'completion', or 'cloudtrail'")
			}

			err = runAWSCommand("logs", "tail", logGroup)
			if err != nil {
				log.Printf("Failed to get logs: %v", err)
			}
		},
	})

	return metricsDebugCmd
}

func getMetrics(functionName, namespace string) {
	// Get the last hour of metrics
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	err := runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", namespace,
		"--metric-name", "Invocations",
		"--dimensions", fmt.Sprintf("Name=FunctionName,Value=%s", functionName),
		"--start-time", startTime.Format(time.RFC3339),
		"--end-time", endTime.Format(time.RFC3339),
		"--period", "300",
		"--statistics", "Sum")
	if err != nil {
		log.Printf("Failed to get metrics for %s: %v", functionName, err)
	}

	err = runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", namespace,
		"--metric-name", "Errors",
		"--dimensions", fmt.Sprintf("Name=FunctionName,Value=%s", functionName),
		"--start-time", startTime.Format(time.RFC3339),
		"--end-time", endTime.Format(time.RFC3339),
		"--period", "300",
		"--statistics", "Sum")
	if err != nil {
		log.Printf("Failed to get error metrics for %s: %v", functionName, err)
	}
}

func getQueueMetrics(queueURL string) {
	// Get the last hour of metrics
	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	err := runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", "AWS/SQS",
		"--metric-name", "NumberOfMessagesReceived",
		"--dimensions", fmt.Sprintf("Name=QueueName,Value=%s", queueURL),
		"--start-time", startTime.Format(time.RFC3339),
		"--end-time", endTime.Format(time.RFC3339),
		"--period", "300",
		"--statistics", "Sum")
	if err != nil {
		log.Printf("Failed to get metrics for queue %s: %v", queueURL, err)
	}
}
