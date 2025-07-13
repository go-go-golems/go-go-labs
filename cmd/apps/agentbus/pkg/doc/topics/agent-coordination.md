---
Title: Agent Coordination with AgentBus
Slug: agent-coordination
Short: Learn how to coordinate multiple coding agents using Redis-backed communication
Topics:
- coordination
- agents
- redis
- communication
Commands:
- speak
- overhear
- jot
- recall
- announce
- await
- satisfy
Flags:
- agent
- channel
- timeout
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

AgentBus provides a Redis-backed coordination layer that enables multiple coding agents to work together effectively. This guide explains how to use the three core coordination primitives to build robust multi-agent workflows.

## Overview

AgentBus offers three coordination mechanisms:

1. **Communication Stream** (`speak`/`overhear`) - Real-time messaging on shared channel
2. **Knowledge Snippets** (`jot`/`recall`/`list`) - Shared documentation and notes
3. **Coordination Flags** (`announce`/`await`/`satisfy`) - Dependency management

All coordination operations automatically publish status messages to the shared communication stream for visibility.

## Setting Up Agent Identity

Every agent must identify itself with an `AGENT_ID`. This can be set via:

```bash
export AGENT_ID="build-agent-1"
# or use the --agent flag
agentbus speak --agent "build-agent-1" --channel build --msg "Starting build"
```

All operations are namespaced by agent ID to prevent conflicts and provide clear attribution.

## Communication Stream: Real-time Messaging

All agents share a single communication stream. Messages can include optional topic slugs for categorization.

### Broadcasting Status Updates

Use `speak` to broadcast status updates to other agents:

```bash
# Build agent announces build start
agentbus speak --msg "Starting compilation of main.go" --topic "build"

# Test agent reports test results  
agentbus speak --msg "All unit tests passed ✅" --topic "testing"

# Deploy agent shares deployment status
agentbus speak --msg "Deployment to staging complete" --topic "deploy"
```

### Monitoring Other Agents

Use `overhear` to monitor what other agents are doing. The system now shows **NEW:** indicators for messages you haven't seen before:

```bash
# Check recent activity across all topics
agentbus overhear --max 10
# Output includes:
# NEW: [build-agent-1 2025-01-12 14:30:15] Starting compilation of main.go
# [test-agent-2 2025-01-12 14:28:03] Running unit tests...

# Monitor only build-related messages with real-time updates
agentbus overhear --topic "build" --follow

# Get last 5 deployment messages
agentbus overhear --topic "deploy" --max 5
```

**Understanding Message Format:**
- **NEW:** prefix indicates messages you haven't seen before
- **[agent-id timestamp]** provides automatic metadata for all messages
- Messages without NEW: are ones you've already seen in previous overhear calls

### Common Communication Patterns

**Status Broadcasting:**
```bash
agentbus speak --msg "Agent online and ready" --topic "status"
agentbus speak --msg "Processing batch job 1/10" --topic "status"
agentbus speak --msg "Agent going offline for maintenance" --topic "status"
```

**Error Reporting:**
```bash
agentbus speak --msg "Build failed: syntax error in main.go:42" --topic "errors"
agentbus speak --msg "High memory usage detected" --topic "alerts"
```

## Knowledge Snippets: Shared Documentation

### Storing Knowledge

Use `jot` to store reusable knowledge that other agents can access:

```bash
# Store build patterns
agentbus jot --key "docker-build-cmd" --value "docker build -t myapp:latest ." --tag "docker,build"

# Save configuration examples
agentbus jot --key "nginx-config" --value "$(cat nginx.conf)" --tag "config,nginx"

# Document API endpoints
agentbus jot --key "health-check-api" --value "GET /health" --tag "api,monitoring"

# Store troubleshooting guides
agentbus jot --key "debug-memory-leak" --value "$(cat debug-guide.md)" --tag "troubleshooting,memory"
```

### Retrieving Knowledge

Use `recall` to access stored knowledge:

```bash
# Get specific configuration
agentbus recall --key "docker-build-cmd"

# Find all build-related snippets
agentbus recall --tag "build" --latest 10

# Get troubleshooting guides
agentbus recall --tag "troubleshooting"

# Find API documentation
agentbus recall --tag "api,monitoring"
```

### Discovering Available Knowledge

Use `list` to see what knowledge snippets are available. The list command shows snippet summaries with metadata:

```bash
# List all available snippets with automatic metadata
agentbus list
# Output includes:
# docker-build-cmd (docker,build) - [build-agent 2025-01-12 14:25:32]
# nginx-config (config,nginx) - [deploy-agent 2025-01-12 14:20:15]

# List docker-related snippets
agentbus list --tag "docker" --latest 10

# Browse recent additions to see what's new
agentbus list --latest 20

# Get detailed view of specific snippets
agentbus list --tag "troubleshooting" --latest 5
```

**Effective List Usage:**
- Use `--latest` to see the most recently added snippets
- Combine tags to find specific knowledge: `--tag "docker,production"`
- Check list output regularly to discover what other agents have shared

### Knowledge Organization Tips

**Tag Hierarchically:**
- `docker,build,production` for production Docker builds
- `api,auth,jwt` for JWT authentication docs
- `deploy,staging,rollback` for staging rollback procedures

**Use Descriptive Keys:**
- `postgres-backup-script` not `backup`
- `kubernetes-deployment-yaml` not `k8s`
- `error-handling-pattern` not `errors`

## Coordination Flags: Dependency Management

### Announcing Work in Progress

Use `announce` to declare that your agent is working on something:

