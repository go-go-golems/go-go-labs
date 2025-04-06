# Bug Report: Intermittent Multi-Second Hangs in Bubble Tea Apps Using Glamour/Termenv

**Libraries Involved:** `github.com/charmbracelet/glamour`, `github.com/muesli/termenv`, `github.com/charmbracelet/bubbletea`
**Environment:**
*   OS: Linux (e.g., Ubuntu 22.04)
*   Terminal: `xterm-kitty` (observed issue here, may affect others)
*   Go Version: (Specify Go version, e.g., 1.21.x)
*   Library Versions: (Specify versions if known, otherwise mention "latest" or "recent")

## 1. Summary

Interactive `bubbletea` applications using `glamour` for markdown rendering experience intermittent but significant hangs (5-10+ seconds) under certain conditions. The root cause has been traced to `termenv.HasDarkBackground()`, specifically the synchronous terminal query (`OSC 11`) used to detect the background color, which can block for multiple seconds waiting for a terminal response, especially when invoked during resize events or initial setup.

## 2. Problem Description

We observed two primary manifestations of this issue in different `bubbletea` applications:

**Scenario 1: Startup Hang (`bubbletea-markdown-test` app)**
*   A simple app using `textarea` and `viewport` with `glamour` rendering.
*   **Symptom:** ~5-second hang immediately after application launch before the UI becomes responsive.
*   **Initial Incorrect Diagnosis:** Thought `glamour.NewTermRenderer` was inherently slow.
*   **Actual Cause:** The renderer was being recreated inside the `tea.WindowSizeMsg` handler. Even though `NewTermRenderer` is fast in isolation (~4ms), recreating it during the rapid initial event processing caused the hang.
*   **Fix:** Initialize the `glamour.TermRenderer` **once** when the `model` is created and only update viewport/textarea dimensions in the `WindowSizeMsg` handler. This resolved the *startup* hang. (See Report: `ttmp/2025-04-06/03-report-on-hanging-markdown-rendering-in-bubbletea-applications.md`)

**Scenario 2: Resize Hang (`bobatea` library conversation view)**
*   A more complex component using `glamour` to render chat messages, adapting to terminal width changes.
*   **Symptom:** ~5-9 second hang *during terminal resize operations*.
*   **Setup:** Followed the fix from Scenario 1 (renderer initialized once). A `SetWidth` method was called on `tea.WindowSizeMsg` to update layout and re-render content. Initially, `SetWidth` recreated the renderer using `glamour.NewTermRenderer` with `glamour.WithAutoStyle()` and `glamour.WithWordWrap(newWidth)`.
*   **Investigation:** Extensive logging pinpointed the `glamour.NewTermRenderer` call *within* `SetWidth` as the source of the delay.
*   **Further Tracing:** The delay was isolated to the `glamour.WithAutoStyle()` option.

## 3. Root Cause Analysis: `termenv.HasDarkBackground()` Terminal Query

Detailed tracing with `zerolog` from `glamour.WithAutoStyle()` down into `termenv` revealed the exact source of the delay:

1.  `glamour.WithAutoStyle()` calls `getDefaultStyle("auto")`.
2.  `getDefaultStyle("auto")` calls `termenv.HasDarkBackground()` to determine the appropriate theme (dark/light).
3.  `termenv.HasDarkBackground()` calls `termenv.BackgroundColor()`.
4.  `termenv.BackgroundColor()` (specifically `termenv_unix.go:backgroundColor`) calls `termStatusReport(11)` to query the terminal's background color using an `OSC 11` escape sequence (`\x1b]11;?\x1b\\`).
5.  `termStatusReport` also sends a cursor position query (`CSI 6n`, `\x1b[6n`) and reads back *both* responses synchronously using `readNextResponse`.
6.  **The `readNextResponse` function, while waiting for the terminal's reply to the `OSC 11` query, blocks for several seconds.**

**Log Evidence (Timestamps from `xterm-kitty`):**

