package model

import (
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
)

// ComposeModel handles message composition
type ComposeModel struct {
	styles *view.Styles
	width  int
	height int
	ready  bool

	// Input fields
	recipient textinput.Model
	message   textarea.Model

	// State
	focused   int // 0 = recipient, 1 = message
	submitted bool
}

// NewComposeModel creates a new compose model
func NewComposeModel(styles *view.Styles) *ComposeModel {
	recipient := textinput.New()
	recipient.Placeholder = "Recipient (or 'broadcast')"
	recipient.Focus()
	recipient.CharLimit = 50
	recipient.Width = 30

	message := textarea.New()
	message.Placeholder = "Type your message here..."
	message.SetWidth(50)
	message.SetHeight(10)

	return &ComposeModel{
		styles:    styles,
		recipient: recipient,
		message:   message,
		focused:   0,
	}
}

// Init initializes the compose model
func (m *ComposeModel) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		textarea.Blink,
	)
}

// Update handles messages and updates the model
func (m *ComposeModel) Update(msg tea.Msg) (*ComposeModel, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update input dimensions
		m.recipient.Width = msg.Width - 20
		m.message.SetWidth(msg.Width - 4)
		m.message.SetHeight(msg.Height - 10)

	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if m.focused == 0 {
				m.focused = 1
				m.recipient.Blur()
				m.message.Focus()
			} else {
				m.focused = 0
				m.message.Blur()
				m.recipient.Focus()
			}
		case "shift+tab":
			if m.focused == 1 {
				m.focused = 0
				m.message.Blur()
				m.recipient.Focus()
			} else {
				m.focused = 1
				m.recipient.Blur()
				m.message.Focus()
			}
		}
	}

	// Update inputs
	if m.focused == 0 {
		m.recipient, cmd = m.recipient.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		m.message, cmd = m.message.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the compose view
func (m *ComposeModel) View() string {
	if !m.ready {
		return "Loading compose..."
	}

	title := m.styles.ComposeTitle.Render("Compose Message")

	// Recipient field
	recipientLabel := m.styles.Subtitle.Render("To:")
	recipientInput := m.recipient.View()
	if m.focused == 0 {
		recipientInput = m.styles.Focused.Render(recipientInput)
	} else {
		recipientInput = m.styles.Unfocused.Render(recipientInput)
	}

	recipientSection := lipgloss.JoinVertical(
		lipgloss.Left,
		recipientLabel,
		recipientInput,
	)

	// Message field
	messageLabel := m.styles.Subtitle.Render("Message:")
	messageInput := m.message.View()
	if m.focused == 1 {
		messageInput = m.styles.Focused.Render(messageInput)
	} else {
		messageInput = m.styles.Unfocused.Render(messageInput)
	}

	messageSection := lipgloss.JoinVertical(
		lipgloss.Left,
		messageLabel,
		messageInput,
	)

	// Instructions
	instructions := m.styles.Muted.Render(
		"Tab: Switch fields • Ctrl+S: Send • Esc: Cancel",
	)

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		recipientSection,
		"",
		messageSection,
		"",
		instructions,
	)

	return m.styles.ComposeBox.Render(content)
}

// Value returns the composed message
func (m *ComposeModel) Value() string {
	return m.message.Value()
}

// Recipient returns the recipient
func (m *ComposeModel) Recipient() string {
	return m.recipient.Value()
}

// Reset resets the compose model
func (m *ComposeModel) Reset() {
	m.recipient.SetValue("")
	m.message.SetValue("")
	m.focused = 0
	m.recipient.Focus()
	m.message.Blur()
	m.submitted = false
}

// SetRecipient sets the recipient
func (m *ComposeModel) SetRecipient(recipient string) {
	m.recipient.SetValue(recipient)
}

// SetMessage sets the message
func (m *ComposeModel) SetMessage(message string) {
	m.message.SetValue(message)
}

// IsEmpty returns true if both fields are empty
func (m *ComposeModel) IsEmpty() bool {
	return m.recipient.Value() == "" && m.message.Value() == ""
}

// IsValid returns true if the compose is valid
func (m *ComposeModel) IsValid() bool {
	return m.message.Value() != ""
}

// SetFocused sets the focused field
func (m *ComposeModel) SetFocused(field int) {
	if field == 0 {
		m.focused = 0
		m.recipient.Focus()
		m.message.Blur()
	} else {
		m.focused = 1
		m.recipient.Blur()
		m.message.Focus()
	}
}

// GetFocused returns the focused field
func (m *ComposeModel) GetFocused() int {
	return m.focused
}
