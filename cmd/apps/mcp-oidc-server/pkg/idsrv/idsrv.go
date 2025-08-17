package idsrv

import (
    "database/sql"
    "crypto/rand"
    "crypto/rsa"
    "crypto/x509"
    "encoding/base64"
    "encoding/json"
    "encoding/pem"
    "html/template"
    "math/big"
    "net/http"
    "net/url"
    "strings"
    "sync"
    "time"

    "github.com/ory/fosite"
    "github.com/ory/fosite/compose"
    "github.com/ory/fosite/handler/openid"
    "github.com/ory/fosite/storage"
    "github.com/ory/fosite/token/jwt"
    _ "github.com/mattn/go-sqlite3"
    "github.com/rs/zerolog/log"
)

type Server struct {
    PrivateKey *rsa.PrivateKey
    Issuer     string
    Provider   fosite.OAuth2Provider
    store      *storage.MemoryStore
    cfg        *fosite.Config
    mu         sync.Mutex
    // optional persistence
    // set when InitSQLite is called
    // guarded by mu for writes
    dbPath string
    // demo login
    User string
    Pass string
}

func New(issuer string) (*Server, error) {
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil { return nil, err }

    cfg := &fosite.Config{
        IDTokenIssuer:               issuer,
        EnforcePKCEForPublicClients: true,
        GlobalSecret:                []byte("0123456789abcdef0123456789abcdef"), // 32 bytes
    }

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

    // Use high-level composer: pass the RSA private key (not the HMAC secret)
    provider := compose.ComposeAllEnabled(cfg, mem, privateKey)

    log.Debug().Str("component", "idsrv").Str("issuer", issuer).Str("dev_client_id", devClientID).Str("dev_redirect", devRedirect).Msg("initialized identity server")

    return &Server{
        PrivateKey: privateKey,
        Issuer:     issuer,
        Provider:   provider,
        store:      mem,
        cfg:        cfg,
        User:       "admin",
        Pass:       "password123",
    }, nil
}

func (s *Server) Routes(mux *http.ServeMux) {
    mux.HandleFunc("/.well-known/openid-configuration", s.oidcDiscovery)
    mux.HandleFunc("/.well-known/oauth-authorization-server", s.asMetadata)
    mux.HandleFunc("/jwks.json", s.jwks)
    mux.HandleFunc("/login", s.login)
    mux.HandleFunc("/oauth2/auth", s.authorize)
    mux.HandleFunc("/oauth2/token", s.token)
    mux.HandleFunc("/register", s.register)
}

func (s *Server) ProviderRef() fosite.OAuth2Provider { return s.Provider }

func (s *Server) oidcDiscovery(w http.ResponseWriter, r *http.Request) {
    log.Debug().Str("endpoint", "/.well-known/openid-configuration").Msg("serving OIDC discovery")
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
        "code_challenge_methods_supported":       []string{"S256"},
        "registration_endpoint":                  s.Issuer + "/register",
    }
    writeJSON(w, j)
}

func (s *Server) asMetadata(w http.ResponseWriter, r *http.Request) {
    log.Debug().Str("endpoint", "/.well-known/oauth-authorization-server").Msg("serving AS metadata")
    j := map[string]any{
        "issuer":                 s.Issuer,
        "authorization_endpoint": s.Issuer + "/oauth2/auth",
        "token_endpoint":         s.Issuer + "/oauth2/token",
        "jwks_uri":               s.Issuer + "/jwks.json",
        "code_challenge_methods_supported": []string{"S256"},
        "response_types_supported":          []string{"code"},
        "grant_types_supported":             []string{"authorization_code", "refresh_token"},
        "scopes_supported":                  []string{"openid", "profile", "offline_access"},
        "token_endpoint_auth_methods_supported": []string{"none"},
        "registration_endpoint":               s.Issuer + "/register",
    }
    writeJSON(w, j)
}

