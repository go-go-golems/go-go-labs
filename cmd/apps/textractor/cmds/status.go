package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"
	"github.com/spf13/cobra"
)

func NewStatusCommand() *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "status [jobID]",
		Short: "Check the status of a submitted job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := args[0]

			// Load resources
			stateLoader := pkg.NewStateLoader()
			resources, err := stateLoader.LoadStateFromCommand(cmd)
			if err != nil {
				return fmt.Errorf("failed to load terraform state: %w", err)
			}

			// Initialize AWS session
			sess := session.Must(session.NewSession(&aws.Config{
				Region: aws.String(resources.Region),
			}))

			// Create DynamoDB client
			dbClient := dynamodb.New(sess)

			// Get job details
			result, err := dbClient.GetItem(&dynamodb.GetItemInput{
				TableName: aws.String(resources.JobsTable),
				Key: map[string]*dynamodb.AttributeValue{
					"JobID": {
						S: aws.String(jobID),
					},
				},
			})
			if err != nil {
				return fmt.Errorf("failed to get job status: %w", err)
			}

			if result.Item == nil {
				return fmt.Errorf("job %s not found", jobID)
			}

			var job pkg.TextractJob
			err = dynamodbattribute.UnmarshalMap(result.Item, &job)
			if err != nil {
				return fmt.Errorf("failed to unmarshal job data: %w", err)
			}

			// Display job information based on format
			switch outputFormat {
			case "json":
				jsonData, err := json.MarshalIndent(job, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to marshal job data to JSON: %w", err)
				}
				fmt.Println(string(jsonData))
			default:
				fmt.Printf("Job ID: %s\n", job.JobID)
				fmt.Printf("Status: %s\n", job.Status)
				fmt.Printf("Document Key: %s\n", job.DocumentKey)
				fmt.Printf("Submitted At: %s\n", job.SubmittedAt)
				if job.CompletedAt != nil {
					fmt.Printf("Completed At: %s\n", job.CompletedAt)
				}
				if job.Error != "" {
					fmt.Printf("Error: %s\n", job.Error)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text/json)")
	return cmd
}