```bash
# Claim exclusive access to database migration
agentbus announce database-migration

# Signal start of integration tests
agentbus announce integration-testing

# Lock deployment process
agentbus announce deployment-lock
```

### Waiting for Dependencies

Use `await` to wait for other agents to complete their work:

```bash
# Wait for build to complete (15 minute timeout)
agentbus await build-complete --timeout 900

# Wait for database migration (no timeout)
agentbus await database-migration

# Wait for tests and then clean up the flag
agentbus await integration-testing --timeout 1800 --delete
```

### Completing Work

Use `satisfy` to signal completion and allow waiting agents to proceed:

```bash
# Signal build completion
agentbus satisfy build-complete

# Release database migration lock
agentbus satisfy database-migration

# Mark integration tests as done
agentbus satisfy integration-testing
```

## Multi-Agent Workflow Example

Here's a complete example of a three-agent deployment workflow:

### Agent 1: Build Agent
```bash
# Start build process
agentbus announce building
agentbus speak --msg "Starting build process" --topic "build"

# ... perform build ...

agentbus speak --msg "Build completed successfully" --topic "build"
agentbus satisfy building
```

### Agent 2: Test Agent
```bash
# Wait for build to complete
agentbus await building --timeout 900

agentbus announce testing
agentbus speak --msg "Starting test suite" --topic "testing"

# ... run tests ...

agentbus speak --msg "All tests passed" --topic "testing"
agentbus satisfy testing
```

### Agent 3: Deploy Agent
```bash
# Wait for tests to pass
agentbus await testing --timeout 1800

agentbus announce deploying
agentbus speak --msg "Starting deployment" --topic "deploy"

# ... perform deployment ...

agentbus speak --msg "Deployment successful" --topic "deploy"
agentbus satisfy deploying

# Store deployment info for next time
agentbus jot --key "last-deployment" --value "$(date): v1.2.3 deployed to production" --tag "deploy,production"
```

## Automatic Coordination Publishing

All coordination operations (`jot`, `announce`, `await`, `satisfy`) automatically publish status messages to the communication stream with emoji indicators:

- **🚩** `announce` operations: "Agent claimed flag"
- **🔍** `await` operations: "Agent waiting for flag"  
- **✅** `satisfy` operations: "Flag satisfied, waiting agents notified"
- **📝** `jot` operations: "Knowledge snippet stored"

Monitor coordination activity with:
```bash
# See all coordination events
agentbus overhear --topic "coordination" --max 20

# Follow coordination in real-time
agentbus overhear --topic "coordination" --follow
```

Example coordination stream output:
```
NEW: [build-agent 2025-01-12 14:30:15] 🚩 Announced flag: building
NEW: [test-agent 2025-01-12 14:30:20] 🔍 Awaiting flag: building  
NEW: [build-agent 2025-01-12 14:35:42] ✅ Satisfied flag: building
NEW: [deploy-agent 2025-01-12 14:36:00] 📝 Stored knowledge: deployment-steps
```

## Error Handling and Recovery

### Handling Timeouts
```bash
# Set reasonable timeouts for await commands
agentbus await build-complete --timeout 1800 || {
    echo "Build timeout - investigating"
    agentbus speak --msg "Build timeout detected" --topic "alerts"
    exit 1
}
```

### Cleanup Stale Flags
```bash
# Force override stuck flags if needed
agentbus announce deployment-lock --force

# Or check what's holding a flag
agentbus recall --tag "deploy" --latest 5
```

### Recovery Patterns
```bash
# Check system state
agentbus overhear --topic "status" --max 20
agentbus recall --tag "troubleshooting"

# Signal recovery
agentbus speak --msg "System recovered, resuming operations" --topic "status"
```

## Best Practices

1. **Use Clear Topic Names**: `build`, `test`, `deploy`, `alerts`, `status`, `coordination`
2. **Set Reasonable Timeouts**: Don't wait forever; 15-30 minutes max for most operations
3. **Always Satisfy Flags**: Clean up after yourself to prevent deadlocks
4. **Tag Consistently**: Use a consistent tagging scheme across agents
5. **Include Context**: Make messages and jots descriptive and actionable
6. **Monitor Communication**: Regularly check status and coordination topics
7. **Document Patterns**: Store successful patterns as jots for reuse
8. **Leverage Auto-Publishing**: Use `overhear --topic coordination` to monitor flag operations

### New Feature Best Practices

9. **Watch for NEW Indicators**: Pay attention to NEW: messages in `overhear` output to track what you haven't seen
10. **Use List Command for Discovery**: Regularly run `agentbus list --latest 10` to discover new knowledge snippets
11. **Monitor Coordination Stream**: Keep `agentbus overhear --topic coordination --follow` running to see real-time coordination activity
12. **Leverage Automatic Metadata**: All messages include agent ID and timestamp automatically - no need to add manually
13. **Follow Emoji Patterns**: Use the emoji system (🚩 🔍 ✅ 📝) to quickly identify coordination event types
14. **Reset Read Position**: If you miss important messages, you can start fresh with overhear to see all as NEW
15. **Combine Tools Effectively**: Use `list` to discover, `recall` to retrieve, and `overhear` to monitor activity

## Redis Configuration

AgentBus uses these Redis key patterns:
- `agentbus:ch:main` - Shared communication stream
- `agentbus:last:<agent>` - Agent read positions
- `agentbus:jot:<title>` - Knowledge snippets
- `agentbus:jots_by_tag:<tag>` - Tag indices
- `agentbus:flag:<name>` - Coordination flags

Ensure your Redis instance has sufficient memory and consider setting appropriate TTL policies for production use.
