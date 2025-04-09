package middleware

import orderedmap "github.com/wk8/go-ordered-map/v2"

// Context is a flexible, extensible state object that flows through the middleware pipeline,
// preserving insertion order of keys.
type Context = *orderedmap.OrderedMap[string, interface{}]

// CloneContext creates a deep copy of the context map itself, but a shallow copy
// of the values.
func CloneContext(ctx Context) Context {
	if ctx == nil {
		return orderedmap.New[string, interface{}]()
	}
	newCtx := orderedmap.New[string, interface{}]()
	for pair := ctx.Oldest(); pair != nil; pair = pair.Next() {
		newCtx.Set(pair.Key, pair.Value)
	}
	return newCtx
}

// PromptFragmentMetadata holds metadata associated with a PromptFragment.
type PromptFragmentMetadata struct {
	ID       string   `json:"id,omitempty"`
	Type     string   `json:"type,omitempty"`
	Position string   `json:"position,omitempty"` // "start", "middle", "end"
	Priority int      `json:"priority,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

// PromptFragment represents a structured piece of the prompt with associated metadata.
type PromptFragment struct {
	Content  string                 `json:"content"`
	Metadata PromptFragmentMetadata `json:"metadata"`
}

// Middleware defines the interface for components in the processing pipeline.
type Middleware interface {
	// Prompt transforms context and prompt fragments before sending to an LLM.
	// It returns the potentially modified context and the new list of fragments.
	Prompt(ctx Context, fragments []PromptFragment) (Context, []PromptFragment)

	// Parse processes the LLM response and updates the context.
	// It returns the potentially modified context and the processed response string.
	Parse(ctx Context, response string) (Context, string)

	// ID returns a unique identifier for the middleware.
	ID() string
	// Name returns a human-readable name for the middleware.
	Name() string
	// Description returns a short description of the middleware's purpose.
	Description() string
}
