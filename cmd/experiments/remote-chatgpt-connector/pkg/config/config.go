package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
)

// Load configuration from environment variables
func Load() (*types.Config, error) {
	cfg := &types.Config{
		Port:     8080,
		Host:     "0.0.0.0",
		LogLevel: "info",
	}

	// Parse port if provided
	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = port
	}

	if host := os.Getenv("HOST"); host != "" {
		cfg.Host = host
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	// GitHub OAuth config - required
	cfg.GitHubClientID = os.Getenv("GITHUB_CLIENT_ID")
	if cfg.GitHubClientID == "" {
		return nil, fmt.Errorf("GITHUB_CLIENT_ID is required")
	}

	cfg.GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")
	if cfg.GitHubClientSecret == "" {
		return nil, fmt.Errorf("GITHUB_CLIENT_SECRET is required")
	}

	cfg.AllowedLogin = os.Getenv("GITHUB_ALLOWED_LOGIN")
	if cfg.AllowedLogin == "" {
		return nil, fmt.Errorf("GITHUB_ALLOWED_LOGIN is required")
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func Validate(c *types.Config) error {
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.GitHubClientID == "" {
		return fmt.Errorf("github_client_id is required")
	}

	if c.GitHubClientSecret == "" {
		return fmt.Errorf("github_client_secret is required")
	}

	if c.AllowedLogin == "" {
		return fmt.Errorf("allowed_login is required")
	}

	return nil
}
