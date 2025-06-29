# MCP Remote Connector Architecture

## Overview
A remote MCP (Model Context Protocol) server that allows ChatGPT to connect via OAuth and perform searches/fetches through a secure SSE (Server-Sent Events) transport.

## Component Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ChatGPT       │────│  SSE Transport   │────│   MCP Server    │
│   Client        │    │  Layer           │    │   Core          │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                       ┌────────▼────────┐    ┌─────────▼─────────┐
                       │  Auth Validator │    │  Search/Fetch     │
                       │  (GitHub OAuth) │    │  Handlers         │
                       └─────────────────┘    └───────────────────┘
                                │
                       ┌────────▼────────┐
                       │  Discovery      │
                       │  Service        │
                       │ (.well-known)   │
                       └─────────────────┘
```

## API Interfaces for Parallel Development

### 1. MCP Server Core (`pkg/mcp/`)
**Interface**: `MCPServer`
- Handles JSON-RPC message routing
- Manages search/fetch capability registration
- Independent of transport layer

**Agent Focus**: Message handling, capability registration, business logic

### 2. Authentication (`pkg/auth/`)
**Interface**: `AuthValidator`
- GitHub OAuth token validation
- User verification against allowlist
- Token introspection

**Agent Focus**: OAuth flows, token validation, GitHub API integration

### 3. SSE Transport (`pkg/transport/`)
**Interface**: `Transport`
- Server-Sent Events implementation
- HTTP request/response handling
- Stream management

**Agent Focus**: HTTP transport, SSE protocol, connection management

### 4. Discovery Service (`pkg/discovery/`)
**Interface**: `DiscoveryService`
- `.well-known/ai-plugin.json` endpoint
- `.well-known/oauth-authorization-server` endpoint
- Manifest generation

**Agent Focus**: Static endpoint serving, JSON manifest generation

## Parallel Development Strategy

Each component can be developed independently by implementing its interface:

1. **Mock implementations** for testing (all interfaces mockable)
2. **Interface contracts** define exact behavior
3. **Dependency injection** allows easy composition
4. **Unit testing** per component without dependencies

## Data Flow

1. **Discovery**: ChatGPT fetches `.well-known/ai-plugin.json`
2. **OAuth**: User authorizes via GitHub OAuth flow
3. **Connect**: ChatGPT opens SSE connection with Bearer token
4. **Auth**: Token validated against GitHub API
5. **Stream**: JSON-RPC messages flow over SSE
6. **Search/Fetch**: MCP operations executed and results streamed back

## Configuration

Environment variables:
- `GITHUB_CLIENT_ID` - OAuth app client ID
- `GITHUB_CLIENT_SECRET` - OAuth app client secret  
- `GITHUB_ALLOWED_LOGIN` - GitHub username to allow
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level (default: info)

## Security

- Bearer token validation on every request
- GitHub API verification (not JWT - GitHub uses opaque tokens)
- Single user allowlist (personal connector)
- HTTPS required in production
