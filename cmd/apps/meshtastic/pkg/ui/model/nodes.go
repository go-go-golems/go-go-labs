package model

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
)

// NodeStatus represents the status of a node
type NodeStatus int

const (
	NodeStatusOnline NodeStatus = iota
	NodeStatusOffline
	NodeStatusUnknown
)

// String returns the string representation of the node status
func (s NodeStatus) String() string {
	switch s {
	case NodeStatusOnline:
		return "Online"
	case NodeStatusOffline:
		return "Offline"
	case NodeStatusUnknown:
		return "Unknown"
	default:
		return "Unknown"
	}
}

// Node represents a mesh node
type Node struct {
	ID              string
	Name            string
	LongName        string
	Status          NodeStatus
	LastSeen        time.Time
	Distance        float64
	Battery         int
	SNR             float64
	RSSI            int
	HardwareModel   string
	FirmwareVersion string
	Position        struct {
		Latitude  float64
		Longitude float64
		Altitude  int
	}
}

// NodesModel handles the nodes view
type NodesModel struct {
	styles   *view.Styles
	viewport viewport.Model
	nodes    []Node
	width    int
	height   int
	ready    bool

	// Selection
	selected int
	focused  bool

	// Sorting
	sortBy   string
	sortDesc bool
}

// NewNodesModel creates a new nodes model
func NewNodesModel(styles *view.Styles) *NodesModel {
	vp := viewport.New(0, 0)

	return &NodesModel{
		styles:   styles,
		viewport: vp,
		nodes:    make([]Node, 0),
		focused:  true,
		sortBy:   "name",
	}
}

// Init initializes the nodes model
func (m *NodesModel) Init() tea.Cmd {
	// Add some sample nodes
	m.nodes = []Node{
		{
			ID:              "!1234567890",
			Name:            "BaseStation",
			LongName:        "Base Station Node",
			Status:          NodeStatusOnline,
			LastSeen:        time.Now(),
			Distance:        0.0,
			Battery:         100,
			SNR:             10.5,
			RSSI:            -30,
			HardwareModel:   "TTGO LoRa32",
			FirmwareVersion: "2.3.2",
		},
		{
			ID:              "!2345678901",
			Name:            "Mobile01",
			LongName:        "Mobile Unit 01",
			Status:          NodeStatusOnline,
			LastSeen:        time.Now().Add(-2 * time.Minute),
			Distance:        1.2,
			Battery:         85,
			SNR:             8.2,
			RSSI:            -45,
			HardwareModel:   "Heltec WiFi LoRa 32",
			FirmwareVersion: "2.3.1",
		},
		{
			ID:              "!3456789012",
			Name:            "Remote01",
			LongName:        "Remote Station 01",
			Status:          NodeStatusOffline,
			LastSeen:        time.Now().Add(-15 * time.Minute),
			Distance:        5.8,
			Battery:         45,
			SNR:             3.1,
			RSSI:            -78,
			HardwareModel:   "T-Beam",
			FirmwareVersion: "2.2.21",
		},
		{
			ID:              "!4567890123",
			Name:            "Repeater",
			LongName:        "Mesh Repeater",
			Status:          NodeStatusOnline,
			LastSeen:        time.Now().Add(-30 * time.Second),
			Distance:        2.5,
			Battery:         0, // Solar powered
			SNR:             12.3,
			RSSI:            -25,
			HardwareModel:   "RAK WisBlock",
			FirmwareVersion: "2.3.2",
		},
	}

	m.sortNodes()
	return nil
}

// Update handles messages and updates the model
func (m *NodesModel) Update(msg tea.Msg) (*NodesModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 2 // Account for title
		m.ready = true
		m.updateContent()

	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("j", "down"))):
			if m.selected < len(m.nodes)-1 {
				m.selected++
				m.updateContent()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.selected > 0 {
				m.selected--
				m.updateContent()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("s"))):
			// Cycle through sort options
			switch m.sortBy {
			case "name":
				m.sortBy = "status"
			case "status":
				m.sortBy = "distance"
			case "distance":
				m.sortBy = "lastseen"
			case "lastseen":
				m.sortBy = "battery"
			case "battery":
				m.sortBy = "name"
			}
			m.sortNodes()
			m.updateContent()
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Reverse sort order
			m.sortDesc = !m.sortDesc
			m.sortNodes()
			m.updateContent()
		}

	case NodeUpdateMsg:
		// Update node information
		node := Node(msg)
		found := false
		for i, n := range m.nodes {
			if n.ID == node.ID {
				m.nodes[i] = node
				found = true
				break
			}
		}
		if !found {
			m.nodes = append(m.nodes, node)
		}
		m.sortNodes()
		m.updateContent()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the nodes view
func (m *NodesModel) View() string {
	if !m.ready {
		return "Loading nodes..."
	}

	title := m.styles.Title.Render(
		fmt.Sprintf("Nodes (%d) - Sorted by %s", len(m.nodes), m.sortBy),
	)

	if len(m.nodes) == 0 {
		empty := m.styles.Muted.Render("No nodes found")
		return lipgloss.JoinVertical(lipgloss.Left, title, empty)
	}

	content := m.viewport.View()

	help := m.styles.Muted.Render("s: sort • r: reverse • j/k: navigate")

	return lipgloss.JoinVertical(lipgloss.Left, title, content, help)
}

