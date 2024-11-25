package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/debug"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/utils"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "textractor",
		Short: "Manage Textractor AWS resources and process PDFs",
	}

	// Add persistent flags to root command
	rootCmd.PersistentFlags().String("tf-dir", "terraform", "Directory containing Terraform state")
	rootCmd.PersistentFlags().String("config", "", "JSON config file containing resource configuration")

	// Add commands
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newSubmitCommand())
	rootCmd.AddCommand(debug.NewDebugCommand())

	addDebugVarCommands(rootCmd, "terraform")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Textract jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse flags
			since, _ := cmd.Flags().GetString("since")

			stateLoader := utils.NewStateLoader()
			resources, err := stateLoader.LoadStateFromCommand(cmd)
			if err != nil {
				return fmt.Errorf("failed to load state: %w", err)
			}

			// Initialize AWS session and DynamoDB client
			sess := session.Must(session.NewSession(&aws.Config{
				Region: aws.String(resources.Region),
			}))
			db := dynamodb.New(sess)

			// Query jobs from DynamoDB
			input := &dynamodb.ScanInput{
				TableName: aws.String(resources.JobsTable),
			}

			if since != "" {
				t, err := time.Parse(time.RFC3339, since)
				if err != nil {
					return fmt.Errorf("invalid time format for --since flag: %w", err)
				}
				input.FilterExpression = aws.String("SubmittedAt >= :t")
				input.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
					":t": {S: aws.String(t.Format(time.RFC3339))},
				}
			}

			result, err := db.Scan(input)
			if err != nil {
				return fmt.Errorf("failed to query jobs: %w", err)
			}

			var jobs []utils.TextractJob
			if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &jobs); err != nil {
				return fmt.Errorf("failed to unmarshal jobs: %w", err)
			}

			// Sort jobs by submission time
			sort.Slice(jobs, func(i, j int) bool {
				return jobs[i].SubmittedAt.After(jobs[j].SubmittedAt)
			})

			// Print jobs
			for _, job := range jobs {
				fmt.Printf("Job ID: %s\n", job.JobID)
				fmt.Printf("  Document: %s\n", job.DocumentKey)
				fmt.Printf("  Status: %s\n", job.Status)
				fmt.Printf("  Submitted: %s\n", job.SubmittedAt.Format(time.RFC3339))
				if job.CompletedAt != nil {
					fmt.Printf("  Completed: %s\n", job.CompletedAt.Format(time.RFC3339))
				}
				if job.Error != "" {
					fmt.Printf("  Error: %s\n", job.Error)
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().String("since", "", "Only show jobs submitted after this time (RFC3339 format)")
	return cmd
}
func addDebugVarCommands(rootCmd *cobra.Command, tfDir string) {
	debugVarsCmd := &cobra.Command{
		Use:   "debug-vars",
		Short: "Print environment variables for debugging",
		Run: func(cmd *cobra.Command, args []string) {
			stateLoader := utils.NewStateLoader()
			resources, err := stateLoader.LoadStateFromCommand(cmd)
			if err != nil {
				log.Fatalf("Failed to load Terraform state: %v", err)
			}

			// Print in a format suitable for shell script
			fmt.Printf("export BUCKET_NAME=\"%s\"\n", resources.DocumentS3Bucket)
			fmt.Printf("export INPUT_QUEUE_URL=\"%s\"\n", resources.InputQueue)
			fmt.Printf("export COMPLETION_QUEUE_URL=\"%s\"\n", resources.CompletionQueue)
			fmt.Printf("export NOTIFICATIONS_QUEUE_URL=\"%s\"\n", resources.NotificationsQueue)
			fmt.Printf("export SNS_TOPIC_ARN=\"%s\"\n", resources.SNSTopic)
			fmt.Printf("export AWS_REGION=\"%s\"\n", resources.Region)
			fmt.Printf("export JOBS_TABLE=\"%s\"\n", resources.JobsTable)
			fmt.Printf("export DOCUMENT_PROCESSOR_ARN=\"%s\"\n", resources.DocumentProcessorARN)
			fmt.Printf("export COMPLETION_PROCESSOR_ARN=\"%s\"\n", resources.CompletionProcessorARN)
			fmt.Printf("export INPUT_DLQ_URL=\"%s\"\n", resources.InputDLQURL)
			fmt.Printf("export COMPLETION_DLQ_URL=\"%s\"\n", resources.CompletionDLQURL)

			// Print helper message
			fmt.Println("\n# To use these variables, run:")
			fmt.Println("# eval $(textractor debug-vars)")
		},
	}
	rootCmd.AddCommand(debugVarsCmd)
}
