package main

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/pkg/tui/components"
)

// Model represents the TUI state - now just wraps the coordinator
type Model struct {
	coordinator *components.Coordinator
}

// NewModel creates a new TUI model using the coordinator
func NewModel(client *RedisClient, demoMode bool, refreshRate time.Duration) Model {
	// Adapt RedisClient to components.RedisClient interface
	adaptedClient := &redisClientAdapter{client: client}

	return Model{
		coordinator: components.NewCoordinator(adaptedClient, demoMode, refreshRate),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return m.coordinator.Init()
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	coordinator, cmd := m.coordinator.Update(msg)
	m.coordinator = coordinator.(*components.Coordinator)
	return m, cmd
}

// View implements tea.Model
func (m Model) View() string {
	return m.coordinator.View()
}

// redisClientAdapter adapts RedisClient to components.RedisClient interface
type redisClientAdapter struct {
	client *RedisClient
}

func (a *redisClientAdapter) GetServerInfo(ctx context.Context) (*components.ServerInfo, error) {
	info, err := a.client.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &components.ServerInfo{
		UptimeInSeconds:   info.UptimeInSeconds,
		UsedMemory:        info.UsedMemory,
		TotalSystemMemory: info.TotalSystemMemory,
		Version:           info.Version,
	}, nil
}

func (a *redisClientAdapter) DiscoverStreams(ctx context.Context) ([]string, error) {
	return a.client.DiscoverStreams(ctx)
}

func (a *redisClientAdapter) GetStreamInfo(ctx context.Context, name string) (*components.StreamInfo, error) {
	info, err := a.client.GetStreamInfo(ctx, name)
	if err != nil {
		return nil, err
	}
	return &components.StreamInfo{
		Name:            info.Name,
		Length:          info.Length,
		MemoryUsage:     info.MemoryUsage,
		Groups:          info.Groups,
		LastGeneratedID: info.LastGeneratedID,
	}, nil
}

func (a *redisClientAdapter) GetStreamGroups(ctx context.Context, stream string) ([]components.GroupInfo, error) {
	groups, err := a.client.GetStreamGroups(ctx, stream)
	if err != nil {
		return nil, err
	}

	var result []components.GroupInfo
	for _, g := range groups {
		result = append(result, components.GroupInfo{
			Name:    g.Name,
			Pending: g.Pending,
		})
	}
	return result, nil
}

func (a *redisClientAdapter) GetGroupConsumers(ctx context.Context, stream, group string) ([]components.ConsumerInfo, error) {
	consumers, err := a.client.GetGroupConsumers(ctx, stream, group)
	if err != nil {
		return nil, err
	}

	var result []components.ConsumerInfo
	for _, c := range consumers {
		result = append(result, components.ConsumerInfo{
			Name:    c.Name,
			Pending: c.Pending,
			Idle:    c.Idle,
		})
	}
	return result, nil
}
