package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// StreamData represents data for a single stream
type StreamData struct {
	Name            string
	Length          int64
	MemoryUsage     int64
	Groups          int64
	LastID          string
	ConsumerGroups  []GroupData
	MessageRates    []float64 // for sparkline
}

// GroupData represents data for a consumer group
type GroupData struct {
	Name      string
	Stream    string
	Consumers []ConsumerData
	Pending   int64
}

// ConsumerData represents data for a single consumer
type ConsumerData struct {
	Name    string
	Pending int64
	Idle    time.Duration
}

// ServerData represents Redis server information
type ServerData struct {
	Uptime       time.Duration
	MemoryUsed   int64
	MemoryTotal  int64
	Version      string
	Throughput   float64
}

// Model represents the TUI state
type Model struct {
	client      *RedisClient
	demoMode    bool
	refreshRate time.Duration
	
	// UI state
	width       int
	height      int
	currentView string // "streams", "groups", "metrics"
	selectedIdx int
	
	// Data
	streams    []StreamData
	serverData ServerData
	lastUpdate time.Time
	
	// UI components
	help     help.Model
	keys     keyMap
	
	// Styles
	styles Styles
}

// Styles contains all the lipgloss styles
type Styles struct {
	Header       lipgloss.Style
	Title        lipgloss.Style
	Status       lipgloss.Style
	StreamTable  lipgloss.Style
	GroupTable   lipgloss.Style
	Sparkline    lipgloss.Style
	Border       lipgloss.Style
	Selected     lipgloss.Style
	Unselected   lipgloss.Style
	Memory       lipgloss.Style
	Throughput   lipgloss.Style
}

// NewStyles creates the default styles
func NewStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),
		
		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
		
		StreamTable: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1),
		
		GroupTable: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F25D94")).
			Padding(1),
		
		Sparkline: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
		
		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")),
		
		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),
		
		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
		
		Memory: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F25D94")),
		
		Throughput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
	}
}

// keyMap defines the key bindings
type keyMap struct {
	Refresh key.Binding
	Groups  key.Binding
	Streams key.Binding
	Metrics key.Binding
	Up      key.Binding
	Down    key.Binding
	Quit    key.Binding
	Help    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Refresh, k.Groups, k.Streams, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Refresh, k.Groups, k.Streams, k.Metrics},
		{k.Up, k.Down, k.Quit, k.Help},
	}
}

var keys = keyMap{
	Refresh: key.NewBinding(
		key.WithKeys("r", "R"),
		key.WithHelp("r", "refresh"),
	),
	Groups: key.NewBinding(
		key.WithKeys("g", "G"),
		key.WithHelp("g", "groups view"),
	),
	Streams: key.NewBinding(
		key.WithKeys("s", "S"),
		key.WithHelp("s", "streams view"),
	),
	Metrics: key.NewBinding(
		key.WithKeys("m", "M"),
		key.WithHelp("m", "metrics view"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
}

// NewModel creates a new TUI model
func NewModel(client *RedisClient, demoMode bool, refreshRate time.Duration) Model {
	h := help.New()
	h.ShowAll = false
	
	return Model{
		client:      client,
		demoMode:    demoMode,
		refreshRate: refreshRate,
		currentView: "streams",
		help:        h,
		keys:        keys,
		styles:      NewStyles(),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.fetchData(),
		tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
			return tickMsg{time: t}
		}),
	)
}

// tickMsg represents a refresh tick
type tickMsg struct {
	time time.Time
}

// dataMsg represents fetched data
type dataMsg struct {
	streams    []StreamData
	serverData ServerData
	err        error
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		return m, nil
		
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Refresh):
			return m, m.fetchData()
		case key.Matches(msg, m.keys.Streams):
			m.currentView = "streams"
			return m, nil
		case key.Matches(msg, m.keys.Groups):
			m.currentView = "groups"
			return m, nil
		case key.Matches(msg, m.keys.Metrics):
			m.currentView = "metrics"
			return m, nil
		case key.Matches(msg, m.keys.Up):
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return m, nil
		case key.Matches(msg, m.keys.Down):
			maxIdx := len(m.streams) - 1
			if m.selectedIdx < maxIdx {
				m.selectedIdx++
			}
			return m, nil
		}
		
	case tickMsg:
		return m, tea.Batch(
			m.fetchData(),
			tea.Tick(m.refreshRate, func(t time.Time) tea.Msg {
				return tickMsg{time: t}
			}),
		)
		
	case dataMsg:
		if msg.err == nil {
			m.streams = msg.streams
			m.serverData = msg.serverData
			m.lastUpdate = time.Now()
		}
		return m, nil
	}
	
	return m, nil
}

