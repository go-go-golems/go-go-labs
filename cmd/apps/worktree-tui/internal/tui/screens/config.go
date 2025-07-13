package screens

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
)

var _ tea.Model = (*ConfigModel)(nil)

type ConfigModel struct {
	config        *config.Config
	selectedRepos []config.RepositorySelection
	width         int
	height        int

	// UI components
	nameInput textinput.Model
	pathInput textinput.Model

	// State
	focused int // 0: name, 1: path

	// Key bindings
	keys configKeyMap
}

type configKeyMap struct {
	Continue key.Binding
	Back     key.Binding
	Tab      key.Binding
}

var configKeys = configKeyMap{
	Continue: key.NewBinding(
		key.WithKeys("enter", "c"),
		key.WithHelp("enter/c", "create workspace"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab", "shift+tab"),
		key.WithHelp("tab", "next field"),
	),
}

func NewConfigModel(cfg *config.Config, selectedRepos []config.RepositorySelection) *ConfigModel {
	nameInput := textinput.New()
	nameInput.Placeholder = "workspace-name"
	nameInput.Focus()
	nameInput.CharLimit = 50

	pathInput := textinput.New()
	pathInput.Placeholder = cfg.Workspaces.DefaultBasePath
	pathInput.CharLimit = 200

	// Generate default workspace name from selected repos
	defaultName := generateWorkspaceName(selectedRepos)
	nameInput.SetValue(defaultName)

	// Set default path
	defaultPath := filepath.Join(cfg.Workspaces.DefaultBasePath, defaultName)
	pathInput.SetValue(defaultPath)

	return &ConfigModel{
		config:        cfg,
		selectedRepos: selectedRepos,
		nameInput:     nameInput,
		pathInput:     pathInput,
		focused:       0,
		keys:          configKeys,
	}
}

func (m *ConfigModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Continue):
			if m.isValid() {
				req := &config.WorkspaceRequest{
					Name:         m.nameInput.Value(),
					Path:         m.pathInput.Value(),
					Repositories: m.getRepositories(),
				}
				return m, func() tea.Msg {
					return NavigateToProgressMsg{WorkspaceRequest: req}
				}
			}
			return m, nil

		case key.Matches(msg, m.keys.Tab):
			if msg.String() == "shift+tab" {
				m.focused--
				if m.focused < 0 {
					m.focused = 1
				}
			} else {
				m.focused++
				if m.focused > 1 {
					m.focused = 0
				}
			}
			m.updateFocus()
			return m, nil
		}
	}

	// Update focused input
	var cmd tea.Cmd
	if m.focused == 0 {
		m.nameInput, cmd = m.nameInput.Update(msg)
		cmds = append(cmds, cmd)

		// Update path when name changes
		if m.nameInput.Value() != "" {
			basePath := m.config.Workspaces.DefaultBasePath
			newPath := filepath.Join(basePath, m.nameInput.Value())
			m.pathInput.SetValue(newPath)
		}
	} else {
		m.pathInput, cmd = m.pathInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *ConfigModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(1, 2)

	title := titleStyle.Render("Workspace Configuration")

	// Selected repositories section
	selectedSection := m.renderSelectedRepos()

	// Configuration form
	configSection := m.renderConfigForm()

	// Status and help
	statusSection := m.renderStatus()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		selectedSection,
		configSection,
		statusSection,
	)
}

func (m *ConfigModel) renderSelectedRepos() string {
	sectionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		Render("Selected Repositories:")

	repoList := make([]string, len(m.selectedRepos))
	for i, repo := range m.selectedRepos {
		repoList[i] = fmt.Sprintf("  • %s (%s) - %s",
			repo.Repository.Name,
			repo.Branch,
			repo.Repository.Description)
	}

	content := header + "\n" + strings.Join(repoList, "\n")

	return sectionStyle.Render(content)
}

func (m *ConfigModel) renderConfigForm() string {
	sectionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	// Name field
	nameLabel := "Workspace Name:"
	if m.focused == 0 {
		nameLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true).
			Render("→ " + nameLabel)
	}

	// Path field
	pathLabel := "Workspace Path:"
	if m.focused == 1 {
		pathLabel = lipgloss.NewStyle().
			Foreground(lipgloss.Color("62")).
			Bold(true).
			Render("→ " + pathLabel)
	}

	form := lipgloss.JoinVertical(
		lipgloss.Left,
		nameLabel,
		"  "+m.nameInput.View(),
		"",
		pathLabel,
		"  "+m.pathInput.View(),
	)

	return sectionStyle.Render(form)
}

func (m *ConfigModel) renderStatus() string {
	statusStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("240"))

	var status strings.Builder

	// Validation messages
	if !m.isValid() {
		status.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Render("⚠ Please provide a workspace name"))
		status.WriteString("\n")
	} else {
		status.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("34")).
			Render("✓ Ready to create workspace"))
		status.WriteString("\n")
	}

	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	help := helpStyle.Render("tab next field • enter create • esc back • q quit")
	status.WriteString(help)

	return statusStyle.Render(status.String())
}

func (m *ConfigModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.nameInput.Width = width - 10
	m.pathInput.Width = width - 10
}

func (m *ConfigModel) SetSelectedRepos(repos []config.RepositorySelection) {
	m.selectedRepos = repos

	// Update default name and path
	defaultName := generateWorkspaceName(repos)
	m.nameInput.SetValue(defaultName)

	defaultPath := filepath.Join(m.config.Workspaces.DefaultBasePath, defaultName)
	m.pathInput.SetValue(defaultPath)
}

func (m *ConfigModel) updateFocus() {
	if m.focused == 0 {
		m.nameInput.Focus()
		m.pathInput.Blur()
	} else {
		m.nameInput.Blur()
		m.pathInput.Focus()
	}
}

func (m *ConfigModel) isValid() bool {
	return strings.TrimSpace(m.nameInput.Value()) != ""
}

func (m *ConfigModel) getRepositories() []config.Repository {
	repos := make([]config.Repository, len(m.selectedRepos))
	for i, selection := range m.selectedRepos {
		repos[i] = selection.Repository
	}
	return repos
}

func generateWorkspaceName(repos []config.RepositorySelection) string {
	if len(repos) == 0 {
		return "workspace"
	}

	if len(repos) == 1 {
		return repos[0].Repository.Name
	}

	// For multiple repos, try to find common tags or create a descriptive name
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = repo.Repository.Name
	}

	// If all names are short, join them
	if len(names) <= 3 {
		joined := strings.Join(names, "-")
		if len(joined) <= 30 {
			return joined
		}
	}

	// Otherwise, use a generic name with count
	return fmt.Sprintf("workspace-%d-repos", len(repos))
}
