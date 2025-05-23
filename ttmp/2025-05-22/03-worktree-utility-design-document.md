# Worktree TUI - Quick Workspace Setup Tool

## Overview

A Terminal User Interface (TUI) application that allows developers to quickly create Go workspaces by selecting from a predefined list of repositories and automatically setting up worktrees with `go.work` initialization.

## Purpose

- **Problem**: Setting up development workspaces with multiple related repositories is repetitive and error-prone
- **Solution**: Interactive TUI that automates the process of creating worktrees, and initializing Go workspaces
- **Target Users**: Go developers working with multiple related repositories (microservices, monorepo components, etc.)

## Core Features

### 1. Repository Selection Interface
- **Multi-select list** of available repositories from config
- **Search/filter** functionality to quickly find repos
- **Repository metadata display**: description, last updated, current branch

### 2. Workspace Configuration
- **Workspace name input** with smart defaults (based on selected repos)
- **Target directory selection** with path validation
- **Conflict detection** for existing directories

### 3. Execution & Progress
- **Error handling** with retry options

## Configuration File Format

```yaml
# ~/.config/worktree-tui/config.yaml
workspaces:
  default_base_path: "~/code/workspaces"
  
repositories:
  - name: "pocketflow"
    description: "Minimalist LLM framework"
    local_path: "~/code/others/llms/PocketFlow"  # optional: use existing local repo
    default_branch: "main"
    tags: ["llm", "framework", "core"]
    
  - name: "geppetto"
    description: "Corporate headquarters automation"
    local_path: "~/code/wesen/corporate-headquarters"
    subdirectory: "geppetto"  # for monorepos
    default_branch: "main"
    tags: ["automation", "corporate"]
    
  - name: "ai-tools"
    description: "AI development utilities"
    default_branch: "develop"
    tags: ["ai", "utilities"]

presets:
  - name: "LLM Development"
    description: "Full stack LLM development environment"
    repositories: ["pocketflow", "geppetto", "ai-tools"]
    
  - name: "Corporate Automation"
    description: "Corporate tools and automation"
    repositories: ["geppetto", "ai-tools"]
```

## User Interface Design

### Main Screen
```
┌─ Worktree TUI - Quick Workspace Setup ─────────────────────────────────┐
│                                                                         │
│ Select repositories for your workspace:                                 │
│                                                                         │
│ Search: [llm________________]                                           │
│                                                                         │
│ Presets:                                                                │
│ ○ LLM Development        ○ Corporate Automation                         │
│                                                                         │
│ Repositories:                                                           │
│ ☑ pocketflow            Minimalist LLM framework                       │
│ ☐ geppetto              Corporate headquarters automation               │
│ ☑ ai-tools              AI development utilities                       │
│                                                                         │
│ Workspace Configuration:                                                │
│ Name: [llm-workspace_______________]                                    │
│ Path: [~/code/workspaces/llm-workspace]                                │
│                                                                         │
│ [Create Workspace]  [Cancel]                                           │
└─────────────────────────────────────────────────────────────────────────┘
```

### Progress Screen
```
┌─ Creating Workspace: llm-workspace ────────────────────────────────────┐
│                                                                         │
│ ✓ Creating workspace directory                                         │
│ ✓ Setting up pocketflow worktree...                                    │
│ ⟳ Setting up ai-tools worktree...                                      │
│ ○ Initializing go.work                                                  │
│                                                                         │
│ Current: git worktree add ../ai-tools ~/code/ai-tools                  │
│                                                                         │
│ Logs:                                                                    │
│                                                                         │
│ 2025-05-22 10:00:00 │ Creating workspace directory: ~/code/workspaces/llm-workspace │
│ 2025-05-22 10:00:01 │ git worktree add ../pocketflow ~/code/pocketflow            │
│ 2025-05-22 10:00:02 │ Preparing to clone ai-tools repository...                  │
│ 2025-05-22 10:00:03 │ git worktree add ../ai-tools ~/code/ai-tools                │
│ 2025-05-22 10:00:04 │ Initializing go.work file with selected modules            │
│                                                                         │
│   [Cancel]                                                  │
└─────────────────────────────────────────────────────────────────────────┘
```

## Technical Implementation

### Architecture
```
cmd/
├── main.go                 # CLI entry point
├── tui/
│   ├── app.go             # Main TUI application
│   ├── screens/
│   │   ├── selection.go   # Repository selection screen
│   │   ├── config.go      # Workspace configuration screen
│   │   └── progress.go    # Progress/execution screen
│   └── components/
│       ├── repolist.go    # Repository list component
│       ├── search.go      # Search/filter component
│       └── progress.go    # Progress indicator component
├── config/
│   ├── loader.go          # Configuration file loading
│   └── types.go           # Configuration data structures
├── workspace/
│   ├── manager.go         # Workspace creation logic
│   ├── git.go             # Git operations (worktree)
│   └── golang.go          # Go workspace initialization
```

### Key Dependencies
- **TUI Framework**: [bubbletea](https://github.com/charmbracelet/bubbletea) + [lipgloss](https://github.com/charmbracelet/lipgloss)
- **Configuration**: [viper](https://github.com/spf13/viper) for YAML config loading
- **Git Operations**: [go-git](https://github.com/go-git/go-git) or shell commands
- **CLI Framework**: [cobra](https://github.com/spf13/cobra) for command structure

## Workflow Steps

### 1. Repository Selection
1. Load configuration file
2. Display repository list with search/filter
3. Allow multi-select with keyboard navigation
4. Show preset options for quick selection
5. Validate selection (at least one repo)

### 2. Workspace Configuration
1. Generate default workspace name from selected repos
2. Allow user to customize name and path
3. Validate target directory doesn't exist
4. Show branch selection for each repository
5. Preview final workspace structure

### 3. Workspace Creation
1. Create workspace base directory
2. For each repository:
   - Create worktree in workspace directory
   - Handle subdirectory extraction if needed
3. Initialize `go.work` file with all Go modules
4. Display success summary with next steps

### Recovery Options
- **Partial failure**: Continue with successful repos, report failures
- **Cleanup on abort**: Remove partially created workspace
- **Retry mechanisms**: Allow retrying failed operations

## Configuration Management

### Default Locations
1. `~/.config/worktree-tui/config.yaml`
2. `./worktree-tui.yaml` (project-specific)
3. Environment variable: `WORKTREE_TUI_CONFIG`

### Validation
- Repository URLs are accessible
- Local paths exist and are valid Git repositories
- Branch names exist in repositories
- No circular dependencies in presets
