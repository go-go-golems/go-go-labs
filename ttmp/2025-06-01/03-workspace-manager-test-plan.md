# Workspace Manager Test Plan

## Overview
This document outlines unit and integration tests that can be implemented without significant source code modifications for mocking. The tests focus on real git operations with controlled test repositories.

## Test Strategy

### Philosophy
- Use real git repositories in isolated test environments
- Minimal mocking - only for external dependencies (network, filesystem permissions)
- Integration tests that exercise complete workflows
- Deterministic test data setup and teardown

### Test Environment Setup
```go
// testutils/setup.go
type TestRepo struct {
    Name     string
    Path     string
    Remote   string
    HasGo    bool
    Branches []string
}

type TestWorkspace struct {
    Name  string
    Path  string
    Repos []TestRepo
}

func SetupTestRepos(t *testing.T, repos []TestRepo) string
func CreateTestWorkspace(t *testing.T, ws TestWorkspace) string
func CleanupTestEnv(t *testing.T, basePath string)
```

## Unit Tests

### 1. Repository Discovery Tests

#### File: `discovery_test.go`

```go
func TestRepositoryDiscoverer_LoadRegistry(t *testing.T)
func TestRepositoryDiscoverer_SaveRegistry(t *testing.T)
func TestRepositoryDiscoverer_DiscoverRepositories(t *testing.T)
func TestRepositoryDiscoverer_scanDirectory(t *testing.T)
func TestRepositoryDiscoverer_isGitRepository(t *testing.T)
func TestRepositoryDiscoverer_analyzeRepository(t *testing.T)
func TestRepositoryDiscoverer_categorizeRepository(t *testing.T)
func TestRepositoryDiscoverer_getGitRemoteURL(t *testing.T)
func TestRepositoryDiscoverer_getGitCurrentBranch(t *testing.T)
func TestRepositoryDiscoverer_mergeRepositories(t *testing.T)
```

**Test Cases:**
- Discovery with valid git repositories
- Discovery with non-git directories
- Registry persistence and loading
- Repository categorization based on file content
- Recursive discovery with depth limits
- Discovery with corrupted git repositories
- Registry merging with existing entries

#### File: `discovery_integration_test.go`

```go
func TestDiscovery_RealRepositories(t *testing.T) {
    // Setup test repos with different characteristics
    testRepos := []TestRepo{
        {Name: "go-repo", HasGo: true, Branches: []string{"main", "develop"}},
        {Name: "node-repo", HasGo: false, Branches: []string{"master"}},
        {Name: "empty-repo", HasGo: false, Branches: []string{}},
    }
    
    basePath := SetupTestRepos(t, testRepos)
    defer CleanupTestEnv(t, basePath)
    
    // Test discovery
    // Validate results
}

func TestDiscovery_DeepNesting(t *testing.T)
func TestDiscovery_LargeNumberOfRepos(t *testing.T)
func TestDiscovery_PermissionErrors(t *testing.T)
```

### 2. Workspace Management Tests

#### File: `workspace_test.go`

```go
func TestWorkspaceManager_CreateWorkspace(t *testing.T)
func TestWorkspaceManager_findRepositories(t *testing.T)
func TestWorkspaceManager_shouldCreateGoWorkspace(t *testing.T)
func TestWorkspaceManager_createWorkspaceStructure(t *testing.T)
func TestWorkspaceManager_createWorktree(t *testing.T)
func TestWorkspaceManager_createGoWorkspace(t *testing.T)
func TestWorkspaceManager_copyAgentMD(t *testing.T)
func TestWorkspaceManager_saveWorkspace(t *testing.T)
```

**Test Cases:**
- Workspace creation with multiple repositories
- Worktree creation on existing and new branches
- Go workspace file generation
- AGENT.md template copying
- Workspace configuration persistence
- Error handling for missing repositories
- Dry-run mode validation

#### File: `workspace_integration_test.go`

```go
func TestWorkspace_CompleteCreationFlow(t *testing.T) {
    // Setup repositories
    // Create workspace
    // Validate all worktrees exist
    // Validate go.work file
    // Validate workspace config
}

func TestWorkspace_BranchCreation(t *testing.T)
func TestWorkspace_GoWorkspaceIntegration(t *testing.T)
func TestWorkspace_AgentMDCopying(t *testing.T)
```

### 3. Status Operations Tests

#### File: `status_test.go`

