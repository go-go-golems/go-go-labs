# MCP Component Documentation

This document describes the MCP (Model Context Protocol) wrapper component that integrates the go-go-mcp SDK for the remote ChatGPT connector.

## Overview

The MCP component provides a clean abstraction over the go-go-mcp SDK, specifically designed for building remote MCP servers that work with ChatGPT's connector system. It implements:

- SSE (Server-Sent Events) transport for real-time communication
- Search and fetch handlers for data retrieval
- Authentication middleware integration
- JSON-RPC 2.0 protocol handling
- Demo implementations for testing

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ChatGPT       │◄──►│  MCP Wrapper     │◄──►│  Your Handlers  │
│   Client        │    │  (This Component)│    │  (Search/Fetch) │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌──────────────────┐
                       │  go-go-mcp SDK   │
                       │  (Transport)     │
                       └──────────────────┘
```

## Core Components

### MCPWrapper

The main wrapper class that manages:
- Server lifecycle (Start/Stop)
- Handler registration (Search/Fetch)
- Authentication validation
- HTTP handler provision
- Logging and error handling

### RequestHandler

Internal handler that bridges between the go-go-mcp SDK and your custom handlers:
- Processes JSON-RPC requests
- Handles MCP protocol methods
- Converts between protocol formats
- Manages request/response flow

## API Usage

### Basic Setup

```go
package main

import (
    "context"
    "log"
    
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/mcp"
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
    "github.com/rs/zerolog"
)

