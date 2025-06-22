package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/go-go-labs/cmd/apps/prompt-renderer/shared"
)

// Ensure FormModel implements the Resizable interface
var _ shared.Resizable = (*FormModel)(nil)

// FormModel handles the variables and sections form as a nested Bubble Tea model.
type FormModel struct {
	template *TemplateDefinition
	handler  *FormHandler
	input    *InputHandler
	ui       *UIRenderer
	width    int
	height   int
}

// NewFormModel creates a new form model based on the template and current selection.
func NewFormModel(template *TemplateDefinition, selection *SelectionState) *FormModel {
	h := NewFormHandler()
	h.RebuildFormItems(template, selection)
	i := NewInputHandler()
	ui := NewUIRenderer()
	return &FormModel{template: template, handler: h, input: i, ui: ui}
}

// Init implements tea.Model.Init
func (m *FormModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.Update
func (m *FormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m, m.SetSize(msg.Width, msg.Height)
	case tea.KeyMsg:
		key := msg.String()
		if m.input.IsEditing() {
			// editing a variable
			switch key {
			case "esc":
				m.input.StopEditing()
				return m, nil
			case "enter":
				value := m.input.CommitEdit()
				if item := m.handler.GetFocusedItem(); item != nil && item.Type == "variable" {
					m.handler.UpdateItemValue(value)
					return m, func() tea.Msg {
						return shared.VarChangedMsg{Name: item.Key, Value: value}
					}
				}
				return m, nil
			default:
				m.input.HandleEditingKeypress(msg)
				return m, nil
			}
		}
		// normal (non-editing) input
		switch key {
		case "up", "k":
			m.handler.NavigateUp()
			return m, nil
		case "down", "j":
			m.handler.NavigateDown()
			return m, nil
		case "tab":
			m.handler.NavigateNext()
			return m, nil
		case "shift+tab":
			m.handler.NavigatePrev()
			return m, nil
		case "esc", "left":
			return m, func() tea.Msg { return GoBackMsg{} }
		case "enter":
			if item := m.handler.GetFocusedItem(); item != nil {
				switch item.Type {
				case "variable":
					m.input.StartEditing(item.Value)
					return m, nil
				case "section":
					return m, func() tea.Msg { return shared.SectionVariantMsg{SectionID: item.Key} }
				case "bullet_header":
					return m, nil
				case "toggle":
					return m, func() tea.Msg {
						return shared.ToggleChangedMsg{SectionID: item.SectionID, VariantID: item.VariantID, BulletIndex: nil}
					}
				case "bullet":
					idx := item.BulletIndex
					return m, func() tea.Msg {
						return shared.ToggleChangedMsg{SectionID: item.SectionID, VariantID: item.VariantID, BulletIndex: &idx}
					}
				}
			}
		case "space":
			if item := m.handler.GetFocusedItem(); item != nil && item.Type != "bullet_header" && item.Type != "variable" && item.Type != "section" {
				if item.Type == "toggle" {
					// optimistic UI update
					item.Selected = !item.Selected
					m.handler.Items[m.handler.FocusIndex] = *item
					return m, func() tea.Msg {
						return shared.ToggleChangedMsg{SectionID: item.SectionID, VariantID: item.VariantID, BulletIndex: nil}
					}
				} else if item.Type == "bullet" {
					idx := item.BulletIndex
					// optimistic update
					item.Selected = !item.Selected
					m.handler.Items[m.handler.FocusIndex] = *item
					return m, func() tea.Msg {
						return shared.ToggleChangedMsg{SectionID: item.SectionID, VariantID: item.VariantID, BulletIndex: &idx}
					}
				}
			}
			return m, nil
		}
	}
	return m, nil
}

// View implements tea.Model.View by rendering form items via UIRenderer.
func (m *FormModel) View() string {
	return m.ui.renderFormItems(m.template, m.handler, m.input, m.width, m.height)
}

// SetSize implements shared.Resizable
func (m *FormModel) SetSize(width, height int) tea.Cmd {
	m.width = width
	m.height = height
	return nil
} 