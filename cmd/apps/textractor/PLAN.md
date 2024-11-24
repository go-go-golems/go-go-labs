textrator is an application to submit PDFs to textract and get back structured data.

## Resources

The way it works is that there is a terraform module that configures the following resources:

- S3 bucket
- Two SQS queues:
  - Input queue: Handles new document processing requests
  - Output queue: Receives Textract completion notifications
- SNS topic: Receives Textract completion notifications
- Lambda function
- EventBridge rule to trigger the lambda function on new files in the S3 bucket

See terraform/README.md for more details.

### Storage
- **S3 Bucket**: Stores input PDFs and potentially processed results
- **Input SQS Queue**: Buffers new document processing requests
- **Output SQS Queue**: Buffers Textract completion notifications
- **SNS Topic**: Handles Textract completion notifications


### Processing
- **Lambda Function**: Orchestrates document processing
- **Amazon Textract**: Performs document analysis
- **EventBridge Rule**: Triggers Lambda on new S3 uploads


The lambda function is configured to use the `textract:AnalyzeDocument` API to extract the text from the PDF.

## Workflow

The workflow is:
1. Go program uploads PDF to S3
2. S3 event triggers Lambda via input queue
3. Lambda submits document to Textract
4. Textract processes document and sends completion notification to SNS
5. SNS forwards notification to output queue
6. Go program polls output queue for results
7. When notification received, program calls textract:GetDocumentAnalysis API

## Jobs and Job Orchestration

### Job Structure
Each document processing request is tracked as a job with the following metadata:

```go
type TextractJob struct {
  JobID string // Unique identifier
  DocumentKey string // S3 key of the original PDF
  Status string // UPLOADING, SUBMITTED, PROCESSING, COMPLETED, FAILED, ERROR
  SubmittedAt time.Time
  CompletedAt time.Time
  TextractID string // AWS Textract Job ID
  ResultKey string // S3 key where results are stored
  Error string // Error message if failed
}
```

### Job States

1. UPLOADING: Initial state when job is created and file upload begins
2. SUBMITTED: File successfully uploaded to S3, ready for processing
3. PROCESSING: Document is being processed by Textract
4. COMPLETED: Processing finished successfully
5. FAILED: Textract processing failed
6. ERROR: System error occurred (upload failed, etc.)

### Job Storage
Jobs are tracked in DynamoDB with the following structure:
- Table: TextractorJobs
- Partition Key: JobID
- GSI1: Status-SubmittedAt (for efficient listing by status)
- GSI2: DocumentKey (for looking up jobs by original document)

This design allows:
- Efficient queries by status
- Easy updates as job status changes
- Serverless and scalable operation
- Built-in TTL for old job cleanup

### Job Lifecycle

1. Document Submission:
   - Generate unique JobID
   - Create job record in DynamoDB (status: UPLOADING)
   - Upload PDF to S3
   - On successful upload, update status to SUBMITTED
   - On upload failure, update status to ERROR with error message

2. Processing:
   - Lambda receives message from input queue
   - Updates job status to PROCESSING
   - Submits document to Textract
   - Textract processes document
   - SNS notification sent upon completion

3. Completion:
   - Lambda receives completion notification via SNS/SQS
   - Updates job status to COMPLETED
   - Stores Textract results in S3
   - Updates job record with result location

4. Error Handling:
   - If processing fails, status updated to FAILED
   - If system error occurs, status updated to ERROR
   - Error message stored in job record
   - Allows for retry mechanisms and debugging

### Job Management Commands
The CLI provides several commands for job management:
- `list`: Query and display jobs with filtering options
- `status`: Check individual job status
- `purge`: Clean up old jobs and associated data
- `monitor`: Watch job processing in real-time

This job orchestration system ensures reliable tracking of document processing requests and provides visibility into the processing pipeline.

## Architecture Diagrams

### Infrastructure Flow
```mermaid
graph TD
    CLI[CLI Tool] -->|Upload PDF| S3[S3 Bucket]
    S3 -->|Trigger| Lambda[Lambda Function]
    Lambda -->|Submit| Textract[Amazon Textract]
    Textract -->|Completion| SNS[SNS Topic]
    SNS -->|Notify| OutputSQS[Output SQS Queue]
    CLI -->|Poll| OutputSQS
    CLI -->|Get Results| Textract
```

### Job Management Flow
```mermaid
graph TD
    Submit[Submit Command] -->|Create Job| DDB[(DynamoDB)]
    Submit -->|Upload| S3[S3 Bucket]
    Submit -->|Queue| InputSQS[Input Queue]
    
    Lambda[Lambda Function] -->|Update Status| DDB
    Lambda -->|Process| Textract[Amazon Textract]
    
    Textract -->|Complete| SNS[SNS Topic]
    SNS -->|Notify| OutputSQS[Output Queue]
    
    List[List Command] -->|Query| DDB
    Status[Status Command] -->|Get Job| DDB
    Monitor[Monitor Command] -->|Watch| DDB
    
    style Submit fill:#f9f,stroke:#333
    style List fill:#f9f,stroke:#333
    style Status fill:#f9f,stroke:#333
    style Monitor fill:#f9f,stroke:#333
```

