package search

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	bandcamp "github.com/go-go-golems/go-go-labs/cmd/apps/bandcamp/pkg"
	"github.com/go-go-golems/go-go-labs/cmd/apps/bandcamp/ui"
)

type CloseSearchMsg struct{}
type SelectEntryMsg struct {
	Result *bandcamp.Result
}

type Result bandcamp.Result

func (s *Result) FilterValue() string {
	return fmt.Sprintf("%d", s.ID)
}

func (s *Result) Title() string {
	switch bandcamp.SearchType(s.Type) {
	case bandcamp.FilterTrack:
		return fmt.Sprintf("%s - %s (%s)", s.BandName, s.Name, s.AlbumName)
	case bandcamp.FilterAlbum:
		return fmt.Sprintf("%s - %s", s.BandName, s.Name)
	case bandcamp.FilterBand:
		return s.BandName
	case bandcamp.FilterAll:
		return fmt.Sprintf("%s - %s (%s)", s.BandName, s.Name, s.AlbumName)
	default:
		return s.Name
	}
}

func (s *Result) Description() string {
	return s.ItemURLPath
}

type KeyMap struct {
	// Keybindings used when setting a filter.
	CancelWhileSearching key.Binding
	AcceptWhileSearching key.Binding

	OpenEntry   key.Binding
	SelectEntry key.Binding
	Search      key.Binding
	CloseSearch key.Binding
}

func DefaultKeyMap() KeyMap {
	return KeyMap{
		OpenEntry: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "OpenEntry"),
		),
		// Searching
		CancelWhileSearching: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileSearching: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "run search"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "Search"),
		),
		SelectEntry: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "Select track"),
		),
		CloseSearch: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("q/esc", "Close search"),
		),
	}

}

var (
	titleStyle             = ui.MainTitleStyle
	titleBarStyle          = lipgloss.NewStyle().Padding(0, 0, 1, 2)
	searchInputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))
	searchInputStyle = lipgloss.NewStyle().
				PaddingLeft(2).PaddingBottom(1)
)

type Model struct {
	results []*Result

	client *bandcamp.Client

	l list.Model

	// TODO(manuel, 2023-08-09) We can actually use the help widget from the list
	// by passing our own keys using AdditionalShortHelpKeys and such
	// however, not sure if this allows us to override the whole filtering stuff
	SearchInput textinput.Model

	KeyMap

	// TODO(manuel, 2023-08-09) Add a spinner

	ShowSearch bool
	height     int
	width      int
}

func (m Model) GetResults() []*bandcamp.Result {
	ret := make([]*bandcamp.Result, len(m.results))
	for i, r := range m.results {
		ret[i] = (*bandcamp.Result)(r)
	}

	return ret
}

func (m Model) GetSelectedResult() *bandcamp.Result {
	idx := m.l.Index()
	if idx < 0 || idx >= len(m.results) {
		return nil
	}

	return (*bandcamp.Result)(m.results[idx])
}

func NewModel(client *bandcamp.Client, results []*bandcamp.Result) Model {
	items := make([]list.Item, len(results))
	results_ := make([]*Result, len(results))

	for i, result := range results {
		r := Result(*result)
		items[i] = &r
		results_[i] = &r
	}

	searchInput := textinput.New()
	searchInput.Prompt = "Search bandcamp: "
	searchInput.PromptStyle = searchInputPromptStyle
	searchInput.Focus()

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	l.SetShowTitle(false)
	l.Title = "Select next playlist track"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	keyMap := DefaultKeyMap()
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			keyMap.Search,
			keyMap.SelectEntry,
			keyMap.OpenEntry,
			keyMap.CloseSearch,
			keyMap.CancelWhileSearching,
			keyMap.AcceptWhileSearching,
		}
	}

	return Model{
		client:      client,
		results:     results_,
		l:           l,
		SearchInput: searchInput,
		KeyMap:      keyMap,
	}
}

func (m *Model) updateKeybindings() {
	if m.ShowSearch {
		m.KeyMap.CancelWhileSearching.SetEnabled(true)
		m.KeyMap.AcceptWhileSearching.SetEnabled(m.SearchInput.Value() != "")
		m.KeyMap.Search.SetEnabled(false)
		m.KeyMap.SelectEntry.SetEnabled(false)
		m.KeyMap.OpenEntry.SetEnabled(false)
		m.KeyMap.CloseSearch.SetEnabled(true)
	} else {
		m.KeyMap.CancelWhileSearching.SetEnabled(false)
		m.KeyMap.AcceptWhileSearching.SetEnabled(false)
		m.KeyMap.Search.SetEnabled(true)
		m.KeyMap.SelectEntry.SetEnabled(true)
		m.KeyMap.OpenEntry.SetEnabled(true)
		m.KeyMap.CloseSearch.SetEnabled(false)
	}
}

func (m Model) GetSearchTerm() string {
	return m.SearchInput.Value()
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m *Model) SetSize(width, height int) {
	m.height = height
	m.width = width
	m.SearchInput.Width = width
	m.recomputeHeight()
}

func (m *Model) SetShowSearch(v bool) {
	m.ShowSearch = v
	if !v {
		m.SearchInput.Blur()
	}
	m.recomputeHeight()
	m.updateKeybindings()
}

