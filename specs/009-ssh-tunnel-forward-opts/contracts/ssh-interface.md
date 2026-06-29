# Contract: SSHInterface

## Current

```go
type SSHInterface interface {
    CreateSession(ctx context.Context, sandboxID string) (*SSHSession, error)
    RevokeSession(ctx context.Context, token string) (bool, error)
}
```

## Proposed

```go
type SSHInterface interface {
    CreateSession(ctx context.Context, sandboxID string) (*SSHSession, error)
    RevokeSession(ctx context.Context, token string) (bool, error)
    Tunnel(ctx context.Context, sandboxName string, port uint32, opts ...TunnelOption) (io.ReadWriteCloser, error)
}
```

## New Types

```go
type TunnelOption func(*tunnelConfig)

func WithTunnelServiceID(id string) TunnelOption
```

## Error Codes

| Method | Error Code | Condition |
|--------|-----------|-----------|
| Tunnel | InvalidArgument | port == 0 or port > 65535, or sandboxName is empty |
| Tunnel | NotFound | sandbox name does not resolve to a sandbox |
| Tunnel | Unimplemented | fake client |
| Tunnel | Unavailable | client is closed |
| Tunnel | (upstream) | Any error from CreateSession or ForwardTcp propagated |
