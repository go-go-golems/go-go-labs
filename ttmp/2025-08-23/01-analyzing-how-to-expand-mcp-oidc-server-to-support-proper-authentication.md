## Expanding mcp-oidc-server to support a proper authentication backend

This document analyzes the current `@mcp-oidc-server/` implementation, evaluates robust authentication backends we can integrate, and proposes a concrete refactor/implementation plan. The goals are to replace the hardcoded login with a secure, extensible identity backend while preserving the existing OAuth2/OIDC server behavior used by MCP clients.

### 1) Current setup: what we have today

- **Entrypoint**: `go-go-labs/cmd/apps/mcp-oidc-server/main.go`
  - Boots the server, sets flags (`--addr`, `--issuer`, `--db`, log options), mounts routes via `pkg/server`.
  - Optional SQLite persistence for clients, keys, tokens.
- **Identity provider (IdP) and OAuth2/OIDC server**: `pkg/idsrv/idsrv.go`
  - Uses ORY Fosite (`compose.ComposeAllEnabled`) as a local Authorization Server with PKCE, ID Token RS256 signing, and a default in‑memory client (`dev-client`).
  - Provides OIDC discovery, AS metadata, JWKS, `/oauth2/auth`, `/oauth2/token`, `/register`.
  - Also provides a minimal `/login` with a hardcoded username/password and a cookie session check.
- **Protected resource (MCP)**: `pkg/server/server.go`
  - `/mcp` endpoint is protected by a Bearer token middleware that first attempts Fosite introspection. If introspection fails, it has a development fallback to check opaque tokens persisted in SQLite.

Key hardcoded login fragment (for context):
```159:176:go-go-labs/cmd/apps/mcp-oidc-server/pkg/idsrv/idsrv.go
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        _ = loginTpl.Execute(w, struct{ ReturnTo string }{r.URL.Query().Get("return_to")})
    case http.MethodPost:
        _ = r.ParseForm()
        u := r.FormValue("username")
        p := r.FormValue("password")
        if u == s.User && p == s.Pass {
            http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "ok:"+u, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
            rt := r.FormValue("return_to")
            if rt == "" { rt = "/" }
            http.Redirect(w, r, rt, http.StatusFound)
            return
        }
        http.Error(w, "invalid credentials", http.StatusUnauthorized)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}
```

Bearer-protected MCP flow (introspection > dev fallback):
```188:205:go-go-labs/cmd/apps/mcp-oidc-server/pkg/server/server.go
sess := new(openid.DefaultSession)
// Introspect opaque access token
tt, ar, err := s.ids.ProviderRef().IntrospectToken(r.Context(), raw, fosite.AccessToken, sess)
if err != nil {
    // Dev fallback: accept manual tokens stored in DB
    if tr, ok, derr := s.ids.GetToken(raw); derr == nil && ok {
        if time.Now().Before(tr.ExpiresAt) {
            ctx := setAuthCtx(r.Context(), tr.Subject, tr.ClientID)
            next.ServeHTTP(w, r.WithContext(ctx))
            return
        }
    }
    http.Error(w, "invalid token", http.StatusUnauthorized)
    return
}
```

### 2) Problems and risks with the current login

- **Hardcoded credentials**
  - What it is: The username/password are embedded in binary (`idsrv.Server{User, Pass}`), compared directly on `/login` POST.
  - Risks:
    - Secrets live in code and can leak via VCS, logs, crash dumps, or reverse‑engineering the binary.
    - No password hashing or KDF (bcrypt/argon2id) → if leaked, immediately usable; cannot enforce password policies.
    - No account lifecycle: no rotation, expiry, lockout, or MFA; vulnerable to credential stuffing and brute force.
    - Operationally brittle: rotating requires a code change/redeploy; no per‑environment segregation.
    - Compliance: fails basic security expectations (OWASP ASVS), audit trails are absent.
  - Short‑term mitigations (if we must keep it temporarily):
    - Gate behind `--dev-login` and default it off in non‑dev; add rate limiting and minimal lockout.
    - Hide credentials from logs; ensure binaries are not distributed.
    - Place in an isolated dev environment with no external exposure.

