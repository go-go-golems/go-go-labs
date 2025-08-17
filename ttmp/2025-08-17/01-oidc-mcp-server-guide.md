Below is a **full, end‑to‑end implementation guide** you can hand to a new intern. It walks through building a **single Go binary** that includes:

* an **OIDC/OAuth2 Authorization Server** (IdP) with a very simple login page and hard‑coded credentials,
* **discovery endpoints** (`/.well-known/openid-configuration` and `/.well-known/oauth-authorization-server`) and **JWKS**,
* the **MCP HTTP endpoint** implemented with **mcp-go**,
* **Bearer‑token protection** on the MCP endpoint,
* clear **verification steps**, **what to log**, and **what to check** at each stage.

Where helpful, I cite the relevant specs and library docs.

---

## 0) Big picture & terminology

**What we’re building**

```
+---------------- Self-contained Go binary ------------------+
|                                                            |
|  [OIDC/OAuth IdP]         [MCP Server over HTTP]           |
|  - /.well-known/...       - POST/GET /mcp  (JSON-RPC)      |
|  - /oauth2/auth           - Tools: search, fetch           |
|  - /oauth2/token          - Bearer auth middleware         |
|  - /jwks.json             - Origin check (streamable HTTP) |
|  - /login (hardcoded)                                      |
|                                                            |
+------------------------------------------------------------+
```

**Why these endpoints?**

* **MCP over Streamable HTTP** is the transport ChatGPT uses for remote connectors. It is one **single HTTP endpoint** that accepts POSTs (and optionally supports SSE via GET). ([Model Context Protocol][1])
* **OIDC/OAuth2** is used so ChatGPT can obtain an access token and call your MCP server. MCP’s **Authorization** page expects **OAuth 2.1 (subset)**, supports **PKCE**, and relies on **Authorization Server Metadata (RFC 8414)** for discovery. ([Model Context Protocol][2])
* We’ll also serve **OpenID Connect Discovery** (the traditional `/.well-known/openid-configuration`). Many clients consume both. ([OpenID Foundation][3])

---

## 1) Prerequisites

* Go ≥ 1.21
* A public HTTPS URL for your server (for real use). For local tests we’ll use `http://localhost:8080`.
* Dependencies:

  ```bash
  go get github.com/ory/fosite@latest
  go get github.com/mark3labs/mcp-go@latest
  # Optional if you want to validate JWTs with a stand-alone verifier:
  go get github.com/coreos/go-oidc/v3/oidc@latest
  ```

  **Fosite** is the Go OAuth2/OIDC engine we’ll embed. **mcp-go** is the Go SDK for MCP servers. **go-oidc** is a verifier helper (we can also introspect via Fosite directly). ([GitHub][4], [mcp-go.dev][5], [Go Packages][6])

---

## 2) Project layout

```
cmd/server/main.go
internal/idsrv/idsrv.go         # Fosite config + endpoints + login
internal/idsrv/metadata.go      # .well-known endpoints + JWKS
internal/mcp/mcp.go             # mcp-go server + tools
internal/httpx/middleware.go    # auth + origin check
internal/logx/logx.go           # minimal structured logging
```

This keeps the code readable for the intern while still being a single binary.

---

## 3) Implement the **Identity Provider** (OIDC/OAuth2) with **Fosite**

### 3.1 Wire up Fosite

* **Config**: require PKCE for public clients; set `IDTokenIssuer` to your external base URL.
* **Storage**: use in‑memory storage and **one public client** initially (dev). You can add more clients (including ChatGPT) later.
* **Strategies**: HMAC for code/refresh, RS256 for **ID Tokens**, and either HMAC (opaque access tokens + introspection) or JWT access tokens.

> Fosite provides “compose” helpers. **Important ordering**: register **OpenID Connect Explicit factory after the authorize‑code factory**. ([Go Packages][7])

**Code (idsrv.go):**

