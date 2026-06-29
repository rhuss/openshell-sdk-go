# Research: SSH Tunneling and TCP Forward Options

**Feature**: 009-ssh-tunnel-forward-opts
**Date**: 2026-06-29

## R1: Sandbox Name-to-ID Resolution Pattern

**Decision**: The tunnel method resolves sandbox name to ID by calling
`sandboxClient.Get(ctx, name)` and extracting `sb.ID`, matching the
existing `GetLogs` pattern.

**Rationale**: The SDK convention uses sandbox names in developer-facing
methods. The proto RPCs (`CreateSshSession`, `ForwardTcp`) take sandbox
IDs. The `GetLogs` method in `sandbox_client.go:237-242` demonstrates
the established resolution pattern: call `Get()` first, use `sb.ID`.

**Alternatives considered**:
- Expose sandbox ID on the tunnel method signature: rejected because it
  breaks the SDK's name-based convention.
- Add a dedicated resolve method to SandboxInterface: rejected because
  `Get()` already does this; a separate method adds API surface without
  value.

**Implementation impact**: The `sshClient` struct needs access to the
sandbox client (or the gRPC connection to create its own sandbox
lookup). The cleanest approach is to pass the `SandboxInterface` to
`newSSHClient` so it can call `Get()` during tunnel setup. This mirrors
how `GetLogs` is on `sandboxClient` which already has `Get()` available.

## R2: Functional Options Pattern for Go

**Decision**: Use the standard Go functional options pattern with
separate option types per sub-client: `TunnelOption` for SSH tunnel,
`ForwardOption` for TCP forward.

**Rationale**: Follows Constitution II (Idiomatic Go). The existing SDK
uses this pattern for `LogOption`, `WatchOptions`, etc. Separate types
per method prevent cross-contamination of option namespaces (a
`ForwardOption` should not be passable to `Tunnel` or vice versa).

**Alternatives considered**:
- Shared option type: rejected because it couples SSH and TCP sub-client
  APIs. If TCP forward later gets options SSH tunnel doesn't need (or
  vice versa), shared options force awkward "this option is ignored"
  documentation.
- Options struct: rejected because it doesn't compose well for future
  options and isn't consistent with the existing `LogOption` pattern.

## R3: Tunnel Close/Cleanup Lifecycle

**Decision**: The tunnel wraps a `tcpForwardConn` (or embeds it) and
adds session revocation on Close. Close order: (1) close gRPC stream via
`CloseSend()`, (2) cancel context, (3) wait for background goroutine,
(4) revoke SSH session. Use `sync.Once` for idempotent close.

**Rationale**: Constitution XII mandates graceful shutdown order:
protocol close before context cancel. The existing `tcpForwardConn.Close`
follows this order. The tunnel extends it with session revocation as the
final step. Session revocation uses a background context (not the
stream context) since the stream context is already cancelled at that
point.

**Alternatives considered**:
- Revoke before stream close: rejected because it could cause the
  gateway to terminate the stream server-side, leading to noisy recv
  errors.
- Skip revocation on Close: rejected because it leaks sessions that
  eventually expire but waste gateway resources in the meantime.

## R4: Error Handling During Tunnel Setup

**Decision**: If `CreateSession` succeeds but `ForwardTcp` fails, the
tunnel MUST revoke the session before returning the error. Use a cleanup
function pattern with `defer` to ensure revocation happens on all error
paths.

**Rationale**: Session leaks are the primary risk in the tunnel
lifecycle. The brainstorm explicitly flags this as a key edge case. A
defer-based cleanup is the idiomatic Go approach.

**Implementation pattern**:
```
session, err := ssh.CreateSession(ctx, sandboxID)
if err != nil { return nil, err }

// Clean up session if anything below fails
var success bool
defer func() {
    if !success {
        ssh.RevokeSession(context.Background(), session.Token)
    }
}()

conn, err := tcp.Forward(ctx, sandboxID, port, ...)
if err != nil { return nil, err }

success = true
return &sshTunnel{conn: conn, session: session, ...}, nil
```

## R5: Service ID Proto Field Mapping

**Decision**: The `service_id` field in `TcpForwardInit` maps directly
to the `ServiceId` field in the generated Go proto. Set it from the
functional option when building the init frame.

**Rationale**: The field is a simple string, already present in the
generated proto bindings. No conversion needed.

**Current state**: `tcp_client.go:40-52` builds the `TcpForwardInit`
frame but does not set `ServiceId`. The fix adds `ServiceId: opts.serviceID`
to the init frame construction.
