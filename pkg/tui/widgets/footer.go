package widgets

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/tui/keys"
)

// FooterWidget displays keyboard commands help
type FooterWidget struct {
	width  int
	keys   keys.KeyMap
	styles FooterStyles
}

type FooterStyles struct {
	Container lipgloss.Style
	Commands  lipgloss.Style
}

// NewFooterWidget creates a new footer widget
func NewFooterWidget(keyMap keys.KeyMap, styles FooterStyles) FooterWidget {
	return FooterWidget{
		keys:   keyMap,
		styles: styles,
	}
}

// Init implements tea.Model
func (w FooterWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w FooterWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
	}
	
	return w, nil
}

// View implements tea.Model
func (w FooterWidget) View() string {
	if w.width == 0 {
		return ""
	}
	
	commandsText := "Commands: [R]efresh  [+/-]Speed  [Tab]Focus  [Q]uit"
	commands := w.styles.Commands.Render(commandsText)
	
	return w.styles.Container.Width(w.width).Render(commands)
}

// SetSize implements Widget interface
func (w *FooterWidget) SetSize(width, height int) {
	w.width = width
}

// SetFocused implements Widget interface
func (w *FooterWidget) SetFocused(focused bool) {
	// Footer doesn't need focus handling
}

// MinHeight implements Widget interface
func (w FooterWidget) MinHeight() int {
	return 1
}

// MaxHeight implements Widget interface
func (w FooterWidget) MaxHeight() int {
	return 1
}