```go
package idsrv

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/storage"
	"github.com/ory/fosite/token/jwt"
)

type Server struct {
	Provider   fosite.OAuth2Provider
	PrivateKey *rsa.PrivateKey
	Issuer     string
	// for demo: single hardcoded login
	User string
	Pass string
}

func New(issuer string, devClientID string, devRedirectURI string, user, pass string) *Server {
	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)

	cfg := &fosite.Config{
		IDTokenIssuer:               issuer,
		EnforcePKCEForPublicClients: true,
		// keep defaults otherwise
	}

	mem := storage.NewMemoryStore()
	mem.Clients[devClientID] = &fosite.DefaultClient{
		ID:            devClientID,
		RedirectURIs:  []string{devRedirectURI},
		GrantTypes:    []string{"authorization_code", "refresh_token"},
		ResponseTypes: []string{"code"},
		Scopes:        []string{"openid", "profile", "offline_access"},
		Public:        true, // PKCE; no client_secret
	}

	secret := []byte("dev-hmac-secret-change-me")
	keyGetter := func(context.Context) (interface{}, error) { return privateKey, nil }

	strats := &compose.CommonStrategy{
		CoreStrategy:               compose.NewOAuth2HMACStrategy(cfg, secret, nil),
		OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(keyGetter, cfg),
		JWTStrategy: &jwt.RS256JWTStrategy{
			PrivateKey: privateKey,
		},
	}

	provider := compose.Compose(
		cfg,
		mem,
		strats,
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2PKCEFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		compose.OpenIDConnectExplicitFactory, // must come after code factory
	)

	return &Server{Provider: provider, PrivateKey: privateKey, Issuer: issuer, User: user, Pass: pass}
}
```

* Fosite implements RFC6749 and OIDC pieces; we’re enabling **Authorization Code + PKCE** and **OIDC** (ID tokens). ([GitHub][4], [Go Packages][7])

### 3.2 Minimal login UI and authorization/token endpoints

We’ll add:

* `GET /login` → renders a login form,
* `POST /login` → sets a cookie if `username`/`password` match,
* `GET /oauth2/auth` → standard authorization endpoint (redirects to login if needed, then issues a **code**),
* `POST /oauth2/token` → token endpoint (exchanges code for tokens),
* `GET /userinfo` (optional),
* cookie helper and session struct.

**Code (idsrv.go continued):**

```go
package idsrv

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"math/big"
	"net/http"
	"net/url"
	"strings"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/ory/fosite/token/jwt"
)

const cookieName = "sid"

var loginTpl = template.Must(template.New("login").Parse(`
<!doctype html><meta charset="utf-8"><title>Login</title>
<body style="font-family:sans-serif">
<h3>Sign in</h3>
<form method="post" action="/login">
  <input type="hidden" name="return_to" value="{{.ReturnTo}}">
  <div><label>User <input name="username" autofocus></label></div>
  <div><label>Pass <input type="password" name="password"></label></div>
  <button type="submit">Login</button>
