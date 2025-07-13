package workspace

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitOperations handles Git-related operations for workspace setup
type GitOperations struct{}

// NewGitOperations creates a new GitOperations instance
func NewGitOperations() *GitOperations {
	return &GitOperations{}
}

// CreateWorktreeFromLocal creates a worktree from an existing local repository
func (g *GitOperations) CreateWorktreeFromLocal(localPath, targetPath, branch string) error {
	// Verify the local path is a Git repository
	if !g.isGitRepository(localPath) {
		return fmt.Errorf("path is not a Git repository: %s", localPath)
	}

	// Get the absolute paths
	absLocalPath, err := filepath.Abs(localPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", localPath, err)
	}

	absTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for %s: %w", targetPath, err)
	}

	// Ensure target directory doesn't exist
	if _, err := os.Stat(absTargetPath); err == nil {
		return fmt.Errorf("target path already exists: %s", absTargetPath)
	}

	// Create worktree
	cmd := exec.Command("git", "worktree", "add", absTargetPath, branch)
	cmd.Dir = absLocalPath

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// CreateWorktreeFromRemote clones a repository and creates a worktree
func (g *GitOperations) CreateWorktreeFromRemote(url, targetPath, branch string) error {
	// For remote repositories, we need to first clone to a bare repository
	// then create a worktree from it

	// Create a temporary bare repository directory
	parentDir := filepath.Dir(targetPath)
	repoName := filepath.Base(targetPath)
	bareRepoPath := filepath.Join(parentDir, fmt.Sprintf(".%s-bare", repoName))

	// Clone as bare repository
	cmd := exec.Command("git", "clone", "--bare", url, bareRepoPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to clone repository: %w\nOutput: %s", err, string(output))
	}

	// Create worktree from bare repository
	cmd = exec.Command("git", "worktree", "add", targetPath, branch)
	cmd.Dir = bareRepoPath

	output, err = cmd.CombinedOutput()
	if err != nil {
		// Clean up bare repository on failure
		os.RemoveAll(bareRepoPath)
		return fmt.Errorf("failed to create worktree from remote: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// isGitRepository checks if a directory is a Git repository
func (g *GitOperations) isGitRepository(path string) bool {
	// Check if .git directory or file exists
	gitPath := filepath.Join(path, ".git")
	if _, err := os.Stat(gitPath); err == nil {
		return true
	}

	// Also check by running git command
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Dir = path
	err := cmd.Run()
	return err == nil
}

// GetBranches returns a list of available branches for a repository
func (g *GitOperations) GetBranches(repoPath string) ([]string, error) {
	if !g.isGitRepository(repoPath) {
		return nil, fmt.Errorf("not a Git repository: %s", repoPath)
	}

	cmd := exec.Command("git", "branch", "-r", "--format=%(refname:short)")
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")

	// Clean up branch names (remove origin/ prefix)
	var cleanBranches []string
	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		if strings.HasPrefix(branch, "origin/") {
			branch = strings.TrimPrefix(branch, "origin/")
		}
		if branch != "" && branch != "HEAD" {
			cleanBranches = append(cleanBranches, branch)
		}
	}

	return cleanBranches, nil
}

// ValidateBranch checks if a branch exists in the repository
func (g *GitOperations) ValidateBranch(repoPath, branch string) error {
	branches, err := g.GetBranches(repoPath)
	if err != nil {
		return err
	}

	for _, b := range branches {
		if b == branch {
			return nil
		}
	}

	return fmt.Errorf("branch %s not found in repository", branch)
}

// CleanupWorktree removes a worktree and cleans up the repository
func (g *GitOperations) CleanupWorktree(worktreePath string) error {
	// Get the parent repository path
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = worktreePath

	output, err := cmd.Output()
	if err != nil {
		// If we can't find the repository, just remove the directory
		return os.RemoveAll(worktreePath)
	}

	// Remove the worktree using Git command
	parentRepo := strings.TrimSpace(string(output))
	cmd = exec.Command("git", "worktree", "remove", worktreePath)
	cmd.Dir = parentRepo

	if err := cmd.Run(); err != nil {
		// Fallback to force removal
		cmd = exec.Command("git", "worktree", "remove", "--force", worktreePath)
		cmd.Dir = parentRepo
		if err := cmd.Run(); err != nil {
			// Final fallback: just remove the directory
			return os.RemoveAll(worktreePath)
		}
	}

	return nil
}
