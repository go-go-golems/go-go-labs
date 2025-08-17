Great progress — your discovery + dynamic registration look healthy. The “falls over after /oauth2/auth” symptom is consistent with a **Fosite composer mis‑wiring** in your server plus one ID‑token detail that will bite you in the next step.

Below I’ll (1) pinpoint the immediate cause, (2) give you exact code diffs, (3) show what to log to get first‑class error detail from Fosite, and (4) give you a quick local verification script.

---

## TL;DR (what’s wrong)

1. **You’re passing the wrong thing to `compose.ComposeAllEnabled` and never set the HMAC global secret.**

* In your code you do:

  ```go
  secret := []byte("dev-hmac-secret-change-me")
  provider := compose.ComposeAllEnabled(cfg, mem, secret)
  ```

  The **third parameter** of `ComposeAllEnabled` is expected to be the **RSA private key used to sign OIDC ID Tokens**, *not* the HMAC secret. The HMAC secret must be placed on `cfg.GlobalSecret`, and Fosite expects it to be **exactly 32 bytes**. The official example shows exactly this pattern: set `GlobalSecret` in `Config` and pass the **RSA private key** to `ComposeAllEnabled`. ([Go Packages][1], [GitHub][2])

* Because `cfg.GlobalSecret` is never set (and you gave `[]byte` to the private‑key slot), the **authorize code signer has no HMAC secret**, so `NewAuthorizeResponse` can’t mint the authorization code and returns a generic `server_error`. That matches your log line:

  ```
  ERR failed issuing code error=server_error endpoint=/oauth2/auth
  ```

  (This error will be unwrapped to an RFC6749 error once you add the logging below.)

2. **(Next step issue)** Your ID token **audience** is set to `s.Issuer + "/mcp"`. For OIDC, the **ID Token `aud` MUST contain the OAuth 2.0 `client_id`** of the RP (ChatGPT). If you keep `aud` as the resource URL, ChatGPT will reject the ID Token after the code exchange. Fix below. ([OpenID Foundation][3])

All of this is visible in your current source (composer call, missing `GlobalSecret`, and custom `aud` in session).&#x20;

---

## Exact fixes (minimal diffs)

**File:** `pkg/idsrv/idsrv.go` &#x20;

### 1) Configure Fosite correctly (GlobalSecret + pass RSA private key)

```diff
 func New(issuer string) (*Server, error) {
     privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
     if err != nil { return nil, err }

-    cfg := &fosite.Config{
-        IDTokenIssuer:               issuer,
-        EnforcePKCEForPublicClients: true,
-    }
+    cfg := &fosite.Config{
+        IDTokenIssuer:               issuer,
+        EnforcePKCEForPublicClients: true,
+        // REQUIRED for HMAC-signed auth codes/refresh tokens:
+        // must be EXACTLY 32 bytes (see fosite docs).
+        GlobalSecret: []byte("0123456789abcdef0123456789abcdef"),
+    }

     mem := storage.NewMemoryStore()
     // Dev client for manual tests
     devClientID := "dev-client"
     devRedirect := issuer + "/dev/callback"
     mem.Clients[devClientID] = &fosite.DefaultClient{
         ID:            devClientID,
         RedirectURIs:  []string{devRedirect},
         GrantTypes:    []string{"authorization_code", "refresh_token"},
         ResponseTypes: []string{"code"},
         Scopes:        []string{"openid", "profile", "offline_access"},
         Public:        true,
     }

-    secret := []byte("dev-hmac-secret-change-me")
-    // Use high-level composer compatible with fosite v0.49.0
-    provider := compose.ComposeAllEnabled(cfg, mem, secret)
+    // Pass the RSA private key here — NOT the HMAC secret.
+    provider := compose.ComposeAllEnabled(cfg, mem, privateKey)
```

Why: Fosite’s example sets `Config.GlobalSecret = secret` (32 bytes) and passes an RSA private key to `ComposeAllEnabled`. Your current code does the opposite, which breaks both the code signer and OIDC signing. ([GitHub][2], [Go Packages][1])

### 2) Make the ID Token audience correct

Right now you hard‑code:

```go
Audience:    []string{s.Issuer + "/mcp"},
```

Change it to the **registered client ID** (what ChatGPT dynamically registers), or simply let Fosite set it. Safest explicit fix:

```diff
-    sess := &openid.DefaultSession{
+    sess := &openid.DefaultSession{
         Subject:  user,
         Username: user,
         Claims: &jwt.IDTokenClaims{
             Subject:     user,
             Issuer:      s.Issuer,
             IssuedAt:    now,
             AuthTime:    now,
             RequestedAt: now,
-            Audience:    []string{s.Issuer + "/mcp"},
+            Audience:    []string{ar.GetClient().GetID()},
         },
         Headers: &jwt.Headers{Extra: map[string]any{"kid": "1"}},
     }
```

