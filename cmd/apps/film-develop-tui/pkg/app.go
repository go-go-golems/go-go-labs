package pkg

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/internal/models"
)

// Model represents the main application model
type Model struct {
	stateMachine *models.StateMachine
	screen       Screen
	width        int
	height       int
	ready        bool
}

// NewModel creates a new application model
func NewModel() *Model {
	sm := models.NewStateMachine()
	return &Model{
		stateMachine: sm,
		screen:       GetScreenForState(sm.GetCurrentState()),
		width:        80,
		height:       24,
		ready:        false,
	}
}

// TimerTickMsg represents a timer tick message
type TimerTickMsg struct{}

// Init initializes the application
func (m *Model) Init() tea.Cmd {
	return m.tickCmd()
}

// tickCmd returns a command that sends timer ticks
func (m *Model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return TimerTickMsg{}
	})
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil

	case TimerTickMsg:
		// Continue ticking if we're on the timer screen or if any timer is running
		if m.stateMachine.GetCurrentState() == models.TimerScreenState ||
			(m.stateMachine.GetApplicationState().TimerState != nil &&
				m.stateMachine.GetApplicationState().TimerState.IsRunning) {
			return m, m.tickCmd()
		}
		return m, nil

	case tea.KeyMsg:
		// Handle quit globally
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		// Convert key to string for easier handling
		key := m.normalizeKey(msg)

		// Handle input through the current screen
		if !m.screen.HandleInput(key, m.stateMachine) {
			return m, tea.Quit
		}

		// Update screen if state changed
		newScreen := GetScreenForState(m.stateMachine.GetCurrentState())
		if fmt.Sprintf("%T", newScreen) != fmt.Sprintf("%T", m.screen) {
			m.screen = newScreen
		}

		// Start timer ticks if we enter timer screen
		var cmd tea.Cmd
		if m.stateMachine.GetCurrentState() == models.TimerScreenState {
			cmd = m.tickCmd()
		}

		return m, cmd
	}

	return m, nil
}

// View renders the application
func (m *Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	content := m.screen.Render(m.stateMachine.GetApplicationState())

	// Add some padding if needed
	if m.width > 80 {
		content = lipgloss.NewStyle().
			Width(m.width).
			Align(lipgloss.Center).
			Render(content)
	}

	return content
}

// normalizeKey normalizes tea.KeyMsg to string
func (m *Model) normalizeKey(msg tea.KeyMsg) string {
	switch msg.Type {
	case tea.KeyEnter:
		return "enter"
	case tea.KeyEsc:
		return "esc"
	case tea.KeyUp:
		return "up"
	case tea.KeyDown:
		return "down"
	case tea.KeyLeft:
		return "left"
	case tea.KeyRight:
		return "right"
	case tea.KeySpace:
		return "space"
	case tea.KeyTab:
		return "tab"
	case tea.KeyBackspace:
		return "backspace"
	case tea.KeyDelete:
		return "delete"
	case tea.KeyRunes:
		return msg.String()
	default:
		return msg.String()
	}
}

// GetStateMachine returns the state machine for testing
func (m *Model) GetStateMachine() *models.StateMachine {
	return m.stateMachine
}
