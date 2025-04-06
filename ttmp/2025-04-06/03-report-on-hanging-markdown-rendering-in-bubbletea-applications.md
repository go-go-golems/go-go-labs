# Technical Report: Debugging Startup Hang and Keybinding Issues in a Bubble Tea Markdown Rendering Application

**Date:** 2025-04-06
**Case:** The Case of the Lagging Listener
**Application:** `cmd/apps/bubbletea-markdown-test`

## 1. Introduction: The Problem

A Bubble Tea application (`bubbletea-markdown-test`) designed to provide real-time markdown rendering previews using the `glamour` library exhibited two critical issues upon initial testing:

1.  **Startup Hang:** The application appeared to freeze or hang for approximately 5 seconds immediately after launching before becoming responsive.
2.  **Unresponsive Keybindings:** Specific keyboard shortcuts intended to toggle functionality (`Ctrl+M` for markdown rendering, `Ctrl+H` for help) did not work as expected. `Ctrl+M` was incorrectly interpreted as an `Enter` key press, inserting a newline into the text area.

This report details the investigation process, findings, and solutions implemented to resolve these issues.

## 2. Initial Investigation: Logs and Hypothesis

### 2.1. Log Analysis (Startup Hang)

Analysis of the application's debug logs (`/tmp/external.log`) revealed a significant delay occurring during the initial event processing cycle. Specifically, a ~5-second gap was observed between log entries related to processing the initial `tea.WindowSizeMsg`:

```log
# ... initial startup logs ...
2025-04-06T11:40:56-04:00 DBG ... > Update called ... msgType=tea.WindowSizeMsg
2025-04-06T11:40:56-04:00 DBG ... > Window size changed ... width=121 height=62
2025-04-06T11:40:56-04:00 DBG ... > Calculated component heights ... viewportHeight=50
2025-04-06T11:40:56-04:00 DBG ... > Creating new renderer for new width ... width=121
# --- Approximately 5-second gap here ---
2025-04-06T11:41:01-04:00 DBG ... > Rendering content ...
# ... subsequent logs ...
```

The log message immediately preceding the delay was `Creating new renderer for new width`. This pointed towards the code block within the `tea.WindowSizeMsg` handler responsible for creating a new `glamour.TermRenderer` instance:

```go
// Inside Update func, case tea.WindowSizeMsg:
logWithCaller(zerolog.DebugLevel, "Creating new renderer for new width", ...)
renderer, err := glamour.NewTermRenderer(
    glamour.WithAutoStyle(),
    glamour.WithWordWrap(m.width),
)
// ... rest of handler ...
```

### 2.2. Initial Hypothesis (Startup Hang)

The initial hypothesis was that the `glamour.NewTermRenderer` function itself was computationally expensive or causing blocking behavior, especially when called within the potentially rapid-fire event handling of Bubble Tea's initial startup and window sizing.

### 2.3. Log Analysis (Keybindings)

When pressing `Ctrl+M`, the logs showed:

```log
2025-04-06T11:52:24-04:00 DBG ... > Update called ... msgType=tea.KeyMsg
2025-04-06T11:52:24-04:00 DBG ... > Key pressed ... key=enter
2025-04-06T11:52:24-04:00 DBG ... > Updating textarea ...
```

This indicated that the application was receiving an `Enter` key event, not a distinct `Ctrl+M` event, causing it to fall through the specific keybinding checks and be handled as text input.

## 3. Isolating the Renderer Performance

To test the hypothesis about `NewTermRenderer` performance, a minimal standalone Go application (`cmd/experimences/glamour-renderer-debugging`) was created. This application:

*   Included the same logging setup.
*   Used Cobra for flags.
*   Specifically timed the execution of `glamour.NewTermRenderer`.

Execution logs (`/tmp/glamour-debug.log`) from this test application showed:

```log
2025-04-06T11:50:56-04:00 DBG ... > Calling glamour.NewTermRenderer ... width=80
2025-04-06T11:50:56-04:00 INF ... > glamour.NewTermRenderer call completed duration=4.21369ms duration_ms=4 ...
```

**Finding:** Creating a `glamour.TermRenderer` instance is extremely fast (~4ms) in isolation. The slowness was not inherent to the function itself.

