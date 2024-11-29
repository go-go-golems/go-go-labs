package pkg

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type JobClient struct {
	db        *dynamodb.DynamoDB
	tableName string
}

func NewJobClient(sess *session.Session, tableName string) *JobClient {
	return &JobClient{
		db:        dynamodb.New(sess),
		tableName: tableName,
	}
}

func (c *JobClient) CreateJob(job TextractJob) error {
	av, err := dynamodbattribute.MarshalMap(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job record: %w", err)
	}

	_, err = c.db.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(c.tableName),
		Item:      av,
	})
	return err
}

func (c *JobClient) UpdateJobStatus(jobID string, status string, errorMsg string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(c.tableName),
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
				S: aws.String(status),
			},
			":error": {
				S: aws.String(errorMsg),
			},
		},
	}

	_, err := c.db.UpdateItem(input)
	return err
}

func (c *JobClient) GetJob(jobID string) (*TextractJob, error) {
	result, err := c.db.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(c.tableName),
		Key: map[string]*dynamodb.AttributeValue{
			"JobID": {
				S: aws.String(jobID),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get job details: %w", err)
	}

	if result.Item == nil {
		return nil, fmt.Errorf("job %s not found", jobID)
	}

	var job TextractJob
	err = dynamodbattribute.UnmarshalMap(result.Item, &job)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal job data: %w", err)
	}

	return &job, nil
}

type ListJobsOptions struct {
	Since  *time.Time
	Status string
}

func (c *JobClient) ListJobs(opts ListJobsOptions) ([]TextractJob, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(c.tableName),
	}

	var filterExpressions []string
	expressionValues := map[string]*dynamodb.AttributeValue{}
	expressionNames := map[string]*string{}

	if opts.Since != nil {
		filterExpressions = append(filterExpressions, "SubmittedAt >= :t")
		expressionValues[":t"] = &dynamodb.AttributeValue{S: aws.String(opts.Since.Format(time.RFC3339))}
	}

	if opts.Status != "" {
		filterExpressions = append(filterExpressions, "#statusAlias = :s")
		expressionValues[":s"] = &dynamodb.AttributeValue{S: aws.String(opts.Status)}
		expressionNames["#statusAlias"] = aws.String("Status")
	}

	if len(filterExpressions) > 0 {
		input.FilterExpression = aws.String(joinWithAND(filterExpressions))
		input.ExpressionAttributeValues = expressionValues
		if len(expressionNames) > 0 {
			input.ExpressionAttributeNames = expressionNames
		}
	}

	result, err := c.db.Scan(input)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}

	var jobs []TextractJob
	if err := dynamodbattribute.UnmarshalListOfMaps(result.Items, &jobs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jobs: %w", err)
	}

	return jobs, nil
}

func joinWithAND(expressions []string) string {
	if len(expressions) == 0 {
		return ""
	}
	if len(expressions) == 1 {
		return expressions[0]
	}
	return fmt.Sprintf("(%s)", strings.Join(expressions, " AND "))
}