</form>
</body>`))

func (s *Server) Routes(mux *http.ServeMux) {
	// Discovery + JWKS
	mux.HandleFunc("/.well-known/openid-configuration", s.oidcDiscovery)
	mux.HandleFunc("/.well-known/oauth-authorization-server", s.asMetadata)
	mux.HandleFunc("/jwks.json", s.jwks)

	// Login
	mux.HandleFunc("/login", s.login)

	// OAuth2 endpoints
	mux.HandleFunc("/oauth2/auth", s.authorize)
	mux.HandleFunc("/oauth2/token", s.token)

	// Optional userinfo
	mux.HandleFunc("/userinfo", s.userinfo)
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		_ = loginTpl.Execute(w, struct{ ReturnTo string }{r.URL.Query().Get("return_to")})
	case "POST":
		_ = r.ParseForm()
		u, p := r.Form.Get("username"), r.Form.Get("password")
		if u == s.User && p == s.Pass {
			http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "ok:" + u, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
			rt := r.Form.Get("return_to")
			if rt == "" { rt = "/" }
			http.Redirect(w, r, rt, http.StatusFound)
			return
		}
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func currentUser(r *http.Request) (string, bool) {
	c, err := r.Cookie(cookieName)
	if err != nil || !strings.HasPrefix(c.Value, "ok:") {
		return "", false
	}
	return strings.TrimPrefix(c.Value, "ok:"), true
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ar, err := s.Provider.NewAuthorizeRequest(ctx, r)
	if err != nil {
		s.Provider.WriteAuthorizeError(ctx, w, ar, err)
		return
	}
	// Require login
	user, ok := currentUser(r)
	if !ok {
		http.Redirect(w, r, "/login?return_to="+url.QueryEscape(r.URL.String()), http.StatusFound)
		return
	}
	now := jwt.ToTime(0).Time() // or time.Now()
	sess := &openid.DefaultSession{
		Subject:  user,
		Username: user,
		Claims: &jwt.IDTokenClaims{
			Subject:     user,
			Issuer:      s.Issuer,
			IssuedAt:    now,
			AuthTime:    now,
			RequestedAt: now,
		},
		Headers: &jwt.Headers{Extra: map[string]any{"kid": "1"}},
	}
	resp, err := s.Provider.NewAuthorizeResponse(ctx, ar, sess)
	if err != nil {
		s.Provider.WriteAuthorizeError(ctx, w, ar, err)
		return
	}
	s.Provider.WriteAuthorizeResponse(ctx, w, ar, resp)
}

func (s *Server) token(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess := new(openid.DefaultSession)
	accessReq, err := s.Provider.NewAccessRequest(ctx, r, sess)
	if err != nil {
		s.Provider.WriteAccessError(ctx, w, accessReq, err)
		return
	}
	resp, err := s.Provider.NewAccessResponse(ctx, accessReq)
	if err != nil {
		s.Provider.WriteAccessError(ctx, w, accessReq, err)
		return
	}
	s.Provider.WriteAccessResponse(ctx, w, accessReq, resp)
}
```

### 3.3 Discovery & JWKS

We’ll serve **both** OIDC Discovery and OAuth **Authorization Server Metadata**. The MCP spec **derives** the Authorization Server discovery URL from the MCP server base URL, and expects **RFC 8414** there. We provide both for broad compatibility. ([Model Context Protocol][2], [IETF Datatracker][8])

**Code (metadata.go):**

```go
package idsrv

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
)

func (s *Server) oidcDiscovery(w http.ResponseWriter, r *http.Request) {
	j := map[string]any{
		"issuer":                 s.Issuer,
		"authorization_endpoint": s.Issuer + "/oauth2/auth",
		"token_endpoint":         s.Issuer + "/oauth2/token",
		"jwks_uri":               s.Issuer + "/jwks.json",
		"scopes_supported":       []string{"openid", "profile", "offline_access"},
		"response_types_supported": []string{"code"},
		"grant_types_supported":    []string{"authorization_code", "refresh_token"},
		"id_token_signing_alg_values_supported": []string{"RS256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
	}
	writeJSON(w, j)
}

func (s *Server) asMetadata(w http.ResponseWriter, r *http.Request) {
	j := map[string]any{
		"issuer":                 s.Issuer,                     // RFC 8414
		"authorization_endpoint": s.Issuer + "/oauth2/auth",
		"token_endpoint":         s.Issuer + "/oauth2/token",
		"jwks_uri":               s.Issuer + "/jwks.json",
	}
	writeJSON(w, j)
}

func (s *Server) jwks(w http.ResponseWriter, r *http.Request) {
	pub := &s.PrivateKey.PublicKey
	jwks := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA", "alg": "RS256", "use": "sig", "kid": "1",
				"n": base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
				"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
			},
		},
	}
	writeJSON(w, jwks)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
```

* **OIDC Discovery** (`/.well-known/openid-configuration`) is the OIDC discovery document. ([OpenID Foundation][3])
* **Authorization Server Metadata** (`/.well-known/oauth-authorization-server`) is the OAuth discovery doc required by MCP Authorization. ([Model Context Protocol][2], [IETF Datatracker][8])

> **PKCE** is mandatory for public clients; we already enforced it via Fosite config (MCP also calls out PKCE). ([IETF Datatracker][9], [Model Context Protocol][2])

---

## 4) Implement the **MCP** endpoint with **mcp-go**

