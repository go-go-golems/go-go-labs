package main

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

func addDebugCommands(rootCmd *cobra.Command) {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug commands for Textractor resources",
	}
	rootCmd.AddCommand(debugCmd)

	// Lambda debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "lambda",
		Short: "Debug Lambda function",
		Run:   debugLambda,
	})

	// Queue debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "queue [input|output]",
		Short: "Debug SQS queues",
		Args:  cobra.ExactArgs(1),
		Run:   debugQueue,
	})

	// S3 debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "s3",
		Short: "Debug S3 bucket configuration",
		Run:   debugS3,
	})

	// SNS debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "sns",
		Short: "Debug SNS topic",
		Run:   debugSNS,
	})

	// CloudWatch metrics
	debugCmd.AddCommand(&cobra.Command{
		Use:   "metrics",
		Short: "Show CloudWatch metrics",
		Run:   debugMetrics,
	})

	// End-to-end test
	debugCmd.AddCommand(&cobra.Command{
		Use:   "test [pdf-file]",
		Short: "Run end-to-end test with PDF file",
		Args:  cobra.ExactArgs(1),
		Run:   debugTest,
	})

	// Submit debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "submit-flow",
		Short: "Debug submit command flow",
		Run:   debugSubmitFlow,
	})
}

func runAWSCommand(args ...string) error {
	cmd := exec.Command("aws", args...)
	cmd.Stdout = log.Writer()
	cmd.Stderr = log.Writer()
	return cmd.Run()
}

func debugLambda(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Println("🔍 Debugging Lambda function:", resources.FunctionName)
	
	// Get recent logs
	err = runAWSCommand("logs", "tail", fmt.Sprintf("/aws/lambda/%s", resources.FunctionName), "--follow")
	if err != nil {
		log.Printf("Failed to get logs: %v", err)
	}
}

func debugQueue(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	queueURL := resources.InputQueue
	if args[0] == "output" {
		queueURL = resources.OutputQueue
	}

	fmt.Printf("🔍 Debugging %s queue\n", args[0])

	// Get queue attributes
	err = runAWSCommand("sqs", "get-queue-attributes",
		"--queue-url", queueURL,
		"--attribute-names", "All")
	if err != nil {
		log.Printf("Failed to get queue attributes: %v", err)
	}

	// Receive messages
	err = runAWSCommand("sqs", "receive-message",
		"--queue-url", queueURL,
		"--max-number-of-messages", "10",
		"--wait-time-seconds", "5")
	if err != nil {
		log.Printf("Failed to receive messages: %v", err)
	}
}

func debugS3(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Println("🔍 Debugging S3 bucket:", resources.S3Bucket)

	// Get bucket notification configuration
	err = runAWSCommand("s3api", "get-bucket-notification-configuration",
		"--bucket", resources.S3Bucket)
	if err != nil {
		log.Printf("Failed to get bucket notification configuration: %v", err)
	}

	// List recent CloudTrail events
	err = runAWSCommand("cloudtrail", "lookup-events",
		"--lookup-attributes", fmt.Sprintf("AttributeKey=ResourceName,AttributeValue=%s", resources.S3Bucket))
	if err != nil {
		log.Printf("Failed to get CloudTrail events: %v", err)
	}
}

func debugSNS(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Println("🔍 Debugging SNS topic:", resources.SNSTopic)

	// List subscriptions
	err = runAWSCommand("sns", "list-subscriptions-by-topic",
		"--topic-arn", resources.SNSTopic)
	if err != nil {
		log.Printf("Failed to list subscriptions: %v", err)
	}

	// Get topic attributes
	err = runAWSCommand("sns", "get-topic-attributes",
		"--topic-arn", resources.SNSTopic)
	if err != nil {
		log.Printf("Failed to get topic attributes: %v", err)
	}
}

func debugMetrics(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Println("📊 Getting CloudWatch metrics")

	// Lambda metrics
	err = runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", "AWS/Lambda",
		"--metric-name", "Duration",
		"--statistics", "Average", "Maximum",
		"--period", "300",
		"--dimensions", fmt.Sprintf("Name=FunctionName,Value=%s", resources.FunctionName))
	if err != nil {
		log.Printf("Failed to get Lambda metrics: %v", err)
	}

	// SQS metrics
	err = runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", "AWS/SQS",
		"--metric-name", "ApproximateNumberOfMessagesVisible",
		"--statistics", "Average",
		"--period", "300",
		"--dimensions", "Name=QueueName,Value="+resources.InputQueue)
	if err != nil {
		log.Printf("Failed to get SQS metrics: %v", err)
	}
}

func debugTest(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	pdfFile := args[0]
	fmt.Printf("🧪 Running end-to-end test with file: %s\n", pdfFile)

	// Upload file
	err = runAWSCommand("s3", "cp",
		pdfFile,
		fmt.Sprintf("s3://%s/input/", resources.S3Bucket))
	if err != nil {
		log.Printf("Failed to upload file: %v", err)
		return
	}

	fmt.Println("✅ File uploaded, monitoring processing...")

	// Monitor Lambda logs
	err = runAWSCommand("logs", "tail",
		fmt.Sprintf("/aws/lambda/%s", resources.FunctionName),
		"--follow")
	if err != nil {
		log.Printf("Failed to get Lambda logs: %v", err)
	}
}

func debugSubmitFlow(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Println("🔍 Debugging submit flow")

	// Check S3 bucket permissions
	err = runAWSCommand("s3api", "get-bucket-acl",
		"--bucket", resources.S3Bucket)
	if err != nil {
		log.Printf("Failed to check S3 bucket ACL: %v", err)
	}

	// Check DynamoDB table
	err = runAWSCommand("dynamodb", "describe-table",
		"--table-name", resources.JobsTable)
	if err != nil {
		log.Printf("Failed to describe DynamoDB table: %v", err)
	}

	// Check SQS permissions
	err = runAWSCommand("sqs", "get-queue-attributes",
		"--queue-url", resources.InputQueue,
		"--attribute-names", "Policy")
	if err != nil {
		log.Printf("Failed to check SQS policy: %v", err)
	}
} 