package storage

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in our system
type User struct {
	Username     string
	PasswordHash string
	Subject      string // OIDC subject identifier
}

// MemoryStore implements Fosite's storage interfaces in memory
type MemoryStore struct {
	mu sync.RWMutex

	// Core storage
	Clients       map[string]fosite.Client
	Users         map[string]*User
	AuthCodes     map[string]fosite.Requester
	AccessTokens  map[string]fosite.Requester
	RefreshTokens map[string]fosite.Requester
	IDTokens      map[string]fosite.Requester
	PKCEs         map[string]fosite.Requester

	// Session storage for OIDC
	AuthCodeSessions     map[string]fosite.Session
	AccessTokenSessions  map[string]fosite.Session
	RefreshTokenSessions map[string]fosite.Session
	IDTokenSessions      map[string]fosite.Session
}

// NewMemoryStore creates a new in-memory storage with default user
func NewMemoryStore() *MemoryStore {
	// Hash the default password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret"), bcrypt.DefaultCost)

	store := &MemoryStore{
		Clients:              make(map[string]fosite.Client),
		Users:                make(map[string]*User),
		AuthCodes:            make(map[string]fosite.Requester),
		AccessTokens:         make(map[string]fosite.Requester),
		RefreshTokens:        make(map[string]fosite.Requester),
		IDTokens:             make(map[string]fosite.Requester),
		PKCEs:                make(map[string]fosite.Requester),
		AuthCodeSessions:     make(map[string]fosite.Session),
		AccessTokenSessions:  make(map[string]fosite.Session),
		RefreshTokenSessions: make(map[string]fosite.Session),
		IDTokenSessions:      make(map[string]fosite.Session),
	}

	// Add default user
	store.Users["wesen"] = &User{
		Username:     "wesen",
		PasswordHash: string(hashedPassword),
		Subject:      "wesen-user-id",
	}

	return store
}

// AuthenticateUser verifies username/password credentials
func (s *MemoryStore) AuthenticateUser(username, password string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.Users[username]
	if !exists {
		return nil, errors.New("user not found")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// CreateClient stores a new OAuth2 client
func (s *MemoryStore) CreateClient(client fosite.Client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Clients[client.GetID()] = client
	return nil
}

// GenerateClientID creates a new random client ID
func (s *MemoryStore) GenerateClientID() string {
	return generateRandomID("client_")
}

// GenerateClientSecret creates a new random client secret
func (s *MemoryStore) GenerateClientSecret() string {
	return generateRandomID("secret_")
}

// generateRandomID creates a random ID with prefix
func generateRandomID(prefix string) string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return prefix + base64.URLEncoding.EncodeToString(bytes)[:22] // Remove padding
}

// Fosite ClientManager interface
func (s *MemoryStore) GetClient(_ context.Context, id string) (fosite.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, ok := s.Clients[id]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return client, nil
}

// Fosite AuthorizeCodeSession storage
func (s *MemoryStore) CreateAuthorizeCodeSession(_ context.Context, code string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.AuthCodes[code] = req
	s.AuthCodeSessions[code] = req.GetSession()
	return nil
}

func (s *MemoryStore) GetAuthorizeCodeSession(_ context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.AuthCodes[code]
	if !ok {
		return nil, fosite.ErrNotFound
	}

	// Copy session data
	if storedSession, exists := s.AuthCodeSessions[code]; exists {
		if defaultSession, ok := session.(*openid.DefaultSession); ok {
			if storedDefault, ok := storedSession.(*openid.DefaultSession); ok {
				*defaultSession = *storedDefault
			}
		}
	}

	return req, nil
}

func (s *MemoryStore) DeleteAuthorizeCodeSession(_ context.Context, code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.AuthCodes, code)
	delete(s.AuthCodeSessions, code)
	return nil
}

// Fosite AccessTokenSession storage
func (s *MemoryStore) CreateAccessTokenSession(_ context.Context, signature string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.AccessTokens[signature] = req
	s.AccessTokenSessions[signature] = req.GetSession()
	return nil
}

func (s *MemoryStore) GetAccessTokenSession(_ context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.AccessTokens[signature]
	if !ok {
		return nil, fosite.ErrNotFound
	}

	// Copy session data
	if storedSession, exists := s.AccessTokenSessions[signature]; exists {
		if defaultSession, ok := session.(*openid.DefaultSession); ok {
			if storedDefault, ok := storedSession.(*openid.DefaultSession); ok {
				*defaultSession = *storedDefault
			}
		}
	}

	return req, nil
}

