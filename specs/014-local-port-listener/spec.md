# Feature Specification: Local Port Listener

**Feature Branch**: `014-local-port-listener`
**Created**: 2026-06-30
**Status**: Draft
**Input**: Brainstorm 015 - Local Port Listener

## User Scenarios & Testing

### User Story 1 - Bind a Local Port to a Sandbox Port (Priority: P1)

An SDK consumer wants to expose a service running inside a sandbox on a local port, similar to `ssh -L` port forwarding. They call a single method that binds a local address, and every incoming connection is automatically tunneled to the specified sandbox port. The consumer receives a `net.Listener` that they can use with existing server frameworks and tools.

**Why this priority**: This is the core value proposition. Without it, consumers must write their own accept loop, bridging goroutines, and teardown logic, which is repetitive and error-prone.

**Independent Test**: Create a listener bound to a local port, connect to it, send data through the tunnel, and verify the data arrives at the sandbox service.

**Acceptance Scenarios**:

1. **Given** a running sandbox with a service on port 80, **When** the consumer creates a listener for local port 8080 tunneled to sandbox port 80, **Then** the listener binds successfully and returns a `net.Listener` with the bound address.
2. **Given** a listener is active, **When** a client connects to localhost:8080, **Then** the connection is accepted and data flows bidirectionally between the client and the sandbox service.
3. **Given** a listener is active with multiple concurrent client connections, **When** each client sends data independently, **Then** each connection is tunneled independently without interference.

---

### User Story 2 - OS-Assigned Local Port (Priority: P1)

An SDK consumer needs a listener on an ephemeral port (useful for tests, tools, and avoiding port conflicts). They pass local port 0, the OS assigns an available port, and the consumer discovers the actual port from the listener's address.

**Why this priority**: Ephemeral ports are essential for testing and multi-instance scenarios. This is the same priority as the base listener because it's a zero-cost feature of the listener contract.

**Independent Test**: Create a listener with local port 0, read the assigned address, connect to it, and verify the tunnel works.

**Acceptance Scenarios**:

1. **Given** a sandbox with a service, **When** the consumer creates a listener with local port 0, **Then** the listener binds to an OS-assigned port and the actual address is discoverable from the listener.
2. **Given** an OS-assigned listener, **When** a client connects to the discovered address, **Then** the tunnel functions identically to a fixed-port listener.

---

### User Story 3 - Graceful Shutdown (Priority: P1)

An SDK consumer closes the listener and expects all active connections to be torn down cleanly. No goroutine leaks, no orphaned streams, no blocked callers.

**Why this priority**: Resource cleanup is fundamental to correct SDK usage. Leaking connections or goroutines is a critical defect.

**Independent Test**: Create a listener, establish several connections, close the listener, and verify all connections are terminated and no goroutines leak.

**Acceptance Scenarios**:

1. **Given** a listener with active connections, **When** the consumer closes the listener, **Then** the listener stops accepting new connections and all active connections are closed.
2. **Given** a closed listener, **When** a client attempts to connect, **Then** the connection is refused.
3. **Given** a closed listener, **When** the consumer inspects resource usage, **Then** no goroutines or streams from the listener remain active.

---

### User Story 4 - Custom Bind Address (Priority: P2)

An SDK consumer wants to bind the listener to an address other than localhost (e.g., `0.0.0.0` to accept connections from the network). They provide a bind address option when creating the listener.

**Why this priority**: Important for production deployments and multi-host setups, but most consumers use the default localhost binding.

**Independent Test**: Create a listener with a custom bind address, connect from the appropriate network interface, and verify the tunnel works.

**Acceptance Scenarios**:

1. **Given** a sandbox with a service, **When** the consumer creates a listener with bind address `0.0.0.0`, **Then** the listener accepts connections on all network interfaces.
2. **Given** no bind address is specified, **When** the consumer creates a listener, **Then** the listener binds to `127.0.0.1` by default.

---

### User Story 5 - SSH Tunnel Transport (Priority: P2)

An SDK consumer wants per-connection SSH session authentication instead of plain TCP forwarding. They provide a transport option when creating the listener, and each accepted connection uses an SSH tunnel internally.

**Why this priority**: Required for environments that mandate SSH-based authentication, but most consumers use the default TCP transport.

**Independent Test**: Create a listener with the SSH tunnel option, connect, and verify the tunnel uses SSH-based transport.

**Acceptance Scenarios**:

1. **Given** a sandbox with SSH access, **When** the consumer creates a listener with the SSH tunnel option, **Then** each accepted connection authenticates via SSH.
2. **Given** no transport option is specified, **When** the consumer creates a listener, **Then** the default TCP forward transport is used.

---

### User Story 6 - Fake Implementation for Testing (Priority: P2)

An SDK consumer writes tests against the TCP interface and needs the listener method to be present in the fake implementation. The fake returns an "unimplemented" error, consistent with how other advanced methods are faked.

