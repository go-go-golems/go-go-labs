package main

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
)

// Model represents the TUI state - now uses the new RootModel
type Model struct {
	root models.RootModel
}

// NewModel creates a new TUI model using the new RootModel
func NewModel(client *RedisClient, demoMode bool, refreshRate time.Duration) Model {
	var adaptedClient models.RedisClient
	
	if demoMode || client == nil {
		// Use nil client for demo mode
		adaptedClient = nil
	} else {
		// Adapt RedisClient to models.RedisClient interface
		adaptedClient = &redisClientAdapter{client: client}
	}

	return Model{
		root: models.NewRootModel(adaptedClient, demoMode, refreshRate),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return m.root.Init()
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	rootModel, cmd := m.root.Update(msg)
	m.root = rootModel.(models.RootModel)
	return m, cmd
}

// View implements tea.Model
func (m Model) View() string {
	return m.root.View()
}

// redisClientAdapter adapts RedisClient to models.RedisClient interface
type redisClientAdapter struct {
	client *RedisClient
}

func (a *redisClientAdapter) GetServerInfo(ctx context.Context) (*models.ServerInfo, error) {
	info, err := a.client.GetServerInfo(ctx)
	if err != nil {
		return nil, err
	}
	return &models.ServerInfo{
		UptimeInSeconds:   info.UptimeInSeconds,
		UsedMemory:        info.UsedMemory,
		TotalSystemMemory: info.TotalSystemMemory,
		Version:           info.Version,
	}, nil
}

func (a *redisClientAdapter) DiscoverStreams(ctx context.Context) ([]string, error) {
	return a.client.DiscoverStreams(ctx)
}

func (a *redisClientAdapter) GetStreamInfo(ctx context.Context, name string) (*models.StreamInfo, error) {
	if a.client == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	
	info, err := a.client.GetStreamInfo(ctx, name)
	if err != nil {
		return nil, err
	}
	
	if info == nil {
		return nil, fmt.Errorf("stream info is nil for stream: %s", name)
	}
	
	return &models.StreamInfo{
		Name:            info.Name,
		Length:          info.Length,
		MemoryUsage:     info.MemoryUsage,
		Groups:          info.Groups,
		LastGeneratedID: info.LastGeneratedID,
	}, nil
}

func (a *redisClientAdapter) GetStreamGroups(ctx context.Context, stream string) ([]models.GroupInfo, error) {
	groups, err := a.client.GetStreamGroups(ctx, stream)
	if err != nil {
		return nil, err
	}

	var result []models.GroupInfo
	for _, g := range groups {
		result = append(result, models.GroupInfo{
			Name:    g.Name,
			Pending: g.Pending,
		})
	}
	return result, nil
}

func (a *redisClientAdapter) GetGroupConsumers(ctx context.Context, stream, group string) ([]models.ConsumerInfo, error) {
	consumers, err := a.client.GetGroupConsumers(ctx, stream, group)
	if err != nil {
		return nil, err
	}

	var result []models.ConsumerInfo
	for _, c := range consumers {
		result = append(result, models.ConsumerInfo{
			Name:    c.Name,
			Pending: c.Pending,
			Idle:    c.Idle,
		})
	}
	return result, nil
}
