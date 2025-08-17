package idsrv

import (
    "crypto/rand"
    "crypto/rsa"
    "encoding/base64"
    "encoding/json"
    "html/template"
    "math/big"
    "net/http"
    "time"

    "github.com/rs/zerolog/log"
)

type Server struct {
    PrivateKey *rsa.PrivateKey
    Issuer     string
}

func New(issuer string) (*Server, error) {
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil { return nil, err }
    return &Server{PrivateKey: privateKey, Issuer: issuer}, nil
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
        "code_challenge_methods_supported":       []string{"S256"},
        "registration_endpoint":                  s.Issuer + "/register",
    }
    writeJSON(w, j)
}

func (s *Server) asMetadata(w http.ResponseWriter, r *http.Request) {
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
        _ = loginTpl.Execute(w, struct{ ReturnTo string }{r.URL.Query().Get("return_to")})
    case http.MethodPost:
        http.Redirect(w, r, r.FormValue("return_to"), http.StatusFound)
    default:
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
    }
}

func (s *Server) authorize(w http.ResponseWriter, r *http.Request) {
    log.Info().Str("endpoint", "oauth2/auth").RawJSON("query", []byte("{}")).Msg("authorize request (stub)")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotImplemented)
    _ = json.NewEncoder(w).Encode(map[string]any{"error": "authorize endpoint not implemented yet"})
}

func (s *Server) token(w http.ResponseWriter, r *http.Request) {
    _ = r.ParseForm()
    log.Info().Str("endpoint", "oauth2/token").RawJSON("form", []byte("{}")).Msg("token request (stub)")
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusNotImplemented)
    _ = json.NewEncoder(w).Encode(map[string]any{"error": "token endpoint not implemented yet"})
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
    var payload map[string]any
    _ = json.NewDecoder(r.Body).Decode(&payload)
    id := "dev-client-" + time.Now().Format("150405")
    log.Info().Str("endpoint", "register").Str("client_id", id).Interface("payload", payload).Msg("dynamic registration (stub)")
    resp := map[string]any{
        "client_id":                  id,
        "redirect_uris":              payload["redirect_uris"],
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


