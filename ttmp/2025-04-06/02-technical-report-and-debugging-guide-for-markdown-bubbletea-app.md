# Technical Report and Debugging Guide for Markdown Bubbletea App

## Overview

The Markdown Bubbletea App is a terminal-based application that allows users to test how the `glamour` library renders markdown content in real-time. The application features a split-screen interface with a text editor at the bottom and a rendered preview of that text at the top. This tool is particularly useful for testing how `glamour` handles partial or incomplete markdown.

The application is built using the [Bubble Tea](https://github.com/charmbracelet/bubbletea) framework from Charmbracelet, along with several other libraries from their ecosystem.

## Architecture

The application follows the Model-View-Update (MVU) architecture pattern as implemented by the Bubble Tea framework:

1. **Model**: Maintains the application state
2. **Update**: Handles user input and updates the model
3. **View**: Renders the current state to the terminal

### Project Structure

The project is located at:
- Package path: `github.com/go-go-golems/go-go-labs/cmd/apps/bubbletea-markdown-test`
- Main file: `main.go`

### Dependencies

The application relies on several external libraries:

- `github.com/charmbracelet/bubbletea`: Core TUI framework
- `github.com/charmbracelet/bubbles/textarea`: Text editor component
- `github.com/charmbracelet/bubbles/viewport`: Scrollable viewport component
- `github.com/charmbracelet/glamour`: Markdown renderer
- `github.com/charmbracelet/lipgloss`: Styling and layout
- `github.com/rs/zerolog`: Structured logging
- `github.com/spf13/cobra`: CLI commands and flags

## Key Components

### Model

The application state is encapsulated in the `model` struct:

```go
type model struct {
    viewport       viewport.Model    // Displays rendered markdown
    textarea       textarea.Model    // Text input area
    renderer       *glamour.TermRenderer  // Glamour markdown renderer
    width          int               // Terminal width
    height         int               // Terminal height
    err            error             // Any rendering errors
    renderMarkdown bool              // Toggle for markdown rendering
    showHelp       bool              // Toggle for help display
}
```

### Main Features

1. **Text Input Area**: A multi-line text editor for entering markdown content
2. **Rendering Preview**: A viewport that displays either:
   - Plain text (when markdown rendering is disabled)
   - Glamour-rendered markdown (when markdown rendering is enabled)
3. **Togglable Help Display**: Shows keyboard shortcuts (`Ctrl+H` to toggle)
4. **Markdown Rendering Toggle**: Switch between plain text and rendered markdown (`Ctrl+M` to toggle)
5. **Status Display**: Shows current mode and content stats
6. **Debug Logging**: Extensive zerolog-based logging to `/tmp/external.log`
7. **File Output**: Saves rendered content to `/tmp/rendered.md` for inspection

### Command-Line Interface

The application uses Cobra to provide a CLI with flags:

```
Usage:
  markdown-test [flags]

Flags:
      --initial-text string   Initial text to show in editor
      --log-level string      Log level (debug, info, warn, error) (default "info")
      --render-markdown       Start with markdown rendering enabled
      --show-help             Start with help visible (default true)
```

## Key Functions

### Initialization

1. `main()`: Entry point that sets up the Cobra command
2. `runApp()`: Sets up logging and initializes the Bubble Tea program
3. `initialModel()`: Creates and configures the initial application state
4. `setupLogging()`: Configures zerolog to write to `/tmp/external.log`

### Core Logic

1. `Init()`: Returns the initial command (textarea.Blink)
2. `Update()`: Handles input events and updates the model
3. `View()`: Renders the current state to the terminal
4. `renderContent()`: Updates the viewport with either plain text or rendered markdown
5. `saveRenderedContent()`: Saves the current viewport content to `/tmp/rendered.md`

## Rendering Flow

1. User types in the textarea component
2. On each keystroke, `renderContent()` is called
3. Based on the `renderMarkdown` flag:
   - If `true`: Content is rendered through glamour and displayed in the viewport
   - If `false`: Raw text is displayed in the viewport
4. The rendered content is saved to `/tmp/rendered.md`
5. Status line is updated with the current mode and content stats

## Keyboard Shortcuts

- `Ctrl+M`: Toggle markdown rendering on/off
- `Ctrl+H`: Toggle help display on/off
- `Ctrl+C` or `Esc`: Quit the application

## Debugging Guide

### Known Issues

1. **Hanging on Startup**: The application may appear to hang when first started. This could be related to the initial renderer setup or how the viewport is populated.

2. **Rendering Errors**: Sometimes glamour may have issues rendering incomplete markdown syntax, which will be shown in the error display.

### Debugging Steps

1. **Check Logs**: Review `/tmp/external.log` for detailed runtime information. The logs include:
   - Caller information (file and line number)
   - Timestamped events
   - Detailed state changes
   - Error messages

2. **Examine Rendered Output**: Look at `/tmp/rendered.md` to see exactly what content is being generated.

3. **Run with Debug Logging**: Use `--log-level=debug` for the most verbose logging:
   ```
   ./markdown-test --log-level=debug
   ```

4. **Test Different Start Modes**: 
   - Start with markdown rendering off: `./markdown-test`
   - Start with markdown rendering on: `./markdown-test --render-markdown`
   - Start with predefined text: `./markdown-test --initial-text="# Test Heading"`

5. **Check for Deadlocks**: If the application hangs, the logs may show where the process got stuck.

### Common Log Patterns to Look For

- Missing or out-of-sequence `Update called` events
- Long gaps between logged events
- Repeated error messages
- Viewport or textarea update failures

## Running and Building

**Note**: All building and running of the application happens outside the editor in a terminal. There is no need to call any special tools to do this.

1. **Building**: Navigate to the application directory and run:
   ```bash
   go build -o markdown-test
   ```

2. **Running**: After building, execute the binary:
   ```bash
   ./markdown-test [flags]
   ```

## Implementation Details

### Viewport and Textarea Sizing

The application manages the terminal real estate by allocating space proportionally:
- Text area gets a fixed height (5 lines + border)
- Status display gets 1 line
- Help gets 3 lines when visible
- Viewport gets the remaining space

When the terminal is resized, these proportions are recalculated, and a new renderer is created to accommodate the new width.

### Glamour Rendering

Glamour rendering is controlled by the `renderMarkdown` flag, which can be toggled with `Ctrl+M`. When enabled, the application:
1. Passes the current textarea content to the glamour renderer
2. Updates the viewport with the rendered output
3. Saves the rendered content to `/tmp/rendered.md`
4. Updates the status display to show "Mode: Markdown"

### Error Handling

Errors are handled at multiple levels:
1. Rendering errors are displayed in the UI and logged
2. File I/O errors are logged but may not be visible in the UI
3. Critical errors during initialization will prevent the application from starting

## Future Improvements

Potential improvements that could be made to the application:

1. Add a theme selector for different Glamour styles
2. Add ability to save/load markdown files
3. Implement syntax highlighting in the textarea
4. Add a split-pane view that can be resized
5. Improve error handling and recovery
6. Add undo/redo functionality

## Conclusion

This application serves as a tool for testing how the Glamour library renders markdown in real-time. By allowing toggling between plain text and rendered markdown, it makes it easy to see exactly how the library processes different markdown constructs, particularly when they are incomplete or malformed.

When debugging issues, focus on the logs, examine the rendered output file, and use the debug mode to trace the application flow. 