package model

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/keys"
	"github.com/go-go-golems/go-go-labs/cmd/apps/meshtastic/pkg/ui/view"
	"github.com/pkg/errors"
)

// Tab represents a tab in the TUI
type Tab int

const (
	TabMessages Tab = iota
	TabNodes
	TabStatus
)

// String returns the string representation of the tab
func (t Tab) String() string {
	switch t {
	case TabMessages:
		return "Messages"
	case TabNodes:
		return "Nodes"
	case TabStatus:
		return "Status"
	default:
		return "Unknown"
	}
}

// Mode represents the current mode of the TUI
type Mode int

const (
	ModeView Mode = iota
	ModeCompose
	ModeHelp
)

// RootModel is the main model for the TUI
type RootModel struct {
	keys   *keys.KeyMap
	styles *view.Styles
	width  int
	height int
	ready  bool

	// Current state
	currentTab Tab
	mode       Mode
	showHelp   bool

	// Sub-models
	messages *MessagesModel
	nodes    *NodesModel
	status   *StatusModel
	compose  *ComposeModel

	// Error state
	err error

	// Client for cleanup
	client interface {
		Close() error
	}
}

// NewRootModel creates a new root model
func NewRootModel() *RootModel {
	keyMap := keys.DefaultKeyMap()
	styles := view.DefaultStyles()

	return &RootModel{
		keys:     keyMap,
		styles:   styles,
		messages: NewMessagesModel(styles),
		nodes:    NewNodesModel(styles),
		status:   NewStatusModel(styles),
		compose:  NewComposeModel(styles),
		mode:     ModeView,
	}
}

// NewRootModelWithClient creates a new root model with a client
func NewRootModelWithClient(client interface{ Close() error }) *RootModel {
	m := NewRootModel()
	m.client = client
	return m
}

// Init initializes the model
func (m *RootModel) Init() tea.Cmd {
	return tea.Batch(
		m.messages.Init(),
		m.nodes.Init(),
		m.status.Init(),
		m.compose.Init(),
	)
}

// Update handles messages and updates the model
func (m *RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true

		// Update sub-models with new dimensions
		contentWidth := m.width - 4   // Account for padding
		contentHeight := m.height - 6 // Account for header, tabs, footer

		m.messages, cmd = m.messages.Update(tea.WindowSizeMsg{
			Width:  contentWidth,
			Height: contentHeight,
		})
		cmds = append(cmds, cmd)

		m.nodes, cmd = m.nodes.Update(tea.WindowSizeMsg{
			Width:  contentWidth,
			Height: contentHeight,
		})
		cmds = append(cmds, cmd)

		m.status, cmd = m.status.Update(tea.WindowSizeMsg{
			Width:  contentWidth,
			Height: contentHeight,
		})
		cmds = append(cmds, cmd)

		m.compose, cmd = m.compose.Update(tea.WindowSizeMsg{
			Width:  contentWidth,
			Height: contentHeight,
		})
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		if m.mode == ModeCompose {
			return m.handleComposeMode(msg)
		}

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp
		case key.Matches(msg, m.keys.TabMessages):
			m.currentTab = TabMessages
		case key.Matches(msg, m.keys.TabNodes):
			m.currentTab = TabNodes
		case key.Matches(msg, m.keys.TabStatus):
			m.currentTab = TabStatus
		case key.Matches(msg, m.keys.Tab):
			m.currentTab = (m.currentTab + 1) % 3
		case key.Matches(msg, m.keys.Compose):
			m.mode = ModeCompose
		default:
			// Forward to current tab
			switch m.currentTab {
			case TabMessages:
				m.messages, cmd = m.messages.Update(msg)
				cmds = append(cmds, cmd)
			case TabNodes:
				m.nodes, cmd = m.nodes.Update(msg)
				cmds = append(cmds, cmd)
			case TabStatus:
				m.status, cmd = m.status.Update(msg)
				cmds = append(cmds, cmd)
			}
		}

	case ComposeCompleteMsg:
		m.mode = ModeView
		// Forward to messages model
		m.messages, cmd = m.messages.Update(msg)
		cmds = append(cmds, cmd)

	case ComposeCancelMsg:
		m.mode = ModeView

	case error:
		m.err = msg

	default:
		// Forward to all sub-models
		m.messages, cmd = m.messages.Update(msg)
		cmds = append(cmds, cmd)

		m.nodes, cmd = m.nodes.Update(msg)
		cmds = append(cmds, cmd)

		m.status, cmd = m.status.Update(msg)
		cmds = append(cmds, cmd)

		m.compose, cmd = m.compose.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// handleComposeMode handles keyboard input in compose mode