The diagrams show:
1. The infrastructure flow demonstrates how documents move through the AWS services
2. The job management flow shows how the CLI commands interact with DynamoDB for job tracking while processing occurs

## Detailed Processing Flow

```mermaid
sequenceDiagram
    participant User as User/CLI
    participant DDB as DynamoDB
    participant S3 as S3 Bucket
    participant SQSin as Input SQS
    participant Lambda as Lambda
    participant Textract as Textract
    participant SNS as SNS Topic
    participant SQSout as Output SQS
    participant SQSnotif as Notifications SQS

    User->>DDB: Create job record (UPLOADING)
    Lambda->>SQSnotif: Send upload started notification
    User->>S3: Upload PDF file
    
    alt Upload Failed
        User->>DDB: Update status (ERROR)
        Lambda->>SQSnotif: Send upload failed notification
    else Upload Success
        User->>DDB: Update status (SUBMITTED)
        Lambda->>SQSnotif: Send upload success notification
        S3-->>SQSin: Send ObjectCreated event
        SQSin-->>Lambda: Trigger Lambda function
        
        Lambda->>DDB: Update status (PROCESSING)
        Lambda->>SQSnotif: Send processing started notification
        Lambda->>Textract: Start document analysis
        Lambda->>DDB: Store Textract job ID
        Lambda->>SQSnotif: Send Textract job started notification
        
        Note over Textract: Process document
        
        Textract-->>SNS: Send completion notification
        SNS-->>SQSout: Forward notification
        SNS-->>Lambda: Trigger completion handler
        
        Lambda->>Textract: Get analysis results
        Lambda->>S3: Store results
        
        alt Processing Success
            Lambda->>DDB: Update status (COMPLETED)
            Lambda->>SQSnotif: Send completion notification
        else Processing Failed
            Lambda->>DDB: Update status (FAILED)
            Lambda->>SQSnotif: Send failure notification
        end
    end

    User->>SQSnotif: Long poll for notifications
```

### Notification Types

1. Job Status Updates:
   - `UPLOAD_STARTED`: Initial job creation
   - `UPLOAD_FAILED`: Upload to S3 failed
   - `UPLOAD_COMPLETED`: File successfully uploaded
   - `PROCESSING_STARTED`: Textract analysis started
   - `PROCESSING_COMPLETED`: Analysis finished successfully
   - `PROCESSING_FAILED`: Analysis failed

2. Progress Updates:
   - `TEXTRACT_PROGRESS`: Periodic progress updates
   - `PAGE_PROCESSED`: Individual page completion
   - `OPERATION_STARTED`: Start of specific operation (tables, forms, etc.)
   - `OPERATION_COMPLETED`: Completion of specific operation

3. Error Notifications:
   - `SYSTEM_ERROR`: Infrastructure/system errors
   - `VALIDATION_ERROR`: Document validation issues
   - `PROCESSING_ERROR`: Textract processing errors

### Notification Message Format

```json
{
  "type": "STATUS_UPDATE",
  "jobId": "job-123",
  "status": "PROCESSING",
  "timestamp": "2024-03-21T15:04:05Z",
  "message": "Started Textract analysis",
  "progress": 0.45,
  "details": {
    "pagesProcessed": 5,
    "totalPages": 10,
    "currentOperation": "table_detection",
    "error": "Error message if applicable"
  }
}
```

## Progress Notifications

### Recommended Approach: New Notifications Queue

We recommend implementing a dedicated notifications queue for progress updates:

```mermaid
sequenceDiagram
    participant User as User/CLI
    participant NotifQ as Notifications SQS
    participant Lambda as Lambda
    participant DDB as DynamoDB
    participant Textract as Textract

    Lambda->>DDB: Update job status
    Lambda->>NotifQ: Send status notification
    User->>NotifQ: Long poll for updates
    
    Note over Lambda,NotifQ: Status changes:<br>UPLOADING → SUBMITTED<br>SUBMITTED → PROCESSING<br>PROCESSING → COMPLETED/FAILED
    
    Textract-->>Lambda: Progress updates
    Lambda->>NotifQ: Forward progress %
    User->>NotifQ: Receive progress
```

### Notification Message Structure
```json
{
  "type": "STATUS_UPDATE",
  "jobId": "job-123",
  "status": "PROCESSING",
  "timestamp": "2024-03-21T15:04:05Z",
  "message": "Started Textract analysis",
  "progress": 0.45,  // Optional
  "details": {       // Optional
    "pagesProcessed": 5,
    "totalPages": 10,
    "currentOperation": "table_detection"
  }
}
```

### Implementation Requirements

