package cmds

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg/utils"
	"github.com/spf13/cobra"
)

func NewListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Textract jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Parse flags
			since, _ := cmd.Flags().GetString("since")
			status, _ := cmd.Flags().GetString("status")

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

			// Build filter expression
			var filterExpressions []string
			expressionValues := map[string]*dynamodb.AttributeValue{}

			if since != "" {
				t, err := time.Parse(time.RFC3339, since)
				if err != nil {
					return fmt.Errorf("invalid time format for --since flag: %w", err)
				}
				filterExpressions = append(filterExpressions, "SubmittedAt >= :t")
				expressionValues[":t"] = &dynamodb.AttributeValue{S: aws.String(t.Format(time.RFC3339))}
			}

			if status != "" {
				filterExpressions = append(filterExpressions, "#statusAlias = :s")
				expressionValues[":s"] = &dynamodb.AttributeValue{S: aws.String(status)}
				input.ExpressionAttributeNames = map[string]*string{
					"#statusAlias": aws.String("Status"),
				}
			}

			if len(filterExpressions) > 0 {
				input.FilterExpression = aws.String(strings.Join(filterExpressions, " AND "))
				input.ExpressionAttributeValues = expressionValues
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
				fmt.Printf("  Textract ID: %s\n", job.TextractID)
				fmt.Printf("  Result Key: %s\n", job.ResultKey)
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
	cmd.Flags().String("status", "", "Filter by job status (UPLOADING, SUBMITTED, PROCESSING, COMPLETED, FAILED, ERROR)")
	return cmd
}
