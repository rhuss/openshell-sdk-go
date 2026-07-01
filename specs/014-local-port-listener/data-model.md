# Data Model: Local Port Listener

**Date**: 2026-06-30
**Feature**: 014-local-port-listener

## Entities

### listenConfig (unexported)

Accumulates options for the Listen method. Follows the same pattern as
`forwardConfig` and `tunnelConfig`.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| bindAddress | string | "127.0.0.1" | Local address to bind the listener to |
| useSSHTunnel | bool | false | Use SSH tunnel transport instead of TCP forward |
| serviceID | string | "" | Optional service identifier for audit/correlation |

### ListenOption (exported)

Function type: `func(*listenConfig)`. Public constructors:

- `WithBindAddress(addr string) ListenOption`
- `WithSSHTunnel() ListenOption`
- `WithListenServiceID(id string) ListenOption`

### tunnelListener (unexported)

Implements `net.Listener`. Manages the local TCP listener and bridges
accepted connections to sandbox ports via Forward or Tunnel.

| Field | Type | Description |
|-------|------|-------------|
| inner | net.Listener | The underlying TCP listener bound to the local address |
| ctx | context.Context | Parent context for the listener lifecycle |
| cancel | context.CancelFunc | Cancels the listener context |
| tcp | *tcpClient | Reference to the TCP client for Forward calls |
| ssh | SSHInterface | Reference to the SSH client for Tunnel calls (nil when not using SSH) |
| sandboxName | string | Target sandbox name |
| remotePort | uint32 | Target port inside the sandbox |
| cfg | listenConfig | Resolved listen configuration |
| wg | sync.WaitGroup | Tracks active bridge goroutines |
| closeOnce | sync.Once | Ensures Close is idempotent |
| closeErr | error | Stored error from Close |

### Relationships

```
TCPInterface
  └── Listen() returns net.Listener (tunnelListener)
        ├── Accept() → net.Conn + bridge goroutines
        │     └── internally calls Forward() or Tunnel()
        │           └── returns io.ReadWriteCloser (bridged to net.Conn)
        ├── Close() → stops accepting, waits for goroutines
        └── Addr() → bound local address
```

### Lifecycle

1. **Creation**: `Listen()` validates inputs, creates `net.Listen("tcp", addr)`,
   wraps it in `tunnelListener`, starts a context-watcher goroutine.
2. **Active**: Each `Accept()` call blocks on `inner.Accept()`. On success,
   calls `Forward()`/`Tunnel()` to establish the remote end, then spawns
   bridge goroutines. On tunnel setup failure, closes the local conn and
   retries (Accept returns only successful connections).
3. **Shutdown**: `Close()` closes `inner` (unblocks Accept), cancels context,
   waits on WaitGroup for all bridge goroutines to finish.
