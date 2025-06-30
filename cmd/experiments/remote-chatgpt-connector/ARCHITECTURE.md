# MCP Remote Connector Architecture

## Overview
A self-contained remote MCP (Model Context Protocol) server that implements its own OIDC (OpenID Connect) Authorization Server with dynamic client registration. This allows ChatGPT and other MCP clients to register dynamically and perform authenticated searches/fetches through a secure SSE (Server-Sent Events) transport.

## Component Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   ChatGPT       │────│  SSE Transport   │────│   MCP Server    │
│   Client        │    │  Layer           │    │   Core          │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                       ┌────────▼────────┐    ┌─────────▼─────────┐
                       │  OIDC Auth      │    │  Search/Fetch     │
                       │  Server         │    │  Handlers         │
                       │  (Fosite)       │    └───────────────────┘
                       └─────────────────┘             
                                │
                       ┌────────▼────────┐
                       │  Discovery      │
                       │  Service        │
                       │ (.well-known)   │
                       └─────────────────┘
```

## New Architecture: Self-Contained OIDC Server

After experiencing OAuth client resolution issues with GitHub OAuth (see [INVESTIGATION.md](INVESTIGATION.md)), we're implementing our own complete OIDC Authorization Server using the Ory Fosite library. This provides:

### Core OIDC Components

1. **Dynamic Client Registration** (`/register`)
   - RFC 7591 compliant endpoint
   - Allows ChatGPT to register itself at runtime
   - Returns `client_id` and optionally `client_secret`
   - Supports both public and confidential clients

2. **Authorization Server** (`/authorize`, `/token`)
   - OAuth 2.0 Authorization Code flow with PKCE
   - User authentication with hardcoded credentials (wesen/secret)
   - ID token issuance for OIDC compliance
   - Refresh token support for long-lived sessions

3. **Token Validation** 
   - Access token introspection for protected resources
   - Bearer token validation on all MCP endpoints
   - Scope-based authorization

### API Interfaces for Parallel Development

### 1. MCP Server Core (`pkg/mcp/`)
**Interface**: `MCPServer`
- Handles JSON-RPC message routing
- Manages search/fetch capability registration
- Token-protected resource endpoints

**Agent Focus**: Message handling, capability registration, business logic

### 2. OIDC Authorization Server (`pkg/auth/`)
**Interface**: `OIDCProvider`
- Dynamic client registration (`POST /register`)
- Authorization endpoint (`GET/POST /authorize`) 
- Token endpoint (`POST /token`)
- User authentication and session management
- Token introspection and validation

**Agent Focus**: Fosite integration, OAuth flows, JWT signing, user auth

### 3. Storage Layer (`pkg/storage/`)
**Interface**: `Storage`
- In-memory or SQLite-backed storage
- Client registration persistence
- Authorization codes, access tokens, refresh tokens
- User credentials and sessions
- PKCE challenge storage

**Agent Focus**: Fosite storage interface implementation, data persistence

### 4. SSE Transport (`pkg/transport/`)
**Interface**: `Transport`
- Server-Sent Events implementation
- HTTP request/response handling
- Bearer token extraction and validation
- Stream management

**Agent Focus**: HTTP transport, SSE protocol, connection management

### 5. Discovery Service (`pkg/discovery/`)
**Interface**: `DiscoveryService`
- `.well-known/ai-plugin.json` endpoint
- `.well-known/oauth-authorization-server` endpoint
- OIDC metadata generation

**Agent Focus**: Static endpoint serving, OIDC discovery metadata

## Data Flow

1. **Discovery**: ChatGPT fetches `.well-known/ai-plugin.json`
2. **Dynamic Registration**: ChatGPT calls `POST /register` to obtain client credentials
3. **Authorization**: User visits `/authorize`, logs in with wesen/secret, grants consent
4. **Token Exchange**: ChatGPT exchanges auth code for access/ID/refresh tokens at `/token`
5. **Authenticated Connection**: ChatGPT opens SSE connection with Bearer token
6. **Token Validation**: Server validates token on every MCP request
7. **MCP Operations**: Search/fetch operations executed and results streamed back

## Implementation Strategy

### Phase 1: Core OIDC Server
- Implement Fosite-based OIDC provider
- Dynamic client registration endpoint
- Authorization code flow with PKCE
- In-memory storage for rapid prototyping

### Phase 2: MCP Integration  
- Bearer token validation middleware
- Protected MCP endpoints
- SSE transport with authentication
- Search/fetch handlers

### Phase 3: Production Hardening
- SQLite persistent storage
- Proper JWT signing keys
- Rate limiting and security policies
- Comprehensive error handling

## Configuration

Environment variables:
- `OIDC_ISSUER` - Base URL of the OIDC issuer (default: http://localhost:8080)
- `OIDC_SIGNING_KEY` - RSA private key for JWT signing
- `DEFAULT_USER_USERNAME` - Hardcoded username (default: wesen)
- `DEFAULT_USER_PASSWORD` - Hardcoded password (default: secret)
- `PORT` - Server port (default: 8080)
- `LOG_LEVEL` - Logging level (default: info)
- `STORAGE_TYPE` - memory or sqlite (default: memory)

## Security Features

- **PKCE Required**: All authorization code flows must use PKCE
- **Secure Token Generation**: Cryptographically secure random tokens
- **Bearer Token Validation**: Every MCP request validates access token
- **Scope-Based Authorization**: Fine-grained permission control
- **Redirect URI Validation**: Prevents token leakage attacks
- **Token Introspection**: Real-time token validity checking
- **HTTPS Enforcement**: Required in production environments

## Compliance

- **OAuth 2.1 Draft**: Incorporates latest security recommendations
- **OpenID Connect Core**: Full OIDC compatibility for ID tokens
- **RFC 7591**: Dynamic Client Registration Protocol
- **MCP Specification**: Follows March 2025 MCP auth requirements
