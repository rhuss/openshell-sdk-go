# Feature Specification: Edge Auth (Extra Headers + WebSocket Tunnel)

**Feature Branch**: `016-edge-auth`
**Created**: 2026-07-02
**Status**: Review
**Input**: Brainstorm `brainstorm/018-edge-auth.md`, Issue #20

## User Scenarios & Testing

### User Story 1 - Attach Extra Headers to Any Auth Provider (Priority: P1)

A developer using the Go SDK needs to connect to a gateway that sits behind a zero-trust reverse proxy (e.g., Cloudflare Access, Google IAP, Zscaler). The proxy requires specific headers on every request in addition to the standard application-level bearer token. The developer wraps their existing auth provider with a generic header decorator that adds the required edge headers to every RPC call.

**Why this priority**: This is the foundational mechanism that all edge auth scenarios depend on. Without generic header layering, no edge proxy integration is possible.

**Independent Test**: Can be fully tested by creating a wrapped auth provider with custom headers and verifying that both the base provider's authorization header and the extra headers appear on outgoing RPCs.

**Acceptance Scenarios**:

1. **Given** an existing auth provider (StaticToken, RefreshableToken, or NoAuth), **When** the developer wraps it with extra headers containing edge proxy credentials, **Then** every RPC includes both the base authorization header and all extra headers.
2. **Given** an extra header has the same key as a base provider header, **When** the RPC is sent, **Then** the extra header value takes precedence over the base provider's value for that key.
3. **Given** a wrapped auth provider, **When** the transport security check is queried, **Then** the response delegates to the base provider's security requirements.

---

### User Story 2 - Cloudflare Access Convenience (Priority: P2)

A developer connecting through Cloudflare Access wants a simple, purpose-built constructor rather than manually specifying CF-specific header names and cookie formats. The convenience constructor accepts a base auth provider and a CF edge token, and produces a properly configured auth provider with the correct Cloudflare headers.

**Why this priority**: Cloudflare Access is the most common zero-trust proxy in the target deployment environment. A convenience constructor prevents header-name typos and formats the cookie correctly.

**Independent Test**: Can be tested by creating a Cloudflare Access auth provider and verifying that the `cf-access-jwt-assertion` header and `CF_Authorization` cookie are set with the correct values.

**Acceptance Scenarios**:

1. **Given** a base auth provider and a Cloudflare edge token, **When** the developer uses the Cloudflare Access constructor, **Then** the resulting auth provider sends `cf-access-jwt-assertion` and `cookie: CF_Authorization=<token>` headers alongside the base auth headers.
2. **Given** an empty edge token, **When** the developer creates a Cloudflare Access auth provider, **Then** creation fails immediately with a clear error message indicating the edge token is required.

---

### User Story 3 - WebSocket Tunnel for gRPC Behind Edge Proxies (Priority: P3)

A developer needs to connect to a gateway behind an edge proxy that rejects standard HTTP/2 POST requests (the transport gRPC uses). The developer creates a WebSocket tunnel proxy that bridges the gRPC connection over a WebSocket, allowing gRPC traffic to pass through the proxy. The tunnel carries its own edge token for proxy authentication, independent of the application-level auth provider.

**Why this priority**: Required for environments where the edge proxy does not support HTTP/2 passthrough. Less common than header-only scenarios, but essential for those deployments.

**Independent Test**: Can be tested by creating a tunnel proxy pointed at a test WebSocket endpoint and verifying that gRPC calls are forwarded through the WebSocket connection.

**Acceptance Scenarios**:

1. **Given** a gateway URL and edge token, **When** the developer creates a tunnel proxy, **Then** gRPC calls are routed through a WebSocket connection to the gateway.
2. **Given** an active tunnel proxy with in-flight connections, **When** the developer calls Close, **Then** in-flight connections are drained gracefully before the tunnel shuts down, and all goroutines are cleaned up. If draining exceeds the configured timeout (default 5 seconds), remaining connections are force-closed.
3. **Given** a tunnel proxy, **When** TLS is configured for the WebSocket connection, **Then** the tunnel connects over `wss://` instead of `ws://`.
4. **Given** an invalid gateway URL, **When** the developer creates a tunnel proxy, **Then** creation fails with a descriptive error rather than silently connecting to a wrong endpoint.

