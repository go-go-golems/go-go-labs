package types

import (
	"context"
	"io"
	"net/http"
)

// Core MCP Types
type SearchRequest struct {
	Query   string            `json:"query"`
	Context map[string]string `json:"context,omitempty"`
}

type SearchResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"` // Brief text snippet
	URL   string `json:"url"`
}

type FetchRequest struct {
	ID string `json:"id"`
}

type FetchResult struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Text     string                 `json:"text"` // Full content
	URL      string                 `json:"url"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// MCP Server Interface - allows parallel development
type MCPServer interface {
	// Core capabilities
	RegisterSearch(handler SearchHandler) error
	RegisterFetch(handler FetchHandler) error

	// Server lifecycle
	Start(ctx context.Context) error
	Stop() error

	// Transport integration
	GetHTTPHandler() http.Handler
}

// Handler types for MCP operations
type SearchHandler func(ctx context.Context, req SearchRequest) (<-chan SearchResult, error)
type FetchHandler func(ctx context.Context, req FetchRequest) (FetchResult, error)

// Auth Interface - allows parallel OAuth development
type AuthValidator interface {
	ValidateToken(ctx context.Context, token string) (*UserInfo, error)
	GetAuthEndpoints() AuthEndpoints
}

type UserInfo struct {
	ID       string `json:"id"`
	Login    string `json:"login"`
	Email    string `json:"email,omitempty"`
	Verified bool   `json:"verified"`
}

type AuthEndpoints struct {
	AuthorizeURL string   `json:"authorization_url"`
	TokenURL     string   `json:"token_url"`
	Scopes       []string `json:"scopes"`
}

// Transport Interface - allows parallel SSE development
type Transport interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	SetAuthValidator(validator AuthValidator)
	SetMCPServer(server MCPServer)
}

// Discovery Interface - allows parallel well-known endpoint development
type DiscoveryService interface {
	GetPluginManifest() ([]byte, error)
	GetOAuthConfig() ([]byte, error)
	GetPluginManifestHandler() http.HandlerFunc
	GetOAuthConfigHandler() http.HandlerFunc
}

// Config represents the application configuration
type Config struct {
	// OIDC Configuration
	OIDCIssuer      string `json:"oidc_issuer"`
	DefaultUser     string `json:"default_user"`
	DefaultPassword string `json:"default_password"`

	// Server Configuration
	Host     string `json:"host"`
	Port     int    `json:"port"`
	LogLevel string `json:"log_level"`
}

// Message represents JSON-RPC messages
type JSONRPCMessage struct {
	JSONRPC string      `json:"jsonrpc"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SSE Event represents Server-Sent Events
type SSEEvent struct {
	Event string `json:"event,omitempty"`
	Data  string `json:"data"`
	ID    string `json:"id,omitempty"`
}

// Stream represents a closeable stream
type Stream interface {
	io.Closer
	Send(event SSEEvent) error
}
