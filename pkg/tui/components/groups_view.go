package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// GroupsView manages the consumer groups table display
type GroupsView struct {
	styles  styles.Styles
	streams []models.StreamData
	width   int
	height  int
}

// GroupsDataMsg contains group data updates
type GroupsDataMsg struct {
	Streams []models.StreamData
}

// NewGroupsView creates a new groups view
func NewGroupsView(styles styles.Styles) *GroupsView {
	return &GroupsView{
		styles: styles,
	}
}

// Init implements tea.Model
func (v *GroupsView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (v *GroupsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case GroupsDataMsg:
		v.streams = msg.Streams
	}

	return v, nil
}

// View implements tea.Model
func (v *GroupsView) View() string {
	var rows []string
	header := fmt.Sprintf("%-12s %-12s %-15s %-8s %-12s",
		"Group", "Stream", "Consumers", "Pending", "Idle Time")
	rows = append(rows, v.styles.Selected.Render(header))

	for _, stream := range v.streams {
		for _, group := range stream.ConsumerGroups {
			var consumerNames []string
			var maxIdle time.Duration

			for _, consumer := range group.Consumers {
				consumerNames = append(consumerNames, fmt.Sprintf("%s(%d)", consumer.Name, consumer.Pending))
				if consumer.Idle > maxIdle {
					maxIdle = consumer.Idle
				}
			}

			row := fmt.Sprintf("%-12s %-12s %-15s %-8d %-12s",
				truncateString(group.Name, 12),
				truncateString(group.Stream, 12),
				truncateString(strings.Join(consumerNames, " "), 15),
				group.Pending,
				v.formatDuration(maxIdle))

			rows = append(rows, v.styles.Unselected.Render(row))
		}
	}

	if len(rows) == 1 {
		rows = append(rows, v.styles.Unselected.Render("No consumer groups found"))
	}

	content := strings.Join(rows, "\n")
	return v.styles.GroupTable.Render(content)
}

// formatDuration formats a duration into a human readable string
func (v *GroupsView) formatDuration(d time.Duration) string {
	if d == 0 {
		return "active"
	}

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}
