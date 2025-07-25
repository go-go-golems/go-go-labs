# Commands Reference

## Understanding the System (Start Here)
amp-tasks agent-types list                 # List all agent types in system
amp-tasks agents list                      # List all agents and their types  
amp-tasks projects default                 # Show current project & guidelines
amp-tasks tasks available                  # Tasks ready for assignment

## Agent Type Management (Essential for Task Assignment)
amp-tasks agent-types list                 # List all agent types
amp-tasks agent-types create <n>        # Create new agent type
amp-tasks agent-types assign <task> <type> # Assign task to agent type (flexible)
amp-tasks agent-types show <id>            # Show agent type details

## Agent Management  
amp-tasks agents list                      # List all agents and their types
amp-tasks agents create <n>             # Create new agent
amp-tasks agents workload                  # Agent workload distribution
amp-tasks agents stats                     # Agent performance stats
amp-tasks agents show <id>                 # Show agent details

## Project Management
amp-tasks projects list                    # List all projects
amp-tasks projects create <n>           # Create new project
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
