package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	// Placeholder for internal packages - adjust path if needed
	// config_pkg "github.com/your-org/desktop-organizer-go/internal/config"
	// log_pkg "github.com/your-org/desktop-organizer-go/internal/log"
)

var (
	cfgFile string
	// cliCfg config_pkg.Config // Placeholder for config struct
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "desktop-organizer",
	Short: "Analyzes a directory (like Downloads) to provide insights for organization.",
	Long: `Scans a target directory, identifies file types, sizes, duplicates, 
modification dates, and generates a report. This report can be used 
manually or by an LLM to suggest cleanup and organization rules.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize Logger (before loading config potentially)
		logLevel := zerolog.InfoLevel
		if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
			logLevel = zerolog.DebugLevel
		}
		// TODO: Replace with log_pkg initialization logic
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(logLevel)
		log.Debug().Msg("Logger initialized")

		// Initialize Viper
		if err := initConfig(cmd); err != nil {
			return err
		}

		// TODO: Initialize debug file logging if specified
		// debugLogFile := viper.GetString("debugLog") ...

		log.Debug().Str("configFile", viper.ConfigFileUsed()).Msg("Using config file")

		// TODO: Unmarshal viper config into cliCfg struct
		// if err := viper.Unmarshal(&cliCfg); err != nil {
		// 	return fmt.Errorf("unable to decode config: %w", err)
		// }

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info().Msg("Starting Desktop Organizer Analysis...")

		// TODO: Load full config (Viper + Flags into cliCfg)

		// TODO: Instantiate Analysis Runner with Config

		// TODO: Execute Analysis Runner

		// TODO: Instantiate Reporter based on output format

		// TODO: Generate Report

		log.Info().Msg("Analysis finished successfully.")
		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Error().Err(err).Msg("Command execution failed")
		os.Exit(1)
	}
}

func init() {
	// Persistent flags (available to this command and all subcommands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.desktop-organizer.yaml or ./config.yaml)")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose/debug logging")
	rootCmd.PersistentFlags().String("debug-log", "", "Path to write debug logs to a file")

	// Local flags (only available to this command)
	// Mirroring script flags + additions
	rootCmd.Flags().StringP("downloads-dir", "d", "", "Directory to analyze (required)")
	rootCmd.Flags().StringP("output-file", "o", "", "Output file path (default: stdout)")
	rootCmd.Flags().IntP("sample-per-dir", "s", 0, "Enable sampling: max N files per directory for type analysis (0=disabled)")
	rootCmd.Flags().Int("max-workers", 4, "Number of concurrent workers for file analysis")
	rootCmd.Flags().String("output-format", "text", "Output format (text, json, markdown)")

	// Mark required flags
	_ = rootCmd.MarkFlagRequired("downloads-dir")

	// Bind flags to Viper keys
	_ = viper.BindPFlag("targetDir", rootCmd.Flags().Lookup("downloads-dir"))
	_ = viper.BindPFlag("outputFile", rootCmd.Flags().Lookup("output-file"))
	_ = viper.BindPFlag("debugLog", rootCmd.Flags().Lookup("debug-log"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")) // Bind persistent flag too
	_ = viper.BindPFlag("samplingPerDir", rootCmd.Flags().Lookup("sample-per-dir"))
	_ = viper.BindPFlag("maxWorkers", rootCmd.Flags().Lookup("max-workers"))
	_ = viper.BindPFlag("outputFormat", rootCmd.Flags().Lookup("output-format"))

	// TODO: Add more viper bindings for config struct fields if needed
}

// initConfig reads in config file and ENV variables if set.
func initConfig(cmd *cobra.Command) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".desktop-organizer" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(".") // Also look in the current directory
		viper.SetConfigName(".desktop-organizer")
		viper.SetConfigType("yaml") // Or json, toml
	}

	viper.SetEnvPrefix("DESKTOP_ORGANIZER")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("file", viper.ConfigFileUsed()).Msg("Using config file")
	} else {
		// Don't fail if config file not found, only if parsing failed
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config file %s: %w", viper.ConfigFileUsed(), err)
		}
		log.Debug().Msg("No config file found, using defaults and flags.")
	}
	return nil
}

func main() {
	Execute()
}
