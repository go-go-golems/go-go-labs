package widgets

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HeaderWidget displays server info, uptime, and refresh rate
type HeaderWidget struct {
	width       int
	height      int
	serverData  ServerData
	refreshRate time.Duration
	demoMode    bool
	styles      HeaderStyles
}

type HeaderStyles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Info      lipgloss.Style
}

// NewHeaderWidget creates a new header widget
func NewHeaderWidget(styles HeaderStyles) HeaderWidget {
	return HeaderWidget{
		styles: styles,
		height: 2, // Fixed height
	}
}

// Init implements tea.Model
func (w HeaderWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w HeaderWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width

	case DataUpdateMsg:
		w.serverData = msg.ServerData
	}

	return w, nil
}

// View implements tea.Model
func (w HeaderWidget) View() string {
	if w.width == 0 {
		return ""
	}

	title := w.styles.Title.Render("Redis Streams Monitor (top-like)")

	uptime := formatDuration(w.serverData.Uptime)
	refresh := fmt.Sprintf("Refresh: %v", w.refreshRate)
	mode := ""
	if w.demoMode {
		mode = " [DEMO]"
	}

	info := w.styles.Info.Render(fmt.Sprintf("Uptime: %s   %s%s", uptime, refresh, mode))

	// Create horizontal layout
	contentWidth := w.width - w.styles.Container.GetHorizontalFrameSize()
	spacerWidth := contentWidth - lipgloss.Width(title) - lipgloss.Width(info)
	if spacerWidth < 0 {
		spacerWidth = 0
	}

	spacer := lipgloss.NewStyle().Width(spacerWidth).Render("")

	headerContent := lipgloss.JoinHorizontal(lipgloss.Top, title, spacer, info)

	return w.styles.Container.Width(w.width).Render(headerContent)
}

// SetSize implements Widget interface
func (w *HeaderWidget) SetSize(width, height int) {
	w.width = width
	// Height is fixed at 2 lines for header
}

// SetFocused implements Widget interface
func (w *HeaderWidget) SetFocused(focused bool) {
	// Header doesn't need focus handling
}

// MinHeight implements Widget interface
func (w HeaderWidget) MinHeight() int {
	return 2
}

// MaxHeight implements Widget interface
func (w HeaderWidget) MaxHeight() int {
	return 2
}

// SetRefreshRate sets the refresh rate to display
func (w *HeaderWidget) SetRefreshRate(rate time.Duration) {
	w.refreshRate = rate
}

// SetDemoMode sets demo mode display
func (w *HeaderWidget) SetDemoMode(demo bool) {
	w.demoMode = demo
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return "00:00:00"
	}
	if d < time.Minute {
		return fmt.Sprintf("00:00:%02d", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("00:%02d:%02d", int(d.Minutes()), int(d.Seconds())%60)
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%02d:%02d:%02d", int(d.Hours()), int(d.Minutes())%60, int(d.Seconds())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dd %02d:%02d", days, hours, minutes)
}
