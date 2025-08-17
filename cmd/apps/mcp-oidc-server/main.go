package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type server struct {
	issuer     string
	privateKey *rsa.PrivateKey
}

func newServer(issuer string) (*server, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	return &server{issuer: issuer, privateKey: key}, nil
}

func (s *server) routes(mux *http.ServeMux) {
	// health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		zlog.Debug().Str("endpoint", "healthz").Str("ua", r.UserAgent()).Msg("health check")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// OIDC Discovery
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		zlog.Debug().Str("endpoint", "openid-configuration").Str("ua", r.UserAgent()).Str("qs", r.URL.RawQuery).Msg("serving discovery")
		j := map[string]any{
			"issuer":                 s.issuer,
			"authorization_endpoint": s.issuer + "/oauth2/auth",
			"token_endpoint":         s.issuer + "/oauth2/token",
			"jwks_uri":               s.issuer + "/jwks.json",
			"scopes_supported":       []string{"openid", "profile", "offline_access"},
			"response_types_supported": []string{"code"},
			"grant_types_supported":    []string{"authorization_code", "refresh_token"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"token_endpoint_auth_methods_supported": []string{"none"},
			"code_challenge_methods_supported":       []string{"S256"},
			"registration_endpoint":                  s.issuer + "/register",
		}
		writeJSONWithPreview(w, r, "openid-configuration", j)
		zlog.Info().Str("endpoint", "openid-configuration").Int("fields", len(j)).Msg("served discovery")
	})

	// OAuth 2.0 Authorization Server Metadata (RFC 8414)
	mux.HandleFunc("/.well-known/oauth-authorization-server", func(w http.ResponseWriter, r *http.Request) {
		zlog.Debug().Str("endpoint", "oauth-authorization-server").Str("ua", r.UserAgent()).Str("qs", r.URL.RawQuery).Msg("serving as-metadata")
		j := map[string]any{
			"issuer":                 s.issuer,
			"authorization_endpoint": s.issuer + "/oauth2/auth",
			"token_endpoint":         s.issuer + "/oauth2/token",
			"jwks_uri":               s.issuer + "/jwks.json",
			"code_challenge_methods_supported": []string{"S256"},
			"response_types_supported":          []string{"code"},
			"grant_types_supported":             []string{"authorization_code", "refresh_token"},
			"scopes_supported":                  []string{"openid", "profile", "offline_access"},
			"token_endpoint_auth_methods_supported": []string{"none"},
			"registration_endpoint":               s.issuer + "/register",
		}
		writeJSONWithPreview(w, r, "oauth-authorization-server", j)
		zlog.Info().Str("endpoint", "oauth-authorization-server").Int("fields", len(j)).Msg("served as-metadata")
	})

	// OAuth 2.0 Protected Resource Metadata (RFC 9728)
	mux.HandleFunc("/.well-known/oauth-protected-resource", func(w http.ResponseWriter, r *http.Request) {
		zlog.Debug().Str("endpoint", "oauth-protected-resource").Msg("serving protected resource metadata")
		j := map[string]any{
			"authorization_servers": []string{s.issuer},
			"resource":              s.issuer + "/mcp",
		}
		writeJSONWithPreview(w, r, "oauth-protected-resource", j)
		zlog.Info().Str("endpoint", "oauth-protected-resource").Msg("served protected resource metadata")
	})

	// Stub: OAuth2 Authorization Endpoint (logs query params, not implemented yet)
	mux.HandleFunc("/oauth2/auth", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		zlog.Info().
			Str("endpoint", "oauth2/auth").
			Str("client_id", q.Get("client_id")).
			Str("redirect_uri", q.Get("redirect_uri")).
			Str("response_type", q.Get("response_type")).
			Str("scope", q.Get("scope")).
			Str("state", q.Get("state")).
			Str("code_challenge", q.Get("code_challenge")).
			Str("code_challenge_method", q.Get("code_challenge_method")).
			Str("resource", q.Get("resource")).
			Msg("authorize request (stub)")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "authorize endpoint not implemented yet"})
	})

	// Stub: OAuth2 Token Endpoint (logs form fields, not implemented yet)
	mux.HandleFunc("/oauth2/token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		zlog.Info().
			Str("endpoint", "oauth2/token").
			Str("grant_type", r.Form.Get("grant_type")).
			Str("client_id", r.Form.Get("client_id")).
			Str("code", r.Form.Get("code")).
			Str("redirect_uri", r.Form.Get("redirect_uri")).
			Str("code_verifier", r.Form.Get("code_verifier")).
			Str("resource", r.Form.Get("resource")).
			Msg("token request (stub)")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "token endpoint not implemented yet"})
	})

	// Stub: Dynamic Client Registration endpoint
	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		_ = json.NewDecoder(r.Body).Decode(&payload)
		uris, _ := payload["redirect_uris"].([]any)
		zlog.Info().Str("endpoint", "register").Interface("payload", payload).Interface("redirect_uris", uris).Msg("dynamic client registration (stub)")
		resp := map[string]any{
			"client_id":                           "dev-client-" + time.Now().Format("150405"),
			"redirect_uris":                        payload["redirect_uris"],
			"token_endpoint_auth_method":           "none",
			"grant_types":                          []string{"authorization_code", "refresh_token"},
			"response_types":                       []string{"code"},
			"code_challenge_methods_supported":      []string{"S256"},
		}
		writeJSONWithPreview(w, r, "register", resp)
	})

	// JWKS
	mux.HandleFunc("/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		zlog.Debug().Str("endpoint", "jwks").Str("ua", r.UserAgent()).Msg("serving jwks")
		pub := &s.privateKey.PublicKey
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
		zlog.Info().Str("endpoint", "jwks").Msg("served jwks")
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
			zlog.Warn().Str("endpoint", "mcp").Str("method", r.Method).Str("origin", r.Header.Get("Origin")).Str("contentType", r.Header.Get("Content-Type")).Str("bodyPreview", bodyPreview).Msg("unauthorized - advertising AS metadata")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// No token validation yet; still not implemented
		zlog.Warn().Str("endpoint", "mcp").Str("method", r.Method).Str("origin", r.Header.Get("Origin")).Str("contentType", r.Header.Get("Content-Type")).Str("bodyPreview", bodyPreview).Msg("received MCP request - not implemented")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": "MCP not implemented yet"})
	})
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}