func (s *Server) jwks(w http.ResponseWriter, r *http.Request) {
    log.Debug().Str("endpoint", "/jwks.json").Msg("serving JWKS")
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

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
    switch r.Method {
    case http.MethodGet:
        log.Debug().Str("endpoint", "/login").Str("method", "GET").Str("return_to", r.URL.Query().Get("return_to")).Msg("render login")
        _ = loginTpl.Execute(w, struct{ ReturnTo string }{r.URL.Query().Get("return_to")})
    case http.MethodPost:
        _ = r.ParseForm()
        u := r.FormValue("username")
        log.Debug().Str("endpoint", "/login").Str("method", "POST").Str("username", u).Str("return_to", r.FormValue("return_to")).Msg("attempt login")
        p := r.FormValue("password")
        if u == s.User && p == s.Pass {
            http.SetCookie(w, &http.Cookie{Name: cookieName, Value: "ok:"+u, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode})
            rt := r.FormValue("return_to")
            if rt == "" { rt = "/" }
            log.Info().Str("endpoint", "/login").Str("username", u).Str("return_to", rt).Msg("login success, redirecting")
            http.Redirect(w, r, rt, http.StatusFound)
            return
        }
        log.Warn().Str("endpoint", "/login").Str("username", u).Msg("invalid credentials")
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
    q := r.URL.Query()
    log.Info().
        Str("endpoint", "/oauth2/auth").
        Str("client_id", q.Get("client_id")).
        Str("redirect_uri", q.Get("redirect_uri")).
        Str("response_type", q.Get("response_type")).
        Str("scope", q.Get("scope")).
        Str("state", q.Get("state")).
        Str("code_challenge_method", q.Get("code_challenge_method")).
        Str("code_challenge", q.Get("code_challenge")).
        Str("resource", q.Get("resource")).
        Msg("authorize request")
    ar, err := s.Provider.NewAuthorizeRequest(ctx, r)
    if err != nil {
        log.Error().Err(err).
            Str("endpoint", "/oauth2/auth").
            Str("client_id", q.Get("client_id")).
            Str("redirect_uri", q.Get("redirect_uri")).
            Msg("authorize error")
        s.Provider.WriteAuthorizeError(ctx, w, ar, err)
        return
    }
    user, ok := currentUser(r)
    if !ok {
        log.Debug().Str("endpoint", "/oauth2/auth").Msg("not logged in, redirect to /login")
        http.Redirect(w, r, "/login?return_to="+url.QueryEscape(r.URL.String()), http.StatusFound)
        return
    }
    now := time.Now()
    sess := &openid.DefaultSession{
        Subject:  user,
        Username: user,
        Claims: &jwt.IDTokenClaims{
            Subject:     user,
            Issuer:      s.Issuer,
            IssuedAt:    now,
            AuthTime:    now,
            RequestedAt: now,
            Audience:    []string{ar.GetClient().GetID()},
        },
        Headers: &jwt.Headers{Extra: map[string]any{"kid": "1"}},
    }
    resp, err := s.Provider.NewAuthorizeResponse(ctx, ar, sess)
    if err != nil {
        rfc := fosite.ErrorToRFC6749Error(err)
        log.Error().Err(err).
            Str("endpoint", "/oauth2/auth").
            Str("rfc_error", rfc.ErrorField).
            Str("rfc_hint", rfc.HintField).
            Str("rfc_description", rfc.DescriptionField).
            Msg("failed issuing code")
        s.Provider.WriteAuthorizeError(ctx, w, ar, err)
        return
    }
    log.Info().Str("endpoint", "/oauth2/auth").Str("client_id", q.Get("client_id")).Str("redirect_uri", q.Get("redirect_uri")).Msg("issued authorization code")
    s.Provider.WriteAuthorizeResponse(ctx, w, ar, resp)
}

func (s *Server) token(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    _ = r.ParseForm()
    form := r.PostForm
    log.Info().
        Str("endpoint", "/oauth2/token").
        Str("grant_type", form.Get("grant_type")).
        Str("client_id", form.Get("client_id")).
        Str("redirect_uri", form.Get("redirect_uri")).
        Bool("has_code", form.Get("code") != "").
        Bool("has_code_verifier", form.Get("code_verifier") != "").
        Bool("has_refresh_token", form.Get("refresh_token") != "").
        Msg("token request")
    sess := new(openid.DefaultSession)
    accessReq, err := s.Provider.NewAccessRequest(ctx, r, sess)
    if err != nil {
        rfc := fosite.ErrorToRFC6749Error(err)
        log.Error().Err(err).
            Str("endpoint", "/oauth2/token").
            Str("rfc_error", rfc.ErrorField).
            Str("rfc_hint", rfc.HintField).
            Str("rfc_description", rfc.DescriptionField).
            Msg("access request error")
        s.Provider.WriteAccessError(ctx, w, accessReq, err)
        return
    }
    resp, err := s.Provider.NewAccessResponse(ctx, accessReq)
    if err != nil {
        rfc := fosite.ErrorToRFC6749Error(err)
        log.Error().Err(err).
            Str("endpoint", "/oauth2/token").
            Str("rfc_error", rfc.ErrorField).
            Str("rfc_hint", rfc.HintField).
            Str("rfc_description", rfc.DescriptionField).
            Msg("access response error")
        s.Provider.WriteAccessError(ctx, w, accessReq, err)
        return
    }
    log.Info().Str("endpoint", "/oauth2/token").Str("grant_type", form.Get("grant_type")).Msg("token exchange success")
    s.Provider.WriteAccessResponse(ctx, w, accessReq, resp)
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        RedirectURIs []string `json:"redirect_uris"`
        TokenEndpointAuthMethod string `json:"token_endpoint_auth_method"`
        GrantTypes []string `json:"grant_types"`
        ResponseTypes []string `json:"response_types"`
        ClientName string `json:"client_name"`
        ClientID   string `json:"client_id"`
    }
    _ = json.NewDecoder(r.Body).Decode(&payload)
    log.Debug().Str("endpoint", "/register").Interface("payload", payload).Msg("dynamic registration request")
    if len(payload.RedirectURIs) == 0 {
        http.Error(w, "missing redirect_uris", http.StatusBadRequest)
        return
    }
    id := payload.ClientID
    if id == "" {
        id = "client-" + time.Now().Format("20060102-150405")
    }
    // persist in memory store
    s.mu.Lock()
    s.store.Clients[id] = &fosite.DefaultClient{
        ID:            id,
        RedirectURIs:  payload.RedirectURIs,
        GrantTypes:    []string{"authorization_code", "refresh_token"},
        ResponseTypes: []string{"code"},
        Scopes:        []string{"openid", "profile", "offline_access"},
        Public:        true,
    }
    s.mu.Unlock()
    // persist if enabled
    if err := s.persistClientToSQLite(id, payload.RedirectURIs); err != nil {
        log.Error().Err(err).Str("endpoint", "/register").Str("client_id", id).Msg("failed to persist client")
    }
    log.Info().Str("endpoint", "/register").Str("client_id", id).Interface("redirect_uris", payload.RedirectURIs).Msg("dynamic client registered")
    resp := map[string]any{
        "client_id":                  id,
        "redirect_uris":              payload.RedirectURIs,
        "token_endpoint_auth_method": "none",
        "grant_types":                []string{"authorization_code", "refresh_token"},
        "response_types":             []string{"code"},
        "code_challenge_methods_supported": []string{"S256"},
    }
    writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, v any) {
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(v)
}

