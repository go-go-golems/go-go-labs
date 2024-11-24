# Textractor

Textractor is an application to submit PDFs to AWS Textract and get back structured data.

## Architecture Overview

The application uses a serverless architecture with the following AWS resources:

### Storage Components
- **S3 Bucket**: Stores input PDFs and processed results
- **Input SQS Queue**: Buffers new document processing requests
- **Output SQS Queue**: Receives Textract completion notifications
- **SNS Topic**: Handles Textract completion notifications
- **Notifications SQS Queue**: Provides real-time status updates and progress information

### Processing Components
- **Document Processor Lambda**: Handles new document submissions
- **Completion Processor Lambda**: Processes Textract completion notifications
- **Amazon Textract**: Performs document analysis
- **DynamoDB**: Tracks job status and metadata

## Processing Workflow

1. Go program uploads PDF to S3
2. S3 event triggers Document Processor Lambda via input queue
3. Document Processor submits document to Textract
4. Textract processes document and sends completion notification to SNS
5. SNS triggers Completion Processor Lambda
6. Completion Processor retrieves results and updates job status
7. Status updates sent to notifications queue
8. Go program polls notifications queue for progress

## Job Management

### Job Structure
```go
type TextractJob struct {
  JobID string       // Unique identifier
  DocumentKey string // S3 key of the original PDF
  Status string      // Current processing status
  SubmittedAt time.Time
  CompletedAt time.Time
  TextractID string  // AWS Textract Job ID
  ResultKey string   // S3 key where results are stored
  Error string      // Error message if failed
}
```

### Job States
1. UPLOADING: Initial state when job is created
2. SUBMITTED: File uploaded to S3, ready for processing
3. PROCESSING: Document being analyzed by Textract
4. COMPLETED: Processing finished successfully
5. FAILED: Textract processing failed
6. ERROR: System error occurred

### Job Storage
Jobs are tracked in DynamoDB with the following structure:
- Table: TextractorJobs
- Partition Key: JobID
- GSI1: Status-SubmittedAt (for efficient listing by status)
- GSI2: DocumentKey (for looking up jobs by original document)

Benefits:
- Efficient queries by status
- Easy updates as job status changes
- Serverless and scalable operation
- Built-in TTL for old job cleanup

### Job Management Commands
The CLI provides several commands:
- `list`: Query and display jobs with filtering options
- `status`: Check individual job status
- `purge`: Clean up old jobs and associated data
- `monitor`: Watch job processing in real-time

## Progress Notifications

### Notification Types
1. Job Status Updates:
   - UPLOAD_STARTED: Initial job creation
   - UPLOAD_COMPLETED: File successfully uploaded
   - PROCESSING_STARTED: Textract analysis started
   - PROCESSING_COMPLETED: Analysis finished successfully
   - PROCESSING_FAILED: Analysis failed

2. Progress Updates:
   - TEXTRACT_PROGRESS: Periodic progress updates
   - PAGE_PROCESSED: Individual page completion
   - OPERATION_STARTED: Start of specific operation
   - OPERATION_COMPLETED: Completion of specific operation

3. Error Notifications:
   - SYSTEM_ERROR: Infrastructure/system errors
   - VALIDATION_ERROR: Document validation issues
   - PROCESSING_ERROR: Textract processing errors

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

## Architecture Diagrams

### Infrastructure Flow
```mermaid
graph TD
    CLI[CLI Tool] -->|Upload PDF| S3[S3 Bucket]
    S3 -->|Trigger| DocLambda[Document Processor]
    DocLambda -->|Submit| Textract[Amazon Textract]
    Textract -->|Completion| SNS[SNS Topic]
    SNS -->|Trigger| CompLambda[Completion Processor]
    CompLambda -->|Store Results| S3
    CompLambda -->|Update Status| DDB[(DynamoDB)]
    CompLambda -->|Send Update| NotifQ[Notifications Queue]
    CLI -->|Poll| NotifQ
```

### Job Management Flow
```mermaid
graph TD
    Submit[Submit Command] -->|Create Job| DDB[(DynamoDB)]
    Submit -->|Upload| S3[S3 Bucket]
    
    DocLambda[Document Processor] -->|Update Status| DDB
    DocLambda -->|Process| Textract[Amazon Textract]
    
    CompLambda[Completion Processor] -->|Update Status| DDB
    CompLambda -->|Store Results| S3
    CompLambda -->|Send Update| NotifQ[Notifications Queue]
    
    List[List Command] -->|Query| DDB
    Status[Status Command] -->|Get Job| DDB
    Monitor[Monitor Command] -->|Watch| NotifQ
    
    style Submit fill:#f9f,stroke:#333
    style List fill:#f9f,stroke:#333
    style Status fill:#f9f,stroke:#333
    style Monitor fill:#f9f,stroke:#333
```
