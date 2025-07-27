package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// NavigationView manages the bottom help/navigation bar
type NavigationView struct {
	styles styles.Styles
	help   help.Model
	keys   keyMap
	width  int
	height int
}

// keyMap defines the key bindings
type keyMap struct {
	Refresh   key.Binding
	Groups    key.Binding
	Streams   key.Binding
	Metrics   key.Binding
	Up        key.Binding
	Down      key.Binding
	SpeedUp   key.Binding
	SpeedDown key.Binding
	Quit      key.Binding
	Help      key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Refresh, k.Groups, k.Streams, k.SpeedUp, k.SpeedDown, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Refresh, k.Groups, k.Streams, k.Metrics},
		{k.Up, k.Down, k.SpeedUp, k.SpeedDown},
		{k.Quit, k.Help},
	}
}

// NewNavigationView creates a new navigation view
func NewNavigationView(styles styles.Styles) *NavigationView {
	h := help.New()
	h.ShowAll = false

	keys := keyMap{
		Refresh: key.NewBinding(
			key.WithKeys("r", "R"),
			key.WithHelp("r", "refresh"),
		),
		Groups: key.NewBinding(
			key.WithKeys("g", "G"),
			key.WithHelp("g", "groups view"),
		),
		Streams: key.NewBinding(
			key.WithKeys("s", "S"),
			key.WithHelp("s", "streams view"),
		),
		Metrics: key.NewBinding(
			key.WithKeys("m", "M"),
			key.WithHelp("m", "metrics view"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		SpeedUp: key.NewBinding(
			key.WithKeys(">", "."),
			key.WithHelp(">", "speed up"),
		),
		SpeedDown: key.NewBinding(
			key.WithKeys("<", ","),
			key.WithHelp("<", "speed down"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
	}

	return &NavigationView{
		styles: styles,
		help:   h,
		keys:   keys,
	}
}

// Init implements tea.Model
func (v *NavigationView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (v *NavigationView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.help.Width = msg.Width

	case tea.KeyMsg:
		if msg.String() == "?" {
			v.help.ShowAll = !v.help.ShowAll
		}
	}

	return v, nil
}

// View implements tea.Model
func (v *NavigationView) View() string {
	return v.help.View(v.keys)
}

// GetKeys returns the key bindings
func (v *NavigationView) GetKeys() keyMap {
	return v.keys
}

// ToggleHelp toggles between short and full help
func (v *NavigationView) ToggleHelp() {
	v.help.ShowAll = !v.help.ShowAll
}

// IsHelpVisible returns whether full help is currently shown
func (v *NavigationView) IsHelpVisible() bool {
	return v.help.ShowAll
}