- **Minimal session cookie**
  - What it is: Cookie value `ok:<username>` with no expiry (session cookie), no server‑side session store, SameSite=Lax.
  - Risks:
    - Session fixation and replay: cookie bears identity directly; cannot invalidate or rotate on privilege changes.
    - Missing expiry/idle timeout: sessions can live arbitrarily long in a browser; no logout invalidation on server.
    - Transport protection: `Secure` not guaranteed (depends on deployment); Lax may be too permissive for some flows.
    - CSRF exposure: form POST to `/login` lacks CSRF token; cookie‑based auth later can be targeted by CSRF in other flows.
    - Audit gaps: no server‑side session model → no way to list/kill sessions.
  - Mitigations:
    - Move to server‑side session table with opaque `session_id` + TTL/idle timeout + rotation on login.
    - Set `Secure`, `HttpOnly`, and tuned `SameSite`; add CSRF tokens to form POSTs.
    - Add explicit `/logout` to revoke server session.

- **No identity lifecycle**
  - What it is: No registration, verification, password reset, TOTP/U2F, or deprovisioning; no external IdP support.
  - Risks:
    - Users cannot be onboarded/deactivated cleanly; manual ad‑hoc changes risk drift.
    - No password recovery → risky workarounds in prod; no verification → account takeovers.
    - No MFA → single‑factor compromise leads to total account compromise.
    - No profile/claims management → weak linkage between business identities and tokens.
  - Mitigations:
    - Integrate a proper IdP (Model A or B) to offload lifecycle and MFA.
    - If staying local, at minimum add password reset workflows, email verification, and optional TOTP.

- **Production footguns (dev token fallback)**
  - What it is: When introspection fails in `/mcp`, code checks `oauth_tokens` table and authorizes if a matching unexpired manual token exists.
  - Risks:
    - Bypass of AS policy: tokens not subject to the same issuance controls, scopes, or revocation lists.
    - Drift: operators may create long‑lived tokens and forget to remove them; weak audit trails.
    - Over‑permissive defaults: easy to forget disabling in prod, silently weakening security.
  - Mitigations:
    - Add `--disable-dev-token-fallback` defaulting to true in non‑dev; log a prominent warning when enabled.
    - Scope manual tokens minimally and expire aggressively; log usage and surface metrics/alerts.

### 3) Target architecture overview (keep AS; replace login with proper IdP)

We keep our local Authorization Server roles (discovery, JWKS, `/oauth2/auth`, `/oauth2/token`, `/register`) so MCP clients continue to integrate exactly the same way. We replace the hardcoded `/login` with a proper authentication backend. Two viable models:

- **Model A – Upstream OIDC federation (recommended)**
  - Replace `/login` with a redirect to an upstream OIDC provider (managed or self‑hosted). On callback, verify ID Token using `go-oidc`, create a session in SQLite, and proceed with authorization code issuance via Fosite.
  - Applicable providers: **Auth0**, **Okta**, **Google**, **ZITADEL**, **Keycloak**, **Dex**, corporate IdPs.
  - Pros: minimal surface area to secure (we delegate identity auth), easy to swap providers, no password handling locally.

- **Model B – Self‑service identity via Ory Kratos (alternative)**
  - Keep Fosite locally for OAuth2/OIDC; integrate **Ory Kratos** for user auth flows. `/login` becomes a Kratos browser flow; we verify the Kratos session and map it to our subject.
  - Pros: fully self‑hosted, rich identity features (passwordless, MFA); Cons: more moving parts than Model A.

- **Model C – Minimal local user management (first step)**
  - Keep Fosite for OAuth2/OIDC, and implement a small local user DB with strong password hashing (argon2id), server‑side sessions, CSRF, and admin CLI to add/disable users.
  - Pros: lowest complexity to remove hardcoded login; no external dependencies; clear upgrade path later to Model A or B.

A full migration to **Ory Hydra** (replacing our in‑process Fosite composition) is possible but is a larger topology change (separate AS with login/consent app). We can consider it a future step if we outgrow the embedded AS.

#### Model B moving parts (what you need to run and wire)

- **Kratos service**
  - Binary/container with `kratos.yml` config; secrets for cookies/CSRF; runs alongside our app.
  - **Database**: Postgres recommended (SQLite only for dev). Stores identities, credentials, sessions, recovery tokens.
  - **Public and Admin APIs**: Public for browser/self‑service; Admin for server‑to‑server session/identity checks.
  - **Courier** (optional but typical): SMTP for verification and recovery emails.
- **Self‑service flows**
  - Browser flows for login, registration, verification, recovery, settings. We can:
    - Implement our own minimal UI hitting Kratos Public API, or
    - Use Kratos’ default UI templates, or
    - Proxy through our app to keep a single origin.
