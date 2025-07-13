package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
	"github.com/rs/zerolog/log"
)

// Load loads the configuration from environment variables
func Load() (*types.Config, error) {
	log.Debug().Msg("Loading configuration from environment variables")

	config := &types.Config{}

	// OIDC Configuration
	config.OIDCIssuer = getEnvString("OIDC_ISSUER", "http://localhost:8080")
	config.DefaultUser = getEnvString("DEFAULT_USER_USERNAME", "wesen")
	config.DefaultPassword = getEnvString("DEFAULT_USER_PASSWORD", "secret")

	// Server Configuration
	config.Host = getEnvString("HOST", "0.0.0.0")
	config.Port = getEnvInt("PORT", 8080)
	config.LogLevel = getEnvString("LOG_LEVEL", "info")

	// Validate configuration
	if err := validateConfig(config); err != nil {
		log.Error().
			Err(err).
			Interface("config", redactSensitiveFields(config)).
			Msg("Configuration validation failed")
		return nil, err
	}

	log.Info().
		Interface("config", redactSensitiveFields(config)).
		Msg("Configuration loaded and validated successfully")

	return config, nil
}

// validateConfig validates that required configuration is present
func validateConfig(config *types.Config) error {
	// Validate server configuration
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port number: %d (must be between 1 and 65535)", config.Port)
	}

	// Validate OIDC issuer
	if config.OIDCIssuer == "" {
		return fmt.Errorf("OIDC_ISSUER is required")
	}

	// Validate default user credentials
	if config.DefaultUser == "" {
		return fmt.Errorf("DEFAULT_USER_USERNAME is required")
	}
	if config.DefaultPassword == "" {
		return fmt.Errorf("DEFAULT_USER_PASSWORD is required")
	}

	return nil
}

// redactSensitiveFields creates a copy of the config with sensitive fields redacted for logging
func redactSensitiveFields(config *types.Config) map[string]interface{} {
	return map[string]interface{}{
		"oidc_issuer":      config.OIDCIssuer,
		"default_user":     config.DefaultUser,
		"default_password": redactSecret(config.DefaultPassword),
		"host":             config.Host,
		"port":             config.Port,
		"log_level":        config.LogLevel,
	}
}

// redactSecret redacts secrets for logging while preserving some info for debugging
func redactSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "[REDACTED]"
	}
	return secret[:4] + "[REDACTED]" + secret[len(secret)-4:]
}

// getEnvString gets a string environment variable with a default value
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
		log.Warn().
			Str("key", key).
			Str("value", value).
			Int("default", defaultValue).
			Msg("Invalid integer value for environment variable, using default")
	}
	return defaultValue
}
