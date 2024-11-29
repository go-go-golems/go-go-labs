package debug

import (
	"fmt"
	"log"

	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"
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

			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}

			s3Client, err := pkg.NewS3Client(cmd.Context())
			if err != nil {
				log.Fatalf("Failed to create S3 client: %v", err)
			}

			objects, err := s3Client.ListObjects(cmd.Context(), resources.OutputS3Bucket, pkg.ListObjectsOptions{
				Recursive: true,
				Prefix:    prefix,
			})
			if err != nil {
				log.Fatalf("Failed to list objects: %v", err)
			}

			for _, obj := range objects {
				fmt.Printf("%s\t%d\t%s\n", obj.LastModified.Format("2006-01-02 15:04:05"),
					obj.Size, *obj.Key)
			}
		},
	})

	return outputS3DebugCmd
}
