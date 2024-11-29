package cmds

import (
	"fmt"
	"github.com/go-go-golems/go-go-labs/cmd/apps/textractor/pkg"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	StateUploading = "UPLOADING"
	StateSubmitted = "SUBMITTED"
	StateError     = "ERROR"

	supportedExtensions = ".pdf,.png,.jpg,.jpeg"
)

func NewSubmitCommand() *cobra.Command {
	var recursive bool

	cmd := &cobra.Command{
		Use:   "submit [file/directory]",
		Short: "Submit a PDF file or directory for processing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

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
			s3Client := s3.New(sess)
			dbClient := dynamodb.New(sess)

			// Handle directory vs file
			fileInfo, err := os.Stat(path)
			if err != nil {
				return fmt.Errorf("failed to stat path %s: %w", path, err)
			}

			if fileInfo.IsDir() {
				if !recursive {
					return fmt.Errorf("path is a directory, use --recursive to process directories")
				}
				return submitDirectory(path, resources, s3Client, dbClient)
			}

			return submitFile(path, resources, s3Client, dbClient)
		},
	}

	cmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "Recursively process directories")
	return cmd
}

func submitDirectory(dirPath string, resources *pkg.TextractorResources, s3Client *s3.S3, dbClient *dynamodb.DynamoDB) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && isValidFileType(strings.ToLower(filepath.Ext(path))) {
			if err := submitFile(path, resources, s3Client, dbClient); err != nil {
				log.Printf("Error processing %s: %v", path, err)
				return nil // Continue with next file
			}
		}
		return nil
	})
}

func submitFile(filePath string, resources *pkg.TextractorResources, s3Client *s3.S3, dbClient *dynamodb.DynamoDB) error {
	// Validate file extension
	ext := strings.ToLower(filepath.Ext(filePath))
	if !isValidFileType(ext) {
		return fmt.Errorf("unsupported file type %s. Supported types: %s", ext, supportedExtensions)
	}

	// Generate unique job ID and S3 key
	jobID := uuid.New().String()
	s3Key := fmt.Sprintf("input/%s/%s", jobID, filepath.Base(filePath))

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(resources.Region),
	}))

	jobClient := pkg.NewJobClient(sess, resources.JobsTable)

	// Create initial job record
	err := jobClient.CreateJob(pkg.TextractJob{
		JobID:       jobID,
		DocumentKey: s3Key,
		Status:      StateUploading,
		SubmittedAt: time.Now(),
		TextractID:  "PENDING",
		ResultKey:   "PENDING",
	})
	if err != nil {
		return fmt.Errorf("failed to create initial job record: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		if err := jobClient.UpdateJobStatus(jobID, StateError, fmt.Sprintf("Failed to open file: %v", err)); err != nil {
			log.Printf("Failed to update job status: %v", err)
		}
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Upload to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(resources.DocumentS3Bucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		if err := jobClient.UpdateJobStatus(jobID, StateError, fmt.Sprintf("Failed to upload to S3: %v", err)); err != nil {
			log.Printf("Failed to update job status: %v", err)
		}
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Update job status
	if err := jobClient.UpdateJobStatus(jobID, StateSubmitted, ""); err != nil {
		log.Printf("Warning: Failed to update job status to SUBMITTED: %v", err)
	}

	fmt.Printf("Successfully submitted job %s for %s\n", jobID, filepath.Base(filePath))
	return nil
}

func isValidFileType(ext string) bool {
	validExtensions := strings.Split(supportedExtensions, ",")
	for _, valid := range validExtensions {
		if ext == valid {
			return true
		}
	}
	return false
}
