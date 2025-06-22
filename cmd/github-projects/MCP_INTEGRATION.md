# GitHub GraphQL CLI with Embedded MCP Task Management

## Overview

The GitHub GraphQL CLI has been enhanced with embedded MCP (Model Context Protocol) server capabilities, providing task management functionality alongside the existing GitHub operations. This integration allows you to use a single binary for both GitHub project management and task tracking.

## Features

- **Embedded MCP Server**: Built-in MCP server with task management tools
- **Session Isolation**: Each MCP session maintains separate task storage
- **Multiple Transports**: Support for stdio and SSE transports
- **Existing Functionality Preserved**: All GitHub GraphQL CLI commands still work
- **Integration Benefits**: Coordinate GitHub operations with task management in one tool

## Available MCP Tools

1. **`read_tasks`** - Get all current tasks for the agent session
2. **`write_tasks`** - Replace all tasks with provided tasks (bulk operation)
3. **`add_task`** - Add a single new task to track work items
4. **`update_task`** - Update a specific task's status, priority, or content
5. **`remove_task`** - Remove a specific task by ID

## Usage

### Starting the MCP Server

#### Stdio Transport (Default)
```bash
github-graphql-cli mcp start
```

#### SSE Transport
```bash
github-graphql-cli mcp start --transport sse --port 3001
```

### List Available Tools
```bash
github-graphql-cli mcp list-tools
```

### Example MCP Tool Calls

When connected via MCP protocol, you can use these tools:

#### Add a Task
```json
{
  "name": "add_task",
  "arguments": {
    "content": "Review GitHub project setup",
    "priority": "high"
  }
}
```

#### Read All Tasks
```json
{
  "name": "read_tasks",
  "arguments": {}
}
```

#### Update a Task
```json
{
  "name": "update_task",
  "arguments": {
    "id": "task_1734739234567890123",
    "status": "in-progress",
    "priority": "medium"
  }
}
```

#### Remove a Task
```json
{
  "name": "remove_task",
  "arguments": {
    "id": "task_1734739234567890123"
  }
}
```

#### Bulk Replace Tasks
```json
{
  "name": "write_tasks",
  "arguments": {
    "tasks_json": "[{\"id\":\"task1\",\"content\":\"Setup GitHub project\",\"status\":\"todo\",\"priority\":\"high\"}]"
  }
}
}
```

## Task Structure

Each task has the following structure:

```json
{
  "id": "task_1734739234567890123",
  "content": "Description of the task",
  "status": "todo",  // "todo", "in-progress", "completed"
  "priority": "medium",  // "low", "medium", "high"  
  "created": "2024-12-20T19:20:34.567890123Z",
  "updated": "2024-12-20T19:20:34.567890123Z"
}
```

## Integration Benefits

### Single Binary Solution
- No need to manage separate tools for GitHub operations and task management
- Consistent CLI interface for all operations
- Simplified deployment and distribution

### Session-Based Task Management
- Each MCP session maintains isolated task storage
- Multiple agents can work simultaneously without interference
- Tasks persist for the duration of the MCP session

### Coordinated Workflows
Example workflow combining GitHub operations with task management:

1. **Start MCP server**: `github-graphql-cli mcp start`
2. **Agent adds task**: "Review GitHub project structure"
3. **Agent lists projects**: Uses existing `list-projects` functionality
4. **Agent updates task**: Changes status to "in-progress"
5. **Agent creates issues**: Uses `create-issue` command
6. **Agent completes task**: Updates status to "completed"

### Standard MCP Protocol
- Compatible with any MCP-compliant client or LLM agent
- Standard protocol ensures interoperability
- Built-in session management and error handling

## Example Agent Interaction

Here's how an LLM agent might use the combined functionality:

```
Agent: I need to set up a GitHub project and track my progress.

1. Add task: "Set up GitHub project structure"
2. List existing GitHub projects to understand current setup
3. Create new project if needed
4. Update task status to "in-progress"
5. Set up project fields and items
6. Update task status to "completed"
7. Add new task: "Configure project automation"
```

## Verification

### Test that both functionalities work:

```bash
# Test GitHub functionality
github-graphql-cli viewer
github-graphql-cli list-projects --help

# Test MCP functionality  
github-graphql-cli mcp list-tools
github-graphql-cli mcp start --transport sse --port 3001
```

### Test MCP Protocol Communication

Use any MCP-compatible client to connect to:
- **Stdio**: `github-graphql-cli mcp start`
- **SSE**: `http://localhost:3001` when running with `--transport sse --port 3001`

## Development Notes

- Task storage is in-memory and session-scoped
- Session ID is used to isolate tasks between different MCP connections
- All validation and error handling follows MCP protocol standards
- The implementation uses the `embeddable` package from go-go-mcp for easy integration

## Future Enhancements

Potential improvements for future versions:
- Persistent task storage (file or database)
- Integration with GitHub Issues (sync tasks with issues)  
- Cross-session task sharing options
- Task templates and workflows
- Integration with GitHub Projects v2 items
