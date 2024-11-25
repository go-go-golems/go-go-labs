package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func newS3Command() *cobra.Command {
	s3DebugCmd := &cobra.Command{
		Use:   "s3",
		Short: "Debug S3 bucket configuration",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging S3 buckets\n")
			fmt.Printf("Document Bucket: %s\n", resources.DocumentS3Bucket)
			fmt.Printf("Output Bucket: %s\n", resources.OutputS3Bucket)

			// Check bucket configuration
			err = runAWSCommand("s3api", "get-bucket-location",
				"--bucket", resources.DocumentS3Bucket)
			if err != nil {
				log.Printf("Failed to get document bucket location: %v", err)
			}

			err = runAWSCommand("s3api", "get-bucket-versioning",
				"--bucket", resources.DocumentS3Bucket)
			if err != nil {
				log.Printf("Failed to get document bucket versioning: %v", err)
			}

			err = runAWSCommand("s3api", "get-bucket-notification-configuration",
				"--bucket", resources.DocumentS3Bucket)
			if err != nil {
				log.Printf("Failed to get document bucket notifications: %v", err)
			}
		},
	}

	// Add ls subcommand
	s3DebugCmd.AddCommand(&cobra.Command{
		Use:   "ls [prefix]",
		Short: "List files in document S3 bucket",
		Long:  "List files in document S3 bucket. Optionally specify a prefix to filter results",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			s3Uri := fmt.Sprintf("s3://%s", resources.DocumentS3Bucket)
			if len(args) > 0 {
				s3Uri = fmt.Sprintf("%s/%s", s3Uri, args[0])
			}

			err = runAWSCommand("s3", "ls", s3Uri)
			if err != nil {
				log.Printf("Failed to list bucket contents: %v", err)
			}
		},
	})

	return s3DebugCmd
}
