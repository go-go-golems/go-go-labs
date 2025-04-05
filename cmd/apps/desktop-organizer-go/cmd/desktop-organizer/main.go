package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/analysis"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/config"
	applog "github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/log"
	"github.com/go-go-golems/go-go-labs/cmd/apps/desktop-organizer-go/internal/reporting"
)

var (
	cfgFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "desktop-organizer",
	Short: "Analyzes a directory (like Downloads) to provide insights for organization.",
	Long: `Scans a target directory, identifies file types, sizes, duplicates, 
modification dates, and generates a report. This report can be used 
manually or by an LLM to suggest cleanup and organization rules.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize basic logger for early errors during config loading
		logLevelStr := viper.GetString("logLevel")
		logLevel, err := zerolog.ParseLevel(strings.ToLower(logLevelStr))
		if err != nil {
			zlog.Warn().Str("level", logLevelStr).Msg("Invalid log level provided, defaulting to info")
			logLevel = zerolog.InfoLevel
		}
		zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(logLevel)
		zlog.Debug().Msg("Basic logger initialized")

		// Initialize Viper
		if err := initConfig(cmd); err != nil {
			return err
		}

		zlog.Debug().Str("configFile", viper.ConfigFileUsed()).Msg("Using config file")

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize full-featured logger with context and optional file output
		ctx, err := applog.InitLogger(viper.GetString("logLevel"), viper.GetString("debugLog"))
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		applog.Info(ctx).Msg("Starting Desktop Organizer Analysis...")

		// Process tool paths from flags into Viper map BEFORE loading config
		if cmd.Flags().Changed("tool-path") {
			toolPaths, _ := cmd.Flags().GetStringSlice("tool-path")
			toolPathMap := make(map[string]string)
			for _, pathSpec := range toolPaths {
				parts := strings.SplitN(pathSpec, "=", 2)
				if len(parts) == 2 {
					toolPathMap[parts[0]] = parts[1]
				} else {
					applog.Warn(ctx).Str("spec", pathSpec).Msg("Ignoring invalid tool-path specification")
				}
			}
			viper.Set("toolPaths", toolPathMap)
			applog.Debug(ctx).Interface("toolPaths", toolPathMap).Msg("Processed tool paths from flags")
		}

		// Load configuration
		cfg, err := config.LoadConfig(cmd, cfgFile)
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}

		// Create analyzer registry and register analyzers
		registry := analysis.NewRegistry()

		// Register built-in analyzers
		registry.Register(analysis.NewMagikaTypeAnalyzer())
		// TODO: Register additional analyzers
		// registry.Register(analysis.NewFileTypeAnalyzer())
		// registry.Register(analysis.NewHashingAnalyzer())
		// etc.

		// Create analysis runner
		runner, err := analysis.NewRunner(ctx, cfg, registry)
		if err != nil {
			return fmt.Errorf("failed to create analysis runner: %w", err)
		}

		// Run analysis
		result, err := runner.Run(ctx)
		if err != nil {
			return fmt.Errorf("analysis failed: %w", err)
		}

		// Setup output (file or stdout)
		var outputFile *os.File
		if cfg.OutputFile != "" {
			outputFile, err = os.Create(cfg.OutputFile)
			if err != nil {
				return fmt.Errorf("failed to create output file: %w", err)
			}
			defer outputFile.Close()
		} else {
			outputFile = os.Stdout
		}

		// Create reporter registry and register reporters
		reporterRegistry := reporting.NewRegistry()
		reporterRegistry.Register(reporting.NewJSONReporter(true)) // Pretty JSON
		// TODO: Register additional reporters
		reporterRegistry.Register(reporting.NewTextReporter())
		// reporterRegistry.Register(reporting.NewMarkdownReporter())

		// Get reporter based on format
		reporter, err := reporterRegistry.GetReporter(cfg.OutputFormat)
		if err != nil {
			// Fallback to JSON if requested format is not available
			applog.Warn(ctx).
				Err(err).
				Str("format", cfg.OutputFormat).
				Msg("Requested output format not available, using JSON")
			reporter = reporting.NewJSONReporter(true)
		}

		// Generate report
		if err := reporter.GenerateReport(ctx, result, outputFile); err != nil {
			return fmt.Errorf("failed to generate report: %w", err)
		}

		applog.Info(ctx).Msg("Analysis finished successfully.")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		zlog.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func init() {
	// Persistent flags (available to this command and all subcommands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.desktop-organizer.yaml or ./config.yaml)")
	rootCmd.PersistentFlags().String("log-level", "info", "Set the logging level (trace, debug, info, warn, error)")
	rootCmd.PersistentFlags().String("debug-log", "", "Path to write debug logs to a file")

	// Local flags (only available to this command)
	// Mirroring script flags + additions
	rootCmd.Flags().StringP("downloads-dir", "d", "", "Directory to analyze (required)")
	rootCmd.Flags().StringP("output-file", "o", "", "Output file path (default: stdout)")
	rootCmd.Flags().IntP("sample-per-dir", "s", 0, "Enable sampling: max N files per directory for type analysis (0=disabled)")
	rootCmd.Flags().Int("max-workers", 4, "Number of concurrent workers for file analysis")
	rootCmd.Flags().String("output-format", "text", "Output format (text, json, markdown)")
	rootCmd.Flags().StringSlice("exclude-path", nil, "Glob patterns for paths to exclude (can specify multiple)")
	rootCmd.Flags().Int("large-file-mb", 100, "Threshold in MB to tag files as 'large' (default: 100)")
	rootCmd.Flags().Int("recent-days", 30, "Threshold in days to tag files as 'recent' (default: 30)")
	rootCmd.Flags().StringSlice("tool-path", nil, "Override path for external tools (e.g., --tool-path magika=/usr/local/bin/magika)")
	rootCmd.Flags().StringSlice("enable-analyzer", nil, "Explicitly enable specific analyzers")
	rootCmd.Flags().StringSlice("disable-analyzer", nil, "Explicitly disable specific analyzers")

	// Mark required flags
	_ = rootCmd.MarkFlagRequired("downloads-dir")

	// Bind flags to Viper keys
	_ = viper.BindPFlag("targetDir", rootCmd.Flags().Lookup("downloads-dir"))
	_ = viper.BindPFlag("outputFile", rootCmd.Flags().Lookup("output-file"))
	_ = viper.BindPFlag("debugLog", rootCmd.Flags().Lookup("debug-log"))
	_ = viper.BindPFlag("logLevel", rootCmd.PersistentFlags().Lookup("log-level"))
	_ = viper.BindPFlag("samplingPerDir", rootCmd.Flags().Lookup("sample-per-dir"))
	_ = viper.BindPFlag("maxWorkers", rootCmd.Flags().Lookup("max-workers"))
	_ = viper.BindPFlag("outputFormat", rootCmd.Flags().Lookup("output-format"))
	_ = viper.BindPFlag("excludePaths", rootCmd.Flags().Lookup("exclude-path"))
	_ = viper.BindPFlag("largeFileThreshold", rootCmd.Flags().Lookup("large-file-mb"))
	_ = viper.BindPFlag("recentFileDays", rootCmd.Flags().Lookup("recent-days"))
	_ = viper.BindPFlag("enabledAnalyzers", rootCmd.Flags().Lookup("enable-analyzer"))
	_ = viper.BindPFlag("disabledAnalyzers", rootCmd.Flags().Lookup("disable-analyzer"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		// cobra.CheckErr(err) -- Don't check err here, let viper handle it
		if err != nil {
			// Log warning if home dir cannot be found, but continue
			zlog.Warn().Err(err).Msg("Could not find user home directory for config search")
		} else {
			viper.AddConfigPath(home)
		}

		// Search config in home directory with name ".desktop-organizer" (without extension).
		//viper.AddConfigPath(home) // Moved up
		viper.AddConfigPath(".") // Also look in the current directory
		viper.SetConfigName(".desktop-organizer")
		viper.SetConfigType("yaml") // Or json, toml
	}

	viper.SetEnvPrefix("DESKTOP_ORGANIZER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		zlog.Info().Str("file", viper.ConfigFileUsed()).Msg("Using config file")
	} else {
		// Don't fail if config file not found, only if parsing failed
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file %s: %w", viper.ConfigFileUsed(), err)
		}
		zlog.Debug().Msg("No config file found, using defaults and flags.")
	}
	return nil
}

func main() {
	Execute()
}
