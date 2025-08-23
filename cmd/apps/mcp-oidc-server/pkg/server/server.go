package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"
	"strconv"

	idsrv "github.com/go-go-golems/go-go-labs/cmd/apps/mcp-oidc-server/pkg/idsrv"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/rs/zerolog/log"
)

type Server struct {
	issuer string
	ids    *idsrv.Server
	devTokenFallbackEnabled bool
}

// auth context keys
type ctxKey string

const (
	ctxSubjectKey  ctxKey = "mcp_subject"
	ctxClientIDKey ctxKey = "mcp_client_id"
)

func setAuthCtx(ctx context.Context, subject, clientID string) context.Context {
	ctx = context.WithValue(ctx, ctxSubjectKey, subject)
	ctx = context.WithValue(ctx, ctxClientIDKey, clientID)
	return ctx
}

func getAuthCtx(ctx context.Context) (string, string) {
	subj, _ := ctx.Value(ctxSubjectKey).(string)
	cid, _ := ctx.Value(ctxClientIDKey).(string)
	return subj, cid
}

func New(issuer string) (*Server, error) {
	id, err := idsrv.New(issuer)
	if err != nil {
		return nil, err
	}
	return &Server{issuer: issuer, ids: id, devTokenFallbackEnabled: true}, nil
}

// EnableSQLite enables SQLite persistence for dynamic client registrations.
func (s *Server) EnableSQLite(path string) error {
	if path == "" { return nil }
	if err := s.ids.InitSQLite(path); err != nil { return err }
	log.Info().Str("component", "server").Str("db", path).Msg("sqlite persistence enabled")
	return nil
}

