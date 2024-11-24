provider "aws" {
  region = "us-east-1"  # Change this to your desired region
}

module "textractor" {
  source = "../modules/textractor"

  prefix      = "textractor-dev"
  environment = "dev"
  bucket_name = "textractor-documents-dev-12345"  # Must be globally unique
}

output "bucket_name" {
  value = module.textractor.bucket_name
}

output "input_queue_url" {
  value = module.textractor.input_queue_url
}

output "output_queue_url" {
  value = module.textractor.output_queue_url
}

output "sns_topic_arn" {
  value = module.textractor.sns_topic_arn
}

output "lambda_arn" {
  value = module.textractor.lambda_arn
}

output "function_name" {
  value = module.textractor.function_name
}

output "region" {
  value = module.textractor.region
}

output "lambda_log_group" {
  description = "Name of the CloudWatch Log Group for the Lambda function"
  value       = module.textractor.lambda_log_group
}

output "cloudtrail_log_group" {
  description = "Name of the CloudWatch Log Group for CloudTrail"
  value       = module.textractor.cloudtrail_log_group
}

# Add DynamoDB jobs table outputs
output "jobs_table_name" {
  description = "Name of the DynamoDB jobs table"
  value       = module.textractor.jobs_table_name
}

output "jobs_table_arn" {
  description = "ARN of the DynamoDB jobs table"
  value       = module.textractor.jobs_table_arn
} 