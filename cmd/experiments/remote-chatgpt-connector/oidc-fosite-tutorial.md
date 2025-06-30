Great. I’ll begin writing a complete book that guides readers through building a simple OAuth2/OIDC dynamic registration server in Go using the Fosite library. The server will act as a Model Context Protocol (MCP) connector as per the March 2025 MCP specification.

The book will be structured as a hands-on technical guide with code examples, step-by-step setup instructions, and implementation of an in-memory or SQLite-based authentication system. It will support dynamic client registration, username/password login (hardcoded as wesen/secret), and access verification for MCP integration.

I’ll get started and notify you once the draft is ready.


# Implementing a Simple OIDC Dynamic Client Registration Server (with OAuth2) for MCP

## Introduction to OIDC Dynamic Registration and MCP Context

OpenID Connect (OIDC) Dynamic Client Registration is a mechanism that allows OAuth2/OIDC clients to **register themselves at an Authorization Server at runtime**, instead of requiring manual pre-configuration. In the context of the **Model Context Protocol (MCP)** – an emerging standard for AI models interacting with tools – dynamic client registration is especially important. MCP clients may not know all possible MCP servers ahead of time, so allowing automatic registration streamlines connectivity. The MCP Authorization Spec explicitly recommends that *MCP servers support the OAuth 2.0 Dynamic Client Registration Protocol (RFC 7591)*. This enables an MCP client (like an AI agent or external tool) to **obtain OAuth2 client credentials on the fly**, which is crucial because:

* Clients cannot know all possible servers in advance, especially as new MCP servers/tools appear.
* Manual client registration would create friction for users connecting to new servers.
* Dynamic registration allows seamless, standardized onboarding of new clients, while still letting servers enforce their own policies (e.g. approval or limits).

In simpler terms, **OIDC dynamic registration** lets our OAuth2 Authorization Server create new client entries at runtime via a REST API (typically the `/register` endpoint). The server generates a `client_id` (and a `client_secret` for confidential clients) and returns it to the client, which can then proceed with the normal OAuth flows.

This tutorial will guide you through implementing a **minimal OIDC-compliant Authorization Server** that supports dynamic client registration, using the Go `fosite` library by Ory. Our server will double as an MCP server's auth component, meaning it will expose the standard OAuth2 endpoints (as required by MCP):

* `/.well-known/oauth-authorization-server` – for metadata (we will mention but not fully implement metadata here).
* `/authorize` – the authorization endpoint (for user login and consent).
* `/token` – the token endpoint (for exchanging grants for tokens, refreshing tokens, etc.).
* `/register` – the client registration endpoint (for dynamic client registration).

We'll use **Ory Fosite** to handle much of the OAuth2/OIDC logic for security and spec compliance. The focus will be on implementing:

1. A dynamic client registration handler (`POST /register`) that allows a client to register and get credentials.
2. A simple user authentication (with a single hardcoded username/password: **wesen/secret** for demo) integrated into the authorization flow.
3. The OAuth2 Authorization Code grant flow (with PKCE for security) as an example of logging in and obtaining an access token (and optionally an ID token for OIDC).
4. Token verification on a protected resource endpoint to confirm that the MCP client (on behalf of user "wesen") can access a resource using the obtained token.

Throughout, we'll keep the implementation **minimal**: an in-memory data store (or lightweight SQLite, if preferred) will suffice for storing clients, users, and tokens. This is not production-ready, but is enough to illustrate the concepts. We'll assume our application *is* the MCP server's auth service (i.e., it's both Authorization Server and Resource Server in OAuth terms). Let's get started by understanding the components we need to build.

## Overview of the Authorization Server Components

To implement an OAuth2/OIDC server with dynamic registration, we need to set up several components:

* **Data Models / Storage**: We need to store:

  * *OAuth2 Clients*: Each client has at least a `client_id`, possibly a `client_secret` (for confidential clients), allowed redirect URIs, grant types, and other metadata. In dynamic registration, these are created on the fly.
  * *Users (Resource Owners)*: We need a way to authenticate end-users. For simplicity, we'll hardcode one user account with username "`wesen`" and password "`secret`". In a real app, this would be a user database.
  * *Authorization Codes, Access Tokens, Refresh Tokens*: These will be issued during the auth flows. Fosite can handle generation and verification of these, but it uses our storage interface to persist them (in memory or DB).

