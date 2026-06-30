# Brainstorm: SSH, TCP Forwarding, and Config (Phase 2b-1)

**Date:** 2026-06-29
**Status:** active
**Depends on:** [005-full-api](005-full-api.md) (Phase 2b scope defined there)

## Problem Framing

Phase 2a delivered operator-useful extensions (services, profiles, credential
refresh). The remaining client-facing RPCs split into two groups. This
brainstorm covers the first group: SSH session management, TCP port
forwarding, and gateway/sandbox configuration. These are low-level
connectivity and configuration primitives that round out the SDK's coverage
of the OpenShell gateway API.

## Scope

5 RPCs across 3 new top-level sub-clients:

| Sub-client | Methods | Proto RPCs |
|------------|---------|------------|
| `SSHInterface` | `CreateSession`, `RevokeSession` | `CreateSshSession`, `RevokeSshSession` |
| `TCPInterface` | `Forward` | `ForwardTcp` |
| `ConfigInterface` | `GetSandbox`, `GetGateway`, `Update` | `GetSandboxConfig`, `GetGatewayConfig`, `UpdateConfig` |

Plus fake client stubs for all three sub-clients.

## Design Decisions

### D1: Top-Level Sub-Clients

SSH, TCP, and Config are all top-level sub-clients on `ClientInterface`,
consistent with the existing pattern (Sandboxes, Exec, Files, Services,
Health, Providers). SSH and TCP take a sandbox name as a parameter, same
as Exec and Files.

```go
client.SSH().CreateSession(ctx, sandboxName, ...)
client.TCP().Forward(ctx, sandboxName, remotePort)
client.Config().GetSandbox(ctx, sandboxName)
```

**Why:** Nesting SSH under Sandboxes would break the flat pattern. Exec
and Files already demonstrate that sandbox-scoped operations work fine
as top-level sub-clients with sandbox name as a parameter.

### D2: SSH and TCP as Separate Sub-Clients

SSH (request/response session management) and TCP (bidirectional streaming)
have different interface shapes and usage patterns. Keeping them separate
keeps each interface small and focused.

### D3: TCP Forward Returns io.ReadWriteCloser

`TCPInterface.Forward` returns an `io.ReadWriteCloser`, not a full
`net.Conn`. The SDK wraps the bidirectional gRPC stream into a simple
read/write/close interface.

**Why:** A full `net.Conn` requires synthesizing network addresses and
wiring deadline semantics through gRPC context cancellation. The actual
use case is piping bytes. If `net.Conn` is needed later, we can add a
`TCPConn` type that embeds the `ReadWriteCloser` and adds the extra
methods. That's additive and non-breaking.

### D4: ConfigInterface for Consistency

Config has only 3 methods but gets its own sub-client rather than being
methods on `Client` directly. Health also has just one method and follows
the same pattern. Uniform structure makes the fake client straightforward
and keeps the `ClientInterface` surface predictable.

## Proposed Interfaces

```go
type SSHInterface interface {
    CreateSession(ctx context.Context, sandboxName string) (*SSHSession, error)
    RevokeSession(ctx context.Context, sessionID string) error
}

type TCPInterface interface {
    Forward(ctx context.Context, sandboxName string, remotePort uint32) (io.ReadWriteCloser, error)
}

type ConfigInterface interface {
    GetSandbox(ctx context.Context, sandboxName string) (*SandboxConfig, error)
    GetGateway(ctx context.Context) (*GatewayConfig, error)
    Update(ctx context.Context, config *ConfigUpdate) error
}
```

## Key Requirements

- All three sub-clients added to `ClientInterface`
- Domain types in `v1/types/` (SSHSession, SandboxConfig, GatewayConfig, ConfigUpdate)
- Converters in `internal/converter/`
- Fake stubs returning Unimplemented (same pattern as Phase 2a fakes)
- Thread-safe, deep copy at boundaries, SPDX headers on all files

## Open Questions

- SSHSession type: what fields does the proto return? (resolve during spec from proto inspection)
- ConfigUpdate: does it use a patch/merge semantic or full replace? (resolve during spec from proto inspection)
- TCP Forward: should there be an option for local port binding, or is raw ReadWriteCloser sufficient for v1?
