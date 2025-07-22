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
- ` + "`agent_type_id`" + `: Reference to agent type (optional)

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

1. **Check available tasks** - ` + "`tasks available`" + `
2. **Understand project context** - ` + "`projects default`" + ` 
3. **Read insights from dependencies** - ` + "`deps show <task-id>`" + ` (if any)
4. **Assign work to yourself** - ` + "`tasks assign <task-id> <agent-id>`" + `
5. **Take notes during work** - ` + "`notes add <task-id> 'progress update'`" + `
6. **Complete the work** following project guidelines
7. **Share insights** - ` + "`til create 'learning title' --content 'what you learned'`" + `
8. **Mark completed** - ` + "`tasks status <task-id> completed`" + `
9. **Check new available tasks** (shown automatically)

## Key Principles

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

## 2. Check project context
amp-tasks projects default

## 3. See available work
amp-tasks tasks available

## 4. Read insights from dependencies (if any)
amp-tasks deps show <task-id>

## 5. Assign work to yourself
amp-tasks tasks assign <task-id> <agent-id>

## 6. Track progress with notes
amp-tasks notes add <task-id> "Working on authentication"

## 7. Share insights
amp-tasks til create "Auth Pattern" --content "Use JWT for stateless auth"

## 8. Update status
amp-tasks tasks status <task-id> completed

## 9. Create new work
amp-tasks tasks create "New task" --description "Details"

## 10. Add dependencies
amp-tasks deps add <task-id> <depends-on-id>

## 11. Visualize work
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

2. **Find Available Work**
   amp-tasks tasks available
   - See tasks ready for assignment
   - Check dependencies are met

3. **Read Previous Work (if dependencies exist)**
   amp-tasks deps show <task-id>
   - Review notes from agents who worked on dependency tasks
   - Learn from TIL entries related to this work

4. **Assign Work**
   amp-tasks tasks assign <task-id> <your-agent-id>
   - Or use agent type: amp-tasks agent-types assign <task-id> <type-id>
   - Task status automatically becomes 'in_progress'

5. **Do the Work**
   - Follow project guidelines
   - Complete the task requirements  
   - Take progress notes: amp-tasks notes add <task-id> "progress update"
   - Check task details: amp-tasks tasks show <task-id>

6. **Share Learning**
   amp-tasks til create "insight title" --content "what you learned"
   - Create task-specific TIL: --task <task-id>
   - Share insights that help other agents

7. **Update Status**
   amp-tasks tasks status <task-id> completed
   - System shows newly available tasks
   - Dependencies are automatically resolved

8. **Create Additional Work (if needed)**
   amp-tasks tasks create "New task" --description "Details"
   amp-tasks deps add <new-task> <depends-on-task>

## Key Principles

- Always check project guidelines first
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

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Show all available commands summary",
	Long:  "Display a summary of all available commands organized by category",
	RunE: func(cmd *cobra.Command, args []string) error {
		commands := `# Commands Reference

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
amp-tasks tasks assign <task-id> <agent>   # Assign task to agent
amp-tasks tasks status <id> <status>       # Update task status

## Agent Management
amp-tasks agents list                      # List all agents
amp-tasks agents create <name>             # Create new agent
amp-tasks agents workload                  # Agent workload distribution
amp-tasks agents stats                     # Agent performance stats

## Agent Types
amp-tasks agent-types list                 # List agent types
amp-tasks agent-types create <name>        # Create agent type
amp-tasks agent-types assign <task> <type> # Assign to agent type

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
	docsCmd.AddCommand(commandsCmd)

	// Add flags for raw output
	for _, cmd := range []*cobra.Command{readmeCmd, agentGuideCmd, quickStartCmd, workflowCmd, commandsCmd} {
		cmd.Flags().Bool("raw", false, "Output raw markdown without formatting")
	}
}
