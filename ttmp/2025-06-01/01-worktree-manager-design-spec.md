# Workspace Manager Implementation Guide

## Overview

This guide provides a step-by-step approach to building a workspace management tool that solves the problem of working with multiple related git repositories simultaneously. The tool automates the tedious process of setting up development environments where you need to make coordinated changes across several codebases.

## The Problem This Solves

When working on features that span multiple repositories (like refactoring shared libraries or implementing cross-service functionality), developers typically need to:
- Manually create git worktrees for each repository
- Set up proper directory structures
- Initialize build systems (like go.work files)
- Track changes across multiple repos
- Commit related changes with consistent messaging
- Keep repositories synchronized

This tool automates and streamlines this entire workflow.

## Tier 1: Foundation (Core Workspace Management)

### Phase 1.1: Repository Discovery and Registration

**What it does**: The tool learns about all git repositories in your development environment by scanning your filesystem and building a registry of available projects.

**User workflow**:
1. User runs discovery to scan their code directories
2. Tool finds all git repositories and extracts metadata (remote URLs, branches, tags)
3. Tool categorizes repositories and stores this information for future use
4. User can view all discovered repositories and their status

**CLI Examples**:
```bash
# Discover repositories in code directories
workspace-manager discover ~/code ~/projects --recursive --max-depth 3

# List all discovered repositories
workspace-manager list repos --format table

# Filter repositories by tags
workspace-manager list repos --tags go,ai
```

**Value delivered**: No more manually tracking where repositories live or what state they're in. The tool knows about your entire development ecosystem.

### Phase 1.2: Basic Workspace Creation

**What it does**: Creates a structured workspace where multiple repositories can be worked on together as a cohesive unit.

**User workflow**:
1. User specifies a workspace name and selects repositories to include
2. Tool creates directory structure in a designated workspace area
3. Tool creates git worktrees for each selected repository on specified branches
4. Tool initializes go.work file to link Go modules together
5. Tool copies template files (like AGENT.md) into the workspace

**CLI Examples**:
```bash
# Create workspace with specific repositories
workspace-manager create refactor-conversation \
  --repos geppetto,pinocchio,bobatea \
  --branch task/refactor-conversation \
  --agent-source ~/templates/AGENT.md

# Create workspace interactively
workspace-manager create my-workspace --interactive

# Preview workspace creation
workspace-manager create test-workspace --repos geppetto,bobatea --dry-run
```

**Value delivered**: What used to take 10+ manual git commands and careful directory management now happens with a single command. The workspace is immediately ready for development.

### Phase 1.3: Workspace Status Awareness

**What it does**: Provides comprehensive visibility into the state of all repositories within a workspace.

**User workflow**:
1. User enters a workspace and checks overall status
2. Tool shows git status for all repositories simultaneously
3. Tool indicates which repos have changes, commits to push/pull, conflicts, etc.
4. User gets a unified view of their multi-repo feature development

**CLI Examples**:
```bash
# Check status of current workspace
workspace-manager status

# Check status of specific workspace
workspace-manager status refactor-conversation

# Get short status format
workspace-manager status --short

# Include untracked files in status
workspace-manager status --untracked
```

**Value delivered**: Instead of manually checking git status in multiple directories, get instant awareness of your entire workspace state.

### Phase 1.4: Workspace Cleanup and Management

**What it does**: Provides safe cleanup and removal of workspaces when development is complete, with options to preserve or remove workspace files. Uses `git worktree remove` safely by default, protecting against uncommitted changes.

**User workflow**:
1. User completes work in a workspace and wants to clean it up
2. User initiates workspace deletion with options for file handling
3. Tool removes git worktrees using `git worktree remove` (fails safely if uncommitted changes exist)
4. Tool optionally deletes workspace directory and all files
5. Tool removes workspace configuration
6. User can choose to keep workspace files for archival or completely remove them
7. User can force worktree removal with `--force-worktrees` if they want to override safety checks

**CLI Examples**:
```bash
# Delete workspace configuration only (keep files)
workspace-manager delete refactor-conversation

# Delete workspace and all files
workspace-manager delete refactor-conversation --remove-files

# Force delete without confirmation
workspace-manager delete old-workspace --force --remove-files

# Force worktree removal even with uncommitted changes
workspace-manager delete workspace-name --force-worktrees --remove-files

# Interactive deletion with preview
workspace-manager delete workspace-name --interactive
```

