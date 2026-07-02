# Data Model: Edge Auth

## Entities

### extraHeadersAuth (internal, core SDK)

Decorator wrapping any `AuthProvider` with additional static headers.

| Field | Type | Description |
|-------|------|-------------|
| base | AuthProvider | Wrapped auth provider |
| headers | map[string]string | Extra headers (keys lowercase-normalized) |

**Relationships**: Decorates any AuthProvider. Returned by `WithExtraHeaders()`.

**Validation**: `base` must not be nil. `headers` must not be nil or empty. Empty-string values are filtered at construction time.

### TunnelProxy (public, edge package)

WebSocket tunnel that bridges gRPC connections over WebSocket.

| Field | Type | Description |
|-------|------|-------------|
| listener | net.Listener | Local TCP listener for gRPC client to dial |
| gatewayURL | string | Remote gateway WebSocket endpoint |
| edgeToken | string | Edge proxy authentication token |
| logger | types.Logger | Optional structured logger |
| closeTimeout | time.Duration | Max wait for drain on Close (default 5s) |
| tlsConfig | *tls.Config | Optional TLS for WebSocket (wss://) |
| wg | sync.WaitGroup | Tracks in-flight bridge goroutines |
| closing | bool | Set on Close to reject new connections |

**Lifecycle**: Created (NewTunnelProxy) -> Active (accepting connections) -> Closing (draining) -> Closed.

**State transitions**:
- `Created -> Active`: Implicit on first Accept via Addr()
- `Active -> Closing`: Close() called, stops accepting, starts drain
- `Closing -> Closed`: All bridges drained or timeout reached

### TunnelOption (public, edge package)

Functional option for configuring TunnelProxy.

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| WithTunnelLogger | types.Logger | nil | Structured logger for tunnel events |
| WithTunnelTLS | *tls.Config | nil | TLS config for wss:// connections |
| WithCloseTimeout | time.Duration | 5s | Max drain time on Close |

## Relationships

```
AuthProvider (interface)
  ├── NoAuth (existing)
  ├── StaticToken (existing)
  ├── RefreshableToken (existing)
  └── extraHeadersAuth (new, wraps any of the above)
        └── CloudflareAccess() returns extraHeadersAuth with CF-specific headers

TunnelProxy (standalone, not an AuthProvider)
  ├── owns: net.Listener (local)
  ├── dials: WebSocket (remote gateway)
  └── bridges: goroutine-per-connection (read/write copy loops)
```
