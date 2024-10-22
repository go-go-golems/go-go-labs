## Markdown Responsive Component Specification

### Overview

The **Markdown Responsive Component** is a standalone Terminal User Interface (TUI) component designed using the [Charmbracelet](https://github.com/charmbracelet) suite, specifically leveraging [Bubbletea](https://github.com/charmbracelet/bubbletea) for state management and [Glamour](https://github.com/charmbracelet/glamour) for rendering Markdown content. This component is responsible for rendering Markdown content within a responsive and interactive viewport. It can operate in two modes:

- **Static Mode:** Renders the Markdown content without scrolling, fitting entirely within the allocated space.
- **Scrollable Mode:** Enables vertical scrolling using the `bubbles/viewport` to navigate through content that exceeds the display area.

### Key Features

- **Markdown Rendering:** Utilizes Glamour to render Markdown strings into styled terminal output.
- **Responsive Design:** Automatically adjusts to terminal window size changes, ensuring optimal display of content.
- **Scrolling Capability:** Incorporates `bubbles/viewport` to provide smooth scrolling when content exceeds the visible area.
- **Mode Flexibility:** Can operate either with scrolling enabled or in a static display mode without scrolling.
- **Lightweight Integration:** Designed to be embedded within larger applications or used as a standalone component.

### Functional Requirements

1. **Initialization:**
   - Accept a Markdown string to render.
   - Configure rendering options, including styles and word wrapping.

2. **Rendering:**
   - Render the Markdown content using Glamour.
   - Display the rendered content within a viewport if scrolling is enabled.
   - Adjust the layout based on the terminal's current size.

3. **User Interaction:**
   - In Scrollable Mode:
     - Support navigation using arrow keys (`Up`, `Down`), `Page Up`, and `Page Down`.
     - Optionally support mouse wheel scrolling if enabled.
   - In Static Mode:
     - Disable scrolling interactions.

4. **Resizing:**
   - Detect terminal resize events and adjust the viewport dimensions accordingly.
   - Re-render content to fit the new size without disrupting the current scroll position.

5. **Content Management:**
   - Allow dynamic updating of the Markdown content.
   - Provide methods to enable or disable scrolling based on content size or user preference.

### Non-Functional Requirements

- **Performance:** Ensure smooth rendering and scrolling, even with large Markdown documents.
- **Modularity:** Encapsulate functionality to allow easy integration and reuse within different parts of the application.
- **Maintainability:** Write clean, well-documented code to facilitate future enhancements and debugging.
- **Accessibility:** Ensure that key bindings are intuitive and that visual indicators (e.g., scroll position) are clear.

---

## Markdown Responsive Component Architecture

### Overview

The **Markdown Responsive Component** is architected using the Model-View-Update (MVU) pattern provided by Bubbletea. It integrates the `bubbles/viewport` for handling scrollable content and the `glamour` package for Markdown rendering. The component is designed to be flexible, allowing it to function independently or as part of a larger interface.

### Component Structure

```go
package markdownview

import (
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/glamour"
    "github.com/charmbracelet/lipgloss"
)

// ViewMode defines the mode of the MarkdownView
type ViewMode int

const (
    Static ViewMode = iota // No scrolling
    Scrollable              // Enable scrolling
)

// MarkdownView represents the model for a markdown component
type MarkdownView struct {
    // Configuration
    Content       string           // Raw Markdown content
    Rendered      string           // Rendered Markdown content
    GlamourStyle  string           // Glamour style to use
    ViewMode      ViewMode         // Current view mode
    Renderer      *glamour.TermRenderer // Glamour renderer instance
    Styles        lipgloss.Style   // Lip Gloss styles for the viewport

    // Viewport for scrollable content
    Viewport      viewport.Model

    // Flags
    NeedsRender   bool             // Indicates if re-rendering is needed
}

// NewMarkdownView initializes a new MarkdownView
func NewMarkdownView(content string, glamourStyle string, viewMode ViewMode) (*MarkdownView, error) {
    renderer, err := glamour.NewTermRenderer(
        glamour.WithStandardStyle(glamourStyle),
        glamour.WithWordWrap(80), // Default word wrap, can be adjusted
    )
    if err != nil {
        return nil, err
    }

    rendered, err := renderer.Render(content)
    if err != nil {
        return nil, err
    }

    // Initialize viewport
    vp := viewport.New(0, 0) // Width and Height will be set in Init
    vp.KeyMap = viewport.KeyMap{
        PageDown:     viewport.KeyMap{}.PageDown,
        PageUp:       viewport.KeyMap{}.PageUp,
        HalfPageUp:   viewport.KeyMap{}.HalfPageUp,
        HalfPageDown: viewport.KeyMap{}.HalfPageDown,
        Down:         viewport.KeyMap{}.Down,
        Up:           viewport.KeyMap{}.Up,
    }
    vp.SetContent(rendered)

    // Define styles (optional customization)
    styles := lipgloss.NewStyle().
        Padding(1, 2).
        Border(lipgloss.NormalBorder()).
        BorderForeground(lipgloss.Color("240")) // Light grey border

    return &MarkdownView{
        Content:      content,
        Rendered:     rendered,
        GlamourStyle: glamourStyle,
        ViewMode:     viewMode,
        Renderer:     renderer,
        Styles:       styles,
        Viewport:     vp,
        NeedsRender:  false,
    }, nil
}

// Init initializes the MarkdownView component
func (m *MarkdownView) Init() bubbletea.Cmd {
    // Initial commands can be placed here
    return nil
}

// Update handles incoming messages and updates the model accordingly
func (m *MarkdownView) Update(msg bubbletea.Msg) (bubbletea.Model, bubbletea.Cmd) {
    var cmd bubbletea.Cmd

    switch m.ViewMode {
    case Scrollable:
        m.Viewport, cmd = m.Viewport.Update(msg)
    case Static:
        // No update needed for static mode
    }

    // Handle terminal resize to adjust viewport size
    switch msg := msg.(type) {
    case bubbletea.WindowSizeMsg:
        width := msg.Width
        height := msg.Height

        // Adjust viewport size
        if m.ViewMode == Scrollable {
            m.Viewport.Width = width
            m.Viewport.Height = height
        }

        // In Static mode, we might want to adjust the word wrap or other rendering options
        if m.ViewMode == Static {
            // Re-render content to fit the new size
            rendered, err := m.Renderer.Render(m.Content)
            if err == nil {
                m.Rendered = rendered
                m.Viewport.SetContent(rendered)
            }
        }
    }

    // Re-render if needed
    if m.NeedsRender {
        rendered, err := m.Renderer.Render(m.Content)
        if err == nil {
            m.Rendered = rendered
            if m.ViewMode == Scrollable {
                m.Viewport.SetContent(rendered)
            }
            m.NeedsRender = false
        }
    }

    return m, cmd
}

// View renders the MarkdownView component
func (m *MarkdownView) View() string {
    var content string

    switch m.ViewMode {
    case Scrollable:
        // Apply styles and render viewport
        return m.Styles.Render(m.Viewport.View())
    case Static:
        // Render without viewport
        styledContent := lipgloss.NewStyle().
            Padding(1, 2).
            Border(lipgloss.NormalBorder()).
            BorderForeground(lipgloss.Color("240")).
            Render(m.Rendered)
        return styledContent
    default:
        return "Invalid View Mode"
    }
}

// SetContent updates the Markdown content and triggers re-rendering
func (m *MarkdownView) SetContent(content string) error {
    m.Content = content
    rendered, err := m.Renderer.Render(content)
    if err != nil {
        return err
    }
    m.Rendered = rendered

    if m.ViewMode == Scrollable {
        m.Viewport.SetContent(rendered)
    } else {
        m.NeedsRender = true
    }

    return nil
}

// ToggleScrollMode switches between Static and Scrollable modes
func (m *MarkdownView) ToggleScrollMode() {
    if m.ViewMode == Scrollable {
        m.ViewMode = Static
    } else {
        m.ViewMode = Scrollable
    }
}

// Resize adjusts the component based on new terminal dimensions
func (m *MarkdownView) Resize(width, height int) {
    if m.ViewMode == Scrollable {
        m.Viewport.Width = width
        m.Viewport.Height = height
    }
}
```

### Component Breakdown

#### 1. **Model (`MarkdownView`)**

- **Fields:**
  - `Content`: The raw Markdown string to be rendered.
  - `Rendered`: The Markdown content after being processed by Glamour.
  - `GlamourStyle`: The style name or path used by Glamour for rendering.
  - `ViewMode`: Enum indicating whether the component is in `Static` or `Scrollable` mode.
  - `Renderer`: Instance of Glamour's `TermRenderer` for rendering Markdown.
  - `Styles`: Lip Gloss styles applied to the viewport or static content.
  - `Viewport`: Instance of `bubbles/viewport.Model` used for handling scrollable content.
  - `NeedsRender`: Flag to indicate if re-rendering is required (e.g., after content update).

#### 2. **Initialization (`NewMarkdownView`)**

- **Parameters:**
  - `content`: The Markdown string to render.
  - `glamourStyle`: The style to use for rendering (e.g., "dark", "light").
  - `viewMode`: Initial mode (`Static` or `Scrollable`).

- **Process:**
  - Initializes the Glamour renderer with the specified style.
  - Renders the initial Markdown content.
  - Sets up the viewport with default key bindings.
  - Applies Lip Gloss styles for visual consistency.

#### 3. **Initialization Method (`Init`)**

- **Purpose:** Prepares any initial commands or setup required when the component starts. Currently, it does not perform any actions but is included to satisfy the `tea.Model` interface for composability.

#### 4. **Update Method (`Update`)**

- **Purpose:** Handles incoming messages to update the component's state.

- **Message Handling:**
  - **Scrollable Mode:**
    - Delegates message handling to the `viewport` component.
  - **Static Mode:**
    - Ignores scroll-related messages.

  - **Window Resize (`bubbletea.WindowSizeMsg`):**
    - Adjusts viewport dimensions based on new terminal size.
    - Re-renders content if in `Static` mode to fit the new dimensions.

  - **Re-rendering:**
    - If `NeedsRender` is true, re-renders the Markdown content.

#### 5. **View Method (`View`)**

- **Purpose:** Renders the component's current state to a string that Bubbletea can display.

- **Rendering Logic:**
  - **Scrollable Mode:**
    - Applies Lip Gloss styles and renders the viewport content.
  - **Static Mode:**
    - Renders the entire Markdown content without scrollbars, applying styles directly.

#### 6. **Content Management (`SetContent`)**

- **Purpose:** Updates the Markdown content and triggers a re-render.

- **Process:**
  - Updates the `Content` field with the new Markdown string.
  - Re-renders the content using Glamour.
  - Updates the viewport's content if in `Scrollable` mode.
  - Sets the `NeedsRender` flag if in `Static` mode to trigger re-rendering.

#### 7. **Mode Toggling (`ToggleScrollMode`)**

- **Purpose:** Switches between `Static` and `Scrollable` modes.

- **Process:**
  - Toggles the `ViewMode` field.
  - Adjusts rendering and interaction capabilities based on the new mode.

#### 8. **Resizing (`Resize`)**

- **Purpose:** Adjusts the component's layout based on new terminal dimensions.

- **Process:**
  - Updates the viewport's width and height if in `Scrollable` mode.
  - In `Static` mode, triggers a re-render to fit content within the new size.

### Integration with Multi-Slide View

The `MarkdownView` component is designed to be instantiated multiple times (e.g., six instances for a multi-slide view) and arranged in a grid layout. Each instance operates independently, handling its own content rendering and interaction, while the parent component manages the overall layout and coordination.

### Styling and Theming

- **Lip Gloss:**
  - Applied to the viewport and static content to ensure consistent borders, padding, and colors.
  - Customizable to match the application's overall theme.

- **Glamour:**
  - Renders Markdown with the specified style, enhancing readability and aesthetics.
  - Styles can be dynamically changed based on user preferences or application settings.

### Error Handling

- **Markdown Rendering Errors:**
  - Capture and handle errors from Glamour during the rendering process.
  - Provide fallback content or error messages within the viewport.

- **Viewport Issues:**
  - Handle potential errors related to viewport resizing or rendering.
  - Ensure that the component remains stable even when encountering unexpected states.

### Example Usage

```go
package main

import (
    "os"

    "github.com/charmbracelet/bubbletea"
    "github.com/yourusername/markdownview"
)

func main() {
    markdownContent := `
# Slide Title

This is an example slide.

- Point 1
- Point 2
- Point 3
`

    // Initialize MarkdownView in Scrollable mode
    slideView, err := markdownview.NewMarkdownView(markdownContent, "dark", markdownview.Scrollable)
    if err != nil {
        panic(err)
    }

    // Create the Bubbletea program with the MarkdownView as the initial model
    p := bubbletea.NewProgram(slideView)

    // Start the program
    if err := p.Start(); err != nil {
        os.Exit(1)
    }
}
```

### Summary

The **Markdown Responsive Component** provides a robust and flexible solution for rendering Markdown content within a TUI application. By leveraging Charmbracelet's Bubbletea and Bubbles libraries, along with Glamour for Markdown rendering, the component ensures high performance, responsiveness, and ease of integration. Its ability to operate in both static and scrollable modes makes it adaptable to various display requirements, while its modular design facilitates reuse within larger systems, such as multi-slide editors.
