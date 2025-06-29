package mcp

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
	"github.com/go-go-golems/go-go-mcp/pkg/protocol"
	"github.com/go-go-golems/go-go-mcp/pkg/server"
	"github.com/go-go-golems/go-go-mcp/pkg/transport"
	"github.com/go-go-golems/go-go-mcp/pkg/transport/sse"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// MCPWrapper wraps the go-go-mcp SDK and provides a clean interface
type MCPWrapper struct {
	mu            sync.RWMutex
	logger        zerolog.Logger
	server        *server.Server
	transport     transport.Transport
	searchHandler types.SearchHandler
	fetchHandler  types.FetchHandler
	authValidator types.AuthValidator

	// Server configuration
	serverName    string
	serverVersion string
	port          int

	// State
	isRunning bool
	stopChan  chan struct{}
}

// NewMCPWrapper creates a new MCP server wrapper
func NewMCPWrapper(logger zerolog.Logger, config types.Config) (*MCPWrapper, error) {
	start := time.Now()

	logger.Debug().
		Interface("config", config).
		Msg("Creating new MCP wrapper")

	if config.Port == 0 {
		config.Port = 8080
		logger.Debug().
			Int("default_port", config.Port).
			Msg("Using default port")
	}

	wrapper := &MCPWrapper{
		logger:        logger.With().Str("component", "mcp_wrapper").Logger(),
		serverName:    "Remote ChatGPT Connector",
		serverVersion: "0.1.0",
		port:          config.Port,
		stopChan:      make(chan struct{}),
	}

	wrapper.logger.Debug().
		Str("server_name", wrapper.serverName).
		Str("server_version", wrapper.serverVersion).
		Int("port", wrapper.port).
		Msg("MCP wrapper configuration initialized")

	// Initialize transport
	wrapper.logger.Debug().Msg("Initializing transport layer")
	if err := wrapper.initTransport(); err != nil {
		wrapper.logger.Error().
			Err(err).
			Dur("init_duration", time.Since(start)).
			Msg("Failed to initialize transport")
		return nil, errors.Wrap(err, "failed to initialize transport")
	}

	wrapper.logger.Info().
		Str("server_name", wrapper.serverName).
		Str("server_version", wrapper.serverVersion).
		Int("port", wrapper.port).
		Dur("init_duration", time.Since(start)).
		Msg("MCP wrapper initialized successfully")

	return wrapper, nil
}

// RegisterSearch registers a search handler
func (w *MCPWrapper) RegisterSearch(handler types.SearchHandler) error {
	w.logger.Debug().Msg("Attempting to register search handler")

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		w.logger.Error().
			Bool("is_running", w.isRunning).
			Msg("Cannot register search handler while server is running")
		return errors.New("cannot register handlers while server is running")
	}

	if handler == nil {
		w.logger.Error().Msg("Attempted to register nil search handler")
		return errors.New("search handler cannot be nil")
	}

	w.searchHandler = handler
	w.logger.Info().
		Str("handler_type", "search").
		Msg("Search handler registered successfully")
	return nil
}

// RegisterFetch registers a fetch handler
func (w *MCPWrapper) RegisterFetch(handler types.FetchHandler) error {
	w.logger.Debug().Msg("Attempting to register fetch handler")

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		w.logger.Error().
			Bool("is_running", w.isRunning).
			Msg("Cannot register fetch handler while server is running")
		return errors.New("cannot register handlers while server is running")
	}

	if handler == nil {
		w.logger.Error().Msg("Attempted to register nil fetch handler")
		return errors.New("fetch handler cannot be nil")
	}

	w.fetchHandler = handler
	w.logger.Info().
		Str("handler_type", "fetch").
		Msg("Fetch handler registered successfully")
	return nil
}

// SetAuthValidator sets the authentication validator
func (w *MCPWrapper) SetAuthValidator(validator types.AuthValidator) {
	w.logger.Debug().
		Bool("validator_provided", validator != nil).
		Msg("Setting auth validator")

	w.mu.Lock()
	defer w.mu.Unlock()

	w.authValidator = validator
	w.logger.Info().
		Bool("auth_enabled", validator != nil).
		Msg("Auth validator configured")
}

