# Discovery Service Component

The Discovery Service handles the `.well-known` endpoints required for ChatGPT connector integration, providing both the plugin manifest and OAuth authorization server configuration.

## Overview

This component implements the `DiscoveryService` interface from the types package and provides two critical endpoints:

1. `/.well-known/ai-plugin.json` - ChatGPT plugin manifest
2. `/.well-known/oauth-authorization-server` - OAuth authorization server metadata

## Implementation Details

### Plugin Manifest Structure

The plugin manifest follows ChatGPT's connector specification:

```json
{
  "name": "GitHub MCP Connector",
  "description": "Secure MCP server with GitHub OAuth authentication for personal use",
  "version": "1.0.0",
  "auth": {
    "type": "oauth",
    "authorization_url": "https://github.com/login/oauth/authorize",
    "token_url": "https://github.com/login/oauth/access_token",
    "scopes": ["read:user"]
  },
  "api": {
    "type": "mcp",
    "url": "/sse"
  },
  "contact": {
    "email": "admin@example.com"
  },
  "legal": {
    "privacy_url": "https://example.com/privacy",
    "terms_url": "https://example.com/terms"
  }
}
```

### OAuth Authorization Server Configuration

The OAuth config provides GitHub's OAuth endpoints and capabilities:

```json
{
  "issuer": "https://github.com",
  "authorization_endpoint": "https://github.com/login/oauth/authorize",
  "token_endpoint": "https://github.com/login/oauth/access_token",
  "scopes_supported": ["read:user", "user:email", "public_repo", "repo"],
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code"],
  "code_challenge_methods_supported": ["S256", "plain"],
  "token_endpoint_auth_methods_supported": ["client_secret_post", "client_secret_basic"]
}
```

## Key Features

### GitHub OAuth Integration

- **Authorization URL**: `https://github.com/login/oauth/authorize`
- **Token URL**: `https://github.com/login/oauth/access_token`
- **Required Scope**: `read:user` (minimum for user identification)
- **PKCE Support**: Includes S256 and plain code challenge methods

### Security Headers

All endpoints include proper CORS headers:
- `Content-Type: application/json`
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type, Authorization`

### Comprehensive Logging

Using zerolog for structured logging:
- Request tracking with method, path, user agent, and remote address
- Error logging with context
- Debug logging for development

## Usage

### Initialization

```go
import (
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/discovery"
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/config"
)

cfg, err := config.Load()
if err != nil {
    log.Fatal().Err(err).Msg("failed to load config")
}

discoveryService := discovery.NewService(cfg)
```

### Route Registration

```go
mux := http.NewServeMux()
discoveryService.RegisterRoutes(mux)

server := &http.Server{
    Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
    Handler: mux,
}
```

### Programmatic Access

```go
// Get plugin manifest as JSON bytes
manifest, err := discoveryService.GetPluginManifest()
if err != nil {
    log.Error().Err(err).Msg("failed to get manifest")
}

// Get OAuth config as JSON bytes
oauthConfig, err := discoveryService.GetOAuthConfig()
if err != nil {
    log.Error().Err(err).Msg("failed to get oauth config")
}

// Get URLs for testing
fmt.Println("Manifest URL:", discoveryService.GetManifestURL())
fmt.Println("OAuth Config URL:", discoveryService.GetOAuthConfigURL())
```

## Testing Endpoints

### 1. Test Plugin Manifest

```bash
# Basic curl test
curl -X GET http://localhost:8080/.well-known/ai-plugin.json

# With headers
curl -X GET \
  -H "Accept: application/json" \
  -H "User-Agent: ChatGPT/1.0" \
  http://localhost:8080/.well-known/ai-plugin.json

# Expected response: JSON manifest with GitHub OAuth config
```

### 2. Test OAuth Authorization Server

```bash
# Basic curl test
curl -X GET http://localhost:8080/.well-known/oauth-authorization-server

# With headers
curl -X GET \
  -H "Accept: application/json" \
  http://localhost:8080/.well-known/oauth-authorization-server

# Expected response: JSON with GitHub OAuth endpoints
```

### 3. Validate JSON Schema

```bash
# Install jq for JSON validation
sudo apt-get install jq

