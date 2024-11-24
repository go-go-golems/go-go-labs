# SNS Topics and related policies
resource "aws_sns_topic" "textract_completion" {
  name = "${var.prefix}-textract-completion"
  
  lambda_success_feedback_role_arn    = aws_iam_role.sns_logging.arn
  lambda_success_feedback_sample_rate = 100
  lambda_failure_feedback_role_arn    = aws_iam_role.sns_logging.arn
  
  sqs_success_feedback_role_arn       = aws_iam_role.sns_logging.arn
  sqs_success_feedback_sample_rate    = 100
  sqs_failure_feedback_role_arn       = aws_iam_role.sns_logging.arn
  
  tags = var.tags
}

resource "aws_sns_topic" "notifications" {
  name = "${var.prefix}-notifications"
  
  tags = {
    Environment = var.environment
  }
}

resource "aws_sns_topic_policy" "textract_completion_policy" {
  arn = aws_sns_topic.textract_completion.arn

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = [
            "textract.amazonaws.com",
            "lambda.amazonaws.com"
          ]
        }
        Action = "sns:Publish"
        Resource = aws_sns_topic.textract_completion.arn
      }
    ]
  })
}

resource "aws_sns_topic_subscription" "textract_to_sqs" {
  topic_arn = aws_sns_topic.textract_completion.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.completion_queue.arn
}

resource "aws_sns_topic_subscription" "notifications_sqs" {
  topic_arn = aws_sns_topic.notifications.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.notifications.arn
}

# SNS Logging Role and Policy
resource "aws_iam_role" "sns_logging" {
  name = "${var.prefix}-sns-logging"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "sns_logging_policy" {
  name = "${var.prefix}-sns-logging-policy"
  role = aws_iam_role.sns_logging.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:PutMetricFilter",
          "logs:PutRetentionPolicy"
        ]
        Resource = [
          "${aws_cloudwatch_log_group.textract_sns_logs.arn}:*"
        ]
      }
    ]
  })
}

resource "aws_cloudwatch_log_group" "textract_sns_logs" {
  name              = "/aws/sns/${aws_sns_topic.textract_completion.name}"
  retention_in_days = 14
  
  tags = var.tags
} 