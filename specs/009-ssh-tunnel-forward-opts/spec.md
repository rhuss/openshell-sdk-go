# Feature Specification: SSH Tunneling and TCP Forward Options

**Feature Branch**: `009-ssh-tunnel-forward-opts`
**Created**: 2026-06-29
**Status**: Draft
**Input**: Brainstorm 010-ssh-tunnel-forward-opts

## User Scenarios & Testing

### User Story 1 - SSH Tunnel to a Sandbox in One Call (Priority: P1)

A developer wants to open an SSH tunnel to a sandbox without manually
orchestrating session creation, stream setup, and cleanup. They call a
single method on the SSH sub-client, passing the sandbox name and target
port, and receive a bidirectional byte stream. When they close the stream,
the underlying SSH session is automatically revoked.

**Why this priority**: This is the core value proposition. Today, SSH
tunneling requires 3+ manual steps with raw gRPC orchestration. A single
call eliminates the most common pain point.

**Independent Test**: Can be fully tested by calling the tunnel method
with a valid sandbox name and port, exchanging data over the returned
stream, and verifying the SSH session is revoked after closing.

**Acceptance Scenarios**:

1. **Given** a running sandbox, **When** the developer calls the tunnel
   method with the sandbox name and a valid port, **Then** they receive
   a bidirectional byte stream connected through the SSH relay.
2. **Given** an active tunnel, **When** the developer closes the stream,
   **Then** the underlying gRPC stream is closed first, then the SSH
   session is revoked (graceful shutdown order).
3. **Given** an active tunnel, **When** the parent context is cancelled,
   **Then** the stream is terminated and the SSH session is revoked.
4. **Given** a sandbox name that does not exist, **When** the developer
   calls the tunnel method, **Then** a descriptive error is returned
   without leaking a dangling SSH session.
5. **Given** a valid sandbox, **When** the developer passes an invalid
   port (0 or >65535), **Then** the call is rejected client-side before
   any network request.

---

### User Story 2 - Correlate TCP Forwarding with a Service ID (Priority: P2)

A developer or platform operator wants to tag TCP forwarding streams
with a service identifier for audit trails and log correlation. They
pass an option when calling the forward method, and the service ID is
included in the initial protocol frame sent to the gateway.

**Why this priority**: Audit and correlation are important for
production deployments but do not block the primary SSH tunneling
use case. This is a gap fix on existing functionality.

**Independent Test**: Can be tested by calling the forward method with a
service ID option and verifying the initial frame includes the
service ID field.

**Acceptance Scenarios**:

1. **Given** a valid sandbox and port, **When** the developer calls
   forward with a service ID option, **Then** the service ID is included
   in the initial protocol frame.
2. **Given** a valid sandbox and port, **When** the developer calls
   forward without a service ID option, **Then** the behavior is
   identical to today (no service ID in the frame).
3. **Given** a tunnel call with a service ID option, **When** the tunnel
   is established, **Then** the service ID is passed through to the
   underlying forwarding frame.

---

### User Story 3 - Use Fake Client with Tunnel and Forward Options (Priority: P3)

A developer writing tests against the fake client expects the tunnel
method and forward options to follow the same contract as the real
client. The fake tunnel returns an Unimplemented error (consistent
with other fake SSH methods). The fake forward continues to work as
before, ignoring options gracefully.

**Why this priority**: Fake client parity ensures developers can write
tests without a live gateway. Lower priority because it is a support
concern, not a user-facing feature.

**Independent Test**: Can be tested by calling the fake client's tunnel
method and verifying it returns an Unimplemented error. Can also test
that the fake forward method accepts service ID options without error
(beyond the existing Unimplemented return).

**Acceptance Scenarios**:

1. **Given** a fake client, **When** the developer calls the tunnel
   method, **Then** an Unimplemented error is returned.
2. **Given** a fake client, **When** the developer calls forward with a
   service ID option, **Then** the option is accepted and an
   Unimplemented error is returned (consistent with current behavior).
3. **Given** a fake client that has been closed, **When** the developer
   calls the tunnel method, **Then** an Unavailable error is returned.

---

### Edge Cases

- What happens when the SSH session is created successfully but the
  subsequent TCP forward stream fails? The session must be revoked
  before the error is returned, preventing session leaks.
- What happens when the tunnel's Close method is called multiple times?
  It must be safe to call Close more than once without panicking or
  double-revoking the session.
- What happens when data is written to the tunnel after Close? The write
  must return an error, not panic.
- What happens when the gateway revokes the SSH session externally while
  the tunnel is active? The stream should surface the server-side error
  to the caller on the next Read or Write.

## Requirements

### Functional Requirements

- **FR-001**: The SSH sub-client MUST provide a tunnel method that
  combines session creation and TCP forwarding into a single call,
  returning a bidirectional byte stream.
