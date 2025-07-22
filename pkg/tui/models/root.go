// Package models contains the main TUI models for the Redis monitor
package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/tui/keys"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
	"github.com/go-go-golems/go-go-labs/pkg/tui/widgets"
)

// RedisClient interface for data fetching
type RedisClient interface {
	GetServerInfo(ctx context.Context) (*ServerInfo, error)
	DiscoverStreams(ctx context.Context) ([]string, error)
	GetStreamInfo(ctx context.Context, name string) (*StreamInfo, error)
	GetStreamGroups(ctx context.Context, stream string) ([]GroupInfo, error)
	GetGroupConsumers(ctx context.Context, stream, group string) ([]ConsumerInfo, error)
}

// Data structures for Redis client interface
type ServerInfo struct {
	UptimeInSeconds   int64
	UsedMemory        int64
	TotalSystemMemory int64
	Version           string
}

type StreamInfo struct {
	Name            string
	Length          int64
	MemoryUsage     int64
	Groups          int64
	LastGeneratedID string
}

type GroupInfo struct {
	Name    string
	Pending int64
}

type ConsumerInfo struct {
	Name    string
	Pending int64
	Idle    time.Duration
}

// RootModel is the main bubbletea model that composes all widgets
type RootModel struct {
	// Core state
	width       int
	height      int
	initialized bool
	
	// Data
	redisClient RedisClient
	serverData  widgets.ServerData
	streamsData []widgets.StreamData
	refreshRate time.Duration
	demoMode    bool
	
	// Data tracking for sparklines
	streamLengthHistory map[string][]int64
	lastRefresh         time.Time
	fetchInFlight       bool
	
	// UI state
	focusedWidget string
	refreshRates  []time.Duration
	refreshRateIdx int
	
	// Widgets
	header  widgets.HeaderWidget
	streams widgets.StreamsTableWidget
	groups  widgets.GroupsTableWidget
	alerts  widgets.AlertsWidget
	metrics widgets.MetricsWidget
	footer  widgets.FooterWidget
	
	// Input handling
	keys keys.KeyMap
}

// RefreshTickMsg signals it's time to refresh data
type RefreshTickMsg struct {
	Time time.Time
}

// DataFetchedMsg contains freshly fetched data
type DataFetchedMsg struct {
	ServerData  widgets.ServerData
	StreamsData []widgets.StreamData
	Error       error
}

// Available refresh rates
var defaultRefreshRates = []time.Duration{
	100 * time.Millisecond,  // 0.1s
	200 * time.Millisecond,  // 0.2s
	500 * time.Millisecond,  // 0.5s
	1 * time.Second,         // 1s
	2 * time.Second,         // 2s
	5 * time.Second,         // 5s
	10 * time.Second,        // 10s
}

// NewRootModel creates a new root model
func NewRootModel(client RedisClient, demoMode bool, refreshRate time.Duration) RootModel {
	// Initialize styles
	appStyles := styles.NewStyles()
	
	// Find closest refresh rate
	refreshRateIdx := findClosestRefreshRate(refreshRate)
	
	// Initialize key map
	keyMap := keys.DefaultKeyMap()
	
	return RootModel{
		redisClient:         client,
		demoMode:           demoMode,
		refreshRate:        defaultRefreshRates[refreshRateIdx],
		refreshRateIdx:     refreshRateIdx,
		refreshRates:       defaultRefreshRates,
		focusedWidget:      "streams",
		streamLengthHistory: make(map[string][]int64),
		keys:               keyMap,
		
		// Initialize widgets
		header:  widgets.NewHeaderWidget(appStyles.Header),
		streams: widgets.NewStreamsTableWidget(appStyles.StreamsTable),
		groups:  widgets.NewGroupsTableWidget(appStyles.GroupsTable),
		alerts:  widgets.NewAlertsWidget(appStyles.Alerts),
		metrics: widgets.NewMetricsWidget(appStyles.Metrics),
		footer:  widgets.NewFooterWidget(keyMap, appStyles.Footer),
	}
}

// Init implements tea.Model
func (m RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		m.startRefreshTimer(),
	)
}

