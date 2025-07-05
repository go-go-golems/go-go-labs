package ui

import (
    statepkg "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
    types "github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/types"
)

// Screen represents a screen in the application
type Screen interface {
    Render(state *types.ApplicationState) string
    HandleInput(key string, sm *statepkg.StateMachine) bool
}

// GetScreenForState returns the appropriate screen for the given state
func GetScreenForState(state statepkg.AppState) Screen {
    switch state {
    case statepkg.MainScreenState:
        return &MainScreen{}
    case statepkg.FilmSelectionState:
        return &FilmSelectionScreen{}
    case statepkg.EISelectionState:
        return &EISelectionScreen{}
    case statepkg.RollSelectionState:
        return &RollSelectionScreen{}
    case statepkg.MixedRollInputState:
        return &MixedRollInputScreen{}
    case statepkg.CalculatedScreenState:
        return &CalculatedScreen{}
    case statepkg.TimerScreenState:
        return &TimerScreen{}
    default:
        return &MainScreen{}
    }
} 