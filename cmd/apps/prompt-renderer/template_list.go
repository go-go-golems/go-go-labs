package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TemplateListModel handles the template selection screen
type TemplateListModel struct {
	templates []TemplateDefinition
	cursor    int
	width     int
	height    int
}

// NewTemplateListModel creates a new template list model
func NewTemplateListModel(templates []TemplateDefinition) *TemplateListModel {
	return &TemplateListModel{
		templates: templates,
		cursor:    0,
	}
}

// Init implements tea.Model
func (m *TemplateListModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *TemplateListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.templates)-1 {
				m.cursor++
			}
		case "enter":
			// Return a message to switch to config view
			return m, func() tea.Msg {
				return SelectTemplateMsg{Template: m.templates[m.cursor]}
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			// Direct selection by number
			num := int(msg.String()[0] - '1')
			if num >= 0 && num < len(m.templates) {
				m.cursor = num
				return m, func() tea.Msg {
					return SelectTemplateMsg{Template: m.templates[m.cursor]}
				}
			}
		case "?":
			return m, func() tea.Msg {
				return ShowHelpMsg{}
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}

	return m, nil
}

// View implements tea.Model
func (m *TemplateListModel) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)

	title := titleStyle.Render("PROMPT BUILDER")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Subtitle
	subtitleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Align(lipgloss.Center).
		Width(m.width)

	subtitle := subtitleStyle.Render("Select a Template")
	b.WriteString(subtitle)
	b.WriteString("\n\n")

	// Template list
	for i, template := range m.templates {
		cursor := "  "
		if i == m.cursor {
			cursor = "► "
		}

		number := fmt.Sprintf("%d.", i+1)
		
		// Template line
		templateStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))
		
		if i == m.cursor {
			templateStyle = templateStyle.
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))
		}

		line := fmt.Sprintf("%s %s %s", cursor, number, template.Label)
		if i == m.cursor {
			line += "                     [SELECTED]"
		}
		
		b.WriteString(templateStyle.Render(line))
		b.WriteString("\n")

		// Description/hint line
		if template.Model != "" {
			hintStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#666666")).
				Italic(true).
				MarginLeft(6)
			
			hint := fmt.Sprintf("Model: %s", template.Model)
			b.WriteString(hintStyle.Render(hint))
			b.WriteString("\n")
		}
		
		b.WriteString("\n")
	}

	// Fill remaining space
	contentHeight := strings.Count(b.String(), "\n")
	for i := contentHeight; i < m.height-3; i++ {
		b.WriteString("\n")
	}

	// Status bar
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1).
		Width(m.width)

	status := "↑↓/j/k: Navigate    Enter: Select    q/Ctrl+C: Quit    ?: Help"
	b.WriteString(statusStyle.Render(status))

	return b.String()
}

// SelectTemplateMsg is sent when a template is selected
type SelectTemplateMsg struct {
	Template TemplateDefinition
}

// ShowHelpMsg is sent when help is requested
type ShowHelpMsg struct{}
