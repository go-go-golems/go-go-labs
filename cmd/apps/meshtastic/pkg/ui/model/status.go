package model

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
)

// DeviceStatus represents the status of the device
type DeviceStatus struct {
	// Device info
	DeviceID        string
	NodeName        string
	LongName        string
	HardwareModel   string
	FirmwareVersion string

	// Connection info
	Connected   bool
	SerialPort  string
	BaudRate    int
	ConnectedAt time.Time

	// Radio info
	Region          string
	Frequency       float64
	Bandwidth       int
	SpreadingFactor int
	CodingRate      int
	TxPower         int

	// Network info
	Channel     int
	ChannelName string
	PSK         string
	ModemPreset string

	// Hardware status
	Battery     int
	Voltage     float64
	Temperature float64
	Uptime      time.Duration

	// Statistics
	MessagesReceived int
	MessagesSent     int
	NodesCount       int
	LastActivity     time.Time
}

// StatusModel handles the status view
type StatusModel struct {
	styles   *view.Styles
	viewport viewport.Model
	status   DeviceStatus
	width    int
	height   int
	ready    bool

	// Update ticker
	lastUpdate time.Time
}

// NewStatusModel creates a new status model
func NewStatusModel(styles *view.Styles) *StatusModel {
	vp := viewport.New(0, 0)

	return &StatusModel{
		styles:   styles,
		viewport: vp,
		status:   getDefaultStatus(),
	}
}

// getDefaultStatus returns a default status with sample data
func getDefaultStatus() DeviceStatus {
	return DeviceStatus{
		DeviceID:        "!1234567890",
		NodeName:        "MyNode",
		LongName:        "My Meshtastic Node",
		HardwareModel:   "TTGO LoRa32",
		FirmwareVersion: "2.3.2",

		Connected:   true,
		SerialPort:  "/dev/ttyUSB0",
		BaudRate:    921600,
		ConnectedAt: time.Now().Add(-30 * time.Minute),

		Region:          "US",
		Frequency:       915.0,
		Bandwidth:       125,
		SpreadingFactor: 7,
		CodingRate:      5,
		TxPower:         20,

		Channel:     0,
		ChannelName: "LongFast",
		PSK:         "AQ==",
		ModemPreset: "LongFast",

		Battery:     85,
		Voltage:     4.1,
		Temperature: 25.5,
		Uptime:      48 * time.Hour,

		MessagesReceived: 142,
		MessagesSent:     73,
		NodesCount:       4,
		LastActivity:     time.Now().Add(-2 * time.Minute),
	}
}

// Init initializes the status model
func (m *StatusModel) Init() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

// Update handles messages and updates the model
func (m *StatusModel) Update(msg tea.Msg) (*StatusModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2 // Account for title
		m.ready = true
		m.updateContent()

	case TickMsg:
		// Update uptime and other time-based fields
		m.status.Uptime += time.Second
		m.updateContent()

		// Schedule next tick
		return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return TickMsg(t)
		})

	case StatusUpdateMsg:
		m.status = DeviceStatus(msg)
		m.updateContent()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the status view
