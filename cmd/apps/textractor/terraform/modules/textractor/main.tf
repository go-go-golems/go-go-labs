
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

# Create ZIP file for Lambda function
data "archive_file" "lambda_zip" {
  type        = "zip"
  output_path = "${path.module}/lambda/textract-processor.zip"
  source {
    content  = file("${path.module}/lambda/index.js")
    filename = "index.js"
  }
}

# Update Lambda Function resource
resource "aws_lambda_function" "textract_processor" {
  filename         = data.archive_file.lambda_zip.output_path
  source_code_hash = data.archive_file.lambda_zip.output_base64sha256
  function_name    = "${var.prefix}-textract-processor"
  role            = aws_iam_role.lambda_role.arn
  handler         = "index.handler"
  runtime         = "nodejs16.x"
  timeout         = 30

  environment {
    variables = {
      OUTPUT_QUEUE_URL = aws_sqs_queue.output_queue.url
      SNS_TOPIC_ARN   = aws_sns_topic.textract_completion.arn
      AWS_LAMBDA_ROLE = aws_iam_role.lambda_role.arn
    }
  }
}

# EventBridge Rule
resource "aws_cloudwatch_event_rule" "pdf_upload" {
  name        = "${var.prefix}-pdf-upload"
  description = "Capture PDF uploads to S3"

  event_pattern = jsonencode({
    source      = ["aws.s3"]
    detail-type = ["AWS API Call via CloudTrail"]
    detail = {
      eventSource = ["s3.amazonaws.com"]
      eventName   = ["PutObject"]
      requestParameters = {
        bucketName = [aws_s3_bucket.document_bucket.id]
      }
    }
  })
}

resource "aws_cloudwatch_event_target" "lambda" {
  rule      = aws_cloudwatch_event_rule.pdf_upload.name
  target_id = "SendToLambda"
  arn       = aws_lambda_function.textract_processor.arn
}

# IAM Roles and Policies
resource "aws_iam_role" "lambda_role" {
  name = "${var.prefix}-lambda-role"

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

resource "aws_iam_role_policy" "lambda_policy" {
  name = "${var.prefix}-lambda-policy"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "textract:AnalyzeDocument",
          "textract:GetDocumentAnalysis",
          "s3:GetObject",
          "s3:PutObject",
          "sns:Publish",
          "sqs:SendMessage",
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage"
        ]
        Resource = [
          aws_s3_bucket.document_bucket.arn,
          "${aws_s3_bucket.document_bucket.arn}/*",
          aws_sns_topic.textract_completion.arn,
          aws_sqs_queue.input_queue.arn,
          aws_sqs_queue.output_queue.arn
        ]
      }
    ]
  })
} 