// writeJSONWithPreview writes JSON and logs key request headers and a short response preview
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
	zlog.Debug().Str("endpoint", endpoint).Fields(hdr).Str("resp_preview", preview).Msg("writing JSON response")
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(b)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	var (
		addr      string
		issuer    string
		logFormat string
		logLevel  string
	)

	rootCmd := &cobra.Command{
		Use:   "mcp-oidc-server",
		Short: "MCP + OIDC discovery stub server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Logging setup
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
			switch logFormat {
			case "json":
				// default JSON
			default:
				zlog.Logger = zlog.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
			}
			switch logLevel {
			case "trace":
				zerolog.SetGlobalLevel(zerolog.TraceLevel)
			case "debug":
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			case "info":
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			case "warn":
				zerolog.SetGlobalLevel(zerolog.WarnLevel)
			case "error":
				zerolog.SetGlobalLevel(zerolog.ErrorLevel)
			default:
				zerolog.SetGlobalLevel(zerolog.InfoLevel)
			}

			// Create server
			s, err := newServer(issuer)
			if err != nil {
				zlog.Fatal().Err(err).Msg("failed generating RSA key")
			}

			mux := http.NewServeMux()
			s.routes(mux)

			// Wrap the entire mux to ensure ALL requests (including 404) are logged
			wrapped := s.loggingMiddleware(mux)
			zlog.Info().Str("addr", addr).Str("issuer", issuer).Msg("mcp-oidc-server listening")
			if err := http.ListenAndServe(addr, wrapped); err != nil {
				zlog.Fatal().Err(err).Msg("server exited")
			}
			return nil
		},
	}

	rootCmd.Flags().StringVar(&addr, "addr", getenv("ADDR", ":8080"), "HTTP listen address")
	rootCmd.Flags().StringVar(&issuer, "issuer", getenv("ISSUER", "http://localhost:8080"), "Issuer/base URL")
	rootCmd.Flags().StringVar(&logFormat, "log-format", getenv("LOG_FORMAT", "console"), "Log format: console|json")
	rootCmd.Flags().StringVar(&logLevel, "log-level", getenv("LOG_LEVEL", "info"), "Log level: trace|debug|info|warn|error")

	if err := rootCmd.Execute(); err != nil {
		zlog.Fatal().Err(err).Msg("command error")
	}
}

// loggingMiddleware logs all requests with duration, status and size
func (s *server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		dur := time.Since(start)
		zlog.Info().
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