1. New SQS Queue:
```hcl
resource "aws_sqs_queue" "notifications" {
  name = "${var.prefix}-notifications"
  visibility_timeout_seconds = 30
  message_retention_seconds = 3600  // 1 hour retention
  receive_wait_time_seconds = 20    // Enable long polling
}
```

2. Lambda Updates:
- Send notifications for all status changes
- Forward Textract progress updates
- Include detailed error information
- Add operation-specific details

3. Client Features:
- Long-polling for notifications
- Progress bar support
- Timeout and retry handling
- Filtering by job ID

### Benefits of This Approach

1. **Clean Architecture**
   - Separates progress monitoring from core processing
   - Allows for future expansion (webhooks, email)
   - Won't impact core processing reliability

2. **Flexibility**
   - Can adjust retention and visibility independently
   - Easy to add new notification types
   - Can implement fan-out patterns

3. **Performance**
   - Long-polling reduces API calls
   - No impact on core queues
   - Can scale independently

4. **Cost Control**
   - Shorter retention period than core queues
   - Can be disabled without affecting core functionality
   - Pay only for actual notification volume

### Client Implementation Example
```go
type NotificationPoller struct {
    sqs         *sqs.SQS
    queueURL    string
    jobID       string
    maxWaitTime int64
}

func (p *NotificationPoller) Poll(ctx context.Context) (<-chan Notification, error) {
    ch := make(chan Notification)
    
    go func() {
        defer close(ch)
        
        for {
            select {
            case <-ctx.Done():
                return
            default:
                msgs, err := p.sqs.ReceiveMessageWithContext(ctx, &sqs.ReceiveMessageInput{
                    QueueUrl:            aws.String(p.queueURL),
                    WaitTimeSeconds:     aws.Int64(p.maxWaitTime),
                    MaxNumberOfMessages: aws.Int64(10),
                    MessageAttributeNames: []*string{
                        aws.String("jobId"),
                    },
                })
                
                if err != nil {
                    log.Printf("Error polling notifications: %v", err)
                    time.Sleep(time.Second)
                    continue
                }
                
                for _, msg := range msgs.Messages {
                    // Filter for specific jobId if set
                    if p.jobID != "" && msg.MessageAttributes["jobId"].StringValue != p.jobID {
                        continue
                    }
                    
                    var notif Notification
                    if err := json.Unmarshal([]byte(*msg.Body), &notif); err != nil {
                        log.Printf("Error parsing notification: %v", err)
                        continue
                    }
                    
                    ch <- notif
                    
                    // Delete processed message
                    _, err := p.sqs.DeleteMessage(&sqs.DeleteMessageInput{
                        QueueUrl:      aws.String(p.queueURL),
                        ReceiptHandle: msg.ReceiptHandle,
                    })
                    if err != nil {
                        log.Printf("Error deleting message: %v", err)
                    }
                }
            }
        }
    }()
    
    return ch, nil
}
```

# Lambda Functions

The application uses two Lambda functions for better separation of concerns:

## Document Processor Lambda
Handles new document submissions via SQS from S3 events:
1. Receives S3 event via SQS
2. Updates job status to PROCESSING
3. Submits document to Textract
4. Updates job with Textract ID
5. Sends notifications about processing start

Permissions:
- SQS: Read from input queue
- S3: Read uploaded documents
- Textract: Start analysis
- DynamoDB: Update job status
- SNS: Send notifications

## Completion Processor Lambda
Handles Textract completion notifications via SNS:
1. Receives completion notification from SNS
2. Retrieves full analysis results from Textract
3. Stores results in S3
4. Updates job status in DynamoDB
5. Sends completion notification

Permissions:
- SNS: Receive notifications
- Textract: Get analysis results
- S3: Write results
- DynamoDB: Update job status
- SQS: Send notifications

# Event Flow

```mermaid
sequenceDiagram
    participant CLI as CLI Tool
    participant S3 as S3 Bucket
    participant SQSin as Input SQS
    participant DocLambda as Document Processor
    participant Textract as Textract
    participant SNS as SNS Topic
    participant CompLambda as Completion Processor
    participant SQSnotif as Notifications SQS
    participant DDB as DynamoDB

    CLI->>DDB: Create job record (UPLOADING)
    CLI->>S3: Upload PDF file
    S3-->>SQSin: Send ObjectCreated event
    SQSin-->>DocLambda: Trigger Lambda
    
    DocLambda->>DDB: Update status (PROCESSING)
    DocLambda->>SQSnotif: Send processing notification
    DocLambda->>Textract: Start document analysis
    DocLambda->>DDB: Store Textract job ID
    
    Note over Textract: Process document
    
    Textract-->>SNS: Send completion notification
    SNS-->>CompLambda: Trigger Lambda
    
    CompLambda->>Textract: Get analysis results
    CompLambda->>S3: Store results
    CompLambda->>DDB: Update status (COMPLETED)
    CompLambda->>SQSnotif: Send completion notification
    
    CLI->>SQSnotif: Long poll for notifications
```