- **FR-002**: The tunnel method MUST accept a sandbox name (not sandbox
  ID), consistent with the SDK's developer-facing convention.
- **FR-003**: The tunnel method MUST accept a target port in the range
  1-65535. Out-of-range values MUST be rejected client-side before any
  network request.
- **FR-004**: The tunnel method MUST support an optional service ID for
  audit and correlation purposes, passed through to the forwarding frame.
- **FR-005**: When the tunnel's stream is closed, the system MUST close
  the gRPC stream first, then revoke the SSH session (graceful shutdown
  order: protocol close before context cancel).
- **FR-006**: If the forwarding stream fails to open after the SSH
  session is created, the system MUST revoke the session before
  returning the error to prevent session leaks.
- **FR-007**: Closing the tunnel MUST be safe to call multiple times
  without panicking or double-revoking.
- **FR-008**: The TCP forward method MUST support an optional service ID
  that is included in the initial protocol frame sent to the gateway.
- **FR-009**: When no service ID option is provided, the forward method
  MUST behave identically to the current implementation (backward
  compatible).
- **FR-010**: The fake SSH client MUST implement the tunnel method,
  returning an Unimplemented error consistent with other fake SSH
  methods. The fake MUST perform the same client-side validation as the
  real client (port range check per FR-003) before returning
  Unimplemented, per Constitution XI (Fake-Real Parity).
- **FR-011**: The fake TCP client MUST accept service ID options on the
  forward method without error (beyond the existing Unimplemented
  return).

### Key Entities

- **SSH Tunnel**: A combined resource that wraps an SSH session and a TCP
  forward stream. Owns the lifecycle of both: creating them together and
  tearing them down in the correct order on close.
- **Service ID**: An optional string identifier attached to forwarding
  frames for audit trails and log correlation across gateway components.
- **Tunnel Option**: A functional option applied to the tunnel method
  (initially only service ID, extensible for future options).
- **Forward Option**: A functional option applied to the TCP forward
  method (initially only service ID, extensible for future options).

## Success Criteria

### Measurable Outcomes

- **SC-001**: A developer can establish an SSH tunnel to a sandbox with
  a single method call, reducing the required steps from 3+ manual
  orchestration calls to 1.
- **SC-002**: When the tunnel is closed (explicitly or via context
  cancellation), the SSH session is always revoked, with zero session
  leaks in normal and error paths.
- **SC-003**: The TCP forward method accepts a service ID option that
  appears in the initial protocol frame, enabling end-to-end audit
  correlation.
- **SC-004**: All new methods and options are covered by unit tests,
  including success paths, error paths, and edge cases (double-close,
  invalid port, session leak prevention).
- **SC-005**: The fake client implements all new interface methods,
  maintaining compile-time interface satisfaction and test usability.
- **SC-006**: Existing code using the TCP forward method without options
  continues to work without modification (full backward compatibility).

## Clarifications

### Session 2026-06-29

- No critical ambiguities detected. All categories assessed as Clear or
  deferred to planning (observability details are plan-level concerns,
  not spec-level decisions).

## Assumptions

- The SDK's convention of using sandbox names (not IDs) in
  developer-facing methods applies to the tunnel method. The tunnel
  method will need to resolve sandbox name to ID internally, consistent
  with how other SDK methods (like GetLogs) handle this.
- The service ID option types are separate for SSH tunnel options and TCP
  forward options (not a shared type), keeping each sub-client's option
  namespace independent and avoiding coupling.
- The `SshRelayTarget` proto message is an empty message (no fields to
  configure), so no additional target configuration is needed for the
  SSH relay path.
- The `authorization_token` field in `TcpForwardInit` is populated
  automatically by the tunnel method using the token from
  `CreateSession`. It is not exposed as a user-facing option.
- The tunnel method uses the same `tcpForwardConn` wrapper (or a
  composition of it) for the returned `io.ReadWriteCloser`, extending
  the close behavior to include session revocation.

## Out of Scope

- SSH target or auth token options on the TCP forward method (the tunnel
  method covers the SSH use case; exposing these on Forward would be
  YAGNI).
- Session sharing across multiple tunnels (additive later if needed).
- `net.Conn` upgrade for the returned stream (deferred from prior phases,
  still deferred).
- Local port listener / `net.Listener` pattern (like `ssh -L`), deferred
  to idea inbox.
- Reverse port forwarding, deferred to idea inbox.

## Dependencies

- Phase 2b-1 (spec 007/008): SSH session management and TCP port
  forwarding must be implemented and merged (already delivered).
- Proto definitions for `TcpForwardInit`, `SshRelayTarget`,
  `TcpRelayTarget`, `service_id`, and `authorization_token` must be
  present in the generated Go bindings (already available).
