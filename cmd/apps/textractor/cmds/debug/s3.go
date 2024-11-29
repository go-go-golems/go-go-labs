package debug

import (
	"fmt"
	"log"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"
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

	// Add ls subcommand with new flags
	var (
		recursive bool
		fromDate  string
		toDate    string
	)

	lsCmd := &cobra.Command{
		Use:   "ls [prefix]",
		Short: "List files in document S3 bucket",
		Long:  "List files in document S3 bucket. Optionally specify a prefix to filter results",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			resources, err := LoadResources(cmd)
			if err != nil {
				log.Fatalf("Failed to load resources: %v", err)
			}

			// Parse date filters if provided
			var fromTime, toTime *time.Time
			if fromDate != "" {
				t, err := time.Parse("2006-01-02", fromDate)
				if err != nil {
					log.Fatalf("Invalid --from date format. Use YYYY-MM-DD: %v", err)
				}
				fromTime = &t
			}
			if toDate != "" {
				t, err := time.Parse("2006-01-02", toDate)
				if err != nil {
					log.Fatalf("Invalid --to date format. Use YYYY-MM-DD: %v", err)
				}
				toTime = &t
			}

			prefix := ""
			if len(args) > 0 {
				prefix = args[0]
			}

			s3Client, err := pkg.NewS3Client(cmd.Context())
			if err != nil {
				log.Fatalf("Failed to create S3 client: %v", err)
			}

			fmt.Printf("📦 Listing objects in bucket: %s\n", resources.DocumentS3Bucket)
			if prefix != "" {
				fmt.Printf("Using prefix: %s\n", prefix)
			}

			objects, err := s3Client.ListObjects(cmd.Context(), resources.DocumentS3Bucket, pkg.ListObjectsOptions{
				Recursive: recursive,
				FromDate:  fromTime,
				ToDate:    toTime,
				Prefix:    prefix,
			})
			if err != nil {
				log.Fatalf("Failed to list objects: %v", err)
			}

			// Display sorted objects
			for _, obj := range objects {
				fmt.Printf("%s\t%d\t%s\n", obj.LastModified.Format("2006-01-02 15:04:05"),
					obj.Size, *obj.Key)
			}
		},
	}

	lsCmd.Flags().BoolVarP(&recursive, "recursive", "r", true, "List files recursively")
	lsCmd.Flags().StringVar(&fromDate, "from", "", "List files modified after this date (YYYY-MM-DD)")
	lsCmd.Flags().StringVar(&toDate, "to", "", "List files modified before this date (YYYY-MM-DD)")

	s3DebugCmd.AddCommand(lsCmd)
	return s3DebugCmd
}
