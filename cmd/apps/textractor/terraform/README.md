# Textractor Terraform Module

This module sets up the AWS infrastructure required for the Textractor application, which processes PDFs using Amazon Textract.

## Architecture

The module creates:
- S3 bucket for document storage
- Input and output SQS queues
- SNS topic for Textract notifications
- Lambda function for processing (automatically packaged and deployed)
- EventBridge rule for triggering the Lambda
- Required IAM roles and policies

## Usage 

```hcl
module "textractor" {
  source = "./modules/textractor"
  prefix = "textractor-dev"
  environment = "dev"
  bucket_name = "your-unique-bucket-name"
}
```

## Prerequisites

1. AWS credentials configured
2. Terraform >= 0.13

## Variables

- `prefix`: Prefix for resource names
- `environment`: Environment name (e.g., dev, prod)
- `bucket_name`: Globally unique S3 bucket name

## Outputs

- `bucket_name`: Name of the created S3 bucket
- `input_queue_url`: URL of the input SQS queue
- `output_queue_url`: URL of the output SQS queue
- `sns_topic_arn`: ARN of the SNS topic
- `lambda_function_name`: Name of the Lambda function

## Lambda Function

The Lambda function code is included in the module and will be automatically packaged and deployed. The function:
1. Receives S3 upload notifications
2. Submits documents to Amazon Textract for analysis
3. Configures SNS notifications for job completion