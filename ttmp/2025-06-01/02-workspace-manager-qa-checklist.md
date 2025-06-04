# Workspace Manager QA Checklist

## Overview
This document provides a comprehensive testing checklist for the workspace-manager tool. Each section includes specific test cases, expected behaviors, and edge cases to validate.

## Prerequisites
- Go 1.23+ installed
- Git configured with user.name and user.email
- Multiple git repositories available for testing
- SSH keys configured for git operations (if testing with remote repos)

## Test Environment Setup

### Test Repository Structure
```bash
# Create test repositories
mkdir -p ~/qa-test/{repo1,repo2,repo3}
cd ~/qa-test/repo1 && git init && echo "repo1" > README.md && git add . && git commit -m "initial"
cd ~/qa-test/repo2 && git init && echo "repo2" > README.md && git add . && git commit -m "initial"
cd ~/qa-test/repo3 && git init && echo "repo3" > README.md && git add . && git commit -m "initial"
```

## Phase 1: Foundation Features

### 1.1 Repository Discovery (`discover` command)

#### Basic Discovery
- [ ] **Test**: `workspace-manager discover ~/qa-test`
  - **Expected**: Finds all 3 test repositories
  - **Validate**: Registry file created at `~/.config/workspace-manager/registry.json`
  - **Check**: Repository metadata includes path, name, categories

#### Recursive Discovery
- [ ] **Test**: `workspace-manager discover ~/qa-test --recursive --max-depth 2`
  - **Expected**: Finds repositories at specified depth
  - **Edge Case**: Test with `--max-depth 0` (should only check current directory)

#### Discovery with No Repositories
- [ ] **Test**: `workspace-manager discover /tmp`
  - **Expected**: Completes without error, reports 0 repositories found

#### Discovery Error Cases
- [ ] **Test**: `workspace-manager discover /nonexistent/path`
  - **Expected**: Clear error message about path not existing
- [ ] **Test**: Discovery with insufficient permissions
  - **Expected**: Graceful handling with warning messages

#### Registry Persistence
- [ ] **Test**: Run discovery twice on same directories
  - **Expected**: Second run updates existing entries, maintains consistency
- [ ] **Test**: Check registry file format
  - **Expected**: Valid JSON with all required fields

### 1.2 Repository Listing (`list` command)

#### List All Repositories
- [ ] **Test**: `workspace-manager list repos`
  - **Expected**: Table format with all discovered repositories
  - **Validate**: Columns: NAME, PATH, BRANCH, TAGS, REMOTE

#### List with Filtering
- [ ] **Test**: `workspace-manager list repos --tags go`
  - **Expected**: Only repositories with 'go' tag shown
- [ ] **Test**: `workspace-manager list repos --tags nonexistent`
  - **Expected**: "No repositories found" message

#### JSON Output
- [ ] **Test**: `workspace-manager list repos --format json`
  - **Expected**: Valid JSON output with all repository data
- [ ] **Test**: Pipe JSON output to `jq` for validation

#### Empty Registry
- [ ] **Test**: Delete registry file, then `list repos`
  - **Expected**: Clear message suggesting to run discovery first

### 1.3 Workspace Creation (`create` command)

#### Basic Workspace Creation
- [ ] **Test**: `workspace-manager create test-workspace --repos repo1,repo2`
  - **Expected**: Workspace directory created with worktrees
  - **Validate**: Both repositories available as worktrees
  - **Check**: Original repositories remain unchanged

#### Workspace with Branch
- [ ] **Test**: `workspace-manager create test-branch-workspace --repos repo1 --branch feature/test`
  - **Expected**: Worktree created on specified branch
  - **Validate**: Branch created in worktree repository

#### Go Workspace Creation
- [ ] **Test**: Create workspace with Go repositories
  - **Expected**: `go.work` file created and configured
  - **Validate**: All Go modules listed in go.work

#### AGENT.md Integration
- [ ] **Test**: Create AGENT.md template file, use with `--agent-source`
  - **Expected**: AGENT.md copied to workspace root

#### Dry Run
- [ ] **Test**: `workspace-manager create test-dry --repos repo1 --dry-run`
  - **Expected**: Shows preview without creating anything
  - **Validate**: No directories or files created

#### Error Cases
- [ ] **Test**: Create workspace with non-existent repository
  - **Expected**: Clear error about repository not found
- [ ] **Test**: Create workspace in directory without permissions
  - **Expected**: Appropriate permission error
- [ ] **Test**: Create workspace with empty repository list
  - **Expected**: Error about no repositories specified

