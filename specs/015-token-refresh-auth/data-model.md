# Data Model: Token Refresh with Coalesced Caching

## Types

### AuthProvider (existing, unchanged)

```
AuthProvider
  GetRequestMetadata(ctx, uri...) -> (metadata map, error)
  RequireTransportSecurity() -> bool
```

Defined in `openshell/v1/types/auth.go`. Wraps gRPC `credentials.PerRPCCredentials`.

### RefreshOption (new)

```
RefreshOption = func(*refreshConfig)
```

Functional option pattern. Available options:
- `WithLeeway(d time.Duration)` - refresh buffer before expiry (default: 10s)
- `WithLogger(l types.Logger)` - logger for stale-token warnings (default: nil, silent)

### refreshConfig (internal)

```
refreshConfig
  leeway   time.Duration  (default: 10s)
  logger   types.Logger   (default: nil)
```

Collects resolved options before constructing `refreshableAuth`.

### refreshableAuth (internal)

```
refreshableAuth
  source   oauth2.TokenSource   // underlying token provider
  mu       sync.RWMutex         // guards tok and tokTime
  tok      *oauth2.Token        // cached token (nil until first fetch)
  leeway   time.Duration        // refresh buffer
  logger   types.Logger         // optional warning logger
```

Implements `AuthProvider`. Lifecycle:
1. Constructed by `RefreshableToken()` with no cached token
2. First `GetRequestMetadata` call fetches from source
3. Subsequent calls return cached token if valid (fast path under RLock)
4. When expired/within leeway, acquires Lock, re-checks, fetches if still stale
5. On fetch error with existing cached token: returns stale, logs warning
6. On fetch error with no cached token: returns error

### oauth2.Token (external, from golang.org/x/oauth2)

```
Token
  AccessToken   string
  TokenType     string
  RefreshToken  string
  Expiry        time.Time
```

Used as-is from the oauth2 package. The SDK reads `AccessToken` and `Expiry` only.

### oauth2.TokenSource (external, from golang.org/x/oauth2)

```
TokenSource
  Token() -> (*Token, error)   // must be safe for concurrent use
```

The interface callers implement. The SDK wraps it with caching.

## Constructors

### RefreshableToken

```
RefreshableToken(src oauth2.TokenSource, opts ...RefreshOption) (AuthProvider, error)
```

- Returns error if `src` is nil
- Applies functional options to build refreshConfig
- Returns a `*refreshableAuth` implementing AuthProvider

### Existing (unchanged)

```
NoAuth() AuthProvider
StaticToken(token string) AuthProvider
```

## Relationships

```
RefreshableToken(src, opts...)
       │
       ▼
  refreshableAuth ──implements──▶ AuthProvider
       │                              │
       │ wraps                        │ used by
       ▼                              ▼
  oauth2.TokenSource            grpc.ClientConn
```
