package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
)

var _ tea.Model = (*SelectionModel)(nil)

type SelectionModel struct {
	config     *config.Config
	width      int
	height     int
	
	// UI components
	searchInput    textinput.Model
	repoList       list.Model
	presetButtons  []presetButton
	selectedPreset int
	
	// State
	repositories   []repositoryItem
	filteredRepos  []repositoryItem
	selectedRepos  map[string]bool
	
	// Key bindings
	keys selectionKeyMap
}

type selectionKeyMap struct {
	Toggle    key.Binding
	Search    key.Binding
	Continue  key.Binding
	Preset    key.Binding
	ClearAll  key.Binding
	SelectAll key.Binding
}

var selectionKeys = selectionKeyMap{
	Toggle: key.NewBinding(
		key.WithKeys(" ", "enter"),
		key.WithHelp("space/enter", "toggle selection"),
	),
	Search: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "search"),
	),
	Continue: key.NewBinding(
		key.WithKeys("c", "tab"),
		key.WithHelp("c/tab", "continue"),
	),
	Preset: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "select preset"),
	),
	ClearAll: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "clear all"),
	),
	SelectAll: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "select all"),
	),
}

type repositoryItem struct {
	repo     config.Repository
	selected bool
}

func (r repositoryItem) FilterValue() string {
	return r.repo.Name + " " + r.repo.Description + " " + strings.Join(r.repo.Tags, " ")
}

func (r repositoryItem) Title() string {
	checkbox := "☐"
	if r.selected {
		checkbox = "☑"
	}
	return fmt.Sprintf("%s %s", checkbox, r.repo.Name)
}

func (r repositoryItem) Description() string {
	return r.repo.Description
}

type presetButton struct {
	name     string
	selected bool
	preset   config.Preset
}

func NewSelectionModel(cfg *config.Config) *SelectionModel {
	// Initialize search input
	searchInput := textinput.New()
	searchInput.Placeholder = "Search repositories..."
	searchInput.CharLimit = 50

	// Convert repositories to items
	repos := make([]repositoryItem, len(cfg.Repositories))
	for i, repo := range cfg.Repositories {
		repos[i] = repositoryItem{repo: repo, selected: false}
	}

	// Initialize list
	items := make([]list.Item, len(repos))
	for i, repo := range repos {
		items[i] = repo
	}

	repoList := list.New(items, newRepositoryDelegate(), 0, 0)
	repoList.Title = "Repositories"
	repoList.SetShowStatusBar(false)
	repoList.SetFilteringEnabled(false) // We'll handle our own filtering
	repoList.Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(0, 1)

	// Initialize preset buttons
	presetButtons := make([]presetButton, len(cfg.Presets))
	for i, preset := range cfg.Presets {
		presetButtons[i] = presetButton{
			name:     preset.Name,
			selected: false,
			preset:   preset,
		}
	}

	return &SelectionModel{
		config:        cfg,
		searchInput:   searchInput,
		repoList:      repoList,
		presetButtons: presetButtons,
		repositories:  repos,
		filteredRepos: repos,
		selectedRepos: make(map[string]bool),
		keys:          selectionKeys,
	}
}

func (m *SelectionModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *SelectionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.searchInput.Focused() {
			switch msg.String() {
			case "esc":
				m.searchInput.Blur()
				m.updateFilter()
				return m, nil
			case "enter":
				m.searchInput.Blur()
				m.updateFilter()
				return m, nil
			}
			
			var cmd tea.Cmd
			m.searchInput, cmd = m.searchInput.Update(msg)
			cmds = append(cmds, cmd)
			m.updateFilter()
			return m, tea.Batch(cmds...)
		}

		switch {
		case key.Matches(msg, m.keys.Search):
			m.searchInput.Focus()
			return m, textinput.Blink

		case key.Matches(msg, m.keys.Toggle):
			return m.toggleCurrentRepo(), nil

		case key.Matches(msg, m.keys.Continue):
			if len(m.getSelectedRepositories()) == 0 {
				return m, nil // Need at least one repo
			}
			return m, func() tea.Msg {
				return NavigateToConfigMsg{
					SelectedRepos: m.getSelectedRepositories(),
				}
			}

		case key.Matches(msg, m.keys.ClearAll):
			m.clearAllSelections()
			return m, nil

		case key.Matches(msg, m.keys.SelectAll):
			m.selectAllFiltered()
			return m, nil

		case key.Matches(msg, m.keys.Preset):
			return m.cyclePreset(), nil
		}
	}

	// Update list
	var cmd tea.Cmd
	m.repoList, cmd = m.repoList.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *SelectionModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(1, 2)

	title := titleStyle.Render("Worktree TUI - Quick Workspace Setup")

	searchSection := lipgloss.NewStyle().
		Padding(1, 2).
		Render(fmt.Sprintf("Search: %s", m.searchInput.View()))

	presetSection := ""
	if len(m.presetButtons) > 0 {
		presetSection = m.renderPresets()
	}

	listSection := lipgloss.NewStyle().
		Padding(0, 2).
		Height(m.height - 12). // Reserve space for other sections
		Render(m.repoList.View())

	statusSection := m.renderStatus()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		searchSection,
		presetSection,
		listSection,
		statusSection,
	)
}

