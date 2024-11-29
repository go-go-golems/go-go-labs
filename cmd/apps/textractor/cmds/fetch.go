package cmds

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg/utils"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
)

func NewFetchCommand() *cobra.Command {
	var (
		outputFormat string
		outputFile   string
	)

	cmd := &cobra.Command{
		Use:   "fetch [jobID]",
		Short: "Fetch the results of a processed job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			jobID := args[0]

			// Load resources
			stateLoader := utils.NewStateLoader()
			resources, err := stateLoader.LoadStateFromCommand(cmd)
			if err != nil {
				return fmt.Errorf("failed to load terraform state: %w", err)
			}

			// Initialize AWS session
			sess := session.Must(session.NewSession(&aws.Config{
				Region: aws.String(resources.Region),
			}))

			// Create service clients
			dbClient := dynamodb.New(sess)
			s3Client := s3.New(sess)

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
				return fmt.Errorf("failed to get job details: %w", err)
			}

			if result.Item == nil {
				return fmt.Errorf("job %s not found", jobID)
			}

			var job utils.TextractJob
			err = dynamodbattribute.UnmarshalMap(result.Item, &job)
			if err != nil {
				return fmt.Errorf("failed to unmarshal job data: %w", err)
			}

			if job.Status != "COMPLETED" {
				return fmt.Errorf("job is not completed (current status: %s)", job.Status)
			}

			// Construct the output S3 key
			outputKey := fmt.Sprintf("output/%s/result.json", jobID)

			// Get the results from S3
			output, err := s3Client.GetObject(&s3.GetObjectInput{
				Bucket: aws.String(resources.ResultsS3Bucket),
				Key:    aws.String(outputKey),
			})
			if err != nil {
				return fmt.Errorf("failed to get results from S3: %w", err)
			}
			defer output.Body.Close()

			// Read the results
			data, err := io.ReadAll(output.Body)
			if err != nil {
				return fmt.Errorf("failed to read results: %w", err)
			}

			// Process and output the results based on format
			switch outputFormat {
			case "json":
				// Parse JSON to ensure it's valid
				var parsed interface{}
				if err := json.Unmarshal(data, &parsed); err != nil {
					return fmt.Errorf("failed to parse JSON results: %w", err)
				}

				// Pretty print if no output file specified
				if outputFile == "" {
					formatted, err := json.MarshalIndent(parsed, "", "  ")
					if err != nil {
						return fmt.Errorf("failed to format JSON: %w", err)
					}
					fmt.Println(string(formatted))
				} else {
					if err := os.WriteFile(outputFile, data, 0644); err != nil {
						return fmt.Errorf("failed to write output file: %w", err)
					}
				}

			case "text":
				// TODO: Implement text format conversion
				return fmt.Errorf("text format not yet implemented")

			default:
				return fmt.Errorf("unsupported output format: %s", outputFormat)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "format", "f", "json", "Output format (json/text)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path")
	return cmd
}
