# Amp Tasks

A comprehensive SQLite-based task management system for coding agents with hierarchical task planning, DAG dependencies, project management, and agent typing.

## Installation

```bash
go run ./cmd/experiments/amp-tasks
```

## Quick Start

```bash
# Initialize amp-tasks project
amp-tasks projects create "amp-tasks" --description "Task coordination system development" --guidelines "Follow Go conventions, write tests, document APIs"

# Set as default project
amp-tasks projects set-default amp-tasks

# Create initial development tasks
amp-tasks tasks create "Setup project structure" --description "Initialize Go modules and directory structure"
amp-tasks tasks create "Implement core CLI" --description "Build basic command structure with cobra"
amp-tasks tasks create "Add database layer" --description "Implement SQLite schema and basic CRUD operations"

# See available work
amp-tasks tasks available

# View project context
amp-tasks projects default
```

## Documentation

Full documentation is embedded in the binary:

```bash
# Agent work guide
amp-tasks docs agent-guide

# Complete documentation
amp-tasks docs readme

# Quick start commands
amp-tasks docs quick-start

# All available commands
amp-tasks docs commands
```

## Features

- Project management with guidelines
- Hierarchical task planning  
- DAG dependency management
- Agent types and smart assignment
- Status tracking and workflows
- Knowledge sharing (TIL system)
- Progress tracking with notes
- Multiple output formats (JSON, YAML, CSV)

For complete feature documentation, run `amp-tasks docs readme`.
