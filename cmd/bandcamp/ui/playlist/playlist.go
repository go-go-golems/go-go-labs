package playlist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/ui"
	"github.com/go-go-golems/go-go-labs/cmd/bandcamp/ui/search"
	"github.com/pkg/errors"
	"time"
)

// states

type state int

const (
	stateList             state = iota
	stateFilePickerSave   state = iota
	stateFilePickerExport state = iota
	stateFilePickerLoad   state = iota
	stateSearch           state = iota
)

type Track pkg.Track

func (s *Track) FilterValue() string {
	return s.Name
}

func (s *Track) Title() string {
	// NOTE(manuel, 2023-08-13) we probably need to set up the color here
	return fmt.Sprintf("%s - %s", s.BandName, s.Name)
}

func (s *Track) Description() string {
	return s.ItemURLPath
}

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

	CancelSearch     key.Binding
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
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc/ctrl+c", "Quit"),
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

		CancelSearch: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "cancel search"),
		),
		CancelFilePicker: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "cancel file picker"),
		),
	}
}

type Model struct {
	Playlist *pkg.Playlist

	l list.Model

	filepicker filepicker.Model
	KeyMap     KeyMap
	search     search.Model
	state      state

	selectedFile string

	err error
}

func (m *Model) updateListItems() {
	items := make([]list.Item, len(m.Playlist.Tracks))
	tracks_ := make([]*Track, len(m.Playlist.Tracks))

	for i, track := range m.Playlist.Tracks {
		t := Track(*track)
		tracks_[i] = &t
		items[i] = &t
	}

	m.l.SetItems(items)
}

func NewModel(playlist *pkg.Playlist) Model {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)

	keymap := DefaultKeyMap()

	l.Title = playlist.Title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keymap.MoveEntryUp,
			keymap.MoveEntryDown,

			keymap.AssignColor,
			keymap.OpenSearch,
			keymap.Export,
			keymap.Save,
			keymap.Load,

			keymap.ShowFullHelp,
			keymap.Quit,
		}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keymap.CloseFullHelp,
		}
	}
	l.SetShowHelp(true)

	client := pkg.NewClient()
	s := search.NewModel(client, []*pkg.Result{})

	m := Model{
		Playlist: playlist,
		l:        l,
		KeyMap:   keymap,
		search:   s,
		state:    stateList,
	}
	m.updateListItems()
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.ForceQuit):
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.Quit):
			if m.state == stateList {
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.l.SetSize(msg.Width, msg.Height)
		m.search.SetSize(msg.Width, msg.Height)
		m.filepicker.Height = msg.Height
	}

	switch m.state {
	case stateList:
		cmds_ := m.updateList(msg)
		cmds = append(cmds, cmds_...)
	case stateSearch:
		cmds_ := m.updateSearch(msg)
		cmds = append(cmds, cmds_...)
	case stateFilePickerExport:
		fallthrough
	case stateFilePickerLoad:
		fallthrough
	case stateFilePickerSave:
		cmds_ := m.updateFilePicker(msg)
		cmds = append(cmds, cmds_...)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateList(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.OpenEntry):
			url := m.Playlist.Tracks[m.l.Index()].ItemURLPath
			if err := pkg.OpenURL(url); err != nil {
				return cmds
			}
		case key.Matches(msg, m.KeyMap.MoveEntryUp):
			newIndex := m.Playlist.MoveEntryUp(m.l.Index())
			// updateIndex
			m.updateListItems()
			m.l.Select(newIndex)
		case key.Matches(msg, m.KeyMap.MoveEntryDown):
			newIndex := m.Playlist.MoveEntryDown(m.l.Index())
			m.updateListItems()
			m.l.Select(newIndex)
		case key.Matches(msg, m.KeyMap.DeleteEntry):
			m.Playlist.DeleteEntry(m.l.Index())
			m.updateListItems()

		case key.Matches(msg, m.KeyMap.OpenSearch):
			m.state = stateSearch
			// NOTE(manuel, 2023-08-13) trigger opening the search bar
			cmds = append(cmds, m.search.Init())
		case key.Matches(msg, m.KeyMap.Export):
			m.state = stateFilePickerExport
			cmds = append(cmds, m.filepicker.Init())

		case key.Matches(msg, m.KeyMap.Save):
			// TODO(manuel, 2023-08-13) Handle save as new and normal save, for now always save as new
			// we probably also want to cache the last name saved to so that we can scroll right to it
			m.state = stateFilePickerSave
			cmds = append(cmds, m.filepicker.Init())
		case key.Matches(msg, m.KeyMap.Load):
			m.state = stateFilePickerLoad
			cmds = append(cmds, m.filepicker.Init())

		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CursorUp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CursorDown):
			fallthrough
		case key.Matches(msg, m.KeyMap.NextPage):
			fallthrough
		case key.Matches(msg, m.KeyMap.PrevPage):
			fallthrough
		case key.Matches(msg, m.KeyMap.GoToStart):
			fallthrough
		case key.Matches(msg, m.KeyMap.GoToEnd):
			listModel, cmd := m.l.Update(msg)
			m.l = listModel
			cmds = append(cmds, cmd)
		}
	}

	return cmds
}

func (m *Model) updateSearch(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.CancelSearch):
			m.state = stateList
			return cmds
		}
	}

	searchModel, cmd := m.search.Update(msg)
	cmds = append(cmds, cmd)
	m.search = searchModel

	return cmds
}

func (m *Model) updateFilePicker(msg tea.Msg) []tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.CancelFilePicker):
			m.state = stateList

		default:
			filePickerModel, cmd := m.filepicker.Update(msg)
			cmds = append(cmds, cmd)
			m.filepicker = filePickerModel

			if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
				// Get the path of the selected file.
				m.selectedFile = path

				// TODO(manuel, 2023-08-13) Actually handle what it means to select a file
			}

			if didSelect, path := m.filepicker.DidSelectDisabledFile(msg); didSelect {
				// Let's clear the selectedFile and display an error.
				m.err = errors.New(path + " is not valid.")
				m.selectedFile = ""
				cmds = append(cmds, tea.Batch(cmd, ui.ClearErrorAfter(2*time.Second)))
			}
		}
	}

	return cmds
}

func (m Model) View() string {
	switch m.state {
	case stateList:
		return m.l.View()

	case stateSearch:
		return m.search.View()

	case stateFilePickerExport:
		fallthrough
	case stateFilePickerLoad:
		fallthrough
	case stateFilePickerSave:
		return m.filepicker.View()
	}

	return "DEFAULT\n"
}
