package sparkline

import (
	"fmt"
	"math"
	"testing"
)

func TestSparklineDebug(t *testing.T) {
	config := Config{
		Width:     30,
		Height:    1,
		MaxPoints: 30,
		Style:     StyleBars,
	}

	s := New(config)
	testData := []float64{1, 5, 10, 20, 35, 50}
	s.SetData(testData)

	// Test prepareDisplayData directly
	displayData := s.prepareDisplayData()

	fmt.Printf("=== Debug Analysis ===\n")
	fmt.Printf("Config width: %d\n", config.Width)
	fmt.Printf("Display data length: %d\n", len(displayData))

	// Count NaN vs real values
	nanCount := 0
	realCount := 0
	for i, v := range displayData {
		if math.IsNaN(v) {
			nanCount++
		} else {
			realCount++
			fmt.Printf("Real data[%d] = %f\n", i, v)
		}
	}

	fmt.Printf("NaN count: %d\n", nanCount)
	fmt.Printf("Real count: %d\n", realCount)

	// Now let's manually build what barsView should produce
	bars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	var result string
	for _, value := range displayData {
		if math.IsNaN(value) {
			result += " "
		} else {
			barHeight := s.normalize(value, 0, len(bars)-1)
			result += bars[barHeight]
		}
	}

	fmt.Printf("Manual result: %q\n", result)
	fmt.Printf("Manual length: %d\n", len(result))

	// Compare with actual render
	actual := s.Render()
	fmt.Printf("Actual render: %q\n", actual)
	fmt.Printf("Actual length: %d\n", len(actual))
}
