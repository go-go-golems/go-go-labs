# Consumer Groups with XREADGROUP

## 1. Creating Consumer Groups

```bash
# Create a consumer group (last parameter is the start ID)
XGROUP CREATE mystream mygroup 0

# Create consumer group starting from new messages only
XGROUP CREATE mystream mygroup $

# Check consumer groups
XINFO GROUPS mystream
```

## 2. Basic Consumer Group Operations

```bash
# Add some test messages
XADD mystream * order 1 product "laptop" user alice
XADD mystream * order 2 product "mouse" user bob
XADD mystream * order 3 product "keyboard" user charlie

# Consumer 1 reads
XREADGROUP GROUP mygroup consumer1 COUNT 1 STREAMS mystream >

# Consumer 2 reads
XREADGROUP GROUP mygroup consumer2 COUNT 1 STREAMS mystream >

# Check pending messages
XPENDING mystream mygroup
```

## 3. Understanding XREADGROUP Behavior

```bash
# '>' means "messages never delivered to any consumer"
XREADGROUP GROUP mygroup worker1 STREAMS mystream >

# Specific ID means "read my pending messages"
XREADGROUP GROUP mygroup worker1 STREAMS mystream 0

# Block for new messages
XREADGROUP GROUP mygroup worker1 BLOCK 1000 STREAMS mystream >
```

## 4. Message Acknowledgment

```bash
# After processing, acknowledge the message
# First, read a message and note the ID
XREADGROUP GROUP mygroup worker1 COUNT 1 STREAMS mystream >

# Acknowledge the message (replace ID with actual one)
XACK mystream mygroup <message-id>

# Check pending messages again
XPENDING mystream mygroup
```

## 5. Consumer Group Management

```bash
# List consumers
XINFO CONSUMERS mystream mygroup

# Delete a consumer
XGROUP DELCONSUMER mystream mygroup worker1

# Destroy a consumer group
XGROUP DESTROY mystream mygroup

# Set the consumer group to a new ID
XGROUP SETID mystream mygroup 0
```

## 6. Practical Example: Order Processing

```bash
# Setup for order processing
DEL orders
XGROUP CREATE orders order_group 0 MKSTREAM

# Add orders
XADD orders * order_id 1001 customer "Alice" amount 299.99
XADD orders * order_id 1002 customer "Bob" amount 149.50
XADD orders * order_id 1003 customer "Charlie" amount 599.99

# Worker processes
XREADGROUP GROUP order_group payment_processor COUNT 1 STREAMS orders >
XREADGROUP GROUP order_group inventory_checker COUNT 1 STREAMS orders >

# Acknowledge when processed
XACK orders order_group <message-id>
```

## Practice Scenario

```bash
# Clean up
DEL mystream
XGROUP DESTROY mystream mygroup

# Create new scenario
XGROUP CREATE mystream mygroup 0 MKSTREAM

# Simulate multiple consumers
XADD mystream * event "user_signup" email "john@example.com"
XADD mystream * event "purchase" amount 99.99

# Consumer 1 processes signup
XREADGROUP GROUP mygroup email_service COUNT 1 STREAMS mystream >

# Consumer 2 processes purchases
XREADGROUP GROUP mygroup payment_service COUNT 1 STREAMS mystream >
```