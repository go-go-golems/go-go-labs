package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

// InputHandler manages keyboard input and editing state
type InputHandler struct {
	Editing      bool
	EditingValue string
}

// NewInputHandler creates a new input handler
func NewInputHandler() *InputHandler {
	return &InputHandler{
		Editing:      false,
		EditingValue: "",
	}
}

// StartEditing begins editing with the current value
func (i *InputHandler) StartEditing(currentValue string) {
	i.Editing = true
	i.EditingValue = currentValue
}

// StopEditing cancels editing mode
func (i *InputHandler) StopEditing() {
	i.Editing = false
	i.EditingValue = ""
}

// CommitEdit returns the edited value and stops editing
func (i *InputHandler) CommitEdit() string {
	value := i.EditingValue
	i.Editing = false
	i.EditingValue = ""
	return value
}

// HandleEditingKeypress processes keystrokes during editing
func (i *InputHandler) HandleEditingKeypress(msg tea.KeyMsg) {
	switch msg.String() {
	case "backspace":
		if len(i.EditingValue) > 0 {
			i.EditingValue = i.EditingValue[:len(i.EditingValue)-1]
		}
	default:
		if len(msg.String()) == 1 {
			i.EditingValue += msg.String()
		}
	}
}

// GetDisplayValue returns the value to display (with cursor if editing)
func (i *InputHandler) GetDisplayValue(originalValue string) string {
	if i.Editing {
		return i.EditingValue + "█" // Show cursor
	}
	return originalValue
}

// IsEditing returns whether the handler is in editing mode
func (i *InputHandler) IsEditing() bool {
	return i.Editing
}
