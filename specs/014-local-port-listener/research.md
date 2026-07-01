# Research: Local Port Listener

**Date**: 2026-06-30
**Feature**: 014-local-port-listener

## R1: net.Listener Contract Requirements

**Decision**: The listener implementation must satisfy Go's `net.Listener` interface exactly: `Accept() (net.Conn, error)`, `Close() error`, `Addr() net.Addr`.

**Rationale**: Returning `net.Listener` enables consumers to use the listener with `http.Serve`, `grpc.NewServer`, and any framework that accepts a `net.Listener`. The interface is small (3 methods) and well-defined.

**Alternatives considered**:
- Custom listener interface: rejected because it breaks compatibility with the Go ecosystem.
- Wrapping net.Listener with extra methods: rejected for v1; connection tracking can be added later via options without changing the return type.

## R2: Internal Tunnel Setup per Accept

**Decision**: Each `Accept()` call that receives a local connection internally calls `tcpClient.Forward()` (or `sshClient.Tunnel()` when `WithSSHTunnel()` is set) to create a new gRPC stream, then bridges the local `net.Conn` to the `io.ReadWriteCloser` returned by Forward/Tunnel.

**Rationale**: Reuses the existing, tested Forward/Tunnel plumbing. Each connection gets its own gRPC stream, providing isolation. The listener struct needs access to the tcpClient (and optionally sshClient) internals, which means the `Listen` method lives on `tcpClient` as a concrete method, and `TCPInterface` gains the `Listen` signature.

**Alternatives considered**:
- Multiplexing connections over a single gRPC stream: rejected because the proto API uses one stream per connection (ForwardTcp is a bidirectional stream per call).
- Spawning a standalone goroutine per connection: this is the chosen approach, each accepted connection gets a bridge goroutine.

## R3: Bridge Goroutine Pattern

**Decision**: Each accepted connection spawns two goroutines: one copying local-to-remote, one copying remote-to-local. When either direction encounters an error or EOF, both the local `net.Conn` and the `io.ReadWriteCloser` from Forward/Tunnel are closed.

**Rationale**: Standard Go pattern for bidirectional stream bridging (used in `io.Copy` based proxies). The Forward/Tunnel connection already handles its own gRPC stream lifecycle, so the bridge just needs to relay bytes and close both ends on termination.

**Alternatives considered**:
- Single goroutine with select: not feasible because `io.Copy` is blocking; two goroutines is the standard approach.

## R4: Connection Tracking for Clean Shutdown

**Decision**: The listener maintains a `sync.WaitGroup` tracking all active bridge goroutines. `Close()` closes the underlying `net.Listener`, waits on the WaitGroup, and returns. A `sync.Once` guards against double-close.

**Rationale**: WaitGroup is the simplest mechanism for "wait until all connections are done." Combined with closing the net.Listener (which causes Accept to return an error), this provides clean shutdown with deterministic resource cleanup.

**Alternatives considered**:
- Channel-based tracking: more complex with no benefit over WaitGroup for this use case.
- Context-only cancellation: insufficient because we need to wait for goroutines to finish, not just signal them.

## R5: Context Cancellation Integration

**Decision**: The context passed to `Listen()` is used to derive per-connection contexts. Cancelling the parent context triggers listener close (via a background goroutine watching the context). Each connection's Forward/Tunnel call receives a child context derived from the listener's context.

**Rationale**: Aligns with FR-010 (context cancellation must close the listener). The parent context flows naturally into each Forward/Tunnel call, so gRPC streams respect cancellation.

**Alternatives considered**:
- Ignoring the context after listener creation: rejected because it violates Go conventions and FR-010.

## R6: ListenOption Pattern

**Decision**: Follow the existing `ForwardOption`/`TunnelOption` pattern: `ListenOption` is a `func(*listenConfig)`, with exported constructors `WithBindAddress(addr string)`, `WithSSHTunnel()`, and `WithListenServiceID(id string)`.

**Rationale**: Consistent with the SDK's established functional options pattern (`ForwardOption`, `TunnelOption`). The `listenConfig` struct is unexported; options are the public API.

**Alternatives considered**:
- Struct-based config: rejected for inconsistency with existing SDK patterns.

## R7: Fake Implementation Strategy

**Decision**: Add a `Listen` method to `fakeTCPClient` that validates sandbox name (non-empty) and port (1-65535), then returns `Unimplemented`. This matches the fake Tunnel pattern.

**Rationale**: Constitution XI (Fake-Real Parity) requires fakes to mirror client-side validation. The review-spec gate already updated FR-011 to require input validation before returning Unimplemented.

**Alternatives considered**:
- Functional fake that returns a real net.Listener backed by in-memory pipes: deferred to future iteration; adds significant complexity.

## R8: File Organization

**Decision**: The `Listen` method and `listenConfig`/`ListenOption` types go in `openshell/v1/tcp.go` (interface + options) and `openshell/v1/tcp_client.go` (implementation). The internal `tunnelListener` struct lives in `tcp_client.go` alongside the existing `tcpForwardConn`. The fake goes in `openshell/v1/fake/tcp.go`.

**Rationale**: Follows the existing pattern where `tcp.go` defines the interface and options, `tcp_client.go` holds the implementation, and `fake/tcp.go` holds the fake.

**Alternatives considered**:
- Separate `listen.go` file: rejected because the listener is part of the TCP sub-client, not a standalone concept.
