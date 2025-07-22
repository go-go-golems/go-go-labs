package components

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/pkg/tui/models"
	"github.com/go-go-golems/go-go-labs/pkg/tui/styles"
)

// MetricsView shows server metrics with progress bars and sparklines
type MetricsView struct {
	styles         styles.Styles
	serverData     models.ServerData
	throughputData []float64
	width          int
	height         int
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
	return &MetricsView{
		styles:         styles,
		throughputData: make([]float64, 0, 20),
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
		// Add throughput to sparkline data
		v.addThroughputPoint(msg.ServerData.Throughput)

	case ThroughputDataMsg:
		v.addThroughputPoint(msg.Value)
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
	throughputSparkline := v.renderSparkline(v.throughputData)

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

// addThroughputPoint adds a new throughput data point for sparkline
func (v *MetricsView) addThroughputPoint(value float64) {
	v.throughputData = append(v.throughputData, value)

	// Keep only last 20 points
	if len(v.throughputData) > 20 {
		v.throughputData = v.throughputData[1:]
	}
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

// renderSparkline creates a text-based sparkline for throughput
func (v *MetricsView) renderSparkline(data []float64) string {
	if len(data) == 0 {
		return ""
	}

	// Normalize data to 0-1 range
	var min, max float64
	if len(data) > 0 {
		min, max = data[0], data[0]
		for _, value := range data {
			if value < min {
				min = value
			}
			if value > max {
				max = value
			}
		}
	}

	bars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}
	var result strings.Builder

	for _, value := range data {
		var normalized float64
		if max != min {
			normalized = (value - min) / (max - min)
		} else {
			normalized = 0.5
		}

		if normalized <= 0 {
			result.WriteRune(' ')
		} else if normalized >= 1 {
			result.WriteRune(bars[len(bars)-1])
		} else {
			idx := int(normalized * float64(len(bars)-1))
			result.WriteRune(bars[idx])
		}
	}

	return result.String()
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
