// Package sparkline provides a flexible, feature-rich sparkline component for terminal applications.
//
// Sparklines are small, word-sized charts that can be embedded in text or used in terminal UIs.
// This package supports multiple visual styles, customizable dimensions, color ranges, and
// integrates seamlessly with Bubble Tea applications.
//
// Key features:
//   - Multiple visual styles: bars, dots, lines, filled areas
//   - Customizable colors with value-based ranges
//   - Configurable dimensions and data capacity
//   - Real-time data updates with sliding window behavior
//   - Bubble Tea integration for interactive applications
//   - Memory-efficient with bounded data storage
//
// Example usage:
//
//	config := sparkline.Config{
//	    Width:      50,
//	    Height:     6,
//	    MaxPoints:  100,
//	    Style:      sparkline.StyleBars,
//	    Title:      "CPU Usage (%)",
//	    ShowValue:  true,
//	    ShowMinMax: true,
//	}
//
//	s := sparkline.New(config)
//	s.AddPoint(42.5)
//	fmt.Println(s.Render())
package sparkline

import (
	"fmt"
	"math"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Style defines the visual representation of the sparkline
type Style int

const (
	StyleBars Style = iota
	StyleDots
	StyleLine
	StyleFilled
)

// ColorRange defines a value range and its associated styling
type ColorRange struct {
	Min   float64
	Max   float64
	Style lipgloss.Style
}

// Config holds the configuration for a sparkline
type Config struct {
	Width        int           // Display width in characters
	Height       int           // Display height in characters  
	MaxPoints    int           // Maximum number of data points to keep in memory
	Style        Style         // Visual style (bars, dots, line, filled)
	Title        string        // Optional title displayed above the sparkline
	ShowValue    bool          // Whether to show the current (last) value
	ShowMinMax   bool          // Whether to show min/max values
	ColorRanges  []ColorRange  // Value-based color ranges
	DefaultStyle lipgloss.Style // Default style for values not in color ranges
}

// Sparkline represents a sparkline chart with its data and configuration
type Sparkline struct {
	data   []float64
	min    float64
	max    float64
	config Config
}

// New creates a new sparkline with the given configuration
func New(config Config) *Sparkline {
	// Set default values for required fields
	if config.Width <= 0 {
		config.Width = 20
	}
	if config.Height <= 0 {
		config.Height = 4
	}
	if config.MaxPoints <= 0 {
		config.MaxPoints = config.Width
	}

	return &Sparkline{
		data:   make([]float64, 0, config.MaxPoints),
		min:    math.Inf(1),
		max:    math.Inf(-1),
		config: config,
	}
}

// Render returns the string representation of the sparkline
func (s *Sparkline) Render() string {
	return s.View()
}

// AddPoint adds a new data point to the sparkline.
//
// If the number of points exceeds MaxPoints, the oldest point is removed (FIFO).
// The min and max values are automatically updated.
// Invalid values (NaN, Inf) are rejected to prevent rendering issues.
func (s *Sparkline) AddPoint(value float64) {
	// Reject invalid values that could cause rendering issues
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return
	}

	s.data = append(s.data, value)

	// Maintain rolling window
	if len(s.data) > s.config.MaxPoints {
		s.data = s.data[1:]
	}

	// Update min/max
	s.updateMinMax()
}

// AddPoints adds multiple data points at once.
// This is more efficient than calling AddPoint multiple times.
func (s *Sparkline) AddPoints(values []float64) {
	for _, v := range values {
		s.AddPoint(v)
	}
}

// SetData replaces all data with new values.
// This clears the existing data and recalculates min/max values.
func (s *Sparkline) SetData(data []float64) {
	s.data = make([]float64, 0, s.config.MaxPoints)
	s.min = math.Inf(1)
	s.max = math.Inf(-1)
	s.AddPoints(data)
}

