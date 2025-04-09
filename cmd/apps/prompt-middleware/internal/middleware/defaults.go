package middleware

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// --- SystemInstructionMiddleware ---

type SystemInstructionMiddleware struct {
	Instructions string
}

// NewSystemInstructionMiddleware creates a middleware that adds a system instruction.
func NewSystemInstructionMiddleware(instructions string) *SystemInstructionMiddleware {
	if instructions == "" {
		instructions = "You are a helpful AI assistant. Answer clearly and concisely."
		log.Debug().Str("middlewareId", "system-instruction").Msg("No instructions provided, using default.")
	} else {
		log.Debug().Str("middlewareId", "system-instruction").Msg("Using provided system instructions.")
	}
	return &SystemInstructionMiddleware{Instructions: instructions}
}

var _ Middleware = &SystemInstructionMiddleware{}

func (m *SystemInstructionMiddleware) ID() string   { return "system-instruction" }
func (m *SystemInstructionMiddleware) Name() string { return "System Instruction" }
func (m *SystemInstructionMiddleware) Description() string {
	return "Adds system instructions at the beginning of the prompt."
}

func (m *SystemInstructionMiddleware) Prompt(ctx Context, fragments []PromptFragment) (Context, []PromptFragment) {
	systemFragment := PromptFragment{
		Content: m.Instructions,
		Metadata: PromptFragmentMetadata{
			ID:       m.ID(),
			Type:     "system",
			Position: "start",
			Priority: 100, // High priority to ensure it comes first
		},
	}
	// Prepend the system instruction
	log.Trace().Str("middlewareId", m.ID()).Interface("fragmentAdded", systemFragment).Msg("Prepending system instruction fragment")
	return ctx, append([]PromptFragment{systemFragment}, fragments...)
}

func (m *SystemInstructionMiddleware) Parse(ctx Context, response string) (Context, string) {
	// This middleware doesn't modify the response
	return ctx, response
}

// --- ThinkingModeMiddleware ---

type ThinkingModeMiddleware struct{}

// NewThinkingModeMiddleware creates a middleware that handles <thinking> tags.
func NewThinkingModeMiddleware() *ThinkingModeMiddleware {
	log.Debug().Str("middlewareId", "thinking-mode").Msg("Creating ThinkingModeMiddleware")
	return &ThinkingModeMiddleware{}
}

var _ Middleware = &ThinkingModeMiddleware{}

func (m *ThinkingModeMiddleware) ID() string   { return "thinking-mode" }
func (m *ThinkingModeMiddleware) Name() string { return "Thinking Mode" }
func (m *ThinkingModeMiddleware) Description() string {
	return "Adds thinking instruction if enabled in context, and extracts <thinking> content from response."
}

const ThinkingModeContextKey = "thinkingMode"
const ExtractedThinkingContextKey = "extractedThinking"

func (m *ThinkingModeMiddleware) Prompt(ctx Context, fragments []PromptFragment) (Context, []PromptFragment) {
	thinkingEnabled := false
	if enabled, ok := ctx.Get(ThinkingModeContextKey); ok && enabled.(bool) {
		thinkingEnabled = true
	}

	if thinkingEnabled {
		log.Debug().Str("middlewareId", m.ID()).Msg("Thinking mode enabled, adding instruction fragment.")
		thinkingFragment := PromptFragment{
			Content: "Please think step by step and show your reasoning in <thinking>...</thinking> tags.",
			Metadata: PromptFragmentMetadata{
				ID:       m.ID() + "-instruction",
				Type:     "instruction",
				Position: "middle",
				Priority: 60,
			},
		}
		// Append the thinking instruction
		log.Trace().Str("middlewareId", m.ID()).Interface("fragmentAdded", thinkingFragment).Msg("Appending thinking instruction fragment")
		return ctx, append(fragments, thinkingFragment)
	} else {
		log.Debug().Str("middlewareId", m.ID()).Msg("Thinking mode disabled, skipping instruction fragment.")
	}

	return ctx, fragments
}

var thinkingRegex = regexp.MustCompile(`(?s)<thinking>(.*?)</thinking>`) // (?s) enables dot to match newline

func (m *ThinkingModeMiddleware) Parse(ctx Context, response string) (Context, string) {
	log.Debug().Str("middlewareId", m.ID()).Msg("Parsing response for <thinking> tags")
	match := thinkingRegex.FindStringSubmatch(response)

	// Create a mutable copy of the context
	newCtx := CloneContext(ctx)

	cleanedResponse := response

	if len(match) > 1 {
		thinkingContent := strings.TrimSpace(match[1])
		newCtx.Set(ExtractedThinkingContextKey, thinkingContent)
		// Remove the thinking block from the response
		cleanedResponse = strings.TrimSpace(thinkingRegex.ReplaceAllString(response, ""))
		log.Debug().Str("middlewareId", m.ID()).Str("extractedThinking", thinkingContent).Msg("Extracted thinking content and updated context")
	} else {
		log.Debug().Str("middlewareId", m.ID()).Msg("No <thinking> tags found in response")
	}

	log.Trace().Str("middlewareId", m.ID()).Str("responseBefore", response).Str("responseAfter", cleanedResponse).Interface("contextAfter", newCtx).Msg("ThinkingModeMiddleware parse complete")
	return newCtx, cleanedResponse
}