```log
# Call chain initiated by WithAutoStyle during resize/setup
6:12PM DBG ../thirdparty/termenv/output.go:217 > HasDarkBackground called
6:12PM DBG ../thirdparty/termenv/output.go:187 > BackgroundColor called cacheEnabled=false
...
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:142 > platform backgroundColor called
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:344 > termStatusReport called sequence=11

# --- Sending queries is fast ---
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:425 > termStatusReport: Finished writing OSC query writeDuration=0.000943
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:439 > termStatusReport: Finished writing CSI query writeDuration=0.000792

# --- Reading OSC response takes ~7.3 seconds ---
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:442 > termStatusReport: Reading first response (expecting OSC or CSI)
... (Many fast byte reads logged here) ...
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:323 > readNextResponse finished (OSC, terminated by ST) duration=7.349203 response="\x1b]11;rgb:2020/2020/2020\x1b\\"
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:444 > termStatusReport: First readNextResponse result isOSC=true response="\x1b]11;rgb:2020/2020/2020\x1b\\"

# --- Reading CSI response takes ~0.9 seconds ---
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:466 > termStatusReport: Reading second response (expecting CSI DSR)
... (Many fast byte reads logged here) ...
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:329 > readNextResponse finished (Cursor Position, terminated by R) duration=0.866391 response="\x1b[35;1R"
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:468 > termStatusReport: Second readNextResponse result (CSI expected) isOSC=false response="\x1b[35;1R"

# --- Total time dominated by terminal response delay ---
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:481 > termStatusReport finished successfully oscResponse="\x1b]11;rgb:2020/2020/2020\x1b\\" sequence=11 totalDuration=8.559313
6:12PM DBG ../thirdparty/termenv/termenv_unix.go:152 > platform backgroundColor finished (from OSC) color=#202020 duration=8.643824
6:12PM DBG ../thirdparty/termenv/output.go:199 > BackgroundColor: Platform call finished color=#202020 platformCallDuration=8.669684
6:12PM DBG ../thirdparty/termenv/output.go:222 > HasDarkBackground finished ... duration=8.751599 hasDarkBackground=true ...
6:12PM DBG ../thirdparty/glamour/glamour.go:122 > Option applied successfully duration=8.793006 optionIndex=0
```

**Key Observation:** The Go code (`select`, `read`) is fast *once data is available*. The delay is the time the terminal takes to respond to the `OSC 11` query. This seems highly dependent on the terminal emulator and potentially its current state or load.

## 4. Impact

Calling `glamour.WithAutoStyle()` (and thus `termenv.HasDarkBackground()`) in any frequently-called code path within an interactive application (like event handlers for resize, key presses, or even initial setup) can lead to significant, unpredictable UI freezes, making the application feel unresponsive and broken.

## 5. Workaround

The effective workaround is to avoid calling `termenv.HasDarkBackground()` frequently.

1.  Call `termenv.HasDarkBackground()` **once** during application initialization.
2.  Store the result (e.g., `"dark"` or `"light"` style name).
3.  When creating/recreating `glamour` renderers, use `glamour.WithStandardStyle(cachedStyleName)` instead of `glamour.WithAutoStyle()`.

```go
// Example in Bubble Tea Model
type model struct {
    renderer *glamour.TermRenderer
    determinedStyle string
    // ...
}

func NewModel() model {
    // ...
    style := "light" // Default
    if termenv.HasDarkBackground() { // Called ONCE
         style = "dark"
    }
    m.determinedStyle = style

    // Create initial renderer using determined style
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle(m.determinedStyle),
        // ... other options
    )
    m.renderer = renderer
    // ...
}

// Example in a resize handler
func (m *Model) SetWidth(width int) {
    // Recreate renderer if needed (e.g., for word wrap)
    // Use the CACHED style name
    renderer, _ := glamour.NewTermRenderer(
        glamour.WithStandardStyle(m.determinedStyle),
        glamour.WithWordWrap(newWidth),
    )
    m.renderer = renderer
    // ...
}
```

This ensures the potentially blocking terminal query only happens at startup.

## 6. Suggestions for Library Maintainers

1.  **`termenv` Timeout/Context:** Add a configurable timeout (perhaps via `termenv.OutputOption`) or `context.Context` support to functions like `BackgroundColor()` / `ForegroundColor()` / `termStatusReport`. This would allow consuming applications to decide how long they are willing to wait for a terminal response. A short timeout (e.g., 100-250ms) could default to a standard color if the terminal is unresponsive.
2.  **`termenv` Caching:** Consider adding an optional caching layer within `termenv` itself for background/foreground colors. The cache could be simple (first call wins) or have a short TTL.
3.  **Documentation:** Explicitly document in `termenv` and `glamour` that `termenv.HasDarkBackground()` and `glamour.WithAutoStyle()` perform synchronous terminal I/O and can potentially block for multiple seconds, recommending they not be used in performance-sensitive code paths like event loops without caching the result.
4.  **Alternative Queries (`termenv`):** Investigate if less common but potentially faster/more reliable terminal query methods exist for background color detection, perhaps as fallback options.
5.  **Error Handling (`termenv`):** The current `OSCTimeout` (5s) in `waitForData` returns a generic `fmt.Errorf("timeout")`. Returning a more specific, potentially exported error variable (e.g., `termenv.ErrTimeout`) would allow callers to handle timeouts more gracefully (e.g., by falling back to a default color).

Thank you for maintaining these valuable libraries. We hope this detailed report helps in addressing this performance issue.