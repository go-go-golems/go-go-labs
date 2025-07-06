package ui

import (
	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
)

// Screen represents a screen in the application
type Screen interface {
	Render(state *state.ApplicationState) string
	HandleInput(key string, sm *state.StateMachine) bool
}

// GetScreenForState returns the appropriate screen for the given state
func GetScreenForState(appState state.AppState) Screen {
	switch appState {
	case state.MainScreenState:
		return &MainScreen{}
	case state.FilmSelectionState:
		return &FilmSelectionScreen{}
	case state.EISelectionState:
		return &EISelectionScreen{}
	case state.RollSelectionState:
		return &RollSelectionScreen{}
	case state.MixedRollInputState:
		return &MixedRollInputScreen{}
	case state.CalculatedScreenState:
		return &CalculatedScreen{}
	case state.TimerScreenState:
		return &TimerScreen{}
	default:
		return &MainScreen{}
	}
} 