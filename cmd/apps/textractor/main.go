package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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
}

var (
	tfDir string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "textractor",
		Short: "Manage Textractor AWS resources and process PDFs",
	}

	rootCmd.PersistentFlags().StringVar(&tfDir, "tf-dir", "terraform", "Directory containing Terraform state")

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
			
			// Initialize AWS session and DynamoDB client
			sess := session.Must(session.NewSession())
			db := dynamodb.New(sess)
			
			// Build query based on flags
			input := &dynamodb.QueryInput{
				TableName: aws.String("TextractorJobs"),
			}
			
			if status != "" {
				// Query GSI1 for specific status
				input.IndexName = aws.String("Status-SubmittedAt-Index")
				// ... set key conditions
			}
			
			// Execute query and format results
			// ...
			
			return nil
		},
	}
	
	cmd.Flags().String("status", "", "Filter by job status")
	cmd.Flags().String("since", "", "Show jobs since date (YYYY-MM-DD)")
	
	return cmd
} 