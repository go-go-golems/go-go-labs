output "document_bucket" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.document_bucket.id
}

output "document_bucket_arn" {
  description = "ARN of the S3 bucket for documents"
  value       = aws_s3_bucket.document_bucket.arn
}

output "output_bucket" {
  description = "Name of the S3 bucket for Textract output"
  value       = aws_s3_bucket.output_bucket.id
}

output "output_bucket_arn" {
  description = "ARN of the S3 bucket for Textract output"
  value       = aws_s3_bucket.output_bucket.arn
}

output "input_queue_url" {
  description = "URL of the input SQS queue"
  value       = aws_sqs_queue.input_queue.url
}

output "completion_queue_url" {
  description = "URL of the completion SQS queue"
  value       = aws_sqs_queue.completion_queue.url
}

output "sns_topic_arn" {
  description = "ARN of the SNS topic"
  value       = aws_sns_topic.textract_completion.arn
}

output "notifications_queue_url" {
  description = "URL of the notifications SQS queue"
  value       = aws_sqs_queue.notifications.url
}

output "document_processor_arn" {
  description = "ARN of the document processor Lambda function"
  value       = aws_lambda_function.document_processor.arn
}

output "completion_processor_arn" {
  description = "ARN of the completion processor Lambda function"
  value       = aws_lambda_function.completion_processor.arn
}

output "document_processor_name" {
  description = "Name of the document processor Lambda function"
  value       = aws_lambda_function.document_processor.function_name
}

output "completion_processor_name" {
  description = "Name of the completion processor Lambda function"
  value       = aws_lambda_function.completion_processor.function_name
}

output "region" {
  description = "AWS region"
  value       = data.aws_region.current.name
}

output "jobs_table_name" {
  description = "Name of the DynamoDB jobs table"
  value       = aws_dynamodb_table.jobs.name
}

output "document_processor_log_group" {
  description = "Name of the document processor CloudWatch log group"
  value       = aws_cloudwatch_log_group.document_processor_logs.name
}

output "completion_processor_log_group" {
  description = "Name of the completion processor CloudWatch log group"
  value       = aws_cloudwatch_log_group.completion_processor_logs.name
}

output "cloudtrail_log_group" {
  description = "Name of the CloudTrail log group"
  value       = aws_cloudwatch_log_group.cloudtrail_logs.name
}

# DLQ outputs
output "input_dlq_url" {
  description = "The URL of the input Dead Letter Queue"
  value       = aws_sqs_queue.input_dlq.url
}

output "input_dlq_arn" {
  description = "The ARN of the input Dead Letter Queue"
  value       = aws_sqs_queue.input_dlq.arn
}

output "textract_role_arn" {
  description = "ARN of IAM role for Textract notifications"
  value       = aws_iam_role.textract_role.arn
}

output "completion_dlq_url" {
  description = "The URL of the completion Dead Letter Queue"
  value       = aws_sqs_queue.completion_dlq.url
} 

output "completion_dlq_arn" {
  description = "The ARN of the completion Dead Letter Queue"
  value       = aws_sqs_queue.completion_dlq.arn
}

output "notification_topic_arn" {
  description = "The ARN of the SNS topic for application notifications"
  value       = aws_sns_topic.notifications.arn
}

