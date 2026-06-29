# Implementation Plan: SSH, TCP Forwarding, and Config (Phase 2b-1)

**Branch**: `007-ssh-tcp-config` | **Date**: 2026-06-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/007-ssh-tcp-config/spec.md`

## Summary

Extend the SDK with 6 new RPCs covering SSH session management (2), TCP port forwarding (1), and gateway/sandbox configuration (3). Each domain gets a top-level sub-client on ClientInterface. TCP forward wraps a bidirectional gRPC stream as `io.ReadWriteCloser`. Config uses opaque `[]byte` for SandboxPolicy and merge operations. Fake stubs return Unimplemented.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: None new — SDK itself (`v1/types/`), gRPC, Go stdlib
**Storage**: N/A (SDK wraps gateway RPCs)
**Testing**: Go testing + testify (assert/require), `go test -race`
**Target Platform**: Go library (any OS)
**Project Type**: Library (SDK extension)
**Constraints**: Zero new dependencies, thread-safe, deep copy at boundaries, proto isolation
**Scale/Scope**: ~1200-1600 lines implementation + ~800-1200 lines tests

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | All new domain types in `v1/types/`, converters in `internal/converter/` |
| II. Idiomatic Go | PASS | `io.ReadWriteCloser` for TCP, context propagation, typed accessors |
| III. Test-First (NON-NEGOTIABLE) | PASS | Tests written before each implementation file |
| IV. Upstream Tracking | PASS | All 6 RPCs verified in proto/openshell.proto |
| V. Minimal Dependencies | PASS | Zero new dependencies |
| VI. Secrets Never Leak | PASS | SSHSession.Token is sensitive — never logged or in error messages |
| VII. Deep Copy at Boundaries | PASS | All converter operations deep-copy maps and slices |
| VIII. Doc Examples Compile | PASS | Updated doc.go examples for new sub-clients |
| IX. Agent-Friendly Documentation | PASS | All exported symbols get doc comments with error codes |
| X. Proto-SDK Naming Fidelity | PASS | Field names match proto semantics (sandbox_id→SandboxID, gateway_host→GatewayHost) |

## Project Structure

### Documentation (this feature)

```text
specs/007-ssh-tcp-config/
├── plan.md              # This file
├── research.md          # Research decisions
├── data-model.md        # Entity model
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Task breakdown (created by /speckit.tasks)
```

### Source Code (repository root)

```text
openshell/v1/
├── types/
│   ├── ssh.go               # SSHSession type
│   └── setting.go           # SandboxConfig, GatewayConfig, ConfigUpdate,
│                            # UpdateResult, SettingValue, SettingValueType,
│                            # EffectiveSetting, SettingScope, PolicySource
│                            # (config.go is taken by client Config struct)
├── internal/
│   └── converter/
│       ├── ssh.go           # SSHSession proto↔SDK converter
│       ├── ssh_test.go
│       ├── setting.go       # Config types proto↔SDK converter
│       └── setting_test.go
├── ssh.go                   # SSHInterface definition
├── ssh_client.go            # sshClient implementation
├── ssh_client_test.go       # SSH client tests (mock gRPC)
├── tcp.go                   # TCPInterface definition
├── tcp_client.go            # tcpClient implementation + tcpForwardConn
├── tcp_client_test.go       # TCP client tests (mock gRPC stream)
├── config.go                # ConfigInterface definition
├── config_client.go         # configClient implementation
├── config_client_test.go    # Config client tests (mock gRPC)
├── client.go                # Add SSH(), TCP(), Config() to ClientInterface (existing)
└── doc.go                   # Add SSH/TCP/Config examples (existing)

openshell/v1/fake/
├── ssh.go                   # fakeSSHClient stub
├── ssh_test.go
├── tcp.go                   # fakeTCPClient stub
├── tcp_test.go
├── config.go                # fakeConfigClient stub
├── config_test.go
├── fake.go                  # Add SSH(), TCP(), Config() accessors (existing)
└── provider.go              # Unchanged
```

## Design Decisions

### D1: Top-Level Sub-Clients

SSH, TCP, and Config are top-level sub-clients on `ClientInterface`, consistent with the existing flat pattern. Sandbox ID/name is a parameter, not a scoping accessor.

### D2: TCP Forward as io.ReadWriteCloser

`tcpForwardConn` wraps the bidirectional gRPC stream. On construction:
1. Opens `ForwardTcp` stream
2. Sends `TcpForwardInit` with `sandbox_id` and `TcpRelayTarget{host: "127.0.0.1", port}`
3. Returns the wrapper as `io.ReadWriteCloser`

Read/Write map to stream Recv/Send of `TcpForwardFrame{data}`. Close cancels the stream context.

### D3: Opaque Policy and Merge Operations

`SandboxConfig.Policy` and `ConfigUpdate.Policy` are `[]byte` (serialized proto). `ConfigUpdate.MergeOperations` is `[]byte` (serialized repeated PolicyMergeOperation). Full SDK types for these are deferred to Phase 2b-2.

### D4: SettingValue as Typed Struct

Proto oneof maps to a Go struct with `Type` field (string enum) and typed value fields. Converter sets the appropriate field based on the oneof case.

### D5: No Converter for TCP

TCP has no domain types — it's a stream wrapper. The init frame is constructed directly in `tcpClient.Forward`. No converter file needed.

## Global Constraints

- **FR-015**: All operations MUST be safe for concurrent use.
- **FR-013**: All domain types in `v1/types/`, converters import types not clients.
- **FR-016**: Typed StatusError with appropriate ErrorCode for all error paths.
- **Constitution III**: Tests written before each implementation file.
- **SPDX headers**: Every `.go` file must start with the SPDX license header.
- **Deep copy at boundaries**: All converter operations deep-copy maps and slices.

## Implementation Order

Bottom-up: types → converters → client implementations → interface wiring → fake stubs → polish.

1. **Types**: New domain types in `v1/types/` (ssh, config)
2. **Converters**: Proto↔SDK converters for SSH and Config types
3. **SSH sub-client**: SSHInterface + sshClient
4. **TCP sub-client**: TCPInterface + tcpClient + tcpForwardConn
5. **Config sub-client**: ConfigInterface + configClient
6. **Interface wiring**: Add SSH(), TCP(), Config() to ClientInterface and Client constructor
7. **Fake stubs**: SSH, TCP, Config stubs + fake wiring
8. **Polish**: doc.go updates, type re-exports, lint, CI verification
