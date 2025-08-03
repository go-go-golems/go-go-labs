package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/bobatea/pkg/sparkline"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// MetricsView shows server metrics with progress bars and sparklines
type MetricsView struct {
	styles              styles.Styles
	serverData          models.ServerData
	throughputSparkline *sparkline.Sparkline
	width               int
	height              int
}

// MetricsDataMsg contains server metrics updates
type MetricsDataMsg struct {
	ServerData models.ServerData
}

// ThroughputDataMsg contains throughput sparkline updates
type ThroughputDataMsg struct {
	Value float64
}

// NewMetricsView creates a new metrics view
func NewMetricsView(styles styles.Styles) *MetricsView {
	// Configure sparkline with throughput styling
	sparklineConfig := sparkline.Config{
		Width:        20,
		Height:       1,
		MaxPoints:    20,
		Style:        sparkline.StyleBars,
		DefaultStyle: styles.Throughput,
	}

	return &MetricsView{
		styles:              styles,
		throughputSparkline: sparkline.New(sparklineConfig),
	}
}

// Init implements tea.Model
func (v *MetricsView) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (v *MetricsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height

	case MetricsDataMsg:
		v.serverData = msg.ServerData
		// Add throughput to sparkline
		v.throughputSparkline.AddPoint(msg.ServerData.Throughput)

	case ThroughputDataMsg:
		v.throughputSparkline.AddPoint(msg.Value)
	}

	return v, nil
}

// View implements tea.Model
func (v *MetricsView) View() string {
	memoryPercent := float64(v.serverData.MemoryUsed) / float64(v.serverData.MemoryTotal) * 100
	if v.serverData.MemoryTotal == 0 {
		memoryPercent = 0
	}

	memoryBar := v.renderProgressBar(memoryPercent, 30)
	throughputSparkline := v.throughputSparkline.Render()

	content := fmt.Sprintf(`Server Information:
  Redis Version: %s
  Uptime: %s
  
Memory Usage:
  Used: %s / %s (%.1f%%)
  %s
  
Global Throughput: %.1f msg/s
%s`,
		v.serverData.Version,
		v.formatDuration(v.serverData.Uptime),
		formatBytes(v.serverData.MemoryUsed),
		formatBytes(v.serverData.MemoryTotal),
		memoryPercent,
		v.styles.Memory.Render(memoryBar),
		v.serverData.Throughput,
		v.styles.Throughput.Render(throughputSparkline))

	return v.styles.StreamTable.Render(content)
}

// renderProgressBar creates a progress bar
func (v *MetricsView) renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}

	var bar strings.Builder
	for i := 0; i < width; i++ {
		if i < filled {
			bar.WriteRune('■')
		} else {
			bar.WriteRune('□')
		}
	}

	return bar.String()
}

// formatDuration formats a duration into a human readable string
func (v *MetricsView) formatDuration(d time.Duration) string {
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