// InitSQLite enables persistence of clients to a simple SQLite DB at the given path and loads existing clients on boot.
// This is optional; if not called, in-memory storage is used only.
func (s *Server) InitSQLite(path string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.dbPath = path
    // Create table if not exists and load clients
    db, err := sqlOpen(path)
    if err != nil { return err }
    defer db.Close()
    if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS oauth_clients (
        client_id TEXT PRIMARY KEY,
        redirect_uris TEXT NOT NULL
    );`); err != nil { return err }
    if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS oauth_keys (
        kid TEXT PRIMARY KEY,
        private_pem BLOB NOT NULL,
        created_at TIMESTAMP NOT NULL
    );`); err != nil { return err }
    if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS oauth_tokens (
        token TEXT PRIMARY KEY,
        subject TEXT NOT NULL,
        client_id TEXT NOT NULL,
        scopes TEXT NOT NULL,
        expires_at TIMESTAMP NOT NULL
    );`); err != nil { return err }

    if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS mcp_tool_calls (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ts TIMESTAMP NOT NULL,
        subject TEXT,
        client_id TEXT,
        request_id TEXT,
        tool_name TEXT NOT NULL,
        args_json TEXT,
        result_json TEXT,
        status TEXT NOT NULL,
        duration_ms INTEGER NOT NULL
    );`); err != nil { return err }

    // Load or persist signing key
    var existingPEM []byte
    var existingKID string
    row := db.QueryRow(`SELECT kid, private_pem FROM oauth_keys LIMIT 1`)
    _ = row.Scan(&existingKID, &existingPEM)
    if len(existingPEM) > 0 {
        pk, err := pemDecodeRSAPrivateKey(existingPEM)
        if err != nil { return err }
        s.PrivateKey = pk
        // Recompose provider with persisted key
        s.Provider = compose.ComposeAllEnabled(s.cfg, s.store, s.PrivateKey)
        log.Info().Str("component", "idsrv").Str("db", path).Str("kid", existingKID).Msg("loaded signing key from sqlite")
    } else {
        // Persist current in-memory key
        pemBytes, err := pemEncodeRSAPrivateKey(s.PrivateKey)
        if err != nil { return err }
        if _, err := db.Exec(`INSERT INTO oauth_keys (kid, private_pem, created_at) VALUES (?, ?, ?)`, "1", pemBytes, time.Now()); err != nil { return err }
        log.Info().Str("component", "idsrv").Str("db", path).Str("kid", "1").Msg("persisted new signing key to sqlite")
    }
    // Load existing
    rows, err := db.Query(`SELECT client_id, redirect_uris FROM oauth_clients`)
    if err != nil { return err }
    defer rows.Close()
    loaded := 0
    for rows.Next() {
        var id string
        var uris string
        if err := rows.Scan(&id, &uris); err != nil { return err }
        redirects := splitCSV(uris)
        s.store.Clients[id] = &fosite.DefaultClient{
            ID:            id,
            RedirectURIs:  redirects,
            GrantTypes:    []string{"authorization_code", "refresh_token"},
            ResponseTypes: []string{"code"},
            Scopes:        []string{"openid", "profile", "offline_access"},
            Public:        true,
        }
        loaded++
    }
    log.Info().Str("component", "idsrv").Str("db", path).Int("clients", loaded).Msg("loaded clients from sqlite")
    return nil
}