Reason: OIDC requires `aud` in the ID Token to include the OAuth 2.0 `client_id` of the RP. If it contains your resource URL instead, clients will (correctly) reject it after the token exchange. ([OpenID Foundation][3])

> Tip: You can also omit `Audience` and let the OIDC strategy infer it; being explicit avoids surprises.

---

## Add high‑signal error logs (so the next unknown isn’t “server\_error”)

Fosite exposes RFC6749‑rich errors. Log them in both authorize and token handlers:

```diff
@@ func (s *Server) authorize(w http.ResponseWriter, r *http.Request) {
-    resp, err := s.Provider.NewAuthorizeResponse(ctx, ar, sess)
-    if err != nil {
-        log.Error().Err(err).Str("endpoint", "/oauth2/auth").Msg("failed issuing code")
-        s.Provider.WriteAuthorizeError(ctx, w, ar, err)
-        return
-    }
+    resp, err := s.Provider.NewAuthorizeResponse(ctx, ar, sess)
+    if err != nil {
+        rfc := fosite.ErrorToRFC6749Error(err)
+        log.Error().
+            Err(err).
+            Str("endpoint", "/oauth2/auth").
+            Str("rfc_name", rfc.Name).
+            Str("rfc_hint", rfc.Hint).
+            Str("rfc_description", rfc.Description).
+            Msg("failed issuing code")
+        s.Provider.WriteAuthorizeError(ctx, w, ar, err)
+        return
+    }
```

And similarly in `token`:

```diff
@@ func (s *Server) token(w http.ResponseWriter, r *http.Request) {
-    accessReq, err := s.Provider.NewAccessRequest(ctx, r, sess)
+    accessReq, err := s.Provider.NewAccessRequest(ctx, r, sess)
     if err != nil {
-        log.Error().Err(err).Str("endpoint", "/oauth2/token").Msg("access request error")
+        rfc := fosite.ErrorToRFC6749Error(err)
+        log.Error().Err(err).
+            Str("endpoint", "/oauth2/token").
+            Str("rfc_name", rfc.Name).
+            Str("rfc_hint", rfc.Hint).
+            Str("rfc_description", rfc.Description).
+            Msg("access request error")
         s.Provider.WriteAccessError(ctx, w, accessReq, err)
         return
     }
-    resp, err := s.Provider.NewAccessResponse(ctx, accessReq)
+    resp, err := s.Provider.NewAccessResponse(ctx, accessReq)
     if err != nil {
-        log.Error().Err(err).Str("endpoint", "/oauth2/token").Msg("access response error")
+        rfc := fosite.ErrorToRFC6749Error(err)
+        log.Error().Err(err).
+            Str("endpoint", "/oauth2/token").
+            Str("rfc_name", rfc.Name).
+            Str("rfc_hint", rfc.Hint).
+            Str("rfc_description", rfc.Description).
+            Msg("access response error")
         s.Provider.WriteAccessError(ctx, w, accessReq, err)
         return
     }
```

With this, if anything else is off (PKCE mismatch, redirect mismatch, etc.) you’ll see the **normalized OAuth2 error name**, a **hint**, and a **human‑readable description** in your logs.

---

## Why ChatGPT “stops” after `/oauth2/auth`

* The UI flow is: discovery → **dynamic client registration** → browser sends the **authorize** request (note the UA switching from `aiohttp` to `Firefox …` in your logs) → your server fails at `NewAuthorizeResponse` and issues a 303 redirect with `error=server_error` back to ChatGPT’s redirect URI → the platform treats that as a hard stop (as it should) and doesn’t proceed to `/oauth2/token`.

Your logs show that exact sequence; after you fix the composer/secret, you should see **“issued authorization code”** and then the `/oauth2/token` call from ChatGPT.

---

## Quick local verification (before trying ChatGPT again)

1. **Start server** on your ngrok origin (or localhost for the dev client):

   ```bash
   curl -sSf https://<issuer>/.well-known/oauth-authorization-server | jq .
   # verify:
   # - issuer is exact
   # - authorization_endpoint, token_endpoint, jwks_uri correct
   # - code_challenge_methods_supported includes "S256"
   ```

2. **Dev round‑trip** using your built‑in `dev-client` (no browser):

   ```bash
   # 1) Create PKCE inputs
   VERIFIER=$(python3 - <<'PY'
   ```

import os,base64,hashlib
v=base64.urlsafe\_b64encode(os.urandom(32)).decode().rstrip('=')
print(v)
PY
)
CHAL=\$(python3 - <\<PY
import base64,hashlib,os,sys
v=os.environ\["VERIFIER"].encode()
print(base64.urlsafe\_b64encode(hashlib.sha256(v).digest()).decode().rstrip('='))
PY
)
echo "verifier=\$VERIFIER"
echo "challenge=\$CHAL"

