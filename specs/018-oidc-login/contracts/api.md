# API Contract: openshell/v1/oidc

**Package**: `github.com/rhuss/openshell-sdk-go/openshell/v1/oidc`

## Public Functions

### Login

```go
func Login(ctx context.Context, gatewayName string, opts ...LoginOption) (*oauth2.Token, error)
```

Performs an interactive OIDC login. If `gatewayName` is non-empty, reads OIDC config from gateway metadata and persists tokens to disk. If empty, requires `WithIssuer` and `WithClientID` options.

**Flow**: Auth Code + PKCE (browser), falling back to keyboard flow if browser open fails or `WithKeyboardFlow()` is set.

**Errors**: `ErrDiscovery`, `ErrAuthCode`, `ErrTimeout`, `ErrCallbackServer`, `ErrTokenPersist`, `ErrOIDCConfig`

### DeviceLogin

```go
func DeviceLogin(ctx context.Context, opts ...LoginOption) (*oauth2.Token, error)
```

Performs device code flow (RFC 8628). Requires `WithIssuer` and `WithClientID` (or a gateway name via `WithGateway`). Displays verification URL and user code, polls until authorized.

**Errors**: `ErrDiscovery`, `ErrDeviceCode`, `ErrTimeout`

### ClientCredentials

```go
func ClientCredentials(ctx context.Context, opts ...LoginOption) (*oauth2.Token, error)
```

Performs client credentials grant. Requires `WithIssuer`, `WithClientID`, and `WithClientSecret` (or `WithGateway` combined with `WithClientSecret`). No user interaction. Default interactive scopes (openid, profile, email) are not sent unless explicitly set via `WithScopes`.

**Errors**: `ErrDiscovery`, `ErrClientCredentials`, `ErrOIDCConfig`

## LoginOption Type

```go
type LoginOption func(*loginConfig)
```

### Option Functions

| Function | Description |
|----------|-------------|
| `WithIssuer(url string)` | Set OIDC issuer URL (required for standalone flows) |
| `WithClientID(id string)` | Set OAuth2 client ID (required for standalone flows) |
| `WithClientSecret(secret string)` | Set client secret (required for client credentials) |
| `WithScopes(scopes ...string)` | Override default scopes (openid, profile, email) |
| `WithCallbackPort(port int)` | Set fixed callback port (default: auto 8000/18000) |
| `WithTimeout(d time.Duration)` | Set interactive flow timeout (default: 2 minutes) |
| `WithKeyboardFlow()` | Force keyboard flow (skip browser open) |
| `WithInMemory()` | Skip disk token persistence |
| `WithDisplayFunc(fn func(verificationURL, userCode string))` | Custom device code display |
| `WithGateway(name string)` | Set gateway name for DeviceLogin/ClientCredentials |

## Error Sentinels

```go
var (
    ErrDiscovery         // OIDC discovery fetch or parse failed
    ErrAuthCode          // Authorization code exchange failed
    ErrDeviceCode        // Device code flow failed
    ErrClientCredentials // Client credentials exchange failed
    ErrTimeout           // Interactive flow timed out
    ErrCallbackServer    // Localhost callback server failed to start
    ErrTokenPersist      // Token disk write failed
    ErrOIDCConfig        // Gateway metadata missing OIDC fields
)
```

## Gateway Config Extension

Two new optional fields in `metadata.json` (backward-compatible):

```json
{
    "endpoint": "gateway.example.com:443",
    "auth_mode": "oidc",
    "name": "my-gateway",
    "oidc_issuer": "https://auth.example.com",
    "oidc_client_id": "openshell-cli"
}
```