func (m *RootModel) handleComposeMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case key.Matches(msg, m.keys.Escape):
		m.mode = ModeView
		return m, ComposeCancel()
	case key.Matches(msg, m.keys.Send):
		// Send message
		m.mode = ModeView
		return m, ComposeComplete(m.compose.Value())
	default:
		// Forward to compose model
		m.compose, cmd = m.compose.Update(msg)
		return m, cmd
	}
}

// View renders the model
func (m *RootModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	if m.mode == ModeCompose {
		return m.composeView()
	}

	var content string
	switch m.currentTab {
	case TabMessages:
		content = m.messages.View()
	case TabNodes:
		content = m.nodes.View()
	case TabStatus:
		content = m.status.View()
	}

	header := m.headerView()
	tabs := m.tabsView()
	footer := m.footerView()

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabs,
		content,
		footer,
	)

	if m.showHelp {
		help := m.helpView()
		view = lipgloss.JoinVertical(lipgloss.Left, view, help)
	}

	return m.styles.App.Render(view)
}

// headerView renders the header
func (m *RootModel) headerView() string {
	title := "Meshtastic TUI"
	if m.err != nil {
		title = fmt.Sprintf("%s - Error: %v", title, m.err)
	}
	return m.styles.Header.Render(title)
}

// tabsView renders the tabs
func (m *RootModel) tabsView() string {
	tabs := make([]string, 3)

	for i := Tab(0); i < 3; i++ {
		style := m.styles.TabInactive
		if i == m.currentTab {
			style = m.styles.TabActive
		}
		tabs[i] = style.Render(fmt.Sprintf("%d. %s", i+1, i.String()))
	}

	return m.styles.TabBar.Render(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
}

// footerView renders the footer
func (m *RootModel) footerView() string {
	help := m.keys.HelpModel()
	return m.styles.Footer.Render(help.View(m.keys))
}

// composeView renders the compose view
func (m *RootModel) composeView() string {
	content := m.compose.View()

	header := m.headerView()
	footer := m.footerView()

	view := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		footer,
	)

	return m.styles.App.Render(view)
}

// helpView renders the help view
func (m *RootModel) helpView() string {
	help := m.keys.HelpModel()
	help.ShowAll = true
	return m.styles.Help.Render(help.View(m.keys))
}

// SetError sets an error on the model
func (m *RootModel) SetError(err error) {
	m.err = errors.Wrap(err, "root model error")
}

// ClearError clears the error on the model
func (m *RootModel) ClearError() {
	m.err = nil
}

// GetCurrentTab returns the current tab
func (m *RootModel) GetCurrentTab() Tab {
	return m.currentTab
}

// SetCurrentTab sets the current tab
func (m *RootModel) SetCurrentTab(tab Tab) {
	m.currentTab = tab
}

// GetMode returns the current mode
func (m *RootModel) GetMode() Mode {
	return m.mode
}

// SetMode sets the current mode
func (m *RootModel) SetMode(mode Mode) {
	m.mode = mode
}

// Cleanup performs cleanup operations
func (m *RootModel) Cleanup() error {
	if m.client != nil {
		if err := m.client.Close(); err != nil {
			return errors.Wrap(err, "failed to close client")
		}
	}
	return nil
}
