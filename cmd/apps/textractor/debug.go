package main

import (
	"fmt"
	"log"
	"os/exec"
	"context"
	"os"
	"time"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func addDebugCommands(rootCmd *cobra.Command) {
	debugCmd := &cobra.Command{
		Use:   "debug",
		Short: "Debug commands for Textractor resources",
	}
	rootCmd.AddCommand(debugCmd)

	// Lambda debugging with processor selection
	debugCmd.AddCommand(&cobra.Command{
		Use:   "lambda [document|completion]",
		Short: "Debug Lambda functions",
		Long: `Debug Lambda functions. Specify which processor to debug:
  document    - Document processor Lambda (handles new uploads)
  completion  - Completion processor Lambda (handles Textract completion)`,
		Args:  cobra.MaximumNArgs(1),
		Run:   debugLambda,
	})

	// Queue debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "queue [input|completion]",
		Short: "Debug SQS queues",
		Args:  cobra.ExactArgs(1),
		Run:   debugQueue,
	})

	// S3 debugging
	s3DebugCmd := &cobra.Command{
		Use:   "s3",
		Short: "Debug S3 bucket configuration",
		Run:   debugS3,
	}
	debugCmd.AddCommand(s3DebugCmd)

	s3DebugCmd.AddCommand(&cobra.Command{
		Use:   "ls [prefix]",
		Short: "List files in S3 bucket",
		Long:  "List files in S3 bucket. Optionally specify a prefix to filter results",
		Args:  cobra.MaximumNArgs(1),
		Run:   debugS3List,
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

	// Add notifications queue debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "notifications",
		Short: "Debug notifications queue",
		Run:   debugNotifications,
	})

	// Add DLQ debugging
	debugCmd.AddCommand(&cobra.Command{
		Use:   "dlq [input|completion]",
		Short: "Debug Dead Letter Queues",
		Long: `Debug Dead Letter Queues. Specify which DLQ to debug:
  input       - Input queue DLQ
  completion  - Completion queue DLQ`,
		Args:  cobra.ExactArgs(1),
		Run:   debugDLQ,
	})

	// Add Textract job status command
	debugCmd.AddCommand(&cobra.Command{
		Use:   "textract-job [jobId]",
		Short: "Check status of a Textract job",
		Long:  "Shows the current status and details of a Textract document analysis job",
		Args:  cobra.ExactArgs(1),
		Run:   debugTextractJob,
	})

	// Add output S3 debugging
	outputS3DebugCmd := &cobra.Command{
		Use:   "output-s3",
		Short: "Debug output S3 bucket configuration",
		Run:   debugOutputS3,
	}
	debugCmd.AddCommand(outputS3DebugCmd)

	outputS3DebugCmd.AddCommand(&cobra.Command{
		Use:   "ls [prefix]",
		Short: "List files in output S3 bucket",
		Long:  "List files in output S3 bucket. Optionally specify a prefix to filter results",
		Args:  cobra.MaximumNArgs(1),
		Run:   debugOutputS3List,
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

	// Allow specifying which processor to debug
	processor := "document"
	if len(args) > 0 {
		processor = args[0]
	}

	functionName := resources.DocumentProcessorName
	if processor == "completion" {
		functionName = resources.CompletionProcessorName
	}

	fmt.Printf("🔍 Debugging %s processor Lambda: %s\n", processor, functionName)
	
	// Get recent logs
	err = runAWSCommand("logs", "tail", 
		fmt.Sprintf("/aws/lambda/%s", functionName),
		"--follow")
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
	if args[0] == "completion" {
		queueURL = resources.CompletionQueue
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

func debugS3List(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Printf("📂 Listing files in bucket: %s\n", resources.S3Bucket)

	lsArgs := []string{"s3", "ls", fmt.Sprintf("s3://%s", resources.S3Bucket)}
	if len(args) > 0 {
		lsArgs = append(lsArgs, fmt.Sprintf("s3://%s/%s", resources.S3Bucket, args[0]))
	}
	lsArgs = append(lsArgs, "--recursive")

	err = runAWSCommand(lsArgs...)
	if err != nil {
		log.Printf("Failed to list bucket contents: %v", err)
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

	// Document processor metrics
	err = runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", "AWS/Lambda",
		"--metric-name", "Duration",
		"--statistics", "Average", "Maximum",
		"--period", "300",
		"--dimensions", 
		fmt.Sprintf("Name=FunctionName,Value=%s", resources.DocumentProcessorName))
	if err != nil {
		log.Printf("Failed to get document processor metrics: %v", err)
	}

	// Completion processor metrics
	err = runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", "AWS/Lambda",
		"--metric-name", "Duration",
		"--statistics", "Average", "Maximum",
		"--period", "300",
		"--dimensions",
		fmt.Sprintf("Name=FunctionName,Value=%s", resources.CompletionProcessorName))
	if err != nil {
		log.Printf("Failed to get completion processor metrics: %v", err)
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

	// Start monitoring both processors and notifications in parallel
	errChan := make(chan error, 3)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Monitor document processor logs
	go func() {
		cmd := exec.CommandContext(ctx, "aws", "logs", "tail",
			fmt.Sprintf("/aws/lambda/%s", resources.DocumentProcessorName),
			"--follow")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		errChan <- cmd.Run()
	}()

	// Monitor completion processor logs
	go func() {
		cmd := exec.CommandContext(ctx, "aws", "logs", "tail",
			fmt.Sprintf("/aws/lambda/%s", resources.CompletionProcessorName),
			"--follow")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		errChan <- cmd.Run()
	}()

	// Monitor notifications queue
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				errChan <- nil
				return
			case <-ticker.C:
				cmd := exec.Command("aws", "sqs", "receive-message",
					"--queue-url", resources.NotificationsQueue,
					"--max-number-of-messages", "10",
					"--wait-time-seconds", "5",
					"--attribute-names", "All")
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					errChan <- fmt.Errorf("failed to receive notifications: %v", err)
					return
				}
			}
		}
	}()

	// Wait for user interrupt
	fmt.Println("\n📋 Monitoring processing (Ctrl+C to stop)...")
	fmt.Println("- Document processor logs")
	fmt.Println("- Completion processor logs")
	fmt.Println("- Notifications queue")

	// Handle interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		fmt.Println("\n🛑 Stopping monitoring...")
		cancel()
	case err := <-errChan:
		if err != nil {
			log.Printf("Error during monitoring: %v", err)
		}
		cancel()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 3; i++ {
		if err := <-errChan; err != nil {
			log.Printf("Error during shutdown: %v", err)
		}
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

func debugNotifications(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Println("🔔 Debugging notifications system")

	// Check SNS topic
	fmt.Println("\n📢 SNS Topic Configuration:")
	err = runAWSCommand("sns", "get-topic-attributes",
		"--topic-arn", resources.NotificationTopic)
	if err != nil {
		log.Printf("Failed to get topic attributes: %v", err)
	}

	// Check SQS subscription
	fmt.Println("\n📬 SQS Queue Messages:")
	err = runAWSCommand("sqs", "receive-message",
		"--queue-url", resources.NotificationsQueue,
		"--max-number-of-messages", "10",
		"--wait-time-seconds", "20",
		"--attribute-names", "All")
	if err != nil {
		log.Printf("Failed to receive messages: %v", err)
	}
}

func debugDLQ(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	queueType := args[0]
	var queueURL string
	
	switch queueType {
	case "input":
		queueURL = resources.InputDLQURL
	case "completion":
		queueURL = resources.CompletionDLQURL
	default:
		log.Fatalf("Invalid queue type: %s. Must be 'input' or 'completion'", queueType)
	}

	fmt.Printf("🔍 Debugging %s Dead Letter Queue\n", queueType)

	// Get queue attributes
	fmt.Println("\n📊 Queue Attributes:")
	err = runAWSCommand("sqs", "get-queue-attributes",
		"--queue-url", queueURL,
		"--attribute-names", "All")
	if err != nil {
		log.Printf("Failed to get queue attributes: %v", err)
	}

	// Show messages in DLQ
	fmt.Println("\n📬 Messages in DLQ:")
	err = runAWSCommand("sqs", "receive-message",
		"--queue-url", queueURL,
		"--max-number-of-messages", "10",
		"--wait-time-seconds", "5",
		"--attribute-names", "All",
		"--message-attribute-names", "All")
	if err != nil {
		log.Printf("Failed to receive messages: %v", err)
	}

	// Show CloudWatch metrics for the DLQ
	fmt.Println("\n📈 CloudWatch Metrics (last hour):")
	err = runAWSCommand("cloudwatch", "get-metric-statistics",
		"--namespace", "AWS/SQS",
		"--metric-name", "ApproximateNumberOfMessagesVisible",
		"--statistics", "Sum",
		"--period", "300",
		"--start-time", time.Now().Add(-1*time.Hour).Format(time.RFC3339),
		"--end-time", time.Now().Format(time.RFC3339),
		"--dimensions", fmt.Sprintf("Name=QueueName,Value=%s", queueURL))
	if err != nil {
		log.Printf("Failed to get CloudWatch metrics: %v", err)
	}

	fmt.Println("\n💡 Tips:")
	fmt.Println("- Use 'aws sqs purge-queue --queue-url <url>' to clear the DLQ")
	fmt.Println("- Use 'aws sqs delete-message' to remove individual messages")
	fmt.Println("- Check CloudWatch logs for more details about failures")
}

func debugTextractJob(cmd *cobra.Command, args []string) {
	jobId := args[0]
	fmt.Printf("🔍 Checking Textract job status for: %s\n", jobId)

	err := runAWSCommand("textract", "get-document-analysis",
		"--job-id", jobId,
		"--max-results", "1")
	if err != nil {
		log.Printf("Failed to get Textract job status: %v", err)
		return
	}
}

func debugOutputS3(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Println("🔍 Debugging output S3 bucket:", resources.OutputS3Bucket)

	// Get bucket notification configuration
	err = runAWSCommand("s3api", "get-bucket-notification-configuration",
		"--bucket", resources.OutputS3Bucket)
	if err != nil {
		log.Printf("Failed to get bucket notification configuration: %v", err)
	}

	// List recent CloudTrail events
	err = runAWSCommand("cloudtrail", "lookup-events",
		"--lookup-attributes", fmt.Sprintf("AttributeKey=ResourceName,AttributeValue=%s", resources.OutputS3Bucket))
	if err != nil {
		log.Printf("Failed to get CloudTrail events: %v", err)
	}
}

func debugOutputS3List(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	fmt.Printf("📂 Listing files in output bucket: %s\n", resources.OutputS3Bucket)

	lsArgs := []string{"s3", "ls", fmt.Sprintf("s3://%s", resources.OutputS3Bucket)}
	if len(args) > 0 {
		lsArgs = append(lsArgs, fmt.Sprintf("s3://%s/%s", resources.OutputS3Bucket, args[0]))
	}
	lsArgs = append(lsArgs, "--recursive")

	err = runAWSCommand(lsArgs...)
	if err != nil {
		log.Printf("Failed to list bucket contents: %v", err)
	}
} 