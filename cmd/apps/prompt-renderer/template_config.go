package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TemplateConfigModel handles template configuration screen using submodels
type TemplateConfigModel struct {
	formHandler  *FormHandler
	inputHandler *InputHandler
	stateManager *StateManager
	uiRenderer   *UIRenderer
	toastMessage string
	toastExpiry  time.Time
}

// NewTemplateConfigModel creates a new template config model using submodels
func NewTemplateConfigModel(template *TemplateDefinition, renderer *PromptRenderer) *TemplateConfigModel {
	m := &TemplateConfigModel{
		formHandler:  NewFormHandler(),
		inputHandler: NewInputHandler(),
		stateManager: NewStateManager(template, renderer),
		uiRenderer:   NewUIRenderer(),
	}

	// Initialize form with template data
	m.formHandler.RebuildFormItems(template, m.stateManager.Selection)

	return m
}

// Init implements tea.Model
func (m *TemplateConfigModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (m *TemplateConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.inputHandler.IsEditing() {
			return m.handleEditingInput(msg)
		}
		return m.handleNormalInput(msg)
	case tea.WindowSizeMsg:
		m.uiRenderer.SetSize(msg.Width, msg.Height)
	case CopyDoneMsg:
		m.showToast("✓ Copied to clipboard!")
		return m, tea.Tick(750*time.Millisecond, func(time.Time) tea.Msg {
			return ClearToastMsg{}
		})
	case ClearToastMsg:
		m.toastMessage = ""
	}

	return m, nil
}

// handleNormalInput handles input when not in editing mode
func (m *TemplateConfigModel) handleNormalInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc", "left":
		return m, func() tea.Msg {
			return GoBackMsg{}
		}
	case "up", "k":
		m.formHandler.NavigateUp()
	case "down", "j":
		m.formHandler.NavigateDown()
	case "tab":
		m.formHandler.NavigateNext()
	case "shift+tab":
		m.formHandler.NavigatePrev()
	case "space":
		if item := m.formHandler.GetFocusedItem(); item != nil {
			if item.Type != "bullet_header" && item.Type != "variable" && item.Type != "section" {
				m.handleToggle()
			}
		}
	case "enter":
		if item := m.formHandler.GetFocusedItem(); item != nil {
			switch item.Type {
			case "variable":
				m.inputHandler.StartEditing(item.Value)
			case "section":
				m.stateManager.CycleSectionVariant(item.Key)
				m.formHandler.RebuildFormItems(m.stateManager.Template, m.stateManager.Selection)
			case "bullet_header":
				// Headers are not interactive, skip
			default:
				// For bullet and toggle items, Enter should also toggle
				m.handleToggle()
			}
		}
	case "c":
		return m, func() tea.Msg {
			return CopyPromptMsg{Prompt: m.stateManager.Preview}
		}
	case "s":
		return m, func() tea.Msg {
			return SaveSelectionMsg{Selection: m.stateManager.Selection}
		}
	case "ctrl+r":
		m.stateManager.UpdatePreview()
	case "?":
		return m, func() tea.Msg {
			return ShowHelpMsg{}
		}
	}

	return m, nil
}

// handleEditingInput handles input when in editing mode
func (m *TemplateConfigModel) handleEditingInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.inputHandler.StopEditing()
	case "enter":
		value := m.inputHandler.CommitEdit()
		if item := m.formHandler.GetFocusedItem(); item != nil && item.Type == "variable" {
			m.stateManager.UpdateVariable(item.Key, value)
			m.formHandler.UpdateItemValue(value)
		}
	default:
		m.inputHandler.HandleEditingKeypress(msg)
	}

	return m, nil
}

// handleToggle handles toggling of bullets and toggle items
func (m *TemplateConfigModel) handleToggle() {
	if item := m.formHandler.GetFocusedItem(); item != nil {
		m.stateManager.HandleToggle(item)
		m.formHandler.RebuildFormItems(m.stateManager.Template, m.stateManager.Selection)
	}
}

// showToast displays a temporary message
func (m *TemplateConfigModel) showToast(message string) {
	m.toastMessage = message
	m.toastExpiry = time.Now().Add(750 * time.Millisecond)
}

// View implements tea.Model
func (m *TemplateConfigModel) View() string {
	return m.uiRenderer.RenderView(
		m.stateManager.Template,
		m.formHandler,
		m.inputHandler,
		m.stateManager.Preview,
		m.toastMessage,
		m.toastExpiry,
	)
}

// Expose properties and methods for compatibility with main.go

// SetSize updates the model dimensions
func (m *TemplateConfigModel) SetSize(width, height int) {
	m.uiRenderer.SetSize(width, height)
}

// GetSelection returns the current selection state
func (m *TemplateConfigModel) GetSelection() *SelectionState {
	return m.stateManager.Selection
}

// RebuildFormItems rebuilds the form items
func (m *TemplateConfigModel) RebuildFormItems() {
	m.formHandler.RebuildFormItems(m.stateManager.Template, m.stateManager.Selection)
}

// UpdatePreview updates the preview
func (m *TemplateConfigModel) UpdatePreview() {
	m.stateManager.UpdatePreview()
}

// Properties for compatibility
func (m *TemplateConfigModel) Width() int {
	return m.uiRenderer.Width
}

func (m *TemplateConfigModel) Height() int {
	return m.uiRenderer.Height
}

// Message types
type CopyPromptMsg struct {
	Prompt string
}

type CopyDoneMsg struct{}

type ClearToastMsg struct{}

type SaveSelectionMsg struct {
	Selection *SelectionState
}

type GoBackMsg struct{}

// ShowHelpMsg is already declared in template_list.go

// Methods for testing

// GetFormItems returns the form items for testing
func (m *TemplateConfigModel) GetFormItems() []FormItem {
	return m.formHandler.Items
}

// GetFocusIndex returns the current focus index for testing
func (m *TemplateConfigModel) GetFocusIndex() int {
	return m.formHandler.FocusIndex
}

// SetFocusIndex sets the focus index for testing
func (m *TemplateConfigModel) SetFocusIndex(index int) {
	m.formHandler.FocusIndex = index
}

// HandleToggle exposes toggle functionality for testing
func (m *TemplateConfigModel) HandleToggle() {
	m.handleToggle()
}