// GetHTTPHandler returns the HTTP handler for the MCP server
func (w *MCPWrapper) GetHTTPHandler() http.Handler {
	w.logger.Debug().Msg("Creating HTTP handler for MCP server")

	// Get handlers from the SSE transport
	sseTransport := w.transport.(*sse.SSETransport)
	handlers := sseTransport.GetHandlers()

	w.logger.Debug().
		Bool("has_sse_handler", handlers.SSEHandler != nil).
		Bool("has_message_handler", handlers.MessageHandler != nil).
		Msg("Retrieved SSE transport handlers")

	// Create a mux to handle different paths
	mux := http.NewServeMux()

	// Create authentication middleware wrapper
	authWrapper := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			start := time.Now()

			w.logger.Debug().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.Header.Get("User-Agent")).
				Msg("HTTP request received")

			// Apply authentication if validator is set
			if w.authValidator != nil {
				token := r.Header.Get("Authorization")
				if token == "" {
					w.logger.Warn().
						Str("path", r.URL.Path).
						Str("remote_addr", r.RemoteAddr).
						Dur("duration", time.Since(start)).
						Msg("Missing authorization header")
					http.Error(rw, "Unauthorized", http.StatusUnauthorized)
					return
				}

				w.logger.Debug().
					Str("path", r.URL.Path).
					Bool("has_bearer_prefix", len(token) > 7 && token[:7] == "Bearer ").
					Msg("Processing authentication token")

				// Remove "Bearer " prefix
				if len(token) > 7 && token[:7] == "Bearer " {
					token = token[7:]
				}

				// Validate token
				authStart := time.Now()
				userInfo, err := w.authValidator.ValidateToken(r.Context(), token)
				if err != nil {
					w.logger.Warn().
						Err(err).
						Str("path", r.URL.Path).
						Str("remote_addr", r.RemoteAddr).
						Dur("auth_duration", time.Since(authStart)).
						Dur("total_duration", time.Since(start)).
						Msg("Token validation failed")
					http.Error(rw, "Unauthorized", http.StatusUnauthorized)
					return
				}

				w.logger.Debug().
					Str("user_id", userInfo.ID).
					Str("user_login", userInfo.Login).
					Str("path", r.URL.Path).
					Dur("auth_duration", time.Since(authStart)).
					Msg("User authenticated successfully")
			} else {
				w.logger.Debug().
					Str("path", r.URL.Path).
					Msg("No authentication required")
			}

			// Call the original handler
			w.logger.Debug().
				Str("path", r.URL.Path).
				Msg("Forwarding to handler")
			handler(rw, r)

			w.logger.Debug().
				Str("path", r.URL.Path).
				Dur("total_duration", time.Since(start)).
				Msg("HTTP request completed")
		}
	}

	// Register SSE endpoints with authentication
	w.logger.Debug().Msg("Registering SSE endpoints")
	mux.HandleFunc("/sse", authWrapper(handlers.SSEHandler))
	mux.HandleFunc("/messages", authWrapper(handlers.MessageHandler))

	w.logger.Info().
		Bool("auth_enabled", w.authValidator != nil).
		Msg("HTTP handler configured with endpoints: /sse, /messages")

	return mux
}

