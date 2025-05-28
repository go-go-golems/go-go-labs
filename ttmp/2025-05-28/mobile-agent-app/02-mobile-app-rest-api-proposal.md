# Agent Fleet Management API Specification

## Base URL
```
https://api.agentfleet.dev/v1
```

## Authentication
All endpoints require Bearer token authentication:
```
Authorization: Bearer <token>
```

---

## Data Models

### Agent
```json
{
  "id": "string",
  "name": "string",
  "status": "active|idle|waiting_feedback|error",
  "current_task": "string",
  "worktree": "string",
  "files_changed": "integer",
  "lines_added": "integer", 
  "lines_removed": "integer",
  "last_commit": "string (ISO 8601)",
  "progress": "integer (0-100)",
  "pending_question": "string|null",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)"
}
```

### Event
```json
{
  "id": "string",
  "agent_id": "string",
  "type": "start|commit|question|success|error|info|command",
  "message": "string",
  "metadata": "object",
  "timestamp": "string (ISO 8601)"
}
```

### TodoItem
```json
{
  "id": "string",
  "agent_id": "string", 
  "text": "string",
  "completed": "boolean",
  "current": "boolean",
  "order": "integer",
  "created_at": "string (ISO 8601)",
  "completed_at": "string (ISO 8601)|null"
}
```

### Task
```json
{
  "id": "string",
  "title": "string",
  "description": "string",
  "assigned_agent_id": "string|null",
  "status": "pending|assigned|in_progress|completed|failed",
  "priority": "low|medium|high|urgent",
  "created_at": "string (ISO 8601)",
  "assigned_at": "string (ISO 8601)|null",
  "completed_at": "string (ISO 8601)|null"
}
```

### Command
```json
{
  "id": "string",
  "agent_id": "string",
  "content": "string",
  "type": "instruction|feedback|question",
  "response": "string|null",
  "status": "sent|acknowledged|completed",
  "sent_at": "string (ISO 8601)",
  "responded_at": "string (ISO 8601)|null"
}
```

---

## REST Endpoints

### Agents

#### `GET /agents`
List all agents with optional filtering
- **Query Parameters:**
  - `status` - Filter by status
  - `limit` - Number of results (default: 50, max: 100)
  - `offset` - Pagination offset
- **Response:** `200 OK`
```json
{
  "agents": [Agent],
  "total": "integer",
  "limit": "integer", 
  "offset": "integer"
}
```

#### `GET /agents/{agent_id}`
Get detailed agent information
- **Response:** `200 OK` - Agent object
- **Error:** `404 Not Found`

#### `POST /agents`
Create a new agent
- **Body:**
```json
{
  "name": "string",
  "worktree": "string"
}
```
- **Response:** `201 Created` - Agent object

#### `PATCH /agents/{agent_id}`
Update agent information
- **Body:** Partial Agent object
- **Response:** `200 OK` - Updated Agent object

#### `DELETE /agents/{agent_id}`
Remove an agent
- **Response:** `204 No Content`

### Agent Events

#### `GET /agents/{agent_id}/events`
Get event history for an agent
- **Query Parameters:**
  - `type` - Filter by event type
  - `since` - ISO 8601 timestamp
  - `limit` - Number of results (default: 100)
  - `offset` - Pagination offset
- **Response:** `200 OK`
```json
{
  "events": [Event],
  "total": "integer"
}
```

#### `POST /agents/{agent_id}/events`
Create a new event (typically used by agents)
- **Body:**
```json
{
  "type": "string",
  "message": "string", 
  "metadata": "object"
}
```
- **Response:** `201 Created` - Event object

### Agent Todo Lists

#### `GET /agents/{agent_id}/todos`
Get todo list for an agent
- **Response:** `200 OK`
```json
{
  "todos": [TodoItem]
}
```

#### `POST /agents/{agent_id}/todos`
Add todo item
- **Body:**
```json
{
  "text": "string",
  "order": "integer"
}
```
- **Response:** `201 Created` - TodoItem object

#### `PATCH /agents/{agent_id}/todos/{todo_id}`
Update todo item
- **Body:**
```json
{
  "completed": "boolean",
  "current": "boolean",
  "text": "string"
}
```
- **Response:** `200 OK` - Updated TodoItem object

#### `DELETE /agents/{agent_id}/todos/{todo_id}`
Remove todo item
- **Response:** `204 No Content`

### Commands

#### `GET /agents/{agent_id}/commands`
Get command history for an agent
- **Query Parameters:**
  - `status` - Filter by command status
  - `limit` - Number of results (default: 50)
- **Response:** `200 OK`
```json
{
  "commands": [Command]
}
```