func (s *Server) Routes(mux *http.ServeMux) {
	// health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("endpoint", "healthz").Str("ua", r.UserAgent()).Msg("health check")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// root page: login status and links
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		subj, ok := s.ids.SubjectFromRequest(r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if ok {
			_, _ = w.Write([]byte("<h3>MCP OIDC Server</h3><p>Status: <b>Logged in</b> as <b>" + subj + "</b></p><ul><li><a href=\"/whoami\">Who am I</a></li><li><a href=\"/logout\">Logout</a></li></ul>"))
			return
		}
		_, _ = w.Write([]byte("<h3>MCP OIDC Server</h3><p>Status: <b>Not logged in</b></p><ul><li><a href=\"/login\">Login</a></li><li><a href=\"/whoami\">Who am I</a></li></ul>"))
	})

	// Delegate all IdP routes (discovery, jwks, oauth2, login, register) to idsrv
	s.ids.Routes(mux)

	// OAuth 2.0 Protected Resource Metadata (RFC 9728)
	mux.HandleFunc("/.well-known/oauth-protected-resource", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("endpoint", "oauth-protected-resource").Msg("serving protected resource metadata")
		j := map[string]any{
			"authorization_servers": []string{s.issuer},
			"resource":              s.issuer + "/mcp",
		}
		writeJSONWithPreview(w, r, "oauth-protected-resource", j)
		log.Info().Str("endpoint", "oauth-protected-resource").Interface("response", j).Msg("served protected resource metadata")
	})

	// Dev callback for manual testing: echoes code/state
	mux.HandleFunc("/dev/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		resp := map[string]any{
			"code":  q.Get("code"),
			"state": q.Get("state"),
		}
		log.Info().Str("endpoint", "/dev/callback").Interface("resp", resp).Msg("dev callback")
		writeJSONWithPreview(w, r, "dev-callback", resp)
	})

	// MCP endpoint: JSON-RPC with Bearer protection
	mux.Handle("/mcp", s.mcpAuthMiddleware(http.HandlerFunc(s.handleMCP)))

	// whoami: simple test page
	mux.HandleFunc("/whoami", func(w http.ResponseWriter, r *http.Request) {
		subject, ok := s.ids.SubjectFromRequest(r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if !ok {
			_, _ = w.Write([]byte("<h3>Not logged in</h3><p><a href=\"/login\">Login</a></p>"))
			return
		}
		// basic tools listing (same as tools/list payload)
		tools := []map[string]any{
			{
				"name": "search",
				"description": "Search corpus and return candidate items",
				"require_approval": "never",
			},
			{
				"name": "fetch",
				"description": "Fetch a record by ID",
				"require_approval": "never",
			},
		}
		b, _ := json.MarshalIndent(tools, "", "  ")
		// include dev token info if generated
		devTok := r.URL.Query().Get("dev_token")
		devMsg := ""
		if devTok != "" {
			devMsg = "<h4>Dev API token</h4><p>Use in Authorization header:</p><pre>Authorization: Bearer " + devTok + "</pre>"
			if !s.devTokenFallbackEnabled {
				devMsg += "<p><i>Note: dev token fallback is disabled on the server, this token won't work for /mcp unless enabled.</i></p>"
			}
		}
		html := "<h3>Logged in</h3><p>User: <b>" + subject + "</b></p>" + devMsg + "<form method=\"post\" action=\"/dev/token\"><button type=\"submit\">Generate API token (dev)</button></form><h4>Tools</h4><pre>" + string(b) + "</pre><p><a href=\"/logout\">Logout</a></p>"
		_, _ = w.Write([]byte(html))
	})

	// dev token generator: creates an API token for current user
	mux.HandleFunc("/dev/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		subject, ok := s.ids.SubjectFromRequest(r)
		if !ok {
			http.Error(w, "not logged in", http.StatusUnauthorized)
			return
		}
		// defaults
		clientID := r.URL.Query().Get("client_id")
		if clientID == "" { clientID = "manual-client" }
		scopes := r.URL.Query().Get("scopes")
		if scopes == "" { scopes = "openid,profile" }
		ttlStr := r.URL.Query().Get("ttl")
		if ttlStr == "" { ttlStr = "24h" }
		ttl, err := time.ParseDuration(ttlStr)
		if err != nil { ttl = 24 * time.Hour }
		// generate token
		buf := make([]byte, 24)
		_, _ = rand.Read(buf)
		token := hex.EncodeToString(buf)
		// persist
		err = s.ids.PersistToken(idsrv.TokenRecord{Token: token, Subject: subject, ClientID: clientID, Scopes: strings.Split(scopes, ","), ExpiresAt: time.Now().Add(ttl)})
		if err != nil {
			log.Error().Str("endpoint", "/dev/token").Str("subject", subject).Err(err).Msg("persist token error")
			http.Error(w, "server error", http.StatusInternalServerError)
			return
		}
		log.Info().Str("endpoint", "/dev/token").Str("subject", subject).Str("client_id", clientID).Dur("ttl", ttl).Msg("created dev token")
		http.Redirect(w, r, "/whoami?dev_token="+token, http.StatusFound)
	})

	// REST API: tokens (list/create/delete) scoped to current subject
	mux.HandleFunc("/api/tokens", func(w http.ResponseWriter, r *http.Request) {
		subject, ok := s.ids.SubjectFromRequest(r)
		if !ok { http.Error(w, "unauthorized", http.StatusUnauthorized); return }
		switch r.Method {
		case http.MethodGet:
			list, err := s.ids.ListTokensBySubject(subject)
			if err != nil { http.Error(w, "server error", http.StatusInternalServerError); return }
			writeJSONWithPreview(w, r, "/api/tokens", list)
		case http.MethodPost:
			var body struct { ClientID string `json:"client_id"`; Scopes []string `json:"scopes"`; TTL string `json:"ttl"` }
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil { http.Error(w, "bad request", http.StatusBadRequest); return }
			if body.ClientID == "" { body.ClientID = "manual-client" }
			if len(body.Scopes) == 0 { body.Scopes = []string{"openid","profile"} }
			ttl := 24 * time.Hour
			if body.TTL != "" { if d, err := time.ParseDuration(body.TTL); err == nil { ttl = d } }
			buf := make([]byte, 24); _, _ = rand.Read(buf); token := hex.EncodeToString(buf)
			if err := s.ids.PersistToken(idsrv.TokenRecord{Token: token, Subject: subject, ClientID: body.ClientID, Scopes: body.Scopes, ExpiresAt: time.Now().Add(ttl)}); err != nil {
				http.Error(w, "server error", http.StatusInternalServerError); return
			}
			writeJSONWithPreview(w, r, "/api/tokens", map[string]any{"token": token})
		case http.MethodDelete:
			tok := r.URL.Query().Get("token")
			if tok == "" { http.Error(w, "token required", http.StatusBadRequest); return }
			// ensure token belongs to subject before delete
			list, _ := s.ids.ListTokensBySubject(subject)
			owned := false
			for _, tr := range list { if tr.Token == tok { owned = true; break } }
			if !owned { http.Error(w, "not found", http.StatusNotFound); return }
			if err := s.ids.DeleteToken(tok); err != nil { http.Error(w, "server error", http.StatusInternalServerError); return }
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// UI: tokens management (Bootstrap + minimal JS)
	mux.HandleFunc("/tokens", func(w http.ResponseWriter, r *http.Request) {
		subject, ok := s.ids.SubjectFromRequest(r)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if !ok { _, _ = w.Write([]byte("<link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css\"><div class=\"container p-4\"><h3>Tokens</h3><div class=\"alert alert-warning\">Please <a href=\"/login\">login</a>.</div></div>")); return }
		html := "<!doctype html><meta charset=\"utf-8\"><link rel=\"stylesheet\" href=\"https://cdn.jsdelivr.net/npm/bootstrap@5.3.3/dist/css/bootstrap.min.css\"><div class=\"container p-4\"><h3>Tokens for " + subject + "</h3><div class=\"mb-3\"><label class=\"form-label\">Client ID</label><input id=\"client_id\" class=\"form-control\" placeholder=\"manual-client\" value=\"manual-client\"></div><div class=\"mb-3\"><label class=\"form-label\">Scopes (comma)</label><input id=\"scopes\" class=\"form-control\" placeholder=\"openid,profile\" value=\"openid,profile\"></div><div class=\"mb-3\"><label class=\"form-label\">TTL</label><input id=\"ttl\" class=\"form-control\" placeholder=\"24h\" value=\"24h\"></div><button id=\"create\" class=\"btn btn-primary\">Create token</button><hr><h5>Existing tokens</h5><table class=\"table table-sm\" id=\"tbl\"><thead><tr><th>Token</th><th>Client</th><th>Scopes</th><th>Expires</th><th></th></tr></thead><tbody></tbody></table><p><a href=\"/whoami\">Back</a></p></div><script>function esc(s){return String(s).replace(/[&<>\"]/g,function(c){return {'&':'&amp;','<':'&lt;','>':'&gt;','\"':'&quot;'}[c]||c;});}async function refresh(){ var res=await fetch('/api/tokens'); var data=await res.json(); var tb=document.querySelector('#tbl tbody'); tb.innerHTML=''; for(var i=0;i<data.length;i++){ var t=data[i]; var tr=document.createElement('tr'); tr.innerHTML = '<td><code>'+esc(t.token)+'</code></td><td>'+esc(t.client_id||t.clientID||'')+'</td><td>'+esc((t.scopes||[]).join(','))+'</td><td>'+esc(t.expires_at||t.expiresAt||'')+'</td><td><button class=\"btn btn-sm btn-danger\" data-token=\"'+esc(t.token)+'\">Delete</button></td>'; tb.appendChild(tr);} }document.addEventListener('click', function(e){ if(e.target && e.target.matches('button[data-token]')){ var tok=e.target.getAttribute('data-token'); fetch('/api/tokens?token='+encodeURIComponent(tok), { method:'DELETE' }).then(function(){ refresh(); }); }});document.getElementById('create').addEventListener('click', function(){ var client_id=(document.getElementById('client_id').value||'manual-client'); var scopes=(document.getElementById('scopes').value||'openid,profile').split(',').map(function(s){return s.trim();}).filter(Boolean); var ttl=(document.getElementById('ttl').value||'24h'); fetch('/api/tokens', { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify({ client_id: client_id, scopes: scopes, ttl: ttl })}).then(function(res){ if(res.ok){ refresh(); } }); });refresh();</script>"
		_, _ = w.Write([]byte(html))
	})
}

// ConfigureAuth allows the caller to configure local user auth and dev-token fallback.
func (s *Server) ConfigureAuth(localUsers bool, sessionTTL time.Duration, devTokenFallbackEnabled bool) {
	s.ids.LocalUsersEnabled = localUsers
	if sessionTTL > 0 {
		s.ids.SessionTTL = sessionTTL
	}
	s.devTokenFallbackEnabled = devTokenFallbackEnabled
}

// User management proxies (Model C)
func (s *Server) CreateUser(username, email, password string) error { return s.ids.CreateUser(username, email, password) }
func (s *Server) DisableUser(username string) error { return s.ids.DisableUser(username) }
func (s *Server) SetPassword(username, password string) error { return s.ids.SetPassword(username, password) }
func (s *Server) ListUsers() ([]idsrv.User, error) { return s.ids.ListUsers() }

func writeJSONWithPreview(w http.ResponseWriter, r *http.Request, endpoint string, v any) {
	b, _ := json.Marshal(v)
	preview := string(b)
	if len(preview) > 500 {
		preview = preview[:500] + "..."
	}
	hdr := map[string]string{
		"accept":          r.Header.Get("Accept"),
		"content_type":    r.Header.Get("Content-Type"),
		"user_agent":      r.Header.Get("User-Agent"),
		"x_forwarded_for": r.Header.Get("X-Forwarded-For"),
	}
	log.Debug().Str("endpoint", endpoint).Fields(hdr).Str("resp_preview", preview).Msg("writing JSON response")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

// LoggingMiddleware logs all requests with duration, status and size
func (s *Server) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		dur := time.Since(start)
		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", rw.status).
			Int("bytes", rw.written).
			Dur("duration", dur).
			Str("ua", r.UserAgent()).
			Str("origin", r.Header.Get("Origin")).
			Str("accept", r.Header.Get("Accept")).
			Str("host", r.Host).
			Str("remote", r.RemoteAddr).
			Str("referer", r.Referer()).
			Msg("request")
	})
}