* **Endpoints**: We must implement the HTTP endpoints that make up the OAuth2 Authorization Server:

  1. **Dynamic Client Registration (`POST /register`)** – accepts client metadata and creates a new client record, returning the client credentials. This is defined by [RFC 7591](https://datatracker.ietf.org/doc/html/rfc7591).
  2. **Authorization Endpoint (`GET/POST /authorize`)** – handles the start of an OAuth2 Authorization Code flow. It will authenticate the user (show a login form for username/password, or accept credentials via POST for simplicity) and obtain user consent for scopes, then issue an authorization code.
  3. **Token Endpoint (`POST /token`)** – handles exchanging an authorization code for tokens, or other grant types like refreshing tokens or client credentials. In our case, it will primarily handle the code exchange (and possibly the Resource Owner Password Credentials grant for simplicity).
  4. (Optionally, **Token Introspection/Revocation** – for validating or revoking tokens, and **.well-known metadata** – for discovery. These are important in real systems, but we might skip or briefly mention them in this simple implementation.)

* **MCP Protected Resource**: Since our MCP server is also the resource server, we should have a sample protected endpoint (e.g., `/v1/contexts` or `/mcp/data`) that requires a valid access token. The MCP spec mandates that every request from client to server include an `Authorization: Bearer <token>` header once authenticated. We will demonstrate that after login, the user "wesen" can successfully access a protected resource with the issued token, proving the flow works.

* **Security Configuration**: OAuth2/OIDC flows have security requirements:

  * We must support PKCE (Proof Key for Code Exchange) for the authorization code flow (in fact, the MCP spec *requires* PKCE for all clients for security). Fosite will enforce this by default for public clients.
  * We will generate cryptographically secure tokens. Fosite by default uses opaque tokens (random strings with an HMAC signature) for access and refresh tokens. For OIDC ID Tokens (if we issue them), we will need a signing key (RSA or EC) to sign JWTs.
  * Redirect URI validation: The server must only redirect to allowed URLs registered by the client, to prevent phishing or token leakage. We'll enforce that in our client registration and authorization flow (fosite helps with this too).
  * All endpoints should be served over HTTPS in production, but for our example, we'll assume a local server ([http://localhost](http://localhost)) environment for testing.

Next, we will set up our project and incorporate Ory Fosite to handle OAuth2 details.

## Setting Up the Project with Ory Fosite

**Ory Fosite** is an OAuth2/OIDC SDK for Go that handles the heavy lifting of the protocols. Using Fosite allows us to avoid writing the entire OAuth logic from scratch and ensures compliance with the specs. Fosite is designed to be **extensible** and allows plugging in various "grant handlers" depending on which flows you want to support. For example, it has built-in factories for Authorization Code grant, Implicit grant, Client Credentials, Refresh Tokens, and even OIDC-specific extensions like ID tokens.

We'll outline a basic Go project structure:

* **main.go**: Initializes the Fosite provider and HTTP routes.
* **store.go** (or a simple in-memory store struct): Implements Fosite's storage interfaces for clients, users, tokens.
* **handlers.go**: Our HTTP handlers for the endpoints (/register, /authorize, /token, etc).

We'll use minimal dependencies: primarily the `fosite` library and the Go standard library (or a minimal router like net/http or Gin for convenience). In the interest of focus, pseudocode or simplified code snippets will illustrate key parts rather than a full program listing.

**Initialize Fosite**: In `main.go`, we'll configure Fosite with a global secret for token HMAC signing and register the OAuth2 grant handlers we need. For example:

```go
import (
    "github.com/ory/fosite"
    "github.com/ory/fosite/compose"
)

// ...

// Our global HMAC secret for signing tokens (32 or 64 bytes recommended)
var globalSecret = []byte("a-very-long-32-byte-secret-value....")

// Fosite configuration
config := &fosite.Config{
    GlobalSecret: globalSecret,
    // You can customize other settings like token lifetimes here
}

// Instantiate our storage (in-memory in this example)
store := NewMemoryStore()  // We'll implement this shortly

// Choose an appropriate strategy for generating tokens
strategy := compose.NewOAuth2HMACStrategy(config) // HMAC for opaque tokens

// Compose the OAuth2 provider with the grant types we want to support
oauth2Provider := compose.Compose(
    config,
    store,
    strategy,
    compose.OAuth2AuthorizeExplicitFactory,    // Authorization Code Grant
    compose.OAuth2TokenExchangeFactory,        // Token endpoint for auth code exchange
    compose.OAuth2RefreshTokenGrantFactory,    // Support refresh tokens
    compose.OAuth2ClientCredentialsGrantFactory, // (optional, for client_credentials flow)
    compose.OpenIDConnectExplicitFactory,      // OIDC support for Authorization Code
    compose.OAuth2PKCEFactory,                 // PKCE support for code flow
    // ... we can include other factories like Implicit if needed
)
```

In the above snippet:

* We set a `GlobalSecret` used by Fosite to sign tokens (for integrity). This secret must be kept safe; it's used in HMAC signing of tokens.
* We use `NewMemoryStore()` to get a storage that implements `fosite.Storage`. We'll create a simple version of this that stores data in memory (with maybe a backing map or two). Fosite will call this to persist and lookup clients, auth codes, tokens, etc.
* We register the **Authorization Code grant** (`OAuth2AuthorizeExplicitFactory`) along with PKCE (`OAuth2PKCEFactory`), so that our server supports the standard OAuth2 code flow with PKCE (suitable for public clients, such as MCP clients running in user applications).
* We include `OpenIDConnectExplicitFactory` to enable OIDC features (specifically ID Token issuance in the code flow when `openid` scope is requested).
* Refresh tokens and client credentials are also included (refresh tokens allow long-lived sessions; client credentials grant could be useful for non-user scenarios, e.g., an MCP client acting on its own).
* We could include Resource Owner Password Credentials grant (`OAuth2ResourceOwnerPasswordCredentialsFactory`) if we wanted to allow direct username/password exchange at `/token`. However, this grant is deprecated in OAuth2.1 and not recommended. Since our scenario involves a user login, we'll stick with the Authorization Code flow to handle username/password via the `/authorize` web endpoint (which is more secure, since the client never sees the password, only the auth server does).

Now that we have an `oauth2Provider` (an instance of `fosite.OAuth2Provider`), we can set up the HTTP server routes and integrate Fosite. Before that, let's implement our storage and data models.

## In-Memory Storage for Clients, Users, and Sessions

**Clients**: We'll maintain a map of client ID to client information. Fosite expects us to implement the `ClientManager` interface (methods like `GetClient(ctx, id) (fosite.Client, error)`). For simplicity, we'll define a struct for our clients:

```go
type Client struct {
    ID            string
    Secret        string // hashed secret if using confidential clients
    RedirectURIs  []string
    GrantTypes    []string
    ResponseTypes []string
    Scopes        []string
    Public        bool   // true if public client (no secret required for token requests)
}
```

We can make our `Client` struct implement `fosite.Client` interface (which includes methods like `GetID()`, `IsPublic()`, etc.). Alternatively, we can use Fosite's default `fosite.DefaultClient` struct. For brevity, using `fosite.DefaultClient` is convenient: it has fields for ID, secret, redirect URIs, scopes, etc. We'll use that for storing clients dynamically.

**Users**: We'll hardcode a single user. A simple struct can hold user info:

```go
type User struct {
    Username string
    PasswordHash string  // we could store plain for demo, but let's assume a hash
    // maybe an ID or other claims like full name if needed
}
```

We can pre-create `User{Username: "wesen", PasswordHash: "<hashOf('secret')>"}`. The `Authenticate(username, password)` function will just check against this one record.

**Authorization Codes and Tokens**: Fosite's storage interface also covers temporary credentials:

* `AuthorizeCode` storage (for pending authorization codes issued and not yet exchanged).
* `AccessToken` storage and `RefreshToken` storage.
* `PKCE` storage (if PKCE is used, to store the code challenge until exchange).
* `Implicit`/`State` management, etc.
* We also need to store OIDC sessions (to recall data for ID token claims).

Because this is complex to implement fully, we have options:

* Use an existing in-memory implementation provided by Fosite (Fosite in the past had an `MemoryStore` mainly for testing, but it's not part of the public API; we might implement minimal functions ourselves).
* Implement a subset of the storage interface. For our case, we will implement enough to get through a basic code flow:

  * Store authorization code requests in a map (keyed by code signature).
  * Store access tokens in a map (keyed by token signature) for introspection/verification.
  * Store refresh tokens similarly.
  * Each stored item maps to a `fosite.Request` object or equivalent that contains the details (client ID, scopes, session, etc.). The Charles Muchogo article demonstrates how to store and retrieve these objects from a DB using GORM, but we'll do a simpler in-memory approach.

For brevity, let's outline a **MemoryStore** struct:

```go
type MemoryStore struct {
    Clients       map[string]fosite.Client               // client_id -> client
    Users         map[string]User                        // username -> User
    AuthCodes     map[string]fosite.Requester            // code signature -> request
    AccessTokens  map[string]fosite.Requester            // access token signature -> request
    RefreshTokens map[string]fosite.Requester            // refresh token signature -> request
    PKCEs         map[string]fosite.Requester            // code challenge signature -> request (if needed)
}
```

We initialize it with `NewMemoryStore()` which sets up the maps and maybe pre-registers some initial data:

* It will create the `Users` map with the `"wesen"` user.
* It could also pre-create a test client if we want one default client (but since we support dynamic reg, we might start with none, or perhaps a special "admin" client if dynamic reg required an initial token – but the MCP spec suggests open registration can be allowed).

In dynamic registration, there's a question: do we require the registration endpoint to be authenticated (with an initial access token or API key), or open to the public? The OIDC Dynamic Client Registration spec allows **open registration** or protected registration (with an initial "software statement" or admin token). The MCP spec does not seem to mandate an initial token for registration (and it says MCP clients *should* attempt metadata discovery, but then can call `/register` presumably openly unless otherwise stated). For simplicity, we'll allow unauthenticated access to `POST /register` (anyone can register a client). The server could impose rate limits or policies, but we won't cover that here.

So our `/register` handler will not require auth in our example. (In production, some servers might restrict who can register clients.)

Now let's implement the key methods of `MemoryStore` for Fosite:

* `GetClient(ctx, id) (fosite.Client, error)`: Look up the client in the `Clients` map by ID. Return it if found, or `fosite.ErrNotFound` if not.
* `CreateClient(ctx, client) error`: (not part of Fosite's interface by default, but we might use it internally when registering new clients).
* The token session methods:

  * `CreateAuthorizeCodeSession(ctx, code, request) error` and `GetAuthorizeCodeSession(ctx, code, session) (fosite.Requester, error)` – store and retrieve auth code data. In memory, we can store `request` (which is a `fosite.Request` containing all details including session) keyed by a code-signature. **Important**: Fosite will give us a "code signature" (not the raw code) as the key, to avoid storing raw codes. We'll store using that signature as the map key.
  * `DeleteAuthorizeCodeSession(ctx, code) error` – remove a code once used (to prevent re-use).
  * Similarly, `CreateAccessTokenSession`, `GetAccessTokenSession`, `DeleteAccessTokenSession` for access tokens, and same for refresh tokens.
  * `CreatePKCERequestSession` and `GetPKCERequestSession` for PKCE if using an authorizations store.

Implementing all of these is a bit involved but straightforward with maps. For example:

```go
func (m *MemoryStore) GetClient(_ context.Context, id string) (fosite.Client, error) {
    client, ok := m.Clients[id]
    if !ok {
        return nil, fosite.ErrNotFound
    }
    return client, nil
}

func (m *MemoryStore) CreateAuthorizeCodeSession(_ context.Context, code string, req fosite.Requester) error {
    m.AuthCodes[code] = req
    return nil
}
func (m *MemoryStore) GetAuthorizeCodeSession(_ context.Context, code string, session fosite.Session) (fosite.Requester, error) {
    req, ok := m.AuthCodes[code]
    if !ok {
        return nil, fosite.ErrNotFound
    }
    // Notice: We might need to set the session pointer to req.GetSession() here 
    // if using concrete types. In fosite, the session passed in is to be populated.
    return req, nil
}
func (m *MemoryStore) DeleteAuthorizeCodeSession(_ context.Context, code string) error {
    delete(m.AuthCodes, code)
    return nil
}
```

And similarly for `AccessTokenSession` and `RefreshTokenSession`. In those, the key will be the token's signature (Fosite passes a signature of the token, not the actual token string, for security). We treat it similarly: store the request by signature, and remove when asked.

For PKCE, Fosite will call e.g. `CreatePKCERequestSession(ctx, signature, req)` when an auth code with PKCE is initiated, and `GetPKCERequestSession(ctx, signature, session)` when exchanging the code. We can implement those to store/retrieve from a `PKCEs` map.

Finally, we need to handle user authentication. Fosite does not manage user verification; we have to do that before calling `NewAuthorizeResponse`. Typically:

* The `/authorize` handler will call `oauth2Provider.NewAuthorizeRequest(ctx, httpRequest)` to parse and validate the incoming request (client ID, redirect URI, scope, etc.). If the request is invalid, we return an error.
* If it's valid, the next step is to authenticate the user (login). In our simple case, we check if the HTTP request has `username` and `password` fields (which might be sent via a POST from a login form). If not yet provided, we should present a login form to the user.
* Once credentials are provided, we verify them against our user store. If correct (username = "wesen" and password = "secret"), we consider the user authenticated. If not, we show an error or login page again.
* After authentication (and consent, if we had scopes to approve), we "grant" the requested scopes and issue an authorization code. With Fosite, this is done by creating a **session** object that holds user data (for OIDC, we want user identity in the ID token) and then calling `oauth2Provider.NewAuthorizeResponse(ctx, authRequest, session)`.

**Session data**: For OIDC, we should provide an OIDC-compliant session containing the user’s identity (at least a "subject" claim and potentially more). Fosite provides `openid.DefaultSession` which has fields for ID token claims and access token extra claims. We can use that, or define a custom struct embedding `DefaultSession` and adding our own fields.

For example, using `fosite/handler/openid`:

```go
import oidc "github.com/ory/fosite/handler/openid"

// ...

session := oidc.NewDefaultSession()  // this has fields for IDTokenClaims, etc.
session.Subject = userID  // the user's unique ID, or username as unique identifier
session.Claims = &oidc.IDTokenClaims{
    Subject: userID,
    Issuer:  "http://localhost:8080/", // the issuer URL of our server
    // you can set other standard claims like audience, expiry will be set automatically
}
session.Headers = &oidc.IDTokenHeaders{ // contains e.g. signing algorithm info
    Extra: map[string]interface{}{},
}
```

Here, `userID` could be "wesen" or an internal user UUID. The Issuer should be the base URL of our auth server (in production, a HTTPS URL). We will also need to ensure Fosite knows our signing key for ID tokens: typically we would configure an RSA private key in the compose (using e.g. `compose.NewOpenIDConnectStrategy` instead of the default HMAC for ID tokens). But to keep it simple, we might skip actual JWT signing and rely on opaque tokens only. (However, an OIDC ID token by definition is a JWT, so let's assume we configure an RSA key offline and pass it to Fosite's OIDC factory. For brevity, details of key generation are omitted.)

Now, let's proceed to implement each endpoint using these components.

## Implementing the Dynamic Client Registration Endpoint (`/register`)

The dynamic client registration endpoint is typically an unauthenticated `POST` where the client sends a JSON payload describing itself. According to RFC 7591, the request can include fields like:

* `redirect_uris` (required for an OIDC client that will use authorization\_code or implicit flows),
* `grant_types` (if omitted, defaults are usually based on `redirect_uris`; but we can infer if `redirect_uris` are given, we'll assume code flow),
* `client_name`, `scope`, `token_endpoint_auth_method` (like `client_secret_post` or `none` for public clients), etc.

For simplicity, we will require at least `redirect_uris` in the registration request (since our server will primarily support the code flow). We might also allow specifying `client_type` or an auth method to decide if the client is public or confidential.

**Example request** (from an MCP client perspective) to register:

```http
POST /register HTTP/1.1
Content-Type: application/json

{
  "client_name": "Sample MCP Client",
  "redirect_uris": ["http://localhost:8080/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "scope": "openid offline_access"
}
```

This requests a client that can do authorization code flow (with refresh tokens), expects an ID token (scope `openid`) and offline access (refresh token scope). We will interpret this and create a new client.

Our `/register` handler will:

1. Parse the JSON body.
2. Validate required fields. For an OIDC client, at least one redirect URI is typically required. We also ensure the URI is valid (proper URL format). We might ensure `grant_types` and `response_types` are consistent (if not provided, we can set defaults: e.g. if `response_types` includes "code", we include "authorization\_code" in grant\_types).
3. Determine if the client is public or confidential. By default, we can make clients **public** (no client secret) if they only use `authorization_code` with PKCE. If the registration explicitly asks for `token_endpoint_auth_method` = "client\_secret\_basic" or similar, then it's a confidential client and we'll issue a secret.
4. Generate a new `client_id`. This could be a random GUID or a securely generated string. (Fosite's default client expects a string ID, often a random value.) Generate a `client_secret` if needed (e.g., 32-character random string).
5. Store the new client in `MemoryStore.Clients`.
6. Respond with a JSON containing at least:

   * `client_id`
   * `client_secret` (if it’s a confidential client; if public client, omit or return `"client_secret": null`).
   * Registered metadata echo (like `redirect_uris`, `grant_types`, etc).
   * Optionally, `client_id_issued_at` (timestamp) and `client_secret_expires_at` (often 0 if it never expires).
   * **Registration Access Token and URI**: The OIDC spec (and Hydra) returns a `registration_access_token` and a `registration_client_uri` which the client can use to later *read/update/delete* its client registration. In this simple implementation, we will **omit** those (i.e. we won't implement the management API), but it's good to be aware of. In a real server, after registration, you'd return a one-time token that the client must use to authenticate to update or delete its registration in the future.

Let's sketch the code for `/register`:

```go
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    var regReq struct {
        RedirectURIs   []string `json:"redirect_uris"`
        GrantTypes     []string `json:"grant_types,omitempty"`
        ResponseTypes  []string `json:"response_types,omitempty"`
        ClientName     string   `json:"client_name,omitempty"`
        TokenEndpointAuthMethod string `json:"token_endpoint_auth_method,omitempty"`
        Scope          string   `json:"scope,omitempty"`
    }
    if err := json.NewDecoder(r.Body).Decode(&regReq); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }
    if len(regReq.RedirectURIs) == 0 {
        http.Error(w, "redirect_uris is required", http.StatusBadRequest)
        return
    }
    // Basic validation of redirect URIs (e.g., ensure proper URL format)...
    // Set defaults if necessary:
    if len(regReq.GrantTypes) == 0 {
        // Default to "authorization_code" (and "refresh_token" if OIDC scope implies offline_access)
        regReq.GrantTypes = []string{"authorization_code", "refresh_token"}
    }
    if len(regReq.ResponseTypes) == 0 {
        // Default to "code" for authorization_code grant
        regReq.ResponseTypes = []string{"code"}
    }
    // Determine token endpoint auth method
    var publicClient bool
    if regReq.TokenEndpointAuthMethod == "" {
        // If not specified, default:
        // If a secret is needed for code exchange? We'll assume public by default for simplicity.
        publicClient = true 
    } else if regReq.TokenEndpointAuthMethod == "none" {
        publicClient = true
    } else {
        // could be "client_secret_post" or "client_secret_basic"
        publicClient = false
    }
    // Generate client_id (random string) and maybe client_secret
    clientId := generateRandomID()  // e.g., 20-byte random base64 string or UUID
    var clientSecret string
    if !publicClient {
        clientSecret = generateSecret() // e.g., 32-byte random string
    }
    // Create a fosite.DefaultClient (from fosite) to represent this client
    client := &fosite.DefaultClient{
        ID:            clientId,
        RedirectURIs:  regReq.RedirectURIs,
        ResponseTypes: regReq.ResponseTypes,
        GrantTypes:    regReq.GrantTypes,
        Scopes:        fosite.Arguments{}, // we can parse regReq.Scope by splitting
        Public:        publicClient,
    }
    if !publicClient {
        // Set the hashed secret. Fosite expects the secret stored hashed (e.g. bcrypt)
        hashedSecret := fosite.HashSHA256([]byte(clientSecret))
        client.Secret = hashedSecret
    }
    // Store the client in our memory store
    memoryStore.Clients[clientId] = client

    // Prepare response
    resp := map[string]interface{}{
        "client_id": clientId,
        "client_id_issued_at": time.Now().Unix(),
    }
    if !publicClient {
        resp["client_secret"] = clientSecret
        resp["client_secret_expires_at"] = 0  // never expires
    } else {
        resp["token_endpoint_auth_method"] = "none"
    }
    if regReq.ClientName != "" {
        resp["client_name"] = regReq.ClientName
    }
    resp["redirect_uris"] = regReq.RedirectURIs
    resp["grant_types"] = regReq.GrantTypes
    resp["response_types"] = regReq.ResponseTypes
    if regReq.Scope != "" {
        resp["scope"] = regReq.Scope
    }

    // Typically we'd also return registration_access_token and registration_client_uri here, 
    // but we'll omit those for simplicity.
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}
```

This handler creates a new client with the provided parameters. We choose whether it's public or confidential based on the `token_endpoint_auth_method`. By default, if none provided, we treat it as public (which means no secret required when exchanging the code). If a secret is issued, we hash it before storing (since we should not store plaintext secrets on the server).

After calling `/register`, the client (MCP client) will get the JSON response with its credentials. For example, a successful response might look like:

```json
{
  "client_id": "oauth2-client-12345",
  "client_id_issued_at": 1688000000,
  "client_secret": "s3cr3t-abc-xyz", 
  "client_secret_expires_at": 0,
  "redirect_uris": ["http://localhost:8080/callback"],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "scope": "openid offline_access"
}
```

Now the MCP client has registered itself. Next, it will want to start the OAuth2 authorization flow to let the user (wesen) log in and authorize it.

## Implementing the Authorization Endpoint (`/authorize`) with User Login

The `/authorize` endpoint is where the user gets involved. In a standard OAuth2 Authorization Code flow:

1. The client directs the user's browser to the authorization URL (with query parameters response\_type=code, client\_id, redirect\_uri, scope, state, etc.).
2. The server (our implementation) needs to authenticate the user. If the user isn't logged in, it should prompt for login (username/password in our case).
3. After login (and typically consent to scopes), the server will generate an authorization code and redirect the user's browser to the client's `redirect_uri` with that code.

For our simple setup, we'll simulate this flow. We might not build a full HTML UI, but conceptually:

* If the request is GET (user coming in initially), we serve a simple HTML form where the user can enter username and password.
* If the request is POST (user submitted credentials), we verify them, then continue.

Using a web framework like Gin or net/http with a templating can help. For brevity, let's outline using net/http:

```go
func AuthorizeHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Let fosite parse the authorize request
    authReq, err := oauth2Provider.NewAuthorizeRequest(ctx, r)
    if err != nil {
        // If any parameters are wrong (unknown client, redirect URI mismatch, etc.), 
        // write an error using Fosite helper:
        oauth2Provider.WriteAuthorizeError(w, authReq, err)
        return
    }

    // Check if user is already authenticated (e.g., via session cookie).
    // For simplicity, we assume not, and require login for each request.
    if r.Method == "GET" {
        // Show login page (simple form)
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, `<html><form method="POST" action="/authorize?%s">
            <h3>Login</h3>
            <label>Username: <input name="username"></label><br>
            <label>Password: <input name="password" type="password"></label><br>
            <button type="submit">Log In</button>
            </form></html>`, r.URL.RawQuery)
        // We include the original query params in the form action so they carry through on POST.
        return
    }

    // If POST, handle credentials
    username := r.PostFormValue("username")
    password := r.PostFormValue("password")
    if username == "" || password == "" {
        // Missing credentials, show error or re-show form
        http.Error(w, "Username and password required", http.StatusBadRequest)
        return
    }
    // Authenticate the user
    user, ok := memoryStore.Users[username]
    if !ok || !CheckPassword(password, user.PasswordHash) {
        // Invalid credentials
        // In a real app, you might redirect back to login with error message
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    // At this point, user "wesen" is authenticated.

    // (We could handle user consent for scopes here. For simplicity, assume user consents to all requested scopes.)

    // Grant the requested scopes
    for _, scope := range authReq.GetRequestedScopes() {
        authReq.GrantScope(scope)
    }

    // Create a session for the user
    session := oidc.NewDefaultSession()
    session.Subject = user.Username  // using username as subject (or a fixed ID if available)
    session.Claims.Issuer = "http://localhost:8080/"
    session.Claims.Subject = user.Username
    session.Claims.AuthTime = time.Now().Unix()
    // (We could set session.Claims to include more, e.g., user.Name, etc., if needed)
    
    // Now, create an authorization response (which will include generating the code)
    response, err := oauth2Provider.NewAuthorizeResponse(ctx, authReq, session)
    if err != nil {
        // Handle errors (e.g., if some scope was not granted properly, etc.)
        oauth2Provider.WriteAuthorizeError(w, authReq, err)
        return
    }

    // If successful, Fosite will have stored the auth code and prepared a redirect.
    oauth2Provider.WriteAuthorizeResponse(w, authReq, response)
}
```

A few things to note in this handler:

* We call `NewAuthorizeRequest` to validate the incoming request. This checks the client\_id is valid and redirect URI matches the client's allowed URIs, among other things. If it fails, we use `WriteAuthorizeError` to send an error back to the user (which typically would redirect to the client's redirect URI with an error parameter, if possible).
* We handle GET vs POST differently. On GET, we output a simple HTML form. On POST, we process the form inputs.
* We verify the username/password. We only have "wesen"/"secret" as valid credentials. `CheckPassword` could simply compare plaintext for our demo (since storing plain "secret" is fine for a demo), or use a proper hash check if we stored a hash.
* After authentication, we **grant scopes**. Fosite will have parsed `scope` from the request (for example, the client might request `openid` or others). We should explicitly call `authReq.GrantScope()` for each scope the user (or server) approves. Here we just grant all requested scopes without a user consent prompt. (In production, you'd show a consent page listing scopes like "This app can access your profile/email" etc., and let the user approve or deny.)
* We then create a `session` object for the user. We used `oidc.NewDefaultSession()` since the client might have requested `openid` and we want to issue an ID token. We set the `Subject` (user ID) in both `session.Subject` and in the `session.Claims.Subject`. We also set `Issuer` claim. (Fosite will handle adding issue time, expiration, etc., and signing the ID token if configured with a signing key.)
* Finally, we call `NewAuthorizeResponse`, which will create an authorization code, store it (via our `MemoryStore.CreateAuthorizeCodeSession`), possibly create an ID token and sign it (which will be stored temporarily until token exchange), and prepare a redirect to the client's redirect URI with `?code=<code>&state=<state>` (and `id_token` if we were using implicit flow or hybrid, but since we are using code flow, the ID token will be delivered from the token endpoint, not directly here).
* We then call `WriteAuthorizeResponse` to send the HTTP redirect to the user's browser. Under the hood, this sets a Location header to the redirect URI. The user’s browser will follow that redirect, sending them back to the client application along with the code (and state).

At this point, the user "wesen" has logged in and authorized the client. The client application (MCP client) will receive the authorization code in the callback.

Now we implement the token endpoint to handle exchanging that code for tokens.

## Implementing the Token Endpoint (`/token`)

The token endpoint is called by the client server-to-server (not via the user's browser) to exchange the auth code for an access token (and refresh token, and ID token if applicable). It also handles refreshing tokens and other grant types (like client\_credentials or password, if implemented).

In our scenario, after the user authorized and the client got the `code`, the client will make a POST request to our `/token`, for example:

```
POST /token HTTP/1.1
Content-Type: application/x-www-form-urlencoded
Authorization: Basic base64(client_id:client_secret)   (if confidential client)

grant_type=authorization_code&
code=SplxlOBeZQQYbYS6WxSbIA&
redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&
code_verifier=plain-text-PKCE-verifier
```

Key points:

* The client must authenticate to the token endpoint. For public clients, they don't use an Authorization header (they should not have a client\_secret). For confidential clients, we expect HTTP Basic auth or `client_id` and `client_secret` in the form (depending on `token_endpoint_auth_method`). Fosite will handle checking the auth if the client has a secret.
* The `grant_type` here is "authorization\_code". The code, redirect\_uri, and PKCE code\_verifier must all be provided and valid.
* Fosite will verify:

  * The code exists and is not expired (using our storage's `GetAuthorizeCodeSession`).
  * The code was issued to the same client that's presenting it (and client is authenticated if required).
  * The `redirect_uri` matches what was used in the authorize request.
  * The PKCE verifier matches the challenge (for public clients).
* If valid, Fosite will mark the code as used (delete it), and then issue an access token (and refresh token if offline\_access/scope present) and an ID token if openid scope was granted.

Our `/token` handler in code:

```go
func TokenHandler(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    // Fosite expects URL form-encoded body, so ensure that's parsed:
    if err := r.ParseForm(); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }
    accessRequest, err := oauth2Provider.NewAccessRequest(ctx, r, new(oidc.DefaultSession)) 
    // We pass a new DefaultSession here; fosite will populate it with session from code.
    if err != nil {
        oauth2Provider.WriteAccessError(w, accessRequest, err)
        return
    }
    // Now create a response
    response, err := oauth2Provider.NewAccessResponse(ctx, accessRequest)
    if err != nil {
        oauth2Provider.WriteAccessError(w, accessRequest, err)
        return
    }
    // on success, send tokens
    oauth2Provider.WriteAccessResponse(w, accessRequest, response)
}
```

What happens here:

* `NewAccessRequest` will handle all the validation for us. We provided a `new(oidc.DefaultSession)` as a session. If the code being exchanged had an OIDC session with user info, Fosite will merge that into this session for token generation.
* If there's an error (e.g., invalid code, wrong client auth), we write an error with `WriteAccessError`.
* If successful, `NewAccessResponse` will generate the tokens. In particular, it will generate:

  * An Access Token (opaque string or JWT depending on config; by default, an opaque token signed with our HMAC secret).
  * A Refresh Token (if the scope "offline\_access" or similar is present, indicating offline use).
  * An ID Token (a JWT containing the user identity, if the scope "openid" was granted).
* `WriteAccessResponse` will output a JSON response containing the tokens. For example:

Successful token response (example):

```json
{
  "access_token": "ZG9uJ3QgdHJ5IHRvIGRlY29kZSB0aGlzCg==.LzK... (opaque token)",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "I6yZ... (if issued)",
  "id_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9... (JWT if openid scope)"
}
```

Fosite takes care of crafting this JSON and including only the fields that apply (it will include `id_token` if an OIDC session was present and openid scope, include `refresh_token` if one was issued, etc.).

At this point, the OAuth2 dance is complete: the MCP client has obtained an `access_token` it can use to call protected APIs on the MCP server on behalf of user *wesen*. It possibly also has an `id_token` that represents the user's identity and a `refresh_token` to get new access tokens when the current one expires.

## Verifying the Token on Protected MCP Endpoints

Now we demonstrate the **resource access** step. The MCP server likely offers various endpoints that require authorization – for example, listing available tools/contexts (`GET /v1/contexts`) or executing actions. The MCP spec requires that the client sends the access token in the `Authorization: Bearer <token>` header of each request. The server must validate the token for each request.

In our implementation, we can add a simple protected endpoint, e.g. `GET /hello` or `GET /v1/contexts`, which:

* Checks for an `Authorization` header.
* If missing or invalid, returns `401 Unauthorized` (as per spec requirement when auth is needed and not provided).
* If present, validate the token:

  * If we used opaque tokens, we need to introspect them or store them in a map to check if they are active. With Fosite, we can use `oauth2Provider.IntrospectToken` to validate the token and get the corresponding request (which includes the granted scopes and session info).
  * If we used JWT access tokens (not in this example), we'd verify the signature and claims.
* Check the scopes or permissions if necessary for the resource.
* If valid, proceed to respond with the resource data.

Since we included the `compose.OAuth2TokenIntrospectionFactory` in our provider, we have introspection capability. But we can also directly leverage our `MemoryStore` by looking up the token signature.

Using Fosite's approach:

```go
func ProtectedResourceHandler(w http.ResponseWriter, r *http.Request) {
    // Example protected endpoint
    tokenStr := ExtractBearerToken(r.Header.Get("Authorization"))
    if tokenStr == "" {
        w.Header().Set("WWW-Authenticate", `Bearer realm="MCP", error="invalid_request"`)
        http.Error(w, "Missing access token", http.StatusUnauthorized)
        return
    }
    ctx := r.Context()
    // Introspect the token (validate and get identity)
    var session oidc.DefaultSession
    requester, err := oauth2Provider.IntrospectToken(ctx, tokenStr, fosite.AccessToken, &session)
    if err != nil {
        w.Header().Set("WWW-Authenticate", `Bearer realm="MCP", error="invalid_token"`)
        http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
        return
    }
    // If introspection succeeds, we have the token's requester (contains client, scopes, etc.)
    username := session.Claims.Subject  // we set subject to username earlier
    // We can also check scopes if needed:
    if !requester.GetGrantedScopes().Has("desired_scope") {
        // if the resource requires a specific scope that wasn't granted
        http.Error(w, "Forbidden - scope required", http.StatusForbidden)
        return
    }
    // Otherwise, token is valid and authorized:
    fmt.Fprintf(w, "Hello %s, your request to MCP is authorized!", username)
}
```

This handler uses `IntrospectToken` which will:

* Verify the token signature with our HMAC secret.
* Check if the token is not expired or revoked.
* Load the stored request from our store (`GetAccessTokenSession`) to reconstruct the details.
* Ensure it's an access token (not an ID or refresh token) by the type we pass (`fosite.AccessToken`).
  If the token is valid, we retrieve the associated `DefaultSession` (which we stored when issuing it) – it contains the `Claims.Subject` which we set to "wesen". We greet the user or proceed to provide the actual resource data.

If the token was missing or invalid, we return a `401 Unauthorized`. If the token is valid but the user lacks a required scope, we return `403 Forbidden` (the MCP spec notes 403 is for invalid scope or insufficient permissions).

This demonstrates verifying that *user "wesen" is indeed accessing the resource with a valid token*, meaning our whole flow succeeded.

## Testing the End-to-End Flow

Let's summarize the steps and how we can test each part, using curl or a REST client:

1. **Dynamic Client Registration**:

   * **Request**: `POST /register` with JSON body as described earlier. No auth required.
   * **Response**: JSON containing `client_id` (and `client_secret` if applicable). **Save these**, as the client will need them.
   * *Example*:

     ```bash
     curl -X POST http://localhost:8080/register -H "Content-Type: application/json" -d '{
       "client_name": "Test MCP Client",
       "redirect_uris": ["http://localhost:8080/callback"],
       "grant_types": ["authorization_code"],
       "response_types": ["code"],
       "scope": "openid offline_access"
     }'
     ```

     (This assumes our server runs on port 8080.) The output gives a new `client_id` and maybe a `client_secret`. Let's say it returns:

     ```json
     { "client_id": "abcd1234", "client_secret": "efgh5678", ... }
     ```

     For this example, assume it's a public client, so no secret returned (token\_endpoint\_auth\_method "none").

2. **Authorization Code Request (User Login)**:

   * Normally, the client would open the user's browser to the `/authorize` URL. We can simulate this with a browser or with curl in two steps since we have to handle cookies (for session) or follow redirects manually. E.g.:

     ```
     http://localhost:8080/authorize?response_type=code&client_id=abcd1234&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&scope=openid+offline_access&state=xyz&code_challenge=...&code_challenge_method=S256
     ```

     This URL includes the client\_id from step 1, the redirect URI (URL-encoded), scopes, a random state, and a PKCE code challenge. For simplicity, if we didn't implement PKCE verification manually, Fosite would have required it for public clients. We should generate a code verifier and challenge (S256). For testing, one might use a tool or script; but in a simplified case, you could disable PKCE requirement or provide a dummy challenge of method "plain" if allowed. The MCP spec and security best practice is to use PKCE, so let's assume we do:

     * Code verifier (random string), code challenge = SHA256(verifier).

   * When you hit that URL in a browser, our server will show the login form. Enter "wesen" / "secret" and submit.

   * The server will then redirect back to the redirect\_uri (`http://localhost:8080/callback?code=XXXXX&state=xyz`). Since our example redirect URI is just a placeholder, we can simulate it by running a small temporary server or by examining the Location header from the authorize response.

   * If using a command-line approach: one could use curl to post credentials:

     ```bash
     curl -X POST -L -d "username=wesen&password=secret" "http://localhost:8080/authorize?response_type=code&client_id=abcd1234&redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback&scope=openid+offline_access&state=xyz&code_challenge=CHALLENGE&code_challenge_method=plain"
     ```

     The `-L` follows redirects. We used a plain code challenge for simplicity here. The final output might be the HTML of a blank page at the callback or an error since our callback isn't actually running a server. However, you can extract the `code` from the redirect URL (curl with `-v` to see Location header, for instance).

   * Let's say we got an auth code `abcde`. Now the client has this code.

3. **Token Exchange**:

   * The client now calls `POST /token` with the code. For a public client, it doesn't send client\_secret. For PKCE, it must send the code\_verifier.
   * Example:

     ```bash
     curl -X POST http://localhost:8080/token -H "Content-Type: application/x-www-form-urlencoded" -d "grant_type=authorization_code&code=abcde&redirect_uri=http://localhost:8080/callback&code_verifier=CHALLENGE"
     ```

     (If the client were confidential, we would add `-u client_id:client_secret` or include them in the form.)
   * Response: JSON with `access_token`, etc. e.g.

     ```json
     {
       "access_token": "2YotnFZFEjr1zCsicMWpAA...", 
       "token_type": "Bearer",
       "expires_in": 3600,
       "refresh_token": "tGzv3JOkF0XG5Qx2TlKWIA",
       "id_token": "<JWT if present>"
     }
     ```

     Now we have the access token.

4. **Access Protected Resource**:

   * Use the access token in an authorized request:

     ```bash
     curl -H "Authorization: Bearer 2YotnFZFEjr1zCsicMWpAA" http://localhost:8080/hello
     ```

     If the token is valid, our server's protected handler will respond with:

     ```
     Hello wesen, your request to MCP is authorized!
     ```

     If the token was missing or wrong, you'd get `401 Unauthorized`. If you intentionally test scope restrictions, you might get `403 Forbidden` for missing scope.

This sequence confirms that:

* The dynamic client registration allowed a new client to be created without prior config.
* The user "wesen" could log in and authorize the client.
* The client obtained a valid token and accessed the MCP server's resource on behalf of "wesen".

## Conclusion and Best Practices

We have implemented a basic OIDC-compliant Authorization Server with dynamic client registration using Ory Fosite. This solution addresses the needs of an MCP server acting as its own auth server, allowing dynamic onboarding of clients and secure user authorization.

A few important points and best practices to note:

* **Security Measures**: Always enforce PKCE for public clients (we did). Use HTTPS in real deployments (our example uses http for simplicity). Use secure random generators for tokens and client secrets. The MCP spec also suggests limiting token lifetimes and implementing rotation for refresh tokens.
* **Dynamic Registration Management**: In a full implementation, consider issuing a `registration_access_token` and `registration_client_uri` in the `/register` response. This allows the client to update or delete its registration. In our simple server, the registration is permanent unless the server admin removes it manually.
* **User Consent and Scope**: We automatically granted all requested scopes. In practice, implement a consent page so users can approve what access the client is requesting (especially for sensitive scopes). For example, an MCP client might request scopes like "read\:tools" or "execute\:commands", and the user should consent.
* **Persistent Storage**: We used in-memory maps, which means restarting the server loses all clients and tokens. For a more realistic setup, use a database (SQLite or others). Fosite's storage interface can be implemented with a DB, as shown by others using Gorm/SQLite. That would allow clients to remain registered across restarts.
* **MCP Considerations**: The MCP spec positions the server as both Authorization and Resource Server, which is what we did. It also allows delegating auth to an external server, but that still requires the MCP server to issue its own tokens in the end. Our implementation is self-contained, but if one wanted, they could integrate an external IdP (like using Ory Hydra or another OAuth provider) and then just validate tokens at the MCP server.
* **OAuth2.1 Draft Alignment**: OAuth 2.1 (in draft) is basically OAuth2 + PKCE + security recommendations. Our server follows those (requiring PKCE, not using implicit flow, etc.). We also expose the `.well-known/oauth-authorization-server` metadata endpoint in a real scenario so that clients can discover the URLs instead of using defaults. For brevity, we did not implement the metadata endpoint here, but one should for full compliance (it’s a static JSON listing the auth URL, token URL, registration URL, supported scopes, etc.).

By following this guide, you can implement a lightweight OAuth2 Authorization Server tailored for MCP or similar use-cases. The use of Fosite ensures that much of the tricky parts (validating requests, forming responses, security checks) are handled according to spec. You can extend this foundation with more features as needed, such as support for additional grant types, refresh token rotation, logout and revocation endpoints, and so on.

With this "full book" overview and step-by-step implementation, you should have a solid starting point for enabling secure, dynamic client registration and OAuth2 login in your MCP server. Happy coding!


