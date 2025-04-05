# CLI Usage Instructions

This document provides detailed instructions for using the LumonStream CLI tool.

## Installation

To use the CLI tool, you need to build it first:

```bash
cd LumonStream/cli
go build -o lumonstream-cli
```

You can then move the binary to a location in your PATH for easier access:

```bash
sudo mv lumonstream-cli /usr/local/bin/lumonstream
```

## Global Flags

The CLI tool supports the following global flags:

- `--server`: Specify the server URL (default: "http://localhost:8080")

## Commands

### Get Stream Information

Retrieve the current stream information from the server.

```bash
lumonstream get
```

### Update Stream Information

Update the stream information on the server.

```bash
lumonstream update [flags]
```

Flags:
- `--title`: Stream title
- `--description`: Stream description
- `--language`: Programming language/framework
- `--github-repo`: GitHub repository URL
- `--viewer-count`: Current viewer count
- `--reset-timer`: Reset the stream timer

Examples:
```bash
# Update the stream title
lumonstream update --title "New Stream Title"

# Update multiple fields
lumonstream update --title "New Stream Title" --description "New description" --viewer-count 100

# Reset the timer
lumonstream update --reset-timer
```

### Task Management

Add or update tasks for the stream.

#### Add a new task

```bash
lumonstream task add [flags]
```

Flags:
- `--content`: Task content (required)
- `--status`: Task status (completed, active, or upcoming) (default: "upcoming")

Example:
```bash
lumonstream task add --content "Implement new feature" --status "upcoming"
```

#### Update a task's status

```bash
lumonstream task update [flags]
```

Flags:
- `--id`: Task ID (required)
- `--status`: Task status (completed, active, or upcoming) (required)

Example:
```bash
lumonstream task update --id 3 --status "active"
```

### Server Control

Control the LumonStream server.

#### Check server status

```bash
lumonstream server status
```

#### Start the server

```bash
lumonstream server start [flags]
```

Flags:
- `--port`: Port to run the server on (default: 8080)

Example:
```bash
lumonstream server start --port 9000
```

## Examples

Here are some common usage examples:

### Complete Workflow

```bash
# Start the server
lumonstream server start

# Check if the server is running
lumonstream server status

# Get current stream information
lumonstream get

# Update stream title and description
lumonstream update --title "Building a CLI Tool" --description "Creating a command-line interface with Cobra"

# Add a new task
lumonstream task add --content "Research Cobra library" --status "completed"

# Add another task
lumonstream task add --content "Implement basic commands" --status "active"

# Add an upcoming task
lumonstream task add --content "Write tests"

# Update a task status
lumonstream task update --id 3 --status "completed"
```

### Using with a Remote Server

```bash
# Specify a different server URL
lumonstream --server "http://remote-server:8080" get

# Update stream information on the remote server
lumonstream --server "http://remote-server:8080" update --title "Remote Stream"
```
