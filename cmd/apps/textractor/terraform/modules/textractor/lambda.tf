# Lambda Functions and related resources
data "archive_file" "document_processor_zip" {
  type        = "zip"
  output_path = "${path.module}/lambda/document-processor.zip"
  source_dir  = "${path.module}/lambda"
  excludes    = ["completion-processor.js", "*.zip"]
}

data "archive_file" "completion_processor_zip" {
  type        = "zip"
  output_path = "${path.module}/lambda/completion-processor.zip"
  source_dir  = "${path.module}/lambda"
  excludes    = ["document-processor.js", "*.zip"]
}

# Document processor Lambda
resource "aws_lambda_function" "document_processor" {
  filename         = data.archive_file.document_processor_zip.output_path
  source_code_hash = data.archive_file.document_processor_zip.output_base64sha256
  function_name    = "${var.prefix}-document-processor"
  role            = aws_iam_role.document_processor_role.arn
  handler         = "document-processor.handler"
  runtime         = "nodejs16.x"
  timeout         = 30

  environment {
    variables = {
      JOBS_TABLE = aws_dynamodb_table.jobs.name
      SNS_TOPIC_ARN = aws_sns_topic.textract_completion.arn
      TEXTRACT_ROLE_ARN = aws_iam_role.textract_role.arn
      NOTIFICATION_TOPIC_ARN = aws_sns_topic.notifications.arn
      OUTPUT_BUCKET = aws_s3_bucket.output_bucket.id
    }
  }

  tags = var.tags
}

# Completion processor Lambda
resource "aws_lambda_function" "completion_processor" {
  filename         = data.archive_file.completion_processor_zip.output_path
  source_code_hash = data.archive_file.completion_processor_zip.output_base64sha256
  function_name    = "${var.prefix}-completion-processor"
  role            = aws_iam_role.completion_processor_role.arn
  handler         = "completion-processor.handler"
  runtime         = "nodejs16.x"
  timeout         = 30

  environment {
    variables = {
      JOBS_TABLE = aws_dynamodb_table.jobs.name
      NOTIFICATION_TOPIC_ARN = aws_sns_topic.notifications.arn
      OUTPUT_BUCKET = aws_s3_bucket.output_bucket.id
    }
  }

  tags = var.tags
}

# Lambda Triggers
resource "aws_lambda_permission" "allow_input_sqs" {
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

resource "aws_lambda_permission" "allow_completion_sqs" {
  statement_id  = "AllowExecutionFromSQS"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.completion_processor.function_name
  principal     = "sqs.amazonaws.com"
  source_arn    = aws_sqs_queue.completion_queue.arn
}

resource "aws_lambda_event_source_mapping" "completion_sqs_trigger" {
  event_source_arn = aws_sqs_queue.completion_queue.arn
  function_name    = aws_lambda_function.completion_processor.arn
  batch_size       = 1
  enabled          = true
}

# CloudWatch Logs
resource "aws_cloudwatch_log_group" "document_processor_logs" {
  name              = "/aws/lambda/${aws_lambda_function.document_processor.function_name}"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "completion_processor_logs" {
  name              = "/aws/lambda/${aws_lambda_function.completion_processor.function_name}"
  retention_in_days = 14
}

# Lambda IAM Roles and Policies
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

# Lambda Policies
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
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject"
        ]
        Resource = [
          aws_s3_bucket.output_bucket.arn,
          "${aws_s3_bucket.output_bucket.arn}/*"
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
        Resource = [
          aws_sns_topic.notifications.arn,
          aws_sns_topic.textract_completion.arn
        ]
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
          "sns:Publish"
        ]
        Resource = [
          aws_sns_topic.notifications.arn,
          aws_sns_topic.textract_completion.arn
        ]
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

resource "aws_iam_role_policy" "lambda_logging" {
  name = "${var.prefix}-lambda-logging"
  role = aws_iam_role.document_processor_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "document_processor_logs" {
  role       = aws_iam_role.document_processor_role.name
  policy_arn = aws_iam_policy.lambda_logging.arn
}

resource "aws_iam_role_policy_attachment" "completion_processor_logs" {
  role       = aws_iam_role.completion_processor_role.name
  policy_arn = aws_iam_policy.lambda_logging.arn
} 