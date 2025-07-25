package components

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/sparkline"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// RedisClient interface for Redis operations
type RedisClient interface {
	GetServerInfo(ctx context.Context) (*ServerInfo, error)
	DiscoverStreams(ctx context.Context) ([]string, error)
	GetStreamInfo(ctx context.Context, name string) (*StreamInfo, error)
	GetStreamGroups(ctx context.Context, stream string) ([]GroupInfo, error)
	GetGroupConsumers(ctx context.Context, stream, group string) ([]ConsumerInfo, error)
}

// ServerInfo represents Redis server information
type ServerInfo struct {
	UptimeInSeconds   int64
	UsedMemory        int64
	TotalSystemMemory int64
	Version           string
}

// StreamInfo represents Redis stream information
type StreamInfo struct {
	Name            string
	Length          int64
	MemoryUsage     int64
	Groups          int64
	LastGeneratedID string
}

// GroupInfo represents Redis consumer group information
type GroupInfo struct {
	Name    string
	Pending int64
}

// ConsumerInfo represents Redis consumer information
type ConsumerInfo struct {
	Name    string
	Pending int64
	Idle    time.Duration
}

// Coordinator manages switching between views and coordinates updates
type Coordinator struct {
	client      RedisClient
	demoMode    bool
	refreshRate time.Duration

	// UI state
	width       int
	height      int
	currentView string // "streams", "groups", "metrics"
	lastUpdate  time.Time

	// Data
	streams    []models.StreamData
	serverData models.ServerData

	// Message rate tracking
	streamLengthHistory map[string][]int64 // Track stream lengths over time
	lengthHistorySize   int                // Number of data points to keep

	// Submodels
	streamsView    *StreamsView
	groupsView     *GroupsView
	metricsView    *MetricsView
	headerView     *HeaderView
	navigationView *NavigationView

	// Progress bars and sparklines for the comprehensive view
	memoryProgress    progress.Model
	globalSparkline   *sparkline.Sparkline
	streamSparklines  map[string]*sparkline.Sparkline
	throughputHistory []float64

	// Styles
	styles styles.Styles

	// Refresh rate control
	refreshRateIndex int // Current index in refreshRates array
}

// Available refresh rates in seconds (0.1s to 10s)
var refreshRates = []time.Duration{
	100 * time.Millisecond,  // 0.1s
	200 * time.Millisecond,  // 0.2s
	500 * time.Millisecond,  // 0.5s
	1 * time.Second,         // 1s
	1500 * time.Millisecond, // 1.5s
	2 * time.Second,         // 2s
	3 * time.Second,         // 3s
	5 * time.Second,         // 5s
	10 * time.Second,        // 10s
}

// tickMsg represents a refresh tick
type tickMsg struct {
	time time.Time
}

// dataMsg represents fetched data
type dataMsg struct {
	streams    []models.StreamData
	serverData models.ServerData
	err        error
}

// NewCoordinator creates a new coordinator model
func NewCoordinator(client RedisClient, demoMode bool, refreshRate time.Duration) *Coordinator {
	styles := styles.NewStyles()

	// Find the closest refresh rate index
	refreshRateIndex := findClosestRefreshRateIndex(refreshRate)

	// Initialize progress bar for memory usage
	memoryProgress := progress.New(progress.WithDefaultGradient())
	memoryProgress.Width = 40

	// Initialize global throughput sparkline
	globalSparkline := sparkline.New(sparkline.Config{
		Width:     30,
		Height:    1,
		MaxPoints: 30,
		Style:     sparkline.StyleBars,
	})

	return &Coordinator{
		client:           client,
		demoMode:         demoMode,
		refreshRate:      refreshRates[refreshRateIndex], // Use the actual rate from our list
		refreshRateIndex: refreshRateIndex,
		currentView:      "streams",
		styles:           styles,

		// Initialize message rate tracking
		streamLengthHistory: make(map[string][]int64),
		lengthHistorySize:   20, // Keep last 20 data points

		// Initialize progress bars and sparklines
		memoryProgress:    memoryProgress,
		globalSparkline:   globalSparkline,
		streamSparklines:  make(map[string]*sparkline.Sparkline),
		throughputHistory: make([]float64, 0, 30),

		// Initialize submodels
		streamsView:    NewStreamsView(styles),
		groupsView:     NewGroupsView(styles),
		metricsView:    NewMetricsView(styles),
		headerView:     NewHeaderView(styles),
		navigationView: NewNavigationView(styles),
	}
}

