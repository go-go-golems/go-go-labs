# DynamoDB table for job tracking
resource "aws_dynamodb_table" "jobs" {
  name           = "${var.prefix}-jobs"
  billing_mode   = "PAY_PER_REQUEST"
  hash_key       = "JobID"
  
  attribute {
    name = "JobID"
    type = "S"
  }
  
  attribute {
    name = "Status"
    type = "S"
  }
  
  attribute {
    name = "SubmittedAt"
    type = "S"
  }
  
  attribute {
    name = "DocumentKey"
    type = "S"
  }

  # GSI1: Status-SubmittedAt for listing by status
  global_secondary_index {
    name               = "Status-SubmittedAt-Index"
    hash_key           = "Status"
    range_key          = "SubmittedAt"
    projection_type    = "ALL"
  }

  # GSI2: DocumentKey for looking up jobs by document
  global_secondary_index {
    name               = "DocumentKey-Index"
    hash_key           = "DocumentKey"
    projection_type    = "ALL"
  }

  tags = var.tags

  point_in_time_recovery {
    enabled = true
  }

  ttl {
    enabled        = true
    attribute_name = "TTL"
  }
} 