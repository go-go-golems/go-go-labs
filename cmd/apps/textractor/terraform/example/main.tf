provider "aws" {
  region = "us-east-1"  # Change this to your desired region
}

module "textractor" {
  source = "../modules/textractor"

  prefix      = "textractor-dev"
  environment = "dev"
  bucket_name = "textractor-documents-dev-12345"  # Must be globally unique
}

# S3 and SQS outputs
output "bucket_name" {
  value = module.textractor.bucket_name
}

output "input_queue_url" {
  value = module.textractor.input_queue_url
}

output "output_queue_url" {
  value = module.textractor.output_queue_url
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