# SQS Queues and related policies
resource "aws_sqs_queue" "input_queue" {
  name = "${var.prefix}-input-queue"

  visibility_timeout_seconds = 30
  message_retention_seconds = 86400
  
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.input_dlq.arn
    maxReceiveCount     = 3
  })
  
  tags = {
    Environment = var.environment
  }
}

resource "aws_sqs_queue" "completion_queue" {
  name = "${var.prefix}-completion-queue"

  visibility_timeout_seconds = 30
  message_retention_seconds = 86400
  
  redrive_policy = jsonencode({
    deadLetterTargetArn = aws_sqs_queue.completion_dlq.arn
    maxReceiveCount     = 3
  })
  
  tags = {
    Environment = var.environment
  }
}

resource "aws_sqs_queue" "notifications" {
  name = "${var.prefix}-notifications"
  visibility_timeout_seconds = 30
  message_retention_seconds = 3600  // 1 hour retention
  receive_wait_time_seconds = 20    // Enable long polling
  
  tags = {
    Environment = var.environment
  }
}

resource "aws_sqs_queue" "input_dlq" {
  name = "${var.prefix}-input-dlq"
  message_retention_seconds = 1209600  # 14 days
  
  tags = {
    Environment = var.environment
  }
}

resource "aws_sqs_queue" "completion_dlq" {
  name = "${var.prefix}-completion-dlq"
  message_retention_seconds = 1209600  # 14 days
  
  tags = {
    Environment = var.environment
  }
}

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

resource "aws_sqs_queue_policy" "notifications_policy" {
  queue_url = aws_sqs_queue.notifications.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action = "sqs:SendMessage"
        Resource = aws_sqs_queue.notifications.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn": aws_sns_topic.notifications.arn
          }
        }
      }
    ]
  })
} 