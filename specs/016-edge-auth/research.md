# Research: Edge Auth

## WebSocket Library Selection

**Decision**: `github.com/coder/websocket` (v2)

**Rationale**: Pure Go, minimal dependencies, supports `context.Context` natively, binary frame support, actively maintained. Used by Grafana, CockroachDB, and other production Go projects. No CGO requirement.

**Alternatives considered**:
- `gorilla/websocket`: Popular but archived (2022), no context support, larger API surface.
- `golang.org/x/net/websocket`: Part of x/net but lacks binary frame control and modern features.
- `gobwas/ws`: Zero-allocation but lower-level, requires manual frame handling.

**Constitution V compliance**: New dependency justified because the stdlib has no WebSocket support. The dependency is isolated to the `openshell/v1/edge/` package and does not affect the core module. SDK consumers who don't import the edge package take no transitive dependency.

## Header Merge Strategy

**Decision**: Lowercase normalization + map merge with extra-wins precedence.

**Rationale**: HTTP/2 (RFC 9113 Section 8.2) mandates lowercase field names. gRPC metadata keys are also lowercase. Normalizing to lowercase before merge prevents case-variant duplicates and matches the transport layer's behavior.

**Implementation**: `GetRequestMetadata` on the wrapper calls the base provider first, then overwrites with extra headers (both normalized to lowercase). This is a simple map merge in a single pass.

## Tunnel Protocol

**Decision**: Raw binary WebSocket frames carrying HTTP/2 bytes (byte-stream tunneling).

**Rationale**: Matches the Rust CLI's `EdgeTunnelProxy` implementation. The WebSocket connection acts as a transparent byte pipe. No gRPC-specific sub-protocol negotiation is required. This is the simplest approach that works with all edge proxies.

**Implementation**: The tunnel proxy listens on a local address, accepts gRPC connections, and for each connection opens a WebSocket to the gateway. Binary frames carry raw bytes bidirectionally. The gRPC client dials the local listener address instead of the remote gateway.

## Tunnel Lifecycle

**Decision**: Goroutine-per-connection with graceful drain on Close.

**Rationale**: Matches Rust implementation. Each accepted connection spawns two goroutines (read/write copy loops). Close sets a closing flag, stops accepting, cancels the listener context, and waits for in-flight bridges with a configurable timeout (default 5s). After timeout, force-close.

**Pattern reference**: The existing `tunnelListener` in `tcp_client.go` uses the same pattern (sync.WaitGroup + closeOnce + context cancellation). The edge tunnel proxy follows this established pattern.

## Package Layout

**Decision**: `WithExtraHeaders` in core `openshell/v1/`, everything else in `openshell/v1/edge/`.

**Rationale**: The header wrapper is a generic decorator that any auth provider can use, so it belongs in the core package alongside `NoAuth`, `StaticToken`, and `RefreshableToken`. The Cloudflare convenience constructor and WebSocket tunnel proxy are vendor/transport-specific and belong in the optional edge package. This keeps the core module dependency-free and makes edge functionality opt-in.
