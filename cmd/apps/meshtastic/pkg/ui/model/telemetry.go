package model

import (
	"fmt"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
)

// TelemetryData represents a telemetry data point
type TelemetryData struct {
	NodeID    uint32
	NodeName  string
	Timestamp time.Time
	Type      string
	Data      interface{}
}

// TelemetryModel represents the telemetry view model
type TelemetryModel struct {
	styles *view.Styles
	width  int
	height int
	ready  bool

	// UI components
	table    table.Model
	viewport viewport.Model

	// Data
	telemetryData []TelemetryData
	selectedNode  uint32

	// View state
	showDetails bool

	// Logging
	logger zerolog.Logger
}

// NewTelemetryModel creates a new telemetry model
func NewTelemetryModel(styles *view.Styles) *TelemetryModel {
	// Create table
	columns := []table.Column{
		{Title: "Node", Width: 20},
		{Title: "Type", Width: 15},
		{Title: "Time", Width: 10},
		{Title: "Data", Width: 40},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Create viewport
	vp := viewport.New(80, 10)

	// Create logger
	logger := log.With().Str("component", "telemetry-ui").Logger()

	return &TelemetryModel{
		styles:        styles,
		table:         t,
		viewport:      vp,
		telemetryData: make([]TelemetryData, 0),
		logger:        logger,
	}
}

// Init initializes the telemetry model
func (m *TelemetryModel) Init() tea.Cmd {
	m.logger.Debug().Msg("Initializing telemetry model")
	return nil
}

// Update handles messages and updates the model
func (m *TelemetryModel) Update(msg tea.Msg) (*TelemetryModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update table dimensions
		m.table.SetWidth(m.width)
		m.table.SetHeight(m.height - 8) // Account for header and footer

		// Update viewport dimensions
		m.viewport.Width = m.width
		m.viewport.Height = m.height - 8

		m.logger.Debug().Int("width", m.width).Int("height", m.height).Msg("Telemetry window resized")

	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			m.showDetails = !m.showDetails
			m.logger.Debug().Bool("show_details", m.showDetails).Msg("Toggled details view")
		case "r":
			m.refreshData()
		case "c":
			m.clearData()
		default:
			if m.showDetails {
				m.viewport, cmd = m.viewport.Update(msg)
				cmds = append(cmds, cmd)
			} else {
				m.table, cmd = m.table.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case TelemetryMsg:
		m.addTelemetryData(msg.NodeID, msg.NodeName, msg.Type, msg.Data)
		m.updateTable()
		m.logger.Debug().Uint32("node_id", msg.NodeID).Str("type", msg.Type).Msg("Received telemetry data")

	case RefreshTelemetryMsg:
		m.refreshData()
		m.logger.Debug().Msg("Refreshing telemetry data")
	}

	return m, tea.Batch(cmds...)
}

// View renders the telemetry view
func (m *TelemetryModel) View() string {
	if !m.ready {
		return m.styles.StatusText.Render("Loading telemetry...")
	}

	if m.showDetails {
		return m.detailsView()
	}

	return m.tableView()
}

// tableView renders the table view
func (m *TelemetryModel) tableView() string {
	header := m.styles.Header.Render("Telemetry Data")

	help := m.styles.Help.Render("↑/↓: navigate • d: details • r: refresh • c: clear")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.table.View(),
		help,
	)

	return content
}

// detailsView renders the details view
func (m *TelemetryModel) detailsView() string {
	header := m.styles.Header.Render("Telemetry Details")

	var content string
	if m.selectedNode != 0 {
		content = m.getNodeTelemetryDetails(m.selectedNode)
	} else {
		content = "No node selected"
	}

	m.viewport.SetContent(content)

	help := m.styles.Help.Render("↑/↓: scroll • d: back to table • r: refresh • c: clear")

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.viewport.View(),
		help,
	)

	return view
}

// addTelemetryData adds new telemetry data
func (m *TelemetryModel) addTelemetryData(nodeID uint32, nodeName, telemetryType string, data interface{}) {
	telemetryData := TelemetryData{
		NodeID:    nodeID,
		NodeName:  nodeName,
		Timestamp: time.Now(),
		Type:      telemetryType,
		Data:      data,
	}

	m.telemetryData = append(m.telemetryData, telemetryData)

	// Keep only last 1000 entries
	if len(m.telemetryData) > 1000 {
		m.telemetryData = m.telemetryData[len(m.telemetryData)-1000:]
	}

	// Sort by timestamp (newest first)
	sort.Slice(m.telemetryData, func(i, j int) bool {
		return m.telemetryData[i].Timestamp.After(m.telemetryData[j].Timestamp)
	})
}

