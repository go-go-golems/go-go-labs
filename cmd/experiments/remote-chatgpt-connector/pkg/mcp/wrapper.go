package mcp

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/middleware"
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
	mu             sync.RWMutex
	logger         zerolog.Logger
	server         *server.Server
	transport      transport.Transport
	searchHandler  types.SearchHandler
	fetchHandler   types.FetchHandler
	authMiddleware *middleware.AuthMiddleware

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

// SetAuthMiddleware sets the authentication middleware
func (w *MCPWrapper) SetAuthMiddleware(authMiddleware *middleware.AuthMiddleware) {
	w.logger.Debug().
		Bool("middleware_provided", authMiddleware != nil).
		Msg("Setting auth middleware")

	w.mu.Lock()
	defer w.mu.Unlock()

	w.authMiddleware = authMiddleware
	w.logger.Info().
		Bool("auth_enabled", authMiddleware != nil).
		Msg("Auth middleware configured")
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

	// Register SSE endpoints with authentication middleware
	w.logger.Debug().Msg("Registering SSE endpoints")

	if w.authMiddleware != nil {
		// Protected endpoints
		mux.HandleFunc("/sse", w.authMiddleware.RequireAuth(handlers.SSEHandler))
		mux.HandleFunc("/messages", w.authMiddleware.RequireAuth(handlers.MessageHandler))
		w.logger.Info().Msg("SSE endpoints configured with authentication middleware")
	} else {
		// Unprotected endpoints (for development/testing)
		mux.HandleFunc("/sse", handlers.SSEHandler)
		mux.HandleFunc("/messages", handlers.MessageHandler)
		w.logger.Warn().Msg("SSE endpoints configured WITHOUT authentication (development mode)")
	}

	w.logger.Info().
		Bool("auth_enabled", w.authMiddleware != nil).
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
		Bool("auth_middleware_set", w.authMiddleware != nil).
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
		Str("result_title", result.Title).
		Int("text_length", len(result.Text)).
		Dur("fetch_duration", time.Since(fetchStart)).
		Msg("Fetch handler completed successfully")

	// Convert to protocol format
	content := protocol.ResourceContent{
		URI:      result.ID,
		MimeType: "text/plain",
		Text:     result.Text,
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
				ID:    "doc-001",
				Title: "Introduction to Model Context Protocol",
				Text:  fmt.Sprintf("The Model Context Protocol (MCP) is an open standard for connecting AI assistants to data sources. Your search for '%s' matches this comprehensive guide covering MCP architecture, server implementation, and best practices for building context-aware AI applications.", req.Query),
				URL:   "https://modelcontextprotocol.io/introduction",
			},
			{
				ID:    "doc-002",
				Title: "OAuth 2.1 and Dynamic Client Registration",
				Text:  "OAuth 2.0 Dynamic Client Registration Protocol allows clients to register with authorization servers automatically. This specification defines how clients can obtain registration information and credentials without manual intervention, enabling scalable OAuth deployments.",
				URL:   "https://datatracker.ietf.org/doc/html/rfc7591",
			},
			{
				ID:    "doc-003",
				Title: "Building Remote MCP Servers with Go",
				Text:  "Learn how to build production-ready MCP servers using Go. This tutorial covers HTTP transports, SSE streaming, JWT authentication, and deployment patterns for remote MCP connectors that integrate with ChatGPT and other AI assistants.",
				URL:   "https://github.com/go-go-golems/go-go-mcp/examples",
			},
			{
				ID:    "doc-004",
				Title: "ChatGPT Plugin Development Guide",
				Text:  "Comprehensive guide for developing ChatGPT plugins and connectors. Covers manifest configuration, OAuth flows, API design patterns, and debugging techniques for creating reliable AI integrations.",
				URL:   "https://platform.openai.com/docs/plugins",
			},
			{
				ID:    "doc-005",
				Title: "Auth0 OIDC Configuration",
				Text:  "Configure Auth0 for OpenID Connect with dynamic client registration. Step-by-step instructions for enabling DCR, setting up JWKS endpoints, and implementing JWT validation for modern OAuth workflows.",
				URL:   "https://auth0.com/docs/get-started/applications/dynamic-client-registration",
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
				Int("text_length", len(result.Text)).
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

	// Generate realistic content based on document ID
	var result types.FetchResult

	switch req.ID {
	case "doc-001":
		result = types.FetchResult{
			ID:    req.ID,
			Title: "Introduction to Model Context Protocol",
			Text:  "# Introduction to Model Context Protocol\n\nThe Model Context Protocol (MCP) is an open standard that enables AI assistants to securely connect to data sources. MCP standardizes how AI applications can access context from various sources including databases, APIs, file systems, and other services.\n\n## Key Features\n\n- **Standardized Interface**: Uniform protocol for data access\n- **Security**: Built-in authentication and authorization\n- **Flexibility**: Supports multiple transport layers\n- **Scalability**: Designed for production deployments\n\n## Architecture\n\nMCP uses a client-server architecture where:\n- **MCP Servers** expose data sources and tools\n- **MCP Clients** (like ChatGPT) consume these resources\n- **Transports** handle communication (HTTP, SSE, WebSocket)\n\n## Getting Started\n\n1. Choose a transport layer\n2. Implement server or client\n3. Configure authentication\n4. Define resources and tools\n\nFor more information, visit the official documentation.",
			URL:   "https://modelcontextprotocol.io/introduction",
			Metadata: map[string]interface{}{
				"document_type": "guide",
				"last_updated":  "2024-12-01",
				"word_count":    245,
				"tags":          []string{"mcp", "protocol", "introduction"},
			},
		}
	case "doc-002":
		result = types.FetchResult{
			ID:    req.ID,
			Title: "OAuth 2.1 and Dynamic Client Registration",
			Text:  "# OAuth 2.0 Dynamic Client Registration Protocol (RFC 7591)\n\n## Abstract\n\nThis specification defines methods for clients to dynamically register with OAuth 2.0 authorization servers. This allows clients to obtain client identifiers and optionally client secrets without requiring manual intervention.\n\n## Protocol Flow\n\n1. **Registration Request**: Client sends registration metadata to registration endpoint\n2. **Server Processing**: Authorization server validates and processes the request\n3. **Registration Response**: Server returns client credentials and metadata\n4. **Optional Management**: Client can update or delete registration\n\n## Registration Endpoint\n\nThe registration endpoint accepts HTTP POST requests with client metadata:\n\n```json\n{\n  \"redirect_uris\": [\"https://client.example.org/callback\"],\n  \"token_endpoint_auth_method\": \"client_secret_basic\",\n  \"grant_types\": [\"authorization_code\", \"refresh_token\"],\n  \"response_types\": [\"code\"],\n  \"client_name\": \"My Example Client\",\n  \"scope\": \"openid profile email\"\n}\n```\n\n## Security Considerations\n\n- Validate all client metadata\n- Implement rate limiting\n- Consider client authentication for updates\n- Monitor for abuse patterns",
			URL:   "https://datatracker.ietf.org/doc/html/rfc7591",
			Metadata: map[string]interface{}{
				"document_type": "rfc",
				"rfc_number":    7591,
				"status":        "proposed_standard",
				"category":      "oauth",
			},
		}
	case "doc-003":
		result = types.FetchResult{
			ID:    req.ID,
			Title: "Building Remote MCP Servers with Go",
			Text:  "# Building Remote MCP Servers with Go\n\n## Overview\n\nThis guide demonstrates building production-ready MCP servers using Go and the go-go-mcp SDK.\n\n## Prerequisites\n\n- Go 1.21 or later\n- Basic understanding of HTTP servers\n- Familiarity with OAuth 2.0\n\n## Implementation Steps\n\n### 1. Server Setup\n\n```go\npackage main\n\nimport (\n    \"github.com/go-go-golems/go-go-mcp\"\n)\n\nfunc main() {\n    server := mcp.NewServer()\n    // Configure server...\n}\n```\n\n### 2. Transport Configuration\n\n- **HTTP**: Standard REST API\n- **SSE**: Server-Sent Events for streaming\n- **WebSocket**: Bidirectional communication\n\n### 3. Authentication\n\n- JWT validation\n- OAuth 2.0 flows\n- Custom auth schemes\n\n### 4. Resource Handlers\n\nImplement search and fetch handlers:\n\n```go\nfunc searchHandler(ctx context.Context, req SearchRequest) (<-chan SearchResult, error) {\n    // Implementation\n}\n\nfunc fetchHandler(ctx context.Context, req FetchRequest) (FetchResult, error) {\n    // Implementation\n}\n```\n\n### 5. Deployment\n\n- Docker containers\n- Cloud platforms\n- Load balancing\n- Monitoring\n\n## Best Practices\n\n- Use structured logging\n- Implement health checks\n- Handle graceful shutdown\n- Monitor performance metrics",
			URL:   "https://github.com/go-go-golems/go-go-mcp/examples",
			Metadata: map[string]interface{}{
				"document_type":  "tutorial",
				"difficulty":     "intermediate",
				"language":       "go",
				"estimated_time": "30 minutes",
			},
		}
	case "doc-004":
		result = types.FetchResult{
			ID:    req.ID,
			Title: "ChatGPT Plugin Development Guide",
			Text:  "# ChatGPT Plugin Development Guide\n\n## Introduction\n\nChatGPT plugins extend ChatGPT's capabilities by connecting it to external services and data sources. This guide covers the complete development lifecycle.\n\n## Plugin Architecture\n\n### Manifest File\n\nEvery plugin needs an ai-plugin.json manifest:\n\n```json\n{\n  \"schema_version\": \"v1\",\n  \"name_for_human\": \"My Plugin\",\n  \"name_for_model\": \"my_plugin\",\n  \"description_for_human\": \"Human readable description\",\n  \"description_for_model\": \"Model instruction description\",\n  \"auth\": {\n    \"type\": \"oauth\",\n    \"authorization_url\": \"https://auth.example.com/authorize\",\n    \"token_url\": \"https://auth.example.com/token\",\n    \"scope\": \"read write\"\n  },\n  \"api\": {\n    \"type\": \"openapi\",\n    \"url\": \"https://api.example.com/openapi.yaml\"\n  }\n}\n```\n\n### Authentication Types\n\n1. **None**: No authentication required\n2. **Service Level**: API key authentication\n3. **User Level**: OAuth for user authorization\n\n### API Specification\n\nDefine your API using OpenAPI 3.0:\n\n- Clear endpoint descriptions\n- Parameter validation\n- Response schemas\n- Error handling\n\n## Development Workflow\n\n1. Design plugin functionality\n2. Create manifest and API spec\n3. Implement backend services\n4. Test with ChatGPT\n5. Deploy and monitor\n\n## Testing\n\n- Use ChatGPT plugin developer tools\n- Test OAuth flows\n- Validate API responses\n- Check error handling\n\n## Best Practices\n\n- Keep responses concise\n- Handle rate limiting\n- Implement proper error codes\n- Use descriptive function names\n- Provide clear documentation",
			URL:   "https://platform.openai.com/docs/plugins",
			Metadata: map[string]interface{}{
				"document_type": "documentation",
				"platform":      "openai",
				"last_updated":  "2024-11-15",
				"version":       "2.0",
			},
		}
	case "doc-005":
		result = types.FetchResult{
			ID:    req.ID,
			Title: "Auth0 OIDC Configuration",
			Text:  "# Auth0 OIDC Configuration for Dynamic Client Registration\n\n## Overview\n\nAuth0 supports OpenID Connect Dynamic Client Registration, allowing applications to register themselves programmatically.\n\n## Enable Dynamic Registration\n\n### Dashboard Configuration\n\n1. Go to Auth0 Dashboard\n2. Navigate to Settings → Advanced\n3. Enable \"OIDC Dynamic Application Registration\"\n4. Save changes\n\n### API Configuration\n\n```bash\ncurl -X PATCH 'https://YOUR_DOMAIN/api/v2/tenants/settings' \\\n  -H 'Authorization: Bearer MGMT_API_TOKEN' \\\n  -H 'Content-Type: application/json' \\\n  -d '{\n    \"flags\": {\n      \"enable_dynamic_client_registration\": true\n    }\n  }'\n```\n\n## Registration Endpoint\n\nOnce enabled, clients can register at:\n`https://YOUR_DOMAIN/oidc/register`\n\n### Example Registration\n\n```bash\ncurl -X POST 'https://dev-example.auth0.com/oidc/register' \\\n  -H 'Content-Type: application/json' \\\n  -d '{\n    \"client_name\": \"My Application\",\n    \"redirect_uris\": [\"https://app.example.com/callback\"],\n    \"grant_types\": [\"authorization_code\"],\n    \"response_types\": [\"code\"],\n    \"token_endpoint_auth_method\": \"client_secret_post\"\n  }'\n```\n\n### Response\n\n```json\n{\n  \"client_id\": \"generated_client_id\",\n  \"client_secret\": \"generated_secret\",\n  \"client_id_issued_at\": 1640995200,\n  \"client_secret_expires_at\": 0,\n  \"redirect_uris\": [\"https://app.example.com/callback\"],\n  \"grant_types\": [\"authorization_code\"],\n  \"response_types\": [\"code\"]\n}\n```\n\n## JWT Validation\n\nValidate JWTs using Auth0's JWKS endpoint:\n\n1. Fetch keys from `/.well-known/jwks.json`\n2. Verify JWT signature\n3. Validate claims (iss, aud, exp)\n4. Extract user information\n\n## Security Considerations\n\n- Monitor registration patterns\n- Implement rate limiting\n- Validate redirect URIs\n- Use HTTPS only\n- Rotate secrets regularly",
			URL:   "https://auth0.com/docs/get-started/applications/dynamic-client-registration",
			Metadata: map[string]interface{}{
				"document_type": "configuration_guide",
				"provider":      "auth0",
				"difficulty":    "beginner",
				"category":      "authentication",
			},
		}
	default:
		result = types.FetchResult{
			ID:    req.ID,
			Title: fmt.Sprintf("Document %s", req.ID),
			Text:  fmt.Sprintf("# Document %s\n\nContent for document '%s' is not available. This is a placeholder response from the demo fetch handler.\n\nIn a real implementation, this would contain the actual content retrieved from the data source identified by the ID.", req.ID, req.ID),
			URL:   fmt.Sprintf("https://example.com/docs/%s", req.ID),
		}
	}

	logger.Debug().
		Str("resource_id", req.ID).
		Str("result_id", result.ID).
		Str("result_title", result.Title).
		Str("result_url", result.URL).
		Int("text_length", len(result.Text)).
		Dur("fetch_duration", time.Since(start)).
		Msg("DemoFetchHandler completed successfully")

	return result, nil
}
