package main

import (
	"flag" // Added for command-line flags
	"net/http"
	"os"   // Added for zerolog console writer
	"sync" // Added for mutex
	"time" // Added for zerolog timestamp

	"github.com/go-go-golems/go-go-labs/cmd/apps/prompt-middleware/internal/middleware"
	"github.com/go-go-golems/go-go-labs/cmd/apps/prompt-middleware/internal/ui"
	"github.com/rs/zerolog"     // Added for logging
	"github.com/rs/zerolog/log" // Added for logging
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// MiddlewareState holds the actual middleware instance and its enabled status.
type MiddlewareState struct {
	Instance middleware.Middleware
	Enabled  bool
}

// AppState holds the application state, including the middleware pipeline configuration.
type AppState struct {
	// Store the original list of potential middlewares and their enabled status
	ConfiguredMiddlewares map[string]*MiddlewareState // Map middleware ID to its state
	Order                 []string                    // Keep track of middleware order for UI
	CurrentContext        middleware.Context
	InitialFragments      []middleware.PromptFragment
	UserQuery             string
	FinalPrompt           string
	LLMResponse           string
	ProcessedResponse     string
	FinalContext          middleware.Context
	mu                    sync.RWMutex // Protect concurrent access to state
}

func main() {
	// --- Logging Setup ---
	logLevelStr := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	flag.Parse()

	logLevel, err := zerolog.ParseLevel(*logLevelStr)
	if err != nil {
		logLevel = zerolog.InfoLevel
		// Use standard log here since zerolog isn't fully configured yet
		println("Invalid log level specified, defaulting to info")
	}

	// Pretty console logging
	output := zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}
	log.Logger = zerolog.New(output).Level(logLevel).With().Timestamp().Caller().Logger()

	log.Info().Str("logLevel", logLevel.String()).Msg("Logger initialized")

	// Initialize default middlewares and their initial state
	defaultMiddlewares := []middleware.Middleware{
		middleware.NewSystemInstructionMiddleware(""),
		middleware.NewThinkingModeMiddleware(),
		middleware.NewTokenCounterMiddleware(),
	}

	configuredMiddlewares := make(map[string]*MiddlewareState)
	order := make([]string, 0, len(defaultMiddlewares))
	log.Debug().Msg("Configuring default middlewares")
	for _, mw := range defaultMiddlewares {
		id := mw.ID()
		configuredMiddlewares[id] = &MiddlewareState{
			Instance: mw,
			Enabled:  true, // Start with all default middlewares enabled
		}
		order = append(order, id)
		log.Debug().
			Str("middlewareId", id).
			Str("middlewareName", mw.Name()).
			Bool("enabled", true).
			Msg("Configured middleware")
	}
	log.Info().Strs("middlewareOrder", order).Msg("Initial middleware order")

	// Initial state
	initialCtx := orderedmap.New[string, interface{}]()
	initialCtx.Set(middleware.ThinkingModeContextKey, true) // Start with thinking mode enabled

	appState := &AppState{
		ConfiguredMiddlewares: configuredMiddlewares,
		Order:                 order,
		CurrentContext:        initialCtx,
		UserQuery:             "Explain Go interfaces.",
		InitialFragments:      []middleware.PromptFragment{},
		FinalContext:          orderedmap.New[string, interface{}](), // Initialize FinalContext
	}

	log.Debug().Msg("Initial state created")

	// Run initial processing to populate state
	log.Info().Msg("Running initial pipeline processing...")
	appState.processPipeline()
	log.Info().Msg("Initial pipeline processing complete")

	// --- HTTP Handlers ---
	http.HandleFunc("/", appState.handleIndex)                            // Use the dynamic handler method
	http.HandleFunc("/process", appState.handleProcess)                   // Endpoint to trigger processing
	http.HandleFunc("/toggleMiddleware", appState.handleToggleMiddleware) // Endpoint to toggle middleware
	http.HandleFunc("/updateQuery", appState.handleUpdateQuery)           // Endpoint to update user query

	log.Info().Msg("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("ListenAndServe failed") // Use zerolog fatal
	}
}

// handleIndex renders the main layout dynamically.
func (s *AppState) handleIndex(w http.ResponseWriter, r *http.Request) {
	log.Debug().Str("method", r.Method).Str("path", r.URL.Path).Msg("Handling index request")
	// Render the full layout dynamically on each request
	pageData := s.createPageData() // Get current state using the receiver 's'
	component := ui.Layout(pageData)
	component.Render(r.Context(), w)
}

