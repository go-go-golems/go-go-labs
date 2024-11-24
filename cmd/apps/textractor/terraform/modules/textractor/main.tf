# S3 Bucket
resource "aws_s3_bucket" "document_bucket" {
  bucket = var.bucket_name

  tags = {
    Name        = "Textractor Document Bucket"
    Environment = var.environment
  }
}

# SQS Queue Policy
resource "aws_sqs_queue_policy" "input_queue_policy" {
  queue_url = aws_sqs_queue.input_queue.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "s3.amazonaws.com"
        }
        Action = "sqs:SendMessage"
        Resource = aws_sqs_queue.input_queue.arn
        Condition = {
          ArnLike = {
            "aws:SourceArn": aws_s3_bucket.document_bucket.arn
          }
        }
      }
    ]
  })
}

resource "aws_s3_bucket_notification" "bucket_notification" {
  bucket = aws_s3_bucket.document_bucket.id

  queue {
    queue_arn     = aws_sqs_queue.input_queue.arn
    events        = ["s3:ObjectCreated:*"]
    filter_suffix = ".pdf"
  }

  depends_on = [aws_sqs_queue_policy.input_queue_policy]
}

# SQS Queues
resource "aws_sqs_queue" "input_queue" {
  name = "${var.prefix}-input-queue"

  visibility_timeout_seconds = 30
  message_retention_seconds = 86400
  
  tags = {
    Environment = var.environment
  }
}

resource "aws_sqs_queue" "output_queue" {
  name = "${var.prefix}-output-queue"

  visibility_timeout_seconds = 30
  message_retention_seconds = 86400
  
  tags = {
    Environment = var.environment
  }
}

# SNS Topic
resource "aws_sns_topic" "textract_completion" {
  name = "${var.prefix}-textract-completion"
}

resource "aws_sns_topic_subscription" "textract_to_sqs" {
  topic_arn = aws_sns_topic.textract_completion.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.output_queue.arn
}

# Create ZIP files for Lambda functions
data "archive_file" "document_processor_zip" {
  type        = "zip"
  output_path = "${path.module}/lambda/document-processor.zip"
  source {
    content  = file("${path.module}/lambda/document-processor.js")
    filename = "index.js"
  }
}

data "archive_file" "completion_processor_zip" {
  type        = "zip"
  output_path = "${path.module}/lambda/completion-processor.zip"
  source {
    content  = file("${path.module}/lambda/completion-processor.js")
    filename = "index.js"
  }
}

# Document processor Lambda
resource "aws_lambda_function" "document_processor" {
  filename         = data.archive_file.document_processor_zip.output_path
  source_code_hash = data.archive_file.document_processor_zip.output_base64sha256
  function_name    = "${var.prefix}-document-processor"
  role            = aws_iam_role.document_processor_role.arn
  handler         = "index.handler"
  runtime         = "nodejs16.x"
  timeout         = 30

  environment {
    variables = {
      SNS_TOPIC_ARN   = aws_sns_topic.textract_completion.arn
      AWS_LAMBDA_ROLE = aws_iam_role.document_processor_role.arn
      NOTIFICATIONS_QUEUE_URL = aws_sqs_queue.notifications.url
      JOBS_TABLE     = aws_dynamodb_table.jobs.name
      BUCKET_NAME    = aws_s3_bucket.document_bucket.id
    }
  }
}

# Completion processor Lambda
resource "aws_lambda_function" "completion_processor" {
  filename         = data.archive_file.completion_processor_zip.output_path
  source_code_hash = data.archive_file.completion_processor_zip.output_base64sha256
  function_name    = "${var.prefix}-completion-processor"
  role            = aws_iam_role.completion_processor_role.arn
  handler         = "index.handler"
  runtime         = "nodejs16.x"
  timeout         = 30

  environment {
    variables = {
      NOTIFICATIONS_QUEUE_URL = aws_sqs_queue.notifications.url
      JOBS_TABLE     = aws_dynamodb_table.jobs.name
      BUCKET_NAME    = aws_s3_bucket.document_bucket.id
    }
  }
}

# Add SNS trigger for completion processor
resource "aws_lambda_permission" "allow_sns" {
  statement_id  = "AllowExecutionFromSNS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.completion_processor.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.textract_completion.arn
}

resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.textract_completion.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.completion_processor.arn
}

# Update SQS trigger to point to document processor
resource "aws_lambda_permission" "allow_sqs" {
  statement_id  = "AllowExecutionFromSQS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.document_processor.function_name
  principal     = "sqs.amazonaws.com"
  source_arn    = aws_sqs_queue.input_queue.arn
}

resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn = aws_sqs_queue.input_queue.arn
  function_name    = aws_lambda_function.document_processor.arn
  batch_size       = 1
  enabled          = true
}

# IAM Roles and Policies
resource "aws_iam_role" "document_processor_role" {
  name = "${var.prefix}-document-processor-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# Document Processor Policy
resource "aws_iam_role_policy" "document_processor_policy" {
  name = "${var.prefix}-document-processor-policy"
  role = aws_iam_role.document_processor_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ChangeMessageVisibility"
        ]
        Resource = [aws_sqs_queue.input_queue.arn]
      },
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.document_bucket.arn,
          "${aws_s3_bucket.document_bucket.arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "textract:StartDocumentAnalysis"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = [aws_sns_topic.textract_completion.arn]
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:UpdateItem",
          "dynamodb:GetItem"
        ]
        Resource = [aws_dynamodb_table.jobs.arn]
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage"
        ]
        Resource = [aws_sqs_queue.notifications.arn]
      }
    ]
  })
}

# Completion Processor IAM Role
resource "aws_iam_role" "completion_processor_role" {
  name = "${var.prefix}-completion-processor-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }]
  })
}

# Completion Processor Policy
resource "aws_iam_role_policy" "completion_processor_policy" {
  name = "${var.prefix}-completion-processor-policy"
  role = aws_iam_role.completion_processor_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "textract:GetDocumentAnalysis"
        ]
        Resource = "*"
      },
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject"
        ]
        Resource = "${aws_s3_bucket.document_bucket.arn}/*"
      },
      {
        Effect = "Allow"
        Action = [
          "dynamodb:UpdateItem",
          "dynamodb:GetItem"
        ]
        Resource = [aws_dynamodb_table.jobs.arn]
      },
      {
        Effect = "Allow"
        Action = [
          "sqs:SendMessage"
        ]
        Resource = [aws_sqs_queue.notifications.arn]
      }
    ]
  })
}

# CloudWatch Logs for Document Processor
resource "aws_cloudwatch_log_group" "document_processor_logs" {
  name              = "/aws/lambda/${aws_lambda_function.document_processor.function_name}"
  retention_in_days = 14
}

resource "aws_iam_role_policy_attachment" "document_processor_logs" {
  role       = aws_iam_role.document_processor_role.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

# CloudWatch Logs for Completion Processor
resource "aws_cloudwatch_log_group" "completion_processor_logs" {
  name              = "/aws/lambda/${aws_lambda_function.completion_processor.function_name}"
  retention_in_days = 14
}

resource "aws_iam_role_policy_attachment" "completion_processor_logs" {
  role       = aws_iam_role.completion_processor_role.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

# Add necessary permissions to Lambda role
resource "aws_iam_policy" "lambda_logging" {
  name        = "${var.prefix}-lambda-logging"
  description = "IAM policy for logging from a lambda"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:GetLogEvents",
          "logs:StartQuery",
          "logs:GetQueryResults",
          "logs:DescribeLogStreams",
          "cloudtrail:LookupEvents"
        ]
        Resource = [
          # Lambda function logs
          "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:/aws/lambda/${var.prefix}-textract-processor:*",
          # Allow Logs Insights queries across all log groups
          "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "cloudtrail:LookupEvents"
        ]
        Resource = ["*"]
      }
    ]
  })
}

# Create CloudWatch Log Group for CloudTrail
resource "aws_cloudwatch_log_group" "cloudtrail_logs" {
  name              = "/aws/cloudtrail/${var.prefix}-textractor"
  retention_in_days = 14
}

# Add data sources if not already present
data "aws_region" "current" {}
data "aws_caller_identity" "current" {}

# Add DynamoDB table for job tracking
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

# Add DynamoDB permissions to Lambda role
resource "aws_iam_role_policy" "lambda_dynamodb" {
  name = "${var.prefix}-lambda-dynamodb"
  role = aws_iam_role.document_processor_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query",
          "dynamodb:Scan"
        ]
        Resource = [
          aws_dynamodb_table.jobs.arn,
          "${aws_dynamodb_table.jobs.arn}/index/*"
        ]
      }
    ]
  })
}

# Add notifications queue
resource "aws_sqs_queue" "notifications" {
  name = "${var.prefix}-notifications"
  visibility_timeout_seconds = 30
  message_retention_seconds = 3600  // 1 hour retention
  receive_wait_time_seconds = 20    // Enable long polling
  
  tags = {
    Environment = var.environment
  }
}

# Add notifications queue policy
resource "aws_sqs_queue_policy" "notifications_policy" {
  queue_url = aws_sqs_queue.notifications.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
        Action = "sqs:SendMessage"
        Resource = aws_sqs_queue.notifications.arn
      }
    ]
  })
} 