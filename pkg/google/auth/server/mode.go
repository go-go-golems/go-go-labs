package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
)

// ServerMode determines how the OAuth callback is handled
type ServerMode interface {
	// GetCallbackURL returns the full callback URL
	GetCallbackURL() string

	// Setup prepares the callback handler
	Setup(ctx context.Context, handler http.HandlerFunc) error

	// Cleanup performs any necessary cleanup
	Cleanup(ctx context.Context) error
}

// StandaloneServer implements ServerMode with its own HTTP server
type StandaloneServer struct {
	port     int
	path     string
	server   *http.Server
	listener net.Listener
}

// NewStandaloneServer creates a standalone server mode
func NewStandaloneServer(port int, path string) *StandaloneServer {
	return &StandaloneServer{
		port: port,
		path: path,
	}
}

func (s *StandaloneServer) GetCallbackURL() string {
	return fmt.Sprintf("http://localhost:%d%s", s.port, s.path)
}

func (s *StandaloneServer) Setup(ctx context.Context, handler http.HandlerFunc) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.path, handler)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Create listener before starting server to ensure port is available
	listener, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		return fmt.Errorf("failed to create listener: %w", err)
	}
	s.listener = listener

	errChan := make(chan error, 1)
	go func() {
		if err := s.server.Serve(listener); err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
		close(errChan)
	}()

	// Wait for either context cancellation or server error
	select {
	case err := <-errChan:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	default:
		// Continue if no immediate error
	}

	return nil
}

func (s *StandaloneServer) Cleanup(ctx context.Context) error {
	if s.server != nil {
		// First attempt to close the listener to prevent new connections
		if s.listener != nil {
			if err := s.listener.Close(); err != nil {
				return fmt.Errorf("failed to close listener: %w", err)
			}
		}

		// Then shutdown the server with the provided context
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown server: %w", err)
		}
	}
	return nil
}

// MuxHandler implements ServerMode using an existing http.ServeMux
type MuxHandler struct {
	mux     *http.ServeMux
	path    string
	baseURL string
}

// NewMuxHandler creates a mux handler mode
func NewMuxHandler(mux *http.ServeMux, path string, baseURL string) *MuxHandler {
	return &MuxHandler{
		mux:     mux,
		path:    path,
		baseURL: baseURL,
	}
}

func (h *MuxHandler) GetCallbackURL() string {
	return h.baseURL + h.path
}

func (h *MuxHandler) Setup(ctx context.Context, handler http.HandlerFunc) error {
	h.mux.HandleFunc(h.path, handler)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func (h *MuxHandler) Cleanup(ctx context.Context) error {
	// No cleanup needed for mux handler
	return nil
}

// CustomServer implements ServerMode with a custom setup function
type CustomServer struct {
	callbackURL string
	setup       func(context.Context, http.HandlerFunc) error
}

// NewCustomServer creates a custom server mode
func NewCustomServer(callbackURL string, setup func(context.Context, http.HandlerFunc) error) *CustomServer {
	return &CustomServer{
		callbackURL: callbackURL,
		setup:       setup,
	}
}

func (s *CustomServer) GetCallbackURL() string {
	return s.callbackURL
}

func (s *CustomServer) Setup(ctx context.Context, handler http.HandlerFunc) error {
	return s.setup(ctx, handler)
}

func (s *CustomServer) Cleanup(ctx context.Context) error {
	// Custom server must handle its own cleanup
	return nil
}
