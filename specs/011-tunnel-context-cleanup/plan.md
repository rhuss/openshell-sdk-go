# Implementation Plan: SSH Tunnel Context Cancellation Cleanup

**Branch**: `011-tunnel-context-cleanup` | **Date**: 2026-06-30 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/011-tunnel-context-cleanup/spec.md`

## Summary

Add a background cleanup goroutine to `sshTunnel` that watches the
`tcpForwardConn.done` channel and calls `Close()` when it fires. This
ensures SSH sessions are auto-revoked when the parent context is cancelled,
even if the caller never explicitly calls `Close()`. The existing
`closeOnce` pattern guarantees idempotency.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: gRPC, protobuf (existing)
**Storage**: N/A
**Testing**: Go testing + testify (assert/require)
**Target Platform**: Linux/macOS (library)
**Project Type**: Library (Go SDK)
**Performance Goals**: N/A (correctness fix, no performance-sensitive path)
**Constraints**: One additional goroutine per tunnel (acceptable for long-lived tunnels)
**Scale/Scope**: Two file changes (`ssh_client.go`, `tcp_client.go`) + test additions

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | No proto changes needed |
| II. Idiomatic Go | PASS | Goroutine + channel watching is idiomatic Go cleanup |
| III. Test-First | PASS | Tests written for all acceptance scenarios |
| V. Minimal Dependencies | PASS | No new dependencies |
| VI. Secrets Never Leak | PASS | Token handling unchanged |
| IX. Agent-Friendly Docs | PASS | Doc comment on cleanup goroutine |
| XII. Graceful Shutdown | PASS | Close() already does protocol close before cancel; cleanup goroutine calls Close() which preserves this order |

No violations. No complexity tracking needed.

## Project Structure

### Documentation (this feature)

```text
specs/011-tunnel-context-cleanup/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output (minimal for this fix)
└── tasks.md             # Phase 2 output
```

### Source Code (repository root)

```text
openshell/v1/
├── ssh_client.go        # MODIFY: add cleanup goroutine to sshTunnel
├── tcp_client.go        # MODIFY: add sendMu lock around CloseSend()
└── ssh_client_test.go   # MODIFY: add tests for auto-revoke on context cancel
```

**Structure Decision**: No new files needed. This is a targeted fix to
existing code in `ssh_client.go` with corresponding test additions.

## Design

### Change to `sshTunnel`

The `sshTunnel` struct gains no new fields. The cleanup goroutine is
launched in the `Tunnel()` method immediately after constructing the
`sshTunnel` value, before returning it to the caller.

```
Tunnel() creates sshTunnel
  └── launches cleanup goroutine
       └── blocks on <-conn.done
            └── when done closes: calls sshTunnel.Close()
                 └── closeOnce ensures exactly-once execution
                      ├── tcpForwardConn.Close() (protocol close)
                      └── revokeFunc() (best-effort session revoke)
```

### Interaction Matrix

| Scenario | Cleanup goroutine | Explicit Close() | Result |
|----------|------------------|-------------------|--------|
| Context cancelled, no Close() | Calls Close() | Never called | Session revoked (goroutine) |
| Context cancelled, then Close() | Calls Close() first | closeOnce no-op | Session revoked once |
| Close() called, no cancel | closeOnce no-op (done closes after Close) | Calls Close() | Session revoked (explicit) |
| Close() + cancel race | closeOnce arbitrates | closeOnce arbitrates | Session revoked once |

### Why `done` channel, not `streamCtx.Done()`

The cleanup goroutine watches `conn.done` (closed by readLoop) rather
than `streamCtx.Done()` (cancelled when context fires). Reason: `done`
fires after the readLoop has fully exited, which means the gRPC stream
is completely shut down. Watching `streamCtx.Done()` would race with
readLoop's exit and could trigger Close() while the stream is still
draining.
