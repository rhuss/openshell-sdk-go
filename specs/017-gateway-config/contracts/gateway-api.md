# Public API Contract: openshell/v1/gateway

## Package

```go
package gateway // import "github.com/rhuss/openshell-sdk-go/openshell/v1/gateway"
```

## Types

### Config

```go
type Config struct {
    Name     string       // Validated gateway name
    Endpoint string       // Host:port of the gateway
    AuthMode AuthMode     // Resolved auth mode
    Source   ConfigSource // User or System origin
    Dir      string       // Absolute path to gateway config directory
}
```

### Info

```go
type Info struct {
    Name   string       // Gateway name from directory listing
    Active bool         // Whether this is the active gateway
    Source ConfigSource // User or System origin
}
```

### AuthMode

```go
type AuthMode string

const (
    AuthModeNone          AuthMode = ""
    AuthModePlaintext     AuthMode = "plaintext"
    AuthModeCloudflareJWT AuthMode = "cloudflare_jwt"
    AuthModeOIDC          AuthMode = "oidc"
    AuthModeMTLS          AuthMode = "mtls"
)
```

### ConfigSource

```go
type ConfigSource string

const (
    SourceUser   ConfigSource = "user"
    SourceSystem ConfigSource = "system"
)
```

### ClientOption

```go
type ClientOption func(*clientConfig)

func WithLogger(l types.Logger) ClientOption
func WithTimeout(d time.Duration) ClientOption
func WithTLS(cfg *types.TLSConfig) ClientOption
func WithAuth(provider types.AuthProvider) ClientOption
func WithRetryPolicy(p *types.RetryPolicy) ClientOption
```

## Functions

### NewClient

```go
func NewClient(name string, opts ...ClientOption) (*v1.Client, error)
```

Creates a fully configured SDK client for the named gateway. If name
is empty, resolves the active gateway. Loads gateway config, resolves
auth provider from auth_mode, and calls v1.NewClient internally.

**Errors**: ErrGatewayNotFound, ErrConfigParse, ErrTokenLoad,
ErrUnsupportedAuthMode, ErrInvalidGatewayName, ErrNoActiveGateway

### LoadConfig

```go
func LoadConfig(name string) (*Config, error)
```

Parses gateway configuration without creating a client. If name is
empty, resolves the active gateway. Returns a frozen snapshot.

**Errors**: ErrGatewayNotFound, ErrConfigParse, ErrInvalidGatewayName,
ErrNoActiveGateway

### ListGateways

```go
func ListGateways() ([]Info, error)
```

Enumerates all available gateways from user and system directories.
Returns empty slice (not error) when no gateways are configured.
User gateways appear before system gateways. Active status is resolved.

**Errors**: Only on unexpected I/O errors reading directories.

## Errors

```go
var (
    ErrGatewayNotFound    = errors.New("gateway: not found")
    ErrConfigParse        = errors.New("gateway: config parse error")
    ErrTokenLoad          = errors.New("gateway: token load error")
    ErrUnsupportedAuthMode = errors.New("gateway: unsupported auth mode")
    ErrInvalidGatewayName = errors.New("gateway: invalid gateway name")
    ErrNoActiveGateway    = errors.New("gateway: no active gateway")
)
```

All errors wrap the sentinel with `fmt.Errorf("...: %w", err)` so
callers can use `errors.Is(err, gateway.ErrGatewayNotFound)`.

## Thread Safety

All exported functions are safe for concurrent use from multiple
goroutines. Token loading uses sync.Once internally.
