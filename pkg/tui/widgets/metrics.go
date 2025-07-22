package widgets

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/sparkline"
)

// MetricsWidget displays global throughput and memory usage
type MetricsWidget struct {
	width             int
	height            int
	serverData        ServerData
	streamsData       []StreamData
	throughputHistory []float64
	sparkline         *sparkline.Sparkline
	memoryProgress    progress.Model
	styles            MetricsStyles
}

type MetricsStyles struct {
	Container lipgloss.Style
	Label     lipgloss.Style
	Value     lipgloss.Style
}

// NewMetricsWidget creates a new metrics widget
func NewMetricsWidget(styles MetricsStyles) MetricsWidget {
	// Initialize sparkline for throughput
	sparklineConfig := sparkline.Config{
		Width:     30,
		Height:    1,
		MaxPoints: 30,
		Style:     sparkline.StyleBars,
	}
	sl := sparkline.New(sparklineConfig)
	
	// Initialize progress bar for memory
	memProgress := progress.New(progress.WithDefaultGradient())
	memProgress.Width = 40
	
	return MetricsWidget{
		sparkline:         sl,
		memoryProgress:    memProgress,
		throughputHistory: make([]float64, 0, 30),
		styles:            styles,
	}
}

// Init implements tea.Model
func (w MetricsWidget) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (w MetricsWidget) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		
		// Update progress bar width based on available space
		progressWidth := w.width - 40 // Leave space for labels and values
		if progressWidth > 20 && progressWidth < 60 {
			w.memoryProgress.Width = progressWidth
		}
		
	case DataUpdateMsg:
		w.serverData = msg.ServerData
		w.streamsData = msg.StreamsData
		w.updateMetrics()
	}
	
	// Update progress bar
	progressModel, progressCmd := w.memoryProgress.Update(msg)
	w.memoryProgress = progressModel.(progress.Model)
	cmd = tea.Sequence(cmd, progressCmd)
	
	return w, cmd
}

// View implements tea.Model
func (w MetricsWidget) View() string {
	if w.width == 0 {
		return ""
	}
	
	// Calculate total throughput
	totalThroughput := w.calculateTotalThroughput()
	
	// Throughput line with sparkline
	throughputSparkline := w.sparkline.Render()
	throughputLine := fmt.Sprintf("Global Throughput %s %s",
		throughputSparkline,
		w.styles.Value.Render(fmt.Sprintf("%d msg/s", int(totalThroughput))))
	
	// Memory usage line with progress bar
	memoryBar := w.memoryProgress.View()
	memoryLine := fmt.Sprintf("Memory Usage    %s %s",
		memoryBar,
		w.styles.Value.Render(fmt.Sprintf("%s / %s",
			formatBytes(w.serverData.MemoryUsed),
			formatBytes(w.serverData.MemoryTotal))))
	
	content := strings.Join([]string{throughputLine, memoryLine}, "\n")
	return w.styles.Container.Width(w.width).Render(content)
}

// SetSize implements Widget interface
func (w *MetricsWidget) SetSize(width, height int) {
	w.width = width
	w.height = height
	
	// Update progress bar width
	progressWidth := width - 40
	if progressWidth > 20 && progressWidth < 60 {
		w.memoryProgress.Width = progressWidth
	}
}

// SetFocused implements Widget interface
func (w *MetricsWidget) SetFocused(focused bool) {
	// Metrics widget doesn't need focus handling
}

// MinHeight implements Widget interface
func (w MetricsWidget) MinHeight() int {
	return 2
}

// MaxHeight implements Widget interface
func (w MetricsWidget) MaxHeight() int {
	return 2
}

// updateMetrics updates internal metrics calculations
func (w *MetricsWidget) updateMetrics() {
	// Calculate total throughput
	totalThroughput := w.calculateTotalThroughput()
	
	// Update throughput history
	w.throughputHistory = append(w.throughputHistory, totalThroughput)
	if len(w.throughputHistory) > 30 {
		w.throughputHistory = w.throughputHistory[1:]
	}
	
	// Update sparkline
	w.sparkline.SetData(w.throughputHistory)
	
	// Update memory progress
	var memoryPercent float64
	if w.serverData.MemoryTotal > 0 {
		memoryPercent = float64(w.serverData.MemoryUsed) / float64(w.serverData.MemoryTotal)
	}
	w.memoryProgress.SetPercent(memoryPercent)
}

// calculateTotalThroughput sums up throughput from all streams
func (w MetricsWidget) calculateTotalThroughput() float64 {
	var total float64
	for _, stream := range w.streamsData {
		if len(stream.MessageRates) > 0 {
			// Use the latest message rate
			total += stream.MessageRates[len(stream.MessageRates)-1]
		}
	}
	return total
}