func main() {
    logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
    
    config := types.Config{
        Port:     8080,
        LogLevel: "info",
    }
    
    // Create MCP wrapper
    mcpWrapper, err := mcp.NewMCPWrapper(logger, config)
    if err != nil {
        log.Fatal(err)
    }
    
    // Register handlers
    err = mcpWrapper.RegisterSearch(mcp.DemoSearchHandler)
    if err != nil {
        log.Fatal(err)
    }
    
    err = mcpWrapper.RegisterFetch(mcp.DemoFetchHandler)
    if err != nil {
        log.Fatal(err)
    }
    
    // Start server
    ctx := context.Background()
    if err := mcpWrapper.Start(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Keep running
    select {}
}
```

### Custom Handler Implementation

#### Search Handler

```go
func MySearchHandler(ctx context.Context, req types.SearchRequest) (<-chan types.SearchResult, error) {
    results := make(chan types.SearchResult, 10)
    
    go func() {
        defer close(results)
        
        // Your search logic here
        searchResults := performSearch(req.Query, req.Context)
        
        for _, result := range searchResults {
            select {
            case results <- types.SearchResult{
                ID:    result.ID,
                Title: result.Title,
                URL:   result.URL,
                Chunk: result.Summary,
            }:
            case <-ctx.Done():
                return
            }
        }
    }()
    
    return results, nil
}
```

#### Fetch Handler

```go
func MyFetchHandler(ctx context.Context, req types.FetchRequest) (types.FetchResult, error) {
    // Your fetch logic here
    content, err := fetchContentByID(req.ID)
    if err != nil {
        return types.FetchResult{}, err
    }
    
    return types.FetchResult{
        ID:   req.ID,
        Data: content,
    }, nil
}
```

### Authentication Integration

```go
// Set up authentication validator
authValidator := &MyAuthValidator{
    // Your auth implementation
}

mcpWrapper.SetAuthValidator(authValidator)
```

### HTTP Integration

```go
// Get HTTP handler for integration with web server
httpHandler := mcpWrapper.GetHTTPHandler()

// Use with standard HTTP server
http.Handle("/sse", httpHandler)

// Or with a router like gorilla/mux
router := mux.NewRouter()
router.Handle("/sse", httpHandler)
```

## Handler Signatures

### SearchHandler

```go
type SearchHandler func(ctx context.Context, req SearchRequest) (<-chan SearchResult, error)

type SearchRequest struct {
    Query   string            `json:"query"`
    Context map[string]string `json:"context,omitempty"`
}

type SearchResult struct {
    ID    string `json:"id"`
    Title string `json:"title"`
    URL   string `json:"url"`
    Chunk string `json:"chunk"`  // Brief text snippet
}
```

**Expected Behavior:**
- Return a channel that streams search results
- Close the channel when search is complete
- Handle context cancellation
- Provide meaningful IDs for fetch operations
- Keep chunks brief (summary/snippet)

### FetchHandler

```go
type FetchHandler func(ctx context.Context, req FetchRequest) (FetchResult, error)

type FetchRequest struct {
    ID string `json:"id"`
}

type FetchResult struct {
    ID   string `json:"id"`
    Data string `json:"data"`  // Full content
}
```

**Expected Behavior:**
- Return full content for the given ID
- Handle cases where ID is not found
- Support any ID format returned by search
- Return complete data (not just snippets)

## Mock Server Setup

For testing other components, you can create a mock MCP server:

```go
func CreateMockMCPServer(t *testing.T) *mcp.MCPWrapper {
    logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
    
    config := types.Config{
        Port:     0, // Random port for testing
        LogLevel: "disabled",
    }
    
    wrapper, err := mcp.NewMCPWrapper(logger, config)
    require.NoError(t, err)
    
    // Register mock handlers
    err = wrapper.RegisterSearch(func(ctx context.Context, req types.SearchRequest) (<-chan types.SearchResult, error) {
        results := make(chan types.SearchResult, 1)
        go func() {
            defer close(results)
            results <- types.SearchResult{
                ID:    "mock-1",
                Title: "Mock Result",
                URL:   "https://mock.example.com",
                Chunk: "Mock search result for testing",
            }
        }()
        return results, nil
    })
    require.NoError(t, err)
    
    err = wrapper.RegisterFetch(func(ctx context.Context, req types.FetchRequest) (types.FetchResult, error) {
        return types.FetchResult{
            ID:   req.ID,
            Data: "Mock content for " + req.ID,
        }, nil
    })
    require.NoError(t, err)
    
    return wrapper
}
```

## Expected Response Formats

### Search Response

```json
[
  {
    "id": "doc-123",
    "title": "Document Title",
    "url": "https://example.com/doc-123",
    "chunk": "Brief summary or excerpt from the document..."
  },
  {
    "id": "page-456",
    "title": "Another Document",
    "url": "https://example.com/page-456", 
    "chunk": "Another brief summary..."
  }
]
```

### Fetch Response

```json
[
  {
    "uri": "doc-123",
    "mimeType": "text/plain",
    "text": "Full document content goes here...\n\nThis would be the complete text of the document or resource."
  }
]
```

## MCP Protocol Methods

The wrapper handles these MCP protocol methods:

- `initialize` - Returns server capabilities
- `resources/search` - Delegates to your search handler
- `resources/read` - Delegates to your fetch handler
- `resources/list` - Returns empty list (using search paradigm)

## Dependencies

### Required Dependencies

- `github.com/go-go-golems/go-go-mcp` - Core MCP SDK
- `github.com/rs/zerolog` - Logging
- `github.com/pkg/errors` - Error handling

### Optional Dependencies

- `github.com/gorilla/mux` - HTTP routing (if using with web server)
- `github.com/stretchr/testify` - Testing utilities

## Configuration

The wrapper accepts a `types.Config` struct:

```go
type Config struct {
    Port     int    `json:"port"`      // Server port (default: 8080)
    Host     string `json:"host"`      // Server host (default: all interfaces)
    LogLevel string `json:"log_level"` // Logging level
    
    // GitHub OAuth config (for auth validator)
    GitHubClientID     string `json:"github_client_id"`
    GitHubClientSecret string `json:"github_client_secret"`
    AllowedLogin       string `json:"allowed_login"`
}
```

## Error Handling

The component provides comprehensive error handling:

- **Initialization errors**: Returned during `NewMCPWrapper`
- **Handler registration errors**: Returned during `RegisterSearch`/`RegisterFetch`
- **Runtime errors**: Logged and returned as JSON-RPC error responses
- **Authentication errors**: Return HTTP 401 Unauthorized

## Logging

Structured logging is provided throughout:

```go
// Set log level
logger := zerolog.New(os.Stdout).Level(zerolog.InfoLevel)

// Component logs include:
// - component: "mcp_wrapper"
// - method: JSON-RPC method name
// - user_id/user_login: For authenticated requests
// - batch_size: For batch requests
// - Error details with stack traces
```

## Testing

### Unit Testing

```go
func TestMCPWrapper(t *testing.T) {
    logger := zerolog.New(os.Stdout).Level(zerolog.Disabled)
    config := types.Config{Port: 0}
    
    wrapper, err := mcp.NewMCPWrapper(logger, config)
    require.NoError(t, err)
    
    // Test handler registration
    err = wrapper.RegisterSearch(mcp.DemoSearchHandler)
    assert.NoError(t, err)
    
    err = wrapper.RegisterFetch(mcp.DemoFetchHandler)
    assert.NoError(t, err)
    
    // Test server lifecycle
    ctx := context.Background()
    err = wrapper.Start(ctx)
    assert.NoError(t, err)
    
    assert.True(t, wrapper.IsRunning())
    
    err = wrapper.Stop()
    assert.NoError(t, err)
    
    assert.False(t, wrapper.IsRunning())
}
```

### Integration Testing

```go
func TestMCPIntegration(t *testing.T) {
    // Start server
    wrapper := CreateMockMCPServer(t)
    ctx := context.Background()
    
    err := wrapper.Start(ctx)
    require.NoError(t, err)
    defer wrapper.Stop()
    
    // Test HTTP endpoints
    httpHandler := wrapper.GetHTTPHandler()
    
    // Test SSE connection
    // Test JSON-RPC requests
    // Test authentication
}
```

## Production Considerations

### Security

- Always validate authentication tokens
- Implement proper CORS policies
- Use HTTPS in production
- Rate limit requests
- Sanitize input data

### Performance

- Use buffered channels for search results
- Implement context cancellation
- Set appropriate timeouts
- Monitor memory usage for large result sets
- Consider connection pooling

### Monitoring

- Log all requests and responses
- Monitor authentication failures
- Track response times
- Alert on error rates

## Troubleshooting

### Common Issues

1. **"Search handler not registered"**
   - Call `RegisterSearch()` before `Start()`

2. **"Cannot register handlers while server is running"**
   - Register all handlers before calling `Start()`

3. **Authentication failures**
   - Verify token format and validation logic
   - Check OAuth configuration

4. **SSE connection issues**
   - Verify port accessibility
   - Check firewall settings
   - Ensure proper CORS headers

### Debug Logging

Enable debug logging to troubleshoot:

```go
logger := zerolog.New(os.Stdout).Level(zerolog.DebugLevel)
```

This will show:
- Incoming requests and parameters
- Handler execution flow
- Authentication attempts
- Transport-level events

## Integration Points

This component is designed to integrate with:

1. **Auth Component**: Provides `AuthValidator` interface
2. **Discovery Component**: Uses the HTTP handler for /.well-known endpoints
3. **Transport Component**: Implements the `MCPServer` interface
4. **Main Application**: Provides the complete MCP server functionality

The integration agent should wire these components together using the interfaces defined in `pkg/types/types.go`.
