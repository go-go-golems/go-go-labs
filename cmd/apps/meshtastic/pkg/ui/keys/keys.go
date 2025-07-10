package keys

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
)

// KeyMap contains all key bindings for the TUI
type KeyMap struct {
	// Navigation
	Up    key.Binding
	Down  key.Binding
	Left  key.Binding
	Right key.Binding
	Tab   key.Binding
	Enter key.Binding

	// Tabs
	TabMessages  key.Binding
	TabNodes     key.Binding
	TabStatus    key.Binding
	TabTelemetry key.Binding
	TabConfig    key.Binding
	TabPosition  key.Binding

	// Actions
	Compose key.Binding
	Send    key.Binding

	// System
	Help key.Binding
	Quit key.Binding

	// Compose mode
	Escape key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() *KeyMap {
	return &KeyMap{
		// Navigation
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("↑/k", "move up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("↓/j", "move down"),
		),
		Left: key.NewBinding(
			key.WithKeys("h", "left"),
			key.WithHelp("←/h", "move left"),
		),
		Right: key.NewBinding(
			key.WithKeys("l", "right"),
			key.WithHelp("→/l", "move right"),
		),
		Tab: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next tab"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),

		// Tabs
		TabMessages: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "messages"),
		),
		TabNodes: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "nodes"),
		),
		TabStatus: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "status"),
		),
		TabTelemetry: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "telemetry"),
		),
		TabConfig: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "config"),
		),
		TabPosition: key.NewBinding(
			key.WithKeys("6"),
			key.WithHelp("6", "position"),
		),

		// Actions
		Compose: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "compose message"),
		),
		Send: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "send message"),
		),

		// System
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),

		// Compose mode
		Escape: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
	}
}

// ShortHelp returns the short help for the key map
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit, k.Tab, k.Compose}
}

// FullHelp returns the full help for the key map
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.Enter, k.Escape},
		{k.TabMessages, k.TabNodes, k.TabStatus, k.TabTelemetry, k.TabConfig, k.TabPosition},
		{k.Compose, k.Send},
		{k.Help, k.Quit},
	}
}

// HelpModel returns a help model for the key map
func (k KeyMap) HelpModel() help.Model {
	h := help.New()
	h.ShowAll = false
	return h
}
