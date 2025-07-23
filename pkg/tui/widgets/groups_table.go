package widgets

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GroupsTableWidget displays consumer groups detail
type GroupsTableWidget struct {
	width   int
	height  int
	focused bool
	groups  []GroupData
	table   table.Model
	styles  GroupsTableStyles
}

type GroupsTableStyles struct {
	Container   lipgloss.Style
	Table       lipgloss.Style
	HeaderRow   lipgloss.Style
	Row         lipgloss.Style
	SelectedRow lipgloss.Style
	Title       lipgloss.Style
}

// NewGroupsTableWidget creates a new groups table widget
func NewGroupsTableWidget(styles GroupsTableStyles) GroupsTableWidget {
	// Create columns for the table
	columns := []table.Column{
		{Title: "Group", Width: 15},
		{Title: "Stream", Width: 15},
		{Title: "Consumers", Width: 20},
		{Title: "Pending", Width: 10},
		{Title: "Idle Time", Width: 12},
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

	return GroupsTableWidget{
		styles: styles,
		table:  t,
	}
}

// Init implements tea.Model
func (w GroupsTableWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w GroupsTableWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Don't handle WindowSizeMsg directly - let the root model handle size allocation
		// The root model will call SetSize() with the proper allocated dimensions

	case DataUpdateMsg:
		// Flatten all groups from all streams
		w.groups = nil
		for _, stream := range msg.StreamsData {
			w.groups = append(w.groups, stream.ConsumerGroups...)
		}
		w.updateTableRows()

	case tea.KeyMsg:
		if w.focused {
			// Let the table handle navigation
			w.table, cmd = w.table.Update(msg)
		}
	}

	return w, cmd
}

// View implements tea.Model
func (w GroupsTableWidget) View() string {
	if w.width == 0 {
		return w.styles.Container.Render("Groups Detail:\nNo consumer groups found")
	}

	if len(w.groups) == 0 {
		return w.styles.Container.Render("Groups Detail:\nNo consumer groups found")
	}

	// Add title and table view
	title := w.styles.Title.Render("Groups Detail:")
	tableView := w.table.View()
	
	content := title + "\n" + tableView
	return w.styles.Container.Width(w.width).Height(w.height).Render(content)
}

// SetSize implements Widget interface
func (w *GroupsTableWidget) SetSize(width, height int) {
	w.width = width
	w.height = height
	w.updateTableSize()
}

// SetFocused implements Widget interface
func (w *GroupsTableWidget) SetFocused(focused bool) {
	w.focused = focused
	if focused {
		w.table.Focus()
	} else {
		w.table.Blur()
	}
}

// MinHeight implements Widget interface
func (w GroupsTableWidget) MinHeight() int {
	// Title + header + borders + at least 1 group
	return 5
}

// MaxHeight implements Widget interface
func (w GroupsTableWidget) MaxHeight() int {
	// Title + header + borders + all groups
	return 4 + len(w.groups)
}

// GetSelectedGroup returns the currently selected group
func (w GroupsTableWidget) GetSelectedGroup() *GroupData {
	selectedIdx := w.table.Cursor()
	if selectedIdx >= 0 && selectedIdx < len(w.groups) {
		return &w.groups[selectedIdx]
	}
	return nil
}

// updateTableSize updates the table dimensions and column widths
func (w *GroupsTableWidget) updateTableSize() {
	if w.width <= 0 || w.height <= 0 {
		return
	}

	// Ensure table height respects widget bounds
	tableHeight := w.height - 2 // Account for title and container padding
	if tableHeight < 3 {
		tableHeight = 3 // Minimum viable height
	}

	w.table.SetWidth(w.width)
	w.table.SetHeight(tableHeight)

	// Calculate responsive column widths
	cols := w.calculateColumnWidths()

	columns := []table.Column{
		{Title: "Group", Width: cols.group},
		{Title: "Stream", Width: cols.stream},
		{Title: "Consumers", Width: cols.consumers},
		{Title: "Pending", Width: cols.pending},
		{Title: "Idle Time", Width: cols.idleTime},
	}

	w.table.SetColumns(columns)
}

// updateTableRows converts group data to table rows
func (w *GroupsTableWidget) updateTableRows() {
	rows := make([]table.Row, len(w.groups))

	for i, group := range w.groups {
		// Build consumer info string
		var consumerStrs []string
		var maxIdle time.Duration

		for _, consumer := range group.Consumers {
			consumerStrs = append(consumerStrs, fmt.Sprintf("%s(%d)", consumer.Name, consumer.Pending))
			if consumer.Idle > maxIdle {
				maxIdle = consumer.Idle
			}
		}
		consumersStr := strings.Join(consumerStrs, " ")

		rows[i] = table.Row{
			truncateString(group.Name, w.getColumnWidth("group")),
			truncateString(group.Stream, w.getColumnWidth("stream")),
			truncateString(consumersStr, w.getColumnWidth("consumers")),
			fmt.Sprintf("%d", group.Pending),
			formatDurationShort(maxIdle),
		}
	}

	w.table.SetRows(rows)
}

// Column widths structure for groups table
type groupColumnWidths struct {
	group     int
	stream    int
	consumers int
	pending   int
	idleTime  int
}

// calculateColumnWidths determines responsive column widths based on terminal size and content
func (w GroupsTableWidget) calculateColumnWidths() groupColumnWidths {
	// Calculate content-based widths
	maxGroupLen := len("Group") // Header width as minimum
	maxStreamLen := len("Stream") // Header width as minimum
	
	for _, group := range w.groups {
		if len(group.Name) > maxGroupLen {
			maxGroupLen = len(group.Name)
		}
		if len(group.Stream) > maxStreamLen {
			maxStreamLen = len(group.Stream)
		}
	}
	
	// Add some padding for content
	maxGroupLen += 2
	maxStreamLen += 2
	
	// Minimum column widths
	minWidths := groupColumnWidths{
		group:     maxGroupLen,  // Content-based group name column
		stream:    maxStreamLen, // Content-based stream name column
		consumers: 15,
		pending:   8,
		idleTime:  12,
	}

	// Account for table padding and borders
	overhead := 15 // Approximate overhead for table styling
	availableWidth := w.width - overhead

	if availableWidth <= 0 {
		return minWidths
	}

	// Calculate total minimum width needed
	totalMinWidth := minWidths.group + minWidths.stream + minWidths.consumers + minWidths.pending + minWidths.idleTime

	if availableWidth <= totalMinWidth {
		return minWidths
	}

	// Distribute extra space proportionally
	extraSpace := availableWidth - totalMinWidth
	result := minWidths

	// Give extra space to consumers column primarily
	result.consumers += extraSpace

	return result
}

// getColumnWidth returns the width of a specific column
func (w GroupsTableWidget) getColumnWidth(columnName string) int {
	cols := w.calculateColumnWidths()
	switch columnName {
	case "group":
		return cols.group
	case "stream":
		return cols.stream
	case "consumers":
		return cols.consumers
	case "pending":
		return cols.pending
	case "idleTime":
		return cols.idleTime
	default:
		return 10
	}
}

// formatDurationShort formats time.Duration for display in the table (shorter format)
func formatDurationShort(d time.Duration) string {
	if d == 0 {
		return "00:00:00"
	}
	if d < time.Second {
		return "00:00:00"
	}
	if d < time.Minute {
		return fmt.Sprintf("00:00:%02d", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("00:%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
}
