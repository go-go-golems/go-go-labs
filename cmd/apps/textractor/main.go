package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/cobra"
)

// TextractorResources represents all AWS resources needed for the Textractor application
type TextractorResources struct {
	S3Bucket     string `json:"s3_bucket"`
	InputQueue   string `json:"input_queue_url"`
	OutputQueue  string `json:"output_queue_url"`
	SNSTopic     string `json:"sns_topic_arn"`
	LambdaARN    string `json:"lambda_arn"`
	Region       string `json:"region"`
	FunctionName string `json:"function_name"`
	JobsTable    string `json:"jobs_table_name"`
}

// Add TextractJob struct as defined in PLAN.md
type TextractJob struct {
	JobID        string     `json:"job_id" dynamodbav:"JobID"`
	DocumentKey  string     `json:"document_key" dynamodbav:"DocumentKey"`
	Status       string     `json:"status" dynamodbav:"Status"`
	SubmittedAt  time.Time  `json:"submitted_at" dynamodbav:"SubmittedAt"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" dynamodbav:"CompletedAt,omitempty"`
	TextractID   string     `json:"textract_id" dynamodbav:"TextractID"`
	ResultKey    string     `json:"result_key" dynamodbav:"ResultKey"`
	Error        string     `json:"error,omitempty" dynamodbav:"Error,omitempty"`
}

