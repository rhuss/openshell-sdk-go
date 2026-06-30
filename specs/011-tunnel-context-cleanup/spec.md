# Feature Specification: SSH Tunnel Context Cancellation Cleanup

**Feature Branch**: `011-tunnel-context-cleanup`
**Created**: 2026-06-30
**Status**: Draft
**Input**: Brainstorm 012: context-cancel-cleanup

## User Scenarios & Testing

### User Story 1 - Auto-cleanup on context cancellation (Priority: P1)

As an SDK consumer, when my parent context is cancelled (timeout, shutdown,
crash path) while an SSH tunnel is active, the SDK automatically revokes the
SSH session on the server so I do not leak server-side resources.

**Why this priority**: This is a correctness fix. Leaked SSH sessions consume
server resources and are invisible to callers. Two independent reviewers
flagged this gap in PR #7.

**Independent Test**: Create a tunnel, cancel the parent context without
calling Close(), verify the SSH session was revoked on the server.

**Acceptance Scenarios**:

1. **Given** an active SSH tunnel with a valid session, **When** the parent
   context is cancelled, **Then** the SSH session is revoked automatically
   within the cleanup goroutine.
2. **Given** an active SSH tunnel, **When** the parent context is cancelled
   and the caller also calls Close(), **Then** exactly one revocation occurs
   (idempotent via closeOnce).
3. **Given** an active SSH tunnel, **When** the caller calls Close() before
   any context cancellation, **Then** the session is revoked normally and the
   cleanup goroutine is a no-op.

---

### User Story 2 - Best-effort revocation (Priority: P2)

As an SDK consumer, when auto-cleanup fires but the revocation RPC fails
(network error, server unavailable), the failure is silently discarded and
does not surface as an error to my application.

**Why this priority**: Callers should not be surprised by errors from cleanup
paths they did not initiate. The existing manual revokeFunc already uses
fire-and-forget semantics.

**Independent Test**: Create a tunnel, cancel context, simulate revocation
failure, verify no error propagates and no panic occurs.

**Acceptance Scenarios**:

1. **Given** an active SSH tunnel whose session revocation will fail, **When**
   the parent context is cancelled, **Then** the revocation error is silently
   discarded.

---

### Edge Cases

- Context cancelled before the tunnel is fully established: the existing
  defer in Tunnel() already handles this by revoking the session if setup
  fails. No change needed.
- Close() and context cancellation race: closeOnce guarantees exactly-once
  execution. No double-revoke possible.
- done channel already closed before cleanup goroutine starts: the goroutine
  reads from a closed channel immediately and proceeds to call Close().
  This is safe Go behavior.

## Requirements

### Functional Requirements

- **FR-001**: The SDK MUST automatically revoke the SSH session when the
  parent context is cancelled on an active tunnel.
- **FR-002**: The auto-revocation MUST be idempotent with explicit Close()
  calls (no double-revoke, no panic).
- **FR-003**: The auto-revocation MUST use best-effort semantics (errors
  from the revocation RPC are silently discarded).
- **FR-004**: The auto-revocation MUST NOT change the existing Close()
  public API or return type.
- **FR-005**: The fix MUST apply only to SSH Tunnel (sshTunnel), not TCP
  Forward (tcpForwardConn). TCP Forward has no server-side session to
  revoke, so auto-cleanup is not applicable.

### Key Entities

- **sshTunnel**: Wraps tcpForwardConn with SSH session lifecycle. Owns
  closeOnce and revokeFunc. Needs a cleanup goroutine that watches the
  done channel and calls Close() when it fires.
- **tcpForwardConn**: Lower-level stream wrapper. Owns the done channel
  and readLoop goroutine. When the parent context is cancelled, the gRPC
  stream terminates, readLoop exits, and done is closed. This is the
  signal the cleanup goroutine uses to trigger auto-revocation.
- **done channel**: A `chan struct{}` on tcpForwardConn, closed when
  readLoop exits. Acts as the bridge between context cancellation (which
  kills the gRPC stream) and cleanup (which revokes the SSH session).

## Success Criteria

### Measurable Outcomes

- **SC-001**: After context cancellation, the SSH session is revoked without
  the caller explicitly calling Close().
- **SC-002**: No resource leaks (goroutines, sessions) when tunnels are
  abandoned via context cancellation.
- **SC-003**: Existing tests continue to pass with no behavioral changes
  for callers who already call Close().
- **SC-004**: Race detector passes with concurrent Close() and context
  cancellation (closeOnce prevents double-revoke, sendMu serializes
  CloseSend() with concurrent Write() calls).

### Non-Functional Requirements

- **NFR-001**: The cleanup mechanism adds at most one goroutine per
  active tunnel. This is acceptable given that tunnels are long-lived
  and few in number.

## Assumptions

- The existing closeOnce pattern in sshTunnel is sufficient for idempotency
  (no new synchronization primitives needed).
- The done channel on tcpForwardConn is reliably closed when readLoop exits
  (verified by reading the current implementation).
- Best-effort revocation is acceptable (matches the existing revokeFunc
  pattern which discards errors).
- One additional goroutine per tunnel is an acceptable cost.
