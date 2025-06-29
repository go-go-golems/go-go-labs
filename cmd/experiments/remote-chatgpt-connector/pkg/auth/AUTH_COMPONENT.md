# Auth Component Documentation

## Overview

The auth component implements GitHub OAuth token validation for the remote ChatGPT MCP connector. It validates GitHub access tokens via the GitHub API and enforces user allowlist controls.

## Token Validation Flow

### 1. Token Extraction
- Tokens are expected in HTTP `Authorization` header as `Bearer <token>`
- The `Bearer ` prefix is automatically stripped during validation
- GitHub tokens are opaque (not JWTs) and require API validation

### 2. GitHub API Validation 
- Makes authenticated `GET /user` request to GitHub API
- Validates token authenticity and retrieves user information
- Returns structured user data including ID, login, and email

### 3. Allowlist Check
- Compares GitHub `login` against `GITHUB_ALLOWED_LOGIN` environment variable
- Only allows exact matches (case-sensitive)
- Rejects unauthorized users with descriptive error

### 4. Error Handling
The validator handles various error cases gracefully:

```go
switch resp.StatusCode {
case http.StatusUnauthorized:
    return fmt.Errorf("invalid or expired github token")
case http.StatusForbidden:  
    return fmt.Errorf("github token lacks required permissions")
case http.StatusTooManyRequests:
    return fmt.Errorf("github api rate limit exceeded")
default:
    return fmt.Errorf("github api error: %s", resp.Status)
}
```

## GitHub API Rate Limits

### Rate Limit Details
- **Authenticated requests**: 5,000 requests per hour per user
- **Rate limit headers**: Check `X-RateLimit-Remaining` and `X-RateLimit-Reset`
- **Token validation cost**: 1 request per validation

### Recommended Caching Strategy
To minimize API calls and respect rate limits:

1. **Token Caching**: Cache valid tokens for their lifetime (typically 1 hour)
2. **User Info Caching**: Cache user information to avoid repeated lookups
3. **Rate Limit Monitoring**: Track remaining requests and implement backoff

Example caching implementation:
```go
type CachedValidator struct {
    validator types.AuthValidator
    cache     map[string]*CacheEntry
    mu        sync.RWMutex
}

type CacheEntry struct {
    UserInfo  *types.UserInfo
    ExpiresAt time.Time
}

func (c *CachedValidator) ValidateToken(ctx context.Context, token string) (*types.UserInfo, error) {
    c.mu.RLock()
    if entry, exists := c.cache[token]; exists && time.Now().Before(entry.ExpiresAt) {
        c.mu.RUnlock()
        return entry.UserInfo, nil
    }
    c.mu.RUnlock()
    
    // Validate and cache result
    userInfo, err := c.validator.ValidateToken(ctx, token)
    if err == nil {
        c.mu.Lock()
        c.cache[token] = &CacheEntry{
            UserInfo:  userInfo,
            ExpiresAt: time.Now().Add(30 * time.Minute), // Cache for 30 minutes
        }
        c.mu.Unlock()
    }
    return userInfo, err
}
```

## Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `GITHUB_CLIENT_ID` | GitHub OAuth App Client ID | `Iv1.xxxxxxxxxxxxx` |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth App Client Secret | `xxxxxxxxxxxxxxxx` |
| `GITHUB_ALLOWED_LOGIN` | GitHub username to allow access | `manuelod` |

### GitHub OAuth App Setup
1. Go to GitHub Settings → Developer settings → OAuth Apps
2. Create new OAuth App with:
   - **Authorization callback URL**: `http://127.0.0.1/ignore` (temporary)
   - **Scopes**: `read:user` (minimum required)
3. Copy Client ID and Client Secret to environment variables

## Mock Auth Validator

For testing other components without GitHub API dependencies:

```go
mockAuth := auth.NewMockAuthValidator(true, logger)
// Always succeeds and returns mock user info

mockAuth := auth.NewMockAuthValidator(false, logger)  
// Always fails for testing error paths
```

Mock returns consistent test data:
- **ID**: `mock-123`
- **Login**: `mockuser`
- **Email**: `mock@example.com`
- **Verified**: `true`

## Integration with HTTP Middleware

### Basic Middleware Integration
```go
func AuthMiddleware(validator types.AuthValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "missing authorization header", http.StatusUnauthorized)
                return
            }
            
            // Validate token
            userInfo, err := validator.ValidateToken(r.Context(), authHeader)
            if err != nil {
                http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
                return
            }
            
            // Add user info to request context
            ctx := context.WithValue(r.Context(), "user", userInfo)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Usage in HTTP Server
```go
// Initialize validator
authValidator := auth.NewGitHubAuthValidator(
    config.GitHubClientID,
    config.GitHubClientSecret, 
    config.AllowedLogin,
    logger,
)

// Create middleware
authMW := AuthMiddleware(authValidator)

// Apply to protected routes
mux.Handle("/sse", authMW(sseHandler))
mux.Handle("/api/search", authMW(searchHandler))
```

### Extracting User Context
```go
func protectedHandler(w http.ResponseWriter, r *http.Request) {
    userInfo := r.Context().Value("user").(*types.UserInfo)
    logger.Info().
        Str("user_login", userInfo.Login).
        Str("user_id", userInfo.ID).
        Msg("authenticated request")
    
    // Handle request with user context
}
```

## Error Logging

The validator uses structured logging with zerolog:

```go
// Debug level - token validation attempts
g.logger.Debug().Msg("validating github token")

// Info level - successful validations  
g.logger.Info().Str("github_login", ghUser.Login).Msg("github token validated successfully")

// Warn level - authorization failures
g.logger.Warn().Str("github_login", ghUser.Login).Msg("github user not in allowlist")

// Error level - API or network failures
g.logger.Error().Err(err).Msg("github api request failed")
```

Set log level via `LOG_LEVEL` environment variable (`debug`, `info`, `warn`, `error`).

## Security Considerations

1. **Token Storage**: Never log or store GitHub tokens in plaintext
2. **HTTPS Only**: Use HTTPS in production - GitHub requires secure callbacks
3. **Scope Minimization**: Only request `read:user` scope (minimum required)
4. **Client Secret Protection**: Store client secret securely (env vars, secrets manager)
5. **Rate Limit Compliance**: Implement caching to avoid hitting API limits
6. **Token Expiration**: GitHub tokens can expire - handle gracefully

## Testing

### Unit Tests
```go
func TestGitHubValidator(t *testing.T) {
    validator := auth.NewGitHubAuthValidator(
        "test-client-id",
        "test-client-secret", 
        "testuser",
        zerolog.Nop(),
    )
    
    // Test cases for valid/invalid tokens, rate limits, etc.
}
```

### Integration Tests
- Test against GitHub API with test tokens
- Verify rate limit handling
- Test allowlist enforcement
- Validate error responses

### Mock Testing
Use `MockAuthValidator` for testing dependent components without GitHub API calls.
