# Agent Fleet Backend Server

A backend server for managing and monitoring AI coding agent fleets with REST API and real-time updates.

## Features

- **REST API**: Complete CRUD operations for agents, tasks, todos, commands, and events
- **Real-time Updates**: Server-Sent Events (SSE) for live fleet monitoring
- **SQLite Storage**: Lightweight, embedded database with automatic migrations
- **Web Interface**: Bootstrap-based web UI for fleet monitoring
- **Authentication**: Bearer token authentication for API endpoints

## Quick Start

### Build and Run

```bash
# Build the application
go build .

# Run with default settings
./backend

# Run with custom configuration
./backend --port 9000 --database ./my-fleet.db --log-level debug
```

### Configuration Options

```bash
Usage:
  agent-fleet-backend [flags]

Flags:
      --config string      config file (default is $HOME/.agent-fleet.yaml)
  -d, --database string    SQLite database file path (default "./agent-fleet.db")
      --dev                Development mode
  -h, --help               help for agent-fleet-backend
  -H, --host string        Host to bind the server to (default "localhost")
  -l, --log-level string   Log level (trace, debug, info, warn, error) (default "info")
  -p, --port string        Port to run the server on (default "8080")
```

## API Usage

### Authentication

All API endpoints require Bearer token authentication:

```bash
curl -H "Authorization: Bearer fleet-agent-token-123" \
     http://localhost:8080/v1/agents
```

### Create an Agent

```bash
curl -X POST http://localhost:8080/v1/agents \
  -H "Authorization: Bearer fleet-agent-token-123" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Code Reviewer",
    "worktree": "/path/to/project"
  }'
```

### Update Agent Status

```bash
curl -X PATCH http://localhost:8080/v1/agents/{agent-id} \
  -H "Authorization: Bearer fleet-agent-token-123" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "active",
    "current_task": "Reviewing pull request #123",
    "progress": 45
  }'
```

### Add Todo Item

```bash
curl -X POST http://localhost:8080/v1/agents/{agent-id}/todos \
  -H "Authorization: Bearer fleet-agent-token-123" \
  -H "Content-Type: application/json" \
  -d '{
    "text": "Review code changes in auth module",
    "order": 1
  }'
```

### Send Command to Agent

```bash
curl -X POST http://localhost:8080/v1/agents/{agent-id}/commands \
  -H "Authorization: Bearer fleet-agent-token-123" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Please focus on security issues in the review",
    "type": "feedback"
  }'
```

### Create Task

```bash
curl -X POST http://localhost:8080/v1/tasks \
  -H "Authorization: Bearer fleet-agent-token-123" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Implement user authentication",
    "description": "Add JWT-based authentication to the API",
    "priority": "high"
  }'
```

## Real-time Updates

### Server-Sent Events

Connect to the SSE endpoint for real-time updates:

```javascript
const eventSource = new EventSource('/v1/stream', {
  headers: {
    'Authorization': 'Bearer fleet-agent-token-123'
  }
});

eventSource.onmessage = function(event) {
  const data = JSON.parse(event.data);
  console.log('Update:', data);
};
```

### Event Types

- `agent_status_changed`: Agent status updated
- `agent_event_created`: New agent activity logged
- `agent_question_posted`: Agent needs feedback
- `agent_progress_updated`: Agent progress metrics updated
- `todo_updated`: Todo item created/updated/deleted
- `task_assigned`: Task assigned to agent
- `command_received`: Command sent to agent

## Web Interface

Access the web interface at:
- **Dashboard**: http://localhost:8080/
- **Agents**: http://localhost:8080/agents

The web interface provides:
- Fleet status overview
- Real-time agent monitoring
- Recent activity feed
- Agent detail cards with progress tracking

## Database Schema

The application automatically creates the following tables:
- `agents`: Agent information and status
- `events`: Agent activity log
- `todo_items`: Agent todo lists
- `tasks`: Task queue
- `commands`: Commands sent to agents

## Development

### Adding New Features

1. Update models in `models/models.go`
2. Add database operations in `database/`
3. Create API handlers in `handlers/`
4. Update web templates in `templates/` (run `templ generate`)
5. Add routes in `main.go`

### API Status Codes

- `200 OK`: Successful GET/PATCH request
- `201 Created`: Successful POST request
- `204 No Content`: Successful DELETE request
- `400 Bad Request`: Invalid request data
- `401 Unauthorized`: Missing/invalid authentication
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Security Note

The current implementation uses a hardcoded bearer token for simplicity. In production, implement proper JWT validation or database-backed authentication.