* Mount a **single path** (e.g., `/mcp`) that accepts POSTs (and optional GET for SSE).
* Add **two tools** expected by ChatGPT’s “Deep Research” shape: `search` and `fetch`. ([mcp-go.dev][5])
* Protect with **Bearer token** middleware; validate against our in‑process Fosite provider (introspection) or verify JWTs.

**Code (internal/mcp/mcp.go):**

```go
package mcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/pkg/mcp"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
)

type Server struct {
	mux      *http.ServeMux
	provider fosite.OAuth2Provider
}

func New(provider fosite.OAuth2Provider) *Server {
	s := &Server{mux: http.NewServeMux(), provider: provider}

	// Build mcp-go server
	srv := mcp.NewServer(
		mcp.WithServerInfo("go-mcp", "0.1.0"),
	)

	// Tool: search
	srv.AddTool(mcp.NewTool("search", "Search corpus and return IDs",
		mcp.WithSchema(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`),
		mcp.WithHandler(func(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
			q, _ := req.Args["query"].(string)
			ids := []string{} // TODO: implement your lookup
			return mcp.NewToolResponse().Text(strings.Join(ids, ",")), nil
		}),
	))

	// Tool: fetch
	srv.AddTool(mcp.NewTool("fetch", "Fetch a record by ID",
		mcp.WithSchema(`{"type":"object","properties":{"id":{"type":"string"}},"required":["id"]}`),
		mcp.WithHandler(func(ctx context.Context, req *mcp.ToolRequest) (*mcp.ToolResponse, error) {
			id, _ := req.Args["id"].(string)
			rec := map[string]any{"id": id, "status": "ok"} // TODO
			return mcp.NewToolResponse().JSON(rec), nil
		}),
	))

	// Protect /mcp with Bearer + introspection
	s.mux.Handle("/mcp", s.authMiddleware(srv))
	return s
}

func (s *Server) Handler() http.Handler { return s.mux }

// Bearer auth via Fosite introspection (works with opaque tokens)
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			http.Error(w, "missing bearer", http.StatusUnauthorized)
			return
		}
		raw := strings.TrimPrefix(authz, "Bearer ")
		// validate access token (not ID token) against our provider
		sess := new(openid.DefaultSession)
		_, err := s.provider.IntrospectToken(r.Context(), raw, fosite.AccessToken, sess, nil)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		// (Optional) Add Origin header checks per MCP transport guidance
		next.ServeHTTP(w, r)
	})
}
```

* mcp-go simplifies the MCP handshake (`initialize`, `tools/list`, `tools/call`). ([mcp-go.dev][5])
* MCP **Streamable HTTP** has **Origin** considerations; validate `Origin` as recommended. Add an origin-check in middleware if you expose GET/SSE. ([Model Context Protocol][1])

---

## 5) Wire it together (cmd/server/main.go)

```go
package main

import (
	"log"
	"net/http"
	"os"

	"yourmod/internal/idsrv"
	"yourmod/internal/mcp"
)

