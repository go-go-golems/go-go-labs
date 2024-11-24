variable "prefix" {
  description = "Prefix to be used for all resource names"
  type        = string
}

variable "environment" {
  description = "Environment name for tagging"
  type        = string
}

variable "bucket_name" {
  description = "Name of the S3 bucket for documents"
  type        = string
}

variable "tags" {
  description = "Tags to be applied to all resources"
  type        = map(string)
  default     = {}
} 