// buildActivePipelineLocked creates a MiddlewarePipeline containing only the currently enabled middlewares.
// It assumes the caller holds the necessary lock (read or write).
func (s *AppState) buildActivePipelineLocked() *middleware.MiddlewarePipeline {
	pipeline := middleware.NewMiddlewarePipeline()
	activeIDs := []string{}
	log.Debug().Msg("Building active middleware pipeline (lock already held)")
	for _, id := range s.Order {
		if state, exists := s.ConfiguredMiddlewares[id]; exists && state.Enabled {
			log.Debug().Str("middlewareId", id).Msg("Adding enabled middleware to active pipeline")
			pipeline.Use(state.Instance)
			activeIDs = append(activeIDs, id)
		} else {
			// Check if state exists before accessing Enabled field
			enabled := false
			if state != nil {
				enabled = state.Enabled
			}
			log.Debug().Str("middlewareId", id).Bool("exists", exists).Bool("enabled", enabled).Msg("Skipping disabled or non-existent middleware")
		}
	}
	log.Info().Strs("activeMiddlewares", activeIDs).Msg("Active pipeline built")
	return pipeline
}

// buildActivePipeline acquires the read lock and then builds the pipeline.
// Use this when the lock is not already held.
func (s *AppState) buildActivePipeline() *middleware.MiddlewarePipeline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buildActivePipelineLocked()
}

// createPageData generates the data needed by the main page template.
func (s *AppState) createPageData() ui.PageData {
	log.Debug().Msg("Creating page data")
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Extract middleware data for the UI, respecting the order
	mwData := make([]ui.MiddlewareData, 0, len(s.Order))
	for _, id := range s.Order {
		if state, exists := s.ConfiguredMiddlewares[id]; exists {
			mwData = append(mwData, ui.MiddlewareData{
				ID:          state.Instance.ID(),
				Name:        state.Instance.Name(),
				Description: state.Instance.Description(),
				Enabled:     state.Enabled,
			})
		}
	}

	// Make copies of context maps to avoid race conditions if UI renders concurrently
	initialCtxCopy := middleware.CloneContext(s.CurrentContext)
	finalCtxCopy := middleware.CloneContext(s.FinalContext)

	log.Debug().Msg("Page data created successfully")
	return ui.PageData{
		Middlewares:       mwData,
		InitialContext:    initialCtxCopy,
		UserQuery:         s.UserQuery,
		FinalPrompt:       s.FinalPrompt,
		LLMResponse:       s.LLMResponse,
		ProcessedResponse: s.ProcessedResponse,
		FinalContext:      finalCtxCopy,
	}
}

// processPipeline runs the full middleware pipeline using only enabled middlewares and updates the app state.
func (s *AppState) processPipeline() {
	log.Info().Msg("Starting pipeline processing")
	s.mu.Lock() // Use full lock as we are modifying state
	defer func() {
		log.Info().Msg("Finished pipeline processing")
		s.mu.Unlock()
	}()

	// Build pipeline with currently enabled middlewares
	activePipeline := s.buildActivePipelineLocked() // Logging happens inside this function

	// Create initial fragment from user query
	queryFragment := middleware.PromptFragment{
		Content: s.UserQuery,
		Metadata: middleware.PromptFragmentMetadata{
			ID:       "user-query",
			Type:     "query",
			Position: "middle",
			Priority: 50,
		},
	}
	// Combine with any other initial fragments if they exist
	initialFrags := append([]middleware.PromptFragment{}, s.InitialFragments...)
	initialFrags = append(initialFrags, queryFragment)
	log.Debug().Int("initialFragmentCount", len(initialFrags)).Interface("initialContext", s.CurrentContext).Msg("Prepared initial fragments and context for prompt phase")

	// --- Prompt Phase ---
	log.Info().Msg("Executing prompt phase...")
	promptCtx, _, finalPrompt := activePipeline.ExecutePromptPhase(s.CurrentContext, initialFrags)
	s.FinalPrompt = finalPrompt
	log.Info().Msg("Prompt phase completed")
	log.Debug().Str("finalPrompt", s.FinalPrompt).Interface("contextAfterPrompt", promptCtx).Msg("Prompt phase results")

	// --- LLM Call (Mock) ---
	log.Info().Msg("Executing mock LLM call...")
	llmResponse := middleware.MockLLM(promptCtx, s.FinalPrompt, s.UserQuery)
	s.LLMResponse = llmResponse // Store the raw response
	log.Info().Msg("Mock LLM call completed")
	log.Debug().Str("llmResponse", s.LLMResponse).Msg("LLM response received")

	// --- Parse Phase ---
	log.Info().Msg("Executing parse phase...")
	finalCtx, processedResp := activePipeline.ExecuteParsePhase(promptCtx, llmResponse)
	s.ProcessedResponse = processedResp
	s.FinalContext = finalCtx
	log.Info().Msg("Parse phase completed")
	log.Debug().Str("processedResponse", s.ProcessedResponse).Interface("finalContext", s.FinalContext).Msg("Parse phase results")
}

