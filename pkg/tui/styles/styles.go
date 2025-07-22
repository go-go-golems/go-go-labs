package styles

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/tui/widgets"
)

// Styles contains all the styles for the TUI widgets
type Styles struct {
	Header       widgets.HeaderStyles
	StreamsTable widgets.StreamsTableStyles
	GroupsTable  widgets.GroupsTableStyles
	Alerts       widgets.AlertsStyles
	Metrics      widgets.MetricsStyles
	Footer       widgets.FooterStyles
}

// NewStyles creates a new Styles instance with default styling
func NewStyles() Styles {
	// Define common colors
	primaryColor := lipgloss.Color("57")   // Blue
	accentColor := lipgloss.Color("69")    // Cyan
	borderColor := lipgloss.Color("240")   // Gray
	textColor := lipgloss.Color("15")      // White
	mutedColor := lipgloss.Color("246")    // Light gray
	errorColor := lipgloss.Color("196")    // Red
	warningColor := lipgloss.Color("214")  // Orange
	infoColor := lipgloss.Color("39")      // Light blue

	return Styles{
		Header: widgets.HeaderStyles{
			Container: lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(textColor).
				Bold(true).
				Padding(0, 1),
			Title: lipgloss.NewStyle().
				Foreground(textColor).
				Bold(true),
			Info: lipgloss.NewStyle().
				Foreground(textColor),
		},

		StreamsTable: widgets.StreamsTableStyles{
			Container: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1),
			Table: lipgloss.NewStyle(),
			HeaderRow: lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(textColor).
				Bold(true),
			Row: lipgloss.NewStyle().
				Foreground(textColor),
			SelectedRow: lipgloss.NewStyle().
				Background(accentColor).
				Foreground(textColor).
				Bold(true),
			SparklineRow: lipgloss.NewStyle().
				Foreground(mutedColor),
		},

		GroupsTable: widgets.GroupsTableStyles{
			Container: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1),
			Table: lipgloss.NewStyle(),
			HeaderRow: lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(textColor).
				Bold(true),
			Row: lipgloss.NewStyle().
				Foreground(textColor),
			SelectedRow: lipgloss.NewStyle().
				Background(accentColor).
				Foreground(textColor).
				Bold(true),
			Title: lipgloss.NewStyle().
				Foreground(textColor).
				Bold(true),
		},

		Alerts: widgets.AlertsStyles{
			Container: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1),
			Title: lipgloss.NewStyle().
				Foreground(textColor).
				Bold(true),
			Alert: lipgloss.NewStyle().
				Foreground(errorColor),
			Warning: lipgloss.NewStyle().
				Foreground(warningColor),
			Info: lipgloss.NewStyle().
				Foreground(infoColor),
		},

		Metrics: widgets.MetricsStyles{
			Container: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(borderColor).
				Padding(1),
			Label: lipgloss.NewStyle().
				Foreground(textColor).
				Bold(true),
			Value: lipgloss.NewStyle().
				Foreground(accentColor).
				Bold(true),
		},

		Footer: widgets.FooterStyles{
			Container: lipgloss.NewStyle().
				Background(borderColor).
				Foreground(textColor).
				Padding(0, 1),
			Commands: lipgloss.NewStyle().
				Foreground(textColor),
		},
	}
}