// Update implements tea.Model
func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.initialized = true
		
		// Update widget sizes
		m.updateWidgetSizes()
		
		// Propagate to widgets
		cmds = append(cmds, m.updateWidgets(msg)...)
		
	case tea.KeyMsg:
		// Handle global keys first
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
			
		case key.Matches(msg, m.keys.Refresh):
			if !m.fetchInFlight {
				cmds = append(cmds, m.fetchData())
			}
			
		case key.Matches(msg, m.keys.RefreshUp):
			cmds = append(cmds, m.speedUp())
			
		case key.Matches(msg, m.keys.RefreshDown):
			cmds = append(cmds, m.speedDown())
			
		case key.Matches(msg, m.keys.FocusNext):
			m.cycleFocus(1)
			
		case key.Matches(msg, m.keys.FocusPrev):
			m.cycleFocus(-1)
			
		default:
			// Pass to focused widget
			cmds = append(cmds, m.updateWidgets(msg)...)
		}
		
	case RefreshTickMsg:
		if !m.fetchInFlight {
			cmds = append(cmds, m.fetchData())
		}
		cmds = append(cmds, m.startRefreshTimer())
		
	case DataFetchedMsg:
		m.fetchInFlight = false
		if msg.Error != nil {
			log.Printf("Error fetching data: %v", msg.Error)
		} else {
			m.serverData = msg.ServerData
			m.streamsData = msg.StreamsData
			m.lastRefresh = time.Now()
			
			// Update widgets with new data
			dataUpdate := widgets.DataUpdateMsg{
				ServerData:  msg.ServerData,
				StreamsData: msg.StreamsData,
				Timestamp:   time.Now(),
			}
			cmds = append(cmds, m.updateWidgets(dataUpdate)...)
		}
		
	default:
		// Pass other messages to widgets
		cmds = append(cmds, m.updateWidgets(msg)...)
	}
	
	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m RootModel) View() string {
	if !m.initialized {
		return "Initializing..."
	}
	
	var sections []string
	
	// Render each widget
	sections = append(sections, m.header.View())
	sections = append(sections, m.streams.View())
	sections = append(sections, m.groups.View())
	sections = append(sections, m.alerts.View())
	sections = append(sections, m.metrics.View())
	sections = append(sections, m.footer.View())
	
	return lipgloss.JoinVertical(lipgloss.Top, sections...)
}

// updateWidgetSizes calculates and sets sizes for all widgets
func (m *RootModel) updateWidgetSizes() {
	if m.height == 0 {
		return
	}
	
	// Calculate available height for dynamic widgets
	fixedHeight := m.header.MinHeight() + m.footer.MinHeight()
	availableHeight := m.height - fixedHeight
	
	// Distribute remaining space
	if availableHeight > 0 {
		// Give priority to streams table, then groups, then alerts
		streamHeight := min(availableHeight/2, m.streams.MaxHeight())
		remainingHeight := availableHeight - streamHeight
		
		groupHeight := min(remainingHeight/2, m.groups.MaxHeight())
		remainingHeight -= groupHeight
		
		alertHeight := min(remainingHeight-m.metrics.MinHeight(), m.alerts.MaxHeight())
		metricHeight := m.metrics.MinHeight()
		
		// Set widget sizes
		m.header.SetSize(m.width, m.header.MinHeight())
		m.streams.SetSize(m.width, streamHeight)
		m.groups.SetSize(m.width, groupHeight)
		m.alerts.SetSize(m.width, alertHeight)
		m.metrics.SetSize(m.width, metricHeight)
		m.footer.SetSize(m.width, m.footer.MinHeight())
	}
	
	// Update focus
	m.updateWidgetFocus()
}

// updateWidgetFocus sets focus on the appropriate widget
func (m *RootModel) updateWidgetFocus() {
	m.streams.SetFocused(m.focusedWidget == "streams")
	m.groups.SetFocused(m.focusedWidget == "groups")
	// Other widgets don't support focus
}

// updateWidgets sends a message to all widgets and collects commands
func (m *RootModel) updateWidgets(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	var model tea.Model
	
	// Update header widget
	m.header.SetRefreshRate(m.refreshRate)
	m.header.SetDemoMode(m.demoMode)
	model, cmd = m.header.Update(msg)
	m.header = model.(widgets.HeaderWidget)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Update streams widget
	model, cmd = m.streams.Update(msg)
	m.streams = model.(widgets.StreamsTableWidget)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Update groups widget
	model, cmd = m.groups.Update(msg)
	m.groups = model.(widgets.GroupsTableWidget)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Update alerts widget
	model, cmd = m.alerts.Update(msg)
	m.alerts = model.(widgets.AlertsWidget)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Update metrics widget
	model, cmd = m.metrics.Update(msg)
	m.metrics = model.(widgets.MetricsWidget)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	// Update footer widget
	model, cmd = m.footer.Update(msg)
	m.footer = model.(widgets.FooterWidget)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	
	return cmds
}

