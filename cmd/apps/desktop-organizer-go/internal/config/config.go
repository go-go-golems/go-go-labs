package config

import (
	"fmt"
	"runtime"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config holds all configuration settings for the application.
type Config struct {
	TargetDir          string                    `mapstructure:"targetDir"`
	OutputFile         string                    `mapstructure:"outputFile"`
	DebugLog           string                    `mapstructure:"debugLog"`
	Verbose            bool                      `mapstructure:"verbose"`
	SamplingPerDir     int                       `mapstructure:"samplingPerDir"`
	MaxWorkers         int                       `mapstructure:"maxWorkers"`
	OutputFormat       string                    `mapstructure:"outputFormat"`
	ExcludePaths       []string                  `mapstructure:"excludePaths"`
	LargeFileThreshold int64                     `mapstructure:"largeFileThreshold"`
	RecentFileDays     int                       `mapstructure:"recentFileDays"`
	ToolPaths          map[string]string         `mapstructure:"toolPaths"`
	EnabledAnalyzers   []string                  `mapstructure:"enabledAnalyzers"`
	DisabledAnalyzers  []string                  `mapstructure:"disabledAnalyzers"`
	AnalyzerConfigs    map[string]map[string]any `mapstructure:"analyzerConfigs"`
}

// LoadConfig initializes Viper and unmarshals the configuration.
func LoadConfig(cmd *cobra.Command, cfgFile string) (*Config, error) {
	// Set default values
	viper.SetDefault("maxWorkers", runtime.NumCPU())
	viper.SetDefault("outputFormat", "text")
	viper.SetDefault("largeFileThreshold", 100) // In MB
	viper.SetDefault("recentFileDays", 30)
	viper.SetDefault("toolPaths", map[string]string{})
	viper.SetDefault("excludePaths", []string{})

	// Don't need to call viper.SetConfigFile, AddConfigPath, etc.
	// That should be done in the calling code (main.go initConfig)

	// Unmarshal into Config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config: %w", err)
	}

	// Post-processing/validation
	if config.TargetDir == "" {
		return nil, fmt.Errorf("target directory is required")
	}

	// Convert MB to bytes for threshold
	config.LargeFileThreshold = config.LargeFileThreshold * 1024 * 1024

	// Ensure max workers is reasonable
	if config.MaxWorkers <= 0 {
		config.MaxWorkers = runtime.NumCPU()
		log.Debug().Int("maxWorkers", config.MaxWorkers).Msg("Invalid maxWorkers value, using CPU count")
	}

	return &config, nil
}