// handleProcess recalculates the pipeline and returns the updated UI fragment.
func (s *AppState) handleProcess(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("Received process request")
	if r.Method != http.MethodPost {
		log.Warn().Str("method", r.Method).Msg("Method not allowed for /process")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.Lock()
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse form")
		s.mu.Unlock()
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	s.UserQuery = r.FormValue("userQuery")
	thinkingMode := r.FormValue("thinkingMode") == "on"
	s.CurrentContext.Set(middleware.ThinkingModeContextKey, thinkingMode) // Use Set for orderedmap
	log.Info().Str("userQuery", s.UserQuery).Bool("thinkingMode", thinkingMode).Msg("Updated state from process request")
	s.mu.Unlock()

	// Re-run the pipeline (acquires lock internally)
	log.Info().Msg("Triggering pipeline processing from /process handler")
	s.processPipeline()

	// Render only the results part of the page using templ
	log.Debug().Msg("Rendering results panel component")
	component := ui.ResultsPanel(s.createPageData())
	err = component.Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Error rendering results panel component")
	} else {
		log.Debug().Msg("Successfully rendered results panel component")
	}
}

// handleToggleMiddleware toggles the enabled state of a middleware.
func (s *AppState) handleToggleMiddleware(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("Received toggle middleware request")
	if r.Method != http.MethodPost {
		log.Warn().Str("method", r.Method).Msg("Method not allowed for /toggleMiddleware")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse form")
		s.mu.Unlock()
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	middlewareID := r.FormValue("id")
	log.Debug().Str("middlewareId", middlewareID).Msg("Received toggle request")

	if state, exists := s.ConfiguredMiddlewares[middlewareID]; exists {
		state.Enabled = !state.Enabled
		log.Info().Str("middlewareId", middlewareID).Bool("enabled", state.Enabled).Msg("Toggled middleware state")
	} else {
		log.Warn().Str("middlewareId", middlewareID).Msg("Middleware ID not found for toggling")
	}
	s.mu.Unlock()

	// Re-run the pipeline with the updated enabled state
	log.Info().Msg("Triggering pipeline processing after toggle")
	s.processPipeline()

	// Get the updated page data
	pageData := s.createPageData()

	// Render both components with OOB swaps
	err = ui.MiddlewareList(pageData.Middlewares).Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Error rendering middleware list component")
		return
	}

	err = ui.ResultsPanelOOB(pageData).Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Error rendering results panel component")
		return
	}
}

// handleUpdateQuery updates the user query and re-processes.
func (s *AppState) handleUpdateQuery(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("method", r.Method).Str("path", r.URL.Path).Msg("Received update query request")
	if r.Method != http.MethodPost {
		log.Warn().Str("method", r.Method).Msg("Method not allowed for /updateQuery")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	err := r.ParseForm()
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse form")
		s.mu.Unlock()
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	s.UserQuery = r.FormValue("userQuery")
	log.Info().Str("newUserQuery", s.UserQuery).Msg("Updating user query")
	s.mu.Unlock()

	// Re-run the pipeline
	log.Info().Msg("Triggering pipeline processing after query update")
	s.processPipeline()

	// Re-render the results panel
	log.Debug().Msg("Rendering results panel component after query update")
	component := ui.ResultsPanel(s.createPageData())
	err = component.Render(r.Context(), w)
	if err != nil {
		log.Error().Err(err).Msg("Error rendering results panel component")
	} else {
		log.Debug().Msg("Successfully rendered results panel component")
	}
}
