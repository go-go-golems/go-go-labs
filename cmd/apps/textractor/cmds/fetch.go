package cmds

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/textract"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func NewFetchCommand() *cobra.Command {
	var (
		outputFormat     string
		outputFile       string
		documentAnalysis bool
	)

	cmd := &cobra.Command{
		Use:   "fetch [jobID]",
		Short: "Fetch the results of a processed job",
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

			var job pkg.TextractJob
			err = dynamodbattribute.UnmarshalMap(result.Item, &job)
			if err != nil {
				return fmt.Errorf("failed to unmarshal job data: %w", err)
			}

			if job.Status != "COMPLETED" {
				return fmt.Errorf("job is not completed (current status: %s)", job.Status)
			}

			if documentAnalysis {
				// Create Textract client
				textractClient := textract.New(sess)

				var allBlocks []*textract.Block
				var nextToken *string
				var documentMetadata *textract.DocumentMetadata

				for {
					input := &textract.GetDocumentAnalysisInput{
						JobId: aws.String(job.TextractID),
					}
					if nextToken != nil {
						input.NextToken = nextToken
					}

					result, err := textractClient.GetDocumentAnalysis(input)
					if err != nil {
						return fmt.Errorf("failed to get document analysis: %w", err)
					}

					// Capture first DocumentMetadata we see
					if documentMetadata == nil && result.DocumentMetadata != nil {
						documentMetadata = result.DocumentMetadata
					}

					allBlocks = append(allBlocks, result.Blocks...)
					nextToken = result.NextToken

					if nextToken == nil {
						break
					}
				}

				// Output results based on format
				switch outputFormat {
				case "json":
					output := map[string]interface{}{
						"DocumentMetadata": documentMetadata,
						"Blocks":           allBlocks,
					}

					if outputFile == "" {
						formatted, err := json.MarshalIndent(output, "", "  ")
						if err != nil {
							return fmt.Errorf("failed to format JSON: %w", err)
						}
						fmt.Println(string(formatted))
					} else {
						combined, err := json.Marshal(output)
						if err != nil {
							return fmt.Errorf("failed to combine JSON results: %w", err)
						}
						if err := os.WriteFile(outputFile, combined, 0644); err != nil {
							return fmt.Errorf("failed to write output file: %w", err)
						}
					}
				case "text":
					return fmt.Errorf("text format not yet implemented")
				default:
					return fmt.Errorf("unsupported output format: %s", outputFormat)
				}

				return nil
			}

			// List all objects under the result prefix
			outputPrefix := job.ResultKey
			log.Info().Msgf("Listing results from S3: %s/%s", resources.OutputS3Bucket, outputPrefix)

			var allData [][]byte
			err = s3Client.ListObjectsV2Pages(&s3.ListObjectsV2Input{
				Bucket: aws.String(resources.OutputS3Bucket),
				Prefix: aws.String(outputPrefix),
			}, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
				for _, obj := range page.Contents {
					// Trim the outputPrefix from the key before checking
					relativeKey := strings.TrimPrefix(*obj.Key, outputPrefix+"/")

					// skip if starting with .
					if strings.HasPrefix(relativeKey, ".") {
						continue
					}
					log.Debug().Msgf("Fetching object: %s", *obj.Key)
					output, err := s3Client.GetObject(&s3.GetObjectInput{
						Bucket: aws.String(resources.OutputS3Bucket),
						Key:    obj.Key,
					})
					if err != nil {
						log.Error().Err(err).Msgf("failed to get object %s", *obj.Key)
						continue
					}
					defer output.Body.Close()

					data, err := io.ReadAll(output.Body)
					if err != nil {
						log.Error().Err(err).Msgf("failed to read object %s", *obj.Key)
						continue
					}
					allData = append(allData, data)
				}
				return true
			})
			if err != nil {
				return fmt.Errorf("failed to list results from S3: %w", err)
			}

			if len(allData) == 0 {
				return fmt.Errorf("no results found for job %s", jobID)
			}

			// Process and output the results based on format
			switch outputFormat {
			case "json":
				// Parse all JSON results
				var allBlocks []interface{}
				var documentMetadata interface{}

				for _, data := range allData {
					var parsed struct {
						DocumentMetadata interface{}   `json:"DocumentMetadata"`
						Blocks           []interface{} `json:"Blocks"`
					}
					if err := json.Unmarshal(data, &parsed); err != nil {
						return fmt.Errorf("failed to parse JSON results: %w", err)
					} else {
						// Capture first DocumentMetadata we see
						if documentMetadata == nil && parsed.DocumentMetadata != nil {
							documentMetadata = parsed.DocumentMetadata
						}
						allBlocks = append(allBlocks, parsed.Blocks...)
					}
				}

				output := map[string]interface{}{
					"DocumentMetadata": documentMetadata,
					"Blocks":           allBlocks,
				}

				// Pretty print if no output file specified
				if outputFile == "" {
					formatted, err := json.MarshalIndent(output, "", "  ")
					if err != nil {
						return fmt.Errorf("failed to format JSON: %w", err)
					}
					fmt.Println(string(formatted))
				} else {
					// Write combined results to file
					combined, err := json.Marshal(output)
					if err != nil {
						return fmt.Errorf("failed to combine JSON results: %w", err)
					}
					if err := os.WriteFile(outputFile, combined, 0644); err != nil {
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
	cmd.Flags().BoolVar(&documentAnalysis, "document-analysis", false, "Fetch results directly from Textract instead of S3")
	return cmd
}
