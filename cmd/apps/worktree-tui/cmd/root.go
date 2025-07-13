package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/tui"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "worktree-tui",
		Short: "Quick workspace setup tool using git worktrees",
		Long: `A Terminal User Interface (TUI) application that allows developers to quickly 
create Go workspaces by selecting from a predefined list of repositories and 
automatically setting up worktrees with go.work initialization.`,
		RunE: runTUI,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig, initLogging)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.config/worktree-tui/config.yaml)")
}

func initLogging() {
	// Set up console logging with debug level and caller info
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).With().Caller().Logger()

	log.Debug().Msg("Debug logging initialized")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		configDir := filepath.Join(home, ".config", "worktree-tui")
		viper.AddConfigPath(configDir)
		viper.AddConfigPath(".")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvPrefix("WORKTREE_TUI")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
			os.Exit(1)
		}
		log.Debug().Msg("No config file found, using defaults")
	} else {
		log.Debug().Str("config_file", viper.ConfigFileUsed()).Msg("Loaded config file")
	}
}

func runTUI(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	app := tui.NewApp(cfg)

	p := tea.NewProgram(app, tea.WithAltScreen())
	_, err = p.Run()

	return err
}
