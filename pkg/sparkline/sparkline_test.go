package sparkline

import (
	"fmt"
	"strings"
	"testing"
)

func TestSparklineWidth(t *testing.T) {
	// Test with width 30 as used in the table
	config := Config{
		Width:     30,
		Height:    1,
		MaxPoints: 30,
		Style:     StyleBars,
	}
	
	s := New(config)
	
	// Add some test data (6 points like in our test)
	testData := []float64{1, 5, 10, 20, 35, 50}
	s.SetData(testData)
	
	output := s.Render()
	
	fmt.Printf("=== Sparkline Width Test ===\n")
	fmt.Printf("Config width: %d\n", config.Width)
	fmt.Printf("Data points: %d\n", len(testData))
	fmt.Printf("Raw output: %q\n", output)
	fmt.Printf("Raw length: %d\n", len(output))
	
	trimmed := strings.TrimSpace(output)
	fmt.Printf("Trimmed: %q\n", trimmed)
	fmt.Printf("Trimmed length: %d\n", len(trimmed))
	
	// Check if the raw output is exactly the configured width
	if len(output) != config.Width {
		t.Errorf("Expected output length %d, got %d", config.Width, len(output))
	}
}