func pemEncodeRSAPrivateKey(pk *rsa.PrivateKey) ([]byte, error) {
    b := x509.MarshalPKCS1PrivateKey(pk)
    blk := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: b}
    return pem.EncodeToMemory(blk), nil
}

func pemDecodeRSAPrivateKey(p []byte) (*rsa.PrivateKey, error) {
    blk, _ := pem.Decode(p)
    if blk == nil || blk.Type != "RSA PRIVATE KEY" { return nil, fosite.ErrServerError.WithHint("invalid PEM for RSA private key") }
    return x509.ParsePKCS1PrivateKey(blk.Bytes)
}

func (s *Server) persistClientToSQLite(id string, redirects []string) error {
    if s.dbPath == "" { return nil }
    db, err := sqlOpen(s.dbPath)
    if err != nil { return err }
    defer db.Close()
    _, err = db.Exec(`INSERT OR REPLACE INTO oauth_clients (client_id, redirect_uris) VALUES (?, ?)`, id, joinCSV(redirects))
    return err
}

func splitCSV(s string) []string { if s == "" { return nil }; return strings.Split(s, ",") }
func joinCSV(ss []string) string { return strings.Join(ss, ",") }

// sqlite open indirection for single-file path
func sqlOpen(path string) (*sql.DB, error) { return sql.Open("sqlite3", path) }

