package server

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
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

func (s *Server) Routes(mux *http.ServeMux) {
	// health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("endpoint", "healthz").Str("ua", r.UserAgent()).Msg("health check")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Delegate discovery/AS metadata to idsrv
	mux.HandleFunc("/.well-known/openid-configuration", s.ids.RoutesDiscovery)
	mux.HandleFunc("/.well-known/oauth-authorization-server", s.ids.RoutesASMetadata)

	// OAuth 2.0 Protected Resource Metadata (RFC 9728)
	mux.HandleFunc("/.well-known/oauth-protected-resource", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("endpoint", "oauth-protected-resource").Msg("serving protected resource metadata")
		j := map[string]any{
			"authorization_servers": []string{s.issuer},
			"resource":              s.issuer + "/mcp",
		}
		writeJSONWithPreview(w, r, "oauth-protected-resource", j)
		log.Info().Str("endpoint", "oauth-protected-resource").Msg("served protected resource metadata")
	})

	mux.HandleFunc("/oauth2/auth", s.ids.Authorize)
	mux.HandleFunc("/oauth2/token", s.ids.Token)
	mux.HandleFunc("/register", s.ids.Register)
	mux.HandleFunc("/login", s.ids.Login)

	// JWKS
	mux.HandleFunc("/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("endpoint", "jwks").Str("ua", r.UserAgent()).Msg("serving jwks")
		pub := &s.ids.PrivateKey.PublicKey
		j := map[string]any{
			"keys": []map[string]any{
				{
					"kty": "RSA", "alg": "RS256", "use": "sig", "kid": "1",
					"n": base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
					"e": base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
				},
			},
		}
		writeJSONWithPreview(w, r, "jwks", j)
		log.Info().Str("endpoint", "jwks").Msg("served jwks")
	})

	// MCP endpoint: for now, advertise authorization via 401 + WWW-Authenticate per RFC 9728
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
			w.Header().Set("WWW-Authenticate", "Bearer realm=\"mcp\", resource=\""+s.issuer+"/mcp\""+
				", authorization_uri=\""+asMeta+"\", resource_metadata=\""+prm+"\"")
			log.Warn().Str("endpoint", "mcp").Str("method", r.Method).Str("origin", r.Header.Get("Origin")).Str("contentType", r.Header.Get("Content-Type")).Str("bodyPreview", bodyPreview).Msg("unauthorized - advertising AS metadata")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// No token validation yet; still not implemented
		log.Warn().Str("endpoint", "mcp").Str("method", r.Method).Str("origin", r.Header.Get("Origin")).Str("contentType", r.Header.Get("Content-Type")).Str("bodyPreview", bodyPreview).Msg("received MCP request - not implemented")
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


