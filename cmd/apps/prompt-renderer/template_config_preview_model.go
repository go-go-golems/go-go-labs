package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/prompt-renderer/shared"
)

// Ensure PreviewModel implements the Resizable interface
var _ shared.Resizable = (*PreviewModel)(nil)

// PreviewModel handles the live prompt preview as a nested Bubble Tea model.
type PreviewModel struct {
	ui      *UIRenderer
	content string
	width   int
	height  int
}

// NewPreviewModel creates a preview model with the initial preview text.
func NewPreviewModel(initial string) *PreviewModel {
	ui := NewUIRenderer()
	return &PreviewModel{ui: ui, content: initial}
}

// Init implements tea.Model.Init
func (m *PreviewModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update
func (m *PreviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case shared.PreviewUpdatedMsg:
		m.content = msg.Text
		return m, nil
	case tea.WindowSizeMsg:
		return m, m.SetSize(msg.Width, msg.Height)
	}
	return m, nil
}

// View implements tea.Model.View by rendering the preview section.
func (m *PreviewModel) View() string {
	// renderPreview is unexported but available within main package
	return m.ui.renderPreview(m.content, m.width, m.height)
}

// SetSize implements shared.Resizable
func (m *PreviewModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	m.ui.SetSize(width, height)
	return nil
} 