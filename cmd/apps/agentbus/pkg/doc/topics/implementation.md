---
Title: AgentBus Implementation Details
Slug: implementation
Short: Technical details of how AgentBus is built on top of Redis
Topics:
- implementation
- redis
- architecture
- technical
Commands:
- all
Flags:
- redis-url
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

AgentBus is built on top of Redis, leveraging its data structures to provide a robust coordination layer for coding agents. This document explains the technical implementation details.

## Redis Data Structures Used

### 1. Redis Streams (Communication Channel)

**Key Pattern:** `<PROJECT_PREFIX>:ch:main`

AgentBus uses a single Redis Stream to handle all agent communication. This provides:
- **Ordering**: Messages are automatically ordered by timestamp
- **Persistence**: Messages persist until explicitly trimmed
- **Scalability**: Streams can handle high throughput
- **Per-consumer tracking**: Each agent tracks its own read position

**Message Structure:**
```json
{
  "agent_id": "build-agent-1",
  "message": "Build completed successfully",
  "topic": "build"    // optional
}
```

**Implementation Benefits:**
- Single stream simplifies coordination
- Topic filtering happens at read time
- No message loss between agent restarts
- Natural ordering for event reconstruction

### 2. Redis Hashes (Knowledge Snippets)

**Key Pattern:** `<PROJECT_PREFIX>:jot:<title>`

Each knowledge snippet is stored as a Redis Hash containing:
```json
{
  "value": "The actual content/documentation",
  "author": "documentation-agent",
  "timestamp": "1642694400",
  "tags": "docker,build,deployment"
}
```

**Why Hashes:**
- Atomic updates of all fields
- Efficient storage for structured data
- Easy to retrieve specific fields
- Built-in serialization

### 3. Redis Sorted Sets (Tag Indices)

**Key Pattern:** `<PROJECT_PREFIX>:jots_by_tag:<tag>`

For each tag, a sorted set maps jot keys to timestamps:
```
docker -> {"docker-build-cmd": 1642694400, "docker-deploy": 1642694500}
build  -> {"docker-build-cmd": 1642694400, "gradle-build": 1642694300}
```

**Score:** Unix timestamp (for chronological ordering)
**Member:** Jot key

**Benefits:**
- Fast tag-based lookups: O(log N)
- Automatic chronological ordering
- Efficient range queries for "latest N"
- Support for multiple tags per jot

### 4. Redis Strings (Coordination Flags)

**Key Pattern:** `<PROJECT_PREFIX>:flag:<name>`

Simple string values containing agent ID and timestamp:
```
<PROJECT_PREFIX>:flag:building -> "build-agent-1 @ 2024-01-20T15:30:00Z"
```

**Why Strings:**
- Atomic operations (SETNX, DEL)
- Simple existence checks
- Minimal overhead
- Clear semantics for locks

### 5. Redis Strings (Read Position Tracking)

**Key Pattern:** `<PROJECT_PREFIX>:last:<agent_id>`

Stores the last stream ID read by each agent:
```
<PROJECT_PREFIX>:last:build-agent-1 -> "1642694400123-0"
```

**Benefits:**
- Each agent tracks independently
- No interference between agents
- Resume from exact position after restart
- Efficient incremental reads

## Key Design Decisions

### Single Communication Channel

**Decision:** Use one shared Redis Stream instead of multiple channels per topic.

**Rationale:**
- Simplifies agent implementation (no channel management)
- Topics become message metadata instead of infrastructure
- Easier to monitor all agent activity
- Reduces Redis memory overhead
- Simplifies backup/restore operations

**Trade-offs:**
- Slightly more network traffic (agents see all messages)
- Filtering happens client-side
- Single point of scaling bottleneck

### Pull-Based Message Consumption

**Decision:** Agents pull messages using `XRANGE`/`XREAD` instead of pub/sub.

**Rationale:**
- Message persistence across agent restarts
- No message loss during network issues
- Each agent controls its read rate
- Easy to replay message history
- Supports backpressure naturally

**Trade-offs:**
- Slightly higher latency than pub/sub
- Requires agents to track read positions
- More Redis memory usage for message history

### Hierarchical Key Naming

**Decision:** Use prefixed, structured key names (`<PROJECT_PREFIX>:type:identifier`).

