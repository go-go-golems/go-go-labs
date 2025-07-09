package cmd

import (
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// GlobalConfig holds global configuration
type GlobalConfig struct {
	Port        string
	Host        string
	LogLevel    string
	Timeout     time.Duration
	DebugSerial bool
	HexDump     bool
}

var globalConfig GlobalConfig

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "meshtastic",
	Short: "A comprehensive CLI for Meshtastic devices",
	Long: `A comprehensive command-line interface for interacting with Meshtastic devices.
Supports serial, TCP, and BLE connections with robust error handling and retry logic.

Examples:
  meshtastic info --port /dev/ttyUSB0
  meshtastic nodes --show-fields id,user,snr
  meshtastic connect --host 192.168.1.100
  meshtastic discover --serial-only`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		return setupLogging(globalConfig.LogLevel)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Error().Err(err).Msg("Command failed")
		os.Exit(1)
	}
}

func setupLogging(level string) error {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Configure console writer for human-readable output
	output := zerolog.ConsoleWriter{Out: os.Stderr}
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		return errors.Errorf("invalid log level: %s", level)
	}

	return nil
}

func init() {
	// Global persistent flags
	rootCmd.PersistentFlags().StringVarP(&globalConfig.Port, "port", "p", "/dev/ttyUSB0", "Serial port for Meshtastic device")
	rootCmd.PersistentFlags().StringVar(&globalConfig.Host, "host", "", "TCP/IP host (e.g., 192.168.1.100)")
	rootCmd.PersistentFlags().StringVar(&globalConfig.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().DurationVar(&globalConfig.Timeout, "timeout", 10*time.Second, "Operation timeout")
	rootCmd.PersistentFlags().BoolVar(&globalConfig.DebugSerial, "debug-serial", false, "Enable verbose serial communication logging")
	rootCmd.PersistentFlags().BoolVar(&globalConfig.HexDump, "hex-dump", false, "Enable hex dump logging of raw serial data")

	// Add subcommands
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(discoverCmd)
	rootCmd.AddCommand(infoCmd)
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(simpleConfigCmd)
	rootCmd.AddCommand(simpleChannelCmd)

	// Device management and telemetry commands are added in their respective init() functions
}
