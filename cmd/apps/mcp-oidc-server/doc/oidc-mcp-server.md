# OIDC + MCP Server (Go) – Architecture, Implementation, and Ops Playbook

This document gives a new developer full context on the OIDC/MCP server implemented in this repo: what it does, how it’s structured, and exactly how to build, test, and operate it. It follows the tutorial/documentation style used in our Glazed docs (clear goals, structured sections, concrete examples).

## 1) What we’re building

- An OAuth 2.1 / OIDC Authorization Server (using Fosite) with:
  - Discovery: `/.well-known/openid-configuration` and `/.well-known/oauth-authorization-server`
  - JWKS: `/jwks.json`
  - Auth endpoints: `/oauth2/auth`, `/oauth2/token`
  - Minimal login UI: `/login` (demo credentials)
  - Dynamic Client Registration: `/register`
- An MCP server over HTTP at `/mcp` implementing JSON‑RPC with tools:
  - `search`: returns a list of matching items
  - `fetch`: returns a single item by `id`
- Bearer token protection on `/mcp` via Fosite introspection, plus a development fallback using tokens stored in SQLite.
- SQLite persistence for clients, tokens, and signing keys; optional table for MCP tool call logs.

## 2) Repository layout (relevant files)

- `go-go-labs/cmd/apps/mcp-oidc-server/main.go`
  - Cobra entrypoint; server flags and subcommands.
- `go-go-labs/cmd/apps/mcp-oidc-server/pkg/server/server.go`
  - HTTP mux, `/mcp` JSON‑RPC implementation, Bearer auth middleware.
- `go-go-labs/cmd/apps/mcp-oidc-server/pkg/idsrv/idsrv.go`
  - Fosite provider wiring, endpoints (discovery/JWKS/login/auth/token/register), SQLite persistence (clients, keys, tokens), helper APIs.

## 3) Identity Provider (OIDC/OAuth2) – key implementation details

File: `pkg/idsrv/idsrv.go`

- Provider composition (Fosite):
  - We generate an RSA private key at startup and pass it to `compose.ComposeAllEnabled`.
  - We set a 32‑byte `GlobalSecret` on `*fosite.Config` for HMAC‑signed authorization codes/refresh tokens.
  - PKCE is enforced for public clients.

- Discovery & JWKS:
  - `/.well-known/openid-configuration`
  - `/.well-known/oauth-authorization-server`
  - `/jwks.json`

- Login:
  - `/login` GET shows a simple form; POST sets a cookie and redirects back to `/oauth2/auth`.
  - Demo credentials: username `admin`, password `password123`.

- Authorization Code + PKCE flow:
  - `/oauth2/auth` issues an authorization code after successful login.
  - `/oauth2/token` exchanges code + code_verifier for `access_token` and (when `openid` scope is requested) `id_token`.
  - ID Token `aud` is set to the OAuth 2.0 `client_id` of the RP.

- Dynamic Client Registration:
  - `POST /register` with: `{ redirect_uris, token_endpoint_auth_method:"none", grant_types:["authorization_code","refresh_token"], response_types:["code"], client_id?(optional) }`.
  - We persist registrations in SQLite, and also support optional `client_id` to match clients that arrive with a preset id.

- Logging (identity):
  - Rich logs added for authorize/token including OAuth error unwrapping via `fosite.ErrorToRFC6749Error`.

References:
- Fosite compose docs: https://pkg.go.dev/github.com/ory/fosite/compose
- OIDC Discovery: https://openid.net/specs/openid-connect-discovery-1_0.html
- RFC 8414 (Authorization Server Metadata): https://datatracker.ietf.org/doc/html/rfc8414
- PKCE (RFC 7636): https://datatracker.ietf.org/doc/html/rfc7636

## 4) MCP endpoint – tools and transport

File: `pkg/server/server.go`

- Endpoint: `/mcp` (HTTP JSON‑RPC), with support for:
  - `initialize`
  - `notifications/initialized` (as sent by client; we log requests generically)
  - `tools/list`
  - `tools/call`

- Tools:
  - `search` (require_approval: "never")
    - Args: `{ query: string }`
    - Returns an array of items with `{ id, title, text, url }`.
  - `fetch` (require_approval: "never")
    - Args: `{ id: string }`
    - Returns a single item `{ id, title, text, url }`.
  - For now, these return data from a small in‑memory sample corpus (see `sampleDocs`). Replace with your dataset.

- Security:
  - Bearer tokens are validated using Fosite introspection.
  - As a development fallback, if introspection fails, we look up the token in SQLite (`oauth_tokens`) and accept it if not expired. Logs show when the dev fallback is used.
  - We advertise `WWW-Authenticate` metadata on 401 per RFC 9728.