```go
func TestStatusChecker_GetWorkspaceStatus(t *testing.T)
func TestStatusChecker_getRepositoryStatus(t *testing.T)
func TestStatusChecker_getCurrentBranch(t *testing.T)
func TestStatusChecker_getModifiedFiles(t *testing.T)
func TestStatusChecker_getStagedFiles(t *testing.T)
func TestStatusChecker_getUntrackedFiles(t *testing.T)
func TestStatusChecker_getAheadBehind(t *testing.T)
func TestStatusChecker_hasConflicts(t *testing.T)
func TestStatusChecker_calculateOverallStatus(t *testing.T)
```

**Test Cases:**
- Status with clean repositories
- Status with various types of changes
- Status with staged and unstaged changes
- Status with untracked files
- Ahead/behind calculation with remote tracking
- Conflict detection
- Overall status calculation logic

### 4. Git Operations Tests

#### File: `git_operations_test.go`

```go
func TestGitOperations_GetWorkspaceChanges(t *testing.T)
func TestGitOperations_getRepositoryChanges(t *testing.T)
func TestGitOperations_StageFile(t *testing.T)
func TestGitOperations_UnstageFile(t *testing.T)
func TestGitOperations_CommitChanges(t *testing.T)
func TestGitOperations_previewCommit(t *testing.T)
func TestGitOperations_stageAllFiles(t *testing.T)
func TestGitOperations_hasStagedChanges(t *testing.T)
func TestGitOperations_GetDiff(t *testing.T)
func TestGitOperations_getRepositoryDiff(t *testing.T)
```

**Test Cases:**
- Change detection across repositories
- File staging and unstaging
- Commit operations with various options
- Diff generation for staged and unstaged changes
- Dry-run commit previews
- Error handling for git command failures

### 5. Sync Operations Tests

#### File: `sync_operations_test.go`

```go
func TestSyncOperations_SyncWorkspace(t *testing.T)
func TestSyncOperations_syncRepository(t *testing.T)
func TestSyncOperations_pullRepository(t *testing.T)
func TestSyncOperations_pushRepository(t *testing.T)
func TestSyncOperations_getAheadBehind(t *testing.T)
func TestSyncOperations_hasConflicts(t *testing.T)
func TestSyncOperations_CreateBranch(t *testing.T)
func TestSyncOperations_SwitchBranch(t *testing.T)
func TestSyncOperations_GetWorkspaceLog(t *testing.T)
```

**Test Cases:**
- Sync operations with pull and push
- Branch creation across repositories
- Branch switching validation
- Conflict detection and handling
- Log aggregation across repositories
- Error handling for network issues (mocked)

## Integration Tests

### 1. End-to-End Workflow Tests

#### File: `e2e_test.go`

```go
func TestE2E_DiscoveryToWorkspaceCreation(t *testing.T) {
    // 1. Setup test repositories
    // 2. Run discovery
    // 3. Create workspace
    // 4. Validate workspace structure
    // 5. Test status operations
}

func TestE2E_CompleteGitWorkflow(t *testing.T) {
    // 1. Create workspace
    // 2. Make changes in multiple repos
    // 3. Stage and commit changes
    // 4. Create branches
    // 5. Sync operations
    // 6. Validate all operations
}

func TestE2E_ConflictResolutionWorkflow(t *testing.T) {
    // 1. Setup repositories with remotes
    // 2. Create conflicting changes
    // 3. Test sync behavior
    // 4. Validate conflict reporting
}

func TestE2E_TUIWorkflow(t *testing.T) {
    // Test TUI operations programmatically
    // Using tea.ProgramTest for bubble tea testing
}
```

### 2. Command Integration Tests

#### File: `cmd_integration_test.go`

```go
func TestCmdIntegration_Discover(t *testing.T)
func TestCmdIntegration_Create(t *testing.T)
func TestCmdIntegration_Status(t *testing.T)
func TestCmdIntegration_Commit(t *testing.T)
func TestCmdIntegration_Sync(t *testing.T)
func TestCmdIntegration_Branch(t *testing.T)
func TestCmdIntegration_Diff(t *testing.T)
func TestCmdIntegration_Log(t *testing.T)
```

**Test Strategy:**
- Execute actual CLI commands using `os/exec`
- Parse command output and validate results
- Test error conditions and edge cases
- Validate file system state after operations

### 3. Configuration and Persistence Tests

#### File: `config_test.go`

```go
func TestConfig_RegistryPersistence(t *testing.T)
func TestConfig_WorkspacePersistence(t *testing.T)
func TestConfig_ConfigDirectoryHandling(t *testing.T)
func TestConfig_RegistryMigration(t *testing.T)
```

## Performance Tests

### File: `performance_test.go`