// Start starts the MCP server
func (w *MCPWrapper) Start(ctx context.Context) error {
	start := time.Now()

	w.logger.Info().
		Int("port", w.port).
		Msg("Starting MCP server initialization")

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		w.logger.Error().
			Bool("is_running", w.isRunning).
			Msg("Server is already running")
		return errors.New("server is already running")
	}

	w.logger.Debug().Msg("Validating registered handlers")

	// Validate that handlers are registered
	if w.searchHandler == nil {
		w.logger.Error().Msg("Search handler not registered")
		return errors.New("search handler must be registered before starting")
	}
	if w.fetchHandler == nil {
		w.logger.Error().Msg("Fetch handler not registered")
		return errors.New("fetch handler must be registered before starting")
	}

	w.logger.Debug().
		Bool("search_handler_registered", w.searchHandler != nil).
		Bool("fetch_handler_registered", w.fetchHandler != nil).
		Bool("auth_validator_set", w.authValidator != nil).
		Msg("Handler validation completed")

	// Initialize server now that we have handlers
	w.logger.Debug().Msg("Initializing MCP server")
	if err := w.initServer(); err != nil {
		w.logger.Error().
			Err(err).
			Dur("startup_duration", time.Since(start)).
			Msg("Failed to initialize server")
		return errors.Wrap(err, "failed to initialize server")
	}

	w.isRunning = true

	w.logger.Info().
		Int("port", w.port).
		Dur("init_duration", time.Since(start)).
		Msg("MCP server initialized, starting transport")

	// Start the server in a goroutine
	go func() {
		serverStart := time.Now()
		w.logger.Debug().Msg("Starting server transport layer")

		if err := w.server.Start(ctx); err != nil {
			w.logger.Error().
				Err(err).
				Dur("server_runtime", time.Since(serverStart)).
				Msg("Server transport failed")
		} else {
			w.logger.Info().
				Dur("server_runtime", time.Since(serverStart)).
				Msg("Server transport stopped normally")
		}
	}()

	w.logger.Info().
		Int("port", w.port).
		Dur("startup_duration", time.Since(start)).
		Msg("MCP server started successfully")

	return nil
}

// Stop stops the MCP server
func (w *MCPWrapper) Stop() error {
	start := time.Now()

	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isRunning {
		w.logger.Debug().
			Bool("is_running", w.isRunning).
			Msg("Server is already stopped")
		return nil
	}

	w.logger.Info().Msg("Initiating MCP server shutdown")

	stopTimeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), stopTimeout)
	defer cancel()

	w.logger.Debug().
		Dur("timeout", stopTimeout).
		Msg("Stopping server with timeout")

	if err := w.server.Stop(ctx); err != nil {
		w.logger.Error().
			Err(err).
			Dur("shutdown_duration", time.Since(start)).
			Dur("timeout", stopTimeout).
			Msg("Error stopping server")
		return errors.Wrap(err, "failed to stop server")
	}

	w.logger.Debug().Msg("Closing stop channel")
	close(w.stopChan)
	w.isRunning = false

	w.logger.Info().
		Dur("shutdown_duration", time.Since(start)).
		Msg("MCP server stopped successfully")
	return nil
}

// IsRunning returns whether the server is currently running
func (w *MCPWrapper) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	w.logger.Debug().
		Bool("is_running", w.isRunning).
		Msg("Server running status checked")

	return w.isRunning
}

// initTransport initializes the SSE transport
func (w *MCPWrapper) initTransport() error {
	start := time.Now()

	w.logger.Debug().
		Int("port", w.port).
		Str("addr", fmt.Sprintf(":%d", w.port)).
		Msg("Initializing SSE transport")

	var err error
	w.transport, err = sse.NewSSETransport(
		transport.WithSSEOptions(transport.SSEOptions{
			Addr: fmt.Sprintf(":%d", w.port),
		}),
		transport.WithLogger(w.logger),
	)
	if err != nil {
		w.logger.Error().
			Err(err).
			Int("port", w.port).
			Dur("init_duration", time.Since(start)).
			Msg("Failed to create SSE transport")
		return errors.Wrap(err, "failed to create SSE transport")
	}

	w.logger.Debug().
		Int("port", w.port).
		Dur("init_duration", time.Since(start)).
		Msg("SSE transport initialized successfully")
	return nil
}

// initServer initializes the MCP server with handlers
func (w *MCPWrapper) initServer() error {
	start := time.Now()

	w.logger.Debug().
		Str("server_name", w.serverName).
		Str("server_version", w.serverVersion).
		Msg("Initializing MCP server")

	// Create resource provider that bridges to our handlers
	w.logger.Debug().Msg("Creating resource provider")
	resourceProvider := &ResourceProvider{
		logger:        w.logger.With().Str("component", "resource_provider").Logger(),
		searchHandler: w.searchHandler,
		fetchHandler:  w.fetchHandler,
	}

	w.logger.Debug().
		Bool("has_search_handler", w.searchHandler != nil).
		Bool("has_fetch_handler", w.fetchHandler != nil).
		Msg("Resource provider configured")

	w.server = server.NewServer(
		w.logger,
		w.transport,
		server.WithServerName(w.serverName),
		server.WithServerVersion(w.serverVersion),
		server.WithResourceProvider(resourceProvider),
	)

	w.logger.Debug().
		Str("server_name", w.serverName).
		Str("server_version", w.serverVersion).
		Dur("init_duration", time.Since(start)).
		Msg("MCP server initialized successfully")
	return nil
}