#### `POST /agents/{agent_id}/commands`
Send command to an agent
- **Body:**
```json
{
  "content": "string",
  "type": "instruction|feedback|question"
}
```
- **Response:** `201 Created` - Command object

#### `PATCH /agents/{agent_id}/commands/{command_id}`
Update command (typically agent responding)
- **Body:**
```json
{
  "response": "string",
  "status": "acknowledged|completed"
}
```
- **Response:** `200 OK` - Updated Command object

### Tasks

#### `GET /tasks`
List all tasks
- **Query Parameters:**
  - `status` - Filter by task status
  - `assigned_agent_id` - Filter by assigned agent
  - `priority` - Filter by priority
  - `limit` - Number of results (default: 50)
  - `offset` - Pagination offset
- **Response:** `200 OK`
```json
{
  "tasks": [Task],
  "total": "integer"
}
```

#### `POST /tasks`
Create a new task
- **Body:**
```json
{
  "title": "string",
  "description": "string",
  "priority": "low|medium|high|urgent",
  "assigned_agent_id": "string|null"
}
```
- **Response:** `201 Created` - Task object

#### `GET /tasks/{task_id}`
Get task details
- **Response:** `200 OK` - Task object

#### `PATCH /tasks/{task_id}`
Update task
- **Body:** Partial Task object
- **Response:** `200 OK` - Updated Task object

#### `DELETE /tasks/{task_id}`
Cancel/remove task
- **Response:** `204 No Content`

### Fleet Operations

#### `GET /fleet/status`
Get overall fleet status summary
- **Response:** `200 OK`
```json
{
  "total_agents": "integer",
  "active_agents": "integer", 
  "pending_tasks": "integer",
  "agents_needing_feedback": "integer",
  "total_files_changed": "integer",
  "total_commits_today": "integer"
}
```

#### `GET /fleet/recent-updates`
Get recent updates across all agents
- **Query Parameters:**
  - `limit` - Number of results (default: 20)
  - `since` - ISO 8601 timestamp
- **Response:** `200 OK`
```json
{
  "updates": [Event]
}
```

---

## Server-Sent Events (SSE)

### Connection
```
GET /stream
Accept: text/event-stream
Authorization: Bearer <token>
```

### Event Types

#### `agent_status_changed`
Agent status has changed
```json
{
  "event": "agent_status_changed",
  "data": {
    "agent_id": "string",
    "old_status": "string",
    "new_status": "string",
    "agent": Agent
  }
}
```

#### `agent_event_created`
New event created for an agent
```json
{
  "event": "agent_event_created", 
  "data": {
    "agent_id": "string",
    "event": Event
  }
}
```

#### `agent_question_posted`
Agent has a question requiring feedback
```json
{
  "event": "agent_question_posted",
  "data": {
    "agent_id": "string",
    "question": "string",
    "agent": Agent
  }
}
```

#### `agent_progress_updated`
Agent progress or metrics updated
```json
{
  "event": "agent_progress_updated",
  "data": {
    "agent_id": "string", 
    "progress": "integer",
    "files_changed": "integer",
    "lines_added": "integer",
    "lines_removed": "integer"
  }
}
```

#### `todo_updated`
Todo item changed
```json
{
  "event": "todo_updated",
  "data": {
    "agent_id": "string",
    "todo": TodoItem,
    "action": "created|updated|deleted"
  }
}
```

#### `task_assigned`
Task assigned to an agent
```json
{
  "event": "task_assigned",
  "data": {
    "task": Task,
    "agent_id": "string"
  }
}
```

#### `command_received`
Command sent to agent
```json
{
  "event": "command_received",
  "data": {
    "agent_id": "string",
    "command": Command
  }
}
```

---

## Error Responses

All error responses follow this format:
```json
{
  "error": {
    "code": "string",
    "message": "string",
    "details": "object|null"
  }
}
```

### Status Codes
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflict (e.g., agent name already exists)
- `422 Unprocessable Entity` - Valid format but invalid data
- `429 Too Many Requests` - Rate limit exceeded
- `500 Internal Server Error` - Server error

---

## Rate Limits
- **General API**: 1000 requests per hour per user
- **SSE Connection**: 1 concurrent connection per user
- **Command Sending**: 60 commands per minute per agent

## Pagination
List endpoints support cursor-based pagination:
- Use `limit` and `offset` for page-based pagination
- Maximum `limit` is 100
- Include `total` count in responses when possible

## Webhooks (Optional)
For external integrations, webhooks can be configured to receive the same events as SSE:

#### `POST /webhooks`
Register a webhook endpoint
```json
{
  "url": "string",
  "events": ["string"],
  "secret": "string"
}
```