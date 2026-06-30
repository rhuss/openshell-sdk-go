# Research: SSH Tunnel Context Cancellation Cleanup

## R1: Goroutine cleanup patterns in Go

**Decision**: Use a goroutine that blocks on a `chan struct{}` and calls
Close() when it fires.

**Rationale**: This is the standard Go pattern for cleanup goroutines.
The `done` channel is already closed by `readLoop` when it exits, making
it the natural signal. The `closeOnce` in `sshTunnel.Close()` handles
the race between the cleanup goroutine and explicit Close() calls.

**Alternatives considered**:
- `runtime.SetFinalizer`: Unreliable, GC-dependent, not suitable for
  resource cleanup that must be timely.
- `context.AfterFunc` (Go 1.21+): Could work but adds complexity since
  we need to wait for readLoop to finish, not just context cancellation.

## R2: Race safety with closeOnce

**Decision**: No new synchronization needed. The existing `sync.Once`
in `sshTunnel.Close()` is sufficient.

**Rationale**: `sync.Once.Do` guarantees that the closure runs exactly
once, regardless of how many goroutines call it concurrently. Both the
cleanup goroutine and explicit Close() calls go through `closeOnce.Do`,
so the revocation happens exactly once.

## R3: Constitution XII compliance (Graceful Shutdown Order)

**Decision**: The cleanup goroutine calls `sshTunnel.Close()`, which
already implements the correct shutdown order (protocol close before
context cancel). No change needed to the shutdown sequence.

**Rationale**: Constitution XII requires protocol-level close before
context cancellation. `sshTunnel.Close()` does:
1. `tcpForwardConn.Close()` (calls `CloseSend()`, then `cancel()`)
2. `revokeFunc()` (best-effort RPC with `context.Background()`)

The cleanup goroutine reuses this path, so the order is preserved.