func (m *StatusModel) View() string {
	if !m.ready {
		return "Loading status..."
	}

	title := m.styles.Title.Render("Device Status")
	content := m.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

// updateContent updates the viewport content
func (m *StatusModel) updateContent() {
	if !m.ready {
		return
	}

	var sections []string

	// Device Information
	deviceInfo := m.renderSection("Device Information", []StatusItem{
		{"Device ID", m.status.DeviceID},
		{"Node Name", m.status.NodeName},
		{"Long Name", m.status.LongName},
		{"Hardware", m.status.HardwareModel},
		{"Firmware", m.status.FirmwareVersion},
	})
	sections = append(sections, deviceInfo)

	// Connection Status
	connectionStatus := "Disconnected"
	if m.status.Connected {
		connectionStatus = "Connected"
	}

	connectedSince := ""
	if m.status.Connected {
		connectedSince = fmt.Sprintf("Since: %s", m.status.ConnectedAt.Format("15:04:05"))
	}

	connectionInfo := m.renderSection("Connection", []StatusItem{
		{"Status", connectionStatus},
		{"Serial Port", m.status.SerialPort},
		{"Baud Rate", fmt.Sprintf("%d", m.status.BaudRate)},
		{"Connected", connectedSince},
	})
	sections = append(sections, connectionInfo)

	// Radio Configuration
	radioInfo := m.renderSection("Radio Configuration", []StatusItem{
		{"Region", m.status.Region},
		{"Frequency", fmt.Sprintf("%.1f MHz", m.status.Frequency)},
		{"Bandwidth", fmt.Sprintf("%d kHz", m.status.Bandwidth)},
		{"Spreading Factor", fmt.Sprintf("%d", m.status.SpreadingFactor)},
		{"Coding Rate", fmt.Sprintf("4/%d", m.status.CodingRate)},
		{"TX Power", fmt.Sprintf("%d dBm", m.status.TxPower)},
	})
	sections = append(sections, radioInfo)

	// Network Configuration
	networkInfo := m.renderSection("Network", []StatusItem{
		{"Channel", fmt.Sprintf("%d (%s)", m.status.Channel, m.status.ChannelName)},
		{"Modem Preset", m.status.ModemPreset},
		{"PSK", m.status.PSK},
	})
	sections = append(sections, networkInfo)

	// Hardware Status
	batteryStatus := fmt.Sprintf("%d%%", m.status.Battery)
	if m.status.Battery == 0 {
		batteryStatus = "External Power"
	}

	hardwareInfo := m.renderSection("Hardware Status", []StatusItem{
		{"Battery", batteryStatus},
		{"Voltage", fmt.Sprintf("%.1f V", m.status.Voltage)},
		{"Temperature", fmt.Sprintf("%.1f °C", m.status.Temperature)},
		{"Uptime", formatUptime(m.status.Uptime)},
	})
	sections = append(sections, hardwareInfo)

	// Statistics
	statsInfo := m.renderSection("Statistics", []StatusItem{
		{"Messages Received", fmt.Sprintf("%d", m.status.MessagesReceived)},
		{"Messages Sent", fmt.Sprintf("%d", m.status.MessagesSent)},
		{"Nodes Count", fmt.Sprintf("%d", m.status.NodesCount)},
		{"Last Activity", formatDuration(time.Since(m.status.LastActivity))},
	})
	sections = append(sections, statsInfo)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)
	m.viewport.SetContent(content)
}

// StatusItem represents a status key-value pair
type StatusItem struct {
	Key   string
	Value string
}

// renderSection renders a status section
func (m *StatusModel) renderSection(title string, items []StatusItem) string {
	sectionTitle := m.styles.Subtitle.Render(title)

	var itemStrings []string
	for _, item := range items {
		if item.Value == "" {
			continue
		}

		key := m.styles.StatusKey.Render(item.Key + ":")
		value := m.styles.StatusValue.Render(item.Value)

		itemStr := m.styles.StatusItem.Render(
			lipgloss.JoinHorizontal(lipgloss.Left, key, " ", value),
		)
		itemStrings = append(itemStrings, itemStr)
	}

	section := lipgloss.JoinVertical(lipgloss.Left, sectionTitle)
	if len(itemStrings) > 0 {
		section = lipgloss.JoinVertical(lipgloss.Left, section,
			lipgloss.JoinVertical(lipgloss.Left, itemStrings...))
	}

	return section + "\n"
}

// formatUptime formats uptime in a human-readable way
func formatUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm %ds", days, hours, minutes, seconds)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// GetStatus returns the current status
func (m *StatusModel) GetStatus() DeviceStatus {
	return m.status
}

// SetStatus updates the status
func (m *StatusModel) SetStatus(status DeviceStatus) {
	m.status = status
	m.updateContent()
}

// UpdateStatus updates specific status fields
func (m *StatusModel) UpdateStatus(update StatusUpdateMsg) {
	m.status = DeviceStatus(update)
	m.updateContent()
}

// TickMsg is sent every second for time updates
type TickMsg time.Time

// StatusUpdateMsg is sent when status is updated
type StatusUpdateMsg DeviceStatus

// StatusUpdate creates a status update message
func StatusUpdate(status DeviceStatus) tea.Cmd {
	return func() tea.Msg {
		return StatusUpdateMsg(status)
	}
}
