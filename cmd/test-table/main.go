package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/tui/widgets"
)

// Simple model that just wraps the streams table widget
type testModel struct {
	table  widgets.StreamsTableWidget
	width  int
	height int
}

func newTestModel() testModel {
	// Create test styles
	styles := widgets.StreamsTableStyles{
		Container:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1),
		Table:        lipgloss.NewStyle(),
		HeaderRow:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		Row:          lipgloss.NewStyle(),
		SelectedRow:  lipgloss.NewStyle().Background(lipgloss.Color("#555555")),
		SparklineRow: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	}

	// Create the widget
	table := widgets.NewStreamsTableWidget(styles)

	return testModel{
		table: table,
	}
}

func (m testModel) Init() tea.Cmd {
	// Send test data immediately
	return func() tea.Msg {
		return widgets.DataUpdateMsg{
			StreamsData: []widgets.StreamData{
				{
					Name:         "test_stream_1",
					Length:       1500,
					MemoryUsage:  1024 * 1024, // 1MB
					Groups:       3,
					LastID:       "1234567890123-0",
					MessageRates: []float64{1.0, 2.0, 3.0, 4.0, 5.0, 4.0, 3.0, 2.0, 1.0, 2.0},
				},
				{
					Name:         "test_stream_2",
					Length:       8750,
					MemoryUsage:  2 * 1024 * 1024, // 2MB
					Groups:       5,
					LastID:       "9876543210987-1",
					MessageRates: []float64{5.0, 4.0, 6.0, 3.0, 7.0, 2.0, 8.0, 1.0, 9.0, 10.0},
				},
			},
		}
	}
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.SetSize(msg.Width, msg.Height-4) // Leave some space for borders

	case widgets.DataUpdateMsg:
		// Update the table with data
		updatedTable, cmd := m.table.Update(msg)
		m.table = updatedTable.(widgets.StreamsTableWidget)
		return m, cmd

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
		
		// Pass other keys to the table
		updatedTable, cmd := m.table.Update(msg)
		m.table = updatedTable.(widgets.StreamsTableWidget)
		return m, cmd
	}

	return m, nil
}

func (m testModel) View() string {
	if m.width == 0 {
		return "Initializing..."
	}

	content := fmt.Sprintf("Simple Table Test (Press 'q' to quit)\nTerminal size: %dx%d\n\n", m.width, m.height)
	content += m.table.View()
	
	return content
}

func main() {
	p := tea.NewProgram(newTestModel(), tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
