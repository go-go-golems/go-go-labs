# Multi-Agent Tmux Orchestrator with TUI

A demonstration of a multi-agent system with beautiful real-time terminal user interfaces (TUI), featuring mock LLM agents working concurrently on different tasks.

## Features

🤖 **Multiple Agent Types**:
- **Research Agent**: Conducts research and gathers information
- **Analysis Agent**: Analyzes data and provides insights  
- **Writing Agent**: Creates written content and documentation
- **Review Agent**: Reviews and provides feedback on work

🎨 **Beautiful TUI Visualization**:
- Each agent gets its own styled TUI in tmux panes
- Real-time updates via Unix socket communication
- Color-coded message types with emojis
- Clean, professional output using Bubbletea framework
- No more messy echo commands or log tailing!

⚡ **Advanced Architecture**:
- Unix socket-based communication protocol
- JSON message passing with structured data
- Concurrent execution across agents
- Graceful error handling and resource cleanup
- Custom TUI components with responsive layouts

## Architecture

```
Orchestrator Agent
├── Socket Server (Unix Domain Sockets)
│   ├── orchestrator.sock → Status TUI
│   ├── research-001.sock → Research TUI  
│   ├── analysis-001.sock → Analysis TUI
│   ├── writing-001.sock → Writing TUI
│   └── review-001.sock → Review TUI
├── Tmux Manager
│   ├── Pane 1: Status TUI (Bubbletea)
│   ├── Pane 2: Research Agent TUI (Bubbletea)
│   ├── Pane 3: Analysis Agent TUI (Bubbletea)
│   ├── Pane 4: Writing Agent TUI (Bubbletea)
│   └── Pane 5: Review Agent TUI (Bubbletea)
├── Mock LLM Agents
│   ├── Research Agent (research-001)
│   ├── Analysis Agent (analysis-001)
│   ├── Writing Agent (writing-001)
│   └── Review Agent (review-001)
└── Protocol Layer (JSON over Unix sockets)
```

## Usage

### Basic Usage
```bash
# Run with default settings (4 tasks, 30s duration)
go run ./cmd/experiments/multi-agent-tmux

# Customize session and task count
go run ./cmd/experiments/multi-agent-tmux --session my-agents --tasks 6

# Debug mode with longer duration
go run ./cmd/experiments/multi-agent-tmux --log-level debug --duration 60s
```

### Interactive Mode
```bash
# Start in interactive mode - waits for user input
go run ./cmd/experiments/multi-agent-tmux --interactive --session demo
```

### Viewing Agent Output
```bash
# Attach to the tmux session to see beautiful TUI output
tmux attach -t multi-agent

# List all sessions
tmux list-sessions

# Navigate between panes in tmux:
# Ctrl+b + arrow keys (move between panes)
# Ctrl+b + o (cycle through panes)
# Ctrl+b + q (show pane numbers)

# Within each TUI:
# Press 'q' to quit individual TUI
# Ctrl+c to force quit
```

### TUI Features
- **Color-coded Messages**: Status (green), Progress (yellow), Results (purple), Errors (red)
- **Real-time Updates**: Messages appear instantly via socket communication
- **Scrollable History**: Last 50 messages per agent with automatic scrolling
- **Responsive Layout**: Adapts to terminal size changes
- **Clean Design**: Professional styling with borders and headers

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `--session` | Tmux session name | `multi-agent` |
| `--tasks` | Number of tasks to execute | `4` |
| `--duration` | Maximum duration for execution | `30s` |
| `--log-level` | Log level (debug/info/warn/error) | `info` |
| `--interactive` | Wait for user input before starting | `false` |

## Agent Behavior

### Research Agent
- Simulates academic research workflow
- Steps: database search, publication review, source validation
- Outputs: findings, source citations, research summaries

### Analysis Agent  
- Performs data analysis and statistical methods
- Steps: data preprocessing, correlation analysis, pattern identification
- Outputs: statistical results, trends, recommendations

### Writing Agent
- Creates structured written content
- Steps: planning, outlining, writing, editing, formatting
- Outputs: word counts, section completions, bibliography updates

### Review Agent
- Reviews content for quality and accuracy
- Steps: structure check, fact verification, grammar review
- Outputs: feedback, suggestions, quality assessments

## Example Session Output

```
[13:07:15] 🚀 Multi-Agent Orchestrator Initialized
[13:07:15] 📊 Session: demo-agents
[13:07:15] 🤖 Agents: 4 registered
[13:07:16] 🎯 Starting execution of 3 tasks
[13:07:16] ▶️  Starting task task-001 on Writing Agent
[13:07:16] ▶️  Starting task task-002 on Review Agent
[13:07:16] ▶️  Starting task task-003 on Analysis Agent
[13:07:45] ✅ Task task-001 completed by Writing Agent
[13:07:48] ✅ Task task-002 completed by Review Agent
[13:07:52] ✅ Task task-003 completed by Analysis Agent
[13:07:52] 🎉 All tasks completed successfully!
```

## Implementation Details

### Core Components

1. **Agent Interface**: Common interface for all agent types with Execute method
2. **SocketMessage**: JSON protocol for Unix socket communication
3. **TUIModel**: Bubbletea-based terminal user interface with real-time updates
4. **SocketServer**: Unix domain socket server managing multiple connections
5. **Orchestrator**: Coordinates agents, sockets, and tmux panes
6. **Task**: Work units with ID, description, and target agent type

### Socket Communication Protocol

```json
{
  "type": "agent_update",
  "agent_id": "research-001", 
  "agent_name": "Research Agent",
  "agent_role": "Conducts research and gathers information",
  "content": "Found relevant study on distributed systems",
  "msg_type": "result",
  "timestamp": "2025-06-16T13:45:30Z"
}
```

### TUI Architecture

- **Bubbletea Framework**: Modern TUI framework with clean abstractions
- **Lipgloss Styling**: Beautiful colors and formatting
- **Event-driven Updates**: Socket messages trigger UI refreshes
- **Responsive Design**: Adapts to terminal resize events
- **Message History**: Scrollable buffer with automatic cleanup

### Tmux Integration

- Uses `github.com/GianlucaP106/gotmux` for tmux control
- Creates panes instead of windows for better layout
- Launches TUI processes in each pane
- Automatic socket path management and cleanup

### Concurrency Model

- Goroutines for agent execution and socket handling
- Channels for message passing between agents and orchestrator
- Unix sockets for inter-process communication with TUIs
- Context-based cancellation and graceful shutdown

## Extension Points

The system is designed for easy extension:

1. **New Agent Types**: Implement the `Agent` interface
2. **Custom Tasks**: Extend the `Task` struct and generation logic
3. **Message Types**: Add new message types for different output formats
4. **Visualization**: Enhance tmux output with colors, formatting, or additional panes

## Dependencies

- `github.com/GianlucaP106/gotmux` - Tmux control library
- `github.com/charmbracelet/bubbletea` - Modern TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling and layout
- `github.com/spf13/cobra` - CLI framework
- `github.com/rs/zerolog` - Structured logging
- Standard Go libraries for Unix sockets, JSON, and concurrency