// updateTable updates the table with current telemetry data
func (m *TelemetryModel) updateTable() {
	rows := make([]table.Row, 0)

	for _, td := range m.telemetryData {
		nodeName := td.NodeName
		if nodeName == "" {
			nodeName = fmt.Sprintf("Node %d", td.NodeID)
		}

		timeStr := td.Timestamp.Format("15:04:05")
		dataStr := m.formatTelemetryData(td.Type, td.Data)

		rows = append(rows, table.Row{
			nodeName,
			td.Type,
			timeStr,
			dataStr,
		})
	}

	m.table.SetRows(rows)
}

// formatTelemetryData formats telemetry data for display
func (m *TelemetryModel) formatTelemetryData(telemetryType string, data interface{}) string {
	switch telemetryType {
	case "device":
		if telemetry, ok := data.(*pb.DeviceMetrics); ok {
			return fmt.Sprintf("Battery: %.1f%%, Voltage: %.2fV, Util: %.1f%%",
				float64(telemetry.GetBatteryLevel()), telemetry.GetVoltage(), telemetry.GetChannelUtilization())
		}
	case "environment":
		if telemetry, ok := data.(*pb.EnvironmentMetrics); ok {
			return fmt.Sprintf("Temp: %.1f°C, Humidity: %.1f%%, Pressure: %.1fhPa",
				telemetry.GetTemperature(), telemetry.GetRelativeHumidity(), telemetry.GetBarometricPressure())
		}
	case "power":
		if telemetry, ok := data.(*pb.PowerMetrics); ok {
			return fmt.Sprintf("CH1: %.2fV, CH2: %.2fV, CH3: %.2fV",
				telemetry.GetCh1Voltage(), telemetry.GetCh2Voltage(), telemetry.GetCh3Voltage())
		}
	}

	return fmt.Sprintf("%v", data)
}

// getNodeTelemetryDetails gets detailed telemetry for a specific node
func (m *TelemetryModel) getNodeTelemetryDetails(nodeID uint32) string {
	var content string

	nodeName := fmt.Sprintf("Node %d", nodeID)
	for _, td := range m.telemetryData {
		if td.NodeID == nodeID && td.NodeName != "" {
			nodeName = td.NodeName
			break
		}
	}

	content += fmt.Sprintf("=== %s ===\n\n", nodeName)

	// Group by type
	types := make(map[string][]TelemetryData)
	for _, td := range m.telemetryData {
		if td.NodeID == nodeID {
			types[td.Type] = append(types[td.Type], td)
		}
	}

	for telemetryType, dataList := range types {
		content += fmt.Sprintf("## %s\n", telemetryType)
		for _, td := range dataList {
			content += fmt.Sprintf("%s: %s\n",
				td.Timestamp.Format("15:04:05"),
				m.formatTelemetryData(td.Type, td.Data))
		}
		content += "\n"
	}

	return content
}

// refreshData refreshes the telemetry data
func (m *TelemetryModel) refreshData() {
	// This would typically trigger a refresh from the client
	m.logger.Info().Msg("Refreshing telemetry data")
}

// clearData clears all telemetry data
func (m *TelemetryModel) clearData() {
	m.telemetryData = make([]TelemetryData, 0)
	m.updateTable()
	m.logger.Info().Msg("Cleared telemetry data")
}

// Messages for telemetry updates
type TelemetryMsg struct {
	NodeID   uint32
	NodeName string
	Type     string
	Data     interface{}
}

type RefreshTelemetryMsg struct{}

// Commands for telemetry
func UpdateTelemetry(nodeID uint32, nodeName, telemetryType string, data interface{}) tea.Cmd {
	return func() tea.Msg {
		return TelemetryMsg{
			NodeID:   nodeID,
			NodeName: nodeName,
			Type:     telemetryType,
			Data:     data,
		}
	}
}

func RefreshTelemetry() tea.Cmd {
	return func() tea.Msg {
		return RefreshTelemetryMsg{}
	}
}
