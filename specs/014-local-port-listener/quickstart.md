# Quickstart: Local Port Listener

**Date**: 2026-06-30
**Feature**: 014-local-port-listener

## Basic Usage

```go
// Bind local port 8080 to sandbox port 80
listener, err := client.TCP().Listen(ctx, "my-sandbox", 80, 8080)
if err != nil {
    log.Fatal(err)
}
defer listener.Close()

fmt.Println("Listening on", listener.Addr())

// Accept drives the tunnel: each accepted connection is automatically
// bridged to sandbox port 80. The caller manages lifecycle (Close),
// not data transfer (the bridge goroutines handle read/write).
for {
    conn, err := listener.Accept()
    if err != nil {
        break // listener closed
    }
    _ = conn // bridge runs automatically; close conn to tear down
}
```

## Ephemeral Port

```go
// Use port 0 for OS-assigned port
listener, err := client.TCP().Listen(ctx, "my-sandbox", 80, 0)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Assigned port:", listener.Addr())
```

## Transparent Proxy

```go
// The listener acts as a transparent proxy: external clients connecting
// to the local port are bridged to the sandbox. The caller holds the
// listener open and drives Accept; the bridge handles data transfer.
listener, err := client.TCP().Listen(ctx, "my-sandbox", 80, 8080)
if err != nil {
    log.Fatal(err)
}
defer listener.Close()

// Run accept loop in background
go func() {
    for {
        if _, err := listener.Accept(); err != nil {
            return
        }
    }
}()

// Block until shutdown signal
<-ctx.Done()
```

## SSH Tunnel Transport

```go
listener, err := client.TCP().Listen(ctx, "my-sandbox", 80, 8080,
    v1.WithSSHTunnel(),
)
```

## Custom Bind Address

```go
listener, err := client.TCP().Listen(ctx, "my-sandbox", 80, 8080,
    v1.WithBindAddress("0.0.0.0"),
)
```

## Testing with Fakes

```go
fc := fake.NewClient()
_, err := fc.TCP().Listen(ctx, "sandbox", 80, 8080)
// Returns Unimplemented error (expected in tests)
```
