# Mock Agent Fleet Agent

A sophisticated mock agent that simulates realistic AI coding agent behavior for testing and demonstrating the Agent Fleet backend system.

## Features

### Core Functionality
- **State Machine**: Implements realistic agent states (idle, active, waiting_feedback, error)
- **API Integration**: Full integration with Agent Fleet backend REST API
- **Command Processing**: Handles instructions, feedback, and questions from the fleet management system
- **Todo Management**: Creates and manages realistic todo lists for tasks
- **Progress Tracking**: Simulates work progress with metrics (files changed, lines added/removed)
- **Event Logging**: Comprehensive logging of agent activities and state changes

### Randomized Behavior Mode
- **Dynamic State Changes**: Random state transitions based on configurable probabilities
- **Scenario Execution**: Predefined realistic work scenarios with different complexity levels
- **Error Simulation**: Realistic error conditions and recovery patterns
- **Question Generation**: Context-aware questions requiring human feedback
- **Metric Fluctuations**: Realistic work progress simulation

### Predefined Scenarios
- **Bug Fix**: Authentication issue resolution with testing and deployment
- **Feature Development**: User dashboard implementation with frontend/backend work
- **Performance Optimization**: Database query optimization with benchmarking
- **Security Audit**: API endpoint security review and vulnerability fixes
- **Code Refactoring**: Legacy code modernization with architectural improvements
- **Infrastructure**: CI/CD pipeline enhancement with automation
- **Data Migration**: Database schema updates with zero-downtime strategies
- **Integration**: Third-party service integration with error handling

## Quick Start

### Build and Run

```bash
# Build the mock agent
cd cmd/apps/agent-fleet/mock-agent
go build -o mock-agent .

# Run with default settings (connects to localhost:8080)
./mock-agent

# Run with randomized behavior
./mock-agent --randomized

# Run with custom configuration
./mock-agent \
  --server http://backend-server:8080 \
  --name "Advanced Code Reviewer" \
  --worktree /projects/my-app \
  --randomized \
  --tick-interval 3 \
  --log-level debug
```

### Configuration Options

```bash
Usage:
  mock-agent [flags]

Flags:
      --command-check-interval int   Command check interval in seconds (default 2)
      --config string                config file
  -h, --help                         help for mock-agent
  -l, --log-level string             Log level (trace, debug, info, warn, error) (default "info")
  -n, --name string                  Agent name (random if not specified)
  -r, --randomized                   Enable randomized behavior mode
  -s, --server string                Fleet backend server URL (default "http://localhost:8080")
      --tick-interval int            Agent tick interval in seconds (default 5)
  -t, --token string                 Authentication token (default "fleet-agent-token-123")
  -w, --worktree string              Agent worktree path (default "/tmp/mock-project")
```

## Usage Examples

### Basic Agent
```bash
# Simple agent that registers and stays mostly idle
./mock-agent --name "Simple Helper"
```

### Randomized Agent
```bash
# Agent with full randomized behavior
./mock-agent --randomized --name "Chaos Tester" --tick-interval 2
```

### Development Testing
```bash
# Fast-paced agent for development testing
./mock-agent \
  --randomized \
  --tick-interval 1 \
  --command-check-interval 1 \
  --log-level debug \
  --name "Dev Test Agent"
```

### Production Simulation
```bash
# Realistic agent for production-like testing
./mock-agent \
  --randomized \
  --tick-interval 10 \
  --name "Production Simulator" \
  --worktree /var/projects/main-app
```

## Agent Behavior

### State Machine
- **Idle**: Looking for work, may pick up available tasks
- **Active**: Working on tasks, making progress, occasionally asking questions
- **Waiting Feedback**: Posted a question and waiting for human response
- **Error**: Encountered an error, attempting recovery
- **Shutting Down**: Graceful shutdown in progress

### Work Simulation
1. **Task Selection**: Picks up pending tasks or starts random scenarios
2. **Todo Creation**: Generates realistic todo lists based on work type
3. **Progress Tracking**: Simulates incremental progress with metrics
4. **Commit Simulation**: Periodic commits with realistic messages
5. **Question Posting**: Context-aware questions when encountering decisions
6. **Completion**: Marks todos complete and logs success events