// Init implements tea.Model
func (c *Coordinator) Init() tea.Cmd {
	return tea.Batch(
		c.fetchData(),
		tea.Tick(c.refreshRate, func(t time.Time) tea.Msg {
			return tickMsg{time: t}
		}),
		c.streamsView.Init(),
		c.groupsView.Init(),
		c.metricsView.Init(),
		c.headerView.Init(),
		c.navigationView.Init(),
	)
}

// Update implements tea.Model
func (c *Coordinator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		c.width = msg.Width
		c.height = msg.Height

		// Update all submodels with window size
		_, cmd := c.streamsView.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = c.groupsView.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = c.metricsView.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = c.headerView.Update(msg)
		cmds = append(cmds, cmd)
		_, cmd = c.navigationView.Update(msg)
		cmds = append(cmds, cmd)

		return c, tea.Batch(cmds...)

	case tea.KeyMsg:
		keys := c.navigationView.GetKeys()

		switch {
		case key.Matches(msg, keys.Quit):
			return c, tea.Quit
		case key.Matches(msg, keys.Help):
			_, cmd := c.navigationView.Update(msg)
			return c, cmd
		case key.Matches(msg, keys.Refresh):
			return c, c.fetchData()
		case key.Matches(msg, keys.Streams):
			c.currentView = "streams"
			return c, c.updateHeaderView()
		case key.Matches(msg, keys.Groups):
			c.currentView = "groups"
			return c, c.updateHeaderView()
		case key.Matches(msg, keys.Metrics):
			c.currentView = "metrics"
			return c, c.updateHeaderView()
		case key.Matches(msg, keys.SpeedUp):
			return c, c.speedUp()
		case key.Matches(msg, keys.SpeedDown):
			return c, c.speedDown()
		}

		// Pass key events to the current view
		switch c.currentView {
		case "streams":
			_, cmd := c.streamsView.Update(msg)
			cmds = append(cmds, cmd)
		}

		return c, tea.Batch(cmds...)

	case tickMsg:
		return c, tea.Batch(
			c.fetchData(),
			tea.Tick(c.refreshRate, func(t time.Time) tea.Msg {
				return tickMsg{time: t}
			}),
		)

	case dataMsg:
		if msg.err == nil {
			c.streams = msg.streams
			c.serverData = msg.serverData
			c.lastUpdate = time.Now()

			// Update all submodels with new data
			_, cmd := c.streamsView.Update(StreamsDataMsg{Streams: msg.streams})
			cmds = append(cmds, cmd)
			_, cmd = c.groupsView.Update(GroupsDataMsg{Streams: msg.streams})
			cmds = append(cmds, cmd)
			_, cmd = c.metricsView.Update(MetricsDataMsg{ServerData: msg.serverData})
			cmds = append(cmds, cmd)
			cmds = append(cmds, c.updateHeaderView())
		}
		return c, tea.Batch(cmds...)
	}

	return c, tea.Batch(cmds...)
}

// View implements tea.Model - comprehensive layout matching design specification
func (c *Coordinator) View() string {
	if c.width == 0 {
		return "Loading..."
	}

	var sections []string

	// Header section with uptime and refresh rate
	header := c.renderHeader()
	sections = append(sections, header)

	// Main streams table with sparklines
	streamsTable := c.renderStreamsTable()
	sections = append(sections, streamsTable)

	// Groups detail section
	groupsDetail := c.renderGroupsDetail()
	sections = append(sections, groupsDetail)

	// Memory alerts and trim warnings
	memoryAlerts := c.renderMemoryAlerts()
	sections = append(sections, memoryAlerts)

	// Global metrics section
	globalMetrics := c.renderGlobalMetrics()
	sections = append(sections, globalMetrics)

	// Footer with commands
	footer := c.navigationView.View()
	sections = append(sections, footer)

	return strings.Join(sections, "\n\n")
}

