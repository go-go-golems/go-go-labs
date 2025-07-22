package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/pkg/sparkline"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

var (
	// Color styles for different value ranges
	lowStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	medStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow
	highStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")) // Red
	defaultStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // White

	// Layout styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("6")).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			MarginTop(1)

	containerStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("240")).
			Padding(1).
			Margin(1)
)

// DataGenerator generates different types of data patterns
type DataGenerator interface {
	Next() float64
	Name() string
}

// RandomGenerator generates random data
type RandomGenerator struct {
	min, max float64
}

func (r *RandomGenerator) Next() float64 {
	return r.min + rand.Float64()*(r.max-r.min)
}

func (r *RandomGenerator) Name() string {
	return "Random"
}

// SineWaveGenerator generates sine wave data
type SineWaveGenerator struct {
	amplitude, frequency, phase float64
	time                        float64
}

func (s *SineWaveGenerator) Next() float64 {
	value := s.amplitude * math.Sin(2*math.Pi*s.frequency*s.time+s.phase)
	s.time += 0.1
	return value
}

func (s *SineWaveGenerator) Name() string {
	return "Sine Wave"
}

// TrendGenerator generates trending data
type TrendGenerator struct {
	base, trend, noise float64
	time               float64
}

func (t *TrendGenerator) Next() float64 {
	value := t.base + t.trend*t.time + t.noise*(rand.Float64()-0.5)
	t.time += 0.1
	return value
}

func (t *TrendGenerator) Name() string {
	return "Trending"
}

// SpikeGenerator generates data with occasional spikes
type SpikeGenerator struct {
	base, spikeProb, spikeHeight float64
}

func (s *SpikeGenerator) Next() float64 {
	if rand.Float64() < s.spikeProb {
		return s.base + s.spikeHeight*(rand.Float64()-0.5)
	}
	return s.base + (rand.Float64()-0.5)*0.1
}

func (s *SpikeGenerator) Name() string {
	return "Spiky"
}

// SparklineDemo holds a sparkline and its data generator
type SparklineDemo struct {
	sparkline *sparkline.Sparkline
	generator DataGenerator
	id        string
}

// Model represents the main application model
type Model struct {
	sparklines    []SparklineDemo
	currentStyle  sparkline.Style
	styleNames    []string
	paused        bool
	speed         time.Duration
	width, height int
}

// TickMsg is sent periodically to update data
type TickMsg time.Time

// tickCmd returns a command that sends TickMsg after a delay
func (m Model) tickCmd() tea.Cmd {
	return tea.Tick(m.speed, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func initialModel() Model {
	width := 50
	height := 8

	// Color ranges for different values
	colorRanges := []sparkline.ColorRange{
		{Min: -math.Inf(1), Max: -0.5, Style: lowStyle},
		{Min: -0.5, Max: 0.5, Style: medStyle},
		{Min: 0.5, Max: math.Inf(1), Style: highStyle},
	}

	// Create different sparkline configurations
	configs := []struct {
		title     string
		generator DataGenerator
		style     sparkline.Style
	}{
		{
			title:     "CPU Usage (%)",
			generator: &RandomGenerator{min: 0, max: 100},
			style:     sparkline.StyleBars,
		},
		{
			title:     "Memory Usage (GB)",
			generator: &SineWaveGenerator{amplitude: 4, frequency: 0.05, phase: 0},
			style:     sparkline.StyleDots,
		},
		{
			title:     "Network I/O (MB/s)",
			generator: &TrendGenerator{base: 10, trend: 0.1, noise: 2},
			style:     sparkline.StyleLine,
		},
		{
			title:     "Disk Activity",
			generator: &SpikeGenerator{base: 20, spikeProb: 0.1, spikeHeight: 50},
			style:     sparkline.StyleFilled,
		},
	}

	sparklines := make([]SparklineDemo, len(configs))
	for i, config := range configs {
		sparklineConfig := sparkline.Config{
			Width:        width,
			Height:       height,
			MaxPoints:    width * 2, // Keep more history than display width
			Style:        config.style,
			Title:        config.title,
			ShowValue:    true,
			ShowMinMax:   true,
			ColorRanges:  colorRanges,
			DefaultStyle: defaultStyle,
		}

		sparklines[i] = SparklineDemo{
			sparkline: sparkline.New(sparklineConfig),
			generator: config.generator,
			id:        fmt.Sprintf("sparkline_%d", i),
		}

		// Add some initial data
		for j := 0; j < 20; j++ {
			sparklines[i].sparkline.AddPoint(config.generator.Next())
		}
	}

	return Model{
		sparklines:   sparklines,
		currentStyle: sparkline.StyleBars,
		styleNames:   []string{"Bars", "Dots", "Line", "Filled"},
		paused:       false,
		speed:        200 * time.Millisecond,
		width:        width,
		height:       height,
	}
}

func (m Model) Init() tea.Cmd {
	return m.tickCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case " ":
			m.paused = !m.paused
		case "s":
			// Switch style for all sparklines
			m.currentStyle = (m.currentStyle + 1) % 4

			// Titles for each sparkline
			titles := []string{
				"CPU Usage (%)",
				"Memory Usage (GB)",
				"Network I/O (MB/s)",
				"Disk Activity",
			}

			for i := range m.sparklines {
				// Create new sparkline with different style
				config := sparkline.Config{
					Width:      m.width,
					Height:     m.height,
					MaxPoints:  m.width * 2,
					Style:      m.currentStyle,
					Title:      titles[i],
					ShowValue:  true,
					ShowMinMax: true,
					ColorRanges: []sparkline.ColorRange{
						{Min: -math.Inf(1), Max: -0.5, Style: lowStyle},
						{Min: -0.5, Max: 0.5, Style: medStyle},
						{Min: 0.5, Max: math.Inf(1), Style: highStyle},
					},
					DefaultStyle: defaultStyle,
				}

				// Preserve data
				oldData := m.sparklines[i].sparkline.GetData()
				m.sparklines[i].sparkline = sparkline.New(config)
				m.sparklines[i].sparkline.SetData(oldData)
			}
		case "r":
			// Reset all data
			for i := range m.sparklines {
				m.sparklines[i].sparkline.Clear()
			}
		case "1":
			m.speed = 50 * time.Millisecond
		case "2":
			m.speed = 200 * time.Millisecond
		case "3":
			m.speed = 500 * time.Millisecond
		case "4":
			m.speed = 1000 * time.Millisecond
		}

	case TickMsg:
		// Update data if not paused
		if !m.paused {
			for i := range m.sparklines {
				value := m.sparklines[i].generator.Next()
				m.sparklines[i].sparkline.AddPoint(value)
			}
		}
		return m, m.tickCmd()
	}

	return m, nil
}

