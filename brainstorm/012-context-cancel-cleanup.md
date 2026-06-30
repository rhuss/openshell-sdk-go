# Brainstorm: Context Cancellation Session Cleanup

**Date:** 2026-06-30
**Status:** active

## Problem Framing

When a parent context is cancelled on an active SSH tunnel, the gRPC stream
terminates and the readLoop goroutine exits, but `sshTunnel.Close()` is never
called automatically. This means the SSH session token remains alive on the
server until it expires. Two independent code reviewers (copilot and devin)
flagged this gap in PR #7 between the spec's acceptance scenario 3 ("context
cancellation revokes the session") and the implementation (Go convention:
caller must call Close()).

SSH sessions are server-side resources. If a caller's context is cancelled in
a crash path or timeout, the caller may never reach `defer tunnel.Close()`,
causing silent resource leaks that are invisible and hard to debug.

## Approaches Considered

### A: Auto-revoke on context cancel

A background goroutine watches the `done` channel (closed when readLoop exits)
and calls `Close()`, which triggers session revocation. Since `Close()` uses
`closeOnce`, explicit `Close()` calls remain safe and idempotent.

- Pros: matches spec, prevents resource leaks, minimal implementation (one goroutine)
- Cons: adds a goroutine per tunnel

### B: Caller must call Close()

Standard Go pattern. Document that `Close()` is mandatory even after context
cancellation. Accept the spec gap.

- Pros: follows Go convention, no additional complexity
- Cons: violates spec acceptance scenario 3, leaks sessions on crash/timeout paths

### C: Auto-revoke + Close() safety net

Same as A but with explicit emphasis that `Close()` remains valid and
idempotent for callers who `defer tunnel.Close()` as standard practice.

- Pros: best of both worlds, matches spec, safe for Go idioms
- Cons: marginally more complex (one goroutine)

## Decision

Option C: auto-revoke on context cancel with idempotent Close(). The existing
`closeOnce` pattern already protects against double-revoke, so adding a
background goroutine watching the `done` channel is the only change needed.

## Key Requirements

- Auto-revoke SSH session when parent context is cancelled
- Close() remains idempotent (closeOnce already handles it)
- SSH Tunnel only, not TCP Forward (no session to revoke)
- Best-effort revocation (fire-and-forget with context.Background(), matches existing revokeFunc pattern)
- Implementation: goroutine watching the `done` channel, calls Close() on exit

## Open Questions

- None. The scope is well-defined and the implementation path is clear.