**Why this priority**: Maintains testability of consumer code without requiring a live sandbox.

**Independent Test**: Call the listener method on the fake implementation and verify it returns an unimplemented error.

**Acceptance Scenarios**:

1. **Given** a fake TCP client, **When** the consumer calls the listener method, **Then** an unimplemented error is returned.

---

### Edge Cases

- What happens when the requested local port is already in use? The listener creation fails with a clear error indicating the port conflict.
- What happens when the sandbox becomes unreachable after the listener is created? Individual connection attempts fail, but the listener continues accepting new connections. Failed connections return errors to the respective callers.
- What happens when the context passed to the listener is cancelled? The listener closes and all active connections are torn down, same as explicit close.
- What happens when connection bridging fails mid-stream (e.g., the sandbox-side service drops)? The affected connection is closed with an error. Other connections are unaffected.

## Clarifications

### Session 2026-06-30

- Q: When a tunnel setup fails for an incoming connection, should Accept() return the error or handle it internally? → A: The listener handles tunnel setup failures internally and only returns successfully established connections from Accept(). Mid-stream failures on already-accepted connections are reported through the connection read/write interface.
- Q: Should Close() block until all active connections are terminated, or return immediately? → A: Close() blocks until all active connections are closed, providing deterministic cleanup.
- Q: Can multiple goroutines call Accept() concurrently on the same listener? → A: Yes, concurrent Accept() calls are safe, consistent with the standard listener contract.

## Requirements

### Functional Requirements

- **FR-001**: The SDK MUST provide a `Listen` method on `TCPInterface` that binds a local address and tunnels accepted connections to a specified sandbox port.
- **FR-002**: The listener MUST return a `net.Listener` compatible with existing server frameworks (e.g., `http.Serve`).
- **FR-003**: Each accepted connection MUST create an independent tunnel stream to the sandbox.
- **FR-004**: The listener MUST support caller-specified local ports, including port 0 for OS-assigned ephemeral ports.
- **FR-005**: When port 0 is used, the actual bound address MUST be discoverable from the returned listener.
- **FR-006**: The listener MUST default to binding on `127.0.0.1`.
- **FR-007**: The listener MUST support a bind address option to override the default.
- **FR-008**: The listener MUST support a `WithSSHTunnel()` option (of type `ListenOption`) to use SSH tunnel instead of TCP forward for per-connection authentication.
- **FR-009**: Closing the listener MUST stop accepting new connections and tear down all active connections.
- **FR-010**: Context cancellation MUST close the listener and all active connections.
- **FR-011**: The fake TCP implementation MUST include the listener method. It MUST validate inputs (sandbox name non-empty, remotePort in range 1-65535, localPort in range 0-65535) before returning an unimplemented error, consistent with fake-real parity (Constitution XI).
- **FR-012**: The listener MUST support a service ID option, consistent with existing forward/tunnel options.
- **FR-013**: Individual connection failures (tunnel setup, mid-stream drops) MUST NOT stop the listener from accepting new connections.
- **FR-014**: Tunnel setup failures for incoming connections MUST be handled internally by the listener; Accept() MUST only return successfully established connections.
- **FR-015**: Mid-stream failures on established connections MUST be reported through the connection's read/write interface.
- **FR-016**: Close() MUST block until all active connections are terminated.
- **FR-017**: The listener MUST be safe for concurrent `Accept()` calls from multiple goroutines, consistent with the `net.Listener` contract.

### Key Entities

- **Listener**: Binds a local address and accepts incoming connections, each tunneled to a sandbox port. Implements `net.Listener` (Accept, Close, Addr).
- **Listener Options**: Configuration for bind address, transport mode (TCP forward vs. SSH tunnel), and service ID.
- **Tunneled Connection**: A single bidirectional stream between a local client and the sandbox service, managed by the listener.

## Success Criteria

### Measurable Outcomes

- **SC-001**: SDK consumers can establish a local-to-sandbox port tunnel with a single method call, eliminating the need for manual accept loops and bridging logic.
- **SC-002**: The listener handles 10+ concurrent connections without interference between streams.
- **SC-003**: Closing the listener tears down all active connections within 5 seconds with zero goroutine leaks (5-second bound is a test timeout; Close() itself blocks per FR-016).
- **SC-004**: SDK consumers can use the listener with port 0 and discover the assigned port programmatically.
- **SC-005**: The listener integrates with existing server frameworks through the `net.Listener` interface without adaptation code.

## Assumptions

- The existing TCP forward and SSH tunnel mechanisms are stable and tested; the listener builds on top of them.
- Connection tracking and max-connection limits are deferred to a future iteration; the internal design should accommodate adding them later as non-breaking options.
- An error callback for connection-level failures is deferred to a future iteration; transient failures are surfaced through the accept mechanism.
- The listener does not implement reconnection or retry logic for the underlying tunnel streams; each connection is independent and failures are reported individually.
