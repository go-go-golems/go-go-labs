package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	logger     zerolog.Logger
	logFile    *os.File
	logLevel   string
	renderMd   bool
	showHelp   bool
	initialMsg string
)

type keyMap struct {
	ToggleHelp     key.Binding
	ToggleMarkdown key.Binding
	Quit           key.Binding
}

// Default key bindings
var defaultKeyMap = keyMap{
	ToggleHelp: key.NewBinding(
		key.WithKeys("ctrl+h"),
		key.WithHelp("ctrl+h", "toggle help"),
	),
	ToggleMarkdown: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "toggle markdown"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("ctrl+c/esc", "quit"),
	),
}

type model struct {
	viewport       viewport.Model
	textarea       textarea.Model
	renderer       *glamour.TermRenderer
	width          int
	height         int
	err            error
	renderMarkdown bool
	showHelp       bool
	keys           keyMap
}

func setupLogging(level string) error {
	// Create log file
	var err error
	logFile, err = os.OpenFile("/tmp/external.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Parse log level
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return fmt.Errorf("invalid log level: %w", err)
	}

	// Configure zerolog
	zerolog.TimeFieldFormat = time.RFC3339
	consoleWriter := zerolog.ConsoleWriter{Out: logFile, TimeFormat: time.RFC3339}
	logger = zerolog.New(consoleWriter).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()

	log.Logger = logger

	logger.Info().Str("level", level).Msg("Logging initialized")
	return nil
}

func logWithCaller(level zerolog.Level, msg string, fields map[string]interface{}) {
	// Get caller information
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		file = "unknown"
		line = 0
	}

	// Create log event based on level
	var event *zerolog.Event
	switch level {
	case zerolog.DebugLevel:
		event = logger.Debug()
	case zerolog.InfoLevel:
		event = logger.Info()
	case zerolog.WarnLevel:
		event = logger.Warn()
	case zerolog.ErrorLevel:
		event = logger.Error()
	default:
		event = logger.Info()
	}

	// Add fields
	event.Str("file", file).Int("line", line)
	for k, v := range fields {
		switch val := v.(type) {
		case string:
			event.Str(k, val)
		case int:
			event.Int(k, val)
		case bool:
			event.Bool(k, val)
		default:
			event.Interface(k, v)
		}
	}

	event.Msg(msg)
}

func initialModel() model {
	logWithCaller(zerolog.InfoLevel, "Initializing model", nil)

	// Create and configure textarea
	ta := textarea.New()
	ta.Placeholder = "Enter markdown here..."
	ta.SetWidth(80)
	ta.SetHeight(5)
	ta.Focus()

	if initialMsg != "" {
		ta.SetValue(initialMsg)
		logWithCaller(zerolog.InfoLevel, "Set initial textarea value", map[string]interface{}{
			"value": initialMsg,
		})
	}

	// Initial viewport with empty content
	vp := viewport.New(80, 20)
	vp.SetContent("")

	// Create glamour renderer
	logWithCaller(zerolog.InfoLevel, "Creating glamour renderer", nil)
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		logWithCaller(zerolog.ErrorLevel, "Failed to create glamour renderer", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}

	m := model{
		textarea:       ta,
		viewport:       vp,
		renderer:       renderer,
		width:          80,
		height:         25,
		renderMarkdown: renderMd,
		showHelp:       showHelp,
		keys:           defaultKeyMap,
	}

	// Initial render of content
	logWithCaller(zerolog.InfoLevel, "Performing initial render", map[string]interface{}{
		"renderMarkdown": m.renderMarkdown,
	})
	m.renderContent()

	return m
}

func (m model) Init() tea.Cmd {
	logWithCaller(zerolog.InfoLevel, "Initializing tea model", nil)
	return textarea.Blink
}

// Save rendered content to file
func (m model) saveRenderedContent() {
	if content := m.viewport.View(); content != "" {
		err := os.WriteFile("/tmp/rendered.md", []byte(content), 0644)
		if err != nil {
			logWithCaller(zerolog.ErrorLevel, "Failed to save rendered content", map[string]interface{}{
				"error": err.Error(),
			})
			m.err = err
		} else {
			logWithCaller(zerolog.DebugLevel, "Saved rendered content", map[string]interface{}{
				"size": len(content),
			})
		}
	}
}

