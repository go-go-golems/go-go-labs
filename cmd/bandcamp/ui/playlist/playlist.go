package playlist

import (
	"fmt"
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

func (m *Model) updateListItems() tea.Cmd {
	items := make([]list.Item, len(m.Playlist.Tracks))
	tracks_ := make([]*Track, len(m.Playlist.Tracks))

	for i, track := range m.Playlist.Tracks {
		t := Track(*track)
		tracks_[i] = &t
		items[i] = &t
	}
	hasItems := len(items) > 0

	m.KeyMap.AssignColor.SetEnabled(hasItems)
	m.KeyMap.MoveEntryDown.SetEnabled(hasItems)
	m.KeyMap.MoveEntryUp.SetEnabled(hasItems)
	m.KeyMap.DeleteEntry.SetEnabled(hasItems)
	m.KeyMap.OpenEntry.SetEnabled(hasItems)

	if m.l.Index() >= len(items) {
		m.l.Select(len(items) - 1)
	}
	return m.l.SetItems(items)
}

var (
	playlistNameStyle = ui.MainTitleStyle
	titleStyle        = lipgloss.NewStyle().
				Foreground(lipgloss.Color("230"))
	appStyle = lipgloss.NewStyle().
			Margin(1, 1, 1, 1)
)

func NewModel(playlist *pkg.Playlist) Model {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)

	keymap := DefaultKeyMap()

	l.DisableQuitKeybindings()
	l.Styles.Title = titleStyle
	l.Title = fmt.Sprintf(
		"%s%s",
		"Edit Playlist: ",
		playlistNameStyle.Render(playlist.Title))
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keymap.AssignColor,
			keymap.DeleteEntry,
			keymap.OpenSearch,
			keymap.Quit,
		}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			keymap.MoveEntryUp,
			keymap.MoveEntryDown,

			keymap.Export,
			keymap.Save,
			keymap.Load,
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
		h, v := appStyle.GetFrameSize()
		newWidth := msg.Width - h
		newHeight := msg.Height - v
		m.l.SetSize(newWidth, newHeight)
		m.search.SetSize(newWidth, newHeight)
		m.filepicker.Height = newHeight

	case ui.InsertPlaylistEntryMsg:
		m.Playlist.InsertTrack(msg.Track, m.l.Index())
		m.state = stateList
		cmd := m.updateListItems()
		return m, cmd
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
			cmd := m.updateListItems()
			cmds = append(cmds, cmd)
			m.l.Select(newIndex)
		case key.Matches(msg, m.KeyMap.MoveEntryDown):
			newIndex := m.Playlist.MoveEntryDown(m.l.Index())
			cmd := m.updateListItems()
			cmds = append(cmds, cmd)
			m.l.Select(newIndex)

		case key.Matches(msg, m.KeyMap.DeleteEntry):
			m.Playlist.DeleteEntry(m.l.Index())
			cmd := m.updateListItems()
			cmds = append(cmds, cmd)

		case key.Matches(msg, m.KeyMap.OpenSearch):
			m.state = stateSearch
			// NOTE(manuel, 2023-08-13) trigger opening the search bar
			m.search.SetShowSearch(true)
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

	switch v := msg.(type) {
	case search.SelectEntryMsg:
		track := &pkg.Track{
			BackgroundColor: "black",
			LinkColor:       "white",
			AlbumID:         v.Result.AlbumID,
			Name:            v.Result.Name,
			BandName:        v.Result.BandName,
			ItemURLPath:     v.Result.ItemURLPath,
		}
		m.state = stateList
		m.updateListItems()
		return []tea.Cmd{
			func() tea.Msg {
				return ui.InsertPlaylistEntryMsg{Track: track}
			},
		}
	case search.CloseSearchMsg:
		m.state = stateList
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
	res := ""

	switch m.state {
	case stateList:
		res = m.l.View()

	case stateSearch:
		res = m.search.View()

	case stateFilePickerExport:
		fallthrough
	case stateFilePickerLoad:
		fallthrough
	case stateFilePickerSave:
		res = m.filepicker.View()
	}

	return appStyle.Render(res)
}
