package models

// Screen represents a screen in the application
type Screen interface {
	Render(state *ApplicationState) string
	HandleInput(key string, sm *StateMachine) bool
}
