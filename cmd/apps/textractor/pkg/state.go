package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/spf13/cobra"
)

// TextractorResources represents all AWS resources needed for the Textractor application
type TextractorResources struct {
	DocumentS3Bucket            string `json:"document_bucket"`
	DocumentS3BucketARN         string `json:"document_bucket_arn"`
	OutputS3Bucket              string `json:"output_bucket"`
	OutputS3BucketARN           string `json:"output_bucket_arn"`
	InputQueue                  string `json:"input_queue_url"`
	CompletionQueue             string `json:"completion_queue_url"`
	NotificationsQueue          string `json:"notifications_queue_url"`
	SNSTopic                    string `json:"sns_topic_arn"`
	DocumentProcessorARN        string `json:"document_processor_arn"`
	DocumentProcessorName       string `json:"document_processor_name"`
	DocumentProcessorLogGroup   string `json:"document_processor_log_group"`
	CompletionProcessorARN      string `json:"completion_processor_arn"`
	CompletionProcessorName     string `json:"completion_processor_name"`
	CompletionProcessorLogGroup string `json:"completion_processor_log_group"`
	Region                      string `json:"region"`
	JobsTable                   string `json:"jobs_table_name"`
	CloudTrailLogGroup          string `json:"cloudtrail_log_group"`
	InputDLQURL                 string `json:"input_dlq_url"`
	CompletionDLQURL            string `json:"completion_dlq_url"`
	NotificationTopic           string `json:"notification_topic_arn"`
}

// StateLoader handles loading Textractor state from various sources
type StateLoader struct {
}

// NewStateLoader creates a new StateLoader
func NewStateLoader() *StateLoader {
	return &StateLoader{}
}

// LoadStateFromCommand loads the Textractor state using flags from the given command
func (s *StateLoader) LoadStateFromCommand(cmd *cobra.Command) (*TextractorResources, error) {
	tfDir, err := cmd.Flags().GetString("tf-dir")
	if err != nil {
		return nil, fmt.Errorf("failed to get tf-dir flag: %w", err)
	}

	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, fmt.Errorf("failed to get config flag: %w", err)
	}

	return s.LoadState(tfDir, configFile)
}

// LoadState loads the Textractor state from either config file or Terraform state
func (s *StateLoader) LoadState(tfDir, configFile string) (*TextractorResources, error) {
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
		if err := s.validateResources(&resources); err != nil {
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

	// Helper function to check and set output
	setOutput := func(field *string, key string) {
		if value, ok := outputMap[key]; ok {
			*field = value
		} else {
			missingOutputs = append(missingOutputs, key)
		}
	}

	// Map all required outputs
	setOutput(&resources.DocumentS3Bucket, "document_bucket")
	setOutput(&resources.InputQueue, "input_queue_url")
	setOutput(&resources.CompletionQueue, "completion_queue_url")
	setOutput(&resources.NotificationsQueue, "notifications_queue_url")
	setOutput(&resources.SNSTopic, "sns_topic_arn")
	setOutput(&resources.Region, "region")
	setOutput(&resources.JobsTable, "jobs_table_name")
	setOutput(&resources.DocumentProcessorARN, "document_processor_arn")
	setOutput(&resources.CompletionProcessorARN, "completion_processor_arn")
	setOutput(&resources.DocumentProcessorName, "document_processor_name")
	setOutput(&resources.CompletionProcessorName, "completion_processor_name")
	setOutput(&resources.DocumentProcessorLogGroup, "document_processor_log_group")
	setOutput(&resources.CompletionProcessorLogGroup, "completion_processor_log_group")
	setOutput(&resources.CloudTrailLogGroup, "cloudtrail_log_group")
	setOutput(&resources.InputDLQURL, "input_dlq_url")
	setOutput(&resources.CompletionDLQURL, "completion_dlq_url")
	setOutput(&resources.NotificationTopic, "notification_topic_arn")
	setOutput(&resources.OutputS3Bucket, "output_bucket")

	if len(missingOutputs) > 0 {
		return nil, fmt.Errorf("missing required terraform outputs: %s", strings.Join(missingOutputs, ", "))
	}

	return resources, nil
}

// validateResources checks if all required fields are present
func (s *StateLoader) validateResources(r *TextractorResources) error {
	required := []struct {
		field string
		value string
	}{
		{"document_bucket", r.DocumentS3Bucket},
		{"input_queue_url", r.InputQueue},
		{"completion_queue_url", r.CompletionQueue},
		{"notifications_queue_url", r.NotificationsQueue},
		{"sns_topic_arn", r.SNSTopic},
		{"region", r.Region},
		{"jobs_table_name", r.JobsTable},
		{"document_processor_arn", r.DocumentProcessorARN},
		{"completion_processor_arn", r.CompletionProcessorARN},
		{"document_processor_name", r.DocumentProcessorName},
		{"completion_processor_name", r.CompletionProcessorName},
		{"document_processor_log_group", r.DocumentProcessorLogGroup},
		{"completion_processor_log_group", r.CompletionProcessorLogGroup},
		{"cloudtrail_log_group", r.CloudTrailLogGroup},
		{"input_dlq_url", r.InputDLQURL},
		{"completion_dlq_url", r.CompletionDLQURL},
		{"notification_topic_arn", r.NotificationTopic},
		{"output_bucket", r.OutputS3Bucket},
	}

	var missing []string
	for _, req := range required {
		if req.value == "" {
			missing = append(missing, req.field)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}