// Token persistence helpers
type TokenRecord struct {
    Token     string
    Subject   string
    ClientID  string
    Scopes    []string
    ExpiresAt time.Time
}

func (s *Server) PersistToken(tr TokenRecord) error {
    if s.dbPath == "" { return fosite.ErrServerError.WithHint("db not enabled") }
    db, err := sqlOpen(s.dbPath)
    if err != nil { return err }
    defer db.Close()
    _, err = db.Exec(`INSERT OR REPLACE INTO oauth_tokens (token, subject, client_id, scopes, expires_at) VALUES (?, ?, ?, ?, ?)`,
        tr.Token, tr.Subject, tr.ClientID, joinCSV(tr.Scopes), tr.ExpiresAt)
    return err
}

func (s *Server) GetToken(token string) (TokenRecord, bool, error) {
    var out TokenRecord
    if s.dbPath == "" { return out, false, fosite.ErrServerError.WithHint("db not enabled") }
    db, err := sqlOpen(s.dbPath)
    if err != nil { return out, false, err }
    defer db.Close()
    row := db.QueryRow(`SELECT token, subject, client_id, scopes, expires_at FROM oauth_tokens WHERE token = ?`, token)
    var scopes string
    if err := row.Scan(&out.Token, &out.Subject, &out.ClientID, &scopes, &out.ExpiresAt); err != nil {
        if err == sql.ErrNoRows { return out, false, nil }
        return out, false, err
    }
    out.Scopes = splitCSV(scopes)
    return out, true, nil
}

func (s *Server) ListTokens() ([]TokenRecord, error) {
    if s.dbPath == "" { return nil, fosite.ErrServerError.WithHint("db not enabled") }
    db, err := sqlOpen(s.dbPath)
    if err != nil { return nil, err }
    defer db.Close()
    rows, err := db.Query(`SELECT token, subject, client_id, scopes, expires_at FROM oauth_tokens ORDER BY expires_at DESC`)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []TokenRecord
    for rows.Next() {
        var tr TokenRecord
        var scopes string
        if err := rows.Scan(&tr.Token, &tr.Subject, &tr.ClientID, &scopes, &tr.ExpiresAt); err != nil { return nil, err }
        tr.Scopes = splitCSV(scopes)
        out = append(out, tr)
    }
    return out, nil
}

// MCP tool call logging
type MCPCallLog struct {
    Timestamp  time.Time
    Subject    string
    ClientID   string
    RequestID  string
    ToolName   string
    ArgsJSON   string
    ResultJSON string
    Status     string
    DurationMs int64
}

func (s *Server) LogMCPCall(entry MCPCallLog) error {
    if s.dbPath == "" { return fosite.ErrServerError.WithHint("db not enabled") }
    db, err := sqlOpen(s.dbPath)
    if err != nil { return err }
    defer db.Close()
    if entry.Timestamp.IsZero() { entry.Timestamp = time.Now() }
    _, err = db.Exec(`INSERT INTO mcp_tool_calls (ts, subject, client_id, request_id, tool_name, args_json, result_json, status, duration_ms)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
        entry.Timestamp, entry.Subject, entry.ClientID, entry.RequestID, entry.ToolName, entry.ArgsJSON, entry.ResultJSON, entry.Status, entry.DurationMs)
    return err
}
