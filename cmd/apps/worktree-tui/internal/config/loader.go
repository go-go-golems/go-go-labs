package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Load reads and validates the configuration
func Load() (*Config, error) {
	log.Debug().Msg("Starting configuration load")
	
	// Get the config file path from viper
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		return nil, fmt.Errorf("no config file found")
	}
	
	log.Debug().Str("config_file", configFile).Msg("Loading config file directly")
	
	// Read the file directly
	data, err := os.ReadFile(configFile)
	if err != nil {
		log.Error().Err(err).Str("file", configFile).Msg("Failed to read config file")
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}
	
	log.Debug().Str("raw_yaml", string(data)).Msg("Raw YAML content")
	
	var cfg Config
	
	// Parse YAML directly
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal YAML config")
		return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
	}

	log.Debug().Int("repositories", len(cfg.Repositories)).Msg("Config unmarshaled")
	
	// Debug: Print the parsed config structure
	log.Debug().Interface("parsed_config", cfg).Msg("Parsed configuration structure")

	// Set defaults if not provided
	if cfg.Workspaces.DefaultBasePath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		cfg.Workspaces.DefaultBasePath = filepath.Join(home, "code", "workspaces")
		log.Debug().Str("default_path", cfg.Workspaces.DefaultBasePath).Msg("Set default workspace path")
	}

	// Expand tilde in paths BEFORE validation
	originalBasePath := cfg.Workspaces.DefaultBasePath
	cfg.Workspaces.DefaultBasePath = expandPath(cfg.Workspaces.DefaultBasePath)
	log.Debug().Str("original", originalBasePath).Str("expanded", cfg.Workspaces.DefaultBasePath).Msg("Expanded workspace base path")
	
	for i := range cfg.Repositories {
		repo := &cfg.Repositories[i]
		log.Debug().Str("name", repo.Name).Str("local_path", repo.LocalPath).Str("url", repo.URL).Msg("Processing repository")
		
		if repo.LocalPath != "" {
			originalPath := repo.LocalPath
			repo.LocalPath = expandPath(repo.LocalPath)
			log.Debug().Str("name", repo.Name).Str("original", originalPath).Str("expanded", repo.LocalPath).Msg("Expanded repository path")
		}
		if repo.DefaultBranch == "" {
			repo.DefaultBranch = "main"
		}
	}

	log.Debug().Msg("Starting config validation")
	if err := validateConfig(&cfg); err != nil {
		log.Error().Err(err).Msg("Config validation failed")
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	log.Debug().Msg("Configuration loaded successfully")
	return &cfg, nil
}

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	
	// Handle ~/path correctly by removing ~/ and joining with home
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	
	// Handle ~ alone
	if path == "~" {
		return home
	}
	
	// Handle ~path (without slash)
	return filepath.Join(home, path[1:])
}

// validateConfig performs basic validation on the configuration
func validateConfig(cfg *Config) error {
	log.Debug().Int("repositories", len(cfg.Repositories)).Msg("Validating configuration")
	
	if len(cfg.Repositories) == 0 {
		return fmt.Errorf("no repositories configured")
	}

	// Validate repository names are unique
	repoNames := make(map[string]bool)
	for _, repo := range cfg.Repositories {
		log.Debug().
			Str("name", repo.Name).
			Str("local_path", repo.LocalPath).
			Str("url", repo.URL).
			Msg("Validating repository")
			
		if repo.Name == "" {
			return fmt.Errorf("repository name cannot be empty")
		}
		if repoNames[repo.Name] {
			return fmt.Errorf("duplicate repository name: %s", repo.Name)
		}
		repoNames[repo.Name] = true

		// Validate that either LocalPath or URL is provided
		if repo.LocalPath == "" && repo.URL == "" {
			log.Error().Str("name", repo.Name).Msg("Repository has neither local_path nor url")
			return fmt.Errorf("repository %s must have either local_path or url", repo.Name)
		}

		// If local path is provided, check if it exists
		if repo.LocalPath != "" {
			log.Debug().Str("name", repo.Name).Str("path", repo.LocalPath).Msg("Checking if local path exists")
			if _, err := os.Stat(repo.LocalPath); os.IsNotExist(err) {
				log.Error().Str("name", repo.Name).Str("path", repo.LocalPath).Msg("Local path does not exist")
				return fmt.Errorf("local path for repository %s does not exist: %s", repo.Name, repo.LocalPath)
			}
			log.Debug().Str("name", repo.Name).Str("path", repo.LocalPath).Msg("Local path exists")
		}
	}

	// Validate presets reference existing repositories
	for _, preset := range cfg.Presets {
		if preset.Name == "" {
			return fmt.Errorf("preset name cannot be empty")
		}
		for _, repoName := range preset.Repositories {
			if !repoNames[repoName] {
				return fmt.Errorf("preset %s references unknown repository: %s", preset.Name, repoName)
			}
		}
	}

	return nil
}

// GetRepositoryByName returns a repository by name
func (c *Config) GetRepositoryByName(name string) (*Repository, bool) {
	for _, repo := range c.Repositories {
		if repo.Name == name {
			return &repo, true
		}
	}
	return nil, false
}