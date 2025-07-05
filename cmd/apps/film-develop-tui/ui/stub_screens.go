package ui

import (
	"strings"

	"github.com/go-go-golems/go-go-labs/cmd/apps/film-develop-tui/state"
)

// EISelectionScreen stub
type EISelectionScreen struct{}

func (s *EISelectionScreen) Render(appState *state.ApplicationState) string {
	return TitleStyle.Render("🎞️  EI Selection - TODO")
}

func (s *EISelectionScreen) HandleInput(key string, sm *state.StateMachine) bool {
	switch strings.ToLower(key) {
	case "esc":
		sm.GoBack()
		return true
	case "q":
		return false
	}
	return true
}

// RollSelectionScreen stub
type RollSelectionScreen struct{}

func (s *RollSelectionScreen) Render(appState *state.ApplicationState) string {
	return TitleStyle.Render("🎞️  Roll Selection - TODO")
}

func (s *RollSelectionScreen) HandleInput(key string, sm *state.StateMachine) bool {
	switch strings.ToLower(key) {
	case "esc":
		sm.GoBack()
		return true
	case "q":
		return false
	}
	return true
}

// MixedRollInputScreen stub
type MixedRollInputScreen struct{}

func (s *MixedRollInputScreen) Render(appState *state.ApplicationState) string {
	return TitleStyle.Render("🎞️  Mixed Roll Input - TODO")
}

func (s *MixedRollInputScreen) HandleInput(key string, sm *state.StateMachine) bool {
	switch strings.ToLower(key) {
	case "esc":
		sm.GoBack()
		return true
	case "q":
		return false
	}
	return true
}

// CalculatedScreen stub
type CalculatedScreen struct{}

func (s *CalculatedScreen) Render(appState *state.ApplicationState) string {
	return TitleStyle.Render("🎞️  Calculated Results - TODO")
}

func (s *CalculatedScreen) HandleInput(key string, sm *state.StateMachine) bool {
	switch strings.ToLower(key) {
	case "esc":
		sm.GoBack()
		return true
	case "q":
		return false
	}
	return true
}

// TimerScreen stub
type TimerScreen struct{}

func (s *TimerScreen) Render(appState *state.ApplicationState) string {
	return TitleStyle.Render("🎞️  Timer - TODO")
}

func (s *TimerScreen) HandleInput(key string, sm *state.StateMachine) bool {
	switch strings.ToLower(key) {
	case "esc":
		sm.GoBack()
		return true
	case "q":
		return false
	}
	return true
} 