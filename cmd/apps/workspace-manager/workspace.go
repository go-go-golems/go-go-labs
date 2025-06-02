package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// WorkspaceManager handles workspace creation and management
type WorkspaceManager struct {
	config       *WorkspaceConfig
	discoverer   *RepositoryDiscoverer
	workspaceDir string
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager() (*WorkspaceManager, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to load config")
	}

	registryPath, err := getRegistryPath()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get registry path")
	}

	discoverer := NewRepositoryDiscoverer(registryPath)
	if err := discoverer.LoadRegistry(); err != nil {
		return nil, errors.Wrap(err, "failed to load registry")
	}

	return &WorkspaceManager{
		config:       config,
		discoverer:   discoverer,
		workspaceDir: config.WorkspaceDir,
	}, nil
}

// CreateWorkspace creates a new multi-repository workspace
func (wm *WorkspaceManager) CreateWorkspace(ctx context.Context, name string, repoNames []string, branch string, agentSource string, dryRun bool) (*Workspace, error) {
	// Validate input
	if name == "" {
		return nil, errors.New("workspace name is required")
	}

	// Find repositories
	repos, err := wm.findRepositories(repoNames)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find repositories")
	}

	// Create workspace directory path
	workspacePath := filepath.Join(wm.workspaceDir, name)

	workspace := &Workspace{
		Name:         name,
		Path:         workspacePath,
		Repositories: repos,
		Branch:       branch,
		Created:      time.Now(),
		GoWorkspace:  wm.shouldCreateGoWorkspace(repos),
		AgentMD:      agentSource,
	}

	if dryRun {
		return workspace, nil
	}

	// Create workspace
	if err := wm.createWorkspaceStructure(ctx, workspace); err != nil {
		return nil, errors.Wrap(err, "failed to create workspace structure")
	}

	// Save workspace configuration
	if err := wm.saveWorkspace(workspace); err != nil {
		return nil, errors.Wrap(err, "failed to save workspace configuration")
	}

	return workspace, nil
}

// findRepositories finds repositories by name
func (wm *WorkspaceManager) findRepositories(repoNames []string) ([]Repository, error) {
	allRepos := wm.discoverer.GetRepositories()
	repoMap := make(map[string]Repository)
	
	for _, repo := range allRepos {
		repoMap[repo.Name] = repo
	}

	var repos []Repository
	var notFound []string

	for _, name := range repoNames {
		if repo, exists := repoMap[name]; exists {
			repos = append(repos, repo)
		} else {
			notFound = append(notFound, name)
		}
	}

	if len(notFound) > 0 {
		return nil, errors.Errorf("repositories not found: %s", strings.Join(notFound, ", "))
	}

	return repos, nil
}

// shouldCreateGoWorkspace determines if go.work should be created
func (wm *WorkspaceManager) shouldCreateGoWorkspace(repos []Repository) bool {
	for _, repo := range repos {
		for _, category := range repo.Categories {
			if category == "go" {
				return true
			}
		}
	}
	return false
}

// createWorkspaceStructure creates the physical workspace structure
func (wm *WorkspaceManager) createWorkspaceStructure(ctx context.Context, workspace *Workspace) error {
	log.Info().Str("workspace", workspace.Name).Msg("Creating workspace structure")

	// Create workspace directory
	if err := os.MkdirAll(workspace.Path, 0755); err != nil {
		return errors.Wrapf(err, "failed to create workspace directory: %s", workspace.Path)
	}

	// Create worktrees for each repository
	for _, repo := range workspace.Repositories {
		if err := wm.createWorktree(ctx, workspace, repo); err != nil {
			return errors.Wrapf(err, "failed to create worktree for %s", repo.Name)
		}
	}

	// Create go.work file if needed
	if workspace.GoWorkspace {
		if err := wm.createGoWorkspace(workspace); err != nil {
			return errors.Wrap(err, "failed to create go.work file")
		}
	}

	// Copy AGENT.md if specified
	if workspace.AgentMD != "" {
		if err := wm.copyAgentMD(workspace); err != nil {
			return errors.Wrap(err, "failed to copy AGENT.md")
		}
	}

	return nil
}

