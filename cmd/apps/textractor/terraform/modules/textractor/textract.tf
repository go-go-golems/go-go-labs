# Textract related resources
resource "aws_iam_role" "textract_role" {
  name = "${var.prefix}-textract-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "textract.amazonaws.com"
        }
      }
    ]
  })

  tags = var.tags
}

resource "aws_iam_role_policy" "textract_sns" {
  name = "${var.prefix}-textract-sns"
  role = aws_iam_role.textract_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sns:Publish",
          "sns:PublishBatch"
        ]
        Resource = [
          aws_sns_topic.textract_completion.arn,
          aws_sns_topic.notifications.arn
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy" "textract_s3" {
  name = "${var.prefix}-textract-s3"
  role = aws_iam_role.textract_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:ListBucket",
          "s3:PutObject"
        ]
        Resource = [
          "${aws_s3_bucket.output_bucket.arn}/*"
        ]
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
    ]
  })
} 