// ResourceProvider implements the pkg.ResourceProvider interface
type ResourceProvider struct {
	logger        zerolog.Logger
	searchHandler types.SearchHandler
	fetchHandler  types.FetchHandler
}

// ListResources returns a list of available resources with optional pagination
func (rp *ResourceProvider) ListResources(ctx context.Context, cursor string) ([]protocol.Resource, string, error) {
	start := time.Now()

	rp.logger.Debug().
		Str("cursor", cursor).
		Msg("ListResources called")

	// For the search/fetch paradigm, we don't expose a static list of resources
	// Instead, we return an empty list and rely on search functionality
	resources := []protocol.Resource{}
	nextCursor := ""

	rp.logger.Debug().
		Str("cursor", cursor).
		Int("resource_count", len(resources)).
		Str("next_cursor", nextCursor).
		Dur("duration", time.Since(start)).
		Msg("ListResources completed - returning empty list for search/fetch paradigm")

	return resources, nextCursor, nil
}

// ReadResource retrieves the contents of a specific resource
func (rp *ResourceProvider) ReadResource(ctx context.Context, uri string) ([]protocol.ResourceContent, error) {
	start := time.Now()

	rp.logger.Debug().
		Str("uri", uri).
		Msg("ReadResource called")

	if rp.fetchHandler == nil {
		rp.logger.Error().
			Str("uri", uri).
			Dur("duration", time.Since(start)).
			Msg("Fetch handler not registered")
		return nil, errors.New("fetch handler not registered")
	}

	// Use our fetch handler to get the content
	fetchReq := types.FetchRequest{ID: uri}

	rp.logger.Debug().
		Str("uri", uri).
		Str("fetch_request_id", fetchReq.ID).
		Msg("Calling fetch handler")

	fetchStart := time.Now()
	result, err := rp.fetchHandler(ctx, fetchReq)
	if err != nil {
		rp.logger.Error().
			Err(err).
			Str("uri", uri).
			Dur("fetch_duration", time.Since(fetchStart)).
			Dur("total_duration", time.Since(start)).
			Msg("Fetch handler failed")
		return nil, errors.Wrap(err, "fetch handler failed")
	}

	rp.logger.Debug().
		Str("uri", uri).
		Str("result_id", result.ID).
		Int("data_length", len(result.Data)).
		Dur("fetch_duration", time.Since(fetchStart)).
		Msg("Fetch handler completed successfully")

	// Convert to protocol format
	content := protocol.ResourceContent{
		URI:      result.ID,
		MimeType: "text/plain",
		Text:     result.Data,
	}

	rp.logger.Debug().
		Str("uri", uri).
		Str("content_uri", content.URI).
		Str("mime_type", content.MimeType).
		Int("content_length", len(content.Text)).
		Dur("total_duration", time.Since(start)).
		Msg("ReadResource completed successfully")

	return []protocol.ResourceContent{content}, nil
}

// ListResourceTemplates returns a list of available resource templates
func (rp *ResourceProvider) ListResourceTemplates(ctx context.Context) ([]protocol.ResourceTemplate, error) {
	start := time.Now()

	rp.logger.Debug().Msg("ListResourceTemplates called")

	// For the search/fetch paradigm, we offer a search template
	template := protocol.ResourceTemplate{
		URITemplate: "search://{query}",
		Name:        "Search",
		Description: "Search for resources by query",
		MimeType:    "text/plain",
	}

	templates := []protocol.ResourceTemplate{template}

	rp.logger.Debug().
		Int("template_count", len(templates)).
		Str("template_uri", template.URITemplate).
		Str("template_name", template.Name).
		Dur("duration", time.Since(start)).
		Msg("ListResourceTemplates completed")

	return templates, nil
}