func (m *SelectionModel) renderPresets() string {
	if len(m.presetButtons) == 0 {
		return ""
	}

	presetStyle := lipgloss.NewStyle().
		Padding(1, 2)

	buttons := make([]string, len(m.presetButtons))
	for i, button := range m.presetButtons {
		style := lipgloss.NewStyle().
			Padding(0, 1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240"))

		if button.selected {
			style = style.
				BorderForeground(lipgloss.Color("62")).
				Foreground(lipgloss.Color("62"))
		}

		buttons[i] = style.Render(button.name)
	}

	return presetStyle.Render(
		"Presets: " + strings.Join(buttons, " "),
	)
}

func (m *SelectionModel) renderStatus() string {
	statusStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("240"))

	selectedCount := len(m.getSelectedRepositories())
	status := fmt.Sprintf("Selected: %d repositories", selectedCount)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	help := helpStyle.Render("/ search • space toggle • c continue • p preset • ctrl+a select all • ctrl+d clear • q quit")

	return statusStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			status,
			help,
		),
	)
}

func (m *SelectionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.repoList.SetSize(width-4, height-12)
	m.searchInput.Width = width - 20
}

func (m *SelectionModel) Reset() {
	m.selectedRepos = make(map[string]bool)
	m.selectedPreset = -1
	for i := range m.presetButtons {
		m.presetButtons[i].selected = false
	}
	m.searchInput.SetValue("")
	m.updateRepositoryList()
}

func (m *SelectionModel) updateFilter() {
	query := strings.ToLower(strings.TrimSpace(m.searchInput.Value()))
	
	if query == "" {
		m.filteredRepos = m.repositories
	} else {
		m.filteredRepos = nil
		for _, repo := range m.repositories {
			if m.matchesFilter(repo, query) {
				m.filteredRepos = append(m.filteredRepos, repo)
			}
		}
	}
	
	m.updateRepositoryList()
}

func (m *SelectionModel) matchesFilter(repo repositoryItem, query string) bool {
	searchText := strings.ToLower(repo.FilterValue())
	return strings.Contains(searchText, query)
}

func (m *SelectionModel) updateRepositoryList() {
	items := make([]list.Item, len(m.filteredRepos))
	for i, repo := range m.filteredRepos {
		repo.selected = m.selectedRepos[repo.repo.Name]
		items[i] = repo
	}
	m.repoList.SetItems(items)
}

func (m *SelectionModel) toggleCurrentRepo() *SelectionModel {
	if m.repoList.SelectedItem() == nil {
		return m
	}
	
	item := m.repoList.SelectedItem().(repositoryItem)
	repoName := item.repo.Name
	
	m.selectedRepos[repoName] = !m.selectedRepos[repoName]
	if !m.selectedRepos[repoName] {
		delete(m.selectedRepos, repoName)
	}
	
	m.updateRepositoryList()
	return m
}

func (m *SelectionModel) clearAllSelections() {
	m.selectedRepos = make(map[string]bool)
	m.selectedPreset = -1
	for i := range m.presetButtons {
		m.presetButtons[i].selected = false
	}
	m.updateRepositoryList()
}

func (m *SelectionModel) selectAllFiltered() {
	for _, repo := range m.filteredRepos {
		m.selectedRepos[repo.repo.Name] = true
	}
	m.updateRepositoryList()
}

func (m *SelectionModel) cyclePreset() *SelectionModel {
	if len(m.presetButtons) == 0 {
		return m
	}

	// Clear current preset selection
	if m.selectedPreset >= 0 {
		m.presetButtons[m.selectedPreset].selected = false
	}

	// Move to next preset (or -1 to clear)
	m.selectedPreset++
	if m.selectedPreset >= len(m.presetButtons) {
		m.selectedPreset = -1
	}

	if m.selectedPreset >= 0 {
		// Select new preset
		m.presetButtons[m.selectedPreset].selected = true
		preset := m.presetButtons[m.selectedPreset].preset
		
		// Clear current selections and apply preset
		m.selectedRepos = make(map[string]bool)
		for _, repoName := range preset.Repositories {
			m.selectedRepos[repoName] = true
		}
	} else {
		// Clear all selections
		m.selectedRepos = make(map[string]bool)
	}

	m.updateRepositoryList()
	return m
}

func (m *SelectionModel) getSelectedRepositories() []config.RepositorySelection {
	var selected []config.RepositorySelection
	
	for repoName := range m.selectedRepos {
		if repo, exists := m.config.GetRepositoryByName(repoName); exists {
			selected = append(selected, config.RepositorySelection{
				Repository: *repo,
				Branch:     repo.DefaultBranch,
				Selected:   true,
			})
		}
	}
	
	return selected
}

// Custom list delegate for repository items
func newRepositoryDelegate() list.DefaultDelegate {
	d := list.NewDefaultDelegate()
	
	d.Styles.SelectedTitle = d.Styles.SelectedTitle.
		Foreground(lipgloss.Color("62")).
		BorderLeftForeground(lipgloss.Color("62"))
	
	d.Styles.SelectedDesc = d.Styles.SelectedDesc.
		Foreground(lipgloss.Color("240")).
		BorderLeftForeground(lipgloss.Color("62"))
	
	return d
}