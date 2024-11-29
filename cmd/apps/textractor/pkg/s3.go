package pkg

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Client struct {
	client *s3.Client
}

func NewS3Client(ctx context.Context) (*S3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &S3Client{
		client: s3.NewFromConfig(cfg),
	}, nil
}

type ListObjectsOptions struct {
	Recursive bool
	FromDate  *time.Time
	ToDate    *time.Time
	Prefix    string
}

func (c *S3Client) ListObjects(ctx context.Context, bucket string, opts ListObjectsOptions) ([]types.Object, error) {
	var allObjects []types.Object
	input := &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &opts.Prefix,
	}

	paginator := s3.NewListObjectsV2Paginator(c.client, input)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range output.Contents {
			// Apply date filters
			if opts.FromDate != nil && obj.LastModified.Before(*opts.FromDate) {
				continue
			}
			if opts.ToDate != nil && obj.LastModified.After(*opts.ToDate) {
				continue
			}

			// Skip non-recursive listing for objects in subdirectories
			if !opts.Recursive && containsSlash(*obj.Key, opts.Prefix) {
				continue
			}

			allObjects = append(allObjects, obj)
		}
	}

	// Sort objects by date (newest first)
	sort.Slice(allObjects, func(i, j int) bool {
		return allObjects[i].LastModified.After(*allObjects[j].LastModified)
	})

	return allObjects, nil
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
