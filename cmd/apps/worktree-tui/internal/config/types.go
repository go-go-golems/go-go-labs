package config

// Config represents the complete configuration structure
type Config struct {
	Workspaces   WorkspaceConfig `yaml:"workspaces"`
	Repositories []Repository    `yaml:"repositories"`
	Presets      []Preset        `yaml:"presets"`
}

// WorkspaceConfig contains workspace-related settings
type WorkspaceConfig struct {
	DefaultBasePath string `yaml:"default_base_path"`
}

// Repository represents a repository configuration
type Repository struct {
	Name          string   `yaml:"name"`
	Description   string   `yaml:"description"`
	LocalPath     string   `yaml:"local_path,omitempty"`
	URL           string   `yaml:"url,omitempty"`
	Subdirectory  string   `yaml:"subdirectory,omitempty"`
	DefaultBranch string   `yaml:"default_branch"`
	Tags          []string `yaml:"tags,omitempty"`
}

// Preset represents a predefined set of repositories
type Preset struct {
	Name         string   `yaml:"name"`
	Description  string   `yaml:"description"`
	Repositories []string `yaml:"repositories"`
}

// WorkspaceRequest represents a workspace creation request
type WorkspaceRequest struct {
	Name         string
	Path         string
	Repositories []Repository
}

// RepositorySelection represents a selected repository with its configuration
type RepositorySelection struct {
	Repository Repository
	Branch     string
	Selected   bool
}