- Logging (MCP):
  - We log the full JSON‑RPC request and response bodies at debug level, plus method/id and durations at info level.
  - Tools list includes `require_approval: "never"` per deep‑research client requirements.

References:
- MCP Transports (Streamable HTTP): https://modelcontextprotocol.io/specification/2025-03-26/basic/transports
- MCP Authorization: https://modelcontextprotocol.io/specification/2025-03-26/basic/authorization

## 5) Persistence (SQLite)

Enabled by passing `--db <path>` (or `DB=<path>`), e.g. `--db /tmp/mcp-oidc.db`.

Tables:

- `oauth_clients(client_id TEXT PRIMARY KEY, redirect_uris TEXT)`
  - Created by `/register` and loaded on startup.
- `oauth_keys(kid TEXT PRIMARY KEY, private_pem BLOB, created_at TIMESTAMP)`
  - We persist the RSA private key so JWKS and token signatures remain stable across restarts.
- `oauth_tokens(token TEXT PRIMARY KEY, subject TEXT, client_id TEXT, scopes TEXT, expires_at TIMESTAMP)`
  - Manual tokens for development and testing; also used by the MCP dev fallback.
- `mcp_tool_calls(id INTEGER PRIMARY KEY AUTOINCREMENT, ts TIMESTAMP, subject TEXT, client_id TEXT, request_id TEXT, tool_name TEXT, args_json TEXT, result_json TEXT, status TEXT, duration_ms INTEGER)`
  - Table is created; a helper API exists to insert logs (`idsrv.LogMCPCall`).
  - Wiring tool‑call persistence is straightforward in `server.go` (see “Future Work” below).

Helper functions (in `idsrv`):
- `InitSQLite(path string)` – sets up tables and loads clients/keys.
- `PersistToken`, `GetToken`, `ListTokens` – manage tokens.
- `LogMCPCall(entry MCPCallLog)` – insert tool call logs (call from MCP handler).

## 6) CLI – flags and subcommands

File: `cmd/apps/mcp-oidc-server/main.go`

Root flags:
- `--addr` (default `:8080`)
- `--issuer` (base URL; for local use `http://localhost:8080`, for ngrok use the public https URL)
- `--log-format` (`console|json`)
- `--log-level` (`trace|debug|info|warn|error`)
- `--db` (SQLite path; enables persistence)

Subcommands:
- `list-clients` (requires `--db`): prints registered clients from `oauth_clients`.
- `tokens list` (requires `--db`): lists tokens from `oauth_tokens`.
- `tokens create` (requires `--db`): creates a manual token.
  - Flags: `--token`, `--subject`, `--client-id`, `--scopes`, `--ttl` (e.g., `24h`).

Examples:

```bash
# Run server locally with persistence
LOG_FORMAT=console ISSUER=http://localhost:8080 \
  go run ./cmd/apps/mcp-oidc-server \
  --log-level debug --addr :8080 --issuer http://localhost:8080 --db /tmp/mcp-oidc.db

# Register a token manually (for dev MCP testing)
DB=/tmp/mcp-oidc.db ./mcp-oidc-server tokens create \
  --token $(openssl rand -hex 24) --subject admin --client-id dev-client --scopes openid,profile --ttl 24h

# List tokens
DB=/tmp/mcp-oidc.db ./mcp-oidc-server tokens list

# List clients
DB=/tmp/mcp-oidc.db ./mcp-oidc-server list-clients
```

## 7) End‑to‑end verification (playbook)

Local (dev client):

1) Discovery

```bash
curl -s http://localhost:8080/.well-known/openid-configuration | jq
curl -s http://localhost:8080/.well-known/oauth-authorization-server | jq
curl -s http://localhost:8080/jwks.json | jq
```

2) Authorization Code + PKCE (manual)

```bash
# Generate verifier/challenge
VER=$(python3 - <<'PY'
import os,base64,hashlib
v=base64.urlsafe_b64encode(os.urandom(32)).decode().rstrip('=')
print(v)
PY
)
CHAL=$(python3 - <<PY
import base64,hashlib,os
v=os.environ['VER'].encode()
print(base64.urlsafe_b64encode(hashlib.sha256(v).digest()).decode().rstrip('='))
PY
)

xdg-open "http://localhost:8080/oauth2/auth?response_type=code&client_id=dev-client&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fdev%2Fcallback&code_challenge_method=S256&code_challenge=${CHAL}&scope=openid&state=s123"

# After logging in (admin/password123), copy the code from /dev/callback
CODE=... # paste here

curl -s -X POST http://localhost:8080/oauth2/token \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -d 'grant_type=authorization_code' \
  -d 'client_id=dev-client' \
  -d 'redirect_uri=http://localhost:8080/dev/callback' \
  -d "code=${CODE}" \
  -d "code_verifier=${VER}" | jq
```

