output "bucket_name" {
  description = "Name of the S3 bucket"
  value       = aws_s3_bucket.document_bucket.id
}

output "input_queue_url" {
  description = "URL of the input SQS queue"
  value       = aws_sqs_queue.input_queue.url
}

output "output_queue_url" {
  description = "URL of the output SQS queue"
  value       = aws_sqs_queue.output_queue.url
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