// updateHeaderView sends updated header data to header view
func (c *Coordinator) updateHeaderView() tea.Cmd {
	_, cmd := c.headerView.Update(HeaderDataMsg{
		ServerData:  c.serverData,
		DemoMode:    c.demoMode,
		RefreshRate: c.refreshRate,
		CurrentView: c.currentView,
	})
	return cmd
}

// fetchData returns a command to fetch fresh data
func (c *Coordinator) fetchData() tea.Cmd {
	if c.demoMode {
		return c.fetchDemoData()
	}

	return func() tea.Msg {
		// Skip Redis operations in demo mode or if client is nil
		if c.demoMode || c.client == nil {
			return c.fetchDemoData()()
		}

		ctx := context.Background()

		// Fetch server data
		serverInfo, err := c.client.GetServerInfo(ctx)
		if err != nil {
			return dataMsg{err: err}
		}

		serverData := models.ServerData{
			Uptime:      time.Duration(serverInfo.UptimeInSeconds) * time.Second,
			MemoryUsed:  serverInfo.UsedMemory,
			MemoryTotal: serverInfo.TotalSystemMemory,
			Version:     serverInfo.Version,
		}

		// Fetch streams
		streamNames, err := c.client.DiscoverStreams(ctx)
		if err != nil {
			return dataMsg{err: err}
		}

		var streams []models.StreamData
		for _, name := range streamNames {
			info, err := c.client.GetStreamInfo(ctx, name)
			if err != nil {
				continue
			}

			groups, err := c.client.GetStreamGroups(ctx, name)
			if err != nil {
				continue
			}

			var groupData []models.GroupData
			for _, group := range groups {
				consumers, err := c.client.GetGroupConsumers(ctx, name, group.Name)
				if err != nil {
					continue
				}

				var consumerData []models.ConsumerData
				for _, consumer := range consumers {
					consumerData = append(consumerData, models.ConsumerData{
						Name:    consumer.Name,
						Pending: consumer.Pending,
						Idle:    consumer.Idle,
					})
				}

				groupData = append(groupData, models.GroupData{
					Name:      group.Name,
					Stream:    name,
					Consumers: consumerData,
					Pending:   group.Pending,
				})
			}

			streams = append(streams, models.StreamData{
				Name:           info.Name,
				Length:         info.Length,
				MemoryUsage:    info.MemoryUsage,
				Groups:         info.Groups,
				LastID:         info.LastGeneratedID,
				ConsumerGroups: groupData,
				MessageRates:   c.calculateMessageRates(info.Name, info.Length),
			})
		}

		return dataMsg{
			streams:    streams,
			serverData: serverData,
		}
	}
}

// fetchDemoData returns demo data
func (c *Coordinator) fetchDemoData() tea.Cmd {
	return func() tea.Msg {
		streams := []models.StreamData{
			{
				Name:         "orders",
				Length:       1243592,
				MemoryUsage:  126177280,
				Groups:       3,
				LastID:       "160123-7",
				MessageRates: []float64{0.8, 0.9, 0.7, 0.6, 0.4, 0.3, 0.2, 0.1, 0.1, 0.1},
				ConsumerGroups: []models.GroupData{
					{Name: "cg-1", Stream: "orders", Pending: 12, Consumers: []models.ConsumerData{
						{Name: "Alice", Pending: 3, Idle: 5 * time.Second},
						{Name: "Bob", Pending: 2, Idle: 1 * time.Second},
					}},
					{Name: "cg-2", Stream: "orders", Pending: 0, Consumers: []models.ConsumerData{
						{Name: "Charlie", Pending: 5, Idle: 10 * time.Second},
					}},
					{Name: "cg-3", Stream: "orders", Pending: 28, Consumers: []models.ConsumerData{
						{Name: "Dave", Pending: 1, Idle: 2*time.Minute + 12*time.Second},
					}},
				},
			},
			{
				Name:         "events",
				Length:       98234,
				MemoryUsage:  9663488,
				Groups:       5,
				LastID:       "160123-3",
				MessageRates: []float64{0.1, 0.1, 0.2, 0.3, 0.4, 0.6, 0.8, 0.9, 0.7, 0.5},
				ConsumerGroups: []models.GroupData{
					{Name: "cg-A", Stream: "events", Pending: 3, Consumers: []models.ConsumerData{
						{Name: "Eve", Pending: 4, Idle: 0},
						{Name: "Frank", Pending: 4, Idle: 30 * time.Second},
					}},
					{Name: "cg-B", Stream: "events", Pending: 47, Consumers: []models.ConsumerData{
						{Name: "Grace", Pending: 2, Idle: 1*time.Minute + 23*time.Second},
					}},
				},
			},
			{
				Name:         "logs",
				Length:       5432100,
				MemoryUsage:  536870912,
				Groups:       1,
				LastID:       "160122-9",
				MessageRates: []float64{1.0, 1.0, 1.0, 1.0, 1.0, 0.9, 0.8, 0.7, 0.6, 0.4},
				ConsumerGroups: []models.GroupData{
					{Name: "cg-logs", Stream: "logs", Pending: 0, Consumers: []models.ConsumerData{
						{Name: "Heidi", Pending: 10, Idle: 0},
					}},
				},
			},
		}

		serverData := models.ServerData{
			Uptime:      1098276 * time.Second, // ~12.7 days
			MemoryUsed:  348127232,             // ~332MB
			MemoryTotal: 1073741824,            // 1GB
			Version:     "7.0.8",
			Throughput:  1023.5,
		}

		return dataMsg{
			streams:    streams,
			serverData: serverData,
		}
	}
}

