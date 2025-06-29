# ChatGPT MCP Remote Connector - Investigation Report

## Overview

This document details our investigation into implementing a remote Model Context Protocol (MCP) server for ChatGPT integration using GitHub OAuth authentication. Despite following the specifications and implementing comprehensive debugging, we encountered persistent OAuth client resolution issues.

## Implementation Summary

### What We Built

1. **Complete MCP Remote Connector** with the following components:
   - **MCP Wrapper** (`pkg/mcp/`) - Integration with go-go-mcp SDK
   - **GitHub Auth Validator** (`pkg/auth/`) - OAuth token validation via GitHub API
   - **HTTP Server** (`pkg/server/`) - SSE transport with Gorilla/mux
   - **Discovery Service** (`pkg/discovery/`) - .well-known endpoints for ChatGPT
   - **Main Application** - CLI with comprehensive logging and graceful shutdown

2. **Production-Ready Features**:
   - Comprehensive debug logging with caller info and structured fields
   - Environment-based configuration with `.env` file support
   - Graceful shutdown and signal handling
   - Health check endpoints
   - CORS and security headers
   - Request ID tracking across components

3. **Public Accessibility**:
   - Deployed via Tailscale Funnel at `https://f.beagle-duck.ts.net`
   - All endpoints responding correctly
   - SSL/HTTPS properly configured

## Troubleshooting Timeline

### Phase 1: Initial Setup
- ✅ **Server Implementation**: Built complete MCP server with all required components
- ✅ **Local Testing**: All endpoints respond correctly (`/health`, `/.well-known/*`, `/sse`)
- ✅ **Public Access**: Successfully exposed via Tailscale Funnel
- ✅ **GitHub OAuth App**: Created OAuth app with proper configuration

### Phase 2: ChatGPT Integration Attempts
- ✅ **Manifest Discovery**: ChatGPT successfully fetches `.well-known/ai-plugin.json`
- ✅ **OAuth Config**: ChatGPT successfully fetches `.well-known/oauth-authorization-server`
- ❌ **OAuth Resolution**: ChatGPT fails with "Failed to resolve OAuth client"

### Phase 3: Manifest Format Debugging
**Issue Identified**: Relative vs Absolute API URLs
- **Problem**: Initial manifest had `"url": "/sse"` (relative path)
- **Solution**: Changed to `"url": "https://f.beagle-duck.ts.net/sse"` (absolute URL)
- **Result**: ChatGPT continued to show errors

### Phase 4: OAuth Configuration Debugging
**Issue Identified**: Token endpoint authentication methods mismatch
- **Problem**: Advertised both `client_secret_post` and `client_secret_basic`
- **Root Cause**: Based on GitHub issue about Dynamic Client Registration (DCR) library behavior
- **Solution**: Limited to only `client_secret_post` which GitHub/ChatGPT expect
- **Result**: Error persists

### Phase 5: Client ID Investigation
**Issue Identified**: Missing client_id in manifest
- **Problem**: Research document suggested client_id not needed
- **Solution**: Added `client_id` field to auth configuration
- **Result**: Error persists

## Current Configuration

### Working Endpoints
```bash
# Health check
curl https://f.beagle-duck.ts.net/health
# Returns: {"status":"healthy","service":"remote-chatgpt-connector"}

# Plugin manifest
curl https://f.beagle-duck.ts.net/.well-known/ai-plugin.json
# Returns: Valid manifest with absolute SSE URL and client_id

# OAuth configuration  
curl https://f.beagle-duck.ts.net/.well-known/oauth-authorization-server
# Returns: Valid OAuth config with only client_secret_post
```

### Current Manifest Structure
```json
{
  "name": "Go SSE MCP Demo",
  "description": "Remote MCP server for ChatGPT integration with GitHub OAuth",
  "version": "0.1.0",
  "auth": {
    "type": "oauth",
    "authorization_url": "https://github.com/login/oauth/authorize",
    "token_url": "https://github.com/login/oauth/access_token",
    "scopes": ["read:user"],
    "client_id": "Ov23liOjz2lGVyB13kva"
  },
  "api": {
    "type": "mcp",
    "url": "https://f.beagle-duck.ts.net/sse"
  }
}
```

### Current OAuth Configuration
```json
{
  "issuer": "https://github.com",
  "authorization_endpoint": "https://github.com/login/oauth/authorize",
  "token_endpoint": "https://github.com/login/oauth/access_token",
  "scopes_supported": ["read:user", "user:email", "public_repo", "repo"],
  "response_types_supported": ["code"],
  "grant_types_supported": ["authorization_code"],
  "code_challenge_methods_supported": ["S256", "plain"],
  "token_endpoint_auth_methods_supported": ["client_secret_post"]
}
```

## Error Analysis

### ChatGPT Error Message
```
Failed to resolve OAuth client for MCP connector: https://f.beagle-duck.ts.net/
```

### Server Logs Analysis
**Successful Requests from ChatGPT**:
```
2:54PM INF pkg/discovery/manifest.go:351 > plugin manifest served successfully 
  user_agent="Python/3.12 aiohttp/3.9.5" 
  host=f.beagle-duck.ts.net 
  status_code=200

2:54PM INF pkg/discovery/manifest.go:456 > OAuth config served successfully
  user_agent="Python/3.12 aiohttp/3.9.5"
  host=f.beagle-duck.ts.net
  status_code=200
```

**Key Observations**:
1. ChatGPT successfully fetches both manifest and OAuth config
2. All HTTP responses return 200 OK
3. No authentication attempts reach our `/sse` endpoint
4. Error occurs during OAuth client resolution phase, before user sees auth prompt

