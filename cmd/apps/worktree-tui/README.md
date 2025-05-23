# Worktree TUI - Quick Workspace Setup Tool

A Terminal User Interface (TUI) application that allows developers to quickly create Go workspaces by selecting from a predefined list of repositories and automatically setting up worktrees with `go.work` initialization.

## Features

- **Interactive Repository Selection**: Multi-select from configured repositories with search/filter
- **Preset Support**: Quick selection of common repository combinations
- **Automatic Worktree Setup**: Creates git worktrees for selected repositories
- **Go Workspace Initialization**: Automatically creates `go.work` file with all Go modules
- **Progress Tracking**: Real-time progress updates during workspace creation
- **Monorepo Support**: Handle subdirectories in monorepo setups

## Installation

```bash
cd cmd/apps/worktree-tui
go build -o worktree-tui .
```

## Configuration

Create a configuration file at `~/.config/worktree-tui/config.yaml`:

```yaml
workspaces:
  default_base_path: "~/code/workspaces"
  
repositories:
  - name: "my-project"
    description: "My main project"
    local_path: "~/code/my-project"  # Use existing local repo
    default_branch: "main"
    tags: ["go", "main"]
    
  - name: "shared-lib"
    description: "Shared library"
    url: "https://github.com/myorg/shared-lib.git"  # Clone from remote
    default_branch: "main"
    tags: ["go", "library"]
    
  - name: "monorepo-tool"
    description: "Tool from monorepo"
    local_path: "~/code/monorepo"
    subdirectory: "tools/my-tool"  # For monorepos
    default_branch: "main"
    tags: ["tools"]

presets:
  - name: "Full Development"
    description: "Complete development environment"
    repositories: ["my-project", "shared-lib"]
```

### Configuration Locations

The tool searches for configuration in this order:
1. File specified with `--config` flag
2. `~/.config/worktree-tui/config.yaml`
3. `./worktree-tui.yaml` (project-specific)
4. Environment variable: `WORKTREE_TUI_CONFIG`

## Usage

```bash
# Run with default config
./worktree-tui

# Run with custom config
./worktree-tui --config ./my-config.yaml
```

### Navigation

- **Arrow keys**: Navigate lists
- **Space/Enter**: Toggle repository selection
- **/** : Search/filter repositories
- **p**: Cycle through presets
- **Ctrl+A**: Select all visible repositories
- **Ctrl+D**: Clear all selections
- **Tab**: Navigate between form fields
- **c/Enter**: Continue to next screen
- **Esc**: Go back
- **q**: Quit

## How It Works

1. **Repository Selection**: Choose repositories from your configured list
2. **Workspace Configuration**: Set workspace name and path
3. **Workspace Creation**: 
   - Creates workspace directory
   - Sets up git worktrees for each repository
   - Initializes `go.work` file with all Go modules
4. **Completion**: Shows success/failure and next steps

## Repository Types

### Local Repositories
Use existing local Git repositories:
```yaml
- name: "local-project"
  local_path: "~/code/my-project"
  default_branch: "main"
```

### Remote Repositories
Clone from remote URLs:
```yaml
- name: "remote-project"
  url: "https://github.com/user/project.git"
  default_branch: "main"
```

### Monorepo Subdirectories
Link to specific subdirectories in monorepos:
```yaml
- name: "tool"
  local_path: "~/code/monorepo"
  subdirectory: "tools/my-tool"
  default_branch: "main"
```

## Output Structure

```
workspace-name/
├── go.work              # Generated Go workspace file
├── repository-1/        # Git worktree
├── repository-2/        # Git worktree
└── ...
```

## Examples

### Example 1: Go Development Workspace
```bash
# Creates: ~/code/workspaces/go-dev/
# - my-service/     (worktree)
# - shared-lib/     (worktree)  
# - go.work         (workspace file)
```

### Example 2: Microservices Workspace
```bash
# Creates: ~/code/workspaces/microservices/
# - api-gateway/    (worktree)
# - user-service/   (worktree)
# - payment-service/(worktree)
# - go.work         (workspace file)
```

## Troubleshooting

### Repository Not Found
- Verify `local_path` exists and is a Git repository
- Check that `url` is accessible
- Ensure branch exists in repository

### Permission Issues
- Check write permissions for workspace directory
- Verify Git access to remote repositories

### Go Workspace Issues
- Ensure repositories contain valid Go modules (`go.mod`)
- Check Go version compatibility

## Development

To extend or modify the tool:

```bash
# Run tests
go test ./...

# Build
go build -o worktree-tui .

# Run with debug output
./worktree-tui --config config.example.yaml
```

## Architecture

```
cmd/apps/worktree-tui/
├── main.go                 # Entry point
├── cmd/
│   └── root.go            # CLI commands
├── internal/
│   ├── config/            # Configuration handling
│   ├── tui/               # TUI application
│   │   ├── app.go        # Main app
│   │   └── screens/      # UI screens
│   └── workspace/         # Workspace operations
│       ├── manager.go    # Main orchestration
│       ├── git.go        # Git operations
│       └── golang.go     # Go workspace setup
└── README.md
```