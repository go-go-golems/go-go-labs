package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
)

// Manager handles workspace creation operations
type Manager struct {
	gitOps    *GitOperations
	goWorkOps *GoWorkspaceOperations
}

// ProgressCallback is called during workspace creation to report progress
type ProgressCallback func(step, total int, currentTask, logMessage string)

// NewManager creates a new workspace manager
func NewManager() *Manager {
	return &Manager{
		gitOps:    NewGitOperations(),
		goWorkOps: NewGoWorkspaceOperations(),
	}
}

// CreateWorkspace creates a new workspace with the specified repositories
func (m *Manager) CreateWorkspace(ctx context.Context, req *config.WorkspaceRequest, progress ProgressCallback) error {
	totalSteps := len(req.Repositories) + 2 // +2 for directory creation and go.work init
	currentStep := 0

	// Step 1: Create workspace directory
	currentStep++
	progress(currentStep, totalSteps, "Creating workspace directory", 
		fmt.Sprintf("Creating workspace directory: %s", req.Path))
	
	if err := m.createWorkspaceDirectory(req.Path); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Step 2-N: Set up repository worktrees
	for _, repo := range req.Repositories {
		currentStep++
		taskDesc := fmt.Sprintf("Setting up %s worktree", repo.Name)
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		progress(currentStep, totalSteps, taskDesc, 
			fmt.Sprintf("Setting up worktree for repository: %s", repo.Name))

		if err := m.setupRepositoryWorktree(ctx, req.Path, repo, progress); err != nil {
			return fmt.Errorf("failed to setup worktree for %s: %w", repo.Name, err)
		}
	}

	// Final step: Initialize go.work
	currentStep++
	progress(currentStep, totalSteps, "Initializing go.work", 
		"Creating go.work file with selected modules")

	if err := m.goWorkOps.InitializeGoWork(req.Path, req.Repositories); err != nil {
		return fmt.Errorf("failed to initialize go.work: %w", err)
	}

	progress(currentStep, totalSteps, "Completed", 
		"Workspace creation completed successfully")

	return nil
}

func (m *Manager) createWorkspaceDirectory(path string) error {
	// Check if directory already exists
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("directory already exists: %s", path)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check directory: %w", err)
	}

	// Create directory with parent directories
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}

func (m *Manager) setupRepositoryWorktree(ctx context.Context, workspacePath string, repo config.Repository, progress ProgressCallback) error {
	repoPath := filepath.Join(workspacePath, repo.Name)

	// Handle subdirectory case for monorepos
	if repo.Subdirectory != "" {
		// For monorepos, we need to handle this differently
		// This is a simplified implementation
		return m.setupMonorepoWorktree(ctx, workspacePath, repo, progress)
	}

	// Regular repository worktree setup
	if repo.LocalPath != "" {
		// Use existing local repository
		return m.gitOps.CreateWorktreeFromLocal(repo.LocalPath, repoPath, repo.DefaultBranch)
	} else if repo.URL != "" {
		// Clone and create worktree from remote
		return m.gitOps.CreateWorktreeFromRemote(repo.URL, repoPath, repo.DefaultBranch)
	}

	return fmt.Errorf("repository %s has neither local_path nor url", repo.Name)
}

func (m *Manager) setupMonorepoWorktree(ctx context.Context, workspacePath string, repo config.Repository, progress ProgressCallback) error {
	// For monorepos with subdirectories, we need to:
	// 1. Create a worktree of the main repo
	// 2. Symlink or copy the subdirectory to the expected location
	
	// This is a simplified implementation
	// In a real implementation, you might want to use sparse-checkout or other Git features
	
	if repo.LocalPath == "" {
		return fmt.Errorf("monorepo subdirectories require a local_path")
	}

	sourcePath := filepath.Join(repo.LocalPath, repo.Subdirectory)
	targetPath := filepath.Join(workspacePath, repo.Name)

	// Check if source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return fmt.Errorf("subdirectory %s does not exist in %s", repo.Subdirectory, repo.LocalPath)
	}

	// Create a symbolic link to the subdirectory
	return os.Symlink(sourcePath, targetPath)
}