package middleware

import (
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
)

// MiddlewarePipeline manages the execution of a sequence of middlewares.
type MiddlewarePipeline struct {
	middlewares []Middleware
}

// NewMiddlewarePipeline creates a new, empty middleware pipeline.
func NewMiddlewarePipeline() *MiddlewarePipeline {
	log.Debug().Msg("Creating new middleware pipeline")
	return &MiddlewarePipeline{
		middlewares: []Middleware{},
	}
}

// Use adds a middleware to the pipeline.
func (p *MiddlewarePipeline) Use(middleware Middleware) *MiddlewarePipeline {
	log.Debug().Str("middlewareId", middleware.ID()).Str("middlewareName", middleware.Name()).Msg("Adding middleware to pipeline")
	p.middlewares = append(p.middlewares, middleware)
	return p
}

// Middlewares returns the list of middlewares currently in the pipeline.
func (p *MiddlewarePipeline) Middlewares() []Middleware {
	// Return a copy to prevent external modification
	result := make([]Middleware, len(p.middlewares))
	copy(result, p.middlewares)
	return result
}

// ExecutePromptPhase runs the prompt transformation phase of the pipeline.
func (p *MiddlewarePipeline) ExecutePromptPhase(
	initialContext Context,
	initialFragments []PromptFragment,
) (Context, []PromptFragment, string) {
	log.Debug().Msg("Executing prompt phase")
	currentContext := initialContext
	// Deep copy context map to avoid modification issues
	if currentContext == nil {
		currentContext = make(Context)
	} else {
		newCtx := make(Context)
		for k, v := range currentContext {
			newCtx[k] = v
		}
		currentContext = newCtx
	}

	currentFragments := make([]PromptFragment, len(initialFragments))
	copy(currentFragments, initialFragments)
	log.Debug().Int("initialFragmentCount", len(currentFragments)).Interface("initialContext", currentContext).Msg("Prompt phase start")

	// Pass through each middleware's prompt function
	for i, middleware := range p.middlewares {
		log.Debug().Int("step", i).Str("middlewareId", middleware.ID()).Msg("Executing middleware prompt function")
		ctxBefore := currentContext
		fragsBefore := currentFragments
		currentContext, currentFragments = middleware.Prompt(currentContext, currentFragments)
		log.Trace(). // Use Trace for potentially very verbose output
				Int("step", i).
				Str("middlewareId", middleware.ID()).
				Interface("contextBefore", ctxBefore).
				Interface("fragmentsBefore", fragsBefore).
				Interface("contextAfter", currentContext).
				Interface("fragmentsAfter", currentFragments).
				Msg("Middleware prompt function state change")
	}

	// Sort fragments by position and priority
	log.Debug().Msg("Sorting fragments")
	sortedFragments := SortFragments(currentFragments) // Logging inside SortFragments

	// Combine into final prompt
	log.Debug().Msg("Combining fragments into final prompt")
	finalPrompt := CombineFragments(sortedFragments) // Logging inside CombineFragments
	log.Debug().Int("sortedFragmentCount", len(sortedFragments)).Str("finalPrompt", finalPrompt).Interface("finalContext", currentContext).Msg("Prompt phase completed")

	return currentContext, sortedFragments, finalPrompt
}

// ExecuteParsePhase runs the response parsing phase of the pipeline.
func (p *MiddlewarePipeline) ExecuteParsePhase(
	context Context,
	llmResponse string,
) (Context, string) {
	log.Debug().Msg("Executing parse phase")
	currentContext := context
	// Deep copy context map
	if currentContext == nil {
		currentContext = make(Context)
	} else {
		newCtx := make(Context)
		for k, v := range currentContext {
			newCtx[k] = v
		}
		currentContext = newCtx
	}
	currentResponse := llmResponse
	log.Debug().Str("initialResponse", currentResponse).Interface("initialContext", currentContext).Msg("Parse phase start")

	// Process in reverse order for symmetry
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		middleware := p.middlewares[i]
		log.Debug().Int("step", i).Str("middlewareId", middleware.ID()).Msg("Executing middleware parse function (reverse order)")
		ctxBefore := currentContext
		respBefore := currentResponse
		currentContext, currentResponse = middleware.Parse(currentContext, currentResponse)
		log.Trace(). // Use Trace for potentially very verbose output
				Int("step", i).
				Str("middlewareId", middleware.ID()).
				Interface("contextBefore", ctxBefore).
				Str("responseBefore", respBefore).
				Interface("contextAfter", currentContext).
				Str("responseAfter", currentResponse).
				Msg("Middleware parse function state change")
	}
	log.Debug().Str("finalResponse", currentResponse).Interface("finalContext", currentContext).Msg("Parse phase completed")

	return currentContext, currentResponse
}

// SortFragments sorts prompt fragments based on position ('start', 'middle', 'end')
// and then by priority (descending).
func SortFragments(fragments []PromptFragment) []PromptFragment {
	log.Debug().Int("fragmentCount", len(fragments)).Msg("Sorting fragments by position and priority")
	positionOrder := map[string]int{
		"start":  0,
		"middle": 1,
		"end":    2,
		"":       1, // Default to middle if empty
	}

	// Create a copy to avoid modifying the original slice potentially passed in
	sortedFragments := make([]PromptFragment, len(fragments))
	copy(sortedFragments, fragments)

	sort.SliceStable(sortedFragments, func(i, j int) bool {
		metaI := sortedFragments[i].Metadata
		metaJ := sortedFragments[j].Metadata

		posI := positionOrder[metaI.Position]
		if metaI.Position == "" {
			posI = positionOrder["middle"]
		}
		posJ := positionOrder[metaJ.Position]
		if metaJ.Position == "" {
			posJ = positionOrder["middle"]
		}

		// First sort by position
		if posI != posJ {
			return posI < posJ
		}

		// Then by priority (higher numbers come first)
		return metaI.Priority > metaJ.Priority
	})

	log.Debug().Int("sortedFragmentCount", len(sortedFragments)).Msg("Fragments sorted")
	return sortedFragments
}

// CombineFragments joins the content of sorted PromptFragments into a single string.
func CombineFragments(fragments []PromptFragment) string {
	log.Debug().Int("fragmentCount", len(fragments)).Msg("Combining fragments into a single string")
	var builder strings.Builder
	first := true
	for _, f := range fragments {
		content := strings.TrimSpace(f.Content)
		if content == "" {
			continue // Skip empty fragments
		}
		if !first {
			builder.WriteString("\n\n") // Join with double newlines
		}
		builder.WriteString(content)
		first = false
	}
	finalString := builder.String()
	log.Debug().Int("combinedStringLength", len(finalString)).Msg("Fragments combined")
	return finalString
}