## 4. Identifying Root Causes and Solutions

### 4.1. Root Cause (Startup Hang)

While `NewTermRenderer` is fast, recreating it **on every `tea.WindowSizeMsg`**, including the initial one triggered at startup, introduced significant overhead within the Bubble Tea event loop. Bubble Tea processes initial messages (like window size) quickly, and inserting a potentially complex object recreation (even if individually fast) into this handler caused the perceived hang.

**Solution:**

The `glamour.TermRenderer` should be initialized **once** when the model is created (`initialModel` function). The `tea.WindowSizeMsg` handler should only update the dimensions of existing components (`viewport`, `textarea`) and trigger a re-render if necessary, but *not* recreate the renderer object.

The problematic code block was removed from the `Update` function's `tea.WindowSizeMsg` case.

```go
// Inside Update func, case tea.WindowSizeMsg:
// --- Removed Block Start ---
// logWithCaller(zerolog.DebugLevel, "Creating new renderer for new width", ...)
// renderer, err := glamour.NewTermRenderer(...)
// if err != nil { ... } else { m.renderer = renderer }
// --- Removed Block End ---

// Keep viewport and textarea resizing
m.viewport.Width = m.width
m.viewport.Height = calculatedViewportHeight
m.textarea.SetWidth(m.width)

// Trigger re-render using the existing renderer
m.renderContent()
```

This resolved the startup hang completely.

### 4.2. Root Cause (Keybindings)

The initial implementation checked key presses using direct string comparison (`msg.String() == "ctrl+m"`). This is brittle.

Even after switching to the recommended Bubble Tea approach using `key.Binding` and `key.Matches`, `Ctrl+M` still failed.

The root cause was identified as **standard terminal behavior**: Many terminal emulators translate the `Ctrl+M` key combination into a Carriage Return character (`\r`), which is the same code sent by the `Enter` key. The application was therefore correctly receiving an `Enter` event from the terminal, causing the `key.Matches(msg, m.keys.ToggleMarkdown)` check (which expected `ctrl+m`) to fail.

**Solution:**

Since the terminal's behavior cannot be reliably changed from within the application, the keybinding itself was modified to use a combination not typically intercepted or translated by terminals.

The binding for toggling markdown was changed from `ctrl+m` to `ctrl+t`.

```go
// In key map definition:
var defaultKeyMap = keyMap{
    // ... other keys
    ToggleMarkdown: key.NewBinding(
        key.WithKeys("ctrl+t"), // Changed from ctrl+m
        key.WithHelp("ctrl+t", "toggle markdown"),
    ),
    // ... other keys
}
```

This resolved the keybinding issue, allowing `Ctrl+T` to correctly toggle markdown rendering mode.

## 5. Conclusion

The investigation revealed two distinct issues stemming from different causes:

1.  The startup hang was caused by inefficiently recreating the `glamour.TermRenderer` within the `WindowSizeMsg` event handler, disrupting the initial event flow. The solution was to create the renderer once during model initialization.
2.  The unresponsive `Ctrl+M` keybinding was due to standard terminal behavior translating `Ctrl+M` to `Enter`. The solution was to change the keybinding to `Ctrl+T` to avoid this translation.

Key takeaways for developers building Bubble Tea applications, especially those involving complex components or rendering:

*   **Initialize expensive objects once:** Avoid recreating components or renderers within frequently triggered event handlers like `WindowSizeMsg` or `KeyMsg` unless absolutely necessary.
*   **Use `key.Binding`:** Leverage Bubble Tea's key binding system (`key.Binding`, `key.Matches`) for robust shortcut handling.
*   **Be aware of terminal behavior:** Understand that terminals can intercept or translate certain key combinations (like `Ctrl+M`, `Ctrl+I`, `Ctrl+J`, `Ctrl+H`) before they reach the application. Choose unambiguous keybindings for custom actions.
*   **Isolate for performance testing:** When suspecting a performance bottleneck, create minimal test cases to isolate the specific component or function call.

## 6. How to Add Togglable Markdown Rendering in a Bubble Tea App

Here's a guide to implementing a feature where the user can toggle between viewing raw text input and its `glamour`-rendered markdown equivalent:

