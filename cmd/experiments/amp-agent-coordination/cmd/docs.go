package cmd

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var readmeContent = `# Amp Agent Coordination

A comprehensive SQLite-based task management system for coding agents with hierarchical task planning, DAG dependencies, project management, agent typing, and knowledge sharing.

## Features

- **Project Management**: Organize work with projects containing guidelines and context
- **Hierarchical Tasks**: Tasks can have parent-child relationships within projects
- **DAG Dependencies**: Tasks can depend on other tasks, preventing circular dependencies
- **Agent Types**: Categorize agents by role (Code Reviewer, Test Runner, etc.)
- **Smart Assignment**: Assign tasks to agent types or specific agents with enhanced status display
- **Status Tracking**: Tasks progress through pending → in_progress → completed/failed states
- **Available Task Discovery**: Automatically finds tasks ready to execute (dependencies met)
- **Knowledge Sharing**: TIL (Today I Learned) system for sharing insights between agents
- **Progress Tracking**: Notes system for documenting task progress and blockers
- **Project Context**: All dual-mode outputs show current project and guidelines
- **Multiple Output Formats**: Table, JSON, YAML, CSV support throughout

## Database Schema

### Projects Table
- ` + "`id`" + `: Unique project identifier (UUID)
- ` + "`name`" + `: Project name
- ` + "`description`" + `: Project description
- ` + "`guidelines`" + `: Work guidelines for agents
- ` + "`author_id`" + `: Optional author agent reference

### Agent Types Table
- ` + "`id`" + `: Unique agent type identifier (UUID)
- ` + "`name`" + `: Agent type name (e.g., "Code Reviewer")
- ` + "`description`" + `: Agent type description
- ` + "`project_id`" + `: Associated project

### Agents Table  
- ` + "`id`" + `: Unique agent identifier (UUID)
- ` + "`name`" + `: Agent name
- ` + "`status`" + `: Agent status (idle, busy, etc.)
- ` + "`agent_type_slug`" + `: Reference to agent type slug (optional)

### Tasks Table
- ` + "`id`" + `: Unique task identifier (UUID)
- ` + "`parent_id`" + `: Reference to parent task (nullable)
- ` + "`title`" + `: Task title
- ` + "`description`" + `: Task description
- ` + "`status`" + `: Task status (pending, in_progress, completed, failed)
- ` + "`agent_id`" + `: Assigned agent UUID (nullable)
- ` + "`project_id`" + `: Associated project
- ` + "`created_at`" + `, ` + "`updated_at`" + `: Timestamps

### Task Dependencies Table
- ` + "`task_id`" + `: Task that has the dependency
- ` + "`depends_on_id`" + `: Task that must be completed first

### Global KV Table
- ` + "`key`" + `: Configuration key (e.g., "default_project")
- ` + "`value`" + `: Configuration value
- ` + "`author_id`" + `: Optional author reference

### TIL (Today I Learned) Table
- ` + "`id`" + `: Unique TIL identifier (UUID)
- ` + "`title`" + `: TIL title/topic
- ` + "`content`" + `: Learning content/insight
- ` + "`project_id`" + `: Associated project
- ` + "`task_id`" + `: Associated task (nullable)
- ` + "`author_id`" + `: Author agent
- ` + "`created_at`" + `: Creation timestamp

### Notes Table
- ` + "`id`" + `: Unique note identifier (UUID)
- ` + "`task_id`" + `: Associated task
- ` + "`content`" + `: Note content
- ` + "`author_id`" + `: Author agent
- ` + "`created_at`" + `: Creation timestamp

## Command Line Interface

### Quick Start
` + "```bash" + `
# Run demo to set up sample data
go run ./cmd/experiments/amp-agent-coordination demo

# See available work
amp-tasks tasks available

# View project context
amp-tasks projects default
` + "```" + `

### Project Management
` + "```bash" + `
# Create project with guidelines  
amp-tasks projects create "My Project" --description "Description" --guidelines "Work collaboratively"

# List all projects
amp-tasks projects list

# Set default project
amp-tasks projects set-default <project-id>

# View current default project and guidelines
amp-tasks projects default
` + "```" + `

### Agent Work Flow

1. **Check project context**: ` + "`amp-tasks projects default`" + `
2. **Find available work**: ` + "`amp-tasks tasks available`" + ` (shows agent types, assignment status)
3. **Read dependency insights**: ` + "`amp-tasks deps show <task-id>`" + ` (notes & TILs)
4. **Assign work**: ` + "`amp-tasks tasks assign <task-id> <agent-id>`" + `
5. **Track progress**: ` + "`amp-tasks notes add <task-id> 'update'`" + `
6. **Share insights**: ` + "`amp-tasks til create 'title' --content 'learning'`" + `
7. **Update status**: ` + "`amp-tasks tasks status <task-id> completed`" + `
8. **See new available tasks** (shown automatically)

## Key Improvements

- **Project Context**: Every dual-mode command shows project name and guidelines
- **Agent Types**: Type-based assignment for flexible workforce management with enhanced status display
- **Knowledge Sharing**: TIL system enables agents to learn from each other's insights
- **Progress Tracking**: Notes system provides transparency into task progress and blockers
- **Status Completion Flow**: Completing tasks automatically shows newly available work
- **Default Project Logic**: Uses latest project or explicitly set default
- **Comprehensive CLI**: Full CRUD operations with consistent patterns
- **Multiple Output Formats**: JSON/YAML/CSV for programmatic integration

## Agent Work Guide

See ` + "`amp-tasks docs agent-guide`" + ` for a concise reference focused on agent workflow and essential commands.`

