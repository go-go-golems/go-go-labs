package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport viewport.Model
	textarea textarea.Model
	renderer *glamour.TermRenderer
	width    int
	height   int
	err      error
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
		textarea: ta,
		viewport: vp,
		renderer: renderer,
		width:    80,
		height:   25,
	}

	// Initial render of empty content
	renderedContent, err := m.renderer.Render(m.textarea.Value())
	if err != nil {
		m.err = err
		m.viewport.SetContent("Error rendering markdown.")
	} else {
		m.viewport.SetContent(renderedContent)
	}

	return m
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
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
		default:
			// Handle textarea input
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)

			// Re-render on textarea change
			if strings.TrimSpace(m.textarea.Value()) == "" {
				m.viewport.SetContent("")
			} else {
				renderedContent, err := m.renderer.Render(m.textarea.Value())
				if err != nil {
					m.err = err
					m.viewport.SetContent(fmt.Sprintf("Render Error:\n%s\n\n%s", err.Error(), m.textarea.Value()))
				} else {
					m.err = nil
					m.viewport.SetContent(renderedContent)
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Calculate proportional heights
		textAreaHeight := 6                             // 5 lines + 1 for border
		viewportHeight := m.height - textAreaHeight - 2 // 2 for margins and error display

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

			// Re-render content with new width
			if strings.TrimSpace(m.textarea.Value()) != "" {
				renderedContent, err := m.renderer.Render(m.textarea.Value())
				if err != nil {
					m.err = err
					m.viewport.SetContent(fmt.Sprintf("Resize Render Error:\n%s\n\n%s", err.Error(), m.textarea.Value()))
				} else {
					m.err = nil
					m.viewport.SetContent(renderedContent)
				}
			}
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

	// Create a divider between viewport and textarea
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("─" + strings.Repeat("─", m.width-2) + "─")

	// Create error display if needed
	errorDisplay := ""
	if m.err != nil {
		errorDisplay = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Combine components
	return lipgloss.JoinVertical(lipgloss.Left,
		viewportStyle.Render(m.viewport.View()),
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
