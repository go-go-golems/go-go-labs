# Workspace Manager Architecture Documentation

## Overview

The workspace-manager is a CLI tool designed to manage multi-repository development workflows. It provides discovery, workspace creation, and git operations across multiple repositories simultaneously. The architecture follows Go best practices with clear separation of concerns and minimal external dependencies.

## Core Design Principles

### 1. **Minimal Dependencies**
- Uses standard library where possible
- External dependencies: cobra (CLI), bubbles/bubbletea (TUI), zerolog (logging), pkg/errors (error handling)
- No heavy ORM or database dependencies

### 2. **Git-Native Operations**
- All git operations use command-line git (no libgit2 bindings)
- Ensures compatibility with user's git configuration
- Leverages git's robustness and feature completeness

### 3. **Stateless Design**
- Configuration stored in JSON files
- No persistent background processes
- Each command execution is independent

### 4. **Error-First Approach**
- Comprehensive error handling with context
- Graceful degradation when individual repositories fail
- Clear error messages for user debugging

## File Structure and Responsibilities

```
cmd/apps/workspace-manager/
├── main.go                 # Entry point and command registration
├── types.go               # Core data structures
├── utils.go              # Shared utilities
├── discovery.go          # Repository discovery logic
├── workspace.go          # Workspace management
├── status.go            # Status checking operations
├── git_operations.go    # Git command abstractions
├── sync_operations.go   # Synchronization operations
├── tui_models.go       # TUI data models and key bindings
├── tui.go              # TUI implementation
├── cmd_*.go            # Individual command implementations
```

### Core Components

#### 1. **Command Layer** (`cmd_*.go`)
**Purpose**: CLI interface and argument parsing
**Files**: `cmd_discover.go`, `cmd_create.go`, `cmd_status.go`, etc.
**Responsibilities**:
- Parse command-line arguments using cobra
- Validate input parameters
- Orchestrate business logic calls
- Format output for users

```go
// Pattern used in all command files
func NewXXXCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "xxx",
        Short: "description",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runXXX(cmd.Context(), /* parsed args */)
        },
    }
    // Flag definitions
    return cmd
}

func runXXX(ctx context.Context, args...) error {
    // Input validation
    // Business logic delegation
    // Output formatting
}
```

#### 2. **Business Logic Layer**
**Purpose**: Core application logic without CLI concerns
**Files**: `discovery.go`, `workspace.go`, `git_operations.go`, `sync_operations.go`

##### Repository Discovery (`discovery.go`)
**Class**: `RepositoryDiscoverer`
**Key Methods**:
- `DiscoverRepositories(paths, recursive, maxDepth)` - Main discovery entry point
- `scanDirectory()` - Recursive directory scanning
- `analyzeRepository()` - Extract git repository metadata
- `categorizeRepository()` - Determine repository type (Go, Node, etc.)

**Design Decisions**:
- **Metadata Caching**: Repository information cached in registry to avoid re-scanning
- **Categorization**: File-based heuristics for technology detection
- **Parallel Scanning**: Could be added without API changes

```go
type RepositoryDiscoverer struct {
    registry     *RepositoryRegistry
    registryPath string
}

// Git operations abstracted through helper methods
func (rd *RepositoryDiscoverer) getGitRemoteURL(ctx, path) (string, error)
func (rd *RepositoryDiscoverer) getGitCurrentBranch(ctx, path) (string, error)
```

##### Workspace Management (`workspace.go`)
**Class**: `WorkspaceManager`
**Key Methods**:
- `CreateWorkspace()` - Complete workspace creation workflow
- `createWorktree()` - Git worktree creation
- `createGoWorkspace()` - Go workspace file generation

**Design Decisions**:
- **Worktree Strategy**: Uses git worktrees for isolation without duplication
- **Go Integration**: Automatic go.work file creation for Go repositories
- **Template System**: AGENT.md copying for consistent workspace setup

```go
type WorkspaceManager struct {
    config       *WorkspaceConfig
    discoverer   *RepositoryDiscoverer
    workspaceDir string
}

// Workspace creation flow
func (wm *WorkspaceManager) CreateWorkspace(ctx, name, repos, branch, agentSource, dryRun) (*Workspace, error) {
    // 1. Validate and find repositories
    // 2. Create workspace directory structure
    // 3. Create worktrees for each repository
    // 4. Initialize Go workspace if applicable
    // 5. Copy templates
    // 6. Save workspace configuration
}
```

##### Git Operations (`git_operations.go`)
**Class**: `GitOperations`
**Key Methods**:
- `GetWorkspaceChanges()` - Aggregate changes across repositories
- `CommitChanges()` - Cross-repository commit operations
- `StageFile()`/`UnstageFile()` - Individual file staging

