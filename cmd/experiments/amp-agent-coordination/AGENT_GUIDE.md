# Agent Work Guide

Quick reference for agents working with the task coordination system.

## Essential Commands

### Finding Work
```bash
# See what's ready to work on
amp-tasks tasks available

# See your assigned tasks 
amp-tasks tasks list --agent <your-agent-id>

# See all tasks in project
amp-tasks tasks list
```

### Taking Work
```bash
# Assign a specific task to yourself
amp-tasks tasks assign <task-id> <your-agent-id>

# Or get assigned by agent type (if configured)
amp-tasks agent-types assign <task-id> <agent-type-id>
```

### Updating Progress
```bash
# Mark task as in progress (automatic when assigned)
amp-tasks tasks status <task-id> in_progress

# Mark task completed (shows new available tasks)
amp-tasks tasks status <task-id> completed

# Mark task failed (if blocked/can't complete)
amp-tasks tasks status <task-id> failed
```

### Understanding Context
```bash
# See project guidelines
amp-tasks projects default

# See task details and dependencies
amp-tasks tasks show <task-id>

# See dependency graph
amp-tasks deps graph
```

## Work Flow

1. **Check available tasks** - `tasks available`
2. **Understand project context** - `projects default` 
3. **Assign work to yourself** - `tasks assign <task-id> <agent-id>`
4. **Complete the work** following project guidelines
5. **Mark completed** - `tasks status <task-id> completed`
6. **Check new available tasks** (shown automatically)

## Key Principles

- **Follow dependencies** - Only available tasks have all dependencies met
- **Read project guidelines** - Each project has specific work guidance
- **Communicate progress** - Update status as you work 
- **Check task details** - Use `tasks show` to understand requirements
- **Work collaboratively** - Coordinate with other agents via status updates

## Output Formats

Add `--output json` to any command for programmatic use:
```bash
amp-tasks tasks available --output json
amp-tasks tasks show <id> --output json
```

## Quick Status Check

```bash
# My current work
amp-tasks tasks list --agent <my-id> --status in_progress

# What's ready next
amp-tasks tasks available

# Project overview
amp-tasks projects default
```

## Creating Work (Advanced)

```bash
# Create new task
amp-tasks tasks create "Task title" --description "Details"

# Add dependencies 
amp-tasks deps add <task-id> <depends-on-id>

# Create subtasks
amp-tasks tasks create "Subtask" --parent <parent-task-id>
```
