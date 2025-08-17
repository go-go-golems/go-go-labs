package server

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	idsrv "github.com/go-go-golems/go-go-labs/cmd/apps/mcp-oidc-server/pkg/idsrv"
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

	// MCP endpoint: advertise authorization via 401 + WWW-Authenticate (intermediate step)
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		// Capture small body for debug visibility
		var bodyPreview string
		if r.Body != nil {
			b, _ := io.ReadAll(io.LimitReader(r.Body, 2048))
			bodyPreview = string(b)
		}
		authz := r.Header.Get("Authorization")
		if authz == "" {
			// RFC 9728 Section 5.1: advertise resource metadata via WWW-Authenticate on 401
			asMeta := s.issuer + "/.well-known/oauth-authorization-server"
			prm := s.issuer + "/.well-known/oauth-protected-resource"
			hdr := "Bearer realm=\"mcp\", resource=\"" + s.issuer + "/mcp\"" +
				", authorization_uri=\"" + asMeta + "\", resource_metadata=\"" + prm + "\""
			w.Header().Set("WWW-Authenticate", hdr)
			log.Warn().
				Str("endpoint", "mcp").
				Str("method", r.Method).
				Str("origin", r.Header.Get("Origin")).
				Str("contentType", r.Header.Get("Content-Type")).
				Str("ua", r.UserAgent()).
				Str("www_authenticate", hdr).
				Str("bodyPreview", bodyPreview).
				Msg("unauthorized - advertising AS metadata")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "MCP not implemented yet"})
	})
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


