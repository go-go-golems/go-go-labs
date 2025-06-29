# HTTP Server Component Documentation

## Overview

The HTTP server component (`pkg/server/http.go`) provides a production-ready HTTP server with SSE (Server-Sent Events) support for the remote MCP (Model Context Protocol) connector. It uses Gorilla Mux for routing and integrates cleanly with the auth validator, MCP server, and transport components.

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ChatGPT       │    │   HTTP Server    │    │   MCP Server    │
│   Client        │◄──►│   Component      │◄──►│   Component     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │   Auth Validator │
                       │   Component      │
                       └──────────────────┘
```

## Route Structure

### Health Check Endpoints (No Auth Required)
- `GET /health` - Basic health check returning JSON status
- `GET /health/ready` - Readiness check validating all components are configured
- `GET /health/live` - Liveness check for load balancer monitoring

### Discovery Endpoints (No Auth Required)
- `GET /.well-known/ai-plugin.json` - ChatGPT plugin manifest
- `GET /.well-known/oauth-authorization-server` - OAuth discovery document

### MCP Endpoints (OAuth Required)
- `GET/POST /sse/*` - SSE transport endpoint with full auth middleware chain

## Middleware Chain

The server implements a comprehensive middleware chain for security and observability:

```go
// Global middleware (applied to all routes)
router.Use(securityHeadersMiddleware)  // Security headers
router.Use(recoveryMiddleware)         // Panic recovery

// SSE-specific middleware  
sseRouter.Use(authMiddleware)          // OAuth token validation
sseRouter.Use(corsMiddleware)          // CORS for SSE
sseRouter.Use(loggingMiddleware)       // Request logging
```

### Auth Middleware

The auth middleware validates OAuth bearer tokens:

1. **Token Extraction**: Extracts `Bearer <token>` from `Authorization` header
2. **Token Validation**: Calls `AuthValidator.ValidateToken()` 
3. **Context Enhancement**: Adds user info to request context
4. **Error Handling**: Returns 401 for invalid/missing tokens

```go
// Usage in your auth validator
func (v *GitHubAuthValidator) ValidateToken(ctx context.Context, token string) (*types.UserInfo, error) {
    // Validate with GitHub API
    // Return user info or error
}
```

### CORS Middleware

Configured for SSE connections with appropriate headers:

```go
Access-Control-Allow-Origin: *              // Configure for production
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type, Accept
Access-Control-Allow-Credentials: true
```

### Security Headers Middleware

Applies security headers following OWASP guidelines:

```go
X-Content-Type-Options: nosniff
X-Frame-Options: DENY  
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
Strict-Transport-Security: max-age=31536000; includeSubDomains  // HTTPS only
```

## SSE Endpoint Configuration

The `/sse/*` endpoint is configured with:

- **Authentication**: OAuth bearer token validation
- **CORS**: Appropriate headers for cross-origin SSE
- **Security**: Full security header suite
- **Logging**: Structured request/response logging
- **Recovery**: Panic recovery with proper error responses

### Testing SSE Connections

1. **Manual Testing with curl**:
```bash
# Health check
curl -X GET http://localhost:8080/health

# SSE connection (requires valid token)
curl -H "Authorization: Bearer <github-token>" \
     -H "Accept: text/event-stream" \
     http://localhost:8080/sse
```

2. **Testing with MCP Inspector**:
```bash
# If using MCP dev tools
mcp-inspector connect http://localhost:8080/sse \
  --auth-header "Authorization: Bearer <token>"
```

3. **Integration Testing**:
```go
func TestSSEEndpoint(t *testing.T) {
    server := setupTestServer()
    req := httptest.NewRequest("GET", "/sse", nil)
    req.Header.Set("Authorization", "Bearer valid-token")
    
    w := httptest.NewRecorder()
    server.router.ServeHTTP(w, req)
    
    assert.Equal(t, http.StatusOK, w.Code)
    assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
}
```

## Health Check Endpoints

### `/health` - Basic Health
Returns basic server status for quick health checks:

```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### `/health/ready` - Readiness Check
Validates all components are properly configured:

```json
{
  "ready": true,
  "components": {
    "auth": true,
    "mcp": true,
    "transport": true,  
    "discovery": true
  },
  "timestamp": "2025-01-15T10:30:00Z"
}
```

### `/health/live` - Liveness Check
Simple liveness indicator for load balancer monitoring:

```json
{
  "alive": true,
  "timestamp": "2025-01-15T10:30:00Z"
}
```

## Component Integration

### Setting Up the Server

```go
package main

import (
    "context"
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/server"
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/types"
)

func main() {
    config := &types.Config{
        Host: "localhost",
        Port: 8080,
        GitHubClientID: os.Getenv("GITHUB_CLIENT_ID"),
        GitHubClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
        AllowedLogin: os.Getenv("GITHUB_ALLOWED_LOGIN"),
    }
    
    // Create server
    httpServer := server.NewHTTPServer(config)
    
    // Wire up components (order matters for dependency injection)
    authValidator := auth.NewGitHubValidator(config)
    mcpServer := mcp.NewServer(config)  
    transport := sse.NewTransport()
    discovery := discovery.NewService(config, authValidator)
    
    // Configure dependencies
    httpServer.SetAuthValidator(authValidator)
    httpServer.SetMCPServer(mcpServer)
    httpServer.SetTransport(transport) // Auto-wires auth and MCP
    httpServer.SetDiscoveryService(discovery)
    
    // Start server
    ctx := context.Background()
    if err := httpServer.Start(ctx); err != nil {
        log.Fatal().Err(err).Msg("Failed to start server")
    }
    
    // Graceful shutdown
    <-ctx.Done()
    httpServer.Stop(ctx)
}
```

### MCP Server Integration

Your MCP server component should implement the `types.MCPServer` interface:

```go
type MCPServer interface {
    RegisterSearch(handler SearchHandler) error
    RegisterFetch(handler FetchHandler) error
    Start(ctx context.Context) error
    Stop() error
    GetHTTPHandler() http.Handler
}
```

### Auth Validator Integration

Your auth validator should implement the `types.AuthValidator` interface:

```go
type AuthValidator interface {
    ValidateToken(ctx context.Context, token string) (*UserInfo, error)
    GetAuthEndpoints() AuthEndpoints
}
```

### Transport Integration

Your SSE transport should implement the `types.Transport` interface:

```go
type Transport interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
    SetAuthValidator(validator AuthValidator)
    SetMCPServer(server MCPServer)
}
```

## Configuration

The server accepts configuration through the `types.Config` struct:

```go
type Config struct {
    Port int    `json:"port"`           // HTTP server port
    Host string `json:"host"`           // HTTP server host
    
    GitHubClientID     string `json:"github_client_id"`     // OAuth client ID
    GitHubClientSecret string `json:"github_client_secret"` // OAuth client secret  
    AllowedLogin       string `json:"allowed_login"`        // Allowed GitHub username
    
    LogLevel string `json:"log_level"`  // Logging level
}
```

## Security Considerations

1. **OAuth Token Validation**: All SSE endpoints require valid bearer tokens
2. **CORS Configuration**: Configure `Access-Control-Allow-Origin` appropriately for production
3. **HTTPS Enforcement**: Use HTTPS in production (HSTS headers are automatically added)
4. **Rate Limiting**: Consider adding rate limiting middleware for production
5. **Request Size Limits**: Consider adding request size limits
6. **Timeout Configuration**: Configured with reasonable timeouts for all operations

## Logging

The server uses structured logging with zerolog:

- **Request Logging**: All HTTP requests are logged with method, path, status, duration
- **Auth Logging**: Token validation attempts are logged (without exposing tokens)
- **Error Logging**: All errors are logged with appropriate context
- **Debug Logging**: Component configuration and internal state changes

## Production Deployment

### Load Balancer Configuration

Configure your load balancer to use the health check endpoints:

```yaml
# Example nginx upstream config
upstream mcp_server {
    server localhost:8080;
    # Health check on /health/live
}
```

### Environment Variables

```bash
# Required
export GITHUB_CLIENT_ID="Iv1.xxxxxxxx"
export GITHUB_CLIENT_SECRET="xxxxxxxx"
export GITHUB_ALLOWED_LOGIN="your-username"

# Optional
export HOST="0.0.0.0"         # Default: localhost
export PORT="8080"            # Default: 8080  
export LOG_LEVEL="info"       # Default: info
```

### Docker Configuration

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server ./cmd/experiments/remote-chatgpt-connector

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/server .
EXPOSE 8080
CMD ["./server"]
```

## Monitoring

The server provides comprehensive monitoring capabilities:

- **Health Endpoints**: `/health`, `/health/ready`, `/health/live`
- **Structured Logging**: JSON-formatted logs for log aggregation
- **Request Metrics**: Duration, status codes, user agents
- **Error Tracking**: Panic recovery with stack traces

## Next Steps

1. **Implement Auth Validator**: Create the GitHub OAuth validator component
2. **Implement MCP Server**: Create the MCP server with search/fetch capabilities  
3. **Implement SSE Transport**: Create the Server-Sent Events transport layer
4. **Implement Discovery Service**: Create the well-known endpoints service
5. **Add Integration Tests**: Test the complete request flow
6. **Add Performance Tests**: Load test the SSE connections

The HTTP server component provides the foundation for a production-ready remote MCP connector with proper security, monitoring, and graceful shutdown handling.