// cycleFocus moves focus between widgets
func (m *RootModel) cycleFocus(direction int) {
	focusableWidgets := []string{"streams", "groups"}
	
	currentIdx := 0
	for i, widget := range focusableWidgets {
		if widget == m.focusedWidget {
			currentIdx = i
			break
		}
	}
	
	newIdx := (currentIdx + direction + len(focusableWidgets)) % len(focusableWidgets)
	m.focusedWidget = focusableWidgets[newIdx]
	m.updateWidgetFocus()
}

// speedUp increases refresh rate
func (m *RootModel) speedUp() tea.Cmd {
	if m.refreshRateIdx > 0 {
		m.refreshRateIdx--
		m.refreshRate = m.refreshRates[m.refreshRateIdx]
	}
	return nil
}

// speedDown decreases refresh rate
func (m *RootModel) speedDown() tea.Cmd {
	if m.refreshRateIdx < len(m.refreshRates)-1 {
		m.refreshRateIdx++
		m.refreshRate = m.refreshRates[m.refreshRateIdx]
	}
	return nil
}

// startRefreshTimer starts the refresh timer
func (m *RootModel) startRefreshTimer() tea.Cmd {
	return tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
		return RefreshTickMsg{Time: t}
	})
}

// fetchData fetches fresh data from Redis
func (m *RootModel) fetchData() tea.Cmd {
	if m.fetchInFlight {
		return nil
	}
	
	m.fetchInFlight = true
	
	if m.demoMode {
		return m.fetchDemoData()
	}
	
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		serverData, streamsData, err := m.fetchRealData(ctx)
		return DataFetchedMsg{
			ServerData:  serverData,
			StreamsData: streamsData,
			Error:       err,
		}
	}
}

// fetchDemoData generates demo data
func (m *RootModel) fetchDemoData() tea.Cmd {
	return func() tea.Msg {
		// Generate demo server data
		serverData := widgets.ServerData{
			Uptime:      time.Duration(1098276) * time.Second, // ~12.7 days
			MemoryUsed:  348127232,                            // ~332MB
			MemoryTotal: 1073741824,                           // 1GB
			Version:     "7.0.8",
			Throughput:  1023.5,
		}
		
		// Generate demo streams data
		streamsData := []widgets.StreamData{
			{
				Name:        "orders",
				Length:      1243592,
				MemoryUsage: 126353408, // ~120.5MB
				Groups:      3,
				LastID:      "160123-7",
				MessageRates: []float64{4.0, 4.0, 4.0, 4.0, 2.0, 2.0, 2.0, 1.0, 1.0, 1.0},
				ConsumerGroups: []widgets.GroupData{
					{Name: "cg-1", Stream: "orders", Pending: 12, Consumers: []widgets.ConsumerData{
						{Name: "Alice", Pending: 3, Idle: 5 * time.Second},
						{Name: "Bob", Pending: 2, Idle: 1 * time.Second},
					}},
					{Name: "cg-2", Stream: "orders", Pending: 0, Consumers: []widgets.ConsumerData{
						{Name: "Charlie", Pending: 5, Idle: 1 * time.Second},
					}},
				},
			},
			{
				Name:        "events", 
				Length:      98234,
				MemoryUsage: 9648128, // ~9.2MB
				Groups:      5,
				LastID:      "160123-3",
				MessageRates: []float64{1.0, 1.0, 1.0, 1.0, 2.0, 2.0, 3.0, 4.0, 4.0, 3.0},
				ConsumerGroups: []widgets.GroupData{
					{Name: "cg-A", Stream: "events", Pending: 3, Consumers: []widgets.ConsumerData{
						{Name: "Eve", Pending: 4, Idle: 0},
						{Name: "Frank", Pending: 4, Idle: 0},
					}},
				},
			},
		}
		
		return DataFetchedMsg{
			ServerData:  serverData,
			StreamsData: streamsData,
			Error:       nil,
		}
	}
}

