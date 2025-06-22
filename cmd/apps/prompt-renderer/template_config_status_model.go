package main

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/prompt-renderer/shared"
)

// Ensure StatusBarModel implements the Resizable interface
var _ shared.Resizable = (*StatusBarModel)(nil)

// StatusBarModel handles the bottom status line and toast messages.
type StatusBarModel struct {
	width   int
	height  int
	message string
	expiry  time.Time
}

// NewStatusBarModel creates a new status bar model.
func NewStatusBarModel() *StatusBarModel {
	return &StatusBarModel{}
}

// Init implements tea.Model.Init
func (m *StatusBarModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update
func (m *StatusBarModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case CopyDoneMsg:
		m.message = "✓ Copied to clipboard!"
		m.expiry = time.Now().Add(750 * time.Millisecond)
		return m, tea.Tick(750*time.Millisecond, func(time.Time) tea.Msg {
			return ClearToastMsg{}
		})
	case ClearToastMsg:
		m.message = ""
		return m, nil
	case tea.WindowSizeMsg:
		return m, m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

// View implements tea.Model.View by rendering the status line.
func (m *StatusBarModel) View() string {
	status := "↑↓/j/k: Navigate  Tab: Next  Space/Enter: Toggle  c: Copy  s: Save  ←/Esc: Back"
	if m.message != "" && time.Now().Before(m.expiry) {
		status = m.message
	}
	// pad or trim to width
	if len(status) < m.width {
		status = status + strings.Repeat(" ", m.width-len(status))
	} else if len(status) > m.width {
		status = status[:m.width]
	}
	return status
}

// SetSize implements shared.Resizable
func (m *StatusBarModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
}
