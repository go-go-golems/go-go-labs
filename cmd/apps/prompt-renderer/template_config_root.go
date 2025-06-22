package main

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/go-go-golems/go-go-labs/cmd/apps/prompt-renderer/shared"
)

const (
	statusHeight = 1
	headerHeight = 3 // title + two blank lines
)

// RootConfigModel composes form, preview, and status submodels
// and manages state updates and size propagation.
type RootConfigModel struct {
	template     *TemplateDefinition
	renderer     *PromptRenderer
	stateManager *StateManager

	form    *FormModel
	preview *PreviewModel
	status  *StatusBarModel

	width  int
	height int
}

// NewRootConfigModel creates and initializes all submodels.
func NewRootConfigModel(template *TemplateDefinition, renderer *PromptRenderer) *RootConfigModel {
	// core state manager
	sm := NewStateManager(template, renderer)
	// submodels
	form := NewFormModel(template, sm.Selection)
	preview := NewPreviewModel(sm.Preview)
	status := NewStatusBarModel()
	return &RootConfigModel{
		template:     template,
		renderer:     renderer,
		stateManager: sm,
		form:         form,
		preview:      preview,
		status:       status,
	}
}

// Init runs initial commands for all submodels.
func (m *RootConfigModel) Init() tea.Cmd {
	return tea.Batch(
		m.form.Init(),
		m.preview.Init(),
		m.status.Init(),
	)
}

// Update handles all incoming messages, routing to submodels or updating state.
func (m *RootConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "c":
			return m, func() tea.Msg {
				return CopyPromptMsg{Prompt: m.stateManager.Preview}
			}
		case "s":
			return m, func() tea.Msg {
				return SaveSelectionMsg{Selection: m.stateManager.Selection}
			}
		case "esc", "left":
			return m, func() tea.Msg {
				return GoBackMsg{}
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		leftW := m.width / 2
		rightW := m.width - leftW
		contentH := m.height - headerHeight - statusHeight
		// propagate sizes
		return m, tea.Batch(
			m.form.SetSize(leftW, contentH),
			m.preview.SetSize(rightW, contentH),
			m.status.SetSize(m.width, statusHeight),
		)

	case shared.ToggleChangedMsg:
		// toggle or bullet changed
		if msg.BulletIndex != nil {
			// bullet index provided
			key := strconv.Itoa(*msg.BulletIndex)
			m.stateManager.ToggleBullet(msg.SectionID, msg.VariantID, key)
		} else {
			m.stateManager.ToggleVariant(msg.SectionID, msg.VariantID)
		}
		// rebuild form items to reflect new selection
		m.form.handler.RebuildFormItems(m.template, m.stateManager.Selection)
		// update preview
		pm, cmd := m.preview.Update(shared.PreviewUpdatedMsg{Text: m.stateManager.Preview})
		m.preview = pm.(*PreviewModel)
		return m, cmd

	case shared.SectionVariantMsg:
		// cycle through section variants
		m.stateManager.CycleSectionVariant(msg.SectionID)
		// rebuild form items after cycle
		m.form.handler.RebuildFormItems(m.template, m.stateManager.Selection)
		// update preview
		pm, cmd := m.preview.Update(shared.PreviewUpdatedMsg{Text: m.stateManager.Preview})
		m.preview = pm.(*PreviewModel)
		return m, cmd

	case shared.VarChangedMsg:
		// variable value changed in form
		m.stateManager.UpdateVariable(msg.Name, msg.Value)
		// rebuild form items to update variable display
		m.form.handler.RebuildFormItems(m.template, m.stateManager.Selection)
		// update preview submodel
		pm, cmd := m.preview.Update(shared.PreviewUpdatedMsg{Text: m.stateManager.Preview})
		m.preview = pm.(*PreviewModel)
		return m, cmd
	}

	// default: delegate to submodels
	var cmds []tea.Cmd

	// form input
	fm, cmd := m.form.Update(msg)
	m.form = fm.(*FormModel)
	cmds = append(cmds, cmd)

	// preview updates only on PreviewUpdatedMsg or resize
	pm, cmd := m.preview.Update(msg)
	m.preview = pm.(*PreviewModel)
	cmds = append(cmds, cmd)

	// status bar
	sm, cmd := m.status.Update(msg)
	m.status = sm.(*StatusBarModel)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View composes the full configuration screen.
func (m *RootConfigModel) View() string {
	var b strings.Builder

	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Width(m.width)
	title := titleStyle.Render(m.template.Label)
	b.WriteString(title)
	b.WriteString("\n\n")

	// Main content
	main := lipgloss.JoinHorizontal(lipgloss.Top,
		m.form.View(),
		m.preview.View(),
	)
	b.WriteString(main)
	b.WriteString("\n")

	// Status bar
	b.WriteString(m.status.View())

	return b.String()
}

// SetSize allows external size setting (e.g., from parent) and propagates.
func (m *RootConfigModel) SetSize(width, height int) tea.Cmd {
	_, cmd := m.Update(tea.WindowSizeMsg{Width: width, Height: height})
	return cmd
}
