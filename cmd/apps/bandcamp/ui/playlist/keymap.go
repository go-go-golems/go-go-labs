package playlist

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	NextPage   key.Binding
	PrevPage   key.Binding
	GoToStart  key.Binding
	GoToEnd    key.Binding

	// Help toggle keybindings.
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding

	CancelFilePicker key.Binding

	ForceQuit key.Binding
	Quit      key.Binding

	OpenEntry   key.Binding
	DeleteEntry key.Binding
	OpenSearch  key.Binding
	AssignColor key.Binding

	MoveEntryUp   key.Binding
	MoveEntryDown key.Binding

	Export key.Binding

	Save key.Binding
	Load key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		CursorUp: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "Move cursor up"),
		),
		CursorDown: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "Move cursor down"),
		),
		ForceQuit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "Force quit"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q"),
			key.WithHelp("q", "Quit"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("left", "h", "pgup", "b", "u"),
			key.WithHelp("←/h/pgup", "prev page"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("right", "l", "pgdown", "f", "d"),
			key.WithHelp("→/l/pgdn", "next page"),
		),
		GoToStart: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "go to start"),
		),
		GoToEnd: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "go to end"),
		),

		OpenEntry: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "OpenEntry"),
		),

		MoveEntryUp: key.NewBinding(
			key.WithKeys("shift+up"),
			key.WithHelp("shift+up", "Move entry up"),
		),
		MoveEntryDown: key.NewBinding(
			key.WithKeys("shift+down"),
			key.WithHelp("shift+down", "Move entry down"),
		),

		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),

		DeleteEntry: key.NewBinding(
			key.WithKeys("delete", "x"),
			key.WithHelp("delete/x", "remove entry"),
		),
		OpenSearch: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		AssignColor: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "assign color"),
		),
		Export: key.NewBinding(
			key.WithKeys("e"),
			key.WithHelp("e", "export"),
		),
		Save: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "save"),
		),
		Load: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "load"),
		),

		CancelFilePicker: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "cancel file picker"),
		),
	}
}
