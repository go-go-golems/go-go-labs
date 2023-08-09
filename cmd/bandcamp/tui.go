package main

import (
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
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
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
		// Filtering.
		CancelWhileSearching: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		AcceptWhileSearching: key.NewBinding(
			key.WithKeys("enter", "tab", "shift+tab", "ctrl+k", "up", "ctrl+j", "down"),
			key.WithHelp("enter", "apply filter"),
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
	l       list.Model
	// TODO(manuel, 2023-08-09) We can actually use the help widget from the list by passing our own keys using AdditionalShortHelpKeys and such
	// however, not sure if this allows us to override the whole filtering stuff
	Help        help.Model
	SearchInput textinput.Model
	height      int
	width       int
	KeyMap
}

func NewModel(results []*Result) Model {
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
		case key.Matches(msg, m.KeyMap.Quit):
			// finish
			return m, tea.Quit
		case key.Matches(msg, m.KeyMap.OpenEntry):
			// open the selected item by using the os open for s.ItemURLPath
			url := m.results[m.l.Cursor()].ItemURLPath
			if err := openURL(url); err != nil {
				return m, tea.Quit
			}

		case key.Matches(msg, m.KeyMap.ShowFullHelp):
			fallthrough
		case key.Matches(msg, m.KeyMap.CloseFullHelp):
			m.Help.ShowAll = !m.Help.ShowAll

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
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.Help.Width = msg.Width
		m.SearchInput.Width = msg.Width
		m.height = msg.Height
		availHeight := m.height
		availHeight -= lipgloss.Height(m.Help.View(m))
		availHeight -= lipgloss.Height(m.SearchInput.View())
		_, v := docStyle.GetFrameSize()
		availHeight -= v
		m.l.SetSize(msg.Width, availHeight)
	}

	return m, tea.Batch(cmds...)
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

	view_ := m.SearchInput.View()
	sections = append(sections, view_)
	availHeight -= lipgloss.Height(view_)

	list_ := docStyle.Render(m.l.View())
	availHeight -= lipgloss.Height(list_)

	sections = append(sections, list_)

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