- **Session/cookies**
  - Kratos issues its own session cookie; our app either trusts Kratos’ session cookie (verify via Admin API) or creates a short‑lived local session linked to a Kratos session ID.
- **Integration points in our app**
  - Replace `/login` with redirect to Kratos self‑service login flow.
  - In `/oauth2/auth`, verify an active Kratos session (Admin API) and map identity → `subject` for Fosite’s `DefaultSession`.
  - Optional `/logout`: call Kratos to revoke and then clear our local session.
- **Security & ops**
  - TLS termination, cookie domain/path alignment, CSRF tokens, CORS for the self‑service UI.
  - Secrets management for Kratos, SMTP, DB creds; backups/migrations for the Kratos DB.
- **Optional social sign‑in**
  - Kratos supports OIDC social providers (Google/GitHub/etc.) via configuration; Kratos handles the OIDC dance, we just consume the session.

- **Model C – Minimal local user management (first step)**

If we prefer a lightweight, local solution without bringing Kratos, we can assemble a small, auditable stack using established Go packages. This will be our first implementation step to remove hardcoded login:

- **Password hashing**: `github.com/alexedwards/argon2id` (argon2id with sane defaults) or `golang.org/x/crypto/bcrypt`.
- **Sessions**: `github.com/alexedwards/scs/v2` (server‑side session store; supports many backends).
- **CSRF**: `github.com/justinas/nosurf` for form POSTs.
- **DB**: keep `database/sql` + SQLite (we already ship SQLite) with a `users` table.
- Optional: **Password strength** `github.com/nbutton23/zxcvbn-go`, **email** `github.com/jordan-wright/email` if you add resets later.

Minimal schema proposal:

```sql
CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  username      TEXT UNIQUE NOT NULL,
  email         TEXT,
  password_hash TEXT NOT NULL,
  disabled      INTEGER NOT NULL DEFAULT 0,
  created_at    TIMESTAMP NOT NULL
);
```

Minimal flow (pseudocode):

```go
// Registration (admin CLI only initially)
hash, _ := argon2id.CreateHash(plaintext, argon2id.DefaultParams)
INSERT INTO users(username, email, password_hash, created_at) VALUES(?, ?, ?, NOW)

// Login handler
u := r.FormValue("username"); p := r.FormValue("password")
row := SELECT password_hash, disabled FROM users WHERE username=?
if disabled { unauthorized }
if !argon2id.CompareHashAndPassword(row.password_hash, p) { unauthorized }
// create server-side session with scs
authSession.Put(ctx, "subject", u)

// Authorize endpoint (/oauth2/auth)
sub, ok := authSession.GetString(ctx, "subject"); if !ok { redirect /login }
issue code as today with Subject=sub
```

Admin operations (using our existing Cobra CLI):
- `mcp-oidc-server users add --username <u> --email <e> --password <p>`
- `mcp-oidc-server users disable --username <u>`
- `mcp-oidc-server users set-password --username <u> --password <p>`
- `mcp-oidc-server users list`

Tradeoffs:
- Pros: very small footprint; easy to reason about; no external dependency.
- Cons: no MFA, no email flows, no SSO; you own password security and UX.

#### Do we need Model A for Google/GitHub SSO?

- **Not strictly.** Model A (upstream OIDC) is the most direct path to add Google/GitHub SSO (use `go-oidc` against Google; or use **Dex** as an OIDC broker to GitHub OAuth).
- **Model B (Kratos)** can also do social sign‑in via its OIDC providers configuration; Kratos handles the SSO and issues a Kratos session we consume.
- Recommendation: If SSO is the immediate goal with minimal moving parts, use **Model A**. If you also want self‑service password logins, recovery, MFA, and identity management, **Model B** is a better long‑term platform.

### 4) Provider/library evaluation (Go-centric)

- **go-oidc (CoreOS/Dex)**: idiomatic OIDC RP library for verifying ID Tokens and doing discovery.
  - Docs: [pkg.go.dev github.com/coreos/go-oidc/v3/oidc](https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc)
  - Use with `golang.org/x/oauth2` for code exchange.
