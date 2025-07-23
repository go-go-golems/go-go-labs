# Sparkline Component

A feature-rich, standalone sparkline component for terminal user interfaces built with [Bubbletea](https://github.com/charmbracelet/bubbletea).

## Features

- 🎨 **Multiple Display Styles**: Bars, dots, lines, and filled areas
- 📊 **Real-time Updates**: Add data points dynamically with rolling window support
- 🎯 **Configurable Dimensions**: Customizable width and height
- 🌈 **Color Ranges**: Define different colors for value ranges
- 📈 **Statistics Display**: Show current value, min/max values
- ⚡ **Performance**: Efficient rendering and memory management
- 🔧 **Highly Configurable**: Extensive customization options

## Quick Start

### Basic Usage

```go
package main

import (
    "github.com/charmbracelet/lipgloss"
    "github.com/go-go-golems/go-go-labs/pkg/tui/components"
)

func main() {
    // Create a basic sparkline configuration
    config := components.SparklineConfig{
        Width:      40,
        Height:     8,
        MaxPoints:  50,
        Style:      components.StyleBars,
        Title:      "CPU Usage (%)",
        ShowValue:  true,
        ShowMinMax: true,
    }
    
    // Create the sparkline
    sparkline := components.NewSparkline(config)
    
    // Add some data
    for i := 0; i < 20; i++ {
        value := 50 + 30*math.Sin(float64(i)*0.3)
        sparkline.AddPoint(value)
    }
    
    // Render
    fmt.Println(sparkline.View())
}
```

### Bubbletea Integration

```go
type Model struct {
    sparkline *components.Sparkline
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case components.SparklineTickMsg:
        if msg.ID == "cpu" {
            m.sparkline.AddPoint(msg.Value)
        }
    }
    return m, nil
}

func (m Model) View() string {
    return m.sparkline.View()
}
```

## Configuration Options

### SparklineConfig

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `Width` | `int` | Display width in characters | 40 |
| `Height` | `int` | Display height in lines | 8 |
| `MaxPoints` | `int` | Maximum data points to keep | Width |
| `Style` | `SparklineStyle` | Visual style (bars, dots, line, filled) | StyleBars |
| `Title` | `string` | Title displayed above the sparkline | "" |
| `ShowValue` | `bool` | Show current value in header | false |
| `ShowMinMax` | `bool` | Show min/max values | false |
| `ColorRanges` | `[]ColorRange` | Color ranges for different values | nil |
| `DefaultStyle` | `lipgloss.Style` | Default color style | White |

### Display Styles

#### StyleBars
```
CPU Usage (%)
Current: 87.23 | Max: 95.67
▁▃▄▇▅▅▅▅▂▆█▇▅▂▂▃█▇▅▄▆▂█▁▄
Min: 12.34
```

#### StyleDots
```
Memory Usage (GB)
Current: 7.4 | Max: 8.1
●     ●   ●       ●
  ●     ●   ● ●     ●
    ● ●       ●   ●
      ●           ●
Min: 4.2
```

#### StyleLine
```
Network I/O
Current: 45.6 | Max: 78.9
    ╱───╲
   ╱     ╲   ╱──●
  ╱       ─╱
 ╱
╱
Min: 15.3
```

#### StyleFilled
```
Disk Activity
Current: 62.1 | Max: 89.4
    ▀▀▀▀▀
   ▀█████▀   ▀▀▀▀
  ▀███████▀▀▀████
 ▀███████████████
▀████████████████
Min: 8.7
```

### Color Ranges

Define color ranges to highlight different value thresholds:

```go
colorRanges := []components.ColorRange{
    {Min: -math.Inf(1), Max: 30, Style: greenStyle},   // Low values
    {Min: 30, Max: 70, Style: yellowStyle},            // Medium values  
    {Min: 70, Max: math.Inf(1), Style: redStyle},      // High values
}

config := components.SparklineConfig{
    // ... other config
    ColorRanges: colorRanges,
}
```

## API Reference

### Creating Sparklines

```go
// Create with configuration
sparkline := components.NewSparkline(config)

// Initialize for Bubbletea
cmd := sparkline.Init() // Returns nil

// Update (in Bubbletea Update method)
model, cmd := sparkline.Update(msg)

// Render
output := sparkline.View()
```

### Data Management

```go
// Add single data point
sparkline.AddPoint(42.5)

// Add multiple points
sparkline.AddPoints([]float64{1.0, 2.0, 3.0})

// Replace all data
sparkline.SetData([]float64{10, 20, 30, 40, 50})

// Clear all data
sparkline.Clear()

// Get current data (copy)
data := sparkline.GetData()

// Get latest value
latest := sparkline.GetLastValue()

// Get min/max values
min, max := sparkline.GetMinMax()
```

### Real-time Updates

```go
// Generate update commands
cmd := components.SparklineUpdateCmd("cpu", cpuUsage)

// Handle in your Update method
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case components.SparklineTickMsg:
        m.sparkline.AddPoint(msg.Value)
    }
    return m, nil
}
```

## Examples

### Monitoring Dashboard

```go
// Create multiple sparklines for different metrics
configs := map[string]components.SparklineConfig{
    "cpu": {
        Width: 50, Height: 6, Style: components.StyleBars,
        Title: "CPU Usage (%)", ShowValue: true, ShowMinMax: true,
    },
    "memory": {
        Width: 50, Height: 6, Style: components.StyleLine,
        Title: "Memory Usage (GB)", ShowValue: true, ShowMinMax: true,
    },
    "network": {
        Width: 50, Height: 6, Style: components.StyleFilled,
        Title: "Network I/O (MB/s)", ShowValue: true, ShowMinMax: true,
    },
}
```

### Custom Data Generators

```go
// Sine wave generator
type SineGenerator struct {
    amplitude, frequency, phase, time float64
}

func (s *SineGenerator) Next() float64 {
    value := s.amplitude * math.Sin(2*math.Pi*s.frequency*s.time + s.phase)
    s.time += 0.1
    return value
}

// Use with sparkline
generator := &SineGenerator{amplitude: 50, frequency: 0.1}
sparkline.AddPoint(generator.Next())
```

### Responsive Layout

```go
// Adjust size based on terminal dimensions
width, height, _ := term.GetSize(int(os.Stdout.Fd()))
sparklineWidth := (width - 10) / 2  // Two sparklines side by side

config.Width = sparklineWidth
config.Height = height / 4
```

## Test Application

Run the interactive demo to see all features:

```bash
# Interactive TUI demo
go run ./cmd/apps/sparkline-test

# Non-interactive demo (works without TTY)
go run ./cmd/apps/sparkline-test demo
```

### Demo Controls

- `SPACE` - Toggle pause/resume
- `s` - Switch sparkline style (bars → dots → line → filled)
- `r` - Reset all data
- `1-4` - Change update speed (1=fast, 4=slow)
- `q` - Quit

## Use Cases

- **System Monitoring**: CPU, memory, disk, network metrics
- **Application Metrics**: Response times, throughput, error rates
- **Financial Data**: Stock prices, trading volumes, portfolio values
- **IoT Dashboards**: Sensor readings, device status
- **Log Analysis**: Event rates, error frequencies
- **Development Tools**: Build times, test coverage, performance metrics

## Performance

- Efficient memory usage with configurable rolling windows
- Optimized rendering for terminal output
- Minimal CPU overhead for real-time updates
- Scales well with high-frequency data

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling and layout

## License

Part of the go-go-labs project.
