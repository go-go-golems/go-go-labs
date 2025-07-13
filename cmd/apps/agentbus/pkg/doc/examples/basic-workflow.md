---
Title: Basic Agent Workflow Example
Slug: basic-workflow
Short: A simple example showing how agents coordinate using AgentBus
Topics:
- coordination
- examples
- workflow
Commands:
- speak
- overhear
- announce
- await
- satisfy
Flags:
- agent
- topic
- timeout
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Example
---

This example demonstrates a basic 3-agent workflow where agents coordinate to build, test, and deploy code.

## Scenario

We have three agents working together:
- `build-agent` - Compiles the code
- `test-agent` - Runs tests after build completes
- `deploy-agent` - Deploys after tests pass

## Step-by-Step Workflow

### 1. Build Agent Starts Work

```bash
# Build agent announces it's starting (using project prefix for isolation)
AGENT_ID=build-agent PROJECT_PREFIX=myproject agentbus announce --flag building
AGENT_ID=build-agent PROJECT_PREFIX=myproject agentbus speak --topic build --msg "Starting compilation"

# ... actual build happens here ...

# Build completes
AGENT_ID=build-agent PROJECT_PREFIX=myproject agentbus speak --topic build --msg "Build completed successfully"
AGENT_ID=build-agent PROJECT_PREFIX=myproject agentbus satisfy --flag building
```

### 2. Test Agent Waits and Runs Tests

```bash
# Wait for build to complete (15 minute timeout)
AGENT_ID=test-agent PROJECT_PREFIX=myproject agentbus await --flag building --timeout 900

# Start testing
AGENT_ID=test-agent PROJECT_PREFIX=myproject agentbus announce --flag testing
AGENT_ID=test-agent PROJECT_PREFIX=myproject agentbus speak --topic test --msg "Running test suite"

# ... tests run here ...

# Tests complete
AGENT_ID=test-agent PROJECT_PREFIX=myproject agentbus speak --topic test --msg "All tests passed ✅"
AGENT_ID=test-agent PROJECT_PREFIX=myproject agentbus satisfy --flag testing
```

### 3. Deploy Agent Waits and Deploys

```bash
# Wait for tests to complete (30 minute timeout)
AGENT_ID=deploy-agent PROJECT_PREFIX=myproject agentbus await --flag testing --timeout 1800

# Start deployment
AGENT_ID=deploy-agent PROJECT_PREFIX=myproject agentbus announce --flag deploying
AGENT_ID=deploy-agent PROJECT_PREFIX=myproject agentbus speak --topic deploy --msg "Starting deployment to production"

# ... deployment happens here ...

# Deployment complete
AGENT_ID=deploy-agent PROJECT_PREFIX=myproject agentbus speak --topic deploy --msg "Deployment successful 🚀"
AGENT_ID=deploy-agent PROJECT_PREFIX=myproject agentbus satisfy --flag deploying
```

### 4. Monitor the Workflow

From another terminal, you can monitor the entire workflow:

```bash
# Monitor all activity in real-time
AGENT_ID=monitor-agent PROJECT_PREFIX=myproject agentbus monitor

# Or monitor specific topics
AGENT_ID=monitor-agent PROJECT_PREFIX=myproject agentbus overhear --topic build --follow
AGENT_ID=monitor-agent PROJECT_PREFIX=myproject agentbus overhear --topic coordination --follow
```

## Expected Output

When monitoring, you'll see messages like:

```
NEW: [build-agent 2025-01-12 15:30:15] 🚩 Announced flag: building
NEW: [build-agent 2025-01-12 15:30:16] Starting compilation
NEW: [build-agent 2025-01-12 15:32:45] Build completed successfully
NEW: [build-agent 2025-01-12 15:32:46] ✅ Satisfied flag: building
NEW: [test-agent 2025-01-12 15:32:47] 🚩 Announced flag: testing
NEW: [test-agent 2025-01-12 15:32:48] Running test suite
NEW: [test-agent 2025-01-12 15:35:12] All tests passed ✅
NEW: [test-agent 2025-01-12 15:35:13] ✅ Satisfied flag: testing
NEW: [deploy-agent 2025-01-12 15:35:14] 🚩 Announced flag: deploying
NEW: [deploy-agent 2025-01-12 15:35:15] Starting deployment to production
NEW: [deploy-agent 2025-01-12 15:37:30] Deployment successful 🚀
NEW: [deploy-agent 2025-01-12 15:37:31] ✅ Satisfied flag: deploying
```

## Key Concepts Demonstrated

1. **Sequential Dependencies**: Test agent waits for build, deploy agent waits for tests
2. **Coordination Flags**: Each agent announces work and satisfies when done
3. **Communication Stream**: Status updates broadcast to all agents
4. **Auto-published Events**: Coordination events appear with emoji indicators
5. **Monitoring**: Real-time visibility into the entire workflow
6. **Project Isolation**: Using PROJECT_PREFIX ensures this workflow won't interfere with others

## Running This Example

1. Start Redis: `docker run -d -p 6379:6379 redis:alpine`
2. Set up environment variables: `export AGENT_ID=your-agent PROJECT_PREFIX=myproject`
3. Run each agent in a separate terminal (with appropriate AGENT_ID values)
4. Monitor from a fourth terminal
5. Agents will coordinate automatically through Redis

**Important:** All agents in the same workflow must use the same PROJECT_PREFIX for proper coordination.

This basic pattern scales to complex multi-agent workflows with dozens of agents and dependencies.
