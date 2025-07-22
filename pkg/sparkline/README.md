# Sparkline - Terminal Data Visualization

A flexible, feature-rich sparkline component for Go terminal applications. Sparklines are small, word-sized charts that provide an at-a-glance view of data trends - perfect for dashboards, monitoring tools, and CLI applications.

## What are Sparklines?

Sparklines are miniature charts designed to be embedded in text or displayed in compact spaces. They show the general shape of data variation without axes, labels, or other chart junk. Originally popularized by Edward Tufte, they're perfect for showing trends, patterns, and data changes over time in a small footprint.

Think of them as "data thumbnails" - they give you the essence of your data's story in just a few characters.

## Features

- **Multiple Visual Styles**: Choose from bars, dots, lines, or filled area charts
- **Real-time Updates**: Perfect for live monitoring with smooth sliding window behavior  
- **Customizable Colors**: Define value-based color ranges for visual alerts
- **Memory Efficient**: Bounded data storage with configurable history limits
- **Bubble Tea Integration**: Works seamlessly with interactive terminal UIs
- **Flexible Configuration**: Adjust dimensions, styling, and display options

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/go-go-golems/go-go-labs/pkg/sparkline"
)

func main() {
    // Create a simple sparkline
    config := sparkline.Config{
        Width:  30,
        Height: 4,
        Style:  sparkline.StyleBars,
        Title:  "CPU Usage",
    }
    
    s := sparkline.New(config)
    
    // Add some data points
    data := []float64{45, 67, 23, 89, 56, 78, 34, 90, 12, 67}
    s.SetData(data)
    
    // Render the sparkline
    fmt.Println(s.Render())
}
```

This produces a compact bar chart showing your data trend:
```
CPU Usage
▃▅▂▇▄▆▃█▁▅▃▆▄▇▅▆▃▇▂▅▃▆▄▇▅▆▃▇▂▅
```

### Real-time Monitoring

For live data feeds, sparklines shine at showing recent trends:

```go
// Configure for real-time use
config := sparkline.Config{
    Width:     50,
    Height:    6,
    MaxPoints: 100,  // Keep last 100 data points
    Style:     sparkline.StyleBars,
    Title:     "Network Traffic (MB/s)",
    ShowValue: true,
}

s := sparkline.New(config)

// Simulate real-time updates
for {
    // Get your metric (CPU, memory, network, etc.)
    value := getCurrentNetworkTraffic()
    
    // Add to sparkline - old data automatically slides out
    s.AddPoint(value)
    
    // Display updated chart
    fmt.Print("\033[H\033[2J") // Clear screen
    fmt.Println(s.Render())
    
    time.Sleep(1 * time.Second)
}
```

## Visual Styles

Choose the style that best fits your data and context:

- **`StyleBars`**: Classic bar chart using Unicode block characters
- **`StyleDots`**: Scatter plot style with dots at value positions  
- **`StyleLine`**: Connected line chart with slope indicators
- **`StyleFilled`**: Area chart showing filled regions under the curve

## Configuration Options

The `Config` struct provides extensive customization:

```go
config := sparkline.Config{
    // Dimensions
    Width:     50,    // Chart width in characters
    Height:    8,     // Chart height in rows
    MaxPoints: 200,   // Maximum data points to keep in memory
    
    // Visual style
    Style: sparkline.StyleBars,
    
    // Display options
    Title:      "Server Response Time",
    ShowValue:  true,  // Show current value
    ShowMinMax: true,  // Show min/max range
    
    // Color coding (requires lipgloss styles)
    ColorRanges: []sparkline.ColorRange{
        {Min: 0, Max: 100, Style: greenStyle},    // Good
        {Min: 100, Max: 200, Style: yellowStyle}, // Warning  
        {Min: 200, Max: math.Inf(1), Style: redStyle}, // Critical
    },
    DefaultStyle: whiteStyle,
}
```

## Color-Coded Alerts

Use color ranges to create visual alerts based on data values:

```go
import "github.com/charmbracelet/lipgloss"

// Define your alert colors
greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2"))   // Green
yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))  // Yellow  
redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("1"))     // Red

config := sparkline.Config{
    Width:  40,
    Height: 5,
    Style:  sparkline.StyleBars,
    Title:  "CPU Temperature (°C)",
    ShowValue: true,
    ColorRanges: []sparkline.ColorRange{
        {Min: 0, Max: 60, Style: greenStyle},      // Safe
        {Min: 60, Max: 80, Style: yellowStyle},    // Warm
        {Min: 80, Max: math.Inf(1), Style: redStyle}, // Hot!
    },
}
```

Now your sparkline will automatically color-code values: green for safe temperatures, yellow for warnings, and red for critical levels.

## Interactive Applications with Bubble Tea

For terminal user interfaces, sparklines integrate seamlessly with [Bubble Tea](https://github.com/charmbracelet/bubbletea):

```go
type Model struct {
    sparkline *sparkline.Sparkline
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case TickMsg:
        // Update with new data
        newValue := getMetric()
        m.sparkline.AddPoint(newValue)
        return m, m.tick()
    }
    return m, nil
}

func (m Model) View() string {
    return m.sparkline.View() // Implements tea.Model interface
}
```

## Memory Management

Sparklines automatically manage memory to prevent unbounded growth:

- **Sliding Window**: When `MaxPoints` is exceeded, oldest data is automatically removed
- **FIFO Behavior**: New points push out old ones, maintaining recent history
- **Bounded Storage**: Memory usage stays constant regardless of runtime duration

This makes sparklines perfect for long-running monitoring applications.

## Use Cases

Sparklines excel in many scenarios:

- **System Monitoring**: CPU, memory, disk, network metrics
- **Application Performance**: Response times, error rates, throughput
- **Business Metrics**: Sales trends, user activity, conversion rates  
- **DevOps Dashboards**: Build times, deployment frequency, uptime
- **IoT Data**: Sensor readings, environmental monitoring
- **Financial Data**: Stock prices, trading volumes, portfolio performance

## Examples

Check out the included examples:

- **Basic Demo**: `go run ./cmd/apps/sparkline-test demo`
- **Interactive TUI**: `go run ./cmd/apps/sparkline-test`

The interactive demo shows all visual styles with real-time data generation, and lets you experiment with different configurations and update speeds.

## Contributing

This sparkline library is part of the [go-go-labs](https://github.com/go-go-golems/go-go-labs) project. Contributions, bug reports, and feature requests are welcome!

## License

Part of the go-go-labs project. See the main repository for license details.
