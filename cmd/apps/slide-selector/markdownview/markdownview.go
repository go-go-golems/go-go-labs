package markdownview

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
)

// ViewMode defines the mode of the MarkdownView
type ViewMode int

const (
	Static     ViewMode = iota // No scrolling
	Scrollable                 // Enable scrolling
)

// MarkdownView represents the model for a markdown component
type MarkdownView struct {
	Content      string
	Rendered     string
	GlamourStyle string
	ViewMode     ViewMode
	Renderer     *glamour.TermRenderer
	Styles       lipgloss.Style
	Viewport     viewport.Model
	NeedsRender  bool
}

// NewMarkdownView initializes a new MarkdownView
func NewMarkdownView(content string, glamourStyle string, viewMode ViewMode) (*MarkdownView, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(glamourStyle),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, err
	}

	rendered, err := renderer.Render(content)
	if err != nil {
		return nil, err
	}

	vp := viewport.New(0, 0)
	vp.SetContent(rendered)

	styles := lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

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
func (m *MarkdownView) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model accordingly
func (m *MarkdownView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg.Width, msg.Height)

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}

	if m.ViewMode == Scrollable {
		m.Viewport, cmd = m.Viewport.Update(msg)
		m.NeedsRender = true
	}

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
	switch m.ViewMode {
	case Scrollable:
		return m.Styles.Render(m.Viewport.View())
	case Static:
		return m.Styles.Render(m.Rendered)
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

// Resize adjusts the component based on new terminal dimensions
func (m *MarkdownView) Resize(width, height int) {
	if m.ViewMode == Scrollable {
		// Calculate border width
		borderWidth := m.Styles.GetBorderLeftSize() + m.Styles.GetBorderRightSize()
		borderHeight := m.Styles.GetBorderTopSize() + m.Styles.GetBorderBottomSize()

		// Adjust viewport size considering border and padding
		m.Viewport.Width = width - borderWidth - m.Styles.GetHorizontalFrameSize()
		m.Viewport.Height = height - borderHeight - m.Styles.GetVerticalFrameSize()
	}
	m.NeedsRender = true
}
