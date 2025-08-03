# Bubbletea Redis Monitor TUI Design

## Overview

Design for a modular, widget-based Redis Streams Monitor TUI using bubbletea models and lipgloss for rendering. The goal is to create a "top-like" interface that shows all information on a single screen with real-time updates.

## Current Architecture Assessment

### What Exists
- **Redis Client Layer**: Well-implemented Redis operations (`redis.go`)
- **CLI Commands**: Working glazed-based CLI commands
- **TUI Components**: Basic component structure in `pkg/tui/components/`
- **Sparkline Library**: Complete sparkline implementation in `bobatea/pkg/sparkline/`
- **Styles**: Basic styles in `pkg/tui/styles/`

### Issues with Current Implementation
- **Over-engineered**: Coordinator pattern is complex for this use case
- **Not truly modular**: Components aren't independent bubbletea models
- **Mixed concerns**: Rendering logic mixed with data management
- **Complex message passing**: Too many custom message types

## Proposed Bubbletea Architecture

### Core Principles
1. **Widget-based**: Each UI section is an independent bubbletea model
2. **Composition over coordination**: Main model composes widgets, no complex coordinator
3. **Event-driven**: Use bubbletea's message system for communication
4. **Responsive**: Widgets adapt to terminal size changes
5. **Data separation**: Clear separation between data fetching and UI rendering

### Widget Hierarchy

```
RootModel (main bubbletea model)
├── HeaderWidget (server info, uptime, refresh rate)
├── StreamsTableWidget (streams with sparklines)
├── GroupsTableWidget (consumer groups detail)
├── AlertsWidget (memory/trim alerts)
├── MetricsWidget (global throughput + memory progress)
└── FooterWidget (keyboard commands)
```

### Model Structure

#### 1. RootModel
```go
type RootModel struct {
    // Data
    redisClient   RedisClient
    serverData    ServerData
    streamsData   []StreamData
    refreshRate   time.Duration
    demoMode      bool
    
    // UI State
    width, height int
    focused       string // which widget has focus
    
    // Widgets
    header  HeaderWidget
    streams StreamsTableWidget
    groups  GroupsTableWidget
    alerts  AlertsWidget
    metrics MetricsWidget
    footer  FooterWidget
    
    // Styles
    styles Styles
}
```

#### 2. Widget Interface
```go
type Widget interface {
    tea.Model
    SetSize(width, height int)
    SetFocused(focused bool)
    Height() int // How much vertical space the widget needs
}
```

#### 3. Individual Widgets

##### HeaderWidget
- Shows title, uptime, refresh rate
- Minimal state, mostly displays passed data
- Fixed height (2 lines)

##### StreamsTableWidget  
- Table with borders using lipgloss
- Embedded sparklines for each stream
- Scrollable if many streams
- Variable height based on content

##### GroupsTableWidget
- Consumer groups with consumer details
- Nested table structure
- Variable height based on content

##### AlertsWidget
- Memory alerts and trim warnings
- Simple text list with bullet points
- Fixed height (4-5 lines max)

##### MetricsWidget
- Global throughput sparkline
- Memory usage progress bar
- Fixed height (2 lines)

##### FooterWidget
- Keyboard commands help
- Fixed height (1 line)

### Message Types

#### Core Messages
```go
// Data update messages
type DataUpdateMsg struct {
    ServerData  ServerData
    StreamsData []StreamData
    Timestamp   time.Time
}

type RefreshTickMsg struct {
    Time time.Time
}

// UI control messages  
type FocusChangeMsg struct {
    Widget string
}

type RefreshRateChangeMsg struct {
    NewRate time.Duration
}
```

#### Widget-specific Messages
```go
// For sparklines
type SparklineUpdateMsg struct {
    StreamName string
    Data       []float64
}

// For progress bars
type ProgressUpdateMsg struct {
    Percent float64
}
```

### Keyboard Mappings

```go
type KeyMap struct {
    Quit         key.Binding // q, ctrl+c
    Refresh      key.Binding // r
    RefreshUp    key.Binding // +, =
    RefreshDown  key.Binding // -, _
    FocusNext    key.Binding // tab
    FocusPrev    key.Binding // shift+tab
    ScrollUp     key.Binding // up, k
    ScrollDown   key.Binding // down, j
    Help         key.Binding // ?
}
```

### Layout Strategy

