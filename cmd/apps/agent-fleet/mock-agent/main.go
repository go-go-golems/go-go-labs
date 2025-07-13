package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/mock-agent/agent"
	"github.com/go-go-golems/go-go-labs/cmd/apps/agent-fleet/mock-agent/client"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "mock-agent",
		Short: "Mock Agent Fleet Agent",
		Long:  `A mock agent that simulates realistic agent behavior for testing the Agent Fleet backend.`,
		Run:   runAgent,
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	rootCmd.PersistentFlags().StringP("server", "s", "http://localhost:8080", "Fleet backend server URL")
	rootCmd.PersistentFlags().StringP("token", "t", "fleet-agent-token-123", "Authentication token")
	rootCmd.PersistentFlags().StringP("name", "n", "", "Agent name (random if not specified)")
	rootCmd.PersistentFlags().StringP("worktree", "w", "/tmp/mock-project", "Agent worktree path")
	rootCmd.PersistentFlags().StringP("log-level", "l", "info", "Log level (trace, debug, info, warn, error)")
	rootCmd.PersistentFlags().BoolP("randomized", "r", false, "Enable randomized behavior mode")
	rootCmd.PersistentFlags().IntP("tick-interval", "", 5, "Agent tick interval in seconds")
	rootCmd.PersistentFlags().IntP("command-check-interval", "", 2, "Command check interval in seconds")

	viper.BindPFlag("server", rootCmd.PersistentFlags().Lookup("server"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("name", rootCmd.PersistentFlags().Lookup("name"))
	viper.BindPFlag("worktree", rootCmd.PersistentFlags().Lookup("worktree"))
	viper.BindPFlag("log-level", rootCmd.PersistentFlags().Lookup("log-level"))
	viper.BindPFlag("randomized", rootCmd.PersistentFlags().Lookup("randomized"))
	viper.BindPFlag("tick-interval", rootCmd.PersistentFlags().Lookup("tick-interval"))
	viper.BindPFlag("command-check-interval", rootCmd.PersistentFlags().Lookup("command-check-interval"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".mock-agent")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.Info().Str("config", viper.ConfigFileUsed()).Msg("Using config file")
	}
}

func setupLogger() {
	level := viper.GetString("log-level")
	logLevel, err := zerolog.ParseLevel(level)
	if err != nil {
		logLevel = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(logLevel)
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		short := file
		for i := len(file) - 1; i > 0; i-- {
			if file[i] == '/' {
				short = file[i+1:]
				break
			}
		}
		return fmt.Sprintf("%s:%d", short, line)
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}).With().Caller().Logger()
}

func generateAgentName() string {
	adjectives := []string{
		"Swift", "Clever", "Diligent", "Sharp", "Focused", "Precise", "Agile", "Smart",
		"Efficient", "Dedicated", "Reliable", "Adaptive", "Creative", "Innovative",
	}

	roles := []string{
		"Coder", "Reviewer", "Tester", "Debugger", "Architect", "Analyst", "Builder",
		"Scanner", "Optimizer", "Validator", "Developer", "Assistant", "Helper",
	}

	adj := adjectives[rand.Intn(len(adjectives))]
	role := roles[rand.Intn(len(roles))]

	return fmt.Sprintf("%s %s", adj, role)
}

func runAgent(cmd *cobra.Command, args []string) {
	setupLogger()

	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Get configuration
	serverURL := viper.GetString("server")
	token := viper.GetString("token")
	name := viper.GetString("name")
	worktree := viper.GetString("worktree")
	randomized := viper.GetBool("randomized")
	tickInterval := viper.GetInt("tick-interval")
	commandCheckInterval := viper.GetInt("command-check-interval")

	if name == "" {
		name = generateAgentName()
	}

	log.Info().
		Str("name", name).
		Str("server", serverURL).
		Str("worktree", worktree).
		Bool("randomized", randomized).
		Msg("Starting mock agent")

	// Create API client
	apiClient := client.New(serverURL, token)

	// Create agent configuration
	config := agent.Config{
		Name:                 name,
		Worktree:             worktree,
		Randomized:           randomized,
		TickInterval:         time.Duration(tickInterval) * time.Second,
		CommandCheckInterval: time.Duration(commandCheckInterval) * time.Second,
	}

	// Create and start agent
	mockAgent := agent.New(apiClient, config)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start agent in goroutine
	go func() {
		if err := mockAgent.Run(ctx); err != nil {
			log.Error().Err(err).Msg("Agent encountered error")
			cancel()
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info().Msg("Received shutdown signal")
	case <-ctx.Done():
		log.Info().Msg("Agent context cancelled")
	}

	// Graceful shutdown
	log.Info().Msg("Shutting down agent...")
	cancel()

	// Give the agent time to clean up
	time.Sleep(2 * time.Second)

	log.Info().Msg("Agent stopped")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Failed to execute command")
	}
}