---

### Edge Cases

- What happens when the base auth provider returns an error? The wrapper propagates the error without adding extra headers.
- What happens when extra headers contain an empty-string value? Empty-string values are silently skipped (not sent as headers), preventing accidental empty headers from reaching the proxy.
- What happens when the WebSocket connection drops mid-RPC? The tunnel proxy surfaces the error to the gRPC caller as a transport-level error. No auto-reconnect; the caller creates a new TunnelProxy.
- What happens when Close is called on a tunnel proxy that was never used? Close returns immediately without error.
- What happens when the tunnel proxy receives concurrent RPCs? Each RPC gets its own goroutine-backed WebSocket stream (goroutine-per-connection model, matching the Rust implementation).

## Clarifications

### Session 2026-07-02

- Q: What should happen when tunnel Close draining exceeds the expected time? → A: Force-close remaining connections after a configurable timeout (default 5 seconds). Prevents indefinite hangs while allowing callers to increase the timeout for long-running RPCs.
- Q: Should the tunnel proxy auto-reconnect after a WebSocket connection drop, or should the caller create a new tunnel? → A: No auto-reconnect. Connection drops surface as transport errors, and the caller is responsible for creating a new TunnelProxy. This matches the Rust implementation and keeps the tunnel proxy stateless between RPCs.
- Q: Should header key comparison for collision detection (FR-002) be case-sensitive or case-insensitive? → A: Case-insensitive, per HTTP/2 specification (RFC 9113 Section 8.2: field names are lowercase). Header keys are normalized to lowercase before merge.

## Requirements

### Functional Requirements