3) Call MCP

```bash
ACCESS=... # use access_token from previous step or DB manual token

curl -s http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS" \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}' | jq

curl -s http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS" \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq

curl -s http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS" \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search","arguments":{"query":"oidc"}}}' | jq

curl -s http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS" \
  -H "Content-Type: application/json" \
  --data '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"fetch","arguments":{"id":"1"}}}' | jq
```

4) ChatGPT connector (via public URL)

- Host the server behind an HTTPS origin (e.g., ngrok).
- Add your MCP URL in ChatGPT → Settings → Connectors → Add custom.
- ChatGPT will read `/.well-known/oauth-authorization-server`, may dynamically register via `/register`, then run the OAuth flow with PKCE.
- Watch logs for incoming `client_id`, `redirect_uri`, and code issuance.

## 8) Operations & management

- Flags and logging:
  - Use `--log-level debug` while integrating; logs cover authorize/token RFC errors and MCP request/response bodies.

- Persistence:
  - Always run with `--db` in real usage to persist clients and signing keys.
  - JWKS stability across restarts is ensured by `oauth_keys`.

- Tokens:
  - For testing, create manual tokens with `tokens create`.
  - Production tokens should come from `/oauth2/token`; consider removing the dev fallback for production deployments.

- Security hardening:
  - Add Origin checks on `/mcp` if exposing SSE/GET.
  - Consider audience enforcement on access tokens.
  - Protect `/register` (DCR) with an initial access token if needed.

## 9) Troubleshooting

- `invalid_client` at `/oauth2/auth`:
  - The `client_id` isn’t registered yet or `redirect_uri` doesn’t match exactly.
  - Fix by POSTing to `/register` with the exact `redirect_uris` used by the client.

- `server_error` at `/oauth2/auth`:
  - Fosite needs `Config.GlobalSecret` (32 bytes) set and RSA private key passed to `ComposeAllEnabled`.
  - Our implementation sets both; logs include RFC fields (`rfc_error`, `rfc_hint`, `rfc_description`).

- `/mcp` returns 401:
  - Missing/invalid Bearer. We advertise `WWW-Authenticate` with authorization metadata per RFC 9728.
  - For dev, create a token via CLI and retry.

## 10) Future work

- Persist MCP tool calls:
  - Table and `LogMCPCall` API exist. Add inserts in `server.go` tools/call handler (after computing results and status). This gives full audit trails for tool usage.
- Replace the sample corpus with your dataset and wire `search`/`fetch` to it.
- Enforce access token audience (`aud`) and scope checks on `/mcp`.

## 11) File/function reference

- `pkg/idsrv/idsrv.go`
  - `New(issuer string) (*Server, error)` – compose Fosite; configure secrets; set demo client.
  - `(*Server) Routes(mux *http.ServeMux)` – discovery, jwks, login, auth/token, register.
  - `(*Server) InitSQLite(path string) error` – create tables and load persisted state.
  - `(*Server) PersistToken/GetToken/ListTokens` – dev token management.
  - `(*Server) LogMCPCall(entry MCPCallLog)` – insert MCP tool call log entries.

- `pkg/server/server.go`
  - `New(issuer string) (*Server, error)` – construct app server and identity server.
  - `(*Server) EnableSQLite(path string) error` – enable persistence by calling `idsrv.InitSQLite`.
  - `(*Server) Routes(mux *http.ServeMux)` – HTTP mux with healthz, discovery, protected resource metadata, dev callback, and `/mcp`.
  - `mcpAuthMiddleware` – Bearer token handling (Fosite introspection + dev fallback).
  - `handleMCP` – JSON‑RPC server (initialize, tools/list, tools/call).

- `cmd/apps/mcp-oidc-server/main.go`
  - Cobra root flags; subcommands: `list-clients`, `tokens list`, `tokens create`.

## 12) References

- Fosite (ORY): https://github.com/ory/fosite
- MCP spec (2025‑03‑26): https://modelcontextprotocol.io/specification/2025-03-26/
- RFC 8414 (AS Metadata): https://datatracker.ietf.org/doc/html/rfc8414
- OIDC Discovery: https://openid.net/specs/openid-connect-discovery-1_0.html
- PKCE (RFC 7636): https://datatracker.ietf.org/doc/html/rfc7636