func (m Model) View() string {
	var content strings.Builder

	// Title
	content.WriteString(titleStyle.Render("🌟 Sparkline Component Demo"))
	content.WriteString("\n\n")

	// Current settings
	status := fmt.Sprintf("Style: %s | Speed: %v | Status: %s",
		m.styleNames[m.currentStyle],
		m.speed,
		map[bool]string{true: "⏸️  Paused", false: "▶️  Running"}[m.paused])
	content.WriteString(status)
	content.WriteString("\n\n")

	// Render sparklines in a 2x2 grid
	row1 := lipgloss.JoinHorizontal(lipgloss.Top,
		containerStyle.Render(m.sparklines[0].sparkline.View()),
		containerStyle.Render(m.sparklines[1].sparkline.View()),
	)

	row2 := lipgloss.JoinHorizontal(lipgloss.Top,
		containerStyle.Render(m.sparklines[2].sparkline.View()),
		containerStyle.Render(m.sparklines[3].sparkline.View()),
	)

	content.WriteString(lipgloss.JoinVertical(lipgloss.Left, row1, row2))
	content.WriteString("\n")

	// Help text
	help := `Controls:
  SPACE    Toggle pause/resume
  s        Switch sparkline style (bars → dots → line → filled)
  r        Reset all data
  1-4      Change update speed (1=fast, 4=slow)
  q        Quit`

	content.WriteString(helpStyle.Render(help))

	return content.String()
}

func main() {
	var logLevel string

	rootCmd := &cobra.Command{
		Use:   "sparkline-test",
		Short: "Test application for the sparkline TUI component",
		Long: `A demonstration application showing various sparkline styles and data patterns.
The application shows real-time data visualization using different sparkline styles
including bars, dots, lines, and filled areas.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Setup logging
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				return fmt.Errorf("invalid log level: %w", err)
			}
			zerolog.SetGlobalLevel(level)

			// Initialize random seed
			rand.Seed(time.Now().UnixNano())

			// Run the TUI
			p := tea.NewProgram(initialModel(), tea.WithAltScreen())
			_, err = p.Run()
			return err
		},
	}

	demoCmd := &cobra.Command{
		Use:   "demo",
		Short: "Run a non-interactive demo of the sparkline component",
		Long:  "Demonstrates the sparkline component capabilities without requiring a TTY",
		RunE: func(cmd *cobra.Command, args []string) error {
			DemoSparkline()
			return nil
		},
	}

	rootCmd.AddCommand(demoCmd)
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Set the log level (debug, info, warn, error)")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
