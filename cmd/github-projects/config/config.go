package config

import (
	"fmt"
	"os"
	"strconv"
)

// GitHubConfig holds configuration for GitHub integration
type GitHubConfig struct {
	Token         string
	Owner         string
	ProjectNumber int
	Repository    string
}

// LoadGitHubConfig loads configuration from environment variables
func LoadGitHubConfig() (*GitHubConfig, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}

	owner := os.Getenv("GITHUB_OWNER")
	if owner == "" {
		return nil, fmt.Errorf("GITHUB_OWNER environment variable is required")
	}

	projectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBER")
	if projectNumberStr == "" {
		return nil, fmt.Errorf("GITHUB_PROJECT_NUMBER environment variable is required")
	}

	projectNumber, err := strconv.Atoi(projectNumberStr)
	if err != nil {
		return nil, fmt.Errorf("invalid GITHUB_PROJECT_NUMBER: %v", err)
	}

	repository := os.Getenv("GITHUB_REPOSITORY")
	if repository == "" {
		return nil, fmt.Errorf("GITHUB_REPOSITORY environment variable is required")
	}

	return &GitHubConfig{
		Token:         token,
		Owner:         owner,
		ProjectNumber: projectNumber,
		Repository:    repository,
	}, nil
}

// GetDefaultOwner returns default owner from env var
func GetDefaultOwner() string {
	return os.Getenv("GITHUB_OWNER")
}

// GetDefaultProjectNumber returns default project number from env var
func GetDefaultProjectNumber() int {
	projectNumberStr := os.Getenv("GITHUB_PROJECT_NUMBER")
	if projectNumberStr == "" {
		return 0
	}
	projectNumber, err := strconv.Atoi(projectNumberStr)
	if err != nil {
		return 0
	}
	return projectNumber
}
