package ui

import (
	"github.com/go-go-golems/go-go-labs/cmd/apps/prompt-middleware/internal/middleware"
)

// MiddlewareData holds the information needed to display a middleware in the UI.
type MiddlewareData struct {
	ID          string
	Name        string
	Description string
	Enabled     bool
}

// PageData aggregates all data needed to render the main UI page.
type PageData struct {
	Middlewares       []MiddlewareData // List of middlewares for configuration panel
	InitialContext    middleware.Context
	UserQuery         string
	FinalPrompt       string
	LLMResponse       string // Raw LLM response (before parsing)
	ProcessedResponse string // Response after middleware parsing
	FinalContext      middleware.Context
	// TODO: Add FragmentStages []FragmentStageData for visualization if needed
}

/*
// Example for fragment visualization if implemented later
type FragmentStageData struct {
	StageName string
	Fragments []middleware.PromptFragment
	Context   middleware.Context
}
*/