func main() {
	issuer := envDefault("ISSUER", "http://localhost:8080")
	// Development OAuth client for manual testing (e.g., an OAuth playground or a simple test page)
	devClientID := envDefault("DEV_CLIENT_ID", "dev-client")
	devRedirect := envDefault("DEV_REDIRECT_URI", issuer+"/dev/callback")

	id := idsrv.New(issuer, devClientID, devRedirect, "admin", "password123")

	mux := http.NewServeMux()
	id.Routes(mux)

	// mount MCP
	mcpSrv := mcp.New(id.Provider)
	mux.Handle("/mcp", mcpSrv.Handler())

	log.Println("listening on", issuer)
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func envDefault(k, v string) string {
	if s := os.Getenv(k); s != "" { return s }
	return v
}
```

---

## 6) Verification steps (do these as you build)

> Each step includes **what to run**, **what a good result looks like**, and **what to log**.

### 6.1 Discovery documents

* **Command**

  ```bash
  curl -s http://localhost:8080/.well-known/openid-configuration | jq
  curl -s http://localhost:8080/.well-known/oauth-authorization-server | jq
  curl -s http://localhost:8080/jwks.json | jq
  ```
* **Good**
  JSON with expected endpoint URLs; JWKS has your RSA public key.
* **Log**
  `GET openid-config`, `GET as-metadata`, `GET jwks` with 200 and response size.
* **Why**
  OIDC Discovery and **RFC 8414** metadata are how the client learns your endpoints. ([OpenID Foundation][3], [IETF Datatracker][8])

### 6.2 Authorization Code + PKCE (manual)

* Open in a browser:

  ```
  http://localhost:8080/oauth2/auth?
    response_type=code&
    client_id=dev-client&
    redirect_uri=http://localhost:8080/dev/callback&
    scope=openid%20profile&
    state=abc123&
    code_challenge=Qf4...&code_challenge_method=S256
  ```

  (Generate a code challenge with any PKCE generator; **PKCE is required**.) ([IETF Datatracker][9])
* You should see your **login form**; log in as `admin/password123`; you’ll be redirected to `/dev/callback?code=...&state=abc123`.
* Exchange the code:

  ```bash
  curl -s -X POST http://localhost:8080/oauth2/token \
    -H 'Content-Type: application/x-www-form-urlencoded' \
    -d 'grant_type=authorization_code&client_id=dev-client' \
    -d 'redirect_uri=http://localhost:8080/dev/callback' \
    -d 'code=REPLACE_ME' \
    -d 'code_verifier=REPLACE_VERIFIER' | jq
  ```
* **Good**
  JSON containing `access_token`, `id_token`, `token_type`, `expires_in`, and optionally `refresh_token`.
* **Log**

  * At `/oauth2/auth`: `client_id`, `redirect_uri`, `scopes`, `code_challenge_method`, `state`.
  * At `/oauth2/token`: `grant_type`, `client_id`, and a result like `issued_access_token=true, issued_id_token=true`.

### 6.3 Call **/mcp** with Bearer token

* Use the `access_token` from step 6.2:

  ```bash
  curl -s http://localhost:8080/mcp \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Accept: application/json" \
    -H "Content-Type: application/json" \
    --data '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}' | jq
  ```
* **Good**
  An `InitializeResult` JSON object (mcp-go will handle this), followed by successful `tools/list` and `tools/call` when you try them. For Streamable HTTP rules, see MCP transports. ([Model Context Protocol][1])
* **Log**
  `mcp initialize` with caller’s `sub` (if you propagate it from introspection), request id, and success.

---

## 7) What to **log** (and sample fields)

1. **/oauth2/auth (authorize)**

   * `ts`, `client_id`, `redirect_uri`, `scopes`, `state`, `code_challenge_method`, **result** (issued code / error).
2. **/oauth2/token**

   * `ts`, `grant_type`, `client_id`, `scopes`, **result** (issued access/id/refresh).
3. **/mcp**

   * `ts`, `subject` (from introspection), `jsonrpc_method` (e.g., `initialize`, `tools/list`, `tools/call`), `tool_name`, `duration_ms`, `status`.
4. **Discovery & JWKS**

   * `ts`, `endpoint`, `status`, `bytes`.

This helps diagnose the usual issues: redirect mismatches, missing PKCE, token exchange errors, or an MCP call missing a Bearer token.

---

## 8) Security & correctness checks (concrete)

* **PKCE present**: log and reject when `code_challenge` or `code_verifier` is missing. (PKCE is required for public clients.) ([IETF Datatracker][9])
* **Redirect URI** matches the registered client’s allowed list (Fosite does this). Log mismatches with the exact `redirect_uri`. (This is standard OAuth behavior.) ([Connect2id][10])
* **Origin validation** at `/mcp` (and SSE GET if you add it): verify `Origin` per MCP transport “Security Warning”. ([Model Context Protocol][1])
* **Token validation** on `/mcp`: use **Fosite introspection** (works with opaque tokens), or if you switch to **JWT access tokens**, validate signature & claims (iss, aud, exp). Fosite and go‑oidc docs show both patterns. ([Go Packages][11])
* **Discovery availability**: both **/.well-known/openid-configuration** and **/.well-known/oauth-authorization-server** resolve and point to your endpoints (MCP Authorization relies on RFC 8414). ([Model Context Protocol][2], [IETF Datatracker][8])

---

## 9) Adding it as a **ChatGPT connector**

* In ChatGPT, go to **Settings → Connectors → Add custom connector**, paste your **/mcp** URL. ChatGPT will discover your authorization endpoints as described in the MCP spec (base URL → `/.well-known/oauth-authorization-server`), then perform OAuth with PKCE. If your client isn’t pre‑registered, add it (or implement Dynamic Client Registration; see next section). ([Model Context Protocol][2], [OpenAI Help Center][12])
* Watch your logs at `/oauth2/auth`: you’ll see the **incoming `client_id` and `redirect_uri`** ChatGPT uses. If you went with **manual client registration**, add this **exact** redirect URI to your allowed list for that client in your server. (Redirect URIs must match *exactly* in OAuth.) ([IETF Datatracker][8], [Microsoft Learn][13])

---

## 10) (Optional, recommended) **Dynamic Client Registration** (RFC 7591)

If you don’t want to pre‑register ChatGPT, add a minimal **`POST /register`** endpoint that accepts:

```json
{
  "redirect_uris": ["https://..."],
  "token_endpoint_auth_method": "none",
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"]
}
```

…and returns:

```json
{
  "client_id": "generated-id-123",
  "redirect_uris": ["..."],
  "token_endpoint_auth_method": "none"
}
```

MCP Authorization **encourages** DCR to remove manual steps. You can store registrations in memory or a tiny DB. ([Model Context Protocol][2], [IETF Datatracker][14])

---

## 11) Troubleshooting matrix (very concrete)

| Symptom                                 | Where to look              | What to check                                                                                                        |
| --------------------------------------- | -------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `invalid_request` at `/oauth2/auth`     | server logs                | Missing `client_id`, `redirect_uri`, `response_type`, `code_challenge`; verify PKCE present. ([IETF Datatracker][9]) |
| Browser loops back to login             | `/login` -> `/oauth2/auth` | Cookie not set; wrong credentials; check cookie domain/path.                                                         |
| `/oauth2/token` returns `invalid_grant` | token logs                 | Mismatched `code_verifier`; state mismatch; code already used or expired.                                            |
| `/mcp` returns 401                      | MCP logs                   | Missing/expired Bearer; token audience/scope; try introspecting token with Fosite. ([Go Packages][11])               |
| ChatGPT can’t connect                   | as‑metadata logs           | Does `/.well-known/oauth-authorization-server` return valid JSON? Endpoints reachable? ([IETF Datatracker][8])       |
| Redirect mismatch error                 | authorize logs             | Your client’s allowed `redirect_uris` doesn’t include the URL ChatGPT sent; add *exact* URI. ([Microsoft Learn][13]) |

---

## 12) Minimal **sequence** you can hand‑check

1. `GET /.well-known/openid-configuration` → 200 JSON. ([OpenID Foundation][3])
2. `GET /.well-known/oauth-authorization-server` → 200 JSON. ([IETF Datatracker][8])
3. `GET /jwks.json` → contains RS256 public key(s).
4. Authorization Code + PKCE (login → code → token). ([IETF Datatracker][9])
5. `POST /mcp` with `Authorization: Bearer <access_token>` → `InitializeResult`. ([Model Context Protocol][1])

---

## 13) References you’ll consult while coding

* **MCP Transports (Streamable HTTP)** — endpoint rules, POST/GET, SSE, **Origin** warning. ([Model Context Protocol][1])
* **MCP Authorization (2025‑03‑26)** — OAuth 2.1 subset, metadata discovery at `/.well-known/oauth-authorization-server`, PKCE required. ([Model Context Protocol][2])
* **mcp-go “Getting Started”** — how to register tools and run an MCP server in Go. ([mcp-go.dev][5])
* **RFC 8414** — Authorization Server Metadata (OAuth discovery). ([IETF Datatracker][8])
* **OIDC Discovery 1.0** — `/.well-known/openid-configuration`. ([OpenID Foundation][3])
* **PKCE (RFC 7636)** — required for public clients. ([IETF Datatracker][9])
* **Fosite** repo & example server — compose factories, example authorize/token flow. ([GitHub][4])
* **go-oidc** (optional) — JWT/ID token verifier. ([Go Packages][6])

---

## 14) Next steps you can assign to the intern

1. **Add Origin checks** to `/mcp`:

   ```go
   allowed := "https://chat.openai.com" // example; set to the real origin your client uses
   if o := r.Header.Get("Origin"); o != "" && o != allowed {
     http.Error(w, "bad origin", http.StatusForbidden); return
   }
   ```

   (MCP transport calls out Origin validation explicitly.) ([Model Context Protocol][1])
2. **Add a tiny “/dev/callback”** page that prints the `code` you received (for manual tests).
3. **(Optional)** Implement **/register** for **Dynamic Client Registration**, storing `client_id`, `redirect_uris`, and `token_endpoint_auth_method` in memory. This lets ChatGPT self‑register. ([IETF Datatracker][14])
4. **(Optional)** Switch to **JWT access tokens** if you want to validate tokens without introspection calls. (Fosite supports JWT strategies; see community Q\&A and docs.) ([Stack Overflow][15])

---

### Appendix: A quick **curl** to exercise MCP after you get a token

```bash
# Initialize
curl -s http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  --data '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26"}}' | jq

# List tools
curl -s http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  --data '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | jq

# Call search
curl -s http://localhost:8080/mcp \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "Accept: application/json" \
  --data '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search","arguments":{"query":"hello"}}}' | jq
```

The **Streamable HTTP** rules specify how your endpoint should respond to POSTs (JSON or SSE). For most servers, returning JSON is fine. ([Model Context Protocol][1])

---

If you want, I can collapse the code above into a single `main.go` and a tiny `go.mod` so the intern can `go run ./cmd/server` and follow the exact verification commands.

[1]: https://modelcontextprotocol.io/specification/2025-03-26/basic/transports "Transports - Model Context Protocol"
[2]: https://modelcontextprotocol.io/specification/2025-03-26/basic/authorization "Authorization - Model Context Protocol"
[3]: https://openid.net/specs/openid-connect-discovery-1_0.html?utm_source=chatgpt.com "OpenID Connect Discovery 1.0 incorporating errata set 2"
[4]: https://github.com/ory/fosite?utm_source=chatgpt.com "ory/fosite: Extensible security first OAuth 2.0 and OpenID ..."
[5]: https://mcp-go.dev/getting-started/?utm_source=chatgpt.com "Getting Started - MCP-Go"
[6]: https://pkg.go.dev/github.com/coreos/go-oidc?utm_source=chatgpt.com "oidc package - github.com/coreos ..."
[7]: https://pkg.go.dev/github.com/ory/fosite/compose?utm_source=chatgpt.com "compose package - github.com/ory/fosite ..."
[8]: https://datatracker.ietf.org/doc/html/rfc8414?utm_source=chatgpt.com "RFC 8414 - OAuth 2.0 Authorization Server Metadata"
[9]: https://datatracker.ietf.org/doc/html/rfc7636?utm_source=chatgpt.com "RFC 7636 - Proof Key for Code Exchange by OAuth Public ..."
[10]: https://connect2id.com/products/nimbus-oauth-openid-connect-sdk/examples/oauth/client-registration?utm_source=chatgpt.com "Client registration · Docs - OAuth 2.0"
[11]: https://pkg.go.dev/github.com/ory/fosite?utm_source=chatgpt.com "fosite package - github.com/ory/fosite - ..."
[12]: https://help.openai.com/en/articles/11487775-connectors-in-chatgpt?utm_source=chatgpt.com "Connectors in ChatGPT"
[13]: https://learn.microsoft.com/en-us/entra/identity-platform/reply-url?utm_source=chatgpt.com "Redirect URI (reply URL) best practices and limitations"
[14]: https://datatracker.ietf.org/doc/html/rfc7591?utm_source=chatgpt.com "RFC 7591 - OAuth 2.0 Dynamic Client Registration Protocol"
[15]: https://stackoverflow.com/questions/66290396/how-to-create-jwt-access-token-in-fosite-oauth2?utm_source=chatgpt.com "How to create JWT access token in Fosite OAuth2?"

