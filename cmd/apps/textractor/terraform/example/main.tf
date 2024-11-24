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