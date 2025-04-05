# How the Fuck Does JWT Work?

JWT (JSON Web Token) is a compact, URL-safe means of representing claims between two parties. This document explains how JWT authentication works in our Friday Talks application.

## JWT Basics

A JWT consists of three parts separated by dots (`.`):
- **Header**: Contains metadata about the token (type, algorithm)
- **Payload**: Contains claims (statements about an entity, typically the user)
- **Signature**: Ensures the token hasn't been altered

The format looks like: `xxxxx.yyyyy.zzzzz`

## Implementation in Friday Talks

Our authentication system is implemented in `auth.go` and provides:

1. User authentication with email/password
2. JWT token generation and validation
3. Middleware for protected routes

## Key Components

### Auth Structure

```go
type Auth struct {
    jwtSecret     []byte
    userRepo      models.UserRepository
    tokenDuration time.Duration
}
```

The `Auth` struct holds:
- JWT secret key for signing tokens
- User repository for accessing user data
- Token duration (24 hours by default)

### JWT Claims

```go
type Claims struct {
    UserID int `json:"user_id"`
    jwt.RegisteredClaims
}
```

Our tokens store:
- User ID
- Standard JWT claims (expiration time, issued time)

## Authentication Flow

1. **User Login**:
   - Application calls `Authenticate(ctx, email, password)`
   - System fetches user by email and validates password
   - On success, generates a JWT token with `GenerateToken(userID)`

2. **Token Storage**:
   - Token is stored in an HTTP cookie named `friday_talks_auth_token`
   - Cookie is HTTP-only (inaccessible to JavaScript) for security
   - Uses SameSite Strict mode to prevent CSRF attacks

3. **Request Authentication**:
   - The `AuthMiddleware` intercepts requests
   - Extracts token from cookie
   - Validates token with `ValidateToken(tokenString)`
   - Fetches user from database and adds to request context

4. **Protected Routes**:
   - `RequireAuth` middleware checks if user exists in context
   - Redirects to login if not authenticated

## Token Generation

```go
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
    // ...
}
```

Key points:
- Uses HMAC-SHA256 algorithm for signing
- Includes expiration time (24 hours from creation)
- Signs token with the secret key

## Token Validation

```go
// ValidateToken validates a JWT token and returns the user ID if valid
func (a *Auth) ValidateToken(tokenString string) (int, error) {
    claims := &Claims{}

    token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, errors.New("unexpected signing method")
        }
        return a.jwtSecret, nil
    })
    // ...
}
```

Validation checks:
- Token is properly signed with our secret
- Token hasn't expired
- Signing method matches what we expect (HMAC)

## Security Considerations

1. **Token Storage**:
   - Uses HTTP-only cookies to prevent JavaScript access
   - SameSite Strict policy to prevent CSRF attacks
   - Should use Secure flag in production (HTTPS only)

2. **Token Expiration**:
   - Tokens expire after 24 hours
   - Expired tokens are rejected during validation

3. **Error Handling**:
   - Invalid tokens are cleared from cookies
   - Users with invalid tokens proceed as unauthenticated

## Advantages of JWT

1. **Stateless**: Server doesn't need to store session data
2. **Scalable**: Works well in distributed systems
3. **Self-contained**: Contains all necessary user information
4. **Secure**: Signed to prevent tampering

## Limitations

1. **Token Revocation**: Can't easily revoke tokens before expiration
2. **Token Size**: More data than session IDs
3. **Secret Management**: Requires secure handling of signing keys

## Conclusion

JWT provides a secure, stateless authentication method for our Friday Talks application. The implementation handles user authentication, token generation/validation, and request authorization through middleware. 