var agentGuideContent = `# Agent Work Guide

Quick reference for agents working with the task coordination system.

## Understanding Agent Types

### Finding Your Place in the System
` + "```bash" + `
# See all available agent types in the system
amp-tasks agent-types list

# See all agents and their types
amp-tasks agents list

# Find specific agent types
amp-tasks agent-types list --name "Code Reviewer"
` + "```" + `

### Assignment Strategies

**By Agent Type (Recommended for flexible work distribution):**
- Use when you want any qualified agent to pick up work
- Use when the specific agent doesn't matter, just the skill set
- Use for load balancing across similar agents

**By Specific Agent (Use when you need specific expertise):**
- Use when you know exactly who should do the work
- Use for specialized knowledge or context
- Use for follow-up work on tasks they've done before

` + "```bash" + `
# Assign to any agent of a specific type (flexible)
amp-tasks agent-types assign <task-id> <agent-type-id>

# Assign to a specific agent (targeted)  
amp-tasks tasks assign <task-id> <agent-id>
` + "```" + `

### When to Use Each Approach

- **Agent Type Assignment**: "Any code reviewer can handle this task"
- **Specific Agent Assignment**: "Alice needs to review this since she wrote the original code"

### Practical Examples

` + "```bash" + `
# Example: Finding your agent type
amp-tasks agents list | grep "your-agent-name"

# Example: See what types exist in the system
amp-tasks agent-types list

# Example: Assign work flexibly
amp-tasks agent-types assign task-123 code-reviewer-type-id

# Example: Assign work specifically  
amp-tasks tasks assign task-456 alice-agent-id
` + "```" + `

## Essential Commands

### Finding Work
` + "```bash" + `
# See what's ready to work on
amp-tasks tasks available

# See your assigned tasks 
amp-tasks tasks list --agent <your-agent-id>

# See all tasks in project
amp-tasks tasks list
` + "```" + `

### Taking Work
` + "```bash" + `
# Assign a specific task to yourself
amp-tasks tasks assign <task-id> <your-agent-id>

# Or get assigned by agent type (if configured)
amp-tasks agent-types assign <task-id> <agent-type-id>
` + "```" + `

### Updating Progress
` + "```bash" + `
# Mark task as in progress (automatic when assigned)
amp-tasks tasks status <task-id> in_progress

# Mark task completed (shows new available tasks)
amp-tasks tasks status <task-id> completed

# Mark task failed (if blocked/can't complete)
amp-tasks tasks status <task-id> failed
` + "```" + `

### Understanding Context
` + "```bash" + `
# See project guidelines
amp-tasks projects default

# See task details and dependencies
amp-tasks tasks show <task-id>

# See dependency graph
amp-tasks deps graph

# Read notes from previous work on dependencies
amp-tasks deps show <task-id>  # Shows parent task notes
` + "```" + `

### Knowledge Sharing

#### TIL (Today I Learned) - For Insights
` + "```bash" + `
# Share project-level insights
amp-tasks til create "Docker optimization" --content "Multi-stage builds reduce image size by 60%"

# Share task-specific learnings
amp-tasks til create "Error handling pattern" --content "Use errors.Wrap for context" --task <task-id>

# List insights from project
amp-tasks til list

# View specific insight
amp-tasks til show <til-id>
` + "```" + `

#### Notes - For Progress Tracking
` + "```bash" + `
# Add progress notes during work
amp-tasks notes add <task-id> "Implemented authentication middleware"
amp-tasks notes add <task-id> "Found issue with database connection pooling"

# Read notes from task
amp-tasks notes list <task-id>

# Read notes from specific agent
amp-tasks notes list <task-id> --agent <agent-id>
` + "```" + `

## Work Flow

1. **Check available agent types** - ` + "`agent-types list`" + ` (understand the system)
2. **See available agents** - ` + "`agents list`" + ` (understand who can work on what)
3. **Check available tasks** - ` + "`tasks available`" + ` (find work ready to do)
4. **Understand project context** - ` + "`projects default`" + ` (read guidelines)
5. **Read insights from dependencies** - ` + "`deps show <task-id>`" + ` (if any)
6. **Assign work strategically**:
   - Type-based: ` + "`agent-types assign <task-id> <type-id>`" + ` (flexible)
   - Specific: ` + "`tasks assign <task-id> <agent-id>`" + ` (targeted)
7. **Take notes during work** - ` + "`notes add <task-id> 'progress update'`" + `
8. **Complete the work** following project guidelines
9. **Share insights** - ` + "`til create 'learning title' --content 'what you learned'`" + `
10. **Mark completed** - ` + "`tasks status <task-id> completed`" + `
11. **Check new available tasks** (shown automatically)

## Key Principles

- **Understand the agent type system** - Know what types exist and how to use them
- **Choose assignment strategy wisely** - Type-based for flexibility, specific for targeted work
- **Follow dependencies** - Only available tasks have all dependencies met
- **Read project guidelines** - Each project has specific work guidance
- **Document progress** - Take notes as you work for transparency
- **Share learnings** - Create TIL entries for insights that help others
- **Learn from others** - Read notes and TILs from dependency tasks
- **Use TIL vs Notes wisely**:
  - **TIL**: Insights, best practices, lessons learned (shareable knowledge)
  - **Notes**: Progress updates, blockers, implementation details (task-specific)
- **Work collaboratively** - Coordinate with other agents via status updates

## Output Formats

Add ` + "`--output json`" + ` to any command for programmatic use:
` + "```bash" + `
amp-tasks tasks available --output json
amp-tasks tasks show <id> --output json
` + "```" + `

## Quick Status Check

` + "```bash" + `
# My current work
amp-tasks tasks list --agent <my-id> --status in_progress

# What's ready next
amp-tasks tasks available

# Project overview
amp-tasks projects default
` + "```" + `

## Creating Work (Advanced)

` + "```bash" + `
# Create new task
amp-tasks tasks create "Task title" --description "Details"

# Add dependencies 
amp-tasks deps add <task-id> <depends-on-id>

# Create subtasks
amp-tasks tasks create "Subtask" --parent <parent-task-id>
` + "```" + ``

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Access embedded documentation",
	Long:  "View README, agent guide, and other documentation directly from the CLI",
}