var (
	tfDir string
	configFile string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "textractor",
		Short: "Manage Textractor AWS resources and process PDFs",
	}

	rootCmd.PersistentFlags().StringVar(&tfDir, "tf-dir", "terraform", "Directory containing Terraform state")
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "JSON config file containing resource configuration")

	// Add save-config subcommand
	saveConfigCmd := &cobra.Command{
		Use:   "save-config",
		Short: "Save resource configuration to JSON file",
		RunE: func(cmd *cobra.Command, args []string) error {
			resources, err := loadTerraformState(tfDir)
			if err != nil {
				return fmt.Errorf("failed to load terraform state: %w", err)
			}

			output := configFile
			if output == "" {
				output = "textractor-config.json"
			}

			data, err := json.MarshalIndent(resources, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			if err := os.WriteFile(output, data, 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}

			fmt.Printf("Configuration saved to %s\n", output)
			return nil
		},
	}
	rootCmd.AddCommand(saveConfigCmd)

	// Add debug-vars subcommand
	debugVarsCmd := &cobra.Command{
		Use:   "debug-vars",
		Short: "Print environment variables for debugging",
		Run:   printDebugVars,
	}
	rootCmd.AddCommand(debugVarsCmd)

	// Add run command
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Run the Textractor application",
		Run:   run,
	}
	rootCmd.AddCommand(runCmd)

	// Add debug commands
	addDebugCommands(rootCmd)

	// Add list command
	listCmd := newListCommand()
	rootCmd.AddCommand(listCmd)

	// Add submit command
	submitCmd := newSubmitCommand()
	rootCmd.AddCommand(submitCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func printDebugVars(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	// Print in a format suitable for shell script
	fmt.Printf("export FUNCTION_NAME=\"%s\"\n", resources.FunctionName)
	fmt.Printf("export BUCKET_NAME=\"%s\"\n", resources.S3Bucket)
	fmt.Printf("export INPUT_QUEUE_URL=\"%s\"\n", resources.InputQueue)
	fmt.Printf("export OUTPUT_QUEUE_URL=\"%s\"\n", resources.OutputQueue)
	fmt.Printf("export TOPIC_ARN=\"%s\"\n", resources.SNSTopic)
	fmt.Printf("export AWS_REGION=\"%s\"\n", resources.Region)

	// Print helper message
	fmt.Println("\n# To use these variables, run:")
	fmt.Println("# eval $(textractor debug-vars)")
}

func run(cmd *cobra.Command, args []string) {
	resources, err := loadTerraformState(tfDir)
	if err != nil {
		log.Fatalf("Failed to load Terraform state: %v", err)
	}

	// Print the loaded resources
	fmt.Printf("Textractor Resources:\n")
	fmt.Printf("  S3 Bucket:      %s\n", resources.S3Bucket)
	fmt.Printf("  Input Queue:    %s\n", resources.InputQueue)
	fmt.Printf("  Output Queue:   %s\n", resources.OutputQueue)
	fmt.Printf("  SNS Topic:      %s\n", resources.SNSTopic)
	fmt.Printf("  Lambda ARN:     %s\n", resources.LambdaARN)
	fmt.Printf("  Region:         %s\n", resources.Region)
	fmt.Printf("  Function Name:  %s\n", resources.FunctionName)
}

func loadTerraformState(tfDir string) (*TextractorResources, error) {
	// First try loading from config file if specified
	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		var resources TextractorResources
		if err := json.Unmarshal(data, &resources); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}

		// Validate required fields
		if err := validateResources(&resources); err != nil {
			return nil, fmt.Errorf("invalid config file: %w", err)
		}

		return &resources, nil
	}

	// Fall back to loading from Terraform state
	absPath, err := filepath.Abs(tfDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	tf, err := tfexec.NewTerraform(absPath, "terraform")
	if err != nil {
		return nil, fmt.Errorf("error running NewTerraform: %w", err)
	}

	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return nil, fmt.Errorf("error running Init: %w", err)
	}

	state, err := tf.Show(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error running Show: %w", err)
	}

	if state.Values == nil || len(state.Values.Outputs) == 0 {
		return nil, fmt.Errorf("no terraform state or outputs found")
	}

	resources := &TextractorResources{}
	outputMap := make(map[string]string)

	// Map all outputs to strings
	for name, output := range state.Values.Outputs {
		if value, ok := output.Value.(string); ok {
			outputMap[name] = value
		}
	}

	// Map outputs to struct fields
	var missingOutputs []string

	if value, ok := outputMap["bucket_name"]; ok {
		resources.S3Bucket = value
	} else {
		missingOutputs = append(missingOutputs, "bucket_name")
	}

	if value, ok := outputMap["input_queue_url"]; ok {
		resources.InputQueue = value
	} else {
		missingOutputs = append(missingOutputs, "input_queue_url")
	}

	if value, ok := outputMap["output_queue_url"]; ok {
		resources.OutputQueue = value
	} else {
		missingOutputs = append(missingOutputs, "output_queue_url")
	}

	if value, ok := outputMap["sns_topic_arn"]; ok {
		resources.SNSTopic = value
	} else {
		missingOutputs = append(missingOutputs, "sns_topic_arn")
	}

	if value, ok := outputMap["lambda_arn"]; ok {
		resources.LambdaARN = value
	} else {
		missingOutputs = append(missingOutputs, "lambda_arn")
	}

	if value, ok := outputMap["function_name"]; ok {
		resources.FunctionName = value
	} else {
		missingOutputs = append(missingOutputs, "function_name")
	}

	if value, ok := outputMap["region"]; ok {
		resources.Region = value
	} else {
		missingOutputs = append(missingOutputs, "region")
	}

	if value, ok := outputMap["jobs_table_name"]; ok {
		resources.JobsTable = value
	} else {
		missingOutputs = append(missingOutputs, "jobs_table_name")
	}

	if len(missingOutputs) > 0 {
		return nil, fmt.Errorf("missing required terraform outputs: %s", strings.Join(missingOutputs, ", "))
	}

	return resources, nil
}

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Textract jobs",
		RunE: func(cmd *cobra.Command, args []string) error {
			status, _ := cmd.Flags().GetString("status")
			since, _ := cmd.Flags().GetString("since")
			
			// Load resources to get table name
			resources, err := loadTerraformState(tfDir)
			if err != nil {
				return fmt.Errorf("failed to load terraform state: %w", err)
			}
			
			// Initialize AWS session and DynamoDB client
			sess := session.Must(session.NewSession())
			db := dynamodb.New(sess)
			
			// Build query based on flags
			input := &dynamodb.QueryInput{
				TableName: aws.String(resources.JobsTable),
			}
			
			if status != "" {
				// Query GSI1 for specific status
				input.IndexName = aws.String("Status-SubmittedAt-Index")
				input.KeyConditionExpression = aws.String("Status = :status")
				input.ExpressionAttributeValues = map[string]*dynamodb.AttributeValue{
					":status": {
						S: aws.String(status),
					},
				}
				
				if since != "" {
					input.KeyConditionExpression = aws.String("Status = :status AND SubmittedAt >= :since")
					input.ExpressionAttributeValues[":since"] = &dynamodb.AttributeValue{
						S: aws.String(since),
					}
				}
			} else {
				// If no status provided, scan the table
				scanInput := &dynamodb.ScanInput{
					TableName: aws.String(resources.JobsTable),
				}
				
				result, err := db.Scan(scanInput)
				if err != nil {
					return fmt.Errorf("failed to scan jobs: %w", err)
				}
				
				printJobs(result.Items)
				return nil
			}
			
			// Execute query
			result, err := db.Query(input)
			if err != nil {
				return fmt.Errorf("failed to query jobs: %w", err)
			}
			
			printJobs(result.Items)
			return nil
		},
	}
	
	cmd.Flags().String("status", "", "Filter by job status (SUBMITTED, PROCESSING, COMPLETED, FAILED)")
	cmd.Flags().String("since", "", "Show jobs since date (YYYY-MM-DD)")
	
	return cmd
}