#### Responsive Layout
```go
func (m RootModel) View() string {
    var sections []string
    
    // Header (fixed)
    headerContent := m.header.View()
    sections = append(sections, headerContent)
    
    // Calculate remaining height
    usedHeight := m.header.Height() + m.footer.Height()
    remainingHeight := m.height - usedHeight
    
    // Distribute remaining space
    streamHeight := min(remainingHeight/2, len(m.streamsData)*2+3) // 2 lines per stream + borders
    groupHeight := min(remainingHeight/3, len(allGroups)+3)
    alertHeight := 4
    metricHeight := 2
    
    // Set widget sizes
    m.streams.SetSize(m.width, streamHeight)
    m.groups.SetSize(m.width, groupHeight)
    m.alerts.SetSize(m.width, alertHeight)
    m.metrics.SetSize(m.width, metricHeight)
    
    // Render widgets
    sections = append(sections, m.streams.View())
    sections = append(sections, m.groups.View())
    sections = append(sections, m.alerts.View())
    sections = append(sections, m.metrics.View())
    
    // Footer (fixed)
    sections = append(sections, m.footer.View())
    
    return lipgloss.JoinVertical(lipgloss.Top, sections...)
}
```

### Data Flow

#### 1. Initialization
```
main() -> NewRootModel() -> Init() -> fetchData() -> tick timer
```

#### 2. Data Updates
```
tick timer -> fetchData() -> DataUpdateMsg -> Update() -> propagate to widgets
```

#### 3. User Input
```
keyboard -> KeyMsg -> Update() -> action or FocusChangeMsg -> Update()
```

#### 4. Widget Updates
```
DataUpdateMsg -> widget.Update() -> widget recalculates view -> View()
```

### Lipgloss Usage

#### Borders and Tables
```go
var (
    tableStyle = lipgloss.NewStyle().
        Border(lipgloss.RoundedBorder()).
        BorderForeground(lipgloss.Color("240"))
    
    headerStyle = lipgloss.NewStyle().
        Background(lipgloss.Color("57")).
        Foreground(lipgloss.Color("15")).
        Bold(true).
        Padding(0, 1)
        
    cellStyle = lipgloss.NewStyle().
        Padding(0, 1)
)
```

#### Progress Bars
```go
// Use charmbracelet/bubbles/progress
progressModel := progress.New(progress.WithDefaultGradient())
progressModel.Width = 40
```

#### Sparklines
```go
// Use existing bobatea/pkg/sparkline with lipgloss integration
sparklineConfig := sparkline.Config{
    Width:     25,
    Height:    1,
    MaxPoints: 25,
    Style:     sparkline.StyleBars,
}
```

### Error Handling

#### Connection Errors
```go
type ErrorMsg struct {
    Err error
    Source string // "redis", "data", etc.
}

// Show error overlay or status in header
```

#### Data Validation
```go
// Validate data before updating widgets
// Handle partial failures gracefully
// Show warnings in alerts widget
```

### Performance Considerations

#### Efficient Rendering
- Only re-render widgets that changed
- Use lipgloss caching where possible
- Limit sparkline data points (20-30 max)
- Throttle updates to prevent flicker

#### Memory Management
- Bounded sparkline history
- Clean up old data periodically
- Avoid creating new styles on each render

### Testing Strategy

#### Widget Testing
```go
func TestStreamTableWidget(t *testing.T) {
    widget := NewStreamTableWidget(testStyles)
    widget.SetSize(80, 10)
    
    // Test with sample data
    widget.Update(DataUpdateMsg{...})
    output := widget.View()
    
    // Assert table format, borders, content
}
```

#### Integration Testing
```go
func TestFullLayout(t *testing.T) {
    model := NewRootModel(demoMode=true)
    model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
    
    view := model.View()
    lines := strings.Split(view, "\n")
    
    // Assert layout structure
    assert.Equal(t, 40, len(lines))
}
```

### Migration Strategy

#### Phase 1: Simplify Current Code
1. Remove coordinator complexity
2. Convert existing components to proper bubbletea models
3. Implement Widget interface

#### Phase 2: Enhance Widgets
1. Add proper lipgloss borders and styling
2. Integrate sparklines and progress bars
3. Implement responsive layout

#### Phase 3: Polish
1. Add keyboard navigation
2. Improve error handling
3. Add animations and transitions

### File Structure

```
pkg/tui/
├── models/
│   ├── root.go              # RootModel
│   └── data_types.go        # Data structures (keep existing)
├── widgets/
│   ├── widget.go            # Widget interface
│   ├── header.go            # HeaderWidget
│   ├── streams_table.go     # StreamsTableWidget
│   ├── groups_table.go      # GroupsTableWidget
│   ├── alerts.go            # AlertsWidget
│   ├── metrics.go           # MetricsWidget
│   └── footer.go            # FooterWidget
├── styles/
│   └── styles.go            # Lipgloss styles (enhance existing)
└── keys/
    └── keys.go              # Keyboard mappings
```

### Benefits of This Design

1. **Simplicity**: Each widget is independent and focused
2. **Testability**: Widgets can be tested in isolation
3. **Maintainability**: Clear separation of concerns
4. **Extensibility**: Easy to add new widgets or modify existing ones
5. **Performance**: Efficient rendering and updates
6. **Responsive**: Adapts to different terminal sizes
7. **Bubbletea-idiomatic**: Uses bubbletea patterns properly

This design removes the complexity of the coordinator pattern while maintaining modularity through proper bubbletea model composition.
