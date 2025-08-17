package idsrv

import "net/http"

// Small exported adapters so other packages can mount these easily
func (s *Server) RoutesDiscovery(w http.ResponseWriter, r *http.Request) { s.oidcDiscovery(w, r) }
func (s *Server) RoutesASMetadata(w http.ResponseWriter, r *http.Request) { s.asMetadata(w, r) }
func (s *Server) Authorize(w http.ResponseWriter, r *http.Request)      { s.authorize(w, r) }
func (s *Server) Token(w http.ResponseWriter, r *http.Request)          { s.token(w, r) }
func (s *Server) Register(w http.ResponseWriter, r *http.Request)       { s.register(w, r) }
func (s *Server) Login(w http.ResponseWriter, r *http.Request)          { s.login(w, r) }