**Value delivered**: Clean workspace lifecycle management ensures git repository integrity by safely removing worktrees with protection against data loss, while providing flexibility in file preservation. Prevents accumulation of stale worktrees and workspace clutter.

## Tier 2: Advanced Git Operations

### Phase 2.1: Cross-Repository Change Management

**What it does**: Enables staging and committing related changes across multiple repositories as a coordinated operation.

**User workflow**:
1. User has made changes across several repositories in their workspace
2. User initiates interactive commit process
3. Tool shows all modified files across all repositories
4. User selectively stages files per repository
5. Tool commits changes with consistent messaging across all repos

**CLI Examples**:
```bash
# Commit changes across all repos with interactive file selection
workspace-manager commit --interactive \
  --message "Refactor conversation handling across repos"

# Commit all changes with the same message
workspace-manager commit --add-all \
  --message "Update shared interfaces" --push

# Preview what would be committed
workspace-manager commit --dry-run --interactive

# Use commit message template
workspace-manager commit --template feature --interactive
```

**Value delivered**: Ensures related changes across repositories are committed together with consistent commit messages, maintaining project history coherence.

### Phase 2.2: Synchronized Repository Operations

**What it does**: Performs git operations (pull, push, branch creation) across all workspace repositories simultaneously.

**User workflow**:
1. User wants to sync their workspace with remote repositories
2. Tool pulls latest changes from all repository remotes
3. Tool handles merge conflicts and reports status
4. User can push all committed changes across repositories in one operation

**CLI Examples**:
```bash
# Sync workspace - pull then push all repos
workspace-manager sync --pull --push

# Pull latest changes with rebase
workspace-manager sync --pull --rebase

# Push all commits across workspace repos
workspace-manager sync --push

# Create feature branch across all repos
workspace-manager branch create feature/new-api --track
```

**Value delivered**: Eliminates the tedium of manually syncing multiple repositories and reduces the chance of working with inconsistent versions.

### Phase 2.3: Workspace-Wide Development Operations

**What it does**: Provides development utilities that work across the entire workspace context.

**User workflow**:
1. User can view unified diff of all changes across repositories
2. User can create feature branches across all repositories simultaneously
3. User can view commit history that spans multiple repositories
4. User can set up git hooks that apply to the entire workspace

**CLI Examples**:
```bash
# View diff across all workspace repositories
workspace-manager diff --staged

# Show diff for specific repository only
workspace-manager diff --repo geppetto

# View commit history across workspace
workspace-manager log --since "1 week ago" --oneline

# Install git hooks across all repos
workspace-manager hooks install pre-commit --script "go fmt ./..."

# Switch all repos to specific branch
workspace-manager branch switch main
```

**Value delivered**: Development operations that naturally span multiple repositories become as easy as single-repository operations.

## Tier 3: Interactive User Experience

### Phase 3.1: Visual Repository Management

**What it does**: Provides an intuitive interface for discovering, selecting, and organizing repositories without memorizing command syntax.

**User workflow**:
1. User launches interactive mode
2. Visual interface shows all available repositories with metadata
3. User can filter, search, and tag repositories
4. User builds workspace by selecting repositories visually
5. User sees real-time preview of workspace that will be created

**TUI Examples**:
```
┌─ Workspace Manager ─────────────────────────────────────────────────────────┐
│                                                                             │
│ [Repositories] [Workspaces] [Templates] [Settings]                         │
│                                                                             │
│ ┌─ Available Repositories ─────────────────────────────────────────────────┐ │
│ │ ● geppetto        ~/code/geppetto        [go, ai, cli]                  │ │
│ │ ● pinocchio       ~/code/pinocchio       [go, tui]                      │ │
│ │ ● bobatea         ~/code/bobatea         [go, bubbles]                  │ │
│ │ ○ another-repo    ~/code/another-repo    [archived]                     │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ ┌─ Actions ───────────────────────────────────────────────────────────────┐ │
│ │ [C] Create Workspace  [D] Delete Workspace  [R] Refresh  [Q] Quit       │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ Filter: [________________] Tags: [go] [ai] [tui] [archived]                │
└─────────────────────────────────────────────────────────────────────────────┘
```