// View implements tea.Model
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}
	
	var content strings.Builder
	
	// Header
	header := m.renderHeader()
	content.WriteString(header + "\n\n")
	
	// Main content based on current view
	switch m.currentView {
	case "streams":
		content.WriteString(m.renderStreamsView())
	case "groups":
		content.WriteString(m.renderGroupsView())
	case "metrics":
		content.WriteString(m.renderMetricsView())
	}
	
	// Footer with help
	footer := "\n" + m.help.View(m.keys)
	content.WriteString(footer)
	
	return content.String()
}

// fetchData returns a command to fetch fresh data
func (m Model) fetchData() tea.Cmd {
	if m.demoMode {
		return m.fetchDemoData()
	}
	
	return func() tea.Msg {
		ctx := context.Background()
		
		// Fetch server data
		serverInfo, err := m.client.GetServerInfo(ctx)
		if err != nil {
			return dataMsg{err: err}
		}
		
		serverData := ServerData{
			Uptime:      time.Duration(serverInfo.UptimeInSeconds) * time.Second,
			MemoryUsed:  serverInfo.UsedMemory,
			MemoryTotal: serverInfo.TotalSystemMemory,
			Version:     serverInfo.Version,
		}
		
		// Fetch streams
		streamNames, err := m.client.DiscoverStreams(ctx)
		if err != nil {
			return dataMsg{err: err}
		}
		
		var streams []StreamData
		for _, name := range streamNames {
			info, err := m.client.GetStreamInfo(ctx, name)
			if err != nil {
				continue
			}
			
			groups, err := m.client.GetStreamGroups(ctx, name)
			if err != nil {
				continue
			}
			
			var groupData []GroupData
			for _, group := range groups {
				consumers, err := m.client.GetGroupConsumers(ctx, name, group.Name)
				if err != nil {
					continue
				}
				
				var consumerData []ConsumerData
				for _, consumer := range consumers {
					consumerData = append(consumerData, ConsumerData{
						Name:    consumer.Name,
						Pending: consumer.Pending,
						Idle:    consumer.Idle,
					})
				}
				
				groupData = append(groupData, GroupData{
					Name:      group.Name,
					Stream:    name,
					Consumers: consumerData,
					Pending:   group.Pending,
				})
			}
			
			streams = append(streams, StreamData{
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
func (m Model) fetchDemoData() tea.Cmd {
	return func() tea.Msg {
		streams := []StreamData{
			{
				Name:        "orders",
				Length:      1243592,
				MemoryUsage: 126177280,
				Groups:      3,
				LastID:      "160123-7",
				MessageRates: []float64{0.8, 0.9, 0.7, 0.6, 0.4, 0.3, 0.2, 0.1, 0.1, 0.1},
				ConsumerGroups: []GroupData{
					{Name: "cg-1", Stream: "orders", Pending: 12, Consumers: []ConsumerData{
						{Name: "Alice", Pending: 3, Idle: 5 * time.Second},
						{Name: "Bob", Pending: 2, Idle: 1 * time.Second},
					}},
					{Name: "cg-2", Stream: "orders", Pending: 0, Consumers: []ConsumerData{
						{Name: "Charlie", Pending: 5, Idle: 10 * time.Second},
					}},
					{Name: "cg-3", Stream: "orders", Pending: 28, Consumers: []ConsumerData{
						{Name: "Dave", Pending: 1, Idle: 2*time.Minute + 12*time.Second},
					}},
				},
			},
			{
				Name:        "events",
				Length:      98234,
				MemoryUsage: 9663488,
				Groups:      5,
				LastID:      "160123-3",
				MessageRates: []float64{0.1, 0.1, 0.2, 0.3, 0.4, 0.6, 0.8, 0.9, 0.7, 0.5},
				ConsumerGroups: []GroupData{
					{Name: "cg-A", Stream: "events", Pending: 3, Consumers: []ConsumerData{
						{Name: "Eve", Pending: 4, Idle: 0},
						{Name: "Frank", Pending: 4, Idle: 30 * time.Second},
					}},
					{Name: "cg-B", Stream: "events", Pending: 47, Consumers: []ConsumerData{
						{Name: "Grace", Pending: 2, Idle: 1*time.Minute + 23*time.Second},
					}},
				},
			},
			{
				Name:        "logs",
				Length:      5432100,
				MemoryUsage: 536870912,
				Groups:      1,
				LastID:      "160122-9",
				MessageRates: []float64{1.0, 1.0, 1.0, 1.0, 1.0, 0.9, 0.8, 0.7, 0.6, 0.4},
				ConsumerGroups: []GroupData{
					{Name: "cg-logs", Stream: "logs", Pending: 0, Consumers: []ConsumerData{
						{Name: "Heidi", Pending: 10, Idle: 0},
					}},
				},
			},
		}
		
		serverData := ServerData{
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

// renderHeader renders the main header
func (m Model) renderHeader() string {
	var status string
	if m.demoMode {
		status = "DEMO MODE"
	} else {
		status = fmt.Sprintf("Uptime: %s", FormatDuration(m.serverData.Uptime))
	}
	
	title := m.styles.Header.Render("Redis Streams Monitor (top-like)")
	statusLine := m.styles.Status.Render(fmt.Sprintf("%s | Refresh: %s | Memory: %s",
		status,
		m.refreshRate,
		FormatBytes(m.serverData.MemoryUsed)))
	
	return lipgloss.JoinVertical(lipgloss.Left, title, statusLine)
}

// renderStreamsView renders the streams overview
func (m Model) renderStreamsView() string {
	var rows []string
	header := fmt.Sprintf("%-15s %-10s %-10s %-8s %-15s %s",
		"Stream", "Entries", "Memory", "Groups", "Last ID", "Msg/s")
	rows = append(rows, m.styles.Selected.Render(header))
	
	for i, stream := range m.streams {
		sparkline := renderSparkline(stream.MessageRates)
		row := fmt.Sprintf("%-15s %-10d %-10s %-8d %-15s %s",
			truncateString(stream.Name, 15),
			stream.Length,
			FormatBytes(stream.MemoryUsage),
			stream.Groups,
			truncateString(stream.LastID, 15),
			sparkline)
		
		if i == m.selectedIdx {
			rows = append(rows, m.styles.Selected.Render(row))
		} else {
			rows = append(rows, m.styles.Unselected.Render(row))
		}
	}
	
	content := strings.Join(rows, "\n")
	return m.styles.StreamTable.Render(content)
}

// renderGroupsView renders the consumer groups view
func (m Model) renderGroupsView() string {
	var rows []string
	header := fmt.Sprintf("%-12s %-12s %-15s %-8s %-12s",
		"Group", "Stream", "Consumers", "Pending", "Idle Time")
	rows = append(rows, m.styles.Selected.Render(header))
	
	for _, stream := range m.streams {
		for _, group := range stream.ConsumerGroups {
			var consumerNames []string
			for _, consumer := range group.Consumers {
				consumerNames = append(consumerNames, fmt.Sprintf("%s(%d)", consumer.Name, consumer.Pending))
			}
			
			row := fmt.Sprintf("%-12s %-12s %-15s %-8d %-12s",
				truncateString(group.Name, 12),
				truncateString(group.Stream, 12),
				truncateString(strings.Join(consumerNames, " "), 15),
				group.Pending,
				"00:00:05") // Mock for now
			
			rows = append(rows, m.styles.Unselected.Render(row))
		}
	}
	
	content := strings.Join(rows, "\n")
	return m.styles.GroupTable.Render(content)
}

// renderMetricsView renders the metrics overview
func (m Model) renderMetricsView() string {
	memoryPercent := float64(m.serverData.MemoryUsed) / float64(m.serverData.MemoryTotal) * 100
	memoryBar := renderProgressBar(memoryPercent, 30)
	
	throughputSparkline := renderSparkline([]float64{0.2, 0.4, 0.6, 0.8, 1.0, 0.9, 0.7, 0.8, 0.9, 1.0})
	
	content := fmt.Sprintf(`Server Information:
  Redis Version: %s
  Uptime: %s
  
Memory Usage:
  Used: %s / %s (%.1f%%)
  %s
  
Global Throughput: %.1f msg/s
%s`,
		m.serverData.Version,
		FormatDuration(m.serverData.Uptime),
		FormatBytes(m.serverData.MemoryUsed),
		FormatBytes(m.serverData.MemoryTotal),
		memoryPercent,
		m.styles.Memory.Render(memoryBar),
		m.serverData.Throughput,
		m.styles.Throughput.Render(throughputSparkline))
	
	return m.styles.StreamTable.Render(content)
}

// renderSparkline creates a text-based sparkline
func renderSparkline(data []float64) string {
	if len(data) == 0 {
		return ""
	}
	
	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var result strings.Builder
	
	for _, value := range data {
		if value <= 0 {
			result.WriteRune(' ')
		} else if value >= 1 {
			result.WriteRune(bars[len(bars)-1])
		} else {
			idx := int(value * float64(len(bars)-1))
			result.WriteRune(bars[idx])
		}
	}
	
	return result.String()
}

// renderProgressBar creates a progress bar
func renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	
	var bar strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteRune('■')
		} else {
			bar.WriteRune('□')
		}
	}
	
	return bar.String()
}

// truncateString truncates a string to a given length with ellipsis
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
