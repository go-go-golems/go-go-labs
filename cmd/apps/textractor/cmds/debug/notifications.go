package debug

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cloudwatch_types "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/spf13/cobra"
)

func newNotificationsCommand() *cobra.Command {
	notificationsCmd := &cobra.Command{
		Use:   "notifications",
		Short: "Debug notifications queue",
	}

	messagesCmd := &cobra.Command{
		Use:   "messages",
		Short: "Debug messages in notifications queue",
		Run:   runMessagesCommand,
	}
	messagesCmd.Flags().BoolP("poll", "p", false, "Continuously poll for messages")
	messagesCmd.Flags().Int32P("max-messages", "m", 10, "Maximum number of messages to receive")
	messagesCmd.Flags().Int32P("wait-time", "w", 5, "Wait time in seconds for long polling")

	metricsCmd := &cobra.Command{
		Use:   "metrics",
		Short: "Show queue metrics",
		Run:   runMetricsCommand,
	}

	notificationsCmd.AddCommand(messagesCmd, metricsCmd)
	return notificationsCmd
}

func runMessagesCommand(cmd *cobra.Command, args []string) {
	resources, err := LoadResources(cmd)
	if err != nil {
		log.Fatalf("Failed to load resources: %v", err)
	}

	poll, _ := cmd.Flags().GetBool("poll")
	maxMessages, _ := cmd.Flags().GetInt32("max-messages")
	waitTime, _ := cmd.Flags().GetInt32("wait-time")

	cfg, err := loadAWSConfig(cmd.Context())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	fmt.Printf("🔍 Debugging messages in notifications queue: %s\n", resources.NotificationsQueue)

	// Get queue attributes
	attrs, err := sqsClient.GetQueueAttributes(cmd.Context(), &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(resources.NotificationsQueue),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if err != nil {
		log.Printf("Failed to get queue attributes: %v", err)
	} else {
		fmt.Println("\nQueue attributes:")
		for key, value := range attrs.Attributes {
			fmt.Printf("%-35s: %s\n", key, value)
		}
	}

	for {
		// Receive messages
		fmt.Println("\nChecking for messages in the queue:")
		resp, err := sqsClient.ReceiveMessage(cmd.Context(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(resources.NotificationsQueue),
			MaxNumberOfMessages: maxMessages,
			WaitTimeSeconds:     waitTime,
			VisibilityTimeout:   30,
		})
		if err != nil {
			log.Printf("Failed to receive messages: %v", err)
		} else {
			for _, msg := range resp.Messages {
				fmt.Printf("Message ID: %s\nBody: %s\n\n", *msg.MessageId, *msg.Body)
			}
		}

		if !poll {
			break
		}
		time.Sleep(time.Second)
	}
}

func runMetricsCommand(cmd *cobra.Command, args []string) {
	resources, err := LoadResources(cmd)
	if err != nil {
		log.Fatalf("Failed to load resources: %v", err)
	}

	cfg, err := loadAWSConfig(cmd.Context())
	if err != nil {
		log.Fatalf("Failed to load AWS config: %v", err)
	}

	cwClient := cloudwatch.NewFromConfig(cfg)

	endTime := time.Now()
	startTime := endTime.Add(-1 * time.Hour)

	fmt.Println("\nChecking queue metrics:")
	resp, err := cwClient.GetMetricStatistics(cmd.Context(), &cloudwatch.GetMetricStatisticsInput{
		Namespace:  aws.String("AWS/SQS"),
		MetricName: aws.String("NumberOfMessagesReceived"),
		Dimensions: []cloudwatch_types.Dimension{
			{
				Name:  aws.String("QueueName"),
				Value: aws.String(resources.NotificationsQueue),
			},
		},
		StartTime:  aws.Time(startTime),
		EndTime:    aws.Time(endTime),
		Period:     aws.Int32(300),
		Statistics: []cloudwatch_types.Statistic{cloudwatch_types.StatisticSum},
	})
	if err != nil {
		log.Printf("Failed to get queue metrics: %v", err)
	} else {
		fmt.Printf("Metrics: %+v\n", resp.Datapoints)
	}
}

// loadAWSConfig loads the AWS configuration
func loadAWSConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return aws.Config{}, fmt.Errorf("unable to load SDK config: %w", err)
	}
	return cfg, nil
}
