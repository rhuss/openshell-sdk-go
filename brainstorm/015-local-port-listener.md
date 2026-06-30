# Brainstorm: Local Port Listener

**Date:** 2026-06-30
**Status:** active

## Problem Framing

The SDK provides `TCP().Forward()` and `SSH().Tunnel()` which return an
`io.ReadWriteCloser` per connection. Callers who want `ssh -L` style
local port binding (bind localhost:8080, tunnel every accepted connection
to sandbox port 80) must write the accept loop, bridge goroutines, and
teardown logic themselves. This is repetitive boilerplate for operators,
MCP tool wrappers, CLI utilities, and developers running locally.

**Origin:** Deferred from brainstorm #008 (ssh-tcp-config) and #010
(ssh-tunnel-forward-opts). Captured in idea-inbox as `local-port-listener`.

## Approaches Considered

### A: Minimal Listener on TCPInterface (Chosen)

Add `Listen()` to `TCPInterface`. Each `Accept()` internally calls
`Forward()` (or `Tunnel()` with `WithSSHTunnel()`). Returns a standard
`net.Listener`. The listener goroutine accepts local connections and
bridges each to a new gRPC stream. `Close()` stops accepting and tears
down all active connections.

- Pros: Familiar `net.Listener` contract. Minimal new types. Reuses
  existing `Forward()`/`Tunnel()` plumbing.
- Cons: Each accepted connection is a new gRPC stream. Under high
  connection rate this could be chatty.

### B: Listener with Connection Tracking

Same as A, but adds `ActiveConnections() int` and an optional
`WithMaxConnections(n)` to cap concurrency.

- Pros: Visibility into connection state. Prevents runaway connections.
- Cons: Extra complexity for a v1. Can be added later as a non-breaking
  option.

### C: Higher-Level PortForwarder Type

Standalone `PortForwarder` struct that composes `TCPInterface` and
`SSHInterface`. Provides `Listen()` plus helpers like `ForwardAndServe()`.

- Pros: Richer API without polluting the core interface.
- Cons: New type to discover. Doesn't follow the sub-client pattern.
  Can't be faked via the existing interface.

## Decision

**Approach A: Minimal Listener.** Simplest thing that works, follows the
existing sub-client pattern. Connection tracking (B) can be added later
as non-breaking options.

## Key Requirements

- **Method**: `TCP().Listen(ctx, sandboxName, remotePort, localPort, ...ListenOption) (net.Listener, error)`
- **Transport**: TCP Forward by default, `WithSSHTunnel()` option for
  per-connection SSH session auth
- **Local port**: Caller-specified. `localPort=0` means OS picks a random
  available port; caller reads `Addr()` to discover it
- **Bind address**: Defaults to `127.0.0.1`, overridable via
  `WithBindAddress("0.0.0.0")`
- **Connection lifecycle**: Each `Accept()` spawns a new `Forward()` (or
  `Tunnel()`) stream internally. Caller gets `net.Conn`. `Listener.Close()`
  tears down all active connections
- **Interface**: Added to `TCPInterface` (not standalone)
- **Fake**: Returns `Unimplemented` error (same pattern as fake `Tunnel()`)
- **Options**: `ListenOption` type with `WithBindAddress()`,
  `WithSSHTunnel()`, `WithListenServiceID()`

## Open Questions

- Should `Listen` propagate per-connection errors (e.g., failed Forward)
  to the caller, or silently retry? Standard `net.Listener` returns errors
  from `Accept()`, but a transient gRPC failure on one connection shouldn't
  kill the listener.
- Should there be a `WithOnError(func(error))` callback for logging
  connection-level failures without stopping the listener?
- Connection limit: deferred to future iteration, but the internal design
  should make it easy to add `WithMaxConnections()` later.
