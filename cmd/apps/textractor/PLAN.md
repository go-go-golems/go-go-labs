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