```go
func BenchmarkDiscovery_100Repos(b *testing.B)
func BenchmarkWorkspaceCreation_10Repos(b *testing.B)
func BenchmarkStatusCheck_LargeWorkspace(b *testing.B)
func BenchmarkCommit_MultipleRepos(b *testing.B)

func TestPerformance_LargeRepositorySet(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping performance test in short mode")
    }
    // Test with 100+ repositories
}

func TestPerformance_DeepDirectoryStructure(t *testing.T)
func TestPerformance_LargeGitHistory(t *testing.T)
```

## Error Handling Tests

### File: `error_handling_test.go`

```go
func TestErrorHandling_NetworkFailures(t *testing.T)
func TestErrorHandling_PermissionDenied(t *testing.T)
func TestErrorHandling_CorruptedRepository(t *testing.T)
func TestErrorHandling_InvalidGitOperations(t *testing.T)
func TestErrorHandling_ConcurrentOperations(t *testing.T)
```

## Test Utilities and Helpers

### File: `testutils/git_helpers.go`

```go
func CreateTestRepo(t *testing.T, path, name string) string
func AddCommitToRepo(t *testing.T, repoPath, message string)
func CreateBranchInRepo(t *testing.T, repoPath, branchName string)
func CreateRemoteRepo(t *testing.T, localPath, remotePath string)
func CreateConflictingChanges(t *testing.T, repo1, repo2 string)
func ValidateWorktree(t *testing.T, workspacePath, repoName string)
func ValidateGoWorkspace(t *testing.T, workspacePath string)
```

### File: `testutils/fs_helpers.go`

```go
func CreateTempDir(t *testing.T) string
func CopyFile(t *testing.T, src, dst string)
func CreateFileWithContent(t *testing.T, path, content string)
func ValidateFileExists(t *testing.T, path string)
func ValidateDirectoryStructure(t *testing.T, basePath string, expected []string)
```

### File: `testutils/cli_helpers.go`

```go
func RunCommand(t *testing.T, args ...string) (stdout, stderr string, exitCode int)
func RunCommandInDir(t *testing.T, dir string, args ...string) (stdout, stderr string, exitCode int)
func ParseTableOutput(t *testing.T, output string) []map[string]string
func ValidateJSONOutput(t *testing.T, output string, target interface{})
```

## Test Data Management

### File: `testdata/repositories.json`
```json
{
  "scenarios": {
    "basic": [
      {"name": "simple-go", "type": "go", "files": ["main.go", "go.mod"]},
      {"name": "node-app", "type": "node", "files": ["package.json", "index.js"]}
    ],
    "complex": [
      {"name": "monorepo", "type": "mixed", "subdirs": ["api", "web", "shared"]},
      {"name": "legacy", "type": "unknown", "files": ["Makefile", "src/"]}
    ]
  }
}
```

### File: `testdata/workspaces.json`
```json
{
  "templates": {
    "fullstack": {
      "repos": ["frontend", "backend", "shared"],
      "branch": "develop"
    },
    "microservices": {
      "repos": ["auth-service", "user-service", "api-gateway"],
      "branch": "main"
    }
  }
}
```

## Continuous Integration Setup

### File: `.github/workflows/test.yml`
```yaml
name: Test Suite
on: [push, pull_request]
jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: [1.23, 1.24]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: ${{ matrix.go-version }}
      - name: Setup Git
        run: |
          git config --global user.name "Test User"
          git config --global user.email "test@example.com"
      - name: Run Unit Tests
        run: go test -v ./...
      - name: Run Integration Tests
        run: go test -v -tags=integration ./...
      - name: Run Performance Tests
        run: go test -v -bench=. -benchmem ./...
```

## Test Execution Strategy

### Local Development
```bash
# Quick unit tests
go test -short ./...

# Full unit test suite
go test ./...

# Integration tests only
go test -tags=integration ./...

# Performance tests
go test -bench=. -benchmem ./...

# Coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### CI/CD Pipeline
1. **Fast feedback**: Unit tests on every commit
2. **Integration tests**: On pull requests
3. **Performance tests**: Nightly builds
4. **Cross-platform tests**: Before releases

## Test Maintenance

### Guidelines
- Keep test data small and focused
- Use table-driven tests for multiple scenarios
- Clean up test resources in `defer` statements
- Use meaningful test names that describe scenarios
- Group related tests in subtests using `t.Run()`

### Review Checklist
- [ ] Tests are deterministic and repeatable
- [ ] Tests clean up after themselves
- [ ] Error cases are covered
- [ ] Performance tests have reasonable thresholds
- [ ] Tests work across platforms
- [ ] Test data is minimal and focused

This test plan provides comprehensive coverage without requiring extensive mocking infrastructure, making it practical to implement and maintain.
