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

// Message represents a message in the TUI
type Message struct {
	ID      string
	From    string
	To      string
	Content string
	Time    time.Time
	IsSent  bool
	IsRead  bool
}

// MessagesModel handles the messages view
type MessagesModel struct {
	styles   *view.Styles
	viewport viewport.Model
	messages []Message
	width    int
	height   int
	ready    bool

	// Selection
	selected int
	focused  bool
}

// NewMessagesModel creates a new messages model
func NewMessagesModel(styles *view.Styles) *MessagesModel {
	vp := viewport.New(0, 0)

	return &MessagesModel{
		styles:   styles,
		viewport: vp,
		messages: make([]Message, 0),
		focused:  true,
	}
}

// Init initializes the messages model
func (m *MessagesModel) Init() tea.Cmd {
	// Add some sample messages
	m.messages = []Message{
		{
			ID:      "1",
			From:    "Node-1234",
			To:      "Broadcast",
			Content: "Hello everyone!",
			Time:    time.Now().Add(-5 * time.Minute),
			IsSent:  false,
			IsRead:  true,
		},
		{
			ID:      "2",
			From:    "Me",
			To:      "Node-1234",
			Content: "Hi there!",
			Time:    time.Now().Add(-3 * time.Minute),
			IsSent:  true,
			IsRead:  true,
		},
		{
			ID:      "3",
			From:    "Node-5678",
			To:      "Broadcast",
			Content: "Weather looks good today",
			Time:    time.Now().Add(-1 * time.Minute),
			IsSent:  false,
			IsRead:  false,
		},
	}

	return nil
}

// Update handles messages and updates the model
func (m *MessagesModel) Update(msg tea.Msg) (*MessagesModel, tea.Cmd) {
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
			if m.selected < len(m.messages)-1 {
				m.selected++
				m.updateContent()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("k", "up"))):
			if m.selected > 0 {
				m.selected--
				m.updateContent()
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.selected < len(m.messages) {
				m.messages[m.selected].IsRead = true
				m.updateContent()
			}
		}

	case ComposeCompleteMsg:
		// Add new message
		newMsg := Message{
			ID:      fmt.Sprintf("msg-%d", len(m.messages)+1),
			From:    "Me",
			To:      "Broadcast",
			Content: string(msg),
			Time:    time.Now(),
			IsSent:  true,
			IsRead:  true,
		}
		m.messages = append(m.messages, newMsg)
		m.selected = len(m.messages) - 1
		m.updateContent()

	case NewMessageMsg:
		// Add incoming message
		m.messages = append(m.messages, Message(msg))
		m.updateContent()
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

// View renders the messages view
func (m *MessagesModel) View() string {
	if !m.ready {
		return "Loading messages..."
	}

	title := m.styles.Title.Render("Messages")

	if len(m.messages) == 0 {
		empty := m.styles.Muted.Render("No messages yet")
		return lipgloss.JoinVertical(lipgloss.Left, title, empty)
	}

	content := m.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}

// updateContent updates the viewport content
func (m *MessagesModel) updateContent() {
	if !m.ready {
		return
	}

	var content []string

	for i, msg := range m.messages {
		var msgStyle lipgloss.Style
		if msg.IsSent {
			msgStyle = m.styles.MessageSent
		} else {
			msgStyle = m.styles.MessageReceived
		}

		if i == m.selected {
			msgStyle = msgStyle.Copy().Border(lipgloss.RoundedBorder(), true).
				BorderForeground(view.Colors.Primary)
		}

		header := m.styles.MessageHeader.Render(
			fmt.Sprintf("%s → %s", msg.From, msg.To),
		)

		body := m.styles.MessageBody.Render(msg.Content)

		timestamp := m.styles.MessageTime.Render(
			msg.Time.Format("15:04:05"),
		)

		readStatus := ""
		if !msg.IsRead && !msg.IsSent {
			readStatus = m.styles.Warning.Render(" [NEW]")
		}

		msgContent := lipgloss.JoinVertical(
			lipgloss.Left,
			header+readStatus,
			body,
			timestamp,
		)

		content = append(content, msgStyle.Render(msgContent))
	}

	m.viewport.SetContent(lipgloss.JoinVertical(lipgloss.Left, content...))
}

// AddMessage adds a new message to the list
func (m *MessagesModel) AddMessage(msg Message) {
	m.messages = append(m.messages, msg)
	m.updateContent()
}

// SetFocused sets the focused state
func (m *MessagesModel) SetFocused(focused bool) {
	m.focused = focused
}

// GetMessages returns all messages
func (m *MessagesModel) GetMessages() []Message {
	return m.messages
}

// GetSelected returns the selected message index
func (m *MessagesModel) GetSelected() int {
	return m.selected
}

// ComposeCompleteMsg is sent when a message is composed
type ComposeCompleteMsg string

// ComposeCancelMsg is sent when compose is cancelled
type ComposeCancelMsg struct{}

// NewMessageMsg is sent when a new message is received
type NewMessageMsg Message

// ComposeComplete creates a compose complete message
func ComposeComplete(content string) tea.Cmd {
	return func() tea.Msg {
		return ComposeCompleteMsg(content)
	}
}

// ComposeCancel creates a compose cancel message
func ComposeCancel() tea.Cmd {
	return func() tea.Msg {
		return ComposeCancelMsg{}
	}
}

// NewMessage creates a new message command
func NewMessage(msg Message) tea.Cmd {
	return func() tea.Msg {
		return NewMessageMsg(msg)
	}
}