**Design Decisions**:
- **Command Abstraction**: Git operations wrapped in Go functions with error handling
- **Batch Operations**: Support for atomic operations across multiple repositories
- **Dry-Run Support**: Preview mode for all destructive operations

```go
type GitOperations struct {
    workspace *Workspace
}

type CommitOperation struct {
    Message     string
    Files       map[string][]FileChange  // repo -> files
    DryRun      bool
    AddAll      bool
    Push        bool
}
```

##### Sync Operations (`sync_operations.go`)
**Class**: `SyncOperations`
**Key Methods**:
- `SyncWorkspace()` - Coordinate sync across all repositories
- `CreateBranch()`/`SwitchBranch()` - Branch management
- `GetWorkspaceLog()` - Aggregated commit history

**Design Decisions**:
- **Result Aggregation**: Detailed per-repository results for user feedback
- **Conflict Detection**: Identify and report merge conflicts
- **Branch Coordination**: Ensure consistent branch states across repositories

#### 3. **Data Layer** (`types.go`)
**Purpose**: Core data structures and persistence
**Key Types**:

```go
type Repository struct {
    Name        string    `json:"name"`
    Path        string    `json:"path"`
    RemoteURL   string    `json:"remote_url"`
    CurrentBranch string  `json:"current_branch"`
    Branches    []string  `json:"branches"`
    Tags        []string  `json:"tags"`
    Categories  []string  `json:"categories"`
}

type Workspace struct {
    Name         string       `json:"name"`
    Path         string       `json:"path"`
    Repositories []Repository `json:"repositories"`
    Branch       string       `json:"branch"`
    Created      time.Time    `json:"created"`
    GoWorkspace  bool         `json:"go_workspace"`
}

type RepositoryStatus struct {
    Repository     Repository `json:"repository"`
    HasChanges     bool       `json:"has_changes"`
    StagedFiles    []string   `json:"staged_files"`
    ModifiedFiles  []string   `json:"modified_files"`
    UntrackedFiles []string   `json:"untracked_files"`
    Ahead          int        `json:"ahead"`
    Behind         int        `json:"behind"`
}
```

**Design Decisions**:
- **JSON Serialization**: All types support JSON for configuration persistence
- **Rich Metadata**: Comprehensive repository information for decision-making
- **Status Snapshots**: Point-in-time status captures for reporting

#### 4. **TUI Layer** (`tui.go`, `tui_models.go`)
**Purpose**: Interactive terminal user interface
**Framework**: Bubble Tea (charm.sh)
**Key Components**:

```go
type mainModel struct {
    state           appState
    discoverer     *RepositoryDiscoverer
    workspaceManager *WorkspaceManager
    repoList       list.Model
    workspaceList  list.Model
    selectedRepos  map[string]bool
    // Form components for workspace creation
}

type appState int
const (
    stateMain appState = iota
    stateRepositories
    stateWorkspaces
    stateCreateWorkspace
    stateWorkspaceForm
)
```

**Design Decisions**:
- **State Machine**: Clear state transitions for navigation
- **Component Reuse**: Bubble Tea list components for consistent UI
- **Form Handling**: Multi-step form workflow for workspace creation

## Configuration Management

### File Locations
- **Registry**: `~/.config/workspace-manager/registry.json`
- **Workspaces**: `~/.config/workspace-manager/workspaces/`
- **Default Workspace Dir**: `~/workspaces/YYYY-MM-DD/`

### Configuration Flow
```go
func loadConfig() (*WorkspaceConfig, error) {
    // Respects XDG directories
    // Falls back to sensible defaults
    // Creates directories as needed
}
```

## Git Command Integration

### Command Execution Pattern
```go
func (component *Component) gitOperation(ctx context.Context, repoPath string) error {
    cmd := exec.CommandContext(ctx, "git", "operation", "args...")
    cmd.Dir = repoPath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return errors.Wrapf(err, "git operation failed: %s", string(output))
    }
    
    return nil
}
```

### Error Handling Strategy
1. **Context Cancellation**: All git operations respect context timeouts
2. **Error Wrapping**: pkg/errors used for error context
3. **Output Capture**: Git command output included in error messages
4. **Graceful Degradation**: Individual repository failures don't stop batch operations

## Concurrency Considerations

### Current Implementation
- **Sequential Operations**: Current implementation processes repositories sequentially
- **Context Support**: All operations support cancellation
- **Future Enhancement**: Architecture supports parallel operations

