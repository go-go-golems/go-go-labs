package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/pkg/sparkline"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// StreamsView handles the streams overview table with sparklines
type StreamsView struct {
	styles      styles.Styles
	streams     []models.StreamData
	sparklines  map[string]*sparkline.Sparkline // Map of stream name to sparkline
	selectedIdx int
	width       int
	height      int
}

// StreamsDataMsg contains stream data updates
type StreamsDataMsg struct {
	Streams []models.StreamData
}

// NewStreamsView creates a new streams view
func NewStreamsView(styles styles.Styles) *StreamsView {
	return &StreamsView{
		styles:     styles,
		sparklines: make(map[string]*sparkline.Sparkline),
	}
}

// Init implements tea.Model
func (v *StreamsView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (v *StreamsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case StreamsDataMsg:
		v.streams = msg.Streams
		// Update sparklines for each stream
		for _, stream := range msg.Streams {
			if _, exists := v.sparklines[stream.Name]; !exists {
				// Create new sparkline for this stream
				config := sparkline.Config{
					Width:        15,
					Height:       1,
					MaxPoints:    20,
					Style:        sparkline.StyleBars,
					DefaultStyle: v.styles.Sparkline,
				}
				v.sparklines[stream.Name] = sparkline.New(config)
			}
			// Set the data for this stream's sparkline
			v.sparklines[stream.Name].SetData(stream.MessageRates)
		}
		// Ensure selected index is within bounds
		if v.selectedIdx >= len(v.streams) {
			v.selectedIdx = len(v.streams) - 1
		}
		if v.selectedIdx < 0 {
			v.selectedIdx = 0
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if v.selectedIdx > 0 {
				v.selectedIdx--
			}
		case "down", "j":
			if v.selectedIdx < len(v.streams)-1 {
				v.selectedIdx++
			}
		}
	}

	return v, nil
}

// View implements tea.Model
func (v *StreamsView) View() string {
	if len(v.streams) == 0 {
		return v.styles.StreamTable.Render("No streams found")
	}

	var rows []string
	header := fmt.Sprintf("%-15s %-10s %-10s %-8s %-15s %s",
		"Stream", "Entries", "Memory", "Groups", "Last ID", "Msg/s")
	rows = append(rows, v.styles.Selected.Render(header))

	for i, stream := range v.streams {
		var sparklineStr string
		if sl, exists := v.sparklines[stream.Name]; exists {
			sparklineStr = sl.Render()
		}
		row := fmt.Sprintf("%-15s %-10d %-10s %-8d %-15s %s",
			truncateString(stream.Name, 15),
			stream.Length,
			formatBytes(stream.MemoryUsage),
			stream.Groups,
			truncateString(stream.LastID, 15),
			sparklineStr)

		if i == v.selectedIdx {
			rows = append(rows, v.styles.Selected.Render(row))
		} else {
			rows = append(rows, v.styles.Unselected.Render(row))
		}
	}

	content := strings.Join(rows, "\n")
	return v.styles.StreamTable.Render(content)
}

// GetSelectedIndex returns the currently selected stream index
func (v *StreamsView) GetSelectedIndex() int {
	return v.selectedIdx
}

// SetSelectedIndex sets the selected stream index
func (v *StreamsView) SetSelectedIndex(idx int) {
	if idx >= 0 && idx < len(v.streams) {
		v.selectedIdx = idx
	}
}

// GetSelectedStream returns the currently selected stream data
func (v *StreamsView) GetSelectedStream() *models.StreamData {
	if v.selectedIdx >= 0 && v.selectedIdx < len(v.streams) {
		return &v.streams[v.selectedIdx]
	}
	return nil
}

// truncateString truncates a string to a given length with ellipsis
func truncateString(s string, length int) string {
	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}

// formatBytes formats bytes into human readable format
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
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
