package components

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
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

	// Submodels
	streamsView    *StreamsView
	groupsView     *GroupsView
	metricsView    *MetricsView
	headerView     *HeaderView
	navigationView *NavigationView

	// Styles
	styles styles.Styles
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

	return &Coordinator{
		client:      client,
		demoMode:    demoMode,
		refreshRate: refreshRate,
		currentView: "streams",
		styles:      styles,

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

// View implements tea.Model
func (c *Coordinator) View() string {
	if c.width == 0 {
		return "Loading..."
	}

	// Header
	header := c.headerView.View()

	// Main content based on current view
	var content string
	switch c.currentView {
	case "streams":
		content = c.streamsView.View()
	case "groups":
		content = c.groupsView.View()
	case "metrics":
		content = c.metricsView.View()
	}

	// Footer with help
	footer := c.navigationView.View()

	return header + "\n\n" + content + "\n" + footer
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
				MessageRates:   generateSparklineData(10), // Mock for now
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

// generateSparklineData creates mock sparkline data
func generateSparklineData(length int) []float64 {
	data := make([]float64, length)
	for i := range data {
		data[i] = float64(i) / float64(length)
	}
	return data
}
