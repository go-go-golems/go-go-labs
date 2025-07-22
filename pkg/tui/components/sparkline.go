package components

import (
	"fmt"
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SparklineStyle defines the visual style of the sparkline
type SparklineStyle int

const (
	StyleBars SparklineStyle = iota
	StyleDots
	StyleLine
	StyleFilled
)

// ColorRange defines color thresholds for values
type ColorRange struct {
	Min   float64
	Max   float64
	Style lipgloss.Style
}

// SparklineConfig holds configuration for the sparkline
type SparklineConfig struct {
	Width        int
	Height       int
	MaxPoints    int
	Style        SparklineStyle
	Title        string
	ShowValue    bool
	ShowMinMax   bool
	ColorRanges  []ColorRange
	DefaultStyle lipgloss.Style
}

// Sparkline represents a sparkline component
type Sparkline struct {
	config SparklineConfig
	data   []float64
	min    float64
	max    float64
}

// NewSparkline creates a new sparkline with the given configuration
func NewSparkline(config SparklineConfig) *Sparkline {
	if config.Width <= 0 {
		config.Width = 40
	}
	if config.Height <= 0 {
		config.Height = 8
	}
	if config.MaxPoints <= 0 {
		config.MaxPoints = config.Width
	}

	return &Sparkline{
		config: config,
		data:   make([]float64, 0, config.MaxPoints),
		min:    math.Inf(1),
		max:    math.Inf(-1),
	}
}

// Init implements tea.Model
func (s *Sparkline) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (s *Sparkline) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

// View implements tea.Model
func (s *Sparkline) View() string {
	if len(s.data) == 0 {
		return s.emptyView()
	}

	switch s.config.Style {
	case StyleBars:
		return s.barsView()
	case StyleDots:
		return s.dotsView()
	case StyleLine:
		return s.lineView()
	case StyleFilled:
		return s.filledView()
	default:
		return s.barsView()
	}
}

// AddPoint adds a new data point to the sparkline
func (s *Sparkline) AddPoint(value float64) {
	s.data = append(s.data, value)

	// Maintain rolling window
	if len(s.data) > s.config.MaxPoints {
		s.data = s.data[1:]
	}

	// Update min/max
	s.updateMinMax()
}

// AddPoints adds multiple data points at once
func (s *Sparkline) AddPoints(values []float64) {
	for _, v := range values {
		s.AddPoint(v)
	}
}

// SetData replaces all data with new values
func (s *Sparkline) SetData(data []float64) {
	s.data = make([]float64, 0, s.config.MaxPoints)
	s.min = math.Inf(1)
	s.max = math.Inf(-1)
	s.AddPoints(data)
}

// Clear removes all data points
func (s *Sparkline) Clear() {
	s.data = s.data[:0]
	s.min = math.Inf(1)
	s.max = math.Inf(-1)
}

// GetData returns a copy of the current data
func (s *Sparkline) GetData() []float64 {
	result := make([]float64, len(s.data))
	copy(result, s.data)
	return result
}

// GetLastValue returns the most recent value, or 0 if no data
func (s *Sparkline) GetLastValue() float64 {
	if len(s.data) == 0 {
		return 0
	}
	return s.data[len(s.data)-1]
}

// GetMinMax returns the current min and max values
func (s *Sparkline) GetMinMax() (float64, float64) {
	if len(s.data) == 0 {
		return 0, 0
	}
	return s.min, s.max
}

// updateMinMax recalculates min and max values
func (s *Sparkline) updateMinMax() {
	if len(s.data) == 0 {
		s.min = math.Inf(1)
		s.max = math.Inf(-1)
		return
	}

	s.min = s.data[0]
	s.max = s.data[0]

	for _, v := range s.data {
		if v < s.min {
			s.min = v
		}
		if v > s.max {
			s.max = v
		}
	}
}

// normalize scales a value to the given range
func (s *Sparkline) normalize(value float64, targetMin, targetMax int) int {
	if s.max == s.min {
		return targetMin
	}

	normalized := (value - s.min) / (s.max - s.min)
	scaled := float64(targetMin) + normalized*float64(targetMax-targetMin)
	return int(math.Round(scaled))
}

// getColorStyle returns the appropriate style for a value based on color ranges
func (s *Sparkline) getColorStyle(value float64) lipgloss.Style {
	for _, colorRange := range s.config.ColorRanges {
		if value >= colorRange.Min && value <= colorRange.Max {
			return colorRange.Style
		}
	}
	return s.config.DefaultStyle
}

// emptyView renders when there's no data
func (s *Sparkline) emptyView() string {
	var b strings.Builder

	if s.config.Title != "" {
		b.WriteString(s.config.Title + "\n")
	}

	emptyLine := strings.Repeat("─", s.config.Width)
	for i := 0; i < s.config.Height; i++ {
		if i == s.config.Height/2 {
			b.WriteString("No data")
		} else {
			b.WriteString(emptyLine)
		}
		if i < s.config.Height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// barsView renders vertical bars
func (s *Sparkline) barsView() string {
	var b strings.Builder

	if s.config.Title != "" {
		b.WriteString(s.config.Title + "\n")
	}

	if s.config.ShowValue || s.config.ShowMinMax {
		b.WriteString(s.renderHeader())
		b.WriteString("\n")
	}

	// Bars characters from bottom to top
	bars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	// Prepare data for display width
	displayData := s.prepareDisplayData()

	// Render bars
	for _, value := range displayData {
		if math.IsNaN(value) {
			b.WriteString(" ")
		} else {
			barHeight := s.normalize(value, 0, len(bars)-1)
			style := s.getColorStyle(value)
			b.WriteString(style.Render(bars[barHeight]))
		}
	}

	if s.config.ShowMinMax {
		b.WriteString("\n")
		b.WriteString(s.renderFooter())
	}

	return b.String()
}

// dotsView renders dots at different heights
func (s *Sparkline) dotsView() string {
	var b strings.Builder

	if s.config.Title != "" {
		b.WriteString(s.config.Title + "\n")
	}

	if s.config.ShowValue || s.config.ShowMinMax {
		b.WriteString(s.renderHeader())
		b.WriteString("\n")
	}

	displayData := s.prepareDisplayData()

	// Create a grid for dots
	grid := make([][]string, s.config.Height)
	for i := range grid {
		grid[i] = make([]string, len(displayData))
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Place dots
	for x, value := range displayData {
		if !math.IsNaN(value) {
			y := s.config.Height - 1 - s.normalize(value, 0, s.config.Height-1)
			if y >= 0 && y < s.config.Height {
				style := s.getColorStyle(value)
				grid[y][x] = style.Render("●")
			}
		}
	}

	// Render grid
	for _, row := range grid {
		b.WriteString(strings.Join(row, ""))
		b.WriteString("\n")
	}

	if s.config.ShowMinMax {
		b.WriteString(s.renderFooter())
	}

	return strings.TrimSuffix(b.String(), "\n")
}

// lineView renders a connected line
func (s *Sparkline) lineView() string {
	var b strings.Builder

	if s.config.Title != "" {
		b.WriteString(s.config.Title + "\n")
	}

	if s.config.ShowValue || s.config.ShowMinMax {
		b.WriteString(s.renderHeader())
		b.WriteString("\n")
	}

	displayData := s.prepareDisplayData()

	// Create a grid for the line
	grid := make([][]string, s.config.Height)
	for i := range grid {
		grid[i] = make([]string, len(displayData))
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Draw line segments
	for x := 0; x < len(displayData)-1; x++ {
		if !math.IsNaN(displayData[x]) && !math.IsNaN(displayData[x+1]) {
			y1 := s.config.Height - 1 - s.normalize(displayData[x], 0, s.config.Height-1)
			y2 := s.config.Height - 1 - s.normalize(displayData[x+1], 0, s.config.Height-1)

			style := s.getColorStyle(displayData[x])

			if y1 == y2 {
				// Horizontal line
				if y1 >= 0 && y1 < s.config.Height {
					grid[y1][x] = style.Render("─")
				}
			} else if y1 < y2 {
				// Going down
				if y1 >= 0 && y1 < s.config.Height {
					grid[y1][x] = style.Render("╲")
				}
			} else {
				// Going up
				if y1 >= 0 && y1 < s.config.Height {
					grid[y1][x] = style.Render("╱")
				}
			}
		}
	}

	// Add final point
	if len(displayData) > 0 && !math.IsNaN(displayData[len(displayData)-1]) {
		x := len(displayData) - 1
		y := s.config.Height - 1 - s.normalize(displayData[x], 0, s.config.Height-1)
		if y >= 0 && y < s.config.Height {
			style := s.getColorStyle(displayData[x])
			grid[y][x] = style.Render("●")
		}
	}

	// Render grid
	for _, row := range grid {
		b.WriteString(strings.Join(row, ""))
		b.WriteString("\n")
	}

	if s.config.ShowMinMax {
		b.WriteString(s.renderFooter())
	}

	return strings.TrimSuffix(b.String(), "\n")
}

// filledView renders filled area under the line
func (s *Sparkline) filledView() string {
	var b strings.Builder

	if s.config.Title != "" {
		b.WriteString(s.config.Title + "\n")
	}

	if s.config.ShowValue || s.config.ShowMinMax {
		b.WriteString(s.renderHeader())
		b.WriteString("\n")
	}

	displayData := s.prepareDisplayData()

	// Create a grid
	grid := make([][]string, s.config.Height)
	for i := range grid {
		grid[i] = make([]string, len(displayData))
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Fill area under each point
	for x, value := range displayData {
		if !math.IsNaN(value) {
			height := s.normalize(value, 0, s.config.Height-1)
			style := s.getColorStyle(value)

			for y := s.config.Height - 1; y >= s.config.Height-1-height; y-- {
				if y >= 0 && y < s.config.Height {
					if y == s.config.Height-1-height {
						grid[y][x] = style.Render("▀")
					} else {
						grid[y][x] = style.Render("█")
					}
				}
			}
		}
	}

	// Render grid
	for _, row := range grid {
		b.WriteString(strings.Join(row, ""))
		b.WriteString("\n")
	}

	if s.config.ShowMinMax {
		b.WriteString(s.renderFooter())
	}

	return strings.TrimSuffix(b.String(), "\n")
}

// prepareDisplayData adjusts data for the display width
func (s *Sparkline) prepareDisplayData() []float64 {
	if len(s.data) <= s.config.Width {
		// Pad with NaN if needed
		result := make([]float64, s.config.Width)
		for i := range result {
			result[i] = math.NaN()
		}

		startIdx := s.config.Width - len(s.data)
		copy(result[startIdx:], s.data)
		return result
	}

	// Sample data if we have more points than width
	step := float64(len(s.data)) / float64(s.config.Width)
	result := make([]float64, s.config.Width)

	for i := 0; i < s.config.Width; i++ {
		idx := int(float64(i) * step)
		if idx >= len(s.data) {
			idx = len(s.data) - 1
		}
		result[i] = s.data[idx]
	}

	return result
}

// renderHeader renders title and current value
func (s *Sparkline) renderHeader() string {
	var parts []string

	if s.config.ShowValue && len(s.data) > 0 {
		value := s.data[len(s.data)-1]
		parts = append(parts, fmt.Sprintf("Current: %.2f", value))
	}

	if s.config.ShowMinMax && len(s.data) > 0 {
		parts = append(parts, fmt.Sprintf("Max: %.2f", s.max))
	}

	return strings.Join(parts, " | ")
}

// renderFooter renders min value
func (s *Sparkline) renderFooter() string {
	if len(s.data) == 0 {
		return ""
	}
	return fmt.Sprintf("Min: %.2f", s.min)
}

// SparklineTickMsg is sent to update sparklines with new data
type SparklineTickMsg struct {
	ID    string
	Value float64
	Time  time.Time
}

// SparklineUpdateCmd sends a tick message for sparkline updates
func SparklineUpdateCmd(id string, value float64) tea.Cmd {
	return func() tea.Msg {
		return SparklineTickMsg{
			ID:    id,
			Value: value,
			Time:  time.Now(),
		}
	}
}
