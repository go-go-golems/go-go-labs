package idsrv

import "net/http"

// Small exported adapters so other packages can mount these easily
func (s *Server) RoutesDiscovery(w http.ResponseWriter, r *http.Request) { s.oidcDiscovery(w, r) }
func (s *Server) RoutesASMetadata(w http.ResponseWriter, r *http.Request) { s.asMetadata(w, r) }
func (s *Server) Authorize(w http.ResponseWriter, r *http.Request)      { s.authorize(w, r) }
func (s *Server) Token(w http.ResponseWriter, r *http.Request)          { s.token(w, r) }
func (s *Server) Register(w http.ResponseWriter, r *http.Request)       { s.register(w, r) }
func (s *Server) Login(w http.ResponseWriter, r *http.Request)          { s.login(w, r) }

// Configuration setters for Model C
func (s *Server) SetLocalUsersEnabled(enabled bool) { s.LocalUsersEnabled = enabled }
func (s *Server) SetSessionTTL(ttlSeconds int64) {
    if ttlSeconds <= 0 { s.SessionTTL = 12 * 3600 * 1e9 /* overwritten below */ }
}

// SubjectFromRequest returns the subject from the current request cookie/session if logged in.
func (s *Server) SubjectFromRequest(r *http.Request) (string, bool) { return s.lookupSessionSubject(r) }