func (s *MemoryStore) DeleteAccessTokenSession(_ context.Context, signature string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.AccessTokens, signature)
	delete(s.AccessTokenSessions, signature)
	return nil
}

// Fosite RefreshTokenSession storage
func (s *MemoryStore) CreateRefreshTokenSession(_ context.Context, signature string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.RefreshTokens[signature] = req
	s.RefreshTokenSessions[signature] = req.GetSession()
	return nil
}

func (s *MemoryStore) GetRefreshTokenSession(_ context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.RefreshTokens[signature]
	if !ok {
		return nil, fosite.ErrNotFound
	}

	// Copy session data
	if storedSession, exists := s.RefreshTokenSessions[signature]; exists {
		if defaultSession, ok := session.(*openid.DefaultSession); ok {
			if storedDefault, ok := storedSession.(*openid.DefaultSession); ok {
				*defaultSession = *storedDefault
			}
		}
	}

	return req, nil
}

func (s *MemoryStore) DeleteRefreshTokenSession(_ context.Context, signature string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.RefreshTokens, signature)
	delete(s.RefreshTokenSessions, signature)
	return nil
}

// Fosite PKCE storage
func (s *MemoryStore) CreatePKCERequestSession(_ context.Context, signature string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.PKCEs[signature] = req
	return nil
}

func (s *MemoryStore) GetPKCERequestSession(_ context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.PKCEs[signature]
	if !ok {
		return nil, fosite.ErrNotFound
	}
	return req, nil
}

func (s *MemoryStore) DeletePKCERequestSession(_ context.Context, signature string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.PKCEs, signature)
	return nil
}

// OIDC ID Token session storage
func (s *MemoryStore) CreateOpenIDConnectSession(_ context.Context, authorizeCode string, req fosite.Requester) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.IDTokens[authorizeCode] = req
	s.IDTokenSessions[authorizeCode] = req.GetSession()
	return nil
}

func (s *MemoryStore) GetOpenIDConnectSession(_ context.Context, authorizeCode string, requester fosite.Requester) (fosite.Requester, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	req, ok := s.IDTokens[authorizeCode]
	if !ok {
		return nil, fosite.ErrNotFound
	}

	return req, nil
}

func (s *MemoryStore) DeleteOpenIDConnectSession(_ context.Context, authorizeCode string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.IDTokens, authorizeCode)
	delete(s.IDTokenSessions, authorizeCode)
	return nil
}

// RevokeRefreshToken implements token revocation
func (s *MemoryStore) RevokeRefreshToken(_ context.Context, requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for signature, req := range s.RefreshTokens {
		if req.GetID() == requestID {
			delete(s.RefreshTokens, signature)
			delete(s.RefreshTokenSessions, signature)
		}
	}
	return nil
}

// RevokeAccessToken implements token revocation
func (s *MemoryStore) RevokeAccessToken(_ context.Context, requestID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for signature, req := range s.AccessTokens {
		if req.GetID() == requestID {
			delete(s.AccessTokens, signature)
			delete(s.AccessTokenSessions, signature)
		}
	}
	return nil
}

// ClientAssertionJWTValid validates client assertion JWTs
func (s *MemoryStore) ClientAssertionJWTValid(_ context.Context, jti string) error {
	// For our simple implementation, we don't support client assertions
	return fosite.ErrJTIKnown
}

// SetClientAssertionJWT stores a client assertion JWT
func (s *MemoryStore) SetClientAssertionJWT(_ context.Context, jti string, exp time.Time) error {
	// For our simple implementation, we don't support client assertions
	return nil
}

// Additional required methods for CoreStorage interface compatibility

// InvalidateAuthorizeCodeSession marks an authorization code as used
func (s *MemoryStore) InvalidateAuthorizeCodeSession(_ context.Context, code string) error {
	return s.DeleteAuthorizeCodeSession(context.Background(), code)
}

// RevokeRefreshTokenMaybeGracePeriod revokes a refresh token with potential grace period
func (s *MemoryStore) RevokeRefreshTokenMaybeGracePeriod(_ context.Context, requestID string, signature string) error {
	return s.RevokeRefreshToken(context.Background(), requestID)
}

// RevokeAccessTokenMaybeGracePeriod revokes an access token with potential grace period
func (s *MemoryStore) RevokeAccessTokenMaybeGracePeriod(_ context.Context, requestID string, signature string) error {
	return s.RevokeAccessToken(context.Background(), requestID)
}
