package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TUI styles
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1).
			Bold(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575"))

	progressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFB700"))

	resultStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7C3AED"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	timestampStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1)

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#625DA6")).
			Padding(1)
)

// TUIModel represents the state of our TUI
type TUIModel struct {
	agentID     string
	agentName   string
	agentRole   string
	messages    []string
	connected   bool
	socketPath  string
	conn        net.Conn
	width       int
	height      int
	maxMessages int
}

// socketMsg represents a message received from the socket
type socketMsg struct {
	message *SocketMessage
	err     error
}

// NewTUIModel creates a new TUI model
func NewTUIModel(socketPath string) *TUIModel {
	return &TUIModel{
		socketPath:  socketPath,
		messages:    make([]string, 0),
		maxMessages: 50, // Keep last 50 messages
		width:       80,
		height:      24,
	}
}

// Init initializes the TUI
func (m *TUIModel) Init() tea.Cmd {
	return tea.Batch(
		m.connectToSocket(),
		m.listenForMessages(),
	)
}

// connectToSocket establishes connection to Unix socket
func (m *TUIModel) connectToSocket() tea.Cmd {
	return func() tea.Msg {
		conn, err := net.Dial("unix", m.socketPath)
		if err != nil {
			return socketMsg{err: fmt.Errorf("failed to connect to socket: %w", err)}
		}
		m.conn = conn
		m.connected = true
		return socketMsg{message: &SocketMessage{Type: "connected"}}
	}
}

// listenForMessages listens for incoming socket messages
func (m *TUIModel) listenForMessages() tea.Cmd {
	return func() tea.Msg {
		if m.conn == nil {
			return nil
		}

		scanner := bufio.NewScanner(m.conn)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			msg, err := UnmarshalSocketMessage([]byte(line))
			if err != nil {
				return socketMsg{err: fmt.Errorf("failed to unmarshal message: %w", err)}
			}

			return socketMsg{message: msg}
		}

		if err := scanner.Err(); err != nil {
			return socketMsg{err: fmt.Errorf("socket read error: %w", err)}
		}

		return socketMsg{message: &SocketMessage{Type: "disconnected"}}
	}
}

// Update handles TUI updates
func (m *TUIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.conn != nil {
				m.conn.Close()
			}
			return m, tea.Quit
		}

	case socketMsg:
		if msg.err != nil {
			m.addMessage(fmt.Sprintf("❌ Error: %v", msg.err), "error")
			return m, tea.Quit
		}

		if msg.message != nil {
			return m.handleSocketMessage(msg.message)
		}
	}

	return m, nil
}

// handleSocketMessage processes incoming socket messages
func (m *TUIModel) handleSocketMessage(msg *SocketMessage) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case "connected":
		m.addMessage("🔗 Connected to orchestrator", "status")

	case "disconnected":
		m.addMessage("🔌 Disconnected from orchestrator", "status")
		return m, tea.Quit

	case "init":
		m.agentID = msg.AgentID
		m.agentName = msg.AgentName
		m.agentRole = msg.AgentRole
		m.addMessage("🚀 Agent initialized", "status")

	case "agent_update":
		// Format the message based on type
		var emoji string
		switch msg.MsgType {
		case "status":
			emoji = "📋"
		case "progress":
			emoji = "⚙️"
		case "result":
			emoji = "💡"
		case "error":
			emoji = "❌"
		default:
			emoji = "📝"
		}

		formattedMsg := fmt.Sprintf("%s %s", emoji, msg.Content)
		m.addMessage(formattedMsg, msg.MsgType)

	case "status_update":
		m.addMessage(fmt.Sprintf("🎯 %s", msg.Content), "status")

	case "shutdown":
		m.addMessage("👋 Shutting down", "status")
		return m, tea.Quit
	}

	// Continue listening for more messages
	return m, m.listenForMessages()
}

// addMessage adds a new message to the display
func (m *TUIModel) addMessage(content, msgType string) {
	timestamp := time.Now().Format("15:04:05")
	formattedMsg := fmt.Sprintf("[%s] %s", timestamp, content)

	// Apply styling based on message type
	switch msgType {
	case "status":
		formattedMsg = statusStyle.Render(formattedMsg)
	case "progress":
		formattedMsg = progressStyle.Render(formattedMsg)
	case "result":
		formattedMsg = resultStyle.Render(formattedMsg)
	case "error":
		formattedMsg = errorStyle.Render(formattedMsg)
	}

	m.messages = append(m.messages, formattedMsg)

	// Keep only the last N messages
	if len(m.messages) > m.maxMessages {
		m.messages = m.messages[len(m.messages)-m.maxMessages:]
	}
}

// View renders the TUI
func (m *TUIModel) View() string {
	var title string
	if m.agentName != "" {
		title = titleStyle.Render(fmt.Sprintf("🤖 %s", m.agentName))
	} else {
		title = titleStyle.Render("🤖 Agent Display")
	}

	var role string
	if m.agentRole != "" {
		role = headerStyle.Render(m.agentRole)
	}

	var status string
	if m.connected {
		status = statusStyle.Render("🟢 Connected")
	} else {
		status = errorStyle.Render("🔴 Disconnected")
	}

	header := lipgloss.JoinVertical(lipgloss.Left, title, role, status, "")

	// Calculate available height for messages
	headerHeight := lipgloss.Height(header)
	availableHeight := m.height - headerHeight - 4 // Account for borders and padding

	// Show only the messages that fit
	var visibleMessages []string
	if len(m.messages) > availableHeight {
		visibleMessages = m.messages[len(m.messages)-availableHeight:]
	} else {
		visibleMessages = m.messages
	}

	messageArea := strings.Join(visibleMessages, "\n")

	// Add padding for any remaining space
	messageLines := len(visibleMessages)
	if messageLines < availableHeight {
		padding := strings.Repeat("\n", availableHeight-messageLines)
		messageArea += padding
	}

	content := lipgloss.JoinVertical(lipgloss.Left, header, messageArea)

	footer := timestampStyle.Render("Press 'q' to quit")

	fullContent := lipgloss.JoinVertical(lipgloss.Left, content, "", footer)

	return borderStyle.Width(m.width - 2).Height(m.height - 2).Render(fullContent)
}

// RunTUI starts the TUI for the given socket path
func RunTUI(socketPath string) error {
	model := NewTUIModel(socketPath)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err := p.Run()
	return err
}