// updateContent updates the viewport content
func (m *NodesModel) updateContent() {
	if !m.ready {
		return
	}

	var content []string

	for i, node := range m.nodes {
		var nodeStyle lipgloss.Style
		switch node.Status {
		case NodeStatusOnline:
			nodeStyle = m.styles.NodeOnline
		case NodeStatusOffline:
			nodeStyle = m.styles.NodeOffline
		default:
			nodeStyle = m.styles.Node
		}

		if i == m.selected {
			nodeStyle = nodeStyle.Copy().Border(lipgloss.RoundedBorder(), true).
				BorderForeground(view.Colors.Primary)
		}

		// Node header
		header := lipgloss.JoinHorizontal(
			lipgloss.Left,
			m.styles.NodeName.Render(node.Name),
			" ",
			m.styles.NodeId.Render(fmt.Sprintf("(%s)", node.ID)),
		)

		// Status and basic info
		status := m.styles.NodeStatus.Render(
			fmt.Sprintf("Status: %s", node.Status.String()),
		)

		lastSeen := m.styles.Muted.Render(
			fmt.Sprintf("Last seen: %s", formatDuration(time.Since(node.LastSeen))),
		)

		// Technical info
		distance := ""
		if node.Distance > 0 {
			distance = fmt.Sprintf("Distance: %.1f km", node.Distance)
		}

		battery := ""
		if node.Battery > 0 {
			battery = fmt.Sprintf("Battery: %d%%", node.Battery)
		} else {
			battery = "Battery: Solar/External"
		}

		signal := fmt.Sprintf("SNR: %.1f dB, RSSI: %d dBm", node.SNR, node.RSSI)

		tech := lipgloss.JoinVertical(
			lipgloss.Left,
			distance,
			battery,
			signal,
		)

		// Hardware info
		hardware := fmt.Sprintf("Hardware: %s", node.HardwareModel)
		firmware := fmt.Sprintf("Firmware: %s", node.FirmwareVersion)

		hw := lipgloss.JoinVertical(
			lipgloss.Left,
			hardware,
			firmware,
		)

		nodeContent := lipgloss.JoinVertical(
			lipgloss.Left,
			header,
			status,
			lastSeen,
			"",
			tech,
			"",
			hw,
		)

		content = append(content, nodeStyle.Render(nodeContent))
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, content...))
}

// sortNodes sorts the nodes based on the current sort criteria
func (m *NodesModel) sortNodes() {
	// Simple bubble sort for demonstration
	for i := 0; i < len(m.nodes)-1; i++ {
		for j := 0; j < len(m.nodes)-1-i; j++ {
			shouldSwap := false

			switch m.sortBy {
			case "name":
				if m.sortDesc {
					shouldSwap = m.nodes[j].Name < m.nodes[j+1].Name
				} else {
					shouldSwap = m.nodes[j].Name > m.nodes[j+1].Name
				}
			case "status":
				if m.sortDesc {
					shouldSwap = m.nodes[j].Status < m.nodes[j+1].Status
				} else {
					shouldSwap = m.nodes[j].Status > m.nodes[j+1].Status
				}
			case "distance":
				if m.sortDesc {
					shouldSwap = m.nodes[j].Distance < m.nodes[j+1].Distance
				} else {
					shouldSwap = m.nodes[j].Distance > m.nodes[j+1].Distance
				}
			case "lastseen":
				if m.sortDesc {
					shouldSwap = m.nodes[j].LastSeen.Before(m.nodes[j+1].LastSeen)
				} else {
					shouldSwap = m.nodes[j].LastSeen.After(m.nodes[j+1].LastSeen)
				}
			case "battery":
				if m.sortDesc {
					shouldSwap = m.nodes[j].Battery < m.nodes[j+1].Battery
				} else {
					shouldSwap = m.nodes[j].Battery > m.nodes[j+1].Battery
				}
			}

			if shouldSwap {
				m.nodes[j], m.nodes[j+1] = m.nodes[j+1], m.nodes[j]
			}
		}
	}
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	return fmt.Sprintf("%dh ago", int(d.Hours()))
}

// SetFocused sets the focused state
func (m *NodesModel) SetFocused(focused bool) {
	m.focused = focused
}

// GetNodes returns all nodes
func (m *NodesModel) GetNodes() []Node {
	return m.nodes
}

// GetSelected returns the selected node index
func (m *NodesModel) GetSelected() int {
	return m.selected
}

// GetSelectedNode returns the selected node
func (m *NodesModel) GetSelectedNode() *Node {
	if m.selected >= 0 && m.selected < len(m.nodes) {
		return &m.nodes[m.selected]
	}
	return nil
}

// AddNode adds a new node
func (m *NodesModel) AddNode(node Node) {
	m.nodes = append(m.nodes, node)
	m.sortNodes()
	m.updateContent()
}

// UpdateNode updates an existing node
func (m *NodesModel) UpdateNode(node Node) {
	for i, n := range m.nodes {
		if n.ID == node.ID {
			m.nodes[i] = node
			m.sortNodes()
			m.updateContent()
			return
		}
	}
	// Node not found, add it
	m.AddNode(node)
}

// NodeUpdateMsg is sent when a node is updated
type NodeUpdateMsg Node

// NodeUpdate creates a node update message
func NodeUpdate(node Node) tea.Cmd {
	return func() tea.Msg {
		return NodeUpdateMsg(node)
	}
}
