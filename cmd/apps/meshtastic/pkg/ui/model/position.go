package model

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	pb "github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/pb"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
)

// PositionData represents a position data point
type PositionData struct {
	NodeID    uint32
	NodeName  string
	Timestamp time.Time
	Position  *pb.Position
	Distance  float64 // Distance from our position
}

// PositionModel represents the position view model
type PositionModel struct {
	styles *view.Styles
	width  int
	height int
	ready  bool

	// UI components
	table    table.Model
	latInput textinput.Model
	lonInput textinput.Model
	altInput textinput.Model

	// Data
	positions  []PositionData
	myPosition *pb.Position

	// View state
	settingPosition bool
	currentInput    int // 0=lat, 1=lon, 2=alt

	// Logging
	logger zerolog.Logger
}

// NewPositionModel creates a new position model
func NewPositionModel(styles *view.Styles) *PositionModel {
	// Create table
	columns := []table.Column{
		{Title: "Node", Width: 20},
		{Title: "Latitude", Width: 12},
		{Title: "Longitude", Width: 12},
		{Title: "Altitude", Width: 10},
		{Title: "Distance", Width: 10},
		{Title: "Time", Width: 10},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	// Create text inputs
	latInput := textinput.New()
	latInput.Placeholder = "Latitude (e.g., 37.7749)"
	latInput.CharLimit = 20

	lonInput := textinput.New()
	lonInput.Placeholder = "Longitude (e.g., -122.4194)"
	lonInput.CharLimit = 20

	altInput := textinput.New()
	altInput.Placeholder = "Altitude (meters)"
	altInput.CharLimit = 10

	// Create logger
	logger := log.With().Str("component", "position-ui").Logger()

	return &PositionModel{
		styles:    styles,
		table:     t,
		latInput:  latInput,
		lonInput:  lonInput,
		altInput:  altInput,
		positions: make([]PositionData, 0),
		logger:    logger,
	}
}

// Init initializes the position model
func (m *PositionModel) Init() tea.Cmd {
	m.logger.Debug().Msg("Initializing position model")
	return nil
}

// Update handles messages and updates the model
func (m *PositionModel) Update(msg tea.Msg) (*PositionModel, tea.Cmd) {
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

		m.logger.Debug().Int("width", m.width).Int("height", m.height).Msg("Position window resized")

	case tea.KeyMsg:
		if m.settingPosition {
			switch msg.String() {
			case "tab":
				m.nextInput()
			case "shift+tab":
				m.prevInput()
			case "enter":
				m.savePosition()
			case "esc":
				m.cancelPosition()
			default:
				m.updateCurrentInput(msg)
			}
		} else {
			switch msg.String() {
			case "s":
				m.startSettingPosition()
			case "r":
				m.requestPositions()
			case "c":
				m.clearPositions()
			case "m":
				m.toggleMyPosition()
			default:
				m.table, cmd = m.table.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case PositionMsg:
		m.addPosition(msg.NodeID, msg.NodeName, msg.Position)
		m.updateTable()
		m.logger.Debug().Uint32("node_id", msg.NodeID).Msg("Received position update")

	case MyPositionMsg:
		m.myPosition = msg.Position
		m.recalculateDistances()
		m.updateTable()
		m.logger.Debug().Msg("Updated my position")
	}

	return m, tea.Batch(cmds...)
}

// View renders the position view
func (m *PositionModel) View() string {
	if !m.ready {
		return m.styles.StatusText.Render("Loading positions...")
	}

	if m.settingPosition {
		return m.positionInputView()
	}

	return m.tableView()
}

// tableView renders the table view
func (m *PositionModel) tableView() string {
	header := m.styles.Header.Render("Node Positions")

	myPosStr := "My Position: Unknown"
	if m.myPosition != nil {
		myPosStr = fmt.Sprintf("My Position: %.6f, %.6f (%.0fm)",
			float64(m.myPosition.GetLatitudeI())/1e7,
			float64(m.myPosition.GetLongitudeI())/1e7,
			float64(m.myPosition.GetAltitude()))
	}

	myPosInfo := m.styles.StatusText.Render(myPosStr)

	help := m.styles.Help.Render("s: set position • r: request positions • c: clear • m: toggle my position")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		myPosInfo,
		m.table.View(),
		help,
	)

	return content
}

// positionInputView renders the position input view
func (m *PositionModel) positionInputView() string {
	header := m.styles.Header.Render("Set My Position")

	inputs := []string{
		m.latInput.View(),
		m.lonInput.View(),
		m.altInput.View(),
	}

	help := m.styles.Help.Render("tab: next field • shift+tab: prev field • enter: save • esc: cancel")

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		inputs[0],
		inputs[1],
		inputs[2],
		help,
	)

	return content
}

// addPosition adds a new position
func (m *PositionModel) addPosition(nodeID uint32, nodeName string, position *pb.Position) {
	// Remove existing position for this node
	for i, pos := range m.positions {
		if pos.NodeID == nodeID {
			m.positions = append(m.positions[:i], m.positions[i+1:]...)
			break
		}
	}

	// Calculate distance from our position
	distance := m.calculateDistance(position)

	// Add new position
	posData := PositionData{
		NodeID:    nodeID,
		NodeName:  nodeName,
		Timestamp: time.Now(),
		Position:  position,
		Distance:  distance,
	}

	m.positions = append(m.positions, posData)

	// Sort by distance
	sort.Slice(m.positions, func(i, j int) bool {
		return m.positions[i].Distance < m.positions[j].Distance
	})
}