### Command Processing
- **Instructions**: `stop`, `continue`, `restart`, `focus on security`
- **Feedback**: Positive reinforcement, change requests, error corrections
- **Questions**: Status inquiries, progress checks, problem diagnosis

### Randomized Events
- State transitions based on probability
- Random error scenarios with recovery
- Spontaneous questions and progress updates
- Metric fluctuations and commit activity
- Scenario switching and task completion

## Scenarios

### Bug Fix - Authentication Issue
- **Duration**: ~45 minutes
- **Steps**: Reproduce → Analyze → Fix → Test → Deploy
- **Error Rate**: 20% (dependency issues, complex JWT logic)
- **Question Rate**: 15% (session invalidation, user notification)

### Feature Development - User Dashboard
- **Duration**: ~2 hours
- **Steps**: Requirements → Backend → Frontend → Testing → Documentation
- **Error Rate**: 10% (framework compatibility, performance)
- **Question Rate**: 25% (customization level, real-time updates)

### Performance Optimization - Database Queries
- **Duration**: ~90 minutes
- **Steps**: Identify → Analyze → Optimize → Benchmark → Monitor
- **Error Rate**: 15% (optimization side effects, memory usage)
- **Question Rate**: 20% (speed vs consistency, caching strategy)

### Security Audit - API Endpoint Review
- **Duration**: ~3 hours
- **Steps**: Catalog → Validate → Analyze → Test → Fix → Document
- **Error Rate**: 30% (false positives, compatibility breaks)
- **Question Rate**: 35% (breaking changes, compatibility trade-offs)

## API Integration

The mock agent fully integrates with the Agent Fleet backend API:

### Registration
```
POST /v1/agents
{
  "name": "Clever Coder",
  "worktree": "/tmp/mock-project"
}
```

### Status Updates
```
PATCH /v1/agents/{id}
{
  "status": "active",
  "current_task": "Implementing user authentication",
  "progress": 45,
  "files_changed": 8,
  "lines_added": 234,
  "lines_removed": 67
}
```

### Event Logging
```
POST /v1/agents/{id}/events
{
  "type": "commit",
  "message": "Add user session management",
  "metadata": {
    "files_changed": 3,
    "lines_added": 89
  }
}
```

### Todo Management
```
POST /v1/agents/{id}/todos
{
  "text": "Implement password validation logic",
  "order": 3
}
```

### Command Processing
```
GET /v1/agents/{id}/commands?status=sent
PATCH /v1/agents/{id}/commands/{cmd_id}
{
  "status": "completed",
  "response": "Acknowledged. Focusing on security aspects."
}
```

## Development

### Adding New Scenarios
1. Edit `scenarios/scenarios.go`
2. Add new scenario to `predefinedScenarios` slice
3. Define steps, error conditions, and questions
4. Set appropriate duration and probabilities

### Customizing Agent Behavior
- Modify state transition logic in `agent/randomized.go`
- Adjust probabilities in `handleRandomizedBehavior()`
- Add new command types in `agent/commands.go`
- Extend metrics tracking in `agent/status.go`

### Testing Integration
- Start backend server: `cd ../backend && ./agent-fleet-backend`
- Run multiple agents: `./mock-agent --randomized --name "Agent-$RANDOM"`
- Monitor via web UI: `http://localhost:8080`
- Watch real-time updates via SSE: `curl -H "Authorization: Bearer fleet-agent-token-123" http://localhost:8080/v1/stream`

## Monitoring

The agent logs comprehensive information about its activities:

```
2025-05-28T15:30:45Z INF Agent registered successfully agent=abc-123 name="Clever Coder"
2025-05-28T15:30:50Z INF Started scenario: Bug Fix - Authentication Issue agent=abc-123 scenario="Bug Fix - Authentication Issue"
2025-05-28T15:31:15Z INF Agent asking question agent=abc-123 question="Should we invalidate all existing user sessions as part of this fix?"
2025-05-28T15:31:20Z INF Processing command agent=abc-123 command=cmd-456 type=feedback content="Yes, invalidate sessions for security"
2025-05-28T15:32:45Z INF Making commit agent=abc-123 message="Fix authentication session validation"
2025-05-28T15:35:30Z INF Completing scenario agent=abc-123 scenario="Bug Fix - Authentication Issue"
```

This mock agent provides comprehensive simulation of realistic AI coding agent behavior, perfect for testing, demonstration, and development of agent fleet management systems.
