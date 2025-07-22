package widgets

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// GroupsTableWidget displays consumer groups detail
type GroupsTableWidget struct {
	width       int
	height      int
	focused     bool
	groups      []GroupData
	selectedIdx int
	styles      GroupsTableStyles
}

type GroupsTableStyles struct {
	Container    lipgloss.Style
	Table        lipgloss.Style
	HeaderRow    lipgloss.Style
	Row          lipgloss.Style
	SelectedRow  lipgloss.Style
	Title        lipgloss.Style
}

// NewGroupsTableWidget creates a new groups table widget
func NewGroupsTableWidget(styles GroupsTableStyles) GroupsTableWidget {
	return GroupsTableWidget{
		styles: styles,
	}
}

// Init implements tea.Model
func (w GroupsTableWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w GroupsTableWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		
	case DataUpdateMsg:
		// Flatten all groups from all streams
		w.groups = nil
		for _, stream := range msg.StreamsData {
			w.groups = append(w.groups, stream.ConsumerGroups...)
		}
		
		// Ensure selected index is valid
		if w.selectedIdx >= len(w.groups) {
			w.selectedIdx = len(w.groups) - 1
		}
		if w.selectedIdx < 0 && len(w.groups) > 0 {
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
				if w.selectedIdx < len(w.groups)-1 {
					w.selectedIdx++
				}
			}
		}
	}
	
	return w, nil
}

// View implements tea.Model
func (w GroupsTableWidget) View() string {
	if w.width == 0 {
		return ""
	}
	
	if len(w.groups) == 0 {
		title := w.styles.Title.Render("Groups Detail:")
		noGroups := w.styles.Container.Render("No consumer groups found")
		return title + "\n" + noGroups
	}
	
	var rows []string
	
	// Title
	rows = append(rows, w.styles.Title.Render("Groups Detail:"))
	
	// Table header with border
	topBorder := "┌─────────┬─────────┬───────────────┬──────────┬──────────────┐"
	headerRow := "│ Group   │ Stream  │ Consumers     │ Pending  │ Idle Time    │"
	divider := "├─────────┼─────────┼───────────────┼──────────┼──────────────┤"
	
	rows = append(rows, topBorder)
	rows = append(rows, w.styles.HeaderRow.Render(headerRow))
	rows = append(rows, divider)
	
	// Group rows
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
		
		row := fmt.Sprintf("│ %-7s │ %-7s │ %-13s │ %8d │ %-12s │",
			truncateString(group.Name, 7),
			truncateString(group.Stream, 7),
			truncateString(consumersStr, 13),
			group.Pending,
			formatDurationShort(maxIdle))
		
		// Apply styling based on selection
		if i == w.selectedIdx && w.focused {
			rows = append(rows, w.styles.SelectedRow.Render(row))
		} else {
			rows = append(rows, w.styles.Row.Render(row))
		}
	}
	
	// Bottom border
	bottomBorder := "└─────────┴─────────┴───────────────┴──────────┴──────────────┘"
	rows = append(rows, bottomBorder)
	
	content := strings.Join(rows, "\n")
	return w.styles.Container.Width(w.width).Render(content)
}

// SetSize implements Widget interface
func (w *GroupsTableWidget) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// SetFocused implements Widget interface
func (w *GroupsTableWidget) SetFocused(focused bool) {
	w.focused = focused
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
	if w.selectedIdx >= 0 && w.selectedIdx < len(w.groups) {
		return &w.groups[w.selectedIdx]
	}
	return nil
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
