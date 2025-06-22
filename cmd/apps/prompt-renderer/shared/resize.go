package shared

import tea "github.com/charmbracelet/bubbletea"

// Resizable is an interface for models that adapt to terminal size changes.
type Resizable interface {
	// SetSize updates the model's width and height and returns any command to execute.
	SetSize(width, height int) tea.Cmd
}