# 2) Get code

AUTH\_URL="https\://<issuer>/oauth2/auth?response\_type=code\&client\_id=dev-client\&redirect\_uri=https%3A%2F%2F<issuer-host-escaped>%2Fdev%2Fcallback\&code\_challenge\_method=S256\&code\_challenge=\$CHAL\&scope=openid\&state=s123"

# Because your /authorize redirects to login when not authenticated, first login to get the cookie:

curl -i "https\://<issuer>/login?return\_to=\$(python3 - <\<PY
import urllib.parse;print(urllib.parse.quote('https\://<issuer>/oauth2/auth?response\_type=code\&client\_id=dev-client\&redirect\_uri=https\://<issuer>/dev/callback\&code\_challenge\_method=S256\&code\_challenge=\$CHAL\&scope=openid\&state=s123'))
PY
)"&#x20;
\| grep -i set-cookie  # grab sid cookie value

# Now POST credentials and follow redirects with cookie:

curl -i -c /tmp/cjar -b /tmp/cjar -X POST "https\://<issuer>/login"&#x20;
-d "username=admin\&password=password123\&return\_to=https\://<issuer>/oauth2/auth?response\_type=code\&client\_id=dev-client\&redirect\_uri=https\://<issuer>/dev/callback\&code\_challenge\_method=S256\&code\_challenge=\$CHAL\&scope=openid\&state=s123"&#x20;
-L

# Response should redirect to /dev/callback with ?code=...\&state=s123

````

3) **Exchange code**:

```bash
curl -sS -X POST "https://<issuer>/oauth2/token" \
  -d "grant_type=authorization_code" \
  -d "client_id=dev-client" \
  -d "redirect_uri=https://<issuer>/dev/callback" \
  -d "code=<code-from-previous-step>" \
  -d "code_verifier=$VERIFIER" \
  | jq .
````

Expect: `access_token`, `refresh_token` (because your client allows it), and `id_token` (since you requested `openid`). If anything fails, your new logs will show **`rfc_name` / `rfc_hint`**.

---

## A few sanity notes for your setup

* **JWKS `kid`**: Your ID token header sets `kid: "1"`, and your JWKS publishes key with `kid: "1"` — that’s fine. Keep them aligned.

* **Scopes**: ChatGPT may omit `scope` in the initial request. That’s legal for OAuth 2.0; you don’t need to force a value to issue a code. If you want an ID token for every login, you can encourage `openid` by documentation, but don’t mutate the request server‑side.

* **Redirect URI exact match**: You’re persisting the redirect URI from dynamic registration; ensure it remains **exact** (no trailing slashes, query params changes). Fosite enforces exact matching. ([GitHub][4])

---

## Why I’m confident this is the root cause

* Fosite’s own quickstart shows: **`GlobalSecret` (32 bytes) goes on `Config`**, and **`ComposeAllEnabled(config, storage, privateKey)`** takes an RSA key, not the HMAC secret. Your code does the inverse. ([GitHub][2], [Go Packages][1])
* Your authorize logs show a fully valid request (PKCE S256, correct redirect, `response_type=code`), so a **server‑side issuance failure** is the only plausible reason for `server_error` before any token step. When the signer can’t mint the authorization code (missing HMAC secret), `NewAuthorizeResponse` fails exactly there.
* Fixing the composer wiring is a classic “first‑run Fosite” pitfall that matches these symptoms.

---

## After the change — what to look for in logs

* `/oauth2/auth`:

  * `issued authorization code` (your existing info log)
  * No more `server_error`; if there is any, you’ll see `rfc_name`, `rfc_hint`, `rfc_description`.

* `/oauth2/token`:

  * `token exchange success` and an `id_token` in the JSON when `openid` is present.

---

If you want, I can sketch a tiny test that asserts `NewAuthorizeResponse` succeeds when `GlobalSecret` is present and fails when it’s blank (to prevent regressions). But the code diffs above should unblock you right now.

[1]: https://pkg.go.dev/github.com/ory/fosite/compose "compose package - github.com/ory/fosite/compose - Go Packages"
[2]: https://github.com/ory/fosite "GitHub - ory/fosite: Extensible security first OAuth 2.0 and OpenID Connect SDK for Go."
[3]: https://openid.net/specs/openid-connect-core-1_0.html?utm_source=chatgpt.com "OpenID Connect Core 1.0 incorporating errata set 2"
[4]: https://github.com/ory/fosite/issues/685?utm_source=chatgpt.com "Investigate missing `redirect_uri` check when performing ..."
