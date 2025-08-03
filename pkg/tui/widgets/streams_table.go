package widgets

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/bobatea/pkg/sparkline"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func init() {
	// Create log file for debugging
	logFile, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logger = zerolog.New(logFile).With().Timestamp().Caller().Str("component", "streams_table").Logger()
	} else {
		logger = zerolog.Nop()
	}
}

// StreamsTableWidget displays the main streams table with sparklines
type StreamsTableWidget struct {
	width      int
	height     int
	focused    bool
	streams    []StreamData
	sparklines map[string]*sparkline.Sparkline
	table      table.Model
	styles     StreamsTableStyles
}

type StreamsTableStyles struct {
	Container lipgloss.Style
	Table     lipgloss.Style
	// The table component handles its own styling, but we keep these for compatibility
	HeaderRow    lipgloss.Style
	Row          lipgloss.Style
	SelectedRow  lipgloss.Style
	SparklineRow lipgloss.Style
}

// NewStreamsTableWidget creates a new streams table widget
func NewStreamsTableWidget(styles StreamsTableStyles) StreamsTableWidget {
	// Create columns for the table
	columns := []table.Column{
		{Title: "Stream", Width: 15},
		{Title: "Entries", Width: 12},
		{Title: "Trend", Width: 20},
		{Title: "Size", Width: 10},
		{Title: "Groups", Width: 8},
		{Title: "Last ID", Width: 15},
		{Title: "Memory RSS", Width: 12},
	}

	// Configure table styles
	tableStyles := table.DefaultStyles()
	tableStyles.Header = styles.HeaderRow
	tableStyles.Selected = styles.SelectedRow
	tableStyles.Cell = styles.Row

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(false),
		table.WithHeight(10),
		table.WithStyles(tableStyles),
	)

	return StreamsTableWidget{
		styles:     styles,
		sparklines: make(map[string]*sparkline.Sparkline),
		table:      t,
	}
}

// Init implements tea.Model
func (w StreamsTableWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w StreamsTableWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Don't handle WindowSizeMsg directly - let the root model handle size allocation
		// The root model will call SetSize() with the proper allocated dimensions
		logger.Debug().Int("width", msg.Width).Int("height", msg.Height).Msg("WindowSizeMsg received - ignoring, root model handles sizing")

	case DataUpdateMsg:
		logger.Info().Int("stream_count", len(msg.StreamsData)).Msg("DataUpdateMsg received")
		for i, stream := range msg.StreamsData {
			logger.Debug().Int("index", i).Str("name", stream.Name).Int64("length", stream.Length).Msg("Stream data")
		}
		w.streams = msg.StreamsData
		w.updateSparklines()
		w.updateTableRows()
		logger.Info().Int("streams_after_update", len(w.streams)).Int("table_rows", len(w.table.Rows())).Msg("Data update completed")

	case tea.KeyMsg:
		if w.focused {
			// Let the table handle navigation
			w.table, cmd = w.table.Update(msg)
		}
	}

	if cmd != nil {
		logger.Warn().Str("cmd_type", fmt.Sprintf("%T", cmd)).Msg("Streams widget returning command")
	}

	return w, cmd
}

// View implements tea.Model
func (w StreamsTableWidget) View() string {
	if w.width == 0 {
		logger.Warn().Msg("View() called with width=0")
		return w.styles.Container.Render("No streams found")
	}
	if len(w.streams) == 0 {
		logger.Warn().Int("width", w.width).Int("height", w.height).Msg("View() called with no streams")
		return w.styles.Container.Render("No streams found")
	}

	// Get the table view and measure it
	tableView := w.table.View()
	tableHeight := lipgloss.Height(tableView)
	tableWidth := lipgloss.Width(tableView)

	// Render with container styling, but ensure height constraint is enforced
	containerStyle := w.styles.Container.Width(w.width).Height(w.height)
	renderedView := containerStyle.Render(tableView)
	renderedHeight := lipgloss.Height(renderedView)
	renderedWidth := lipgloss.Width(renderedView)

	// If the rendered view exceeds our allocated height, we need to truncate
	if renderedHeight > w.height {
		logger.Warn().
			Int("allocated_height", w.height).
			Int("rendered_height", renderedHeight).
			Msg("StreamsTable exceeds allocated height, truncating")

		// Use lipgloss to truncate to fit the allocated height
		lines := strings.Split(renderedView, "\n")
		if len(lines) > w.height {
			lines = lines[:w.height]
		}
		renderedView = strings.Join(lines, "\n")
		renderedHeight = lipgloss.Height(renderedView)
	}

	logger.Info().
		Int("widget_width", w.width).
		Int("widget_height", w.height).
		Int("table_width", tableWidth).
		Int("table_height", tableHeight).
		Int("rendered_width", renderedWidth).
		Int("rendered_height", renderedHeight).
		Int("table_rows", len(w.table.Rows())).
		Int("streams_count", len(w.streams)).
		Msg("StreamsTable View() rendering")

	return renderedView
}