**Rationale:**
- Clear ownership and purpose
- Easy to scan/debug keys
- Supports multiple AgentBus projects in same Redis instance
- Natural data organization
- Simplified backup/restore
- Project isolation through PROJECT_PREFIX environment variable

### Auto-Publishing Coordination Events

**Decision:** Coordination commands automatically publish to the communication channel.

**Rationale:**
- Increases system observability
- Helps with debugging coordination issues
- Provides audit trail of operations
- No additional agent implementation required

### Comprehensive Debug Logging

**Decision:** Log all operations to `/tmp/agentbus.log` with structured logging.

**Rationale:**
- Essential for troubleshooting hanging `announce` operations
- Provides audit trail of Redis operations
- Helps debug timeout issues with `await` commands
- Structured format (JSON) for easy parsing
- Centralized logging location for all agents

### Safe Data Cleanup

**Decision:** Provide `clear` command for complete project data removal.

**Rationale:**
- Safe cleanup of all project data using PROJECT_PREFIX
- Requires explicit `--force` flag to prevent accidental deletion
- Essential for resetting coordination state between test runs
- Removes all streams, flags, jots, and tracking data

## Performance Characteristics

### Memory Usage

- **Streams**: ~32 bytes + message size per entry
- **Hashes**: ~24 bytes + field data per jot
- **Sorted Sets**: ~24 bytes per tag-jot relationship
- **Strings**: ~24 bytes + value size per flag/position

### Time Complexity

- **Speak**: O(1) - Single XADD operation
- **Overhear**: O(N) where N = number of new messages
- **Jot**: O(T) where T = number of tags
- **Recall by key**: O(1) - Single HGETALL
- **Recall by tag**: O(log N + M) where N = jots with tag, M = results
- **Announce/Satisfy**: O(1) - Simple string operations
- **Await**: O(1) per polling cycle

### Scaling Considerations

**Single Redis Instance Limits:**
- ~100K-1M messages/second depending on message size
- Memory limited by total messages + jots stored
- Single point of failure

**Horizontal Scaling Options:**
- Redis Cluster for data partitioning
- Separate Redis instances per agent group
- Stream-specific Redis instances
- Read replicas for overhear operations

## Redis Configuration Recommendations

### Memory Management
```redis
# Limit memory usage
maxmemory 2gb
maxmemory-policy allkeys-lru

# Stream trimming (optional)
# Trim communication stream to last 10000 messages
# XTRIM agentbus:ch:main MAXLEN ~ 10000
```

### Persistence
```redis
# For durability
save 900 1      # Save if at least 1 key changed in 900 seconds
save 300 10     # Save if at least 10 keys changed in 300 seconds  
save 60 10000   # Save if at least 10000 keys changed in 60 seconds

# Or use AOF for better durability
appendonly yes
appendfsync everysec
```

### Performance Tuning
```redis
# Increase client timeout for long-running awaits
timeout 1800

# Optimize for network
tcp-keepalive 300
tcp-nodelay yes
```

## Monitoring and Observability

### Key Metrics to Monitor

1. **Stream Length**: `XLEN <PROJECT_PREFIX>:ch:main`
2. **Memory Usage**: `INFO memory`
3. **Connected Agents**: `CLIENT LIST`
4. **Coordination Flags**: `KEYS <PROJECT_PREFIX>:flag:*`
5. **Knowledge Base Size**: `KEYS <PROJECT_PREFIX>:jot:*`

### Debug Commands

```bash
# Check stream contents
redis-cli XRANGE <PROJECT_PREFIX>:ch:main - + COUNT 10

# List all coordination flags
redis-cli KEYS "<PROJECT_PREFIX>:flag:*"

# Check agent read positions
redis-cli KEYS "<PROJECT_PREFIX>:last:*"

# Monitor real-time activity
redis-cli MONITOR

# Debug logging (check for hanging operations)
tail -f /tmp/agentbus.log
```

## Error Handling and Recovery

### Connection Failures
- Agents should retry with exponential backoff
- Commands are designed to be idempotent where possible
- Client-side connection pooling recommended

### Data Consistency
- All operations use Redis atomic commands
- No multi-step transactions required
- Flag operations use SETNX for conflict detection

### Recovery Procedures
- Agents automatically resume from last read position
- Stale flags can be manually cleaned up
- Knowledge snippets persist indefinitely
- Stream trimming for memory management

This implementation provides a robust, scalable foundation for agent coordination while remaining simple to understand and debug.
