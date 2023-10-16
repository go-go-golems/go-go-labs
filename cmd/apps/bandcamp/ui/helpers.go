package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"time"
)

func ClearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return ClearErrorMsg{}
	})
}
