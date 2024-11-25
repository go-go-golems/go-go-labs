package debug

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

func newOutputS3Command() *cobra.Command {
	outputS3DebugCmd := &cobra.Command{
		Use:   "output-s3",
		Short: "Debug output S3 bucket configuration",
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			fmt.Printf("🔍 Debugging output S3 bucket: %s\n", resources.OutputS3Bucket)

			// List bucket details
			err = runAWSCommand("s3api", "get-bucket-location",
				"--bucket", resources.OutputS3Bucket)
			if err != nil {
				log.Printf("Failed to get bucket location: %v", err)
			}

			// Get bucket versioning
			err = runAWSCommand("s3api", "get-bucket-versioning",
				"--bucket", resources.OutputS3Bucket)
			if err != nil {
				log.Printf("Failed to get bucket versioning: %v", err)
			}
		},
	}

	outputS3DebugCmd.AddCommand(&cobra.Command{
		Use:   "ls [prefix]",
		Short: "List files in output S3 bucket",
		Long:  "List files in output S3 bucket. Optionally specify a prefix to filter results",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			s3Uri := fmt.Sprintf("s3://%s", resources.OutputS3Bucket)
			if len(args) > 0 {
				s3Uri = fmt.Sprintf("%s/%s", s3Uri, args[0])
			}

			err = runAWSCommand("s3", "ls", s3Uri)
			if err != nil {
				log.Printf("Failed to list bucket contents: %v", err)
			}
		},
	})

	return outputS3DebugCmd
}