#### Interactive Mode
- [ ] **Test**: `workspace-manager create test-interactive --interactive`
  - **Expected**: Shows repository selection interface
  - **Validate**: Can select repositories by number or name

### 1.4 Workspace Status (`status` command)

#### Current Workspace Detection
- [ ] **Test**: `cd <workspace>` then `workspace-manager status`
  - **Expected**: Detects current workspace automatically
- [ ] **Test**: Run status from non-workspace directory
  - **Expected**: Clear error about not being in workspace

#### Status Display
- [ ] **Test**: Status with clean repositories
  - **Expected**: All repositories show clean status
- [ ] **Test**: Make changes in workspace repositories, check status
  - **Expected**: Shows modified files, staged changes, etc.

#### Short Format
- [ ] **Test**: `workspace-manager status --short`
  - **Expected**: Condensed one-line-per-repo format

#### Untracked Files
- [ ] **Test**: Create untracked files, use `--untracked` flag
  - **Expected**: Untracked files shown in status

#### Named Workspace
- [ ] **Test**: `workspace-manager status specific-workspace-name`
  - **Expected**: Shows status for named workspace

## Phase 2: Advanced Git Operations

### 2.1 Cross-Repository Commits (`commit` command)

#### Basic Commit
- [ ] **Test**: Make changes in multiple repos, `workspace-manager commit -m "test message"`
  - **Expected**: All changes committed with same message
  - **Validate**: Commit history shows consistent messages

#### Interactive Commit
- [ ] **Test**: `workspace-manager commit --interactive`
  - **Expected**: Shows all changes, allows file selection
  - **Validate**: Only selected files committed

#### Add All
- [ ] **Test**: `workspace-manager commit --add-all -m "add all changes"`
  - **Expected**: All unstaged changes staged and committed

#### Commit with Push
- [ ] **Test**: `workspace-manager commit -m "test" --push`
  - **Expected**: Changes committed and pushed to remotes
  - **Note**: Requires remote repositories

#### Dry Run Commit
- [ ] **Test**: `workspace-manager commit --dry-run --interactive`
  - **Expected**: Shows what would be committed without changes

#### Template Usage
- [ ] **Test**: `workspace-manager commit --template feature`
  - **Expected**: Uses predefined commit message template

#### Error Cases
- [ ] **Test**: Commit with no changes
  - **Expected**: Appropriate message about no changes
- [ ] **Test**: Commit without message in non-interactive mode
  - **Expected**: Error requiring commit message
- [ ] **Test**: Commit from non-workspace directory
  - **Expected**: Error about not being in workspace

### 2.2 Synchronized Operations (`sync` command)

#### Sync All
- [ ] **Test**: `workspace-manager sync all`
  - **Expected**: All repos pulled and pushed
  - **Validate**: Status shows synchronized state

#### Pull Only
- [ ] **Test**: `workspace-manager sync pull`
  - **Expected**: Only pull operations performed
- [ ] **Test**: `workspace-manager sync pull --rebase`
  - **Expected**: Uses rebase instead of merge

#### Push Only
- [ ] **Test**: `workspace-manager sync push`
  - **Expected**: Only push operations performed

#### Conflict Handling
- [ ] **Test**: Create merge conflicts, run sync
  - **Expected**: Reports conflicts, provides guidance

#### Ahead/Behind Tracking
- [ ] **Test**: Create commits locally and remotely, run sync
  - **Expected**: Shows before/after ahead/behind counts

#### Repositories Without Remotes
- [ ] **Test**: Sync workspace with local-only repositories
  - **Expected**: Graceful handling, clear status reporting

#### Network Errors
- [ ] **Test**: Sync with unreachable remote
  - **Expected**: Appropriate error messages, continues with other repos

### 2.3 Branch Management (`branch` command)

#### Create Branch Across Repos
- [ ] **Test**: `workspace-manager branch create feature/new-feature`
  - **Expected**: Branch created in all repositories
  - **Validate**: All repos switched to new branch

#### Switch Branch
- [ ] **Test**: `workspace-manager branch switch main`
  - **Expected**: All repositories switched to main branch

#### Branch with Tracking
- [ ] **Test**: `workspace-manager branch create feature/tracked --track`
  - **Expected**: Branch set up with upstream tracking

#### List Current Branches
- [ ] **Test**: `workspace-manager branch list`
  - **Expected**: Shows current branch for each repository

#### Error Cases
- [ ] **Test**: Switch to non-existent branch
  - **Expected**: Clear error messages for failing repositories
- [ ] **Test**: Create branch that already exists
  - **Expected**: Appropriate error handling

