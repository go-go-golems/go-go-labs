package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	StateUploading = "UPLOADING"
	StateSubmitted = "SUBMITTED"
	StateError     = "ERROR"
)

func newSubmitCommand() *cobra.Command {
	var recursive bool

	cmd := &cobra.Command{
		Use:   "submit [file/directory]",
			Short: "Submit a PDF file or directory for processing",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				path := args[0]

				// Load resources
				resources, err := loadTerraformState(tfDir)
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

func submitDirectory(dirPath string, resources *TextractorResources, s3Client *s3.S3, dbClient *dynamodb.DynamoDB) error {
	return filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.ToLower(filepath.Ext(path)) == ".pdf" {
			if err := submitFile(path, resources, s3Client, dbClient); err != nil {
				log.Printf("Error processing %s: %v", path, err)
				return nil // Continue with next file
			}
		}
		return nil
	})
}

func createJob(jobID, documentKey string, dbClient *dynamodb.DynamoDB, tableName string, state string, errorMsg string) error {
	job := TextractJob{
		JobID:       jobID,
		DocumentKey: documentKey,
		Status:     state,
		SubmittedAt: time.Now(),
	}

	if errorMsg != "" {
		job.Error = errorMsg
	}

	av, err := dynamodbattribute.MarshalMap(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job record: %w", err)
	}

	_, err = dbClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      av,
	})
	return err
}

func updateJobStatus(jobID string, state string, errorMsg string, dbClient *dynamodb.DynamoDB, tableName string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"JobID": {
				S: aws.String(jobID),
			},
		},
		UpdateExpression: aws.String("SET #status = :status, #error = :error"),
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("Status"),
			"#error":  aws.String("Error"),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":status": {
				S: aws.String(state),
			},
			":error": {
				S: aws.String(errorMsg),
			},
		},
	}

	_, err := dbClient.UpdateItem(input)
	return err
}

func submitFile(filePath string, resources *TextractorResources, s3Client *s3.S3, dbClient *dynamodb.DynamoDB) error {
	// Validate file is PDF
	if strings.ToLower(filepath.Ext(filePath)) != ".pdf" {
		return fmt.Errorf("file must be a PDF: %s", filePath)
	}

	// Generate unique job ID and S3 key
	jobID := uuid.New().String()
	s3Key := fmt.Sprintf("input/%s/%s", jobID, filepath.Base(filePath))

	// Create initial job record with UPLOADING status
	err := createJob(jobID, s3Key, dbClient, resources.JobsTable, StateUploading, "")
	if err != nil {
		return fmt.Errorf("failed to create initial job record: %w", err)
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		updateErr := updateJobStatus(jobID, StateError, fmt.Sprintf("Failed to open file: %v", err), dbClient, resources.JobsTable)
		if updateErr != nil {
			log.Printf("Failed to update job status: %v", updateErr)
		}
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Upload to S3
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(resources.S3Bucket),
		Key:         aws.String(s3Key),
		Body:        file,
		ContentType: aws.String("application/pdf"),
	})
	if err != nil {
		updateErr := updateJobStatus(jobID, StateError, fmt.Sprintf("Failed to upload to S3: %v", err), dbClient, resources.JobsTable)
		if updateErr != nil {
			log.Printf("Failed to update job status: %v", updateErr)
		}
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	// Update job status to SUBMITTED
	err = updateJobStatus(jobID, StateSubmitted, "", dbClient, resources.JobsTable)
	if err != nil {
		log.Printf("Warning: Failed to update job status to SUBMITTED: %v", err)
	}

	fmt.Printf("Successfully submitted job %s for %s\n", jobID, filepath.Base(filePath))
	return nil
} 