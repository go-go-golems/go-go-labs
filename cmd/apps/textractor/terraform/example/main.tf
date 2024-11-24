provider "aws" {
  region = "us-east-1"  # Change this to your desired region
}

module "textractor" {
  source = "../modules/textractor"

  prefix      = "textractor-dev"
  environment = "dev"
  bucket_name = "textractor-documents-dev-12345"  # Must be globally unique
}

# S3 outputs
output "bucket_name" {
  value = module.textractor.bucket_name
}

output "document_bucket" {
  description = "The name of the S3 bucket used for document storage"
  value       = module.textractor.document_bucket
}

output "document_bucket_arn" {
  description = "The ARN of the S3 bucket used for document storage"
  value       = module.textractor.document_bucket_arn
}

# Queue outputs
output "input_queue_url" {
  value = module.textractor.input_queue_url
}

output "completion_queue_url" {
  value = module.textractor.completion_queue_url
}

output "notifications_queue_url" {
  value = module.textractor.notifications_queue_url
}

# SNS outputs
output "sns_topic_arn" {
  value = module.textractor.sns_topic_arn
}

# Lambda outputs for document processor
output "document_processor_arn" {
  value = module.textractor.document_processor_arn
}

output "document_processor_name" {
  value = module.textractor.document_processor_name
}

output "document_processor_log_group" {
  value = module.textractor.document_processor_log_group
}

# Lambda outputs for completion processor
output "completion_processor_arn" {
  value = module.textractor.completion_processor_arn
}

output "completion_processor_name" {
  value = module.textractor.completion_processor_name
}

output "completion_processor_log_group" {
  value = module.textractor.completion_processor_log_group
}

# Other outputs
output "region" {
  value = module.textractor.region
}

output "jobs_table_name" {
  value = module.textractor.jobs_table_name
}

output "cloudtrail_log_group" {
  value = module.textractor.cloudtrail_log_group
}

# DLQ outputs
output "input_dlq_url" {
  description = "The URL of the input Dead Letter Queue"
  value       = module.textractor.input_dlq_url
}

output "completion_dlq_url" {
  description = "The URL of the completion Dead Letter Queue"
  value       = module.textractor.completion_dlq_url
}

output "notification_topic_arn" {
  description = "The ARN of the SNS topic for application notifications"
  value = module.textractor.notification_topic_arn
}

# Output bucket outputs
output "output_bucket_name" {
  description = "The name of the S3 bucket used for Textract outputs"
  value       = module.textractor.output_bucket_name
}

output "output_bucket_arn" {
  description = "The ARN of the S3 bucket used for Textract outputs"
  value       = module.textractor.output_bucket_arn
}