func (m *Model) recomputeHeight() {
	availHeight := m.height
	if m.ShowSearch {
		view_ := searchInputStyle.Render(m.SearchInput.View())
		availHeight -= lipgloss.Height(view_)
	} else {
		title_ := titleBarStyle.Render(titleStyle.Render("test"))
		titleHeight := lipgloss.Height(title_)
		availHeight -= titleHeight
	}
	m.l.SetSize(m.width, availHeight)
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.CloseSearch):
			return m, func() tea.Msg { return CloseSearchMsg{} }
		}
	case tea.WindowSizeMsg:
		m.SetSize(msg.Width, msg.Height)

	case ui.UpdateSearchResultsMsg:
		items := make([]list.Item, len(msg.Results))
		results := make([]*Result, len(msg.Results))
		for i, result := range msg.Results {
			items[i] = (*Result)(result)
			results[i] = (*Result)(result)
		}
		m.SetShowSearch(false)
		cmd := m.l.SetItems(items)
		m.results = results
		cmds = append(cmds, cmd)
		m.updateKeyBindings()
	}

	if m.ShowSearch {
		m.updateKeybindings()
		cmds_ := m.updateSearch(msg)
		cmds = append(cmds, cmds_)
	} else {
		m.updateKeyBindings()
		cmds_ := m.updateList(msg)
		cmds = append(cmds, cmds_)
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) updateSearch(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.CancelWhileSearching):
			m.SetShowSearch(false)
			return nil

		case key.Matches(msg, m.KeyMap.AcceptWhileSearching):
			searchTerm := m.SearchInput.Value()

			m.SetShowSearch(false)

			return func() tea.Msg {
				return m.SearchBandcamp(searchTerm)
			}
		}

		newSearchInputModel, inputCmd := m.SearchInput.Update(msg)
		searchChanged := newSearchInputModel.Value() != m.SearchInput.Value()
		m.SearchInput = newSearchInputModel

		if searchChanged {
			m.KeyMap.AcceptWhileSearching.SetEnabled(m.SearchInput.Value() != "")
		}
		cmds = append(cmds, inputCmd)
	}

	return tea.Batch(cmds...)
}

func (m Model) SearchBandcamp(searchTerm string) tea.Msg {
	resp, err := m.client.Search(context.Background(), searchTerm, bandcamp.FilterTrack)
	if err != nil {
		return ui.ErrMsg{Err: err}
	}

	return ui.UpdateSearchResultsMsg{Results: resp.Auto.Results}
}

func (m *Model) updateList(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.CloseSearch):
			// finish
			return func() tea.Msg {
				return CloseSearchMsg{}
			}

		case key.Matches(msg, m.KeyMap.OpenEntry):
			// open the selected item by using the os open for s.ItemURLPath
			idx := m.l.Index()
			if idx < 0 || idx >= len(m.results) {
				return nil
			}
			url := m.results[m.l.Index()].ItemURLPath
			if err := bandcamp.OpenURL(url); err != nil {
				return tea.Quit
			}

		case key.Matches(msg, m.SelectEntry):
			return func() tea.Msg {
				if m.l.Index() < 0 || m.l.Index() >= len(m.results) {
					return CloseSearchMsg{}
				}
				return SelectEntryMsg{
					Result: (*bandcamp.Result)(m.results[m.l.Index()]),
				}
			}

		case key.Matches(msg, m.KeyMap.Search):
			m.SetShowSearch(true)
			m.SearchInput.CursorEnd()
			m.SearchInput.Focus()
			m.SearchInput.SetValue("")

		// forward to list
		default:
			if !m.ShowSearch {
				listModel, cmd := m.l.Update(msg)
				m.l = listModel
				cmds = append(cmds, cmd)
			}
		}
	}

	return tea.Batch(cmds...)
}

func (m *Model) updateKeyBindings() {
	if m.ShowSearch {
		m.KeyMap.Search.SetEnabled(false)
		m.KeyMap.CancelWhileSearching.SetEnabled(true)
		m.KeyMap.AcceptWhileSearching.SetEnabled(m.SearchInput.Value() != "")
		m.KeyMap.CloseSearch.SetEnabled(false)
		m.KeyMap.OpenEntry.SetEnabled(false)
		m.KeyMap.SelectEntry.SetEnabled(false)
	} else {
		hasItems := len(m.results) != 0
		m.KeyMap.OpenEntry.SetEnabled(hasItems)
		m.KeyMap.SelectEntry.SetEnabled(hasItems)

		m.KeyMap.Search.SetEnabled(true)
		m.KeyMap.CancelWhileSearching.SetEnabled(false)
		m.KeyMap.AcceptWhileSearching.SetEnabled(false)

		m.KeyMap.CloseSearch.SetEnabled(true)
	}
}

func (m Model) View() string {
	sections := []string{}

	title := "Search bandcamp:"
	if !m.ShowSearch {
		title = "Add track to playlist:"
	}

	if m.ShowSearch {
		view_ := searchInputStyle.Render(m.SearchInput.View())
		sections = append(sections, view_)
	} else {
		title_ := titleStyle.Render(title)
		title_ = titleBarStyle.Render(title_)
		sections = append(sections, title_)
	}

	list_ := m.l.View()

	sections = append(sections, list_)

	ret := lipgloss.JoinVertical(lipgloss.Left, sections...)

	return ret
}
