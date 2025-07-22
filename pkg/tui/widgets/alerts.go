package widgets

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// AlertsWidget displays memory and trim alerts
type AlertsWidget struct {
	width       int
	height      int
	streamsData []StreamData
	styles      AlertsStyles
}

type AlertsStyles struct {
	Container lipgloss.Style
	Title     lipgloss.Style
	Alert     lipgloss.Style
	Warning   lipgloss.Style
	Info      lipgloss.Style
}

// NewAlertsWidget creates a new alerts widget
func NewAlertsWidget(styles AlertsStyles) AlertsWidget {
	return AlertsWidget{
		styles: styles,
	}
}

// Init implements tea.Model
func (w AlertsWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w AlertsWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		
	case DataUpdateMsg:
		w.streamsData = msg.StreamsData
	}
	
	return w, nil
}

// View implements tea.Model
func (w AlertsWidget) View() string {
	if w.width == 0 {
		return ""
	}
	
	var alerts []string
	
	// Title
	alerts = append(alerts, w.styles.Title.Render("Trim / Memory Alerts:"))
	
	if len(w.streamsData) == 0 {
		alerts = append(alerts, w.styles.Info.Render("  No streams to monitor"))
		return w.styles.Container.Width(w.width).Render(strings.Join(alerts, "\n"))
	}
	
	hasAlerts := false
	
	// Check each stream for alerts
	for _, stream := range w.streamsData {
		alert := w.generateStreamAlert(stream)
		if alert != "" {
			alerts = append(alerts, alert)
			hasAlerts = true
		}
	}
	
	// If no specific alerts, show general status
	if !hasAlerts {
		alerts = append(alerts, w.styles.Info.Render("  All streams within normal parameters"))
	}
	
	content := strings.Join(alerts, "\n")
	return w.styles.Container.Width(w.width).Render(content)
}

// SetSize implements Widget interface
func (w *AlertsWidget) SetSize(width, height int) {
	w.width = width
	w.height = height
}

// SetFocused implements Widget interface
func (w *AlertsWidget) SetFocused(focused bool) {
	// Alerts widget doesn't need focus handling
}

// MinHeight implements Widget interface
func (w AlertsWidget) MinHeight() int {
	return 3 // Title + at least 2 lines of content
}

// MaxHeight implements Widget interface
func (w AlertsWidget) MaxHeight() int {
	// Title + one alert per stream + general status
	return 2 + len(w.streamsData) + 1
}

// generateStreamAlert creates an alert message for a stream based on its characteristics
func (w AlertsWidget) generateStreamAlert(stream StreamData) string {
	const (
		highThreshold = 500000  // 500k entries
		medThreshold  = 100000  // 100k entries
	)
	
	if stream.Length > highThreshold {
		// High usage - show trim rate alert
		rate := (stream.Length - highThreshold) / 10000 // Simulated trim rate
		return w.styles.Alert.Render(fmt.Sprintf("  • %s: maxlen=500k (approx %s entries) → trim rate: %d/s",
			stream.Name, formatNumber(stream.Length), rate))
			
	} else if stream.Length > medThreshold {
		// Medium usage - show threshold warning
		return w.styles.Warning.Render(fmt.Sprintf("  • %s: maxlen=100k (%s entries) → within threshold",
			stream.Name, formatNumber(stream.Length)))
			
	} else if stream.Length > 0 {
		// Low usage - show info
		return w.styles.Info.Render(fmt.Sprintf("  • %s: no maxlen configured",
			stream.Name))
	}
	
	return ""
}

// formatNumber formats large numbers with k/m suffixes
func formatNumber(n int64) string {
	if n < 1000 {
		return fmt.Sprintf("%d", n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%.1fm", float64(n)/1000000)
}
