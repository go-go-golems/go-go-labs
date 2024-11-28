package debug

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/spf13/cobra"
)

func newLambdaCommand() *cobra.Command {
	var tail bool
	lambdaDebugCmd := &cobra.Command{
		Use:   "lambda",
		Short: "Debug Lambda functions",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging Lambda functions\n")
			fmt.Printf("Document Processor: %s\n", resources.DocumentProcessorName)
			fmt.Printf("Completion Processor: %s\n", resources.CompletionProcessorName)

			// Initialize AWS SDK
			cfg, err := config.LoadDefaultConfig(context.Background())
			if err != nil {
				log.Fatalf("Failed to load AWS config: %v", err)
			}
			client := lambda.NewFromConfig(cfg)

			// Get function configurations
			for _, funcName := range []string{resources.DocumentProcessorName, resources.CompletionProcessorName} {
				fmt.Printf("\n📋 Getting configuration for function: %s\n", funcName)
				output, err := client.GetFunction(context.Background(), &lambda.GetFunctionInput{
					FunctionName: &funcName,
				})
				if err != nil {
					log.Printf("Failed to get function config: %v", err)
					continue
				}

				fmt.Printf("Runtime: %s\n", string(output.Configuration.Runtime))
				fmt.Printf("Memory: %d MB\n", output.Configuration.MemorySize)
				fmt.Printf("Timeout: %d seconds\n", output.Configuration.Timeout)
				fmt.Printf("Last Modified: %s\n", *output.Configuration.LastModified)
			}
		},
	}

	// Add logs subcommand
	logsCmd := &cobra.Command{
		Use:   "logs [processor]",
		Short: "View Lambda function logs",
		Long:  "View Lambda function logs. Processor can be 'document' or 'completion'",
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
				fmt.Printf("📋 Fetching logs for document processor from: %s\n", logGroup)
			case "completion":
				logGroup = resources.CompletionProcessorLogGroup
				fmt.Printf("📋 Fetching logs for completion processor from: %s\n", logGroup)
			default:
				log.Fatalf("Invalid processor type. Must be 'document' or 'completion'")
			}

			cfg, err := config.LoadDefaultConfig(context.Background())
			if err != nil {
				log.Fatalf("Failed to load AWS config: %v", err)
			}

			cwClient := cloudwatchlogs.NewFromConfig(cfg)
			if tail {
				err = streamLogs(cmd.Context(), cwClient, logGroup)
			} else {
				err = fetchRecentLogs(cmd.Context(), cwClient, logGroup)
			}
			if err != nil {
				log.Fatalf("Failed to get logs: %v", err)
			}
		},
	}

	logsCmd.Flags().BoolVar(&tail, "tail", false, "Continuously stream new logs")
	lambdaDebugCmd.AddCommand(logsCmd)

	return lambdaDebugCmd
}

func fetchRecentLogs(ctx context.Context, client *cloudwatchlogs.Client, logGroup string) error {
	// Get log streams, sorted by last event time
	streamsOutput, err := client.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: &logGroup,
		OrderBy:      types.OrderByLastEventTime,
		Descending:   aws.Bool(true),
		Limit:        aws.Int32(10),
	})
	if err != nil {
		return fmt.Errorf("failed to describe log streams: %w", err)
	}

	// Fetch logs from each stream
	for _, stream := range streamsOutput.LogStreams {
		fmt.Printf("\n=== Stream: %s ===\n", *stream.LogStreamName)
		
		input := &cloudwatchlogs.GetLogEventsInput{
			LogGroupName:  &logGroup,
			LogStreamName: stream.LogStreamName,
			StartFromHead: aws.Bool(false),
			Limit:         aws.Int32(100),
		}

		paginator := cloudwatchlogs.NewGetLogEventsPaginator(client, input)
		for paginator.HasMorePages() {
			output, err := paginator.NextPage(ctx)
			if err != nil {
				return fmt.Errorf("failed to get log events: %w", err)
			}

			for _, event := range output.Events {
				timestamp := time.UnixMilli(*event.Timestamp)
				message := *event.Message
				if !strings.HasSuffix(message, "\n") {
					message += "\n"
				}
				fmt.Printf("[%s] %s", timestamp.Format("2006-01-02 15:04:05"), message)
			}
		}
	}
	return nil
}

func streamLogs(ctx context.Context, client *cloudwatchlogs.Client, logGroup string) error {
	// Get the most recent log streams
	streamsOutput, err := client.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: &logGroup,
		OrderBy:      types.OrderByLastEventTime,
		Descending:   aws.Bool(true),
		Limit:        aws.Int32(10),
	})
	if err != nil {
		return fmt.Errorf("failed to describe log streams: %w", err)
	}

	// Keep track of the latest timestamp seen for each stream
	streamTokens := make(map[string]*string)
	for _, stream := range streamsOutput.LogStreams {
		streamTokens[*stream.LogStreamName] = nil
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	fmt.Println("🔄 Streaming logs (Ctrl+C to stop)...")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			// Check for new streams
			newStreamsOutput, err := client.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
				LogGroupName: &logGroup,
				OrderBy:      types.OrderByLastEventTime,
				Descending:   aws.Bool(true),
				Limit:        aws.Int32(10),
			})
			if err != nil {
				return fmt.Errorf("failed to describe log streams: %w", err)
			}

			// Add any new streams we haven't seen before
			for _, stream := range newStreamsOutput.LogStreams {
				if _, exists := streamTokens[*stream.LogStreamName]; !exists {
					streamTokens[*stream.LogStreamName] = nil
				}
			}

			// Get new logs from each stream
			for streamName, nextToken := range streamTokens {
				input := &cloudwatchlogs.GetLogEventsInput{
					LogGroupName:  &logGroup,
					LogStreamName: &streamName,
					StartFromHead: aws.Bool(false),
					NextToken:     nextToken,
				}

				output, err := client.GetLogEvents(ctx, input)
				if err != nil {
					log.Printf("Error getting logs for stream %s: %v", streamName, err)
					continue
				}

				// Store the next token for the next iteration
				streamTokens[streamName] = output.NextForwardToken

				// Print any new events
				for _, event := range output.Events {
					timestamp := time.UnixMilli(*event.Timestamp)
					message := *event.Message
					if !strings.HasSuffix(message, "\n") {
						message += "\n"
					}
					fmt.Printf("[%s] [%s] %s", 
						timestamp.Format("2006-01-02 15:04:05"),
						streamName,
						message)
				}
			}
		}
	}
}