**CLI Example**:
```bash
# Launch interactive interface
workspace-manager tui

# Start with specific workspace selected
workspace-manager tui --workspace refactor-conversation
```

**Value delivered**: Makes the tool accessible to users who prefer visual interfaces and reduces cognitive load of remembering repository names and locations.

### Phase 3.2: Interactive Change Management

**What it does**: Provides visual tools for managing changes across multiple repositories with granular control.

**User workflow**:
1. User enters commit mode in the interface
2. Interface shows file-by-file changes across all repositories
3. User can review diffs, select files to stage, and craft commit messages
4. User sees preview of exactly what will be committed where
5. User executes commits with confidence in the outcome

**TUI Examples**:
```
┌─ Interactive Commit ────────────────────────────────────────────────────────┐
│                                                                             │
│ Commit Message: [Refactor conversation handling across repos              ]│
│                                                                             │
│ ┌─ geppetto ──────────────────────────────────────────────────────────────┐ │
│ │ [✓] pkg/conversation/manager.go      M   +45  -12                       │ │
│ │ [✓] pkg/conversation/types.go        M   +23   -5                       │ │
│ │ [ ] internal/debug/logger.go         M   +2    -1                       │ │
│ │ [✓] go.mod                          M   +1     -0                       │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ ┌─ pinocchio ─────────────────────────────────────────────────────────────┐ │
│ │ [✓] pkg/ui/conversation.go           M   +67  -23                       │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ [Commit] [Preview] [Cancel]  Selected: 4 files in 2 repos                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Workspace Status Screen**:
```
┌─ Workspace Status: refactor-geppetto-conversation ─────────────────────────┐
│                                                                             │
│ ┌─ Repository Status ─────────────────────────────────────────────────────┐ │
│ │ ● geppetto        main  ↑2 ↓1  [M:3 S:1 U:2] 🔄                       │ │
│ │ ● pinocchio       main  ↑0 ↓0  [M:1 S:0 U:0] ✓                        │ │
│ │ ● bobatea         main  ↑1 ↓3  [M:0 S:0 U:1] ⚠️                        │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ [C] Commit All  [S] Sync  [D] Show Diff  [I] Interactive Commit            │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Value delivered**: Transforms complex multi-repository change management into an intuitive, visual process that reduces errors and increases confidence.

### Phase 3.3: Workspace Lifecycle Management

**What it does**: Provides complete workspace lifecycle management through visual interfaces.

**User workflow**:
1. User can create, modify, and delete workspaces through guided interfaces
2. User can manage workspace templates for common project setups
3. User can monitor workspace health and get recommendations
4. User can archive completed workspaces while preserving important artifacts

