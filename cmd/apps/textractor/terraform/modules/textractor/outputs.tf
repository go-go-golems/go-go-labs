output "bucket_name" {
  description = "Name of the created S3 bucket"
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

output "lambda_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.textract_processor.arn
}

output "function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.textract_processor.function_name
}

output "region" {
  description = "Region of the Lambda function"
  value       = data.aws_region.current.name
}

output "lambda_log_group" {
  description = "Name of the CloudWatch Log Group for the Lambda function"
  value       = aws_cloudwatch_log_group.lambda_logs.name
}

output "cloudtrail_log_group" {
  description = "Name of the CloudWatch Log Group for CloudTrail"
  value       = aws_cloudwatch_log_group.cloudtrail_logs.name
}

# Add DynamoDB table outputs
output "jobs_table_name" {
  description = "Name of the DynamoDB jobs table"
  value       = aws_dynamodb_table.jobs.name
}

output "jobs_table_arn" {
  description = "ARN of the DynamoDB jobs table"
  value       = aws_dynamodb_table.jobs.arn
} 