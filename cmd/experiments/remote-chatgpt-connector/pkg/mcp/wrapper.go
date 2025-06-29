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
	if config.Port == 0 {
		config.Port = 8080
	}

	wrapper := &MCPWrapper{
		logger:        logger.With().Str("component", "mcp_wrapper").Logger(),
		serverName:    "Remote ChatGPT Connector",
		serverVersion: "0.1.0",
		port:          config.Port,
		stopChan:      make(chan struct{}),
	}

	// Initialize transport
	if err := wrapper.initTransport(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize transport")
	}

	wrapper.logger.Info().
		Str("server_name", wrapper.serverName).
		Str("server_version", wrapper.serverVersion).
		Int("port", wrapper.port).
		Msg("MCP wrapper initialized")

	return wrapper, nil
}

// RegisterSearch registers a search handler
func (w *MCPWrapper) RegisterSearch(handler types.SearchHandler) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		return errors.New("cannot register handlers while server is running")
	}

	w.searchHandler = handler
	w.logger.Info().Msg("Search handler registered")
	return nil
}

// RegisterFetch registers a fetch handler
func (w *MCPWrapper) RegisterFetch(handler types.FetchHandler) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		return errors.New("cannot register handlers while server is running")
	}

	w.fetchHandler = handler
	w.logger.Info().Msg("Fetch handler registered")
	return nil
}

// SetAuthValidator sets the authentication validator
func (w *MCPWrapper) SetAuthValidator(validator types.AuthValidator) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.authValidator = validator
	w.logger.Info().Msg("Auth validator set")
}

// GetHTTPHandler returns the HTTP handler for the MCP server
func (w *MCPWrapper) GetHTTPHandler() http.Handler {
	// Get handlers from the SSE transport
	sseTransport := w.transport.(*sse.SSETransport)
	handlers := sseTransport.GetHandlers()

	// Create a mux to handle different paths
	mux := http.NewServeMux()

	// Create authentication middleware wrapper
	authWrapper := func(handler http.HandlerFunc) http.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request) {
			// Apply authentication if validator is set
			if w.authValidator != nil {
				token := r.Header.Get("Authorization")
				if token == "" {
					w.logger.Warn().Msg("Missing authorization header")
					http.Error(rw, "Unauthorized", http.StatusUnauthorized)
					return
				}

				// Remove "Bearer " prefix
				if len(token) > 7 && token[:7] == "Bearer " {
					token = token[7:]
				}

				// Validate token
				userInfo, err := w.authValidator.ValidateToken(r.Context(), token)
				if err != nil {
					w.logger.Warn().Err(err).Msg("Token validation failed")
					http.Error(rw, "Unauthorized", http.StatusUnauthorized)
					return
				}

				w.logger.Debug().
					Str("user_id", userInfo.ID).
					Str("user_login", userInfo.Login).
					Msg("User authenticated")
			}

			// Call the original handler
			handler(rw, r)
		}
	}

	// Register SSE endpoints with authentication
	mux.HandleFunc("/sse", authWrapper(handlers.SSEHandler))
	mux.HandleFunc("/messages", authWrapper(handlers.MessageHandler))

	return mux
}

// Start starts the MCP server
func (w *MCPWrapper) Start(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.isRunning {
		return errors.New("server is already running")
	}

	// Validate that handlers are registered
	if w.searchHandler == nil {
		return errors.New("search handler must be registered before starting")
	}
	if w.fetchHandler == nil {
		return errors.New("fetch handler must be registered before starting")
	}

	// Initialize server now that we have handlers
	if err := w.initServer(); err != nil {
		return errors.Wrap(err, "failed to initialize server")
	}

	w.isRunning = true

	w.logger.Info().
		Int("port", w.port).
		Msg("Starting MCP server")

	// Start the server in a goroutine
	go func() {
		if err := w.server.Start(ctx); err != nil {
			w.logger.Error().Err(err).Msg("Server failed")
		}
	}()

	return nil
}

// Stop stops the MCP server
func (w *MCPWrapper) Stop() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.isRunning {
		return nil
	}

	w.logger.Info().Msg("Stopping MCP server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := w.server.Stop(ctx); err != nil {
		w.logger.Error().Err(err).Msg("Error stopping server")
		return errors.Wrap(err, "failed to stop server")
	}

	close(w.stopChan)
	w.isRunning = false

	w.logger.Info().Msg("MCP server stopped")
	return nil
}

// IsRunning returns whether the server is currently running
func (w *MCPWrapper) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.isRunning
}

