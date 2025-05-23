package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/go-go-golems/go-go-labs/cmd/apps/worktree-tui/internal/config"
)

var _ tea.Model = (*CompletionModel)(nil)

type CompletionModel struct {
	workspaceReq *config.WorkspaceRequest
	success      bool
	err          error
	width        int
	height       int
	
	// Key bindings
	keys completionKeyMap
}

type completionKeyMap struct {
	NewWorkspace key.Binding
	Quit         key.Binding
}

var completionKeys = completionKeyMap{
	NewWorkspace: key.NewBinding(
		key.WithKeys("n", "enter"),
		key.WithHelp("n/enter", "new workspace"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func NewCompletionModel(req *config.WorkspaceRequest, success bool, err error) *CompletionModel {
	return &CompletionModel{
		workspaceReq: req,
		success:      success,
		err:          err,
		keys:         completionKeys,
	}
}

func (m *CompletionModel) Init() tea.Cmd {
	return nil
}

func (m *CompletionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.NewWorkspace):
			// Navigate back to selection screen to create another workspace
			return m, func() tea.Msg {
				// This will be handled by the app to reset to selection screen
				return QuitMsg{} // For now, just quit
			}
		case key.Matches(msg, m.keys.Quit):
			return m, func() tea.Msg {
				return QuitMsg{}
			}
		}
	}
	
	return m, nil
}

func (m *CompletionModel) View() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("62")).
		Bold(true).
		Padding(1, 2)

	var title string
	if m.success {
		title = titleStyle.Render("✓ Workspace Created Successfully!")
	} else {
		title = titleStyle.Render("✗ Workspace Creation Failed")
	}

	// Result section
	resultSection := m.renderResult()

	// Next steps section
	nextStepsSection := m.renderNextSteps()

	// Help section
	helpSection := m.renderHelp()

	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		resultSection,
		nextStepsSection,
		helpSection,
	)
}

func (m *CompletionModel) renderResult() string {
	sectionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	if m.success {
		successStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("34"))

		content := []string{
			successStyle.Render("Workspace Details:"),
			fmt.Sprintf("  Name: %s", m.workspaceReq.Name),
			fmt.Sprintf("  Path: %s", m.workspaceReq.Path),
			"",
			successStyle.Render("Repositories:"),
		}

		for _, repo := range m.workspaceReq.Repositories {
			content = append(content, fmt.Sprintf("  • %s - %s", repo.Name, repo.Description))
		}

		return sectionStyle.Render(strings.Join(content, "\n"))
	} else {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196"))

		content := []string{
			errorStyle.Render("Error Details:"),
		}

		if m.err != nil {
			content = append(content, fmt.Sprintf("  %s", m.err.Error()))
		} else {
			content = append(content, "  Unknown error occurred")
		}

		if m.workspaceReq != nil {
			content = append(content, "", "Attempted to create:")
			content = append(content, fmt.Sprintf("  Name: %s", m.workspaceReq.Name))
			content = append(content, fmt.Sprintf("  Path: %s", m.workspaceReq.Path))
		}

		return sectionStyle.Render(strings.Join(content, "\n"))
	}
}

func (m *CompletionModel) renderNextSteps() string {
	if !m.success {
		return ""
	}

	sectionStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		MarginBottom(1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62"))

	commandStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Background(lipgloss.Color("236")).
		Padding(0, 1)

	content := []string{
		headerStyle.Render("Next Steps:"),
		"",
		"1. Navigate to your workspace:",
		"   " + commandStyle.Render(fmt.Sprintf("cd %s", m.workspaceReq.Path)),
		"",
		"2. Verify the go.work file:",
		"   " + commandStyle.Render("cat go.work"),
		"",
		"3. Start developing:",
		"   " + commandStyle.Render("go build ./..."),
		"   " + commandStyle.Render("go test ./..."),
	}

	return sectionStyle.Render(strings.Join(content, "\n"))
}

func (m *CompletionModel) renderHelp() string {
	statusStyle := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder(), true, false, false, false).
		BorderForeground(lipgloss.Color("240"))

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	var help string
	if m.success {
		help = helpStyle.Render("n/enter create another workspace • q quit")
	} else {
		help = helpStyle.Render("n/enter try again • q quit")
	}

	return statusStyle.Render(help)
}

func (m *CompletionModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *CompletionModel) SetResult(req *config.WorkspaceRequest, success bool, err error) {
	m.workspaceReq = req
	m.success = success
	m.err = err
}