// Clear removes all data points and resets min/max values.
func (s *Sparkline) Clear() {
	s.data = s.data[:0]
	s.min = math.Inf(1)
	s.max = math.Inf(-1)
}

// GetData returns a copy of the current data points.
// The returned slice is safe to modify without affecting the sparkline.
func (s *Sparkline) GetData() []float64 {
	result := make([]float64, len(s.data))
	copy(result, s.data)
	return result
}

// GetLastValue returns the most recent data point, or 0 if no data exists.
func (s *Sparkline) GetLastValue() float64 {
	if len(s.data) == 0 {
		return 0
	}
	return s.data[len(s.data)-1]
}

// GetMinMax returns the current minimum and maximum values.
// Returns (0, 0) if no data exists.
func (s *Sparkline) GetMinMax() (float64, float64) {
	if len(s.data) == 0 {
		return 0, 0
	}
	return s.min, s.max
}

// GetConfig returns a copy of the current configuration.
func (s *Sparkline) GetConfig() Config {
	return s.config
}

// UpdateConfig updates the sparkline configuration.
// Data is preserved, but the display may change based on new settings.
func (s *Sparkline) UpdateConfig(config Config) {
	if config.Width <= 0 {
		config.Width = s.config.Width
	}
	if config.Height <= 0 {
		config.Height = s.config.Height
	}
	if config.MaxPoints <= 0 {
		config.MaxPoints = config.Width
	}

	s.config = config

	// Trim data if new MaxPoints is smaller
	if len(s.data) > s.config.MaxPoints {
		s.data = s.data[len(s.data)-s.config.MaxPoints:]
		s.updateMinMax()
	}
}

// Bubble Tea interface implementation

// Init implements tea.Model for Bubble Tea integration.
func (s *Sparkline) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model for Bubble Tea integration.
func (s *Sparkline) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return s, nil
}

// View implements tea.Model for Bubble Tea integration.
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

// Internal methods

func (s *Sparkline) updateMinMax() {
	if len(s.data) == 0 {
		s.min = math.Inf(1)
		s.max = math.Inf(-1)
		return
	}

	s.min = s.data[0]
	s.max = s.data[0]

	for _, v := range s.data {
		// Skip invalid values when calculating min/max
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if v < s.min {
			s.min = v
		}
		if v > s.max {
			s.max = v
		}
	}

	// If all values were invalid, reset to safe defaults
	if math.IsInf(s.min, 1) && math.IsInf(s.max, -1) {
		s.min = 0
		s.max = 1
	}
}

func (s *Sparkline) normalize(value float64, targetMin, targetMax int) int {
	// Handle edge cases that could cause invalid indices
	if s.max == s.min || math.IsNaN(value) || math.IsInf(value, 0) {
		return targetMin
	}

	// Ensure value is within bounds
	if value < s.min {
		value = s.min
	}
	if value > s.max {
		value = s.max
	}

	// Normalize to target range
	normalized := (value - s.min) / (s.max - s.min)
	result := targetMin + int(normalized*float64(targetMax-targetMin))

	// Clamp to target range to prevent array bounds issues
	if result < targetMin {
		result = targetMin
	}
	if result > targetMax {
		result = targetMax
	}

	return result
}

func (s *Sparkline) getColorStyle(value float64) lipgloss.Style {
	for _, cr := range s.config.ColorRanges {
		if value >= cr.Min && value < cr.Max {
			return cr.Style
		}
	}
	return s.config.DefaultStyle
}

