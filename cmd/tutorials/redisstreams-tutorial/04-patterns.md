# Real-World Patterns and Best Practices

## 1. Event Sourcing Pattern

```bash
# User activity stream
DEL user_events
XGROUP CREATE user_events user_group 0 MKSTREAM

# Add user events
XADD user_events * type signup user_id 123 email "user@example.com" timestamp 1672531200
XADD user_events * type login user_id 123 ip "192.168.1.1"
XADD user_events * type purchase user_id 123 product_id 456 amount 99.99

# Different consumers for different purposes
# Analytics consumer
XREADGROUP GROUP user_group analytics COUNT 1 STREAMS user_events >

# Email service consumer
XREADGROUP GROUP user_group email_service COUNT 1 STREAMS user_events >

# Audit log consumer
XREADGROUP GROUP user_group audit_log COUNT 1 STREAMS user_events >
```

## 2. Work Queue Pattern

```bash
# Job processing queue
DEL jobs
XGROUP CREATE jobs job_workers 0 MKSTREAM

# Add jobs with priorities
XADD jobs * type email recipient "user1@example.com" priority high template welcome
XADD jobs * type report user_id 123 format pdf period monthly
XADD jobs * type cleanup retention_days 30

# Multiple worker types
XREADGROUP GROUP job_workers email_worker COUNT 1 STREAMS jobs >
XREADGROUP GROUP job_workers report_worker COUNT 1 STREAMS jobs >
```

## 3. Fan-Out Pattern

```bash
# Create multiple consumer groups for same stream
XGROUP CREATE notifications email_group 0 MKSTREAM
XGROUP CREATE notifications sms_group 0
XGROUP CREATE notifications push_group 0

# Add notification
XADD notifications * user_id 123 type order_update order_id 456 status shipped

# Each service processes independently
XREADGROUP GROUP email_group email_service STREAMS notifications >
XREADGROUP GROUP sms_group sms_service STREAMS notifications >
XREADGROUP GROUP push_group push_service STREAMS notifications >
```

## 4. Idempotency Pattern

```bash
# Add unique identifier to prevent duplicates
XADD orders * idempotency_key abc123 user_id 456 amount 100 product laptop

# Consumer checks if already processed
XPENDING orders order_group - + 1000

# Skip if already processed
XACK orders order_group <message-id>
```

## 5. Dead Letter Queue Pattern

```bash
# Main processing stream
DEL main_processing
XGROUP CREATE main_processing main_group 0 MKSTREAM

# Dead letter queue
DEL dead_letter_queue
XGROUP CREATE dead_letter_queue dlq_group 0 MKSTREAM

# Add message
XADD main_processing * data "problematic_payload"

# After max retries, move to DLQ
XADD dead_letter_queue * original_stream main_processing original_id <id> error "max_retries_exceeded"
```

## 6. Monitoring Commands

```bash
# Monitor stream health
XINFO STREAM mystream

# Check consumer lag
XPENDING mystream mygroup

# Monitor consumer activity
XINFO CONSUMERS mystream mygroup

# Check stream length
XLEN mystream
```

## 7. Performance Tuning

```bash
# Batch processing
XREADGROUP GROUP mygroup worker1 COUNT 100 STREAMS mystream >

# Parallel consumers
# Terminal 1
XREADGROUP GROUP mygroup worker1 BLOCK 1000 STREAMS mystream >

# Terminal 2  
XREADGROUP GROUP mygroup worker2 BLOCK 1000 STREAMS mystream >

# Terminal 3
XREADGROUP GROUP mygroup worker3 BLOCK 1000 STREAMS mystream >
```

## 8. Complete Example: E-commerce Order Flow

```bash
# Setup order processing
DEL order_flow
XGROUP CREATE order_flow order_processing 0 MKSTREAM
XGROUP CREATE order_flow inventory_check 0
XGROUP CREATE order_flow payment_process 0

# Add order
XADD order_flow * order_id 1001 user_id 123 items '[{"sku":"LAPTOP","qty":1}]' total 999.99

# Step 1: Inventory check
XREADGROUP GROUP order_flow inventory_worker COUNT 1 STREAMS order_flow >
# Process inventory check
XACK order_flow inventory_check <message-id>

# Step 2: Payment processing
XREADGROUP GROUP order_flow payment_worker COUNT 1 STREAMS order_flow >
# Process payment
XACK order_flow payment_process <message-id>

# Step 3: Final processing
XREADGROUP GROUP order_flow fulfillment_worker COUNT 1 STREAMS order_flow >
```

## 9. Testing Commands

```bash
# Quick test setup
redis-cli FLUSHALL

# Create test environment
XGROUP CREATE test_stream test_group 0 MKSTREAM
XADD test_stream * test "data"
XREADGROUP GROUP test_group test_consumer STREAMS test_stream >
```

## 10. Cleanup Commands

```bash
# Clean up specific stream
DEL mystream

# Remove consumer group
XGROUP DESTROY mystream mygroup

# Clean all (development only)
FLUSHALL

# Remove old messages
XTRIM mystream MAXLEN 1000
```

## Next Steps
- Try implementing a real consumer application
- Experiment with different block times
- Test failover scenarios
- Monitor with redis-cli MONITOR command