# Amp Agent Coordination

A comprehensive SQLite-based task management system for coding agents with hierarchical task planning, DAG dependencies, project management, and agent typing.

## Features

- **Project Management**: Organize work with projects containing guidelines and context
- **Hierarchical Tasks**: Tasks can have parent-child relationships within projects
- **DAG Dependencies**: Tasks can depend on other tasks, preventing circular dependencies
- **Agent Types**: Categorize agents by role (Code Reviewer, Test Runner, etc.)
- **Smart Assignment**: Assign tasks to agent types or specific agents
- **Status Tracking**: Tasks progress through pending → in_progress → completed/failed states
- **Available Task Discovery**: Automatically finds tasks ready to execute (dependencies met)
- **Project Context**: All dual-mode outputs show current project and guidelines
- **Multiple Output Formats**: Table, JSON, YAML, CSV support throughout

## Database Schema

### Projects Table
- `id`: Unique project identifier (UUID)
- `name`: Project name
- `description`: Project description
- `guidelines`: Work guidelines for agents
- `author_id`: Optional author agent reference

### Agent Types Table
- `id`: Unique agent type identifier (UUID)
- `name`: Agent type name (e.g., "Code Reviewer")
- `description`: Agent type description
- `project_id`: Associated project

### Agents Table  
- `id`: Unique agent identifier (UUID)
- `name`: Agent name
- `status`: Agent status (idle, busy, etc.)
- `agent_type_id`: Reference to agent type (optional)

### Tasks Table
- `id`: Unique task identifier (UUID)
- `parent_id`: Reference to parent task (nullable)
- `title`: Task title
- `description`: Task description
- `status`: Task status (pending, in_progress, completed, failed)
- `agent_id`: Assigned agent UUID (nullable)
- `project_id`: Associated project
- `created_at`, `updated_at`: Timestamps

### Task Dependencies Table
- `task_id`: Task that has the dependency
- `depends_on_id`: Task that must be completed first

### Global KV Table
- `key`: Configuration key (e.g., "default_project")
- `value`: Configuration value
- `author_id`: Optional author reference

## Command Line Interface

### Quick Start
```bash
# Run demo to set up sample data
go run ./cmd/experiments/amp-agent-coordination demo

# See available work
amp-tasks tasks available

# View project context
amp-tasks projects default
```

### Project Management
```bash
# Create project with guidelines  
amp-tasks projects create "My Project" --description "Description" --guidelines "Work collaboratively"

# List all projects
amp-tasks projects list

# Set default project
amp-tasks projects set-default <project-id>

# View current default project and guidelines
amp-tasks projects default
```

### Agent Type Management
```bash
# Create agent types for project
amp-tasks agent-types create "Code Reviewer" --description "Reviews code quality"

# List agent types
amp-tasks agent-types list

# Assign task to any available agent of a type
amp-tasks agent-types assign <task-id> <agent-type-id>
```

### Agent Management
```bash
# Create agent with type
amp-tasks agents create "Review Bot" --type <agent-type-id>

# List agents and workload
amp-tasks agents list
amp-tasks agents workload
amp-tasks agents stats
```

### Task Management
```bash
# Create tasks (automatically uses default project)
amp-tasks tasks create "Build feature" --description "Detailed description"

# Create subtasks
amp-tasks tasks create "Unit tests" --parent <parent-task-id>

# List tasks with project context
amp-tasks tasks list
amp-tasks tasks list --status pending --agent <agent-id>

# See tasks ready for work
amp-tasks tasks available

# Assign and update tasks
amp-tasks tasks assign <task-id> <agent-id>
amp-tasks tasks status <task-id> completed  # Shows new available tasks

# View task details
amp-tasks tasks show <task-id>
```

### Dependency Management
```bash
# Add dependencies
amp-tasks deps add <task-id> <depends-on-task-id>

# View dependencies
amp-tasks deps list <task-id>

# Visualize dependency graph
amp-tasks deps graph
amp-tasks deps graph --format dot  # Graphviz format
```

### Output Formats
```bash
# All commands support multiple output formats
amp-tasks tasks list --output json
amp-tasks agents list --output csv
amp-tasks projects list --output yaml
```

## Agent Work Flow

1. **Check project context**: `amp-tasks projects default`
2. **Find available work**: `amp-tasks tasks available`
3. **Assign work**: `amp-tasks tasks assign <task-id> <agent-id>`
4. **Update progress**: `amp-tasks tasks status <task-id> completed`
5. **See new available tasks** (shown automatically)

## Key Improvements

- **Project Context**: Every dual-mode command shows project name and guidelines
- **Agent Types**: Type-based assignment for flexible workforce management  
- **Status Completion Flow**: Completing tasks automatically shows newly available work
- **Default Project Logic**: Uses latest project or explicitly set default
- **Comprehensive CLI**: Full CRUD operations with consistent patterns
- **Multiple Output Formats**: JSON/YAML/CSV for programmatic integration

## Built-in Documentation

All documentation is embedded in the CLI for easy agent access:

```bash
amp-tasks docs quick-start    # Essential getting started commands
amp-tasks docs agent-guide    # Concise work reference for agents  
amp-tasks docs workflow       # Detailed agent workflow
amp-tasks docs commands       # Complete command reference
amp-tasks docs readme         # Full system documentation
```

Use `--raw` flag for markdown output suitable for external processing.
