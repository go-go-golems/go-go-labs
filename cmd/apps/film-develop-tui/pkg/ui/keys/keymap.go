package keys

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the application
type KeyMap struct {
	// Navigation
	Quit   key.Binding
	Back   key.Binding
	Enter  key.Binding
	Up     key.Binding
	Down   key.Binding
	
	// Film Selection
	Film     key.Binding
	Settings key.Binding
	
	// Number keys for selection
	Key1 key.Binding
	Key2 key.Binding
	Key3 key.Binding
	Key4 key.Binding
	Key5 key.Binding
	Key6 key.Binding
	Key7 key.Binding
	
	// Letter keys for 120mm rolls
	KeyA key.Binding
	KeyB key.Binding
	KeyC key.Binding
	KeyD key.Binding
	KeyE key.Binding
	KeyF key.Binding
	
	// Special actions
	Mixed      key.Binding
	UseFixer   key.Binding
	ChangeRoll key.Binding
	Reset      key.Binding
	Plus       key.Binding
	Minus      key.Binding
}

// NewKeyMap creates a new KeyMap with default bindings
func NewKeyMap() KeyMap {
	return KeyMap{
		// Navigation
		Quit:  key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Back:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Enter: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "confirm")),
		Up:    key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:  key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		
		// Main actions
		Film:     key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "film type")),
		Settings: key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "settings")),
		
		// Number keys
		Key1: key.NewBinding(key.WithKeys("1"), key.WithHelp("1", "select 1")),
		Key2: key.NewBinding(key.WithKeys("2"), key.WithHelp("2", "select 2")),
		Key3: key.NewBinding(key.WithKeys("3"), key.WithHelp("3", "select 3")),
		Key4: key.NewBinding(key.WithKeys("4"), key.WithHelp("4", "select 4")),
		Key5: key.NewBinding(key.WithKeys("5"), key.WithHelp("5", "select 5")),
		Key6: key.NewBinding(key.WithKeys("6"), key.WithHelp("6", "select 6")),
		Key7: key.NewBinding(key.WithKeys("7"), key.WithHelp("7", "select 7")),
		
		// Letter keys
		KeyA: key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "120mm 1 roll")),
		KeyB: key.NewBinding(key.WithKeys("b"), key.WithHelp("b", "120mm 2 rolls")),
		KeyC: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "120mm 3 rolls")),
		KeyD: key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "120mm 4 rolls")),
		KeyE: key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "120mm 5 rolls")),
		KeyF: key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "120mm 6 rolls")),
		
		// Special actions
		Mixed:      key.NewBinding(key.WithKeys("m"), key.WithHelp("m", "mixed rolls")),
		UseFixer:   key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "use fixer")),
		ChangeRoll: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "change rolls")),
		Reset:      key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "reset")),
		Plus:       key.NewBinding(key.WithKeys("+", "="), key.WithHelp("+", "increase")),
		Minus:      key.NewBinding(key.WithKeys("-", "_"), key.WithHelp("-", "decrease")),
	}
}

// ShortHelp returns keybindings to be shown in the mini help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Film, k.Settings, k.Quit}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Film, k.Settings, k.Back},
		{k.Up, k.Down, k.Enter},
		{k.Quit},
	}
}
