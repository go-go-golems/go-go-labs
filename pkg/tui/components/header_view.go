package components

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// HeaderView renders the top header with title and status info
type HeaderView struct {
	styles      styles.Styles
	serverData  models.ServerData
	demoMode    bool
	refreshRate time.Duration
	currentView string
	width       int
	height      int
}

// HeaderDataMsg contains header data updates
type HeaderDataMsg struct {
	ServerData  models.ServerData
	DemoMode    bool
	RefreshRate time.Duration
	CurrentView string
}

// NewHeaderView creates a new header view
func NewHeaderView(styles styles.Styles) *HeaderView {
	return &HeaderView{
		styles: styles,
	}
}

// Init implements tea.Model
func (v *HeaderView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (v *HeaderView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case HeaderDataMsg:
		v.serverData = msg.ServerData
		v.demoMode = msg.DemoMode
		v.refreshRate = msg.RefreshRate
		v.currentView = msg.CurrentView
	}

	return v, nil
}

// View implements tea.Model
func (v *HeaderView) View() string {
	var status string
	if v.demoMode {
		status = "DEMO MODE"
	} else {
		status = fmt.Sprintf("Uptime: %s", v.formatDuration(v.serverData.Uptime))
	}

	title := v.styles.Header.Render("Redis Streams Monitor (top-like)")

	// Add current view indicator
	viewIndicator := fmt.Sprintf("[%s]", v.currentView)
	viewStyle := v.styles.Selected.Copy().Bold(true)

	statusLine := v.styles.Status.Render(fmt.Sprintf("%s | %s | Refresh: %s | Memory: %s",
		status,
		viewStyle.Render(viewIndicator),
		v.refreshRate,
		formatBytes(v.serverData.MemoryUsed)))

	return lipgloss.JoinVertical(lipgloss.Left, title, statusLine)
}

// formatDuration formats a duration into a human readable string
func (v *HeaderView) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}

	if d < time.Hour {
		return fmt.Sprintf("%.0fm %.0fs", d.Minutes(), d.Seconds()-60*d.Minutes())
	}

	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}

	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