// createWorktree creates a git worktree for a repository
func (wm *WorkspaceManager) createWorktree(ctx context.Context, workspace *Workspace, repo Repository) error {
	targetPath := filepath.Join(workspace.Path, repo.Name)
	
	log.Info().
		Str("repo", repo.Name).
		Str("branch", workspace.Branch).
		Str("target", targetPath).
		Msg("Creating worktree")

	// Create worktree with specific branch
	var cmd *exec.Cmd
	if workspace.Branch != "" {
		cmd = exec.CommandContext(ctx, "git", "worktree", "add", "-B", workspace.Branch, targetPath, "origin/"+workspace.Branch)
	} else {
		cmd = exec.CommandContext(ctx, "git", "worktree", "add", targetPath)
	}
	
	cmd.Dir = repo.Path
	if output, err := cmd.CombinedOutput(); err != nil {
		// If branch doesn't exist remotely, create it locally
		if strings.Contains(string(output), "invalid reference") && workspace.Branch != "" {
			log.Info().Str("branch", workspace.Branch).Msg("Branch not found remotely, creating locally")
			cmd = exec.CommandContext(ctx, "git", "worktree", "add", "-b", workspace.Branch, targetPath)
			cmd.Dir = repo.Path
			if output, err := cmd.CombinedOutput(); err != nil {
				return errors.Wrapf(err, "failed to create worktree: %s", string(output))
			}
		} else {
			return errors.Wrapf(err, "failed to create worktree: %s", string(output))
		}
	}

	return nil
}

// createGoWorkspace creates a go.work file
func (wm *WorkspaceManager) createGoWorkspace(workspace *Workspace) error {
	goWorkPath := filepath.Join(workspace.Path, "go.work")
	
	log.Info().Str("path", goWorkPath).Msg("Creating go.work file")

	content := "go 1.23\n\nuse (\n"
	
	for _, repo := range workspace.Repositories {
		// Check if repo has go.mod
		goModPath := filepath.Join(workspace.Path, repo.Name, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			content += fmt.Sprintf("\t./%s\n", repo.Name)
		}
	}
	
	content += ")\n"

	if err := os.WriteFile(goWorkPath, []byte(content), 0644); err != nil {
		return errors.Wrapf(err, "failed to write go.work file")
	}

	return nil
}

// copyAgentMD copies AGENT.md file to workspace
func (wm *WorkspaceManager) copyAgentMD(workspace *Workspace) error {
	// Expand ~ in source path
	source := workspace.AgentMD
	if strings.HasPrefix(source, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return errors.Wrap(err, "failed to get home directory")
		}
		source = filepath.Join(home, source[1:])
	}

	target := filepath.Join(workspace.Path, "AGENT.md")
	
	log.Info().Str("source", source).Str("target", target).Msg("Copying AGENT.md")

	data, err := os.ReadFile(source)
	if err != nil {
		return errors.Wrapf(err, "failed to read source file: %s", source)
	}

	if err := os.WriteFile(target, data, 0644); err != nil {
		return errors.Wrapf(err, "failed to write target file: %s", target)
	}

	return nil
}

// saveWorkspace saves workspace configuration
func (wm *WorkspaceManager) saveWorkspace(workspace *Workspace) error {
	workspacesDir := filepath.Join(filepath.Dir(wm.config.RegistryPath), "workspaces")
	if err := os.MkdirAll(workspacesDir, 0755); err != nil {
		return errors.Wrap(err, "failed to create workspaces directory")
	}

	configPath := filepath.Join(workspacesDir, workspace.Name+".json")
	
	data, err := json.MarshalIndent(workspace, "", "  ")
	if err != nil {
		return errors.Wrap(err, "failed to marshal workspace configuration")
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return errors.Wrap(err, "failed to write workspace configuration")
	}

	return nil
}

// loadConfig loads workspace manager configuration
func loadConfig() (*WorkspaceConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	config := &WorkspaceConfig{
		WorkspaceDir: filepath.Join(home, "workspaces", time.Now().Format("2006-01-02")),
		TemplateDir:  filepath.Join(home, "templates"),
		RegistryPath: filepath.Join(configDir, "workspace-manager", "registry.json"),
	}

	return config, nil
}

// loadWorkspaces loads all workspace configurations
func loadWorkspaces() ([]Workspace, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	workspacesDir := filepath.Join(configDir, "workspace-manager", "workspaces")
	
	if _, err := os.Stat(workspacesDir); os.IsNotExist(err) {
		return []Workspace{}, nil
	}

	entries, err := os.ReadDir(workspacesDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read workspaces directory")
	}

	var workspaces []Workspace
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			path := filepath.Join(workspacesDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				log.Warn().Err(err).Str("path", path).Msg("Failed to read workspace file")
				continue
			}

			var workspace Workspace
			if err := json.Unmarshal(data, &workspace); err != nil {
				log.Warn().Err(err).Str("path", path).Msg("Failed to parse workspace file")
				continue
			}

			workspaces = append(workspaces, workspace)
		}
	}

	return workspaces, nil
}
