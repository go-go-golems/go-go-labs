package auth

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/wesen/friday-talks/internal/models"
)

const (
	// TokenCookieName is the name of the cookie used to store the auth token
	TokenCookieName = "friday_talks_auth_token"

	// TokenExpiry is the duration after which a token will expire
	TokenExpiry = 24 * time.Hour

	// ContextUserKey is the key used to store user info in request context
	ContextUserKey = "user"
)

// Auth handles authentication and authorization
type Auth struct {
	jwtSecret     []byte
	userRepo      models.UserRepository
	tokenDuration time.Duration
}

// Claims represents JWT token claims
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// NewAuth creates a new Auth instance
func NewAuth(jwtSecret string, userRepo models.UserRepository) *Auth {
	return &Auth{
		jwtSecret:     []byte(jwtSecret),
		userRepo:      userRepo,
		tokenDuration: TokenExpiry,
	}
}

// Authenticate checks if the provided credentials are valid
func (a *Auth) Authenticate(ctx context.Context, email, password string) (*models.User, error) {
	user, err := a.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.Wrap(err, "failed to find user by email")
	}

	if !models.CheckPassword(password, user.PasswordHash) {
		return nil, errors.New("invalid password")
	}

	return user, nil
}

// GenerateToken creates a new JWT token for the given user
func (a *Auth) GenerateToken(userID int) (string, error) {
	expiresAt := time.Now().Add(a.tokenDuration)

	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", errors.Wrap(err, "failed to sign token")
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the user ID if valid
func (a *Auth) ValidateToken(tokenString string) (int, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return a.jwtSecret, nil
	})

	if err != nil {
		return 0, errors.Wrap(err, "failed to parse token")
	}

	if !token.Valid {
		return 0, errors.New("invalid token")
	}

	return claims.UserID, nil
}

// SetTokenCookie sets the authentication token as a cookie
func (a *Auth) SetTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     TokenCookieName,
		Value:    token,
		Expires:  time.Now().Add(a.tokenDuration),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set to true in production with HTTPS
	})
}

// ClearTokenCookie clears the authentication token cookie
func (a *Auth) ClearTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     TokenCookieName,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
		Secure:   false, // Set to true in production with HTTPS
	})
}

// UserFromContext extracts the user from the request context
func UserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value(ContextUserKey).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// AuthMiddleware creates middleware that validates JWT tokens and adds user to context
func (a *Auth) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from cookie
		cookie, err := r.Cookie(TokenCookieName)
		if err != nil {
			// No token, proceed without authentication
			next.ServeHTTP(w, r)
			return
		}

		// Validate token
		userID, err := a.ValidateToken(cookie.Value)
		if err != nil {
			// Invalid token, clear cookie and proceed without authentication
			a.ClearTokenCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		// Get user from database
		user, err := a.userRepo.FindByID(r.Context(), userID)
		if err != nil {
			// User not found, clear cookie and proceed without authentication
			a.ClearTokenCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), ContextUserKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth creates middleware that requires authentication
func (a *Auth) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
