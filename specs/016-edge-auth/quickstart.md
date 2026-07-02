# Quickstart: Edge Auth

## Extra Headers (any edge proxy)

```go
base := v1.StaticToken("my-app-token")
auth, err := v1.WithExtraHeaders(base, map[string]string{
    "x-custom-proxy-key": "proxy-secret",
    "x-tenant-id":        "acme-corp",
})
client, err := v1.NewClient(v1.Config{
    Address: "gateway.example.com:443",
    Auth:    auth,
})
```

## Cloudflare Access

```go
base := v1.StaticToken("my-app-token")
auth, err := edge.CloudflareAccess(base, "cf-edge-jwt-token")
client, err := v1.NewClient(v1.Config{
    Address: "gateway.example.com:443",
    Auth:    auth,
})
```

## WebSocket Tunnel

```go
tunnel, err := edge.NewTunnelProxy(
    "wss://gateway.example.com/ws",
    "cf-edge-jwt-token",
    edge.WithTunnelTLS(&tls.Config{}),
    edge.WithTunnelLogger(myLogger),
)
defer tunnel.Close()

client, err := v1.NewClient(v1.Config{
    Address: tunnel.Addr(),
    Auth:    auth,
    TLS:     &v1.TLSConfig{Insecure: true}, // local tunnel, no TLS
})
```