// updateTable updates the table with current position data
func (m *PositionModel) updateTable() {
	rows := make([]table.Row, 0)

	for _, pos := range m.positions {
		nodeName := pos.NodeName
		if nodeName == "" {
			nodeName = fmt.Sprintf("Node %d", pos.NodeID)
		}

		lat := float64(pos.Position.GetLatitudeI()) / 1e7
		lon := float64(pos.Position.GetLongitudeI()) / 1e7
		alt := pos.Position.GetAltitude()

		distanceStr := "Unknown"
		if pos.Distance >= 0 {
			if pos.Distance < 1000 {
				distanceStr = fmt.Sprintf("%.0fm", pos.Distance)
			} else {
				distanceStr = fmt.Sprintf("%.1fkm", pos.Distance/1000)
			}
		}

		timeStr := pos.Timestamp.Format("15:04:05")

		rows = append(rows, table.Row{
			nodeName,
			fmt.Sprintf("%.6f", lat),
			fmt.Sprintf("%.6f", lon),
			fmt.Sprintf("%.0fm", float64(alt)),
			distanceStr,
			timeStr,
		})
	}

	m.table.SetRows(rows)
}

// calculateDistance calculates the distance between two positions using the Haversine formula
func (m *PositionModel) calculateDistance(position *pb.Position) float64 {
	if m.myPosition == nil || position == nil {
		return -1
	}

	lat1 := float64(m.myPosition.GetLatitudeI()) / 1e7
	lon1 := float64(m.myPosition.GetLongitudeI()) / 1e7
	lat2 := float64(position.GetLatitudeI()) / 1e7
	lon2 := float64(position.GetLongitudeI()) / 1e7

	return haversineDistance(lat1, lon1, lat2, lon2)
}

// haversineDistance calculates the distance between two points using the Haversine formula
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth's radius in meters

	dLat := toRadians(lat2 - lat1)
	dLon := toRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(toRadians(lat1))*math.Cos(toRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// toRadians converts degrees to radians
func toRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// recalculateDistances recalculates distances for all positions
func (m *PositionModel) recalculateDistances() {
	for i := range m.positions {
		m.positions[i].Distance = m.calculateDistance(m.positions[i].Position)
	}

	// Sort by distance
	sort.Slice(m.positions, func(i, j int) bool {
		return m.positions[i].Distance < m.positions[j].Distance
	})
}

// startSettingPosition starts the position setting process
func (m *PositionModel) startSettingPosition() {
	m.settingPosition = true
	m.currentInput = 0
	m.latInput.Focus()
	m.logger.Debug().Msg("Started setting position")
}

// cancelPosition cancels the position setting process
func (m *PositionModel) cancelPosition() {
	m.settingPosition = false
	m.currentInput = 0
	m.latInput.Blur()
	m.lonInput.Blur()
	m.altInput.Blur()
	m.logger.Debug().Msg("Cancelled setting position")
}

// nextInput moves to the next input field
func (m *PositionModel) nextInput() {
	switch m.currentInput {
	case 0:
		m.latInput.Blur()
		m.lonInput.Focus()
		m.currentInput = 1
	case 1:
		m.lonInput.Blur()
		m.altInput.Focus()
		m.currentInput = 2
	case 2:
		m.altInput.Blur()
		m.latInput.Focus()
		m.currentInput = 0
	}
}

// prevInput moves to the previous input field
func (m *PositionModel) prevInput() {
	switch m.currentInput {
	case 0:
		m.latInput.Blur()
		m.altInput.Focus()
		m.currentInput = 2
	case 1:
		m.lonInput.Blur()
		m.latInput.Focus()
		m.currentInput = 0
	case 2:
		m.altInput.Blur()
		m.lonInput.Focus()
		m.currentInput = 1
	}
}

// updateCurrentInput updates the current input field
func (m *PositionModel) updateCurrentInput(msg tea.KeyMsg) {
	var cmd tea.Cmd

	switch m.currentInput {
	case 0:
		m.latInput, cmd = m.latInput.Update(msg)
	case 1:
		m.lonInput, cmd = m.lonInput.Update(msg)
	case 2:
		m.altInput, cmd = m.altInput.Update(msg)
	}

	// Execute the command if needed
	if cmd != nil {
		// Handle the command
	}
}

// savePosition saves the entered position
func (m *PositionModel) savePosition() {
	// This would typically save the position through the client
	m.logger.Info().
		Str("latitude", m.latInput.Value()).
		Str("longitude", m.lonInput.Value()).
		Str("altitude", m.altInput.Value()).
		Msg("Saving position")

	m.cancelPosition()
}

// requestPositions requests position updates from all nodes
func (m *PositionModel) requestPositions() {
	m.logger.Info().Msg("Requesting position updates from all nodes")
	// This would typically send position requests through the client
}

// clearPositions clears all position data
func (m *PositionModel) clearPositions() {
	m.positions = make([]PositionData, 0)
	m.updateTable()
	m.logger.Info().Msg("Cleared all position data")
}

// toggleMyPosition toggles the display of my position
func (m *PositionModel) toggleMyPosition() {
	m.logger.Info().Msg("Toggling my position display")
	// This would typically toggle the display of our own position
}

// Messages for position updates
type PositionMsg struct {
	NodeID   uint32
	NodeName string
	Position *pb.Position
}

type MyPositionMsg struct {
	Position *pb.Position
}

// Commands for position
func UpdatePosition(nodeID uint32, nodeName string, position *pb.Position) tea.Cmd {
	return func() tea.Msg {
		return PositionMsg{
			NodeID:   nodeID,
			NodeName: nodeName,
			Position: position,
		}
	}
}

func UpdateMyPosition(position *pb.Position) tea.Cmd {
	return func() tea.Msg {
		return MyPositionMsg{
			Position: position,
		}
	}
}