**TUI Examples**:
```
┌─ Create Workspace ──────────────────────────────────────────────────────────┐
│                                                                             │
│ Workspace Name: [refactor-geppetto-conversation]                           │
│ Base Directory: [~/workspaces/2025-06-01/                  ] [Browse]      │
│ Branch Name:    [task/refactor-geppetto-conversation       ]               │
│                                                                             │
│ ┌─ Select Repositories ───────────────────────────────────────────────────┐ │
│ │ [✓] geppetto        ~/code/geppetto                                     │ │
│ │ [✓] pinocchio       ~/code/pinocchio                                    │ │
│ │ [✓] bobatea         ~/code/bobatea                                      │ │
│ │ [ ] another-repo    ~/code/another-repo                                 │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ ┌─ Options ───────────────────────────────────────────────────────────────┐ │
│ │ [✓] Initialize go.work file                                             │ │
│ │ [✓] Copy AGENT.md from: [~/templates/AGENT.md          ] [Browse]       │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ [Create] [Preview] [Cancel]                                                │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Workspace Preview Screen**:
```
┌─ Workspace Preview ─────────────────────────────────────────────────────────┐
│                                                                             │
│ ┌─ Actions to be performed ───────────────────────────────────────────────┐ │
│ │ 1. Create directory structure                                           │ │
│ │ 2. Create worktrees:                                                    │ │
│ │    git worktree add -B task/refactor-conversation geppetto              │ │
│ │    git worktree add -B task/refactor-conversation pinocchio             │ │
│ │ 3. Initialize go.work and add modules                                   │ │
│ │ 4. Copy AGENT.md from ~/templates/AGENT.md                             │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ [Execute] [Back] [Cancel]                                                  │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Delete Workspace Confirmation Screen**:
```
┌─ Delete Workspace Confirmation ────────────────────────────────────────────┐
│                                                                             │
│ Workspace: refactor-conversation                                           │
│ Path: ~/workspaces/2025-06-01/refactor-conversation                       │
│ Repositories: 3                                                           │
│                                                                             │
│ This will:                                                                 │
│   1. Remove git worktrees (git worktree remove)                           │
│      ⚠️  Will fail if there are uncommitted changes                       │
│   2. DELETE the workspace directory and ALL its contents!                 │
│                                                                             │
│ Options:                                                                   │
│   [f] Toggle file deletion (currently: ON - will delete files)            │
│   [w] Toggle force worktrees (currently: OFF - safe removal)              │
│   [y] Confirm deletion                                                     │
│   [n] Cancel                                                               │
│   [esc] Cancel                                                             │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Value delivered**: Workspaces become first-class development artifacts that can be managed, shared, and reused across teams and projects.

## Tier 4: Advanced Workflow Integration

### Phase 4.1: Template and Automation System

**What it does**: Enables teams to codify their multi-repository development patterns into reusable templates.

**User workflow**:
1. User creates workspace templates that encode team practices
2. Templates include repository selections, branch patterns, and initialization scripts
3. New workspaces can be created from templates instantly
4. Teams share templates to standardize development environments

**CLI Examples**:
```bash
# Create template from existing workspace
workspace-manager template create feature-template \
  --from-workspace refactor-conversation

# Create workspace from template
workspace-manager create new-feature --template feature-template

# List available templates
workspace-manager template list

# Edit template configuration
workspace-manager template edit feature-template
```

**TUI Examples**:
```
┌─ Template Management ───────────────────────────────────────────────────────┐
│                                                                             │
│ ┌─ Available Templates ───────────────────────────────────────────────────┐ │
│ │ ● feature-template     [geppetto, pinocchio, bobatea]                   │ │
│ │ ● bugfix-template      [geppetto, pinocchio]                            │ │
│ │ ● full-stack-template  [frontend, backend, shared]                      │ │
│ └─────────────────────────────────────────────────────────────────────────┘ │
│                                                                             │
│ [Create] [Edit] [Delete] [Export] [Import]                                 │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Value delivered**: Team knowledge about complex development setups becomes encoded and shareable, reducing onboarding time and ensuring consistency.

### Phase 4.2: Integration with Development Tools

**What it does**: Connects workspace operations with broader development toolchain.

**User workflow**:
1. Workspace creation triggers IDE workspace configuration
2. Commit operations can trigger CI/CD pipelines across repositories
3. Workspace status integrates with project management tools
4. Branch operations coordinate with code review systems

**CLI Examples**:
```bash
# Configure integration settings
workspace-manager config set integrations.ide "vscode"
workspace-manager config set integrations.ci "github-actions"

# Create workspace with IDE integration
workspace-manager create feature-workspace --repos app,lib \
  --configure-ide

# Commit with CI trigger
workspace-manager commit --message "Add new feature" \
  --trigger-ci --create-pr
```

**Value delivered**: The workspace becomes the central coordination point for all development activities, not just git operations.

## Success Criteria

**Tier 1 Success**: A developer can replace manual worktree setup with a single command and immediately have a functional multi-repository workspace. They can also safely clean up workspaces when development is complete, with full control over file preservation.

**Tier 2 Success**: A developer can manage complex feature development across multiple repositories without leaving their workspace context or manually tracking repository states.

**Tier 3 Success**: A developer who is uncomfortable with complex git operations can confidently manage multi-repository workflows through intuitive visual interfaces.

**Tier 4 Success**: Development teams can standardize and share their multi-repository development practices, making complex workflows accessible to all team members.

## Implementation Strategy

Each tier builds upon the previous one and delivers standalone value. Tier 1 alone provides significant productivity improvements. Each subsequent tier adds more sophisticated capabilities while maintaining the core value proposition of simplifying multi-repository development workflows.