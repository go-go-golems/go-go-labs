// Package keys defines keyboard mappings for the Redis monitor TUI
package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keyboard bindings
type KeyMap struct {
	Quit        key.Binding
	Refresh     key.Binding
	RefreshUp   key.Binding
	RefreshDown key.Binding
	FocusNext   key.Binding
	FocusPrev   key.Binding
	ScrollUp    key.Binding
	ScrollDown  key.Binding
	Help        key.Binding
}

// DefaultKeyMap returns the default key bindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "refresh"),
		),
		RefreshUp: key.NewBinding(
			key.WithKeys("+", "="),
			key.WithHelp("+", "faster refresh"),
		),
		RefreshDown: key.NewBinding(
			key.WithKeys("-", "_"),
			key.WithHelp("-", "slower refresh"),
		),
		FocusNext: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next widget"),
		),
		FocusPrev: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "prev widget"),
		),
		ScrollUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "scroll up"),
		),
		ScrollDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "scroll down"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Refresh, k.RefreshUp, k.RefreshDown, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Refresh, k.RefreshUp, k.RefreshDown},
		{k.FocusNext, k.FocusPrev, k.ScrollUp, k.ScrollDown},
		{k.Help, k.Quit},
	}
}
