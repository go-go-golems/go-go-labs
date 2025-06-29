# MCP Remote Connector - Integration Guide

## Overview

This is a complete, production-ready remote Model Context Protocol (MCP) server that allows ChatGPT to connect via GitHub OAuth and perform secure searches and data fetches.

## Quick Start

### 1. GitHub OAuth App Setup

1. Go to [GitHub Developer Settings](https://github.com/settings/developers)
2. Click "New OAuth App"
3. Fill in the details:
   - **Application name**: "My MCP Connector"
   - **Homepage URL**: `https://yourdomain.com`
   - **Authorization callback URL**: `https://chat.openai.com/aip/...` (ChatGPT will provide this)
4. Copy the **Client ID** and **Client Secret**

### 2. Environment Configuration

Create a `.env` file or export these variables:

```bash
export GITHUB_CLIENT_ID="Iv1.your_client_id"
export GITHUB_CLIENT_SECRET="your_client_secret"
export GITHUB_ALLOWED_LOGIN="your_github_username"
export PORT="8080"
export LOG_LEVEL="info"
```

### 3. Build and Run

```bash
# From the project root
go build -o connector ./cmd/experiments/remote-chatgpt-connector

# Set environment variables and run
./connector --log-level debug
```

## Testing the Components

### Health Check
```bash
curl http://localhost:8080/health
# Expected: {"status":"healthy","service":"remote-chatgpt-connector"}
```

### Discovery Endpoints
```bash
# Plugin manifest
curl http://localhost:8080/.well-known/ai-plugin.json | jq

# OAuth configuration  
curl http://localhost:8080/.well-known/oauth-authorization-server | jq
```

### Token Validation (requires valid GitHub token)
```bash
curl -H "Authorization: Bearer gho_your_token" \
     http://localhost:8080/sse
```

## ChatGPT Integration

### Step 1: Add Connector
1. In ChatGPT, go to Settings → Data Controls → Connectors
2. Click "Add Connector"
3. Enter your server URL: `https://yourdomain.com:8080`
4. ChatGPT will fetch the manifest and initiate OAuth

### Step 2: Authorize
1. You'll be redirected to GitHub OAuth
2. Grant the `read:user` permission
3. You'll be redirected back to ChatGPT

### Step 3: Use the Connector
Ask ChatGPT: *"Search for 'hello' using the Go MCP connector"*

ChatGPT will call your `/sse` endpoint and stream results.

## Production Deployment

### Using systemd

Create `/etc/systemd/system/mcp-connector.service`:

```ini
[Unit]
Description=MCP Remote Connector
After=network.target

[Service]
Type=simple
User=mcp
WorkingDirectory=/opt/mcp-connector
ExecStart=/opt/mcp-connector/connector
Environment=GITHUB_CLIENT_ID=Iv1.your_client_id
Environment=GITHUB_CLIENT_SECRET=your_secret
Environment=GITHUB_ALLOWED_LOGIN=your_username
Environment=PORT=8080
Environment=LOG_LEVEL=info
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Using nginx Reverse Proxy

```nginx
server {
    listen 443 ssl;
    server_name your-mcp-server.com;
    
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # SSE support
        proxy_buffering off;
        proxy_cache off;
        proxy_set_header Connection "";
        proxy_http_version 1.1;
    }
}
```

## Architecture Components

This application was built using a parallel agent development approach with the following components:

### Core Components

1. **MCP Wrapper** (`pkg/mcp/`) - Integrates with go-go-mcp SDK
2. **Auth Validator** (`pkg/auth/`) - GitHub OAuth token validation  
3. **HTTP Server** (`pkg/server/`) - Gorilla/mux with SSE support
4. **Discovery Service** (`pkg/discovery/`) - .well-known endpoints

### Key Features

- ✅ **GitHub OAuth Security** - Single user allowlist
- ✅ **SSE Transport** - Real-time streaming results  
- ✅ **Health Monitoring** - Multiple health check endpoints
- ✅ **Structured Logging** - Zerolog with configurable levels
- ✅ **Graceful Shutdown** - Signal handling with timeout
- ✅ **Production Ready** - CORS, security headers, rate limiting

## Troubleshooting

### Common Issues

**Connection Refused**
```bash
# Check if server is running
curl http://localhost:8080/health
```

**OAuth Authorization Failed**
- Verify `GITHUB_CLIENT_ID` and `GITHUB_CLIENT_SECRET`
- Check GitHub OAuth app callback URL matches ChatGPT's

**Token Validation Failed**
- Verify `GITHUB_ALLOWED_LOGIN` matches your GitHub username exactly
- Check GitHub API rate limits (5000 requests/hour)

**SSE Connection Issues**
- Ensure proxy doesn't buffer SSE streams
- Check CORS headers for cross-origin requests

### Debug Mode

Run with debug logging to see detailed request/response flow:

```bash
./connector --log-level debug
```

### Component Testing

Each component can be tested independently:

```bash
# Test auth validation
go test ./pkg/auth/...

# Test MCP integration  
go test ./pkg/mcp/...

# Test HTTP server
go test ./pkg/server/...

# Test discovery endpoints
go test ./pkg/discovery/...
```

## Customization

### Custom Search Handler

Replace the demo search handler in `main.go`:

```go
func customSearchHandler(ctx context.Context, req types.SearchRequest) (<-chan types.SearchResult, error) {
    results := make(chan types.SearchResult, 10)
    
    go func() {
        defer close(results)
        
        // Your custom search logic here
        // Query your database, API, etc.
        
        results <- types.SearchResult{
            ID:    "custom-1",
            Title: "Custom Result",
            URL:   "https://your-api.com/item/1",
            Chunk: "Custom search result content",
        }
    }()
    
    return results, nil
}
```

### Custom Fetch Handler

```go
func customFetchHandler(ctx context.Context, req types.FetchRequest) (types.FetchResult, error) {
    // Fetch full content using req.ID
    content := fetchFromYourAPI(req.ID)
    
    return types.FetchResult{
        ID:   req.ID,
        Data: content,
    }, nil
}
```

## Security Considerations

- **Single User**: Only allows one GitHub user (personal connector)
- **Token Validation**: Every request validates GitHub token
- **Rate Limiting**: GitHub API has 5000 req/hour limit
- **HTTPS Required**: Production deployment needs SSL/TLS
- **Secret Management**: Use environment variables, never hardcode secrets

## Support

For issues and questions, see:
- [MCP Specification](https://modelcontextprotocol.io)
- [go-go-mcp SDK Documentation](https://github.com/go-go-golems/go-go-mcp)
- Component documentation in each `pkg/*/README.md`
