## Debugging

### Lambda Function
1. Monitor Lambda execution in AWS Console:
   ```bash
   # Get recent log events
   aws logs get-log-events --log-group-name "/aws/lambda/<function-name>" --log-stream-name $(aws logs describe-log-streams --log-group-name "/aws/lambda/<function-name>" --order-by LastEventTime --descending --limit 1 --query 'logStreams[0].logStreamName' --output text)

   # Watch logs in real-time
   aws logs tail "/aws/lambda/<function-name>" --follow

   # Get Lambda metrics for the last hour
   aws cloudwatch get-metric-statistics --namespace AWS/Lambda \
     --metric-name Errors --statistics Sum \
     --start-time $(date -u -v-1H +"%Y-%m-%dT%H:%M:%SZ") \
     --end-time $(date -u +"%Y-%m-%dT%H:%M:%SZ") \
     --period 300 \
     --dimensions Name=FunctionName,Value=<function-name>

   # Test Lambda function directly
   aws lambda invoke --function-name <function-name> \
     --payload '{"Records":[{"s3":{"bucket":{"name":"<bucket-name>"},"object":{"key":"test.pdf"}}}]}' \
     response.json
   ```

### Queue Monitoring
1. Check SQS queues:
   ```bash
   # Get number of messages in queue
   aws sqs get-queue-attributes \
     --queue-url <queue-url> \
     --attribute-names ApproximateNumberOfMessages

   # Receive messages from queue
   aws sqs receive-message --queue-url <queue-url> --max-number-of-messages 10

   # View dead-letter queue messages
   aws sqs receive-message --queue-url <dlq-url> --max-number-of-messages 10

   # Purge queue if needed
   aws sqs purge-queue --queue-url <queue-url>
   ```

### S3 Events
1. Verify S3 event triggering:
   ```bash
   # Enable S3 access logging
   aws s3api put-bucket-logging \
     --bucket <bucket-name> \
     --bucket-logging-status '{"LoggingEnabled":{"TargetBucket":"<logging-bucket>","TargetPrefix":"logs/"}}'

   # Get bucket notification configuration
   aws s3api get-bucket-notification-configuration --bucket <bucket-name>

   # List recent CloudTrail events for bucket
   aws cloudtrail lookup-events \
     --lookup-attributes AttributeKey=ResourceName,AttributeValue=<bucket-name>
   ```

### SNS Topic
1. Monitor SNS delivery:
   ```bash
   # List subscriptions
   aws sns list-subscriptions-by-topic --topic-arn <topic-arn>

   # Publish test message
   aws sns publish \
     --topic-arn <topic-arn> \
     --message '{"test": "message"}'

   # Get topic attributes
   aws sns get-topic-attributes --topic-arn <topic-arn>

   # Check subscription attributes
   aws sns get-subscription-attributes --subscription-arn <subscription-arn>
   ```

### End-to-End Testing
1. Upload a test PDF:
   ```bash
   # Upload file
   aws s3 cp test.pdf s3://<bucket-name>/input/test.pdf

   # Monitor S3 events
   aws s3api get-object-tagging --bucket <bucket-name> --key input/test.pdf

   # Check Lambda logs
   aws logs tail "/aws/lambda/<function-name>" --follow

   # Monitor SQS messages
   aws sqs receive-message --queue-url <output-queue-url> --wait-time-seconds 20
   ```

### Using CloudWatch Logs Insights
Query example to track document processing:
```bash
# Start query
aws logs start-query \
  --log-group-name "/aws/lambda/<function-name>" \
  --start-time $(date -v-1H +%s) \
  --end-time $(date +%s) \
  --query-string 'fields @timestamp, @message | filter @message like /Processing document/ | sort @timestamp desc | limit 20'

# Get query results (use query-id from previous command)
aws logs get-query-results --query-id <query-id>
```

### Metrics to Monitor
```bash
# Get Lambda metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/Lambda \
  --metric-name Duration \
  --statistics Average Maximum \
  --start-time $(date -v-1H +"%Y-%m-%dT%H:%M:%SZ") \
  --end-time $(date +"%Y-%m-%dT%H:%M:%SZ") \
  --period 300 \
  --dimensions Name=FunctionName,Value=<function-name>

# Get SQS metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/SQS \
  --metric-name ApproximateNumberOfMessagesVisible \
  --statistics Average \
  --start-time $(date -v-1H +"%Y-%m-%dT%H:%M:%SZ") \
  --end-time $(date +"%Y-%m-%dT%H:%M:%SZ") \
  --period 300 \
  --dimensions Name=QueueName,Value=<queue-name>

# Get SNS metrics
aws cloudwatch get-metric-statistics \
  --namespace AWS/SNS \
  --metric-name NumberOfNotificationsFailed \
  --statistics Sum \
  --start-time $(date -v-1H +"%Y-%m-%dT%H:%M:%SZ") \
  --end-time $(date +"%Y-%m-%dT%H:%M:%SZ") \
  --period 300 \
  --dimensions Name=TopicName,Value=<topic-name>
```

### Helper Script
Create a debug.sh script:
```bash
#!/bin/bash

FUNCTION_NAME="<function-name>"
BUCKET_NAME="<bucket-name>"
QUEUE_URL="<queue-url>"
TOPIC_ARN="<topic-arn>"

case $1 in
  "lambda")
    aws logs tail "/aws/lambda/$FUNCTION_NAME" --follow
    ;;
  "queue")
    aws sqs get-queue-attributes \
      --queue-url $QUEUE_URL \
      --attribute-names ApproximateNumberOfMessages
    ;;
  "upload")
    aws s3 cp "$2" "s3://$BUCKET_NAME/input/"
    ;;
  "test")
    aws lambda invoke \
      --function-name $FUNCTION_NAME \
      --payload '{"test":"event"}' \
      response.json
    ;;
  *)
    echo "Usage: $0 {lambda|queue|upload|test}"
    exit 1
    ;;
esac
```

### Common Issues
- IAM permissions: Ensure Lambda has correct permissions for S3, Textract, SNS
- Queue visibility timeout: Adjust if processing takes longer than expected
- Lambda timeout: Check if function times out during Textract processing
- S3 event configuration: Verify event notifications are properly set up
- SNS subscription confirmation: Ensure SQS queue subscription is confirmed
