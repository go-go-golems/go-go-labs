# Pending Messages and Recovery

## 1. Understanding Pending Messages

```bash
# Setup
DEL notifications
XGROUP CREATE notifications alert_group 0 MKSTREAM

# Add messages
XADD notifications * alert "disk_full" server "web-01"
XADD notifications * alert "memory_high" server "db-01"
XADD notifications * alert "cpu_spike" server "app-01"

# Consumer reads but doesn't ack
XREADGROUP GROUP alert_group monitor1 COUNT 2 STREAMS notifications >

# Check pending messages
XPENDING notifications alert_group

# Detailed pending info
XPENDING notifications alert_group - + 10
```

## 2. Claiming Stuck Messages

```bash
# Claim messages from a failed consumer
XCLAIM notifications alert_group monitor2 100000 1672531200000-0

# Claim with more options
XCLAIM notifications alert_group monitor2 60000 1672531200000-0 JUSTID

# Auto-claim messages (Redis 6.2+)
XAUTOCLAIM notifications alert_group monitor2 60000 1672531200000-0
```

## 3. Inspecting Failed Deliveries

```bash
# See full pending message details
XPENDING notifications alert_group - + 10 monitor1

# Read the actual pending messages
XREADGROUP GROUP alert_group monitor1 STREAMS notifications 0

# Force claim after timeout
XCLAIM notifications alert_group monitor2 30000 1672531200000-0
```

## 4. Handling Consumer Failures

```bash
# Simulate consumer failure
XADD notifications * alert "disk_critical" server "web-02"

# Consumer reads but crashes before ack
XREADGROUP GROUP alert_group unstable_consumer COUNT 1 STREAMS notifications >

# Check pending (should show 3 messages from unstable_consumer)
XPENDING notifications alert_group

# Recover by claiming to healthy consumer
XCLAIM notifications alert_group stable_consumer 0 -

# Or use auto-claim for all old messages
XAUTOCLAIM notifications alert_group stable_consumer 30000 0-0
```

## 5. Message Retry Logic

```bash
# Setup retry scenario
DEL orders_retry
XGROUP CREATE orders_retry retry_group 0 MKSTREAM

# Add orders
XADD orders_retry * order 1001 status "pending"
XADD orders_retry * order 1002 status "pending"

# Consumer processes but fails to ack
XREADGROUP GROUP retry_group processor1 COUNT 1 STREAMS orders_retry >

# Check delivery count
XPENDING orders_retry retry_group - + 10

# After some time, claim the message for retry
XCLAIM orders_retry retry_group processor2 60000 1672531200000-0

# Acknowledge successful processing
XACK orders_retry retry_group 1672531200000-0
```

## 6. Monitoring and Cleanup

```bash
# Get stream info
XINFO STREAM notifications

# Get consumer group info
XINFO GROUPS notifications

# Get consumer details
XINFO CONSUMERS notifications alert_group

# Trim old messages (be careful!)
XTRIM notifications MAXLEN 1000

# Delete specific consumer
XGROUP DELCONSUMER notifications alert_group unstable_consumer

# Destroy entire group
XGROUP DESTROY notifications alert_group
```

## Practical Recovery Exercise

```bash
# Setup realistic failure scenario
DEL logs
XGROUP CREATE logs log_group 0 MKSTREAM

# Add log entries
XADD logs * level ERROR service web message "timeout"
XADD logs * level WARN service api message "rate_limit"
XADD logs * level ERROR service db message "connection_failed"

# Consumer reads some messages
XREADGROUP GROUP log_group analyzer1 COUNT 2 STREAMS logs >

# Simulate consumer death
# Now recover:

# 1. Check what's pending
XPENDING logs log_group

# 2. See which consumer has it
XPENDING logs log_group - + 10

# 3. Claim messages to new consumer
XCLAIM logs log_group analyzer2 0 -

# 4. Process and acknowledge
XACK logs log_group <message-id>
```