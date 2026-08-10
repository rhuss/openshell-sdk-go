# Feature Specification: Reverse Port Forwarding (ssh -R)

**Feature Branch**: `025-reverse-port-forwarding`
**Created**: 2026-08-05
**Status**: Draft
**Input**: Brainstorm #016 - Reverse Port Forwarding

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Expose Local Service to Sandbox (Priority: P1)

A developer runs a service on their local machine (e.g., an MCP tool server, model inference endpoint, or API server) and needs sandbox code to reach it. They call `TCP().RemoteListen()` with the workspace, sandbox name, the port the sandbox should listen on, and the local address to bridge to. The method blocks, and any connection made to that port inside the sandbox is transparently tunneled back to the developer's local service.

**Why this priority**: This is the core use case that all other scenarios build on. Without the ability to bridge a single local service into a sandbox, none of the higher-level workflows (model serving, debugging, dev testing) are possible.

**Independent Test**: Can be tested by calling `RemoteListen()` with valid parameters, making a connection from inside the sandbox to the specified port, and verifying that data flows end-to-end between the sandbox process and the local service.

**Acceptance Scenarios**:

1. **Given** a running sandbox and a local service on port 8080, **When** the developer calls `TCP().RemoteListen(ctx, "default", "my-sandbox", 8080, "localhost:8080")`, **Then** a process inside the sandbox can connect to `localhost:8080` and reach the developer's local service.
2. **Given** an active reverse tunnel, **When** multiple connections are made from inside the sandbox, **Then** each connection is independently bridged to the local target, and connections do not interfere with each other.
3. **Given** an active reverse tunnel, **When** the developer cancels the context, **Then** all active bridges are torn down and RemoteListen returns `ctx.Err()`.

---

### User Story 2 - Custom Bind Address and Service Identification (Priority: P2)

A developer needs to bind the sandbox-side listener to a specific address (e.g., `0.0.0.0` to accept connections from any interface within the sandbox) and attach a service identifier for audit and correlation purposes. They use `WithRemoteBindAddress()` and `WithRemoteListenServiceID()` options.

**Why this priority**: Options extend the core behavior for real-world deployment scenarios where the default `127.0.0.1` bind address is insufficient or where operational observability requires service tagging.

**Independent Test**: Can be tested by calling `RemoteListen()` with options and verifying that the bind address is passed through to the proto layer and that the service ID appears in connection metadata.

**Acceptance Scenarios**:

1. **Given** a sandbox in workspace "default", **When** `RemoteListen` is called with `WithRemoteBindAddress("0.0.0.0")`, **Then** the sandbox-side listener accepts connections from any interface, not just loopback.
2. **Given** a sandbox in workspace "default", **When** `RemoteListen` is called with `WithRemoteListenServiceID("mcp-proxy")`, **Then** the service ID is included in the proto request for audit and correlation.

---

### User Story 3 - Graceful Error Handling (Priority: P2)

A developer's local service may be temporarily unavailable (restarted, crashed) while the reverse tunnel is active. Per-connection failures (failed dial to local target, broken bridge) must not tear down the entire tunnel. Only permanent errors (sandbox deleted, authentication revoked) should cause `RemoteListen` to return.

**Why this priority**: Resilient error handling is essential for inner-loop development workflows where local services restart frequently. Tearing down the tunnel on every transient failure would break the developer experience.

**Independent Test**: Can be tested by stopping the local service while a reverse tunnel is active, verifying the tunnel remains up, then restarting the service and verifying new connections succeed.

**Acceptance Scenarios**:

1. **Given** an active reverse tunnel, **When** the local target is temporarily unreachable, **Then** the failed connection is dropped but RemoteListen continues accepting new connections.
2. **Given** an active reverse tunnel, **When** the sandbox is deleted, **Then** RemoteListen returns a permanent error.
3. **Given** an active reverse tunnel, **When** authentication credentials are revoked, **Then** RemoteListen returns a permanent error.

---

### User Story 4 - Fake Client Support (Priority: P3)

A developer writing tests against the SDK's fake client calls `TCP().RemoteListen()` and receives an `Unimplemented` error, consistent with how other streaming methods (`Listen()`, `Tunnel()`) behave in the fake.

**Why this priority**: Fake-real parity is an SDK invariant. Adding the method to the fake ensures test code compiles and behaves predictably.

**Independent Test**: Can be tested by calling `RemoteListen()` on a fake client and asserting the returned error is `Unimplemented`.

**Acceptance Scenarios**:

1. **Given** a fake SDK client, **When** `TCP().RemoteListen()` is called, **Then** it returns an `Unimplemented` error.
2. **Given** a fake SDK client, **When** `TCP().RemoteListen()` is called with any combination of options, **Then** input validation runs first (e.g., empty sandbox name returns `InvalidArgument`) and only valid calls return `Unimplemented`.

---

### Edge Cases

