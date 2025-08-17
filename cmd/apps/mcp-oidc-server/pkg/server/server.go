package server

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	idsrv "github.com/go-go-golems/go-go-labs/cmd/apps/mcp-oidc-server/pkg/idsrv"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/rs/zerolog/log"
)

type Server struct {
	issuer string
	ids    *idsrv.Server
}

func New(issuer string) (*Server, error) {
	id, err := idsrv.New(issuer)
	if err != nil {
		return nil, err
	}
	return &Server{issuer: issuer, ids: id}, nil
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
}

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
		_, _, err := s.ids.ProviderRef().IntrospectToken(r.Context(), raw, fosite.AccessToken, sess)
		if err != nil {
			// Dev fallback: accept manual tokens stored in DB
			if tr, ok, derr := s.ids.GetToken(raw); derr == nil && ok {
				if time.Now().Before(tr.ExpiresAt) {
					log.Warn().Str("endpoint", "mcp").Str("reason", "using dev token fallback").Str("subject", tr.Subject).Str("client_id", tr.ClientID).Time("expires_at", tr.ExpiresAt).Msg("authorized via DB token")
					// proceed without Fosite session
					next.ServeHTTP(w, r)
					return
				}
				log.Warn().Str("endpoint", "mcp").Str("reason", "dev token expired").Time("expired_at", tr.ExpiresAt).Msg("unauthorized")
			}
			log.Warn().Str("endpoint", "mcp").Str("reason", "introspection failed").Err(err).Msg("unauthorized")
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		// Optionally enforce audience on access tokens if set
		// Log subject for traceability
		subject := ""
		if sess.Claims != nil { subject = sess.Claims.Subject }
		log.Debug().Str("endpoint", "mcp").Str("subject", subject).Msg("authorized request")
		next.ServeHTTP(w, r)
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
		writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"tools": tools}}, req.Method)
	case "tools/call":
		var p struct{ Name string `json:"name"`; Arguments map[string]any `json:"arguments"` }
		if err := json.Unmarshal(req.Params, &p); err != nil { writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32602, Message: "invalid params"}}, req.Method); return }
		log.Debug().Str("endpoint", "mcp").Str("tool", p.Name).Interface("args", p.Arguments).Msg("tools/call")
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
			writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"content": []any{map[string]any{"type": "application/json", "data": items}}}}, req.Method)
		case "fetch":
			id, _ := p.Arguments["id"].(string)
			for _, d := range sampleDocs {
				if d.id == id {
					item := map[string]any{"id": d.id, "title": d.title, "text": d.text, "url": d.url}
					log.Info().Str("endpoint", "mcp").Str("tool", "fetch").Str("id", id).Int("bytes", len(d.text)).Msg("fetch result")
					writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{"content": []any{map[string]any{"type": "application/json", "data": item}}}}, req.Method)
					return
				}
			}
			log.Warn().Str("endpoint", "mcp").Str("tool", "fetch").Str("id", id).Msg("fetch not found")
			writeRPC(w, rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32004, Message: "not found"}}, req.Method)
		default:
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