**1. Update Model:**
   Add fields to your `model` struct:

   ```go
   import (
       "github.com/charmbracelet/bubbles/textarea"
       "github.com/charmbracelet/bubbles/viewport"
       "github.com/charmbracelet/bubbles/key"
       "github.com/charmbracelet/glamour"
       tea "github.com/charmbracelet/bubbletea"
   )

   type keyMap struct { // Define or reuse your keymap
       ToggleMarkdown key.Binding
       // ... other bindings
   }

   type model struct {
       textarea       textarea.Model    // For text input
       viewport       viewport.Model    // To display raw or rendered text
       renderer       *glamour.TermRenderer // Glamour renderer instance
       renderMarkdown bool              // Flag to track current mode
       keys           keyMap            // Key bindings
       err            error             // To store potential render errors
       width          int
       height         int
       // ... other model fields
   }
   ```

**2. Initialization (`initialModel`):**
   Create the renderer *once* when initializing the model.

   ```go
   func initialModel() model {
       ta := textarea.New()
       ta.Placeholder = "Enter markdown..."
       ta.Focus()

       vp := viewport.New(80, 10) // Initial dimensions

       // Create the renderer ONCE
       renderer, err := glamour.NewTermRenderer(
           glamour.WithAutoStyle(),
           // Add other glamour options as needed
       )
       if err != nil {
           // Handle error appropriately - maybe set m.err
           log.Printf("Error creating glamour renderer: %v", err)
           // Use a placeholder or default renderer if possible
       }

       m := model{
           textarea:       ta,
           viewport:       vp,
           renderer:       renderer,
           renderMarkdown: false, // Start in plain text mode (or true)
           keys:           defaultKeyMap, // Your keymap
           err:            err, // Store initial error if any
           // ... initialize other fields
       }

       // Perform initial render based on starting mode
       m.renderContent()

       return m
   }

   func (m model) Init() tea.Cmd {
       return textarea.Blink
   }
   ```

**3. Update Logic (`Update`):**
   Handle the toggle key and trigger re-renders.

   ```go
   func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
       var cmds []tea.Cmd
       var cmd tea.Cmd

       switch msg := msg.(type) {
       case tea.KeyMsg:
           switch {
           case key.Matches(msg, m.keys.ToggleMarkdown):
               m.renderMarkdown = !m.renderMarkdown // Flip the flag
               m.err = nil // Clear previous render error on mode switch
               m.renderContent() // Re-render the viewport content
               return m, nil // Return early, key handled

           // Handle other keybindings (Quit, Help, etc.)
           case key.Matches(msg, m.keys.Quit):
               return m, tea.Quit

           // Default: Forward keys to focused component (textarea or viewport)
           default:
               if m.textarea.Focused() {
                   m.textarea, cmd = m.textarea.Update(msg)
                   cmds = append(cmds, cmd)
                   // Re-render after textarea changes
                   m.renderContent()
               } else {
                   // Allow viewport scrolling etc. if it's focused
                   m.viewport, cmd = m.viewport.Update(msg)
                   cmds = append(cmds, cmd)
               }
           }

       case tea.WindowSizeMsg:
           m.width = msg.Width
           m.height = msg.Height
           // Recalculate heights for components
           // Example: headerHeight := 1; footerHeight := 1; textAreaHeight := 5
           // vpHeight := m.height - headerHeight - footerHeight - textAreaHeight
           m.viewport.Width = m.width
           m.viewport.Height = vpHeight // Use calculated height
           m.textarea.SetWidth(m.width)

           // IMPORTANT: Re-render content as width affects word wrap
           m.renderContent()

       // Handle other messages (e.g., custom messages, errors)
       }

       // Ensure viewport gets other updates too (like scrolling results)
       // It's often okay to update it even if textarea was focused,
       // as it might process messages like mouse events.
       if !m.textarea.Focused() { // Only if textarea didn't already update viewport
           m.viewport, cmd = m.viewport.Update(msg)
           cmds = append(cmds, cmd)
       }

       return m, tea.Batch(cmds...)
   }
   ```

