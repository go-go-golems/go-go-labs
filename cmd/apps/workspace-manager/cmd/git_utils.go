package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Shared git utility functions used across multiple components

// analyzeRepository extracts metadata from a git repository
func analyzeRepository(ctx context.Context, path string) (*Repository, error) {
	name := filepath.Base(path)
	
	repo := &Repository{
		Name:        name,
		Path:        path,
		LastUpdated: time.Now(),
		Categories:  categorizeRepository(path),
	}

	// Get remote URL
	if remoteURL, err := getGitRemoteURL(ctx, path); err == nil {
		repo.RemoteURL = remoteURL
	}

	// Get current branch
	if branch, err := getGitCurrentBranch(ctx, path); err == nil {
		repo.CurrentBranch = branch
	}

	// Get all branches
	if branches, err := getGitBranches(ctx, path); err == nil {
		repo.Branches = branches
	}

	// Get tags
	if tags, err := getGitTags(ctx, path); err == nil {
		repo.Tags = tags
	}

	// Get last commit
	if lastCommit, err := getGitLastCommit(ctx, path); err == nil {
		repo.LastCommit = lastCommit
	}

	return repo, nil
}

// categorizeRepository determines categories based on repository content
func categorizeRepository(path string) []string {
	var categories []string

	// Check for common language/framework files
	files := map[string]string{
		"go.mod":           "go",
		"package.json":     "node",
		"Cargo.toml":       "rust",
		"setup.py":         "python",
		"requirements.txt": "python",
		"Gemfile":          "ruby",
		"pom.xml":          "java",
		"build.gradle":     "gradle",
		"Makefile":         "make",
		"docker-compose.yml": "docker",
		"Dockerfile":       "docker",
	}

	for file, category := range files {
		if _, err := os.Stat(filepath.Join(path, file)); err == nil {
			categories = append(categories, category)
		}
	}

	// Check for common project types
	dirs := map[string]string{
		"cmd":    "cli",
		"web":    "web",
		"mobile": "mobile",
		"tui":    "tui",
		"api":    "api",
		"server": "server",
		"client": "client",
	}

	for dir, category := range dirs {
		if stat, err := os.Stat(filepath.Join(path, dir)); err == nil && stat.IsDir() {
			categories = append(categories, category)
		}
	}

	if len(categories) == 0 {
		categories = append(categories, "unknown")
	}

	return categories
}

// Git command helpers
func getGitRemoteURL(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getGitCurrentBranch(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getGitBranches(ctx context.Context, path string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "-a")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var branches []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove current branch marker and remote prefixes
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimPrefix(line, "remotes/origin/")
		if !strings.Contains(line, "HEAD ->") {
			branches = append(branches, line)
		}
	}
	
	return branches, nil
}

func getGitTags(ctx context.Context, path string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "tag", "-l")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var tags []string
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			tags = append(tags, line)
		}
	}
	
	return tags, nil
}

func getGitLastCommit(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", "-1", "--pretty=format:%H %s")
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// mergeRepositories merges existing repositories with newly discovered ones
func mergeRepositories(existing, discovered []Repository) []Repository {
	repoMap := make(map[string]Repository)
	
	// Add existing repositories
	for _, repo := range existing {
		repoMap[repo.Path] = repo
	}
	
	// Update with discovered repositories
	for _, repo := range discovered {
		repoMap[repo.Path] = repo
	}
	
	// Convert back to slice
	var result []Repository
	for _, repo := range repoMap {
		result = append(result, repo)
	}
	
	return result
}