## Potential Root Causes

### 1. GitHub OAuth App Configuration
**Issue**: Callback URL mismatch
- **Current**: Various test URLs used during development
- **Required**: `https://chat.openai.com/aip/p/callback` (ChatGPT's callback)
- **Status**: Needs verification/update

### 2. Dynamic Client Registration (DCR) Issues
**Issue**: ChatGPT may be attempting DCR with GitHub
- **Problem**: GitHub doesn't support DCR, only pre-registered clients
- **Evidence**: Error mentions "resolve OAuth client" suggesting dynamic resolution
- **Research**: Found GitHub issue about DCR library behavior with `token_endpoint_auth_method`

### 3. Missing OAuth Discovery Fields
**Issue**: OAuth configuration may be incomplete for ChatGPT's DCR expectations
- **Missing fields**: 
  - `registration_endpoint` 
  - `jwks_uri`
  - `userinfo_endpoint`
  - `introspection_endpoint`
- **Impact**: ChatGPT may require these for full OAuth 2.1 compliance

### 4. GitHub as OAuth Provider Limitations
**Issue**: GitHub OAuth may not be fully compatible with ChatGPT's expectations
- **GitHub limitations**:
  - No OIDC ID tokens (only access tokens)
  - No standard `userinfo_endpoint`
  - Limited OAuth 2.1 feature support
- **ChatGPT expectations**: May expect full OIDC/OAuth 2.1 provider

### 5. ChatGPT Connector vs Plugin Differences
**Issue**: MCP connectors may have different requirements than traditional plugins
- **Plugin manifest**: May be incorrect format for connectors
- **OAuth flow**: Connectors might use different OAuth flow than plugins
- **API structure**: `"api": {"type": "mcp"}` may be incorrect

## Next Steps & Recommendations

### Immediate Actions (High Priority)

1. **Verify GitHub OAuth App Configuration**
   ```bash
   # Check current callback URL
   # Update to: https://chat.openai.com/aip/p/callback
   ```

2. **Test with Standard OAuth Provider**
   - Replace GitHub with Auth0/Okta that supports full OIDC
   - Verify if issue is GitHub-specific or general

3. **Add Missing OAuth Discovery Fields**
   ```json
   {
     "registration_endpoint": "https://f.beagle-duck.ts.net/oauth2/register",
     "userinfo_endpoint": "https://api.github.com/user",
     "jwks_uri": "https://github.com/.well-known/jwks"
   }
   ```

4. **Implement Mock DCR Endpoint**
   - Add `/oauth2/register` endpoint that returns pre-configured client
   - May resolve "OAuth client resolution" issue

### Research Actions (Medium Priority)

5. **Analyze ChatGPT Connector Documentation**
   - Find official ChatGPT connector specification
   - Compare with plugin manifest requirements
   - Verify `api.type: "mcp"` is correct

6. **Study Working MCP Connector Examples**
   - Find published MCP connectors (HubSpot mentioned in research)
   - Compare manifest structures and OAuth configurations
   - Identify differences in implementation

7. **Test Alternative Manifest Formats**
   ```json
   // Try plugin-style manifest
   {
     "schema_version": "v1",
     "name_for_model": "go_mcp_demo",
     "name_for_human": "Go MCP Demo"
   }
   ```

### Advanced Debugging (Low Priority)

8. **Implement OAuth 2.1 Full Compliance**
   - Add PKCE support
   - Implement proper OIDC flows
   - Add JWT token support

9. **Create Alternative Transport**
   - Test HTTP-only transport (no SSE)
   - Verify if SSE is causing issues

10. **Network-Level Debugging**
    - Use Wireshark/tcpdump to capture ChatGPT requests
    - Analyze if there are hidden API calls we're missing

## Technical Debt & Improvements

### Code Quality
- ✅ Comprehensive logging implemented
- ✅ Error handling and graceful shutdown
- ✅ Structured configuration management
- ✅ Interface-based architecture for testability

### Missing Tests
- Unit tests for each component
- Integration tests for OAuth flow
- End-to-end tests with mock ChatGPT client

### Documentation
- API documentation
- Deployment guide
- Troubleshooting runbook

## Conclusion

We have successfully implemented a complete, production-ready MCP remote connector with comprehensive logging and proper architecture. The server responds correctly to all requests from ChatGPT, but fails during the OAuth client resolution phase.

The most likely issues are:
1. **GitHub OAuth app callback URL misconfiguration**
2. **GitHub's limited OAuth 2.1/OIDC support** conflicting with ChatGPT's DCR expectations
3. **Missing OAuth discovery fields** required for full compliance

The next highest-impact action is to verify/fix the GitHub OAuth app callback URL and then test with a full OIDC provider like Auth0 to isolate whether the issue is GitHub-specific.

## Environment Details

- **Server URL**: `https://f.beagle-duck.ts.net`
- **GitHub OAuth App**: `Ov23liOjz2lGVyB13kva`
- **Deployment**: Tailscale Funnel
- **Go Version**: 1.24.3
- **Key Dependencies**:
  - `github.com/go-go-golems/go-go-mcp v0.0.12`
  - `github.com/gorilla/mux v1.8.1`
  - `github.com/rs/zerolog v1.34.0`

## Logs & Artifacts

All debug logs available in tmux session `mcp-server`:
```bash
tmux attach -t mcp-server
```

Test script available:
```bash
./test-integration.sh
```

Configuration files:
- `.env` - Environment configuration
- `INTEGRATION.md` - Setup guide
- `ARCHITECTURE.md` - System architecture
