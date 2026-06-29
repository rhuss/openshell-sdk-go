# Implementation Plan: SSH Tunneling and TCP Forward Options

**Branch**: `009-ssh-tunnel-forward-opts` | **Date**: 2026-06-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/009-ssh-tunnel-forward-opts/spec.md`

## Summary

Add SSH tunnel support via `SSH().Tunnel()` that combines `CreateSession`
and `ForwardTcp(SshRelayTarget)` into a single call with auto-cleanup.
Fix the TCP `Forward()` method to accept a `WithForwardServiceID` option
that populates the `service_id` field in `TcpForwardInit`. Update fake
clients for interface parity.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: google.golang.org/grpc, github.com/stretchr/testify
**Storage**: N/A
**Testing**: go test + testify (assert/require)
**Target Platform**: Linux/macOS (library)
**Project Type**: Library (Go SDK)
**Performance Goals**: N/A (SDK library, performance bound by gRPC transport)
**Constraints**: Minimal dependencies (Constitution V), SPDX headers on all files
**Scale/Scope**: 2 interface changes, 1 new type, ~8 files modified

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | New types (TunnelOption, ForwardOption) are SDK-level. Proto types stay internal. |
| II. Idiomatic Go | PASS | Functional options pattern matches existing LogOption, WatchOptions. |
| III. Test-First | PASS | Tests will be written alongside implementation per TDD cycle. |
| IV. Upstream Tracking | PASS | Uses existing proto fields (service_id, SshRelayTarget) already in generated code. |
| V. Minimal Dependencies | PASS | No new dependencies. |
| VI. Secrets Never Leak | PASS | SSH token handled internally by Tunnel(); never exposed in errors or logs. The token is passed to authorization_token in the proto but never returned to the caller. |
| VII. Deep Copy at Boundaries | PASS | No new mutable reference types crossing boundaries. |
| VIII. Doc Examples Compile | PASS | quickstart.md examples will match final signatures. |
| IX. Agent-Friendly Documentation | PASS | All new public types/methods will have doc comments with error codes. |
| X. Proto-SDK Naming Fidelity | PASS | ServiceID matches proto service_id semantics. |
| XI. Fake-Real Parity | PASS | Fake Tunnel() will validate port range before returning Unimplemented. |
| XII. Graceful Shutdown Order | PASS | Tunnel.Close(): CloseSend -> cancel -> wait -> revoke session. |

## Project Structure

### Documentation (this feature)

```text
specs/009-ssh-tunnel-forward-opts/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   ├── ssh-interface.md
│   └── tcp-interface.md
└── tasks.md
```

### Source Code (repository root)

```text
openshell/v1/
├── ssh.go                    # SSHInterface: add Tunnel(), TunnelOption, WithTunnelServiceID
├── ssh_client.go             # sshClient: implement Tunnel() with session+forward+cleanup
├── ssh_client_test.go        # Tests for Tunnel()
├── tcp.go                    # TCPInterface: add ForwardOption to Forward() signature
├── tcp_client.go             # tcpClient: apply ForwardOption, set service_id in init frame
├── tcp_client_test.go        # Tests for ForwardOption/service_id
├── client.go                 # Pass sandboxes to newSSHClient
├── fake/
│   ├── ssh.go                # Add Tunnel() stub with port validation
│   ├── ssh_test.go           # Test fake Tunnel()
│   ├── tcp.go                # Accept ForwardOption on Forward()
│   └── tcp_test.go           # Test fake Forward() with options
└── doc.go                    # Update package examples if needed
```

**Structure Decision**: All changes are within the existing `openshell/v1/`
package structure. No new packages or directories needed.

## Implementation Approach

### Phase 1: TCP Forward Options (FR-008, FR-009, FR-011)

Add `ForwardOption` type and `WithForwardServiceID` to `tcp.go`. Modify
`tcpClient.Forward()` to accept variadic options and set `ServiceId` on
the init frame. Update the fake `Forward()` to accept options. This is
the simpler change and establishes the functional options pattern.

**Files**: `tcp.go`, `tcp_client.go`, `tcp_client_test.go`, `fake/tcp.go`,
`fake/tcp_test.go`

### Phase 2: SSH Tunnel (FR-001 through FR-007, FR-010)

Add `TunnelOption`, `WithTunnelServiceID`, and `Tunnel()` to `ssh.go`.
Implement `sshClient.Tunnel()` that: resolves sandbox name via
`sandboxes.Get()`, creates session, opens ForwardTcp stream with SSH
target and auth token, wraps in `sshTunnel` struct with cleanup logic.
Update `client.go` to pass sandbox client to SSH client constructor.
Add fake `Tunnel()` with port validation.

**Files**: `ssh.go`, `ssh_client.go`, `ssh_client_test.go`, `client.go`,
`fake/ssh.go`, `fake/ssh_test.go`

### Key Design Decisions

1. **Sandbox resolution**: `sshClient` gets a `SandboxInterface` injected
   via constructor so it can resolve sandbox name to ID. This follows
   the `GetLogs` pattern in `sandbox_client.go`.

2. **Internal Forward call**: The tunnel method needs to build a
   `TcpForwardInit` with `SshRelayTarget` (not `TcpRelayTarget`). It
   cannot reuse the public `Forward()` method because that hardcodes
   `TcpRelayTarget`. The tunnel method will call the gRPC `ForwardTcp`
   RPC directly, building the init frame with `SshRelayTarget`,
   `authorization_token`, and optional `service_id`.

3. **sshTunnel wrapper**: A small struct that embeds or wraps
   `tcpForwardConn`, adding session revocation on Close. Uses
   `sync.Once` for idempotent close. Revocation uses
   `context.Background()` since the stream context is already cancelled.

4. **Cleanup on failure**: If ForwardTcp fails after CreateSession
   succeeds, defer-based cleanup revokes the session before returning
   the error.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
