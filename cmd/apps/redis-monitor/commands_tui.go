package main

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
)

// TUICommand starts the terminal UI for Redis monitoring
type TUICommand struct {
	*cmds.CommandDescription
}

type TUISettings struct {
	RedisAddr     string `glazed.parameter:"redis-addr"`
	RedisPassword string `glazed.parameter:"redis-password"`
	RedisDB       int    `glazed.parameter:"redis-db"`
	RefreshRate   string `glazed.parameter:"refresh-rate"`
	Demo          bool   `glazed.parameter:"demo"`
}

func NewTUICommand() (*TUICommand, error) {
	return &TUICommand{
		CommandDescription: cmds.NewCommandDescription(
			"tui",
			cmds.WithShort("Start the terminal UI for Redis streams monitoring"),
			cmds.WithLong(`Start an interactive terminal UI for monitoring Redis streams.

The TUI provides a top-like interface showing:
- Real-time stream metrics (entries, memory, groups)
- Consumer group details with pending messages
- Memory usage graphs and throughput sparklines
- Server uptime and global statistics

Navigation:
- R: Refresh all data
- G: Filter by group
- S: Sort by different metrics
- Q: Quit

This command provides the same functionality as the CLI commands but in an
interactive format suitable for real-time monitoring.`),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"redis-addr",
					parameters.ParameterTypeString,
					parameters.WithDefault("localhost:6379"),
					parameters.WithHelp("Redis server address"),
				),
				parameters.NewParameterDefinition(
					"redis-password",
					parameters.ParameterTypeString,
					parameters.WithDefault(""),
					parameters.WithHelp("Redis password"),
				),
				parameters.NewParameterDefinition(
					"redis-db",
					parameters.ParameterTypeInteger,
					parameters.WithDefault(0),
					parameters.WithHelp("Redis database number"),
				),
				parameters.NewParameterDefinition(
					"refresh-rate",
					parameters.ParameterTypeString, // Duration will be parsed from string
					parameters.WithDefault("5s"),
					parameters.WithHelp("Refresh rate for auto-updates (e.g., 5s, 10s)"),
				),
				parameters.NewParameterDefinition(
					"demo",
					parameters.ParameterTypeBool,
					parameters.WithDefault(false),
					parameters.WithHelp("Use demo data instead of connecting to Redis"),
				),
			),
		),
	}, nil
}

func (c *TUICommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	s := &TUISettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return fmt.Errorf("failed to initialize settings: %w", err)
	}

	// Parse refresh rate from string
	refreshRate, err := time.ParseDuration(s.RefreshRate)
	if err != nil {
		return fmt.Errorf("invalid refresh-rate format: %w", err)
	}

	var client *RedisClient
	if !s.Demo {
		// Test Redis connection first
		client = NewRedisClient(s.RedisAddr, s.RedisPassword, s.RedisDB)
		defer client.Close()

		if err := client.Ping(ctx); err != nil {
			return fmt.Errorf("failed to connect to Redis: %w", err)
		}
	}

	// Start the bubbletea TUI
	model := NewModel(client, s.Demo, refreshRate)
	p := tea.NewProgram(model, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	return nil
}

var _ cmds.BareCommand = &TUICommand{}