// SetSize implements Widget interface
func (w *StreamsTableWidget) SetSize(width, height int) {
	// Clamp height to reasonable bounds to prevent infinite loops
	maxHeight := 25 // Maximum reasonable height for the table
	originalHeight := height
	if height > maxHeight {
		logger.Warn().Int("requested_height", height).Int("clamped_height", maxHeight).
			Msg("Clamping table height to prevent render issues")
		height = maxHeight
	}

	logger.Info().
		Int("old_width", w.width).Int("old_height", w.height).
		Int("requested_width", width).Int("requested_height", originalHeight).
		Int("final_width", width).Int("final_height", height).
		Int("streams_count", len(w.streams)).
		Int("max_height_limit", w.MaxHeight()).
		Int("min_height_limit", w.MinHeight()).
		Msg("SetSize called")

	w.width = width
	w.height = height
	w.updateTableSize()
}

// SetFocused implements Widget interface
func (w *StreamsTableWidget) SetFocused(focused bool) {
	w.focused = focused
	if focused {
		w.table.Focus()
	} else {
		w.table.Blur()
	}
}

// MinHeight implements Widget interface
func (w StreamsTableWidget) MinHeight() int {
	return 8 // Minimum height to show at least headers and a couple rows
}

// MaxHeight implements Widget interface
func (w StreamsTableWidget) MaxHeight() int {
	streamCount := len(w.streams)
	if streamCount == 0 {
		// When no streams loaded, reserve minimal space
		return 3
	}
	// Header (1) + streams (N) + minimal buffer (2) = N + 3
	return streamCount + 3
}

// updateTableSize updates the table dimensions and column widths
func (w *StreamsTableWidget) updateTableSize() {
	if w.width <= 0 || w.height <= 0 {
		logger.Warn().Int("width", w.width).Int("height", w.height).Msg("updateTableSize: invalid dimensions")
		return
	}

	// Ensure table height respects widget bounds
	tableHeight := w.height - 2 // Account for container padding
	if tableHeight < 3 {
		tableHeight = 3 // Minimum viable height
	}

	logger.Info().
		Int("widget_width", w.width).Int("widget_height", w.height).
		Int("table_width", w.width).Int("table_height", tableHeight).
		Msg("updateTableSize: setting table dimensions")

	w.table.SetWidth(w.width)
	w.table.SetHeight(tableHeight)

	// Calculate responsive column widths
	cols := w.calculateColumnWidths()

	columns := []table.Column{
		{Title: "Stream", Width: cols.stream},
		{Title: "Entries", Width: cols.entries},
		{Title: "Trend", Width: cols.trend},
		{Title: "Size", Width: cols.size},
		{Title: "Groups", Width: cols.groups},
		{Title: "Last ID", Width: cols.lastID},
		{Title: "Memory RSS", Width: cols.memory},
	}

	w.table.SetColumns(columns)

	logger.Info().
		Int("stream_col", cols.stream).
		Int("entries_col", cols.entries).
		Int("size_col", cols.size).
		Int("groups_col", cols.groups).
		Int("lastid_col", cols.lastID).
		Int("memory_col", cols.memory).
		Msg("updateTableSize: column widths calculated")
}

