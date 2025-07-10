package model

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
	"github.com/rs/zerolog/log"
)

// FilterMode represents different filtering modes
type FilterMode int

const (
	FilterAll FilterMode = iota
	FilterSent
	FilterReceived
	FilterPrivate
	FilterPublic
	FilterChannel
)

// Message represents a message in the TUI
type Message struct {
	ID         string
	From       string
	FromNodeID uint32
	To         string
	ToNodeID   uint32
	Content    string
	Time       time.Time
	IsSent     bool
	IsRead     bool
	IsPrivate  bool
	Channel    uint32
	ReplyTo    string // ID of message this is replying to
	ThreadID   string // ID of thread this message belongs to
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

	// Filtering
	filterMode   FilterMode
	filterText   string
	filterActive bool

	// Threading
	showThreads   bool
	currentThread string

	// Channels
	currentChannel uint32
	channels       map[uint32]string
}

// NewMessagesModel creates a new messages model
func NewMessagesModel(styles *view.Styles) *MessagesModel {
	vp := viewport.New(0, 0)

	return &MessagesModel{
		styles:         styles,
		viewport:       vp,
		messages:       make([]Message, 0),
		focused:        true,
		filterMode:     FilterAll,
		channels:       make(map[uint32]string),
		currentChannel: 0,
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
		case key.Matches(msg, key.NewBinding(key.WithKeys("f"))):
			// Cycle through filter modes
			m.filterMode = (m.filterMode + 1) % 6
			m.updateContent()
			log.Debug().Int("filter_mode", int(m.filterMode)).Msg("Changed filter mode")
		case key.Matches(msg, key.NewBinding(key.WithKeys("t"))):
			// Toggle thread view
			m.showThreads = !m.showThreads
			m.updateContent()
			log.Debug().Bool("show_threads", m.showThreads).Msg("Toggled thread view")
		case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
			// Reply to selected message
			if m.selected < len(m.messages) {
				selectedMsg := m.messages[m.selected]
				// This would typically trigger a compose with reply context
				log.Debug().Str("reply_to", selectedMsg.ID).Msg("Replying to message")
			}
		case key.Matches(msg, key.NewBinding(key.WithKeys("p"))):
			// Toggle private message filter
			if m.filterMode == FilterPrivate {
				m.filterMode = FilterAll
			} else {
				m.filterMode = FilterPrivate
			}
			m.updateContent()
			log.Debug().Int("filter_mode", int(m.filterMode)).Msg("Toggled private filter")
		case key.Matches(msg, key.NewBinding(key.WithKeys("c"))):
			// Toggle channel message filter
			if m.filterMode == FilterChannel {
				m.filterMode = FilterAll
			} else {
				m.filterMode = FilterChannel
			}
			m.updateContent()
			log.Debug().Int("filter_mode", int(m.filterMode)).Msg("Toggled channel filter")
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

	// Create title with filter info
	filterInfo := m.getFilterInfo()
	title := m.styles.Title.Render(fmt.Sprintf("Messages %s", filterInfo))

	// Create help text
	help := m.styles.Help.Render("↑/↓: navigate • f: filter • t: threads • r: reply • p: private • c: channel")

	filteredMessages := m.getFilteredMessages()
	if len(filteredMessages) == 0 {
		empty := m.styles.Muted.Render("No messages match current filter")
		return lipgloss.JoinVertical(lipgloss.Left, title, empty, help)
	}

	content := m.viewport.View()

	return lipgloss.JoinVertical(lipgloss.Left, title, content, help)
}

// updateContent updates the viewport content
func (m *MessagesModel) updateContent() {
	if !m.ready {
		return
	}

	var content []string
	filteredMessages := m.getFilteredMessages()

	for i, msg := range filteredMessages {
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

// getFilteredMessages returns messages filtered by current filter mode
func (m *MessagesModel) getFilteredMessages() []Message {
	var filtered []Message

	for _, msg := range m.messages {
		if m.shouldIncludeMessage(msg) {
			filtered = append(filtered, msg)
		}
	}

	return filtered
}

// shouldIncludeMessage checks if a message should be included based on current filter
func (m *MessagesModel) shouldIncludeMessage(msg Message) bool {
	switch m.filterMode {
	case FilterAll:
		return true
	case FilterSent:
		return msg.IsSent
	case FilterReceived:
		return !msg.IsSent
	case FilterPrivate:
		return msg.IsPrivate
	case FilterPublic:
		return !msg.IsPrivate
	case FilterChannel:
		return msg.Channel == m.currentChannel
	default:
		return true
	}
}

// getFilterInfo returns a string describing the current filter
func (m *MessagesModel) getFilterInfo() string {
	switch m.filterMode {
	case FilterAll:
		return "(All)"
	case FilterSent:
		return "(Sent)"
	case FilterReceived:
		return "(Received)"
	case FilterPrivate:
		return "(Private)"
	case FilterPublic:
		return "(Public)"
	case FilterChannel:
		channelName := m.channels[m.currentChannel]
		if channelName == "" {
			channelName = fmt.Sprintf("Channel %d", m.currentChannel)
		}
		return fmt.Sprintf("(%s)", channelName)
	default:
		return ""
	}
}