- **Ory Kratos**: identity system with self‑service auth flows.
  - Docs: [ory.sh/kratos/docs](https://www.ory.sh/kratos/docs/)
- **Dex**: lightweight OIDC broker with many connectors (GitHub, LDAP, SAML).
  - Docs: [dexidp.io/docs](https://dexidp.io/docs/)
- **ZITADEL**: modern cloud/self‑hosted IdP with orgs/projects and strong OIDC support.
  - Docs: [docs.zitadel.com](https://docs.zitadel.com/)
- **Keycloak**: full‑featured self‑hosted IdP.
  - Docs: [keycloak.org/docs/latest](https://www.keycloak.org/docs/latest/)
- **Auth0/Okta/Google**: managed IdPs with excellent OIDC support and examples.
  - Auth0 Go Web Quickstart: [auth0.com/docs/quickstart/webapp/golang](https://auth0.com/docs/quickstart/webapp/golang)

Recommendation: Start with **Model A** using `go-oidc` to federate to an upstream provider (pluggable via flags/env). It’s the smallest, safest change with good ergonomics.

### 5) Detailed implementation and refactor plan

#### 5.0 Implementation track: Model C (first step)

We will implement Model C first to remove hardcoded login safely and quickly. Summary of concrete steps:

- **Config (flags/env):**
  - `--local-users` (bool; default true): enable DB‑backed local users and session login flow.
  - `--session-ttl` (duration; default `12h`): lifetime of server‑side session.
- **Database:**
  - Create `users` table (see section 3 – Model C schema).
  - Reuse `user_sessions` table defined in 5.6 for session tracking across all models.
- **Handlers and wiring:**
  - Update `/login` POST to verify against `users` with argon2id; on success, create server session (`user_sessions`) and cookie.
  - Update `authorize` to require an active server session (subject from session), then issue code as today.
  - Add `/logout` to revoke `user_sessions` and clear cookie.
  - Keep existing `/oauth2/token` and `/register` unchanged.
- **CLI:**
  - Add `users add|disable|set-password|list` subcommands (see examples in section 3).
- **Security:**
  - Enforce `Secure` cookies where applicable, idle timeout, and session rotation on login.
  - Add CSRF protection to `/login` form posts.

The existing sections 5.1–5.7 describe the upstream OIDC (Model A) track and shared pieces; we will implement them later if/when we add SSO.

#### 5.1 Configuration additions (flags/env)

- `--upstream-issuer` / `UPSTREAM_ISSUER` (string): OIDC issuer URL of upstream IdP.
- `--upstream-client-id` / `UPSTREAM_CLIENT_ID` (string)
- `--upstream-client-secret` / `UPSTREAM_CLIENT_SECRET` (string; optional for PKCE public client)
- `--upstream-redirect-url` / `UPSTREAM_REDIRECT_URL` (string): our callback URL (e.g., `${ISSUER}/auth/callback`).
- `--upstream-scopes` / `UPSTREAM_SCOPES` (csv; default `openid,profile,email`)
- `--dev-login` (bool; default false): keep current form login for local dev only.
- `--disable-dev-token-fallback` (bool; default true in prod): disable the DB token fallback in `/mcp`.

Wire these in `main.go`, pass through to `pkg/server` → `pkg/idsrv`.

#### 5.2 Data model (SQLite)

Add a new table for browser sessions created after upstream login:

- `user_sessions(session_id TEXT PRIMARY KEY, subject TEXT NOT NULL, provider TEXT NOT NULL, id_token TEXT, refresh_token TEXT, expires_at TIMESTAMP NOT NULL, created_at TIMESTAMP NOT NULL)`

Notes:
- `subject`: stable unique user identifier from upstream (e.g., `sub` or `email` if guaranteed unique in tenant).
- `session_id`: random, opaque ID stored in a secure cookie; do not store user info in cookie.
- `expires_at`: session expiry; consider renewal using upstream refresh tokens if configured.

#### 5.3 New upstream OIDC flow (login + callback)

Replace the current `/login` POST password check when upstream is enabled:

```go
// Pseudocode (Model A): upstream OIDC RP flow
// In pkg/idsrv/idsrv.go (or a new upstream.go in same package)

type UpstreamOIDC struct {
    Provider     *oidc.Provider
    Verifier     *oidc.IDTokenVerifier
    OAuth2Config *oauth2.Config
    NonceStore   NonceStore // e.g., HMAC-protected state/nonce or DB-backed
}

func (s *Server) initUpstream(ctx context.Context, cfg UpstreamConfig) error {
    p, err := oidc.NewProvider(ctx, cfg.Issuer) // discovery
    if err != nil { return err }
    s.upstream = &UpstreamOIDC{
        Provider: p,
        Verifier: p.Verifier(&oidc.Config{ClientID: cfg.ClientID}),
        OAuth2Config: &oauth2.Config{
            ClientID:     cfg.ClientID,
            ClientSecret: cfg.ClientSecret,
            Endpoint:     p.Endpoint(),
            RedirectURL:  cfg.RedirectURL,
            Scopes:       cfg.Scopes, // include "openid"
        },
    }
    return nil
}

// GET /login → redirect to upstream authorize URL
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
    if s.upstream == nil && !s.devLoginEnabled { http.Error(w, "login disabled", 404); return }
    if s.upstream == nil { /* fall back to dev form (existing) */ }

    state := s.upstream.NonceStore.NewState()
    nonce := s.upstream.NonceStore.NewNonce()
    authURL := s.upstream.OAuth2Config.AuthCodeURL(state, oidc.Nonce(nonce))
    http.Redirect(w, r, authURL, http.StatusFound)
}

// GET /auth/callback → exchange code, verify ID Token, create session
func (s *Server) authCallback(w http.ResponseWriter, r *http.Request) {
    if s.upstream == nil { http.Error(w, "upstream not configured", 500); return }
    if errParam := r.URL.Query().Get("error"); errParam != "" { http.Error(w, errParam, 400); return }

    state := r.URL.Query().Get("state")
    code := r.URL.Query().Get("code")
    if !s.upstream.NonceStore.CheckState(state) { http.Error(w, "bad state", 400); return }

    tok, err := s.upstream.OAuth2Config.Exchange(r.Context(), code)
    if err != nil { http.Error(w, "exchange failed", 400); return }

    rawID := tok.Extra("id_token").(string)
    idt, err := s.upstream.Verifier.Verify(r.Context(), rawID)
    if err != nil { http.Error(w, "invalid id_token", 401); return }

    // Optionally verify nonce claim if set
    // Extract subject and attributes
    var claims struct{ Sub, Email string }
    _ = idt.Claims(&claims)
    subject := claims.Sub
    if subject == "" { subject = claims.Email }

    // Create session in DB and set cookie
    sid := randomID()
    exp := time.Now().Add(12 * time.Hour)
    _ = s.persistSession(Session{SessionID: sid, Subject: subject, Provider: s.upstream.ProviderEndpoint(), IDToken: rawID, RefreshToken: tok.RefreshToken, ExpiresAt: exp})

    http.SetCookie(w, &http.Cookie{Name: cookieName, Value: sid, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, Secure: isHTTPS(s.Issuer), Expires: exp})

    returnTo := r.URL.Query().Get("return_to")
    if returnTo == "" { returnTo = "/" }
    http.Redirect(w, r, returnTo, http.StatusFound)
}

// authorize: replace currentUser() with session lookup by session_id cookie
func (s *Server) authorize(w http.ResponseWriter, r *http.Request) {
    ar, err := s.Provider.NewAuthorizeRequest(r.Context(), r)
    if err != nil { s.Provider.WriteAuthorizeError(r.Context(), w, ar, err); return }

    user, ok := s.lookupSessionSubject(r)
    if !ok { http.Redirect(w, r, "/login?return_to="+url.QueryEscape(r.URL.String()), http.StatusFound); return }

    // continue with DefaultSession creation and NewAuthorizeResponse as today ...
}
```

#### 5.4 Security notes

- Always verify ID Token signature and issuer via discovery. Enforce `aud` on the ID Token verifier.
- Use a dedicated, random `session_id` and store it server‑side; never store user info in the cookie.
- Set `Secure` on cookies when `ISSUER` is HTTPS; consider `SameSite` policy for your deployment.
- Generate and validate `state` and `nonce` for upstream authorization requests.
- Rotate and store `UPSTREAM_CLIENT_SECRET` via env or secrets manager; never commit.
- Consider removing the `/mcp` dev token fallback in production (`--disable-dev-token-fallback`).

#### 5.5 Code edits (high-level map)

- `cmd/apps/mcp-oidc-server/main.go`
  - Add flags/env for upstream OIDC and hardening toggles.
  - Pass config to `pkg/server.New(...)` and down to `idsrv`.
- `pkg/idsrv/idsrv.go`
  - Add upstream config and initialization.
  - Add `/auth/callback` handler.
  - Replace `currentUser` with `lookupSessionSubject` that reads `session_id` → DB.
  - Add SQLite DDL for `user_sessions` in `InitSQLite`.
- `pkg/idsrv/exports.go`
  - Export `AuthCallback` if we keep the adapter pattern.
- `pkg/server/server.go`
  - Keep `/mcp` exactly as-is for AS behavior; optionally expose `/logout` that clears session cookie and deletes the session.

#### 5.6 Database changes (SQLite DDL)

Add the following tables (used across models):

```sql
CREATE TABLE IF NOT EXISTS user_sessions (
  session_id   TEXT PRIMARY KEY,
  subject      TEXT NOT NULL,
  provider     TEXT NOT NULL,
  id_token     TEXT,
  refresh_token TEXT,
  expires_at   TIMESTAMP NOT NULL,
  created_at   TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS users (
  id            INTEGER PRIMARY KEY AUTOINCREMENT,
  username      TEXT UNIQUE NOT NULL,
  email         TEXT,
  password_hash TEXT NOT NULL,
  disabled      INTEGER NOT NULL DEFAULT 0,
  created_at    TIMESTAMP NOT NULL
);
```

#### 5.7 Testing plan

- Local end‑to‑end with Google/Auth0/Zitadel/Keycloak:
  - Configure `UPSTREAM_*` env vars; visit `/oauth2/auth` to trigger the upstream login; verify code issuance and token exchange.
  - Confirm `/mcp` accepts access tokens via introspection and logs subject/client_id.
- Session lifecycle:
  - Verify cookie set/cleared, DB session entries created/deleted, and expiry logic.
- Negative tests:
  - Invalid state/nonce, expired ID Token, wrong issuer/audience.

### 6) Alternative: Ory Kratos integration (Model B)

- Replace `/login` with Kratos browser flow (`/self-service/login/browser`).
- After redirect/callback, use the Kratos Go client to fetch and validate the session; map identity to our `subject`.
- Keep Fosite as our AS and MCP behavior unchanged.
- Good option if we need a self‑hosted identity with passwordless/MFA and profile management without rolling our own.

### 7) Optional future evolution: replace embedded AS with Ory Hydra

- Use **Hydra** as the dedicated OAuth2/OIDC AS (it uses Fosite internally) and create a login/consent app.
- Our service can become only the MCP resource server and drop AS responsibilities.
- Heavier change; not required to fix hardcoded login.

### 8) Rollout plan (checklist)

- [ ] Model C: add `--local-users`, `--session-ttl` flags and config structs.
- [ ] Model C: add `users` DDL and CRUD helpers; argon2id hashing.
- [ ] Model C: implement `/login` DB verification + server‑side sessions in `user_sessions`.
- [ ] Model C: implement `/logout` and session invalidation.
- [ ] Model C: add CLI commands `users add|disable|set-password|list`.
- [ ] Model C: replace `currentUser` with DB session lookup in `authorize`.
- [ ] Security: add CSRF to `/login`, set cookie `Secure`/`HttpOnly`/`SameSite`.
- [ ] Keep `/mcp` middleware; add `--disable-dev-token-fallback` and default it true in non‑dev.
- [ ] Manual E2E: create a user, login, run full auth code + token flow, call `/mcp`.
- [ ] Docs update and ops notes.
- [ ] Later (optional): implement Model A (upstream OIDC) and/or Model B (Kratos) as needed for SSO/MFA.

### 9) References (specs and libraries)

- OIDC Discovery: [openid.net/specs/openid-connect-discovery-1_0.html](https://openid.net/specs/openid-connect-discovery-1_0.html)
- OAuth 2.0 Token Introspection (if needed): [datatracker.ietf.org/doc/html/rfc7662](https://datatracker.ietf.org/doc/html/rfc7662)
- MCP spec (authorization guidance): [modelcontextprotocol.io/specification/2025-03-26/basic/authorization](https://modelcontextprotocol.io/specification/2025-03-26/basic/authorization)
- go-oidc RP library: [pkg.go.dev github.com/coreos/go-oidc/v3/oidc](https://pkg.go.dev/github.com/coreos/go-oidc/v3/oidc)
- Ory Kratos: [ory.sh/kratos/docs](https://www.ory.sh/kratos/docs/)
- Dex: [dexidp.io/docs](https://dexidp.io/docs/)
- ZITADEL: [docs.zitadel.com](https://docs.zitadel.com/)
- Keycloak: [keycloak.org/docs/latest](https://www.keycloak.org/docs/latest/)
- Auth0 Go Quickstart: [auth0.com/docs/quickstart/webapp/golang](https://auth0.com/docs/quickstart/webapp/golang)