// updateTableRows converts stream data to table rows
func (w *StreamsTableWidget) updateTableRows() {
	logger.Info().Int("stream_count", len(w.streams)).Msg("updateTableRows called")
	rows := make([]table.Row, len(w.streams))

	for i, stream := range w.streams {
		// Get sparkline for dedicated trend column
		sparklineStr := ""
		if sl, exists := w.sparklines[stream.Name]; exists {
			rawSparkline := sl.Render()
			// Trim any padding from the sparkline
			sparklineStr = strings.TrimSpace(rawSparkline)
		}

		// Format entries as just numbers now
		entriesContent := formatNumberWithCommas(stream.Length)

		// Format sparkline for trend column
		trendContent := sparklineStr
		if trendContent == "" {
			trendContent = "-" // Show dash if no sparkline data
		}

		rows[i] = table.Row{
			truncateString(stream.Name, w.getColumnWidth("stream")),
			entriesContent,
			truncateString(trendContent, w.getColumnWidth("trend")),
			formatBytes(stream.MemoryUsage),
			formatNumberWithCommas(stream.Groups),
			truncateString(stream.LastID, w.getColumnWidth("lastID")),
			formatBytes(stream.MemoryUsage),
		}
	}

	logger.Info().Int("rows_created", len(rows)).Msg("Setting table rows")
	w.table.SetRows(rows)
	logger.Info().Int("table_rows_after_set", len(w.table.Rows())).Msg("Table rows set complete")
}

// updateSparklines creates or updates sparklines for each stream
func (w *StreamsTableWidget) updateSparklines() {
	// Use the dedicated trend column width for sparklines
	trendWidth := w.getColumnWidth("trend")

	// Leave some margin for padding/borders
	sparklineWidth := trendWidth - 2

	if sparklineWidth < 5 {
		sparklineWidth = 5
	}
	if sparklineWidth > 30 {
		sparklineWidth = 30
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
	selectedIdx := w.table.Cursor()
	if selectedIdx >= 0 && selectedIdx < len(w.streams) {
		return &w.streams[selectedIdx]
	}
	return nil
}

// getColumnWidth returns the width of a specific column
func (w StreamsTableWidget) getColumnWidth(columnName string) int {
	cols := w.calculateColumnWidths()
	switch columnName {
	case "stream":
		return cols.stream
	case "entries":
		return cols.entries
	case "trend":
		return cols.trend
	case "size":
		return cols.size
	case "groups":
		return cols.groups
	case "lastID":
		return cols.lastID
	case "memory":
		return cols.memory
	default:
		return 10
	}
}

// Column widths structure
type columnWidths struct {
	stream  int
	entries int
	trend   int
	size    int
	groups  int
	lastID  int
	memory  int
}

// calculateColumnWidths determines responsive column widths based on terminal size and content
func (w StreamsTableWidget) calculateColumnWidths() columnWidths {
	// Calculate content-based widths
	maxStreamLen := len("Stream")  // Header width as minimum
	maxLastIDLen := len("Last ID") // Header width as minimum

	for _, stream := range w.streams {
		if len(stream.Name) > maxStreamLen {
			maxStreamLen = len(stream.Name)
		}
		if len(stream.LastID) > maxLastIDLen {
			maxLastIDLen = len(stream.LastID)
		}
	}

	// Add some padding for content
	maxStreamLen += 2
	maxLastIDLen += 2

	// Minimum column widths
	minWidths := columnWidths{
		stream:  maxStreamLen, // Content-based stream name column
		entries: 10,           // Just for numbers now, no sparklines
		trend:   20,           // Fixed reasonable width for sparkline column
		size:    8,
		groups:  8,
		lastID:  maxLastIDLen, // Content-based last ID column
		memory:  12,
	}

	// Account for table padding and borders
	overhead := 15 // Approximate overhead for table styling
	availableWidth := w.width - overhead

	if availableWidth <= 0 {
		return minWidths
	}

	// Calculate total minimum width needed
	totalMinWidth := minWidths.stream + minWidths.entries + minWidths.trend + minWidths.size +
		minWidths.groups + minWidths.lastID + minWidths.memory

	if availableWidth <= totalMinWidth {
		return minWidths
	}

	// Distribute extra space proportionally to flexible columns
	extraSpace := availableWidth - totalMinWidth
	result := minWidths

	// Distribute extra space to entries and memory columns
	entriesExtra := extraSpace / 2
	memoryExtra := extraSpace - entriesExtra

	result.entries += entriesExtra
	result.memory += memoryExtra

	return result
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