// --- TokenCounterMiddleware ---

type TokenCounterMiddleware struct{}

// NewTokenCounterMiddleware creates a middleware that estimates token counts.
func NewTokenCounterMiddleware() *TokenCounterMiddleware {
	log.Debug().Str("middlewareId", "token-counter").Msg("Creating TokenCounterMiddleware")
	return &TokenCounterMiddleware{}
}

var _ Middleware = &TokenCounterMiddleware{}

func (m *TokenCounterMiddleware) ID() string   { return "token-counter" }
func (m *TokenCounterMiddleware) Name() string { return "Token Counter" }
func (m *TokenCounterMiddleware) Description() string {
	return "Estimates token usage (simple char/4 method)."
}

const PromptTokensContextKey = "promptTokens"
const ResponseTokensContextKey = "responseTokens"
const TotalTokensContextKey = "totalTokens"

func (m *TokenCounterMiddleware) Prompt(ctx Context, fragments []PromptFragment) (Context, []PromptFragment) {
	var combinedText strings.Builder
	for _, f := range fragments {
		combinedText.WriteString(f.Content)
		combinedText.WriteString(" ") // Add space between fragments
	}
	estimatedTokens := estimateTokens(combinedText.String())

	// Create a mutable copy of the context
	newCtx := CloneContext(ctx)

	newCtx.Set(PromptTokensContextKey, estimatedTokens)
	log.Debug().Str("middlewareId", m.ID()).Int("estimatedPromptTokens", estimatedTokens).Msg("Estimated prompt tokens and updated context")
	return newCtx, fragments
}

func (m *TokenCounterMiddleware) Parse(ctx Context, response string) (Context, string) {
	responseTokens := estimateTokens(response)
	promptTokens := 0
	if pt, ok := ctx.Get(PromptTokensContextKey); ok {
		if ptInt, ok := pt.(int); ok {
			promptTokens = ptInt
		} else {
			log.Error().Str("middlewareId", m.ID()).Interface("promptTokens", pt).Msg("Prompt tokens are not an int")
		}
	}
	totalTokens := promptTokens + responseTokens

	// Create a mutable copy of the context
	newCtx := CloneContext(ctx)

	newCtx.Set(ResponseTokensContextKey, responseTokens)
	newCtx.Set(TotalTokensContextKey, totalTokens)

	log.Debug().Str("middlewareId", m.ID()).Int("estimatedResponseTokens", responseTokens).Int("estimatedTotalTokens", totalTokens).Msg("Estimated response/total tokens and updated context")
	log.Trace().Str("middlewareId", m.ID()).Interface("contextAfter", newCtx).Msg("TokenCounterMiddleware parse complete")
	return newCtx, response
}

// estimateTokens provides a very basic token estimation.
// Replace with a proper tokenizer in a real application.
func estimateTokens(text string) int {
	// Basic estimation: 1 token ~= 4 characters
	if len(text) == 0 {
		return 0
	}
	return (len(text) + 3) / 4 // Ceiling division
}

// --- MockLLM --- (For demonstration purposes)

// MockLLM simulates an LLM call based on context.
func MockLLM(ctx Context, prompt string, userQuery string) string {
	log.Debug().Str("userQuery", userQuery).Interface("context", ctx).Msg("Simulating LLM call")
	baseResponse := fmt.Sprintf("I have processed your request regarding '%s'.", userQuery)
	thinkingBlock := ""

	if _, ok := ctx.Get(ExtractedThinkingContextKey); ok {
		// Thinking was extracted during parse phase, so don't add it again in mock response
	} else if enabled, ok := ctx.Get(ThinkingModeContextKey); ok && enabled.(bool) {
		// Add thinking block if ThinkingMode was enabled in context passed to LLM
		thinkingBlock = fmt.Sprintf(`

<thinking>
Okay, let me think step-by-step about the query "%s":
1. Identify the core request.
2. Recall relevant information.
3. Structure the response clearly.
</thinking>
`, userQuery)
	}

	// Add thinking block *before* the base response, simulating LLM output
	// If thinking block is empty, this just adds the base response.
	finalResponse := thinkingBlock + "\n\n" + baseResponse
	log.Debug().Str("simulatedResponse", finalResponse).Msg("LLM simulation complete")
	return finalResponse
}
