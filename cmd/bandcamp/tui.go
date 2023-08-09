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
	Quit       key.Binding
	Open       key.Binding
	Search     key.Binding
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
		Open: key.NewBinding(
			key.WithKeys("o"),
			key.WithHelp("o", "Open"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "Search"),
		),
	}

}

type Model struct {
	results     []*Result
	cursor      int
	l           list.Model
	Help        help.Model
	SearchInput textinput.Model
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
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			// finish
			return m, tea.Quit
		case "up":
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.results) - 1
			}
			cmd = nil
		case "down":
			m.cursor++
			if m.cursor >= len(m.results) {
				m.cursor = 0
			}
			cmd = nil
		case "o":
			// open the selected item by using the os open for s.ItemURLPath
			cmd = nil
			url := m.results[m.cursor].ItemURLPath
			if err := openURL(url); err != nil {
				return m, tea.Quit
			}

		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.l.SetSize(msg.Width-h, msg.Height-v)
	}

	// Use the list's built in update method to handle list events
	m.l, cmd = m.l.Update(msg)

	return m, cmd
}
func (m Model) ShortHelp() []key.Binding {
	return []key.Binding{
		m.CursorUp,
		m.CursorDown,
		m.Open,
		m.Quit,
	}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{
			m.CursorUp,
			m.CursorDown,
			m.Open,
			m.Quit,
		},
	}
}

func (m Model) helpView() string {
	return helpStyle.Render(m.Help.View(m))
}

func (m Model) View() string {
	//availHeight := m.height
	//help
	return docStyle.Render(m.l.View())
}