var readmeCmd = &cobra.Command{
	Use:   "readme",
	Short: "Show the complete README documentation",
	Long:  "Display the full README with system overview, features, and usage examples",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("README", readmeContent, raw)
	},
}

var agentGuideCmd = &cobra.Command{
	Use:   "agent-guide",
	Short: "Show the agent work guide",
	Long:  "Display the concise agent work guide with essential commands and workflow",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Agent Guide", agentGuideContent, raw)
	},
}

var quickStartCmd = &cobra.Command{
	Use:   "quick-start",
	Short: "Show quick start guide",
	Long:  "Display essential commands to get started with the system",
	RunE: func(cmd *cobra.Command, args []string) error {
		quickStart := `# Quick Start Guide

## 1. Set up sample data
amp-tasks demo

## 2. Understand the agent system
amp-tasks agent-types list
amp-tasks agents list

## 3. Check project context
amp-tasks projects default

## 4. See available work
amp-tasks tasks available

## 5. Read insights from dependencies (if any)
amp-tasks deps show <task-id>

## 6. Assign work strategically
# By agent type (flexible - any qualified agent can pick it up)
amp-tasks agent-types assign <task-id> <agent-type-id>

# By specific agent (targeted - specific expertise needed)
amp-tasks tasks assign <task-id> <agent-id>

## 7. Track progress with notes
amp-tasks notes add <task-id> "Working on authentication"

## 8. Share insights
amp-tasks til create "Auth Pattern" --content "Use JWT for stateless auth"

## 9. Update status
amp-tasks tasks status <task-id> completed

## 10. Create new work
amp-tasks tasks create "New task" --description "Details"

## 11. Add dependencies
amp-tasks deps add <task-id> <depends-on-id>

## 12. Visualize work
amp-tasks deps graph

For detailed help: amp-tasks docs agent-guide
For complete docs: amp-tasks docs readme
`
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Quick Start", quickStart, raw)
	},
}

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Show typical agent workflow",
	Long:  "Display the step-by-step workflow for agents",
	RunE: func(cmd *cobra.Command, args []string) error {
		workflow := `# Agent Workflow

## Daily Work Cycle

1. **Check Context**
   amp-tasks projects default
   - Read project guidelines
   - Understand current objectives

2. **Understand the Agent System**
   amp-tasks agent-types list
   amp-tasks agents list
   - See what agent types exist
   - Understand the workforce structure

3. **Find Available Work**
   amp-tasks tasks available
   - See tasks ready for assignment
   - Check dependencies are met

4. **Read Previous Work (if dependencies exist)**
   amp-tasks deps show <task-id>
   - Review notes from agents who worked on dependency tasks
   - Learn from TIL entries related to this work

5. **Assign Work Strategically**
   # By agent type (flexible distribution)
   amp-tasks agent-types assign <task-id> <type-id>
   
   # By specific agent (targeted assignment)
   amp-tasks tasks assign <task-id> <your-agent-id>
   - Task status automatically becomes 'in_progress'

6. **Do the Work**
   - Follow project guidelines
   - Complete the task requirements  
   - Take progress notes: amp-tasks notes add <task-id> "progress update"
   - Check task details: amp-tasks tasks show <task-id>

7. **Share Learning**
   amp-tasks til create "insight title" --content "what you learned"
   - Create task-specific TIL: --task <task-id>
   - Share insights that help other agents

8. **Update Status**
   amp-tasks tasks status <task-id> completed
   - System shows newly available tasks
   - Dependencies are automatically resolved

9. **Create Additional Work (if needed)**
   amp-tasks tasks create "New task" --description "Details"
   amp-tasks deps add <new-task> <depends-on-task>

## Key Principles

- Always check project guidelines first
- Understand the agent type system before assigning work
- Choose assignment strategy based on work requirements
- Only work on available tasks (dependencies met)
- Read notes from dependency tasks to understand context
- Document your progress with notes for transparency
- Share valuable insights through TIL entries
- Update status promptly
- Create clear, actionable tasks
- Follow the dependency chain

## Getting Help

- amp-tasks docs agent-guide    # Essential commands
- amp-tasks docs readme         # Complete documentation  
- amp-tasks --help              # Command reference
- amp-tasks <command> --help    # Specific command help
`
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Agent Workflow", workflow, raw)
	},
}

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Complete project initialization guide",
	Long:  "Step-by-step guide to initialize a new project with proper structure",
	RunE: func(cmd *cobra.Command, args []string) error {
		setup := `# Project Setup Guide

Complete initialization workflow for setting up a new project with proper structure, agents, and task hierarchy.

## 1. Project Creation

### Create Your Project
` + "```bash" + `
# Create a new project with guidelines
amp-tasks projects create "Web Dashboard" \
  --description "Customer-facing analytics dashboard" \
  --guidelines "Use TypeScript, follow React patterns, write tests for all components, prioritize accessibility"

# View the created project
amp-tasks projects list
` + "```" + `

### Set as Default Project
` + "```bash" + `
# Set as your default project (replace with actual project ID)
amp-tasks projects set-default <project-id>

# Verify it's set correctly
amp-tasks projects default
` + "```" + `

### Guidelines Best Practices
- **Technical Standards**: Language, frameworks, coding patterns
- **Quality Requirements**: Testing, code review, documentation
- **Business Context**: User needs, project goals, constraints
- **Collaboration Rules**: Communication channels, decision processes

## 2. Agent Type Setup

### Choose Agent Types for Your Project
Consider what types of work your project needs:

#### Development Agent Types
` + "```bash" + `
# Frontend specialists
amp-tasks agent-types create "Frontend Developer" \
  --description "React/TypeScript components, UI/UX implementation, responsive design" \
  --project-id <project-id>

# Backend specialists  
amp-tasks agent-types create "Backend Developer" \
  --description "API development, database design, server infrastructure" \
  --project-id <project-id>

# Full-stack generalists
amp-tasks agent-types create "Full Stack Developer" \
  --description "End-to-end feature development, system integration" \
  --project-id <project-id>
` + "```" + `

#### Quality & Review Agent Types
` + "```bash" + `
# Code quality experts
amp-tasks agent-types create "Code Reviewer" \
  --description "Code quality, security review, architectural guidance" \
  --project-id <project-id>

# Testing specialists
amp-tasks agent-types create "Test Engineer" \
  --description "Test automation, quality assurance, test strategy" \
  --project-id <project-id>

# Security experts
amp-tasks agent-types create "Security Specialist" \
  --description "Security analysis, vulnerability assessment, compliance" \
  --project-id <project-id>
` + "```" + `

#### Operations & Infrastructure Agent Types
` + "```bash" + `
# Infrastructure specialists
amp-tasks agent-types create "DevOps Engineer" \
  --description "CI/CD, deployment, monitoring, infrastructure as code" \
  --project-id <project-id>

# Data specialists
amp-tasks agent-types create "Data Engineer" \
  --description "Data pipelines, analytics, database optimization" \
  --project-id <project-id>
` + "```" + `

#### Management & Coordination Agent Types
` + "```bash" + `
# Project coordination
amp-tasks agent-types create "Tech Lead" \
  --description "Technical leadership, architecture decisions, team coordination" \
  --project-id <project-id>

# Documentation and communication
amp-tasks agent-types create "Documentation Specialist" \
  --description "Technical writing, API docs, user guides, knowledge management" \
  --project-id <project-id>
` + "```" + `

### View Your Agent Types
` + "```bash" + `
# See all agent types for your project
amp-tasks agent-types list --project-id <project-id>
` + "```" + `

## 3. Agent Creation

### Create Initial Agents
` + "```bash" + `
# Create agents for each type (examples)
amp-tasks agents create "Alice Frontend" --agent-type-id <frontend-type-id>
amp-tasks agents create "Bob Backend" --agent-type-id <backend-type-id>
amp-tasks agents create "Carol Review" --agent-type-id <code-reviewer-type-id>
amp-tasks agents create "David Test" --agent-type-id <test-engineer-type-id>
amp-tasks agents create "Eve DevOps" --agent-type-id <devops-type-id>

# View your agent workforce
amp-tasks agents list
` + "```" + `

### Agent Naming Best Practices
- **Descriptive Names**: Include role/specialty ("Alice Frontend", "Security-Bob")
- **Team Organization**: Use prefixes for teams ("Team1-Alice", "Core-Bob")
- **Skill Indication**: Include key skills ("React-Alice", "K8s-Bob")

## 4. Initial Task Structure

### Create Epic-Level Tasks (Top-Level Features)
` + "```bash" + `
# Major feature areas
amp-tasks tasks create "User Authentication System" \
  --description "Complete auth flow: login, registration, password reset, session management"

amp-tasks tasks create "Analytics Dashboard" \
  --description "Real-time charts, filtering, data export, customizable views"

amp-tasks tasks create "API Infrastructure" \
  --description "REST endpoints, rate limiting, validation, error handling"

amp-tasks tasks create "Data Pipeline" \
  --description "ETL processes, data validation, monitoring, alerting"
` + "```" + `

### Create Feature-Level Tasks (Under Epics)
` + "```bash" + `
# Get epic task IDs first
amp-tasks tasks list

# Create sub-features under authentication epic
amp-tasks tasks create "Login Component" \
  --description "React component with form validation, error handling" \
  --parent <auth-epic-id>

amp-tasks tasks create "JWT Token Service" \
  --description "Token generation, validation, refresh logic" \
  --parent <auth-epic-id>

amp-tasks tasks create "Password Reset Flow" \
  --description "Email sending, token validation, password update" \
  --parent <auth-epic-id>

# Create sub-features under dashboard epic  
amp-tasks tasks create "Chart Components" \
  --description "Reusable D3/Chart.js components with responsive design" \
  --parent <dashboard-epic-id>

amp-tasks tasks create "Data Fetching Layer" \
  --description "API client, caching, error handling, loading states" \
  --parent <dashboard-epic-id>
` + "```" + `

### Create Implementation Tasks (Specific Work Items)
` + "```bash" + `
# Get feature task IDs
amp-tasks tasks list

# Implementation tasks under login component
amp-tasks tasks create "Login Form UI" \
  --description "Form layout, styling, responsive design" \
  --parent <login-component-id>

amp-tasks tasks create "Form Validation" \
  --description "Client-side validation, error messages, accessibility" \
  --parent <login-component-id>

amp-tasks tasks create "API Integration" \
  --description "Connect form to auth API, handle responses" \
  --parent <login-component-id>

amp-tasks tasks create "Unit Tests" \
  --description "Test form validation, API mocking, edge cases" \
  --parent <login-component-id>
` + "```" + `

### Task Hierarchy Best Practices
- **Epic Level** (3-6 months): Major features or systems
- **Feature Level** (2-4 weeks): Cohesive functionality within epics  
- **Implementation Level** (1-5 days): Specific, actionable work items
- **Task Level** (2-8 hours): Individual commits or small changes

## 5. Dependency Setup

### Add Cross-Feature Dependencies
` + "```bash" + `
# API must exist before frontend can integrate
amp-tasks deps add <login-component-id> <jwt-service-id>
amp-tasks deps add <chart-components-id> <data-fetching-layer-id>

# Infrastructure before application features
amp-tasks deps add <auth-epic-id> <api-infrastructure-id>
amp-tasks deps add <dashboard-epic-id> <data-pipeline-id>

# Foundation before specialization
amp-tasks deps add <form-validation-id> <login-form-ui-id>
amp-tasks deps add <api-integration-id> <form-validation-id>
amp-tasks deps add <unit-tests-id> <api-integration-id>
` + "```" + `

### Add Quality Gates
` + "```bash" + `
# All implementation tasks need testing
amp-tasks tasks create "Code Review" \
  --description "Security review, code quality, architectural compliance" \
  --parent <login-component-id>

amp-tasks deps add <code-review-id> <unit-tests-id>

# Integration tests depend on unit tests
amp-tasks tasks create "Integration Tests" \
  --description "End-to-end auth flow testing" \
  --parent <auth-epic-id>

amp-tasks deps add <integration-tests-id> <code-review-id>
` + "```" + `

### Dependency Best Practices
- **Technical Dependencies**: API before UI, database before API
- **Quality Gates**: Tests before review, review before merge
- **Risk Management**: Core features before nice-to-have features
- **Team Coordination**: Shared components before features using them

## 6. Verification Steps

### Verify Project Structure
` + "```bash" + `
# Check project is set as default
amp-tasks projects default

# View agent types and coverage
amp-tasks agent-types list

# View agent workforce
amp-tasks agents list

# Check task hierarchy
amp-tasks tasks list
` + "```" + `

### Verify Dependencies
` + "```bash" + `
# Visualize dependency graph
amp-tasks deps graph

# Check for circular dependencies (should be none)
amp-tasks deps validate

# See what's ready to start
amp-tasks tasks available
` + "```" + `

### Verify Assignment Strategy
` + "```bash" + `
# Test flexible assignment
amp-tasks agent-types assign <available-task-id> <frontend-developer-type-id>

# Verify assignment worked
amp-tasks tasks show <task-id>

# Test specific assignment  
amp-tasks tasks assign <another-task-id> <alice-agent-id>

# Check workload distribution
amp-tasks agents workload
` + "```" + `

## 7. Start Development Workflow

### Begin First Sprint
` + "```bash" + `
# Check what's ready to work on
amp-tasks tasks available

# Assign foundational tasks first
amp-tasks agent-types assign <api-infrastructure-task> <backend-developer-type>
amp-tasks agent-types assign <ui-foundation-task> <frontend-developer-type>

# Set up monitoring for progress
amp-tasks notes add <task-id> "Sprint 1 started - focusing on foundation"
` + "```" + `

### Establish Knowledge Sharing
` + "```bash" + `
# Document architectural decisions
amp-tasks til create "Project Architecture" \
  --content "Using React + TypeScript frontend, Node.js + Express backend, PostgreSQL database"

# Share setup insights
amp-tasks til create "Development Environment" \
  --content "Use Docker for local DB, pnpm for package management, runs on port 3000"
` + "```" + `

## Quick Setup Template

For a rapid setup, copy and adapt this sequence:

` + "```bash" + `
#!/bin/bash
# Quick project setup script

# 1. Create project
PROJECT_ID=$(amp-tasks projects create "Your Project" --description "Description" --guidelines "Your guidelines" --output json | jq -r '.id')
amp-tasks projects set-default $PROJECT_ID

# 2. Create agent types
FRONTEND_TYPE=$(amp-tasks agent-types create "Frontend Dev" --description "UI/UX" --project-id $PROJECT_ID --output json | jq -r '.id')
BACKEND_TYPE=$(amp-tasks agent-types create "Backend Dev" --description "API/DB" --project-id $PROJECT_ID --output json | jq -r '.id')
REVIEWER_TYPE=$(amp-tasks agent-types create "Code Reviewer" --description "Quality" --project-id $PROJECT_ID --output json | jq -r '.id')

# 3. Create agents
amp-tasks agents create "Alice" --agent-type-id $FRONTEND_TYPE
amp-tasks agents create "Bob" --agent-type-id $BACKEND_TYPE  
amp-tasks agents create "Carol" --agent-type-id $REVIEWER_TYPE

# 4. Create initial tasks
EPIC1=$(amp-tasks tasks create "Core Features" --description "Main functionality" --output json | jq -r '.id')
EPIC2=$(amp-tasks tasks create "Infrastructure" --description "Foundation" --output json | jq -r '.id')

# 5. Add dependencies
amp-tasks deps add $EPIC1 $EPIC2

# 6. Verify
amp-tasks projects default
amp-tasks tasks available
amp-tasks deps graph
` + "```" + `

## Next Steps

After setup is complete:

1. **Start Small**: Begin with foundational tasks
2. **Document Progress**: Use notes and TIL entries actively  
3. **Review Dependencies**: Adjust as you learn more about the work
4. **Iterate Structure**: Add more tasks and agent types as needed
5. **Monitor Workload**: Use ` + "`agents workload`" + ` to balance assignments
6. **Share Knowledge**: Create TIL entries for discoveries and best practices

For ongoing work, see: ` + "`amp-tasks docs agent-guide`" + ` and ` + "`amp-tasks docs workflow`" + `
`
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Project Setup Guide", setup, raw)
	},
}

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Show all available commands summary",
	Long:  "Display a summary of all available commands organized by category",
	RunE: func(cmd *cobra.Command, args []string) error {
		commands := `# Commands Reference

## Understanding the System (Start Here)
amp-tasks agent-types list                 # List all agent types in system
amp-tasks agents list                      # List all agents and their types  
amp-tasks projects default                 # Show current project & guidelines
amp-tasks tasks available                  # Tasks ready for assignment

## Agent Type Management (Essential for Task Assignment)
amp-tasks agent-types list                 # List all agent types
amp-tasks agent-types create <name>        # Create new agent type
amp-tasks agent-types assign <task> <type> # Assign task to agent type (flexible)
amp-tasks agent-types show <id>            # Show agent type details

## Agent Management  
amp-tasks agents list                      # List all agents and their types
amp-tasks agents create <name>             # Create new agent
amp-tasks agents workload                  # Agent workload distribution
amp-tasks agents stats                     # Agent performance stats
amp-tasks agents show <id>                 # Show agent details

## Project Management
amp-tasks projects list                    # List all projects
amp-tasks projects create <name>           # Create new project
amp-tasks projects default                 # Show default project & guidelines
amp-tasks projects set-default <id>        # Set default project

## Task Management  
amp-tasks tasks list                       # List tasks (shows project context)
amp-tasks tasks available                  # Tasks ready for assignment
amp-tasks tasks create <title>             # Create new task
amp-tasks tasks show <id>                  # Task details & dependencies
amp-tasks tasks assign <task-id> <agent>   # Assign task to specific agent (targeted)
amp-tasks tasks status <id> <status>       # Update task status

## Dependencies
amp-tasks deps add <task> <depends-on>     # Add dependency
amp-tasks deps list <task-id>              # List task dependencies  
amp-tasks deps graph                       # Visualize dependency graph
amp-tasks deps show <task-id>              # Show dependency details with notes

## Knowledge Sharing
### TIL Management  
amp-tasks til create <title> --content <c> # Create TIL entry
amp-tasks til create <title> --content <c> --task <id> # Task-specific TIL
amp-tasks til list                         # List TILs for default project
amp-tasks til list --project <id>          # List TILs for specific project
amp-tasks til list --task <id>             # List TILs for specific task
amp-tasks til show <til-id>                # Show detailed TIL

### Notes Management
amp-tasks notes add <task-id> <content>    # Add note to task
amp-tasks notes list <task-id>             # List notes for task
amp-tasks notes list <task-id> --agent <id> # Filter notes by agent
amp-tasks notes show <note-id>             # Show detailed note

## Utilities
amp-tasks demo                             # Create sample data
amp-tasks docs <topic>                     # View documentation

## Assignment Strategy Examples

### Flexible Assignment (Use Agent Types)
# For any qualified agent to pick up work
amp-tasks agent-types assign review-task-123 code-reviewer-type
amp-tasks agent-types assign test-task-456 test-runner-type

### Targeted Assignment (Use Specific Agents)  
# For specific expertise or context
amp-tasks tasks assign auth-task-789 alice-security-expert
amp-tasks tasks assign db-migration-101 bob-database-admin

## Output Formats
Add --output json|yaml|csv to most commands for different formats.

## Common Flags
--db <path>        # Database file path
--log-level <lvl>  # Logging level (debug, info, warn, error)
--help             # Command help
`
		raw, _ := cmd.Flags().GetBool("raw")
		return displayDoc("Commands Reference", commands, raw)
	},
}

func displayDoc(title, content string, raw bool) error {
	if raw {
		fmt.Print(content)
		return nil
	}

	// Format for terminal display
	fmt.Printf("═══ %s ═══\n\n", strings.ToUpper(title))

	// Add some basic formatting for better readability
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		// Skip the top-level title since we already showed it
		if strings.HasPrefix(line, "# ") && strings.Contains(line, title) {
			continue
		}
		fmt.Println(line)
	}

	fmt.Printf("\n═══ END %s ═══\n", strings.ToUpper(title))
	return nil
}

func init() {
	rootCmd.AddCommand(docsCmd)

	// Add subcommands
	docsCmd.AddCommand(readmeCmd)
	docsCmd.AddCommand(agentGuideCmd)
	docsCmd.AddCommand(quickStartCmd)
	docsCmd.AddCommand(workflowCmd)
	docsCmd.AddCommand(setupCmd)
	docsCmd.AddCommand(commandsCmd)

	// Add flags for raw output
	for _, cmd := range []*cobra.Command{readmeCmd, agentGuideCmd, quickStartCmd, workflowCmd, setupCmd, commandsCmd} {
		cmd.Flags().Bool("raw", false, "Output raw markdown without formatting")
	}
}
