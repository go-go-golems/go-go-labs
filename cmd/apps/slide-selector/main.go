package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/slide-selector/markdownview"
)

type model struct {
	mdView *markdownview.MarkdownView
	width  int
	height int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "/":
			if m.mdView.ViewMode == markdownview.Scrollable {
				m.mdView.ViewMode = markdownview.Static
			} else {
				m.mdView.ViewMode = markdownview.Scrollable
			}
			m.mdView.NeedsRender = true
			return m, nil
		}
	}

	var updatedView tea.Model
	updatedView, cmd = m.mdView.Update(msg)
	var ok bool
	m.mdView, ok = updatedView.(*markdownview.MarkdownView)
	if !ok {
		err := fmt.Errorf("failed to cast updatedView to *markdownview.MarkdownView")
		fmt.Printf("Error: %v\n", err)
		panic(err)
	}
	return m, cmd
}

func (m model) View() string {
	// Create a style for the border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	// Create a style for the size display
	sizeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Right)

	// Wrap the Markdown view in a border
	contentView := borderStyle.Render(m.mdView.View())

	// Create the size display
	sizeDisplay := sizeStyle.Render(fmt.Sprintf("Width: %d, Height: %d", m.width, m.height))

	// Combine the content view and size display
	return lipgloss.JoinVertical(lipgloss.Left, contentView, sizeDisplay)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please provide a Markdown file as an argument.")
		os.Exit(1)
	}

	filePath := os.Args[1]
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	mdView, err := markdownview.NewMarkdownView(string(content), "dark", markdownview.Scrollable)
	if err != nil {
		fmt.Printf("Error creating MarkdownView: %v\n", err)
		os.Exit(1)
	}

	m := model{mdView: mdView}
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
