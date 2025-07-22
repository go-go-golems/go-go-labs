package components

import (
	"math"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestNewSparkline(t *testing.T) {
	config := SparklineConfig{
		Width:     40,
		Height:    8,
		MaxPoints: 50,
		Style:     StyleBars,
		Title:     "Test Sparkline",
	}

	sparkline := NewSparkline(config)

	assert.NotNil(t, sparkline)
	assert.Equal(t, config.Width, sparkline.config.Width)
	assert.Equal(t, config.Height, sparkline.config.Height)
	assert.Equal(t, config.MaxPoints, sparkline.config.MaxPoints)
	assert.Equal(t, config.Style, sparkline.config.Style)
	assert.Equal(t, config.Title, sparkline.config.Title)
	assert.Empty(t, sparkline.data)
	assert.True(t, math.IsInf(sparkline.min, 1))
	assert.True(t, math.IsInf(sparkline.max, -1))
}

func TestSparklineDefaults(t *testing.T) {
	// Test with minimal config
	config := SparklineConfig{}
	sparkline := NewSparkline(config)

	assert.Equal(t, 40, sparkline.config.Width)
	assert.Equal(t, 8, sparkline.config.Height)
	assert.Equal(t, 40, sparkline.config.MaxPoints) // Should default to width
}

func TestAddPoint(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 5,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	// Add some data points
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for _, v := range values {
		sparkline.AddPoint(v)
	}

	data := sparkline.GetData()
	assert.Equal(t, values, data)
	assert.Equal(t, 1.0, sparkline.min)
	assert.Equal(t, 5.0, sparkline.max)
	assert.Equal(t, 5.0, sparkline.GetLastValue())
}

func TestRollingWindow(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 3, // Small window for testing
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	// Add more points than the window size
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for _, v := range values {
		sparkline.AddPoint(v)
	}

	data := sparkline.GetData()
	expected := []float64{3.0, 4.0, 5.0} // Should keep only last 3
	assert.Equal(t, expected, data)
	assert.Equal(t, 3.0, sparkline.min)
	assert.Equal(t, 5.0, sparkline.max)
}

func TestAddPoints(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	sparkline.AddPoints(values)

	data := sparkline.GetData()
	assert.Equal(t, values, data)
	assert.Equal(t, 1.0, sparkline.min)
	assert.Equal(t, 5.0, sparkline.max)
}

func TestSetData(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	// Add some initial data
	sparkline.AddPoint(100.0)
	assert.Equal(t, 1, len(sparkline.GetData()))

	// Replace with new data
	newData := []float64{1.0, 2.0, 3.0}
	sparkline.SetData(newData)

	data := sparkline.GetData()
	assert.Equal(t, newData, data)
	assert.Equal(t, 1.0, sparkline.min)
	assert.Equal(t, 3.0, sparkline.max)
}

func TestClear(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	// Add some data
	sparkline.AddPoints([]float64{1.0, 2.0, 3.0})
	assert.Equal(t, 3, len(sparkline.GetData()))

	// Clear data
	sparkline.Clear()

	data := sparkline.GetData()
	assert.Empty(t, data)
	assert.True(t, math.IsInf(sparkline.min, 1))
	assert.True(t, math.IsInf(sparkline.max, -1))
	assert.Equal(t, 0.0, sparkline.GetLastValue())
}

func TestGetMinMax(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	// No data case
	min, max := sparkline.GetMinMax()
	assert.Equal(t, 0.0, min)
	assert.Equal(t, 0.0, max)

	// With data
	sparkline.AddPoints([]float64{-5.0, 0.0, 10.0, 3.0})
	min, max = sparkline.GetMinMax()
	assert.Equal(t, -5.0, min)
	assert.Equal(t, 10.0, max)
}

func TestNormalize(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)
	sparkline.AddPoints([]float64{0.0, 50.0, 100.0})

	// Test normalization
	assert.Equal(t, 0, sparkline.normalize(0.0, 0, 7))   // Min value maps to 0
	assert.Equal(t, 7, sparkline.normalize(100.0, 0, 7)) // Max value maps to 7
	assert.Equal(t, 4, sparkline.normalize(50.0, 0, 7))  // Mid value maps to middle
}

func TestNormalizeIdenticalValues(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)
	sparkline.AddPoints([]float64{50.0, 50.0, 50.0})

	// When min == max, should return targetMin
	assert.Equal(t, 0, sparkline.normalize(50.0, 0, 7))
}

