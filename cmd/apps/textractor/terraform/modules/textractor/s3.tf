# S3 Buckets and related policies
resource "aws_s3_bucket" "document_bucket" {
  bucket = var.bucket_name

  tags = {
    Name        = "Textractor Document Bucket"
    Environment = var.environment
  }
}

resource "aws_s3_bucket" "output_bucket" {
  bucket = "${var.bucket_name}-output"

  tags = {
    Name        = "Textractor Output Bucket"
    Environment = var.environment
  }
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

resource "aws_s3_bucket_policy" "document_bucket_policy" {
  bucket = aws_s3_bucket.document_bucket.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid = "AllowTextractAccess"
        Effect = "Allow"
        Principal = {
          Service = "textract.amazonaws.com"
        }
        Action = [
          "s3:GetObject"
        ]
        Resource = "${aws_s3_bucket.document_bucket.arn}/*"
      }
    ]
  })
}

resource "aws_s3_bucket_policy" "output_bucket_policy" {
  bucket = aws_s3_bucket.output_bucket.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid = "AllowTextractAccess"
        Effect = "Allow"
        Principal = {
          Service = "textract.amazonaws.com"
        }
        Action = [
          "s3:PutObject"
        ]
        Resource = "${aws_s3_bucket.output_bucket.arn}/*"
      }
    ]
  })
} 