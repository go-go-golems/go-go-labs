// Package styles contains all lipgloss styling definitions for the TUI components
package styles

import "github.com/charmbracelet/lipgloss"

// Styles contains all the lipgloss styles used throughout the TUI
type Styles struct {
	Header      lipgloss.Style
	Title       lipgloss.Style
	Status      lipgloss.Style
	StreamTable lipgloss.Style
	GroupTable  lipgloss.Style
	Sparkline   lipgloss.Style
	Border      lipgloss.Style
	Selected    lipgloss.Style
	Unselected  lipgloss.Style
	Memory      lipgloss.Style
	Throughput  lipgloss.Style
}

// NewStyles creates the default styles for the TUI
func NewStyles() Styles {
	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),

		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),

		Status: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),

		StreamTable: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1),

		GroupTable: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F25D94")).
			Padding(1),

		Sparkline: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),

		Border: lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")),

		Selected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),

		Unselected: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),

		Memory: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F25D94")),

		Throughput: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")),
	}
}
