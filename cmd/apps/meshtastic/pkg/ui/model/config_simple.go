package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
)

// SimpleConfigModel represents the configuration view model (simplified)
type SimpleConfigModel struct {
	styles *view.Styles
	width  int
	height int
	ready  bool

	// UI components
	table    table.Model
	viewport viewport.Model

	// Data
	config       *pb.LocalConfig
	moduleConfig *pb.LocalModuleConfig

	// Logging
	logger zerolog.Logger
}

// NewSimpleConfigModel creates a new simple configuration model
func NewSimpleConfigModel(styles *view.Styles) *SimpleConfigModel {
	// Create table
	columns := []table.Column{
		{Title: "Setting", Width: 30},
		{Title: "Value", Width: 50},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Create viewport
	vp := viewport.New(80, 10)

	// Create logger
	logger := log.With().Str("component", "config-ui").Logger()

	return &SimpleConfigModel{
		styles:   styles,
		table:    t,
		viewport: vp,
		logger:   logger,
	}
}

// Init initializes the configuration model
func (m *SimpleConfigModel) Init() tea.Cmd {
	m.logger.Debug().Msg("Initializing simple configuration model")
	return nil
}

// Update handles messages and updates the model
func (m *SimpleConfigModel) Update(msg tea.Msg) (*SimpleConfigModel, tea.Cmd) {
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

		m.logger.Debug().Int("width", m.width).Int("height", m.height).Msg("Config window resized")

	case tea.KeyMsg:
		switch msg.String() {
		case "r":
			m.refreshConfig()
		default:
			m.table, cmd = m.table.Update(msg)
			cmds = append(cmds, cmd)
		}

	case ConfigMsg:
		m.config = msg.Config
		m.moduleConfig = msg.ModuleConfig
		m.updateTable()
		m.logger.Debug().Msg("Received configuration update")
	}

	return m, tea.Batch(cmds...)
}

// View renders the configuration view
func (m *SimpleConfigModel) View() string {
	if !m.ready {
		return m.styles.StatusText.Render("Loading configuration...")
	}

	return m.tableView()
}

// tableView renders the table view
func (m *SimpleConfigModel) tableView() string {
	header := m.styles.Header.Render("Configuration")

	help := m.styles.Help.Render("↑/↓: navigate • r: refresh")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		m.table.View(),
		help,
	)

	return content
}

// updateTable updates the table with current configuration
func (m *SimpleConfigModel) updateTable() {
	var rows []table.Row

	if m.config != nil {
		rows = append(rows, table.Row{"Status", "Configuration loaded"})

		if m.config.Device != nil {
			rows = append(rows, table.Row{"Device Role", m.config.Device.GetRole().String()})
		}

		if m.config.Lora != nil {
			rows = append(rows, table.Row{"LoRa Region", m.config.Lora.GetRegion().String()})
			rows = append(rows, table.Row{"LoRa Preset", m.config.Lora.GetModemPreset().String()})
		}

		if m.config.Bluetooth != nil {
			rows = append(rows, table.Row{"Bluetooth", fmt.Sprintf("%t", m.config.Bluetooth.GetEnabled())})
		}
	} else {
		rows = append(rows, table.Row{"Status", "No configuration loaded"})
	}

	m.table.SetRows(rows)
}

// refreshConfig refreshes the configuration
func (m *SimpleConfigModel) refreshConfig() {
	m.logger.Info().Msg("Refreshing configuration")
	// This would typically trigger a refresh from the client
}

// ConfigMsg represents a configuration update message
type ConfigMsg struct {
	Config       *pb.LocalConfig
	ModuleConfig *pb.LocalModuleConfig
}

// UpdateConfig creates a config update command
func UpdateConfig(config *pb.LocalConfig, moduleConfig *pb.LocalModuleConfig) tea.Cmd {
	return func() tea.Msg {
		return ConfigMsg{
			Config:       config,
			ModuleConfig: moduleConfig,
		}
	}
}