type responseRecorder struct {
	http.ResponseWriter
	status  int
	written int
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += n
	return n, err
}

// Minimal in-memory corpus for search/fetch demo
type doc struct{
	id string
	title string
	url string
	text string
}

var sampleDocs = []doc{
	{ id: "1", title: "Welcome", url: "https://example.com/welcome", text: "Welcome to the MCP demo corpus." },
	{ id: "2", title: "OIDC Notes", url: "https://example.com/oidc", text: "OIDC Authorization Code with PKCE example." },
}

// Bearer auth + audience enforcement, with RFC 9728 advertisement on 401
func (s *Server) mcpAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// small preview for debugging
		var bodyPreview string
		if r.Body != nil {
			b, _ := io.ReadAll(io.LimitReader(r.Body, 2048))
			bodyPreview = string(b)
			r.Body = io.NopCloser(strings.NewReader(bodyPreview))
		}
		authz := r.Header.Get("Authorization")
		if !strings.HasPrefix(authz, "Bearer ") {
			asMeta := s.issuer + "/.well-known/oauth-authorization-server"
			prm := s.issuer + "/.well-known/oauth-protected-resource"
			hdr := "Bearer realm=\"mcp\", resource=\"" + s.issuer + "/mcp\"" + ", authorization_uri=\"" + asMeta + "\", resource_metadata=\"" + prm + "\""
			w.Header().Set("WWW-Authenticate", hdr)
			log.Warn().Str("endpoint", "mcp").Str("reason", "missing bearer").Str("www_authenticate", hdr).Str("bodyPreview", bodyPreview).Msg("unauthorized")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		raw := strings.TrimPrefix(authz, "Bearer ")
		sess := new(openid.DefaultSession)
		// Introspect opaque access token
		tt, ar, err := s.ids.ProviderRef().IntrospectToken(r.Context(), raw, fosite.AccessToken, sess)
		if err != nil {
			// Dev fallback: accept manual tokens stored in DB
			if s.devTokenFallbackEnabled {
			if tr, ok, derr := s.ids.GetToken(raw); derr == nil && ok {
				if time.Now().Before(tr.ExpiresAt) {
					log.Warn().Str("endpoint", "mcp").Str("reason", "using dev token fallback").Str("subject", tr.Subject).Str("client_id", tr.ClientID).Time("expires_at", tr.ExpiresAt).Msg("authorized via DB token")
					ctx := setAuthCtx(r.Context(), tr.Subject, tr.ClientID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				log.Warn().Str("endpoint", "mcp").Str("reason", "dev token expired").Time("expired_at", tr.ExpiresAt).Msg("unauthorized")
			}
			}
			log.Warn().Str("endpoint", "mcp").Str("reason", "introspection failed").Err(err).Msg("unauthorized")
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		_ = tt
		// Trace subject and client_id
		subject := ""
		if sess.Claims != nil { subject = sess.Claims.Subject }
		clientID := ""
		if ar != nil && ar.GetClient() != nil { clientID = ar.GetClient().GetID() }
		log.Debug().Str("endpoint", "mcp").Str("subject", subject).Str("client_id", clientID).Msg("authorized request")
		ctx := setAuthCtx(r.Context(), subject, clientID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// JSON-RPC request/response types (minimal)
type rpcRequest struct{
	JSONRPC string `json:"jsonrpc"`
	ID any `json:"id"`
	Method string `json:"method"`
	Params json.RawMessage `json:"params"`
}

type rpcResponse struct{
	JSONRPC string `json:"jsonrpc"`
	ID any `json:"id"`
	Result any `json:"result,omitempty"`
	Error *rpcError `json:"error,omitempty"`
}

type rpcError struct{
	Code int `json:"code"`
	Message string `json:"message"`
	Data any `json:"data,omitempty"`
}

func (s *Server) handleMCP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// Read full body for debug logging
	bodyBytes, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	var req rpcRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Error().Str("endpoint", "mcp").Err(err).Msg("failed to decode JSON-RPC request")
		writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: nil, Error: &rpcError{Code: -32700, Message: "parse error"}}, "parse-error")
		return
	}
	log.Info().Str("endpoint", "mcp").Any("id", req.ID).Str("method", req.Method).Dur("since", time.Since(start)).Msg("received JSON-RPC")
	log.Debug().Str("endpoint", "mcp").Str("method", req.Method).Any("id", req.ID).RawJSON("request_body", bodyBytes).Msg("jsonrpc request body")
	switch req.Method {
	case "initialize":
		res := map[string]any{
			"protocolVersion": "2025-03-26",
			"serverInfo": map[string]any{"name": "go-mcp", "version": "0.1.0"},
			"capabilities": map[string]any{},
		}
		writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: res}, req.Method)
	case "tools/list":
		tools := []map[string]any{
			{
				"name": "search",
				"description": "Search corpus and return candidate items",
				"require_approval": "never",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{"query": map[string]any{"type": "string"}},
					"required": []string{"query"},
				},
			},
			{
				"name": "fetch",
				"description": "Fetch a record by ID",
				"require_approval": "never",
				"inputSchema": map[string]any{
					"type": "object",
					"properties": map[string]any{"id": map[string]any{"type": "string"}},
					"required": []string{"id"},
				},
			},
		}
		log.Info().Str("endpoint", "mcp").Str("method", "tools/list").Int("tools", len(tools)).Msg("returning tools list")
		resp := rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": tools}}
		subj, cid := getAuthCtx(r.Context())
		_ = s.ids.LogMCPCall(idsrv.MCPCallLog{Timestamp: time.Now(), Subject: subj, ClientID: cid, RequestID: toStringID(req.ID), ToolName: "tools/list", ArgsJSON: "{}", ResultJSON: mustJSON(resp.Result), Status: "ok", DurationMs: time.Since(start).Milliseconds()})
		writeRPC(w, resp, req.Method)
	case "tools/call":
		var p struct{ Name string `json:"name"`; Arguments map[string]any `json:"arguments"` }
		if err := json.Unmarshal(req.Params, &p); err != nil { writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32602, Message: "invalid params"}}, req.Method); return }
		log.Debug().Str("endpoint", "mcp").Str("tool", p.Name).Interface("args", p.Arguments).Msg("tools/call")
		subj, cid := getAuthCtx(r.Context())
		argsBytes, _ := json.Marshal(p.Arguments)
		callStart := time.Now()
		switch p.Name {
		case "search":
			q, _ := p.Arguments["query"].(string)
			var items []map[string]any
			for _, d := range sampleDocs {
				if q == "" || strings.Contains(strings.ToLower(d.text), strings.ToLower(q)) || strings.Contains(strings.ToLower(d.title), strings.ToLower(q)) {
					items = append(items, map[string]any{"id": d.id, "title": d.title, "text": d.text, "url": d.url})
				}
			}
			log.Info().Str("endpoint", "mcp").Str("tool", "search").Str("query", q).Int("results", len(items)).Msg("search results")
			resp := rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"content": []any{map[string]any{"type": "application/json", "data": items}}}}
			_ = s.ids.LogMCPCall(idsrv.MCPCallLog{Timestamp: time.Now(), Subject: subj, ClientID: cid, RequestID: toStringID(req.ID), ToolName: "search", ArgsJSON: string(argsBytes), ResultJSON: mustJSON(resp.Result), Status: "ok", DurationMs: time.Since(callStart).Milliseconds()})
			writeRPC(w, resp, req.Method)
		case "fetch":
			id, _ := p.Arguments["id"].(string)
			for _, d := range sampleDocs {
				if d.id == id {
					item := map[string]any{"id": d.id, "title": d.title, "text": d.text, "url": d.url}
					log.Info().Str("endpoint", "mcp").Str("tool", "fetch").Str("id", id).Int("bytes", len(d.text)).Msg("fetch result")
					resp := rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"content": []any{map[string]any{"type": "application/json", "data": item}}}}
					_ = s.ids.LogMCPCall(idsrv.MCPCallLog{Timestamp: time.Now(), Subject: subj, ClientID: cid, RequestID: toStringID(req.ID), ToolName: "fetch", ArgsJSON: string(argsBytes), ResultJSON: mustJSON(resp.Result), Status: "ok", DurationMs: time.Since(callStart).Milliseconds()})
					writeRPC(w, resp, req.Method)
					return
				}
			}
			log.Warn().Str("endpoint", "mcp").Str("tool", "fetch").Str("id", id).Msg("fetch not found")
			_ = s.ids.LogMCPCall(idsrv.MCPCallLog{Timestamp: time.Now(), Subject: subj, ClientID: cid, RequestID: toStringID(req.ID), ToolName: "fetch", ArgsJSON: string(argsBytes), ResultJSON: "", Status: "not_found", DurationMs: time.Since(callStart).Milliseconds()})
			writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32004, Message: "not found"}}, req.Method)
		default:
			_ = s.ids.LogMCPCall(idsrv.MCPCallLog{Timestamp: time.Now(), Subject: subj, ClientID: cid, RequestID: toStringID(req.ID), ToolName: p.Name, ArgsJSON: string(argsBytes), ResultJSON: "", Status: "unknown_tool", DurationMs: time.Since(callStart).Milliseconds()})
			writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32601, Message: "unknown tool"}}, req.Method)
		}
	default:
		writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32601, Message: "method not found"}}, req.Method)
	}
}

func writeRPC(w http.ResponseWriter, resp rpcResponse, method string) {
	// Encode to buffer to log full response body
	b, err := json.Marshal(resp)
	if err == nil {
		log.Debug().Str("endpoint", "mcp").Str("method", method).Any("id", resp.ID).RawJSON("response_body", b).Msg("jsonrpc response body")
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().Str("endpoint", "mcp").Str("method", method).Any("id", resp.ID).Err(err).Msg("failed writing response")
	}
}

// small helpers
func toStringID(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return strconv.FormatInt(int64(t), 10)
	default:
		return ""
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}