### 2.4 Diff and Log Operations

#### Workspace Diff
- [ ] **Test**: Make changes, `workspace-manager diff`
  - **Expected**: Unified diff across all repositories
- [ ] **Test**: `workspace-manager diff --staged`
  - **Expected**: Shows only staged changes

#### Repository-Specific Diff
- [ ] **Test**: `workspace-manager diff --repo repo1`
  - **Expected**: Shows diff only for specified repository

#### Workspace Log
- [ ] **Test**: `workspace-manager log`
  - **Expected**: Commit history from all repositories
- [ ] **Test**: `workspace-manager log --since "1 week ago" --oneline`
  - **Expected**: Filtered, concise log output

#### Empty Results
- [ ] **Test**: Diff with no changes
  - **Expected**: "No changes found" message
- [ ] **Test**: Log with empty repositories
  - **Expected**: Graceful handling of repos without commits

## TUI Testing

### 3.1 TUI Launch and Navigation

#### Basic Launch
- [ ] **Test**: `workspace-manager tui`
  - **Expected**: TUI launches with main menu
  - **Validate**: Can navigate with keyboard
  - **Check**: Help screen accessible with '?'

#### Repository Browser
- [ ] **Test**: Navigate to repository browser
  - **Expected**: Lists all discovered repositories
  - **Validate**: Can filter and select repositories
  - **Check**: Space key toggles selection

#### Workspace Management
- [ ] **Test**: Navigate to workspace manager
  - **Expected**: Shows created workspaces
  - **Validate**: Can view workspace details

#### Workspace Creation Flow
- [ ] **Test**: Complete workspace creation through TUI
  - **Expected**: Full workflow from selection to creation
  - **Validate**: Form inputs work correctly

### 3.2 TUI Error Handling

#### No Repositories
- [ ] **Test**: Launch TUI with empty registry
  - **Expected**: Clear message about running discovery

#### Terminal Resize
- [ ] **Test**: Resize terminal while TUI running
  - **Expected**: Interface adapts appropriately

#### Invalid Input
- [ ] **Test**: Enter invalid data in forms
  - **Expected**: Validation messages shown

## Cross-Platform Testing

### 4.1 Path Handling
- [ ] **Test**: Workspace creation with various path types
  - **Expected**: Handles absolute/relative paths correctly
- [ ] **Test**: Paths with spaces and special characters
  - **Expected**: Proper escaping and handling

### 4.2 Git Command Compatibility
- [ ] **Test**: All git operations with different git versions
  - **Expected**: Compatible with git 2.20+

### 4.3 Permissions
- [ ] **Test**: Operations with read-only directories
  - **Expected**: Appropriate error handling
- [ ] **Test**: Registry creation in various user directories
  - **Expected**: Respects XDG directories

## Performance Testing

### 5.1 Large Repository Sets
- [ ] **Test**: Discovery with 100+ repositories
  - **Expected**: Reasonable performance, progress indication
- [ ] **Test**: Workspace operations with 10+ repositories
  - **Expected**: Parallel operations where possible

### 5.2 Large Repository Sizes
- [ ] **Test**: Operations on repositories with large histories
  - **Expected**: Timeouts handled gracefully

## Security Testing

### 6.1 Path Injection
- [ ] **Test**: Repository paths with malicious content
  - **Expected**: Proper sanitization and validation

### 6.2 Command Injection
- [ ] **Test**: Repository names with shell metacharacters
  - **Expected**: Proper escaping in git commands

## Cleanup and Validation

### Final Validation Steps
- [ ] **Verify**: No leftover processes after operations
- [ ] **Check**: All temporary files cleaned up
- [ ] **Validate**: Registry file remains consistent
- [ ] **Confirm**: Original repositories unchanged after workspace operations

### Cleanup Commands
```bash
# Clean up test data
rm -rf ~/qa-test
rm -rf ~/.config/workspace-manager
rm -rf ~/workspaces/2025-06-01/test-*
```

## Bug Reporting Template

When issues are found, report with:
- Command run
- Expected behavior
- Actual behavior
- Environment details (OS, Git version, Go version)
- Error messages
- Steps to reproduce

## Performance Benchmarks

Expected performance targets:
- Discovery: <5 seconds for 100 repositories
- Workspace creation: <10 seconds for 5 repositories
- Status check: <3 seconds for 10 repositories
- Sync operations: Network dependent

## Regression Testing

After any code changes, run:
1. Core discovery and listing functionality
2. Basic workspace creation and status
3. One complete sync cycle
4. TUI basic navigation

This ensures no core functionality is broken by changes.