// initTransport initializes the SSE transport
func (w *MCPWrapper) initTransport() error {
	var err error
	w.transport, err = sse.NewSSETransport(
		transport.WithSSEOptions(transport.SSEOptions{
			Addr: fmt.Sprintf(":%d", w.port),
		}),
		transport.WithLogger(w.logger),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create SSE transport")
	}

	w.logger.Debug().Msg("SSE transport initialized")
	return nil
}

// initServer initializes the MCP server with handlers
func (w *MCPWrapper) initServer() error {
	// Create resource provider that bridges to our handlers
	resourceProvider := &ResourceProvider{
		logger:        w.logger.With().Str("component", "resource_provider").Logger(),
		searchHandler: w.searchHandler,
		fetchHandler:  w.fetchHandler,
	}

	w.server = server.NewServer(
		w.logger,
		w.transport,
		server.WithServerName(w.serverName),
		server.WithServerVersion(w.serverVersion),
		server.WithResourceProvider(resourceProvider),
	)

	w.logger.Debug().Msg("MCP server initialized")
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
	rp.logger.Debug().Str("cursor", cursor).Msg("Listing resources")

	// For the search/fetch paradigm, we don't expose a static list of resources
	// Instead, we return an empty list and rely on search functionality
	return []protocol.Resource{}, "", nil
}

// ReadResource retrieves the contents of a specific resource
func (rp *ResourceProvider) ReadResource(ctx context.Context, uri string) ([]protocol.ResourceContent, error) {
	rp.logger.Debug().Str("uri", uri).Msg("Reading resource")

	if rp.fetchHandler == nil {
		return nil, errors.New("fetch handler not registered")
	}

	// Use our fetch handler to get the content
	fetchReq := types.FetchRequest{ID: uri}
	result, err := rp.fetchHandler(ctx, fetchReq)
	if err != nil {
		return nil, errors.Wrap(err, "fetch handler failed")
	}

	// Convert to protocol format
	content := protocol.ResourceContent{
		URI:      result.ID,
		MimeType: "text/plain",
		Text:     result.Data,
	}

	return []protocol.ResourceContent{content}, nil
}

// ListResourceTemplates returns a list of available resource templates
func (rp *ResourceProvider) ListResourceTemplates(ctx context.Context) ([]protocol.ResourceTemplate, error) {
	rp.logger.Debug().Msg("Listing resource templates")

	// For the search/fetch paradigm, we offer a search template
	template := protocol.ResourceTemplate{
		URITemplate: "search://{query}",
		Name:        "Search",
		Description: "Search for resources by query",
		MimeType:    "text/plain",
	}

	return []protocol.ResourceTemplate{template}, nil
}

// SubscribeToResource registers for notifications about resource changes
func (rp *ResourceProvider) SubscribeToResource(ctx context.Context, uri string) (chan struct{}, func(), error) {
	rp.logger.Debug().Str("uri", uri).Msg("Subscribing to resource")

	// For simplicity, we don't support resource subscriptions
	// Return a closed channel and a no-op cleanup function
	ch := make(chan struct{})
	close(ch)

	cleanup := func() {
		rp.logger.Debug().Str("uri", uri).Msg("Cleaning up resource subscription")
	}

	return ch, cleanup, nil
}

// DemoSearchHandler provides a demo search implementation
func DemoSearchHandler(ctx context.Context, req types.SearchRequest) (<-chan types.SearchResult, error) {
	results := make(chan types.SearchResult, 3)

	go func() {
		defer close(results)

		// Send demo results
		results <- types.SearchResult{
			ID:    "demo-1",
			Title: "Hello from Go MCP Server!",
			URL:   "https://example.com/hello",
			Chunk: fmt.Sprintf("This is a demo search result for query: '%s'. The MCP server is working correctly.", req.Query),
		}

		results <- types.SearchResult{
			ID:    "demo-2",
			Title: "Another Demo Result",
			URL:   "https://example.com/another",
			Chunk: "This demonstrates streaming search results through the MCP protocol.",
		}

		results <- types.SearchResult{
			ID:    "demo-3",
			Title: "Third Demo Result",
			URL:   "https://example.com/third",
			Chunk: "Search functionality is working through the remote MCP connector.",
		}
	}()

	return results, nil
}

// DemoFetchHandler provides a demo fetch implementation
func DemoFetchHandler(ctx context.Context, req types.FetchRequest) (types.FetchResult, error) {
	return types.FetchResult{
		ID:   req.ID,
		Data: fmt.Sprintf("Full content for resource '%s'.\n\nThis is placeholder content returned by the demo fetch handler. In a real implementation, this would contain the actual content retrieved from the data source identified by the ID.", req.ID),
	}, nil
}
