package widgets

import (
	"fmt"
	"strconv"
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
	
	// Calculate responsive column widths
	cols := w.calculateColumnWidths()
	
	var rows []string
	
	// Create borders and headers
	topBorder := w.createBorder(cols, "┌", "┬", "┐")
	headerRow := w.createHeaderRow(cols)
	divider := w.createBorder(cols, "├", "┼", "┤")
	bottomBorder := w.createBorder(cols, "└", "┴", "┘")
	
	rows = append(rows, topBorder)
	rows = append(rows, w.styles.HeaderRow.Render(headerRow))
	rows = append(rows, divider)
	
	// Stream rows
	for i, stream := range w.streams {
		// Main data row
		mainRow := w.createDataRow(stream, cols)
		
		// Sparkline row
		sparklineRow := w.createSparklineRow(stream, cols)
		
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
	// Calculate optimal sparkline width based on entries column width
	cols := w.calculateColumnWidths()
	sparklineWidth := cols.entries - len("msg/s: ")
	if sparklineWidth < 5 {
		sparklineWidth = 5 // Minimum sparkline width
	}
	if sparklineWidth > 40 {
		sparklineWidth = 40 // Maximum sparkline width for performance
	}
	
	for _, stream := range w.streams {
		if _, exists := w.sparklines[stream.Name]; !exists {
			config := sparkline.Config{
				Width:     sparklineWidth,
				Height:    1,
				MaxPoints: sparklineWidth,
				Style:     sparkline.StyleBars,
			}
			w.sparklines[stream.Name] = sparkline.New(config)
		} else {
			// Update existing sparkline configuration if width changed
			existing := w.sparklines[stream.Name]
			currentConfig := existing.GetConfig()
			if currentConfig.Width != sparklineWidth {
				config := sparkline.Config{
					Width:     sparklineWidth,
					Height:    1,
					MaxPoints: sparklineWidth,
					Style:     sparkline.StyleBars,
				}
				existing.UpdateConfig(config)
			}
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

// Column widths structure
type columnWidths struct {
	stream   int
	entries  int
	size     int
	groups   int
	lastID   int
	memory   int
}

// calculateColumnWidths determines responsive column widths based on terminal size
func (w StreamsTableWidget) calculateColumnWidths() columnWidths {
	// Minimum column widths
	minWidths := columnWidths{
		stream:  7,
		entries: 12,
		size:    8,
		groups:  8,
		lastID:  9,
		memory:  12,
	}
	
	// Account for borders and padding (6 columns = 7 borders + 12 spaces)
	overhead := 7 + 12
	availableWidth := w.width - overhead
	
	if availableWidth <= 0 {
		return minWidths
	}
	
	// Calculate total minimum width needed
	totalMinWidth := minWidths.stream + minWidths.entries + minWidths.size + 
		minWidths.groups + minWidths.lastID + minWidths.memory
	
	if availableWidth <= totalMinWidth {
		return minWidths
	}
	
	// Distribute extra space proportionally
	extraSpace := availableWidth - totalMinWidth
	result := minWidths
	
	// Give extra space to entries and memory columns first
	entriesExtra := extraSpace / 3
	memoryExtra := extraSpace / 3
	streamExtra := extraSpace - entriesExtra - memoryExtra
	
	result.entries += entriesExtra
	result.memory += memoryExtra
	result.stream += streamExtra
	
	return result
}

// createBorder creates a horizontal border with given characters
func (w StreamsTableWidget) createBorder(cols columnWidths, left, middle, right string) string {
	parts := []string{
		left + strings.Repeat("─", cols.stream),
		middle + strings.Repeat("─", cols.entries),
		middle + strings.Repeat("─", cols.size),
		middle + strings.Repeat("─", cols.groups),
		middle + strings.Repeat("─", cols.lastID),
		middle + strings.Repeat("─", cols.memory) + right,
	}
	return strings.Join(parts, "")
}

// createHeaderRow creates the table header
func (w StreamsTableWidget) createHeaderRow(cols columnWidths) string {
	return fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │",
		cols.stream, "Stream",
		cols.entries, "Entries",
		cols.size, "Size",
		cols.groups, "Groups",
		cols.lastID, "Last ID",
		cols.memory, "Memory RSS",
	)
}

// createDataRow creates a data row for a stream
func (w StreamsTableWidget) createDataRow(stream StreamData, cols columnWidths) string {
	return fmt.Sprintf("│ %-*s │ %*s │ %-*s │ %*s │ %-*s │ %-*s │",
		cols.stream, truncateString(stream.Name, cols.stream),
		cols.entries, formatNumberWithCommas(stream.Length),
		cols.size, formatBytes(stream.MemoryUsage),
		cols.groups, formatNumberWithCommas(stream.Groups),
		cols.lastID, truncateString(stream.LastID, cols.lastID),
		cols.memory, formatBytes(stream.MemoryUsage),
	)
}

// createSparklineRow creates the sparkline row for a stream
func (w StreamsTableWidget) createSparklineRow(stream StreamData, cols columnWidths) string {
	sparklineStr := ""
	if sl, exists := w.sparklines[stream.Name]; exists {
		sparklineStr = sl.Render()
	}
	
	// Calculate available space for sparkline
	prefix := "msg/s: "
	maxSparklineWidth := cols.entries - len(prefix)
	
	// Ensure we have at least some space for the sparkline
	if maxSparklineWidth <= 0 {
		maxSparklineWidth = 0
		sparklineStr = ""
	} else if len(sparklineStr) > maxSparklineWidth {
		// Truncate sparkline to fit available space
		sparklineStr = sparklineStr[:maxSparklineWidth]
	}
	
	sparklineContent := fmt.Sprintf("%s%s", prefix, sparklineStr)
	
	return fmt.Sprintf("│ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │ %-*s │",
		cols.stream, "",
		cols.entries, sparklineContent,
		cols.size, "",
		cols.groups, "",
		cols.lastID, "",
		cols.memory, "",
	)
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

// formatNumberWithCommas formats int64 with commas for thousands separators
func formatNumberWithCommas(n int64) string {
	str := strconv.FormatInt(n, 10)
	
	// Add commas for thousands separators
	if len(str) <= 3 {
		return str
	}
	
	var result strings.Builder
	for i, char := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteRune(char)
	}
	
	return result.String()
}
