package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport       viewport.Model
	textarea       textarea.Model
	renderer       *glamour.TermRenderer
	width          int
	height         int
	err            error
	renderMarkdown bool
	showHelp       bool
}

func initialModel() model {
	// Create and configure textarea
	ta := textarea.New()
	ta.Placeholder = "Enter markdown here..."
	ta.SetWidth(80)
	ta.SetHeight(5)
	ta.Focus()

	// Initial viewport with empty content
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Create glamour renderer
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		log.Fatal(err)
	}

	m := model{
		textarea:       ta,
		viewport:       vp,
		renderer:       renderer,
		width:          80,
		height:         25,
		renderMarkdown: false, // Start with markdown rendering disabled
		showHelp:       true,  // Start with help visible
	}

	// Initial render of empty content (as plain text)
	m.renderContent()

	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

// Save rendered content to file
func (m model) saveRenderedContent() {
	if content := m.viewport.View(); content != "" {
		err := os.WriteFile("/tmp/rendered.md", []byte(content), 0644)
		if err != nil {
			m.err = err
		}
	}
}

// Render content based on current mode
func (m *model) renderContent() {
	if strings.TrimSpace(m.textarea.Value()) == "" {
		m.viewport.SetContent("")
		return
	}

	var content string
	if m.renderMarkdown {
		// Render as markdown
		renderedContent, err := m.renderer.Render(m.textarea.Value())
		if err != nil {
			m.err = err
			content = fmt.Sprintf("Render Error:\n%s\n\n%s", err.Error(), m.textarea.Value())
		} else {
			m.err = nil
			content = renderedContent
		}
	} else {
		// Show as plain text
		m.err = nil
		content = m.textarea.Value()
	}

	m.viewport.SetContent(content)
	m.saveRenderedContent()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "ctrl+m":
			// Toggle markdown rendering
			m.renderMarkdown = !m.renderMarkdown
			m.renderContent()
			return m, nil
		case "ctrl+h":
			// Toggle help
			m.showHelp = !m.showHelp
			return m, nil
		default:
			// Handle textarea input
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)

			// Re-render content
			m.renderContent()
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate proportional heights
		helpHeight := 0
		if m.showHelp {
			helpHeight = 3 // 3 lines for help
		}
		statusHeight := 1                                                           // 1 line for status
		textAreaHeight := 6                                                         // 5 lines + 1 for border
		viewportHeight := m.height - textAreaHeight - helpHeight - statusHeight - 2 // 2 for margins and dividers

		// Resize viewport
		m.viewport.Width = m.width
		m.viewport.Height = viewportHeight

		// Create new renderer with updated width for word wrap
		renderer, err := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
			glamour.WithWordWrap(m.width),
		)
		if err != nil {
			m.err = err
		} else {
			m.renderer = renderer
			m.renderContent()
		}

		// Resize textarea
		m.textarea.SetWidth(m.width)
		m.textarea.SetHeight(5)
	}

	// Also update viewport for scrolling
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	// Define styles
	viewportStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	textareaStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	// Create a divider
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("─" + strings.Repeat("─", m.width-2) + "─")

	// Create status display
	mode := "Plain Text"
	if m.renderMarkdown {
		mode = "Markdown"
	}
	contentLength := len(m.viewport.View())
	linesCount := len(strings.Split(m.viewport.View(), "\n"))
	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Render(fmt.Sprintf("Mode: %s | Content size: %d chars, %d lines", mode, contentLength, linesCount))

	// Create help display
	help := ""
	if m.showHelp {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
		help = helpStyle.Render("ctrl+h: toggle help | ctrl+m: toggle markdown mode | esc/ctrl+c: quit")
		help = lipgloss.JoinVertical(lipgloss.Left, help, divider)
	}

	// Create error display if needed
	errorDisplay := ""
	if m.err != nil {
		errorDisplay = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Combine components
	return lipgloss.JoinVertical(lipgloss.Left,
		help,
		viewportStyle.Render(m.viewport.View()),
		status,
		divider,
		textareaStyle.Render(m.textarea.View()),
		errorDisplay,
	)
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}