### Potential Improvements
```go
// Example of how parallel operations could be added
func (so *SyncOperations) SyncWorkspaceParallel(ctx context.Context, options *SyncOptions) ([]SyncResult, error) {
    var wg sync.WaitGroup
    results := make([]SyncResult, len(so.workspace.Repositories))
    
    for i, repo := range so.workspace.Repositories {
        wg.Add(1)
        go func(i int, repo Repository) {
            defer wg.Done()
            results[i] = so.syncRepository(ctx, repo.Name, filepath.Join(so.workspace.Path, repo.Name), options)
        }(i, repo)
    }
    
    wg.Wait()
    return results, nil
}
```

## Error Handling Architecture

### Error Types and Handling
1. **User Errors**: Invalid input, missing files - clear error messages
2. **System Errors**: Permission issues, disk space - actionable guidance
3. **Git Errors**: Command failures - include git output for debugging
4. **Network Errors**: Remote operations - retry suggestions

### Error Propagation Pattern
```go
func businessLogic() error {
    if err := lowLevelOperation(); err != nil {
        return errors.Wrap(err, "business context for the error")
    }
    return nil
}

func commandHandler() error {
    if err := businessLogic(); err != nil {
        // Log structured error
        log.Error().Err(err).Msg("operation failed")
        // Return user-friendly message
        return fmt.Errorf("operation failed: %v", err)
    }
    return nil
}
```

## Performance Considerations

### Current Optimizations
1. **Registry Caching**: Avoid re-scanning repositories
2. **Selective Operations**: Only process repositories with changes
3. **Streaming Output**: Status information shown as available

### Scalability Limits
- **Repository Count**: Tested with 50+ repositories
- **Repository Size**: Limited by git command performance
- **Concurrent Operations**: Currently sequential, could be parallelized

### Memory Usage
- **Repository Metadata**: ~1KB per repository in memory
- **Status Information**: Loaded on-demand, not cached
- **TUI State**: Minimal memory footprint

## Testing Architecture

### Test Strategy
- **Real Git Operations**: Tests use actual git repositories
- **Temporary Directories**: Isolated test environments
- **Table-Driven Tests**: Multiple scenarios in single test functions

### Test Utilities Pattern
```go
func setupTestRepo(t *testing.T, name string) string {
    dir := t.TempDir()
    repoPath := filepath.Join(dir, name)
    
    cmd := exec.Command("git", "init", repoPath)
    require.NoError(t, cmd.Run())
    
    // Setup basic repo structure
    return repoPath
}
```

## Extension Points

### Adding New Repository Types
1. **Update categorizeRepository()** in `discovery.go`
2. **Add detection patterns** for new file types
3. **Extend workspace creation logic** if needed

### Adding New Git Operations
1. **Create methods in GitOperations** or SyncOperations
2. **Add command handlers** in cmd_*.go files
3. **Update TUI** if interactive support needed

### Adding New Output Formats
1. **Implement in utils.go** (printJSON pattern)
2. **Update command flags** to support new format
3. **Add format-specific logic** in command handlers

## Security Considerations

### Input Validation
- **Path Sanitization**: All file paths validated and cleaned
- **Command Injection**: Git commands use exec.Command with separated arguments
- **Directory Traversal**: Repository paths validated within expected boundaries

### Privilege Model
- **User Permissions**: Operates with user's git and filesystem permissions
- **No Elevation**: Never requires elevated privileges
- **SSH Key Usage**: Leverages user's existing git authentication

## Future Architecture Enhancements

### Planned Improvements
1. **Plugin System**: Extension mechanism for custom operations
2. **Remote API**: REST API for integration with other tools
3. **Database Backend**: Optional database for large-scale deployments
4. **Parallel Operations**: Concurrent git operations for performance

### Backward Compatibility
- **Configuration Migration**: Automatic upgrade of registry format
- **Command Compatibility**: Maintain existing CLI interface
- **Data Preservation**: Never lose user workspace configurations

## Troubleshooting and Debugging

### Logging Strategy
```go
log.Info().
    Str("workspace", workspace.Name).
    Int("repositories", len(repositories)).
    Msg("starting operation")
```

### Debug Information
- **Git Command Output**: Captured and included in errors
- **Structured Logging**: Machine-readable log format with zerolog
- **Context Propagation**: Request IDs and operation context maintained

### Common Issues and Solutions
1. **Git Authentication**: Use user's existing git configuration
2. **Path Issues**: Comprehensive path validation and error messages
3. **Permission Problems**: Clear guidance on required permissions
4. **Repository Corruption**: Graceful handling with recovery suggestions

This architecture provides a solid foundation for multi-repository management while maintaining simplicity and extensibility.
