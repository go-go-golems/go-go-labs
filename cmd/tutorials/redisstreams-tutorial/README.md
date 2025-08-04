# Redis Streams Tutorial

This tutorial teaches Redis Streams concepts through hands-on redis-cli commands. Learn XADD, XREAD, XREADGROUP, consumer groups, and more.

## Prerequisites
- Redis 5.0+ (Streams introduced in 5.0)
- redis-cli available

## Start Redis Server
```bash
# If Redis isn't running, start it:
redis-server
```

## Tutorial Sections
1. [Basic Stream Operations](01-basic-streams.md)
2. [Consumer Groups](02-consumer-groups.md)
3. [Pending Messages](03-pending-messages.md)
4. [Real-world Patterns](04-patterns.md)

## Quick Start
```bash
# Connect to Redis
redis-cli

# Test connection
PING
```