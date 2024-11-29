package debug

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

			// Initialize AWS SDK
			cfg, err := config.LoadDefaultConfig(context.Background())
			if err != nil {
				log.Fatalf("Failed to load AWS config: %v", err)
			}
			client := s3.NewFromConfig(cfg)

			// List objects
			var allObjects []types.Object
			input := &s3.ListObjectsV2Input{
				Bucket: &resources.DocumentS3Bucket,
				Prefix: &prefix,
			}

			fmt.Printf("📦 Listing objects in bucket: %s\n", resources.DocumentS3Bucket)
			if prefix != "" {
				fmt.Printf("Using prefix: %s\n", prefix)
			}

			paginator := s3.NewListObjectsV2Paginator(client, input)
			for paginator.HasMorePages() {
				output, err := paginator.NextPage(context.Background())
				if err != nil {
					log.Fatalf("Failed to list objects: %v", err)
				}

				for _, obj := range output.Contents {
					// Apply date filters
					if fromTime != nil && obj.LastModified.Before(*fromTime) {
						continue
					}
					if toTime != nil && obj.LastModified.After(*toTime) {
						continue
					}

					// Skip non-recursive listing for objects in subdirectories
					if !recursive && containsSlash(*obj.Key, prefix) {
						continue
					}

					allObjects = append(allObjects, obj)
				}
			}

			// Sort objects by date (newest first)
			sort.Slice(allObjects, func(i, j int) bool {
				return allObjects[i].LastModified.After(*allObjects[j].LastModified)
			})

			// Display sorted objects
			for _, obj := range allObjects {
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

// containsSlash returns true if the key contains additional path components after the prefix
func containsSlash(key, prefix string) bool {
	remainder := key[len(prefix):]
	for _, c := range remainder {
		if c == '/' {
			return true
		}
	}
	return false
}