// SubscribeToResource registers for notifications about resource changes
func (rp *ResourceProvider) SubscribeToResource(ctx context.Context, uri string) (chan struct{}, func(), error) {
	start := time.Now()

	rp.logger.Debug().
		Str("uri", uri).
		Msg("SubscribeToResource called")

	// For simplicity, we don't support resource subscriptions
	// Return a closed channel and a no-op cleanup function
	ch := make(chan struct{})
	close(ch)

	cleanup := func() {
		rp.logger.Debug().
			Str("uri", uri).
			Msg("Cleaning up resource subscription")
	}

	rp.logger.Debug().
		Str("uri", uri).
		Dur("duration", time.Since(start)).
		Msg("SubscribeToResource completed - returning closed channel (no subscription support)")

	return ch, cleanup, nil
}

// DemoSearchHandler provides a demo search implementation
func DemoSearchHandler(ctx context.Context, req types.SearchRequest) (<-chan types.SearchResult, error) {
	// Create a logger for demo search handler
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().
		Str("component", "demo_search_handler").
		Str("query", req.Query).
		Logger()

	start := time.Now()

	logger.Debug().
		Str("query", req.Query).
		Msg("DemoSearchHandler called")

	results := make(chan types.SearchResult, 3)

	go func() {
		defer func() {
			close(results)
			logger.Debug().
				Str("query", req.Query).
				Dur("streaming_duration", time.Since(start)).
				Msg("Demo search results streaming completed")
		}()

		demoResults := []types.SearchResult{
			{
				ID:    "demo-1",
				Title: "Hello from Go MCP Server!",
				URL:   "https://example.com/hello",
				Chunk: fmt.Sprintf("This is a demo search result for query: '%s'. The MCP server is working correctly.", req.Query),
			},
			{
				ID:    "demo-2",
				Title: "Another Demo Result",
				URL:   "https://example.com/another",
				Chunk: "This demonstrates streaming search results through the MCP protocol.",
			},
			{
				ID:    "demo-3",
				Title: "Third Demo Result",
				URL:   "https://example.com/third",
				Chunk: "Search functionality is working through the remote MCP connector.",
			},
		}

		logger.Debug().
			Str("query", req.Query).
			Int("result_count", len(demoResults)).
			Msg("Streaming demo search results")

		// Send demo results with logging
		for i, result := range demoResults {
			logger.Debug().
				Str("query", req.Query).
				Str("result_id", result.ID).
				Str("result_title", result.Title).
				Str("result_url", result.URL).
				Int("result_index", i+1).
				Int("total_results", len(demoResults)).
				Int("chunk_length", len(result.Chunk)).
				Msg("Sending demo search result")

			results <- result

			// Small delay to simulate processing time
			time.Sleep(10 * time.Millisecond)
		}
	}()

	logger.Debug().
		Str("query", req.Query).
		Dur("setup_duration", time.Since(start)).
		Msg("DemoSearchHandler setup completed, returning results channel")

	return results, nil
}

// DemoFetchHandler provides a demo fetch implementation
func DemoFetchHandler(ctx context.Context, req types.FetchRequest) (types.FetchResult, error) {
	// Create a logger for demo fetch handler
	logger := zerolog.New(zerolog.NewConsoleWriter()).With().
		Str("component", "demo_fetch_handler").
		Str("resource_id", req.ID).
		Logger()

	start := time.Now()

	logger.Debug().
		Str("resource_id", req.ID).
		Msg("DemoFetchHandler called")

	// Simulate some processing time
	time.Sleep(5 * time.Millisecond)

	result := types.FetchResult{
		ID:   req.ID,
		Data: fmt.Sprintf("Full content for resource '%s'.\n\nThis is placeholder content returned by the demo fetch handler. In a real implementation, this would contain the actual content retrieved from the data source identified by the ID.", req.ID),
	}

	logger.Debug().
		Str("resource_id", req.ID).
		Str("result_id", result.ID).
		Int("data_length", len(result.Data)).
		Dur("fetch_duration", time.Since(start)).
		Msg("DemoFetchHandler completed successfully")

	return result, nil
}
