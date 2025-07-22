package widgets

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/sparkline"
)

// StreamsTableWidget displays the main streams table with sparklines
type StreamsTableWidget struct {
	width       int
	height      int
	focused     bool
	streams     []StreamData
	sparklines  map[string]*sparkline.Sparkline
	selectedIdx int
	styles      StreamsTableStyles
}

type StreamsTableStyles struct {
	Container    lipgloss.Style
	Table        lipgloss.Style
	HeaderRow    lipgloss.Style
	Row          lipgloss.Style
	SelectedRow  lipgloss.Style
	SparklineRow lipgloss.Style
}

// NewStreamsTableWidget creates a new streams table widget
func NewStreamsTableWidget(styles StreamsTableStyles) StreamsTableWidget {
	return StreamsTableWidget{
		styles:     styles,
		sparklines: make(map[string]*sparkline.Sparkline),
	}
}

// Init implements tea.Model
func (w StreamsTableWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w StreamsTableWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		
	case DataUpdateMsg:
		w.streams = msg.StreamsData
		w.updateSparklines()
		
		// Ensure selected index is valid
		if w.selectedIdx >= len(w.streams) {
			w.selectedIdx = len(w.streams) - 1
		}
		if w.selectedIdx < 0 && len(w.streams) > 0 {
			w.selectedIdx = 0
		}
		
	case tea.KeyMsg:
		if w.focused {
			switch msg.String() {
			case "up", "k":
				if w.selectedIdx > 0 {
					w.selectedIdx--
				}
			case "down", "j":
				if w.selectedIdx < len(w.streams)-1 {
					w.selectedIdx++
				}
			}
		}
	}
	
	return w, nil
}

// View implements tea.Model
func (w StreamsTableWidget) View() string {
	if w.width == 0 || len(w.streams) == 0 {
		return w.styles.Container.Render("No streams found")
	}
	
	var rows []string
	
	// Table header with border
	topBorder := "┌─────────┬──────────────┬──────────┬───────────────┬───────────┬──────────────┐"
	headerRow := "│ Stream  │   Entries    │   Size   │   Groups      │  Last ID  │  Memory RSS  │"
	divider := "├─────────┼──────────────┼──────────┼───────────────┼───────────┼──────────────┤"
	
	rows = append(rows, topBorder)
	rows = append(rows, w.styles.HeaderRow.Render(headerRow))
	rows = append(rows, divider)
	
	// Stream rows
	for i, stream := range w.streams {
		// Main data row
		mainRow := fmt.Sprintf("│ %-7s │ %,12d │ %-8s │ %13d │ %-9s │ %-12s │",
			truncateString(stream.Name, 7),
			stream.Length,
			formatBytes(stream.MemoryUsage),
			stream.Groups,
			truncateString(stream.LastID, 9),
			formatBytes(stream.MemoryUsage),
		)
		
		// Sparkline row
		sparklineStr := ""
		if sl, exists := w.sparklines[stream.Name]; exists {
			sparklineStr = sl.Render()
		}
		sparklineRow := fmt.Sprintf("│         │ msg/s: %-25s │          │               │           │              │",
			sparklineStr)
		
		// Apply styling based on selection
		if i == w.selectedIdx && w.focused {
			rows = append(rows, w.styles.SelectedRow.Render(mainRow))
			rows = append(rows, w.styles.SelectedRow.Render(sparklineRow))
		} else {
			rows = append(rows, w.styles.Row.Render(mainRow))
			rows = append(rows, w.styles.SparklineRow.Render(sparklineRow))
		}
		
		// Add divider between streams (except last)
		if i < len(w.streams)-1 {
			rows = append(rows, divider)
		}
	}
	
	// Bottom border
	bottomBorder := "└─────────┴──────────────┴──────────┴───────────────┴───────────┴──────────────┘"
	rows = append(rows, bottomBorder)
	
	content := strings.Join(rows, "\n")
	return w.styles.Container.Width(w.width).Render(content)
}

// SetSize implements Widget interface
func (w *StreamsTableWidget) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// SetFocused implements Widget interface
func (w *StreamsTableWidget) SetFocused(focused bool) {
	w.focused = focused
}

// MinHeight implements Widget interface
func (w StreamsTableWidget) MinHeight() int {
	// At least header + 1 stream (2 rows) + borders
	return 5
}

// MaxHeight implements Widget interface
func (w StreamsTableWidget) MaxHeight() int {
	// Header + borders + 2 rows per stream
	return 3 + (len(w.streams) * 2)
}

// updateSparklines creates or updates sparklines for each stream
func (w *StreamsTableWidget) updateSparklines() {
	for _, stream := range w.streams {
		if _, exists := w.sparklines[stream.Name]; !exists {
			config := sparkline.Config{
				Width:     25,
				Height:    1,
				MaxPoints: 25,
				Style:     sparkline.StyleBars,
			}
			w.sparklines[stream.Name] = sparkline.New(config)
		}
		
		// Update sparkline data
		w.sparklines[stream.Name].SetData(stream.MessageRates)
	}
}

// GetSelectedStream returns the currently selected stream
func (w StreamsTableWidget) GetSelectedStream() *StreamData {
	if w.selectedIdx >= 0 && w.selectedIdx < len(w.streams) {
		return &w.streams[w.selectedIdx]
	}
	return nil
}

// Helper functions

func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
