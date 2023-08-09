package main

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"os/exec"
	"runtime"
)

var (
	docStyle         = lipgloss.NewStyle().Margin(1, 2)
	titleStyle       = lipgloss.NewStyle().MarginLeft(2)
	searchInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))
	searchInputPromptStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205"))
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default: // for linux and unix
		cmd = exec.Command("xdg-open", url)
	}

	return cmd.Start()
}

func (s *Result) FilterValue() string {
	return fmt.Sprintf("%d", s.ID)
}

func (s *Result) Title() string {
	switch SearchType(s.Type) {
	case Track:
		return fmt.Sprintf("%s - %s (%s)", s.BandName, s.Name, s.AlbumName)
	case Album:
		return fmt.Sprintf("%s - %s", s.BandName, s.Name)
	case Band:
		return fmt.Sprintf("%s", s.BandName)
	default:
		return fmt.Sprintf("%s", s.Name)
	}
}

func (s *Result) Description() string {
	return s.ItemURLPath
}

type KeyMap struct {
	CursorUp   key.Binding
	CursorDown key.Binding
	NextPage   key.Binding
	PrevPage   key.Binding
	GoToStart  key.Binding
	GoToEnd    key.Binding

	// Keybindings used when setting a filter.
	CancelWhileSearching key.Binding
	AcceptWhileSearching key.Binding

	// Help toggle keybindings.
	ShowFullHelp  key.Binding
	CloseFullHelp key.Binding

	ForceQuit key.Binding
	Quit      key.Binding
	OpenEntry key.Binding
	Search    key.Binding
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
		OpenEntry: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "OpenEntry"),
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
		// Toggle help.
		ShowFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "more"),
		),
		CloseFullHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "close help"),
		),
	}

}

type Model struct {
	results []*Result

	client *Client

	l list.Model
	// TODO(manuel, 2023-08-09) We can actually use the help widget from the list by passing our own keys using AdditionalShortHelpKeys and such
	// however, not sure if this allows us to override the whole filtering stuff
	Help        help.Model
	SearchInput textinput.Model
	height      int
	width       int
	KeyMap

	// TODO(manuel, 2023-08-09) Add a spinner

	showSearch bool
}

func NewModel(client *Client, results []*Result) Model {
	items := make([]list.Item, len(results))

	for i, result := range results {
		items[i] = result
	}

	h := help.New()

	searchInput := textinput.New()
	searchInput.Prompt = "Search: "
	searchInput.PromptStyle = searchInputStyle
	searchInput.Focus()

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)

	l.Title = "Select next playlist track"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)

	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	return Model{
		client:      client,
		results:     results,
		l:           l,
		Help:        h,
		SearchInput: searchInput,
		KeyMap:      DefaultKeyMap(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.ForceQuit):
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.Help.Width = msg.Width
		m.SearchInput.Width = msg.Width
		m.height = msg.Height
		availHeight := m.height
		availHeight -= lipgloss.Height(m.Help.View(m))
		if m.showSearch {
			availHeight -= lipgloss.Height(m.SearchInput.View())
		}
		_, v := docStyle.GetFrameSize()
		availHeight -= v
		m.l.SetSize(msg.Width, availHeight)
	}

	if m.showSearch {
		cmds_ := m.updateSearch(msg)
		cmds = append(cmds, cmds_)
	} else {
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
			m.showSearch = false
			m.updateKeyBindings()

		case key.Matches(msg, m.KeyMap.AcceptWhileSearching):
			searchTerm := m.SearchInput.Value()
			resp, err := m.client.Search(context.Background(), searchTerm, Track)
			if err != nil {
				return tea.Quit
			}
			items := make([]list.Item, len(resp.Auto.Results))

			for i, result := range resp.Auto.Results {
				items[i] = result
			}

			m.l.SetItems(items)

			m.showSearch = false
			m.updateKeyBindings()
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

func (m *Model) updateList(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Quit):
			// finish
			return tea.Quit
		case key.Matches(msg, m.KeyMap.OpenEntry):
			// open the selected item by using the os open for s.ItemURLPath
			url := m.results[m.l.Index()].ItemURLPath
			if err := openURL(url); err != nil {
				return tea.Quit
			}

		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll

		case key.Matches(msg, m.KeyMap.Search):
			m.showSearch = true
			m.SearchInput.CursorEnd()
			m.SearchInput.Focus()
			m.updateKeyBindings()

			// forward to list
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

	return tea.Batch(cmds...)
}

func (m *Model) updateKeyBindings() {
	if m.showSearch {
		m.KeyMap.CursorUp.SetEnabled(false)
		m.KeyMap.CursorDown.SetEnabled(false)
		m.KeyMap.NextPage.SetEnabled(false)
		m.KeyMap.PrevPage.SetEnabled(false)
		m.KeyMap.GoToStart.SetEnabled(false)
		m.KeyMap.GoToEnd.SetEnabled(false)
		m.KeyMap.Search.SetEnabled(false)
		m.KeyMap.CancelWhileSearching.SetEnabled(true)
		m.KeyMap.AcceptWhileSearching.SetEnabled(m.SearchInput.Value() != "")
		m.KeyMap.Quit.SetEnabled(false)
		m.KeyMap.ShowFullHelp.SetEnabled(false)
		m.KeyMap.CloseFullHelp.SetEnabled(false)
	} else {
		hasItems := len(m.results) != 0
		m.KeyMap.CursorUp.SetEnabled(hasItems)
		m.KeyMap.CursorDown.SetEnabled(hasItems)

		hasPages := m.l.Paginator.TotalPages > 1
		m.KeyMap.NextPage.SetEnabled(hasPages)
		m.KeyMap.PrevPage.SetEnabled(hasPages)

		m.KeyMap.GoToStart.SetEnabled(hasItems)
		m.KeyMap.GoToEnd.SetEnabled(hasItems)

		m.KeyMap.Search.SetEnabled(true)
		m.KeyMap.CancelWhileSearching.SetEnabled(false)
		m.KeyMap.AcceptWhileSearching.SetEnabled(false)
		m.KeyMap.Quit.SetEnabled(true)

		m.KeyMap.ShowFullHelp.SetEnabled(true)
		m.KeyMap.CloseFullHelp.SetEnabled(true)
	}
}

func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.CursorUp,
		m.CursorDown,
		m.OpenEntry,
		m.Quit,
	}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.CursorUp,
			m.CursorDown,
			m.NextPage,
			m.PrevPage,
			m.GoToStart,
			m.GoToEnd,
			m.Search,
			m.ShowFullHelp,

			m.OpenEntry,
			m.Quit,
		},
	}
}

func (m Model) helpView() string {
	return helpStyle.Render(m.Help.View(m))
}

func (m Model) View() string {
	sections := []string{}

	help_ := m.helpView()
	availHeight := m.height
	availHeight -= lipgloss.Height(help_)
	sections = append(sections, help_)

	if m.showSearch {
		view_ := m.SearchInput.View()
		sections = append(sections, view_)
		availHeight -= lipgloss.Height(view_)
	}

	list_ := docStyle.Render(m.l.View())
	availHeight -= lipgloss.Height(list_)

	sections = append(sections, list_)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
