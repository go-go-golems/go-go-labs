# Basic Stream Operations

## 1. Creating Streams with XADD

```bash
# Connect to Redis
redis-cli

# Create a stream with XADD - returns the message ID
XADD mystream * name John age 30
XADD mystream * name Jane age 25 city "New York"
XADD mystream * sensor temperature value 22.5 location kitchen
```

## 2. Reading from Streams with XRANGE

```bash
# Read all messages from the beginning
XRANGE mystream - +

# Read messages from a specific ID
XRANGE mystream 1672531200000-0 +

# Read limited number of messages
XRANGE mystream - + COUNT 2
```

## 3. Reading from Streams with XREAD

```bash
# Read all new messages (blocking read)
XREAD COUNT 100 STREAMS mystream 0

# Read from multiple streams
XREAD STREAMS mystream anotherstream 0 0

# Block for new messages (timeout in ms)
XREAD BLOCK 5000 STREAMS mystream $

# Read from specific ID
XREAD STREAMS mystream 1672531200000-0
```

## 4. Understanding Message IDs

```bash
# Message ID format: milliseconds-sequence
# Example: 1672531200000-0

# Add messages to see IDs
XADD mystream * message "Hello"
XADD mystream * message "World"

# Check the last generated ID
XINFO STREAM mystream
```

## 5. Stream Length and Trimming

```bash
# Check stream length
XLEN mystream

# Add with automatic trimming (maxlen)
XADD mystream MAXLEN 100 * data "new message"

# Add with approximate trimming
XADD mystream MAXLEN ~ 100 * data "approximate"
```

## Practice Commands
Run these commands in redis-cli to understand the basics:

```bash
# Clean up
DEL mystream

# Create new stream
XADD mystream * action login user alice
XADD mystream * action logout user alice
XADD mystream * action login user bob

# Read everything
XRANGE mystream - +
```