func TestEmptyView(t *testing.T) {
	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
		Title:     "Test",
	}

	sparkline := NewSparkline(config)
	view := sparkline.View()

	assert.Contains(t, view, "Test")
	assert.Contains(t, view, "No data")
}

func TestViewWithData(t *testing.T) {
	config := SparklineConfig{
		Width:      10,
		Height:     4,
		MaxPoints:  10,
		Style:      StyleBars,
		Title:      "Test",
		ShowValue:  true,
		ShowMinMax: true,
	}

	sparkline := NewSparkline(config)
	sparkline.AddPoints([]float64{1.0, 2.0, 3.0, 4.0, 5.0})

	view := sparkline.View()

	assert.Contains(t, view, "Test")
	assert.Contains(t, view, "Current: 5.00")
	assert.Contains(t, view, "Max: 5.00")
	assert.Contains(t, view, "Min: 1.00")
}

func TestDifferentStyles(t *testing.T) {
	styles := []SparklineStyle{
		StyleBars,
		StyleDots,
		StyleLine,
		StyleFilled,
	}

	for _, style := range styles {
		config := SparklineConfig{
			Width:     20,
			Height:    6,
			MaxPoints: 20,
			Style:     style,
		}

		sparkline := NewSparkline(config)
		sparkline.AddPoints([]float64{1.0, 5.0, 3.0, 8.0, 2.0})

		view := sparkline.View()
		assert.NotEmpty(t, view, "Style %d should produce output", style)
	}
}

func TestColorRanges(t *testing.T) {
	lowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	highStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))

	config := SparklineConfig{
		Width:     10,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
		ColorRanges: []ColorRange{
			{Min: -math.Inf(1), Max: 49.9, Style: lowStyle},
			{Min: 50.0, Max: math.Inf(1), Style: highStyle},
		},
		DefaultStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("7")),
	}

	sparkline := NewSparkline(config)

	// Test color style selection
	lowStyleResult := sparkline.getColorStyle(25.0)
	assert.Equal(t, lowStyle, lowStyleResult)

	highStyleResult := sparkline.getColorStyle(75.0)
	assert.Equal(t, highStyle, highStyleResult)

	// Value exactly on boundary should match high range
	boundaryStyleResult := sparkline.getColorStyle(50.0)
	assert.Equal(t, highStyle, boundaryStyleResult)

	// Value between ranges should use default style
	defaultStyleResult := sparkline.getColorStyle(49.95)
	assert.Equal(t, config.DefaultStyle, defaultStyleResult)
}

func TestPrepareDisplayData(t *testing.T) {
	config := SparklineConfig{
		Width:     5,
		Height:    4,
		MaxPoints: 10,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	// Case 1: Less data than width (should pad with NaN)
	sparkline.AddPoints([]float64{1.0, 2.0, 3.0})
	displayData := sparkline.prepareDisplayData()

	assert.Equal(t, 5, len(displayData))
	assert.True(t, math.IsNaN(displayData[0]))
	assert.True(t, math.IsNaN(displayData[1]))
	assert.Equal(t, 1.0, displayData[2])
	assert.Equal(t, 2.0, displayData[3])
	assert.Equal(t, 3.0, displayData[4])

	// Case 2: More data than width (should sample)
	sparkline.SetData([]float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0})
	displayData = sparkline.prepareDisplayData()

	assert.Equal(t, 5, len(displayData))
	// Should sample from the data
	assert.False(t, math.IsNaN(displayData[0]))
	assert.False(t, math.IsNaN(displayData[4]))
}

func BenchmarkAddPoint(b *testing.B) {
	config := SparklineConfig{
		Width:     100,
		Height:    10,
		MaxPoints: 1000,
		Style:     StyleBars,
	}

	sparkline := NewSparkline(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sparkline.AddPoint(float64(i))
	}
}

func BenchmarkView(b *testing.B) {
	config := SparklineConfig{
		Width:      100,
		Height:     10,
		MaxPoints:  1000,
		Style:      StyleBars,
		ShowValue:  true,
		ShowMinMax: true,
	}

	sparkline := NewSparkline(config)

	// Add some test data
	for i := 0; i < 100; i++ {
		sparkline.AddPoint(float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sparkline.View()
	}
}
