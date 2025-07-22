package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/sparkline"
)

// DemoSparkline demonstrates the sparkline component without TUI
func DemoSparkline() {
	fmt.Println("🌟 Sparkline Component Demo")
	fmt.Println(strings.Repeat("=", 50))

	// Create color styles
	lowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))     // Green
	medStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))     // Yellow
	highStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))    // Red
	defaultStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // White

	colorRanges := []sparkline.ColorRange{
		{Min: -math.Inf(1), Max: 30, Style: lowStyle},
		{Min: 30, Max: 70, Style: medStyle},
		{Min: 70, Max: math.Inf(1), Style: highStyle},
	}

	// Demo different styles
	styles := []struct {
		name  string
		style sparkline.Style
	}{
		{"Bars", sparkline.StyleBars},
		{"Dots", sparkline.StyleDots},
		{"Line", sparkline.StyleLine},
		{"Filled", sparkline.StyleFilled},
	}

	// Generate sample data patterns
	patterns := []struct {
		name string
		data []float64
	}{
		{
			name: "Random CPU Usage",
			data: generateRandomData(40, 0, 100),
		},
		{
			name: "Sine Wave",
			data: generateSineWave(40, 50, 30, 0.3),
		},
		{
			name: "Trending Up",
			data: generateTrend(40, 20, 1.5, 5),
		},
		{
			name: "Spiky Data",
			data: generateSpikes(40, 25, 0.15, 50),
		},
	}

	// Demonstrate each style with different data patterns
	for _, pattern := range patterns {
		fmt.Printf("\n📊 %s\n", pattern.name)
		fmt.Println(strings.Repeat("-", 30))

		for _, styleInfo := range styles {
			config := sparkline.Config{
				Width:        50,
				Height:       6,
				MaxPoints:    50,
				Style:        styleInfo.style,
				Title:        fmt.Sprintf("%s (%s)", pattern.name, styleInfo.name),
				ShowValue:    true,
				ShowMinMax:   true,
				ColorRanges:  colorRanges,
				DefaultStyle: defaultStyle,
			}

			s := sparkline.New(config)
			s.SetData(pattern.data)

			fmt.Printf("\n%s:\n", styleInfo.name)
			fmt.Println(s.View())
		}
	}

	// Demonstrate real-time updates
	fmt.Println("\n🔄 Real-time Update Demo")
	fmt.Println(strings.Repeat("-", 30))

	config := sparkline.Config{
		Width:        30,
		Height:       5,
		MaxPoints:    30,
		Style:        sparkline.StyleBars,
		Title:        "Live CPU Usage",
		ShowValue:    true,
		ShowMinMax:   true,
		ColorRanges:  colorRanges,
		DefaultStyle: defaultStyle,
	}

	s := sparkline.New(config)

	// Simulate adding data points over time
	fmt.Println("Adding data points (simulated real-time):")
	for i := 0; i < 15; i++ {
		value := 20 + 50*math.Sin(float64(i)*0.3) + rand.Float64()*10
		s.AddPoint(value)

		fmt.Printf("\nStep %d (Value: %.1f):\n", i+1, value)
		fmt.Println(s.View())

		// In real app, this would be time.Sleep
		if i%5 == 4 {
			fmt.Println("\n[Simulated pause...]")
		}
	}

	// Demonstrate configuration changes
	fmt.Println("\n⚙️  Configuration Demo")
	fmt.Println(strings.Repeat("-", 30))

	testData := generateSineWave(25, 50, 25, 0.4)

	configurations := []struct {
		name   string
		config sparkline.Config
	}{
		{
			name: "Compact (20x3)",
			config: sparkline.Config{
				Width:        20,
				Height:       3,
				MaxPoints:    20,
				Style:        sparkline.StyleBars,
				Title:        "Compact View",
				ShowValue:    false,
				ShowMinMax:   false,
				DefaultStyle: defaultStyle,
			},
		},
		{
			name: "Detailed (60x8)",
			config: sparkline.Config{
				Width:        60,
				Height:       8,
				MaxPoints:    60,
				Style:        sparkline.StyleLine,
				Title:        "Detailed View",
				ShowValue:    true,
				ShowMinMax:   true,
				ColorRanges:  colorRanges,
				DefaultStyle: defaultStyle,
			},
		},
		{
			name: "No Title/Values",
			config: sparkline.Config{
				Width:        40,
				Height:       4,
				MaxPoints:    40,
				Style:        sparkline.StyleFilled,
				Title:        "",
				ShowValue:    false,
				ShowMinMax:   false,
				DefaultStyle: defaultStyle,
			},
		},
	}

	for _, cfg := range configurations {
		fmt.Printf("\n%s:\n", cfg.name)
		s := sparkline.New(cfg.config)
		s.SetData(testData)
		fmt.Println(s.View())
	}

	fmt.Println("\n✅ Demo completed!")
	fmt.Println("\nTo run the interactive TUI version:")
	fmt.Println("  go run ./cmd/apps/sparkline-test")
}

// Helper functions to generate test data
func generateRandomData(count int, min, max float64) []float64 {
	data := make([]float64, count)
	for i := range data {
		data[i] = min + rand.Float64()*(max-min)
	}
	return data
}

func generateSineWave(count int, center, amplitude, frequency float64) []float64 {
	data := make([]float64, count)
	for i := range data {
		data[i] = center + amplitude*math.Sin(2*math.Pi*frequency*float64(i)/float64(count))
	}
	return data
}

func generateTrend(count int, start, slope, noise float64) []float64 {
	data := make([]float64, count)
	for i := range data {
		data[i] = start + slope*float64(i) + (rand.Float64()-0.5)*noise
	}
	return data
}

func generateSpikes(count int, base float64, spikeProb, spikeHeight float64) []float64 {
	data := make([]float64, count)
	for i := range data {
		if rand.Float64() < spikeProb {
			data[i] = base + spikeHeight*(rand.Float64()-0.5)
		} else {
			data[i] = base + (rand.Float64()-0.5)*5
		}
	}
	return data
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
