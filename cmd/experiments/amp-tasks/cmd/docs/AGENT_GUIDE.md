# Agent Work Guide

Quick reference for agents working with the task coordination system.

## ⚠️ IMPORTANT: Agent Identity Requirements

**Before starting any work, every agent MUST:**
1. **Be properly defined** in the system with a unique agent entry
2. **Have an assigned agent-type** that exists in the current project
3. **Verify identity** before taking on tasks or making assignments

**For LLM Agents:** You can switch personas while working (e.g., coordinator ↔ worker, or between different specializations), but **every persona must be properly defined as a separate agent with appropriate agent-type**.

## Agent Self-Identification

### FIRST: Check for Existing Agents
```bash
# ALWAYS check if you already exist before creating new agent
amp-tasks agents list | grep -i "your name"
amp-tasks agents list | grep -i "coordinator\|developer\|reviewer"

# Search by partial name or role
amp-tasks agents list --output json | jq '.[] | select(.name | test("(?i)alice|coordinator"))'

# Check all existing agents to avoid duplicates
amp-tasks agents list
```

### Creating Your Agent Entry (Only If Needed)
```bash
# Only create if you don't already exist
amp-tasks agents create "Your Name" --agent-type-id <agent-type-slug>

# Verify your agent was created successfully
amp-tasks agents list | grep "Your Name"
```

### FIRST: Check Existing Agent Types
```bash
# ALWAYS check existing agent types before creating new ones
amp-tasks agent-types list

# Search for similar agent types by keyword
amp-tasks agent-types list --output json | jq '.[] | select(.name | test("(?i)developer|frontend|backend"))'
amp-tasks agent-types list --output json | jq '.[] | select(.description | test("(?i)review|test|coordinate"))'

# Look for agent types that might fit your role
amp-tasks agent-types list | grep -i "frontend\|backend\|developer\|reviewer\|coordinator"
```

### Finding Your Agent Type (Use Existing When Possible)
```bash
# See what agent types are available in current project
amp-tasks agent-types list

# Search for agent types that match your intended role
amp-tasks agent-types list --output json | jq '.[] | select(.description | contains("keyword"))'

# Find your agent record and its type
amp-tasks agents list --output json | jq '.[] | select(.name | contains("Your Name"))'
```

### Understanding Agent Types in Projects
```bash
# See agent types for current project (includes global types)
amp-tasks agent-types list

# See only global agent types (available across all projects)
amp-tasks agent-types list --output json | jq '.[] | select(.global == true)'

# Check which project you're working in
amp-tasks projects default
```

## LLM Agent Persona Switching

### When to Switch Personas
- **Coordination tasks**: Switch to "Project Coordinator" or "Tech Lead" persona
- **Specialized work**: Switch to "Frontend Developer", "Backend Developer", etc.
- **Review tasks**: Switch to "Code Reviewer" or "QA Tester" persona
- **Different project phases**: Switch from "Developer" to "Documenter"

### How to Switch Personas
```bash
# 1. Check if the target persona already exists
amp-tasks agents list | grep -i "coordinator\|tech.*lead"
amp-tasks agents list | grep "Your Name"

# 2. Check if appropriate agent-type exists in current project
amp-tasks agent-types list | grep -i "tech.*lead\|coordinator\|manager"

# 3. Only create persona if it doesn't exist and agent-type is available
amp-tasks agents create "Alice Coordinator" --agent-type-id <existing-tech-lead-type>

# 4. Verify the persona was created successfully
amp-tasks agents list | grep "Alice Coordinator"

# 5. Switch context - now operate as the new persona
# All subsequent task assignments and work should use this agent ID
```

### Persona Best Practices
- **Check First**: Always verify existing agents/types before creating new ones
- **Reuse When Possible**: Use existing agent-types that match your needs
- **Distinct Names**: "Alice Developer", "Alice Coordinator", "Alice Reviewer"
- **Appropriate Types**: Match persona to actual agent-type capabilities
- **Clear Context**: Document when and why you're switching personas
- **Consistent Identity**: Use the same persona for related work streams
- **Avoid Duplicates**: Don't create "Frontend Dev" if "Frontend Developer" exists

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
- **Respect project scoping** - Agent types and tasks are project-specific unless marked global
- **Know your capabilities** - Understand which agent type you belong to and what work you're qualified for

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