**4. Rendering Logic (`renderContent` helper):**
   Create a helper function to centralize viewport updates.

   ```go
   func (m *model) renderContent() {
       if m.renderer == nil {
           m.viewport.SetContent("Error: Glamour renderer not initialized.")
           return
       }

       currentText := m.textarea.Value()

       if m.renderMarkdown {
           rendered, err := m.renderer.Render(currentText)
           if err != nil {
               m.err = err // Store the error
               // Display error nicely in viewport
               m.viewport.SetContent(fmt.Sprintf("Render Error:\n%v\n\n-- Raw Text --\n%s", err, currentText))
           } else {
               m.err = nil // Clear error on success
               m.viewport.SetContent(rendered)
           }
       } else {
           // Plain text mode
           m.err = nil // No render error possible here
           m.viewport.SetContent(currentText)
       }
       // Optional: Move viewport to bottom on update if desired
       // m.viewport.GotoBottom()
   }
   ```

**5. View Logic (`View`):**
   Assemble the UI components.

   ```go
   func (m model) View() string {
       // Use lipgloss or similar for layout
       status := fmt.Sprintf("Mode: %s", map[bool]string{true: "Markdown", false: "Plain Text"}[m.renderMarkdown])
       if m.err != nil {
           status += fmt.Sprintf(" | Error: %v", m.err)
       }

       return lipgloss.JoinVertical(lipgloss.Left,
           "My Markdown Editor", // Header
           m.viewport.View(),
           status, // Status line
           m.textarea.View(),
           // Footer / Help
       )
   }
   ```

This structure provides a robust way to toggle between raw text and Glamour-rendered markdown previews within a Bubble Tea application, addressing the pitfalls discovered during this investigation.

## 7. Addendum: Investigation of Similar Performance Issue in `bobatea`

**Date:** 2025-04-06 (Continued)
**Application:** `bobatea` (`bobatea/pkg/chat/conversation/model.go`)

Subsequent to the initial investigation, a similar performance degradation (multi-second delays) was observed during resize operations within the `bobatea` library's conversation component, which also utilizes `glamour` for markdown rendering.

The debugging approach mirrored the steps taken for `bubbletea-markdown-test`:

1.  **Initial Application of Fix:** The `Initialize once` pattern was applied by moving `glamour.NewTermRenderer` out of the frequently called `renderMessage` function and into the `NewModel` initializer. A new method, `SetWidth`, was introduced to handle terminal width changes.

2.  **Persistent Slowness:** Despite initializing the renderer only once, significant delays (up to ~13 seconds) were observed during calls to `SetWidth`, which was responsible for recreating the renderer with the new width using `glamour.WithWordWrap`.

3.  **Detailed Timing Logs:** Granular logging was added to `SetWidth` and its helper `getRendererContentWidth` to precisely measure the duration of operations:
    *   Logs confirmed that `getRendererContentWidth` (calculating frame sizes and padding) was extremely fast (~3 microseconds).
    *   Logs pinpointed the `glamour.NewTermRenderer` call *itself*, specifically when invoked with the `glamour.WithWordWrap` option inside `SetWidth`, as the source of the multi-second delay.

```log
# Example Log Snippet from bobatea debugging
...
2025-04-06T17:26:36.4243684-04:00 DBG getRendererContentWidth called
2025-04-06T17:26:36.424382961-04:00 DBG getRendererContentWidth finished duration=0.002883 ...
2025-04-06T17:26:43.691608511-04:00 DBG SetWidth: glamour.NewTermRenderer call finished duration=7267.234844 width=80
...
```

**Finding:** The performance issue is specifically tied to calling `glamour.NewTermRenderer` with the `glamour.WithWordWrap` option within the context of the `bobatea` application's `SetWidth` handler, likely triggered during terminal resize events handled by the underlying Bubble Tea framework. This contrasts with the initial standalone tests where `NewTermRenderer` (even with `WithWordWrap`) was fast, suggesting a potential interaction between Glamour's wrapping calculation and the terminal state or environment managed by Bubble Tea during resize events.

**Next Steps (Next Session):**
*   Re-evaluate the proposed workaround (render first without wrapping, then manually wrap the styled output).
*   Consider investigating potential upstream issues or interactions within Glamour or Bubble Tea related to terminal state during resizing. 