// Render content based on current mode
func (m *model) renderContent() {
	logWithCaller(zerolog.DebugLevel, "Rendering content", map[string]interface{}{
		"renderMarkdown": m.renderMarkdown,
		"textLength":     len(m.textarea.Value()),
	})

	if strings.TrimSpace(m.textarea.Value()) == "" {
		logWithCaller(zerolog.DebugLevel, "Empty content, setting viewport to empty", nil)
		m.viewport.SetContent("")
		return
	}

	var content string
	if m.renderMarkdown {
		// Render as markdown
		logWithCaller(zerolog.DebugLevel, "Rendering markdown", nil)
		renderedContent, err := m.renderer.Render(m.textarea.Value())
		if err != nil {
			logWithCaller(zerolog.ErrorLevel, "Error rendering markdown", map[string]interface{}{
				"error": err.Error(),
			})
			m.err = err
			content = fmt.Sprintf("Render Error:\n%s\n\n%s", err.Error(), m.textarea.Value())
		} else {
			logWithCaller(zerolog.DebugLevel, "Markdown rendered successfully", map[string]interface{}{
				"renderedLength": len(renderedContent),
			})
			m.err = nil
			content = renderedContent
		}
	} else {
		// Show as plain text
		logWithCaller(zerolog.DebugLevel, "Using plain text mode", nil)
		m.err = nil
		content = m.textarea.Value()
	}

	m.viewport.SetContent(content)
	m.saveRenderedContent()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	logWithCaller(zerolog.DebugLevel, "Update called", map[string]interface{}{
		"msgType": fmt.Sprintf("%T", msg),
	})

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()
		logWithCaller(zerolog.DebugLevel, "Key pressed", map[string]interface{}{
			"key": keyStr,
		})

		switch {
		case key.Matches(msg, m.keys.Quit):
			logWithCaller(zerolog.InfoLevel, "Quitting application", nil)
			return m, tea.Quit
		case key.Matches(msg, m.keys.ToggleMarkdown):
			// Toggle markdown rendering
			m.renderMarkdown = !m.renderMarkdown
			logWithCaller(zerolog.InfoLevel, "Toggled markdown rendering", map[string]interface{}{
				"renderMarkdown": m.renderMarkdown,
			})
			m.renderContent()
			return m, nil
		case key.Matches(msg, m.keys.ToggleHelp):
			// Toggle help
			m.showHelp = !m.showHelp
			logWithCaller(zerolog.InfoLevel, "Toggled help display", map[string]interface{}{
				"showHelp": m.showHelp,
			})
			return m, nil
		default:
			// Handle textarea input only if not matching other keys
			// Check if the textarea is focused before updating
			if m.textarea.Focused() {
				logWithCaller(zerolog.DebugLevel, "Updating textarea", nil)
				m.textarea, cmd = m.textarea.Update(msg)
				cmds = append(cmds, cmd)
				// Re-render content after textarea update
				m.renderContent()
			} else {
				// If textarea isn't focused, keys might be for viewport scrolling
				// We handle viewport updates later anyway
			}
		}

	case tea.WindowSizeMsg:
		logWithCaller(zerolog.DebugLevel, "Window size changed", map[string]interface{}{
			"width":  msg.Width,
			"height": msg.Height,
		})
		m.width = msg.Width
		m.height = msg.Height

		// Calculate proportional heights
		helpHeight := 0
		if m.showHelp {
			helpHeight = 3 // 3 lines for help
		}
		statusHeight := 1                                                           // 1 line for status
		textAreaHeight := 6                                                         // 5 lines + 1 for border
		viewportHeight := m.height - textAreaHeight - helpHeight - statusHeight - 2 // 2 for margins and dividers

		logWithCaller(zerolog.DebugLevel, "Calculated component heights", map[string]interface{}{
			"helpHeight":     helpHeight,
			"statusHeight":   statusHeight,
			"textAreaHeight": textAreaHeight,
			"viewportHeight": viewportHeight,
			"showHelp":       m.showHelp,
		})

		// Resize viewport
		m.viewport.Width = m.width
		m.viewport.Height = viewportHeight

		// Renderer is initialized once, just re-render content if needed
		// based on the new size affecting layout, but glamour handles wrap internally mostly.
		m.renderContent() // Re-render in case width change affects layout

		// Resize textarea
		m.textarea.SetWidth(m.width)
		m.textarea.SetHeight(5)
	}

	// Also update viewport for scrolling
	logWithCaller(zerolog.DebugLevel, "Updating viewport", nil)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	logWithCaller(zerolog.DebugLevel, "View called", nil)

	// Define styles
	viewportStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	textareaStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(0, 1)

	// Create a divider
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render("─" + strings.Repeat("─", m.width-2) + "─")

	// Create status display
	mode := "Plain Text"
	if m.renderMarkdown {
		mode = "Markdown"
	}
	contentLength := len(m.viewport.View())
	linesCount := len(strings.Split(m.viewport.View(), "\n"))
	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("86")).
		Render(fmt.Sprintf("Mode: %s | Content size: %d chars, %d lines", mode, contentLength, linesCount))

	// Create help display
	help := ""
	if m.showHelp {
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

		helpItems := []string{
			m.keys.ToggleHelp.Help().Key + ": " + m.keys.ToggleHelp.Help().Desc,
			m.keys.ToggleMarkdown.Help().Key + ": " + m.keys.ToggleMarkdown.Help().Desc,
			m.keys.Quit.Help().Key + ": " + m.keys.Quit.Help().Desc,
		}
		help = helpStyle.Render(strings.Join(helpItems, " | "))
		help = lipgloss.JoinVertical(lipgloss.Left, help, divider)
	}

	// Create error display if needed
	errorDisplay := ""
	if m.err != nil {
		errorDisplay = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Combine components
	return lipgloss.JoinVertical(lipgloss.Left,
		help,
		viewportStyle.Render(m.viewport.View()),
		status,
		divider,
		textareaStyle.Render(m.textarea.View()),
		errorDisplay,
	)
}

func runApp(cmd *cobra.Command, args []string) error {
	if err := setupLogging(logLevel); err != nil {
		return err
	}
	defer logFile.Close()

	logWithCaller(zerolog.InfoLevel, "Starting application", map[string]interface{}{
		"renderMarkdown": renderMd,
		"showHelp":       showHelp,
		"initialMsg":     initialMsg,
	})

	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	logWithCaller(zerolog.InfoLevel, "Running tea program", nil)
	if _, err := p.Run(); err != nil {
		logWithCaller(zerolog.ErrorLevel, "Error running program", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	logWithCaller(zerolog.InfoLevel, "Application exiting normally", nil)
	return nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "markdown-test",
		Short: "A TUI markdown renderer test application",
		RunE:  runApp,
	}

	// Flags
	rootCmd.Flags().StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVar(&renderMd, "render-markdown", false, "Start with markdown rendering enabled")
	rootCmd.Flags().BoolVar(&showHelp, "show-help", true, "Start with help visible")
	rootCmd.Flags().StringVar(&initialMsg, "initial-text", "", "Initial text to show in editor")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