- What happens when `remotePort` is 0 or > 65535? Returns `InvalidArgument`.
- What happens when `sandboxName` is empty? Returns `InvalidArgument`.
- What happens when `localTarget` is malformed (missing port, invalid host)? Returns `InvalidArgument`.
- What happens when the client is already closed? Returns `Unavailable`.
- What happens when the same remote port is requested twice on the same sandbox? Behavior depends on the proto layer (likely returns an error from the gateway).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: SDK MUST add a `RemoteListen` method to the `TCPInterface` interface with signature `RemoteListen(ctx context.Context, workspace, sandboxName string, remotePort uint32, localTarget string, opts ...RemoteListenOption) error` that sets up reverse port forwarding from a sandbox back to the client. The method blocks until context cancellation or a permanent error, returning `error`.
- **FR-002**: `RemoteListen` MUST block until context cancellation or an unrecoverable error occurs.
- **FR-003**: `RemoteListen` MUST bridge each accepted connection from the sandbox to the specified `localTarget` on the client side.
- **FR-004**: SDK MUST validate inputs: empty `sandboxName` returns `InvalidArgument`, `remotePort` of 0 or > 65535 returns `InvalidArgument`, `localTarget` that fails `net.SplitHostPort` parsing returns `InvalidArgument`.
- **FR-005**: SDK MUST provide a `WithRemoteBindAddress(addr string)` option that overrides the sandbox-side bind address (default: `127.0.0.1`).
- **FR-006**: SDK MUST provide a `WithRemoteListenServiceID(id string)` option for audit and correlation.
- **FR-007**: SDK MUST treat per-connection errors (failed dial to localTarget, broken bridge) as transient and continue accepting new connections.
- **FR-008**: SDK MUST treat permanent errors (sandbox deleted, auth revoked) as fatal and return from `RemoteListen`.
- **FR-009**: SDK MUST tear down all active bridges when context is cancelled and return `ctx.Err()`.
- **FR-010**: Fake client MUST return `Unimplemented` for `RemoteListen` calls that pass input validation.
- **FR-011**: Fake client MUST perform the same input validation as the real client before returning `Unimplemented`.
- **FR-012**: SDK MUST return `Unavailable` if `RemoteListen` is called on a closed client.

### Key Entities

- **RemoteListenOption**: Functional option type for configuring reverse listen behavior (bind address, service ID).
- **remoteListenConfig**: Internal config struct holding resolved option values.
- **TCPInterface**: Existing sub-client interface extended with the `RemoteListen` method.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `RemoteListen` is callable on the `TCPInterface` and compiles without errors.
- **SC-002**: All input validation cases (empty name, invalid port, malformed target, closed client) return the correct error type.
- **SC-003**: Fake client returns `Unimplemented` for valid `RemoteListen` calls.
- **SC-004**: Both options (`WithRemoteBindAddress`, `WithRemoteListenServiceID`) are accepted and their values are accessible in the internal config.
- **SC-005**: The method signature and option types follow existing SDK patterns (consistent with `Listen`, `Forward`, `Tunnel`).
- **SC-006**: All existing tests continue to pass after the interface change.
- **SC-007**: New unit tests cover all edge cases enumerated above.

## Clarifications

### Session 2026-08-05

- Q: What does "malformed localTarget" mean for input validation? → A: Format validation only via `net.SplitHostPort`. A localTarget that cannot be parsed into host and port components is invalid. Host resolution and reachability are runtime concerns, not input validation.
- Q: Are the brainstorm's deferred options (WithOnError, WithMaxConnections, WithSSHTunnel) in scope? → A: No. All three are explicitly out-of-scope for v1. They can be added as non-breaking extensions later.
- Q: Should RemoteListen support workspace-scoped sandbox names? → A: Yes. Follow existing SDK patterns established in workspace scoping (PR #41). RemoteListen takes `workspace` as an explicit parameter, consistent with `Forward` and `Listen`.

### Non-Functional Requirements

- **NFR-001**: `RemoteListen` MUST NOT leak goroutines. Every goroutine spawned for connection bridging must exit when the connection closes or context is cancelled.
- **NFR-002**: Documentation (doc.go examples, README feature list) MUST be updated in the same PR per Constitution XIII.

## Out of Scope (v1)

- `WithOnError(func(error))` callback for per-connection error reporting (logging is sufficient for v1)
- `WithMaxConnections(n int)` to limit concurrent reverse-tunneled connections
- `WithSSHTunnel()` option for SSH-based reverse tunneling (v1 is direct TCP only)
- Real gRPC implementation (blocked on upstream proto extension)

## Assumptions

- The upstream proto extension for reverse forwarding does not yet exist. This implementation covers the SDK-side API surface, types, options, input validation, and fake client. The real gRPC implementation will be added when the proto support lands.
- The real client's `RemoteListen` method will initially return `Unimplemented` (same as the fake), since there is no proto RPC to call yet. This is a stub that establishes the interface contract.
- The `localTarget` parameter uses standard Go `host:port` format as accepted by `net.Dial`.
- No new dependencies are required for this feature.
- The proto extension sketch in the brainstorm document is informational and does not need to be implemented as part of this spec.