// Update printJobs to use TextractJob struct
func printJobs(items []map[string]*dynamodb.AttributeValue) {
	if len(items) == 0 {
		fmt.Println("No jobs found")
		return
	}

	jobs := make([]TextractJob, 0, len(items))
	for _, item := range items {
		var job TextractJob
		err := dynamodbattribute.UnmarshalMap(item, &job)
		if err != nil {
			log.Printf("Error unmarshaling job: %v", err)
			continue
		}
		jobs = append(jobs, job)
	}

	if len(jobs) == 0 {
		fmt.Println("No valid jobs found")
		return
	}

	// Sort jobs by submission time, newest first
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].SubmittedAt.After(jobs[j].SubmittedAt)
	})

	fmt.Printf("%-36s %-12s %-20s %-20s %-20s %s\n", 
		"Job ID", "Status", "Submitted", "Completed", "Textract ID", "Document")
	fmt.Println(strings.Repeat("-", 120))

	for _, job := range jobs {
		completedAt := "-"
		if job.CompletedAt != nil {
			completedAt = job.CompletedAt.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%-36s %-12s %-20s %-20s %-20s %s\n",
			job.JobID,
			job.Status,
			job.SubmittedAt.Format("2006-01-02 15:04:05"),
			completedAt,
			job.TextractID,
			job.DocumentKey)

		// If there's an error, print it indented on the next line
		if job.Error != "" {
			fmt.Printf("    Error: %s\n", job.Error)
		}
	}
}

// Add validation function
func validateResources(r *TextractorResources) error {
	var missing []string

	if r.S3Bucket == "" {
		missing = append(missing, "s3_bucket")
	}
	if r.InputQueue == "" {
		missing = append(missing, "input_queue_url")
	}
	if r.OutputQueue == "" {
		missing = append(missing, "output_queue_url")
	}
	if r.SNSTopic == "" {
		missing = append(missing, "sns_topic_arn")
	}
	if r.LambdaARN == "" {
		missing = append(missing, "lambda_arn")
	}
	if r.Region == "" {
		missing = append(missing, "region")
	}
	if r.FunctionName == "" {
		missing = append(missing, "function_name")
	}
	if r.JobsTable == "" {
		missing = append(missing, "jobs_table_name")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
} 