// calculateMessageRates calculates message rates for sparkline display
func (c *Coordinator) calculateMessageRates(streamName string, currentLength int64) []float64 {
	// Get or create history for this stream
	if _, exists := c.streamLengthHistory[streamName]; !exists {
		c.streamLengthHistory[streamName] = make([]int64, 0, c.lengthHistorySize)
	}

	history := c.streamLengthHistory[streamName]

	// Add current length to history
	history = append(history, currentLength)

	// Keep only the last lengthHistorySize entries
	if len(history) > c.lengthHistorySize {
		history = history[len(history)-c.lengthHistorySize:]
	}

	// Update the history
	c.streamLengthHistory[streamName] = history

	// Calculate rates (differences between consecutive measurements)
	rates := make([]float64, len(history))
	if len(history) > 1 {
		for i := 1; i < len(history); i++ {
			rate := float64(history[i] - history[i-1])
			if rate < 0 {
				rate = 0 // Handle stream resets/deletions
			}
			rates[i] = rate
		}
	}

	// If we don't have enough data, pad with zeros
	if len(rates) < 10 {
		padded := make([]float64, 10)
		copy(padded[10-len(rates):], rates)
		return padded
	}

	// Return the last 10 rates for display
	if len(rates) > 10 {
		return rates[len(rates)-10:]
	}

	return rates
}

// generateSparklineData creates mock sparkline data (kept for demo mode)
func generateSparklineData(length int) []float64 {
	data := make([]float64, length)
	for i := range data {
		data[i] = float64(i) / float64(length)
	}
	return data
}

// findClosestRefreshRateIndex finds the index of the closest refresh rate
func findClosestRefreshRateIndex(target time.Duration) int {
	minDiff := time.Duration(1<<63 - 1) // Max duration
	bestIndex := 3                      // Default to 1 second (index 3)

	for i, rate := range refreshRates {
		diff := target - rate
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			bestIndex = i
		}
	}

	return bestIndex
}

// speedUp increases the refresh rate (makes it faster)
func (c *Coordinator) speedUp() tea.Cmd {
	if c.refreshRateIndex > 0 {
		c.refreshRateIndex--
		c.refreshRate = refreshRates[c.refreshRateIndex]
		return tea.Batch(
			c.updateHeaderView(),
			// Start new ticker with faster rate
			tea.Tick(c.refreshRate, func(t time.Time) tea.Msg {
				return tickMsg{time: t}
			}),
		)
	}
	return nil
}

// speedDown decreases the refresh rate (makes it slower)
func (c *Coordinator) speedDown() tea.Cmd {
	if c.refreshRateIndex < len(refreshRates)-1 {
		c.refreshRateIndex++
		c.refreshRate = refreshRates[c.refreshRateIndex]
		return tea.Batch(
			c.updateHeaderView(),
			// Start new ticker with slower rate
			tea.Tick(c.refreshRate, func(t time.Time) tea.Msg {
				return tickMsg{time: t}
			}),
		)
	}
	return nil
}