func (s *Sparkline) emptyView() string {
	var b strings.Builder

	if s.config.Title != "" {
		b.WriteString(s.config.Title + "\n")
	}

	if s.config.ShowValue || s.config.ShowMinMax {
		b.WriteString("No data\n")
	}

	// Show empty chart area
	for i := 0; i < s.config.Height; i++ {
		b.WriteString(strings.Repeat("·", s.config.Width))
		if i < s.config.Height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

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

	// Place dots based on values
	for col, value := range displayData {
		if !math.IsNaN(value) {
			row := s.config.Height - 1 - s.normalize(value, 0, s.config.Height-1)
			style := s.getColorStyle(value)
			grid[row][col] = style.Render("●")
		}
	}

	// Render grid from top to bottom
	for i := 0; i < s.config.Height; i++ {
		b.WriteString(strings.Join(grid[i], ""))
		if i < s.config.Height-1 {
			b.WriteString("\n")
		}
	}

	if s.config.ShowMinMax {
		b.WriteString("\n")
		b.WriteString(s.renderFooter())
	}

	return b.String()
}

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

	// Create a grid for line
	grid := make([][]string, s.config.Height)
	for i := range grid {
		grid[i] = make([]string, len(displayData))
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Draw line connecting points
	for col, value := range displayData {
		if !math.IsNaN(value) {
			row := s.config.Height - 1 - s.normalize(value, 0, s.config.Height-1)
			style := s.getColorStyle(value)

			// Use different characters for line connections
			if col == 0 || math.IsNaN(displayData[col-1]) {
				grid[row][col] = style.Render("●") // Start point
			} else {
				prevValue := displayData[col-1]
				prevRow := s.config.Height - 1 - s.normalize(prevValue, 0, s.config.Height-1)

				if prevRow == row {
					grid[row][col] = style.Render("─") // Horizontal line
				} else if prevRow < row {
					grid[row][col] = style.Render("╱") // Up slope
				} else {
					grid[row][col] = style.Render("╲") // Down slope
				}
			}
		}
	}

	// Render grid from top to bottom
	for i := 0; i < s.config.Height; i++ {
		b.WriteString(strings.Join(grid[i], ""))
		if i < s.config.Height-1 {
			b.WriteString("\n")
		}
	}

	if s.config.ShowMinMax {
		b.WriteString("\n")
		b.WriteString(s.renderFooter())
	}

	return b.String()
}

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

	// Create a grid for filled area
	grid := make([][]string, s.config.Height)
	for i := range grid {
		grid[i] = make([]string, len(displayData))
		for j := range grid[i] {
			grid[i][j] = " "
		}
	}

	// Fill area under the curve
	for col, value := range displayData {
		if !math.IsNaN(value) {
			height := s.normalize(value, 0, s.config.Height-1)
			style := s.getColorStyle(value)

			// Fill from bottom up to the value height
			for row := s.config.Height - 1; row >= s.config.Height-1-height; row-- {
				if row == s.config.Height-1-height {
					grid[row][col] = style.Render("▀") // Top edge
				} else {
					grid[row][col] = style.Render("█") // Fill
				}
			}
		}
	}

	// Render grid from top to bottom
	for i := 0; i < s.config.Height; i++ {
		b.WriteString(strings.Join(grid[i], ""))
		if i < s.config.Height-1 {
			b.WriteString("\n")
		}
	}

	if s.config.ShowMinMax {
		b.WriteString("\n")
		b.WriteString(s.renderFooter())
	}

	return b.String()
}

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

	// Take the most recent Width points (sliding window behavior)
	startIdx := len(s.data) - s.config.Width
	result := make([]float64, s.config.Width)
	copy(result, s.data[startIdx:])
	return result
}

func (s *Sparkline) renderHeader() string {
	var parts []string

	if s.config.ShowValue && len(s.data) > 0 {
		value := s.data[len(s.data)-1]
		parts = append(parts, fmt.Sprintf("Current: %.2f", value))
	}

	if s.config.ShowMinMax && len(s.data) > 0 {
		parts = append(parts, fmt.Sprintf("Range: %.2f - %.2f", s.min, s.max))
	}

	return strings.Join(parts, " | ")
}

func (s *Sparkline) renderFooter() string {
	if s.config.ShowMinMax && len(s.data) > 0 {
		return fmt.Sprintf("Min: %.2f, Max: %.2f", s.min, s.max)
	}
	return ""
}