// fetchRealData fetches data from actual Redis
func (m *RootModel) fetchRealData(ctx context.Context) (widgets.ServerData, []widgets.StreamData, error) {
	if m.redisClient == nil {
		return widgets.ServerData{}, nil, fmt.Errorf("redis client not available")
	}
	
	// Fetch server info
	serverInfo, err := m.redisClient.GetServerInfo(ctx)
	if err != nil {
		return widgets.ServerData{}, nil, fmt.Errorf("failed to get server info: %w", err)
	}
	
	serverData := widgets.ServerData{
		Uptime:      time.Duration(serverInfo.UptimeInSeconds) * time.Second,
		MemoryUsed:  serverInfo.UsedMemory,
		MemoryTotal: serverInfo.TotalSystemMemory,
		Version:     serverInfo.Version,
	}
	
	// Discover streams
	streamNames, err := m.redisClient.DiscoverStreams(ctx)
	if err != nil {
		return serverData, nil, fmt.Errorf("failed to discover streams: %w", err)
	}
	
	var streamsData []widgets.StreamData
	for _, streamName := range streamNames {
		streamData, err := m.fetchStreamData(ctx, streamName)
		if err != nil {
			log.Printf("Failed to fetch data for stream %s: %v", streamName, err)
			continue
		}
		streamsData = append(streamsData, streamData)
	}
	
	return serverData, streamsData, nil
}

// fetchStreamData fetches data for a single stream
func (m *RootModel) fetchStreamData(ctx context.Context, streamName string) (widgets.StreamData, error) {
	// Get basic stream info
	streamInfo, err := m.redisClient.GetStreamInfo(ctx, streamName)
	if err != nil {
		return widgets.StreamData{}, err
	}
	
	// Calculate message rates
	messageRates := m.calculateMessageRates(streamName, streamInfo.Length)
	
	// Get groups for this stream
	groups, err := m.redisClient.GetStreamGroups(ctx, streamName)
	if err != nil {
		return widgets.StreamData{}, err
	}
	
	var consumerGroups []widgets.GroupData
	for _, group := range groups {
		consumers, err := m.redisClient.GetGroupConsumers(ctx, streamName, group.Name)
		if err != nil {
			log.Printf("Failed to get consumers for group %s: %v", group.Name, err)
			continue
		}
		
		var consumerData []widgets.ConsumerData
		for _, consumer := range consumers {
			consumerData = append(consumerData, widgets.ConsumerData{
				Name:    consumer.Name,
				Pending: consumer.Pending,
				Idle:    consumer.Idle,
			})
		}
		
		consumerGroups = append(consumerGroups, widgets.GroupData{
			Name:      group.Name,
			Stream:    streamName,
			Pending:   group.Pending,
			Consumers: consumerData,
		})
	}
	
	return widgets.StreamData{
		Name:           streamInfo.Name,
		Length:         streamInfo.Length,
		MemoryUsage:    streamInfo.MemoryUsage,
		Groups:         streamInfo.Groups,
		LastID:         streamInfo.LastGeneratedID,
		ConsumerGroups: consumerGroups,
		MessageRates:   messageRates,
	}, nil
}

// calculateMessageRates calculates message rates for sparklines
func (m *RootModel) calculateMessageRates(streamName string, currentLength int64) []float64 {
	const historySize = 20
	
	// Get or create history for this stream
	if _, exists := m.streamLengthHistory[streamName]; !exists {
		m.streamLengthHistory[streamName] = make([]int64, 0, historySize)
	}
	
	history := m.streamLengthHistory[streamName]
	
	// Add current length
	history = append(history, currentLength)
	if len(history) > historySize {
		history = history[1:]
	}
	m.streamLengthHistory[streamName] = history
	
	// Calculate rates
	rates := make([]float64, len(history))
	for i := 1; i < len(history); i++ {
		rate := float64(history[i] - history[i-1])
		if rate < 0 {
			rate = 0 // Handle resets
		}
		rates[i] = rate
	}
	
	return rates
}

// Helper functions

func findClosestRefreshRate(target time.Duration) int {
	closest := 0
	minDiff := absDuration(defaultRefreshRates[0] - target)
	
	for i, rate := range defaultRefreshRates {
		diff := absDuration(rate - target)
		if diff < minDiff {
			minDiff = diff
			closest = i
		}
	}
	
	return closest
}

func absDuration(d time.Duration) time.Duration {
	if d < 0 {
		return -d
	}
	return d
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
