# Brainstorm: SSH Tunneling and TCP Forward Options (Phase 2b-3)

**Date:** 2026-06-29
**Status:** active
**Depends on:** [008-ssh-tcp-config](008-ssh-tcp-config.md) (Phase 2b-1 delivered SSH/TCP/Config sub-clients)

## Problem Framing

Phase 2b-1 delivered SSH session management (CreateSession/RevokeSession) and
TCP port forwarding (Forward returning io.ReadWriteCloser). However, the two
halves don't connect: there is no way to tunnel SSH through the gateway using
the SDK.

The proto's `TcpForwardInit` has a `target` oneof with both `SshRelayTarget`
and `TcpRelayTarget`. The current `Forward()` implementation hardcodes
`TcpRelayTarget` and ignores three proto fields:

- `target` oneof (SSH vs TCP relay)
- `authorization_token` (SSH session token from CreateSshSession)
- `service_id` (audit/correlation identifier)

A developer who wants an SSH tunnel today must manually orchestrate
CreateSession, build a raw gRPC ForwardTcp stream with SSH target, pass the
token, and handle cleanup. The SDK should make this a single call.

## Approaches Considered

### A: SSH().Tunnel() + Forward() Gap Fix (Chosen)

Add a high-level `SSH().Tunnel()` method that combines CreateSession +
ForwardTcp(SshRelayTarget) into one call with auto-revoke on Close. Fix the
`service_id` gap on Forward() with a functional option. Keep SSH target/auth
token internal to Tunnel(), not exposed on Forward().

- Pros: Clean separation. Tunnel() is the simple path, Forward() stays
  focused on TCP. Minimal new API surface. YAGNI on Forward() SSH options.
- Cons: If someone needs session sharing across tunnels, they can't use
  Tunnel(). But that's additive later.

### B: Extend Forward() Only

Add WithSSHTarget(token), WithServiceID(id) to Forward(). No high-level
Tunnel() helper. Consumers orchestrate CreateSession + Forward() manually.

- Pros: Full proto coverage. One method handles both target types.
- Cons: Poor UX. Every SSH tunnel requires 3+ calls and manual cleanup.
  Misses the point of an SDK.

### C: Full Stack (Tunnel + Forward + net.Listener)

Everything in A, plus a local net.Listener that accepts connections and
tunnels each one through Forward/Tunnel (like `ssh -L`), plus reverse
forwarding.

- Pros: Most complete. Covers all SSH tunneling patterns.
- Cons: Large scope. Local listener and reverse forwarding are different
  features that deserve their own brainstorm. YAGNI for v1.

## Decision

**Approach A.** SSH().Tunnel() for the common case, Forward() gap fix for
audit. Local listener and reverse forwarding deferred to idea inbox.

## Key Requirements

### SSH().Tunnel() (new method on SSHInterface)

```go
type SSHInterface interface {
    CreateSession(ctx context.Context, sandboxID string) (*SSHSession, error)
    RevokeSession(ctx context.Context, token string) (bool, error)
    // New:
    Tunnel(ctx context.Context, sandboxName string, port uint32, opts ...TunnelOption) (io.ReadWriteCloser, error)
}
```

- Internally: CreateSession -> ForwardTcp(SshRelayTarget, token) -> stream
- Close() does: close stream, then revoke session (graceful shutdown order
  per Constitution XII: protocol close before context cancel)
- Returns `io.ReadWriteCloser` (simple, consistent with Forward())
- Functional option: `WithServiceID(id)` for audit/correlation

### TCP().Forward() Gap Fix

- Add `WithServiceID(id)` functional option
- Passes `service_id` through to `TcpForwardInit`
- No other changes to Forward() signature or behavior

### Fake Stubs

- Fake `Tunnel()` on existing fake SSH client (returns Unimplemented)
- Forward() options flow through existing fake without changes

### Out of Scope

- SSH target / auth token options on Forward() (YAGNI, Tunnel covers it)
- Session sharing across tunnels (additive later if needed)
- net.Conn upgrade (deferred in brainstorm 008, still deferred)
- Local port listener / net.Listener (added to idea inbox)
- Reverse forwarding (added to idea inbox)

## Open Questions

- Tunnel() uses sandbox name (consistent with SDK) but CreateSession takes
  sandbox_id. Does Tunnel need to resolve name -> ID internally, like
  GetLogs does? (resolve during spec from proto inspection)
- Should WithServiceID be a shared option type across SSH and TCP, or
  separate per-interface option types? (resolve during spec)