# Validate manifest structure
curl -s http://localhost:8080/.well-known/ai-plugin.json | jq '.'

# Check required fields
curl -s http://localhost:8080/.well-known/ai-plugin.json | jq '.auth.type'
# Should return: "oauth"

curl -s http://localhost:8080/.well-known/ai-plugin.json | jq '.auth.authorization_url'
# Should return: "https://github.com/login/oauth/authorize"
```

### 4. Test CORS Headers

```bash
# Test CORS preflight
curl -X OPTIONS \
  -H "Origin: https://chat.openai.com" \
  -H "Access-Control-Request-Method: GET" \
  -H "Access-Control-Request-Headers: Content-Type" \
  http://localhost:8080/.well-known/ai-plugin.json

# Check CORS headers in response
curl -X GET \
  -H "Origin: https://chat.openai.com" \
  -v http://localhost:8080/.well-known/ai-plugin.json 2>&1 | grep -i "access-control"
```

## GitHub OAuth Flow Documentation

### 1. OAuth App Registration

Before using the connector, register a GitHub OAuth App:

1. Go to **Settings → Developer settings → OAuth Apps**
2. Click **New OAuth App**
3. Fill in the application details:
   - **Application name**: "ChatGPT MCP Connector"
   - **Homepage URL**: Your server URL (e.g., `http://localhost:8080`)
   - **Authorization callback URL**: `https://chat.openai.com/aip/callback` (ChatGPT's callback)
   - **Description**: "Personal MCP connector for ChatGPT"

4. Save the **Client ID** and **Client Secret**

### 2. Environment Configuration

Set the required environment variables:

```bash
export GITHUB_CLIENT_ID="Iv1.your_client_id_here"
export GITHUB_CLIENT_SECRET="your_client_secret_here"
export GITHUB_ALLOWED_LOGIN="your_github_username"
export PORT=8080
export LOG_LEVEL=debug
```

### 3. OAuth Flow Steps

1. **Discovery**: ChatGPT fetches `/.well-known/ai-plugin.json`
2. **Authorization**: User redirected to GitHub OAuth (`/login/oauth/authorize`)
3. **Consent**: User grants `read:user` permission
4. **Token Exchange**: ChatGPT exchanges code for access token
5. **Connection**: ChatGPT connects to `/sse` with `Authorization: Bearer <token>`

### 4. Security Considerations

- **Single User**: Only configured GitHub user can access
- **Token Validation**: Every request validates token with GitHub API
- **Rate Limiting**: GitHub API has 5k requests/hour limit
- **HTTPS**: Use HTTPS in production (GitHub requires it)

## ChatGPT Connector Integration

### Step 1: Start the Server

```bash
# Set environment variables
export GITHUB_CLIENT_ID="your_client_id"
export GITHUB_CLIENT_SECRET="your_client_secret"  
export GITHUB_ALLOWED_LOGIN="your_username"

# Start the server
go run ./cmd/experiments/remote-chatgpt-connector
```

### Step 2: Register with ChatGPT

1. Open ChatGPT and go to **Settings**
2. Navigate to **Beta Features** or **Data Controls**
3. Find **Connectors** or **Plugins** section
4. Click **Add Connector** or **Add Plugin**
5. Enter your server URL: `http://localhost:8080` (or your domain)
6. ChatGPT will fetch the manifest and show the connector details

### Step 3: OAuth Authorization

1. ChatGPT will show "Authorize GitHub MCP Connector"
2. Click **Continue** to start OAuth flow
3. You'll be redirected to GitHub's authorization page
4. Click **Authorize** to grant `read:user` permission
5. You'll be redirected back to ChatGPT
6. The connector should now be active

### Step 4: Test the Connection

Start a new chat and try:

```
"Use the GitHub MCP Connector to search for information about Go programming"
```

ChatGPT should establish an SSE connection to your server and stream search results.

## Troubleshooting

### Common Issues

1. **Manifest Not Found (404)**
   - Check that server is running on correct port
   - Verify the URL path is exactly `/.well-known/ai-plugin.json`
   - Check firewall/network settings

2. **CORS Issues**
   - Ensure CORS headers are properly set
   - Check browser developer tools for CORS errors
   - Try different User-Agent headers

3. **OAuth Configuration Errors**
   - Verify GitHub OAuth app settings
   - Check callback URL configuration
   - Ensure client ID/secret are correct

4. **JSON Parsing Errors**
   - Validate JSON structure with `jq`
   - Check for missing required fields
   - Verify data types match specification

### Debug Commands

```bash
# Check server health
curl -v http://localhost:8080/.well-known/ai-plugin.json

# Test with different user agents
curl -H "User-Agent: Mozilla/5.0" http://localhost:8080/.well-known/ai-plugin.json
curl -H "User-Agent: ChatGPT-User/1.0" http://localhost:8080/.well-known/ai-plugin.json

# Validate JSON structure
curl -s http://localhost:8080/.well-known/ai-plugin.json | python -m json.tool

# Check response headers
curl -I http://localhost:8080/.well-known/ai-plugin.json
```

### Log Analysis

Enable debug logging to see detailed request information:

```bash
export LOG_LEVEL=debug
```

Look for log entries like:
- `serving plugin manifest` - Manifest requests
- `serving oauth config` - OAuth config requests
- `discovery routes registered successfully` - Successful startup

## Production Deployment

### HTTPS Requirements

GitHub OAuth requires HTTPS in production:

```bash
# Using Let's Encrypt with certbot
sudo certbot --nginx -d your-domain.com

# Update OAuth app callback URL
# https://your-domain.com/.well-known/ai-plugin.json
```

### Environment Variables

```bash
# Production settings
export GITHUB_CLIENT_ID="your_production_client_id"
export GITHUB_CLIENT_SECRET="your_production_client_secret"
export GITHUB_ALLOWED_LOGIN="your_username"
export HOST="0.0.0.0"
export PORT=443
export LOG_LEVEL=info
```

### Security Checklist

- [ ] Use HTTPS in production
- [ ] Rotate GitHub client secret regularly
- [ ] Monitor rate limits and implement caching
- [ ] Use environment variables for secrets
- [ ] Implement proper logging and monitoring
- [ ] Set up firewall rules
- [ ] Use reverse proxy (nginx/traefik)

## API Reference

### DiscoveryService Interface

```go
type DiscoveryService interface {
    GetPluginManifest() ([]byte, error)
    GetOAuthConfig() ([]byte, error)
    RegisterRoutes(mux *http.ServeMux)
}
```

### Service Methods

```go
// NewService creates a new discovery service
func NewService(config *types.Config) *Service

// GetPluginManifest returns the ChatGPT plugin manifest
func (s *Service) GetPluginManifest() ([]byte, error)

// GetOAuthConfig returns the OAuth authorization server configuration  
func (s *Service) GetOAuthConfig() ([]byte, error)

// RegisterRoutes registers the discovery endpoints
func (s *Service) RegisterRoutes(mux *http.ServeMux)

// Helper methods for testing
func (s *Service) GetBaseURL() string
func (s *Service) GetManifestURL() string
func (s *Service) GetOAuthConfigURL() string
```

## Example Integration

Here's a complete example of integrating the discovery service:

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/config"
    "github.com/go-go-golems/go-go-labs/cmd/experiments/remote-chatgpt-connector/pkg/discovery"
    "github.com/rs/zerolog"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal("Failed to load config:", err)
    }
    
    // Setup logging
    zerolog.SetGlobalLevel(zerolog.InfoLevel)
    if cfg.LogLevel == "debug" {
        zerolog.SetGlobalLevel(zerolog.DebugLevel)
    }
    
    // Create discovery service
    discoveryService := discovery.NewService(cfg)
    
    // Setup HTTP server
    mux := http.NewServeMux()
    discoveryService.RegisterRoutes(mux)
    
    // Add health check
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    // Start server
    addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
    fmt.Printf("Server starting on %s\n", addr)
    fmt.Printf("Plugin manifest: %s\n", discoveryService.GetManifestURL())
    fmt.Printf("OAuth config: %s\n", discoveryService.GetOAuthConfigURL())
    
    if err := http.ListenAndServe(addr, mux); err != nil {
        log.Fatal("Server failed:", err)
    }
}
```

This discovery service provides the foundation for ChatGPT connector integration, handling all the necessary .well-known endpoints and OAuth configuration for GitHub authentication.