- **FR-001**: The SDK MUST provide a generic header wrapping mechanism that decorates any auth provider with additional static headers per RPC.
- **FR-002**: When extra headers and base provider headers share a key (compared case-insensitively per HTTP/2 spec), the extra header MUST take precedence. Header keys are normalized to lowercase before merge.
- **FR-003**: The header wrapper MUST delegate transport security requirements to the base auth provider.
- **FR-004**: The header wrapper MUST NOT introduce any new external dependencies in the core SDK.
- **FR-005**: A Cloudflare Access convenience constructor MUST format the `cf-access-jwt-assertion` header and `CF_Authorization` cookie from a single edge token parameter.
- **FR-006**: The Cloudflare Access constructor MUST validate that the edge token is non-empty and return an error if it is not.
- **FR-007**: A WebSocket tunnel proxy MUST bridge gRPC connections over WebSocket for environments where HTTP/2 POST is rejected by the edge proxy. The tunnel uses raw binary WebSocket frames to carry HTTP/2 bytes end-to-end (byte-stream tunneling), matching the Rust CLI's tunnel implementation. No gRPC-specific sub-protocol negotiation is required; the WebSocket connection acts as a transparent byte pipe between the gRPC client and the upstream gateway.
- **FR-008**: The tunnel proxy MUST accept its own edge token parameter, independent of the application-level auth provider.
- **FR-009**: The tunnel proxy MUST provide explicit lifecycle management with a Close method that drains in-flight connections before shutting down. If draining exceeds a configurable timeout (default 5 seconds), remaining connections are force-closed.
- **FR-010**: The tunnel proxy MUST clean up all goroutines on Close, preventing leaks.
- **FR-011**: The tunnel proxy MUST support TLS configuration for the WebSocket connection (wss:// vs ws://).
- **FR-012**: The tunnel proxy MUST use a goroutine-per-connection model for concurrent RPC handling.
- **FR-013**: The tunnel proxy MUST support configurable logging using the `types.Logger` interface, consistent with `RefreshableToken` and other SDK components.
- **FR-014**: All edge-specific functionality (Cloudflare Access convenience, WebSocket tunnel) MUST reside in a separate optional package, keeping the WebSocket dependency out of the core module.

### Key Entities

- **AuthProvider**: The existing interface for providing per-RPC authentication credentials. The extra headers wrapper decorates this interface.
- **TunnelProxy**: A new entity representing a WebSocket tunnel connection to the gateway. Manages its own lifecycle (creation, active connections, graceful shutdown).
- **TunnelOption**: Configuration options for the tunnel proxy (TLS settings, logger, connection parameters).

## Success Criteria

### Measurable Outcomes

- **SC-001**: Developers can add edge proxy headers to any existing auth provider with a single function call, requiring no changes to existing auth setup code.
- **SC-002**: Cloudflare Access integration requires exactly two parameters (base auth, edge token) with no manual header name knowledge.
- **SC-003**: WebSocket tunnel proxy supports at least 10 concurrent gRPC streams without connection errors or goroutine leaks.
- **SC-004**: Tunnel proxy Close completes within 5 seconds under normal conditions, draining all in-flight connections.
- **SC-005**: All core SDK changes (header wrapper) introduce zero new external dependencies.
- **SC-006**: Edge package functionality is fully optional, with no impact on existing SDK consumers who do not import it.

## Assumptions

- Edge tokens are static for the lifetime of a connection and do not require independent refresh cycles. If dynamic edge token refresh is needed later, a callback-based mechanism can be added without breaking changes.
- Cloudflare Access is the primary edge proxy vendor; other vendors (Google IAP, Zscaler) will be added as convenience constructors when concrete use cases arise.
- The WebSocket tunnel follows the goroutine-per-connection model consistent with the Rust CLI implementation, which has proven sufficient for production workloads.
- The existing `types.Logger` interface is adequate for tunnel proxy logging, maintaining consistency with `RefreshableToken` and other SDK components.
- The tunnel proxy is a transport concern and does not need a fake implementation for testing. SDK consumers test edge auth by wrapping fake auth providers with the generic header decorator.
- Empty-string header values are treated as "not set" and silently skipped, preventing accidental empty headers from causing proxy validation failures.

## Dependencies

- **Core header wrapper** (FR-001 through FR-004): No new dependencies. Uses only the Go standard library and the existing `types.AuthProvider` interface.
- **Cloudflare Access convenience** (FR-005, FR-006): No new dependencies. Built on top of the core header wrapper.
- **WebSocket tunnel** (FR-007 through FR-013): Requires a WebSocket library. The recommended choice is `github.com/coder/websocket` (MIT license), which provides a minimal, idiomatic Go API with context support and is actively maintained. This dependency is isolated in the separate edge package (FR-014) and does not affect the core module. This is justified per Constitution V (Minimal Dependencies) because: (a) the stdlib has no WebSocket support, (b) the dependency is confined to an optional package, and (c) existing SDK consumers who do not import the edge package incur zero additional dependencies.

## Out of Scope

- **Dynamic edge token refresh**: Edge tokens are assumed static for the connection lifetime. A callback-based refresh mechanism may be added in a future spec without breaking changes.
- **Vendor-specific convenience constructors beyond Cloudflare Access**: Google IAP, Zscaler, and other zero-trust proxies are deferred until concrete use cases arise. Developers can use the generic header wrapper (FR-001) for any vendor.
- **Auto-reconnect for the WebSocket tunnel**: Connection drops surface as transport errors. The caller is responsible for creating a new TunnelProxy. This matches the Rust implementation.
- **Load balancing or connection pooling in the tunnel proxy**: The tunnel is a single-connection transport bridge, not a connection pool.
- **Upstream proto changes**: This feature does not require any proto schema modifications.

## Error Handling

| Scenario | Behavior |
|---|---|
| Base auth provider returns an error during `GetRequestMetadata` | The header wrapper propagates the error without adding extra headers. |
| Extra headers contain an empty-string value | Empty-string values are silently skipped (not sent as headers). |
| Cloudflare Access constructor receives an empty edge token | Returns an error immediately with a descriptive message. |
| Tunnel proxy receives an invalid gateway URL | `NewTunnelProxy` returns an error at creation time with a descriptive message. |
| WebSocket connection drops mid-RPC | The tunnel surfaces the error to the gRPC caller as a transport-level error. |
| `Close` called on a tunnel proxy that was never used | Returns immediately without error. |
| `Close` draining exceeds the configured timeout | Remaining connections are force-closed after the timeout (default 5 seconds). |
| `Close` called concurrently from multiple goroutines | Must be safe for concurrent calls; second and subsequent calls return immediately. |
