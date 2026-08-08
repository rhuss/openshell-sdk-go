# Research: Reverse Port Forwarding

**Feature**: 025-reverse-port-forwarding
**Date**: 2026-08-05

## R1: TCPInterface Extension Pattern

**Decision**: Add `RemoteListen` to the existing `TCPInterface` interface with the same parameter conventions as `Forward` and `Listen`.

**Rationale**: The existing `TCPInterface` in `openshell/v1/tcp.go` has two methods (`Forward`, `Listen`) that both take `(ctx, workspace, sandboxName, ...)`. Adding `RemoteListen` with the same pattern maintains interface consistency. The functional options pattern (`RemoteListenOption`) follows `ForwardOption` and `ListenOption` exactly.

**Alternatives considered**:
- Separate `ReverseTCPInterface`: Would fragment the TCP sub-client unnecessarily. The SDK's pattern is one interface per sub-client (`TCPInterface`, `SSHInterface`, `ExecInterface`).
- Method on `SSHInterface`: Reverse forwarding is conceptually TCP, not SSH. SSH is a transport option for forward tunneling, not a separate forwarding model.

## R2: Fake Client Implementation

**Decision**: Fake `RemoteListen` validates inputs then returns `Unimplemented`, matching `Forward` and `Listen` in `fake/tcp.go`.

**Rationale**: The fake TCP client (`fakeTCPClient`) validates port ranges, empty names, and closed state before returning `Unimplemented`. `RemoteListen` follows the same pattern but adds `localTarget` validation via `net.SplitHostPort`.

**Alternatives considered**:
- Fake that simulates bridging: Over-engineering for a method that requires a real sandbox runtime. `Listen` and `Forward` both return `Unimplemented` in the fake.

## R3: localTarget Validation

**Decision**: Use `net.SplitHostPort` from Go stdlib. If it returns an error, return `InvalidArgument`.

**Rationale**: `net.SplitHostPort` handles all standard host:port formats including IPv6 brackets (`[::1]:8080`). It validates syntax without attempting resolution, which is correct for input validation (resolution is a runtime concern).

**Alternatives considered**:
- Custom regex: Fragile, doesn't handle edge cases like IPv6.
- `net.Dial` probe: Would cause side effects during validation. The local target may not be running yet when `RemoteListen` is called.

## R4: Real Client Stub Behavior

**Decision**: Real client's `RemoteListen` returns `Unimplemented` with a clear message indicating that the upstream proto extension is required.

**Rationale**: There is no `ReverseTcp` or `WaitForReverse` RPC in the current proto definitions. The real client cannot do anything meaningful. Returning `Unimplemented` is consistent with how the SDK would behave if the gateway didn't support the RPC. When proto support lands, the stub is replaced with the real implementation.

**Alternatives considered**:
- Omit from real client entirely: Would break `TCPInterface` since both real and fake must implement it.
- Panic: Violates SDK error handling principles.

## R5: Workspace Parameter Position

**Decision**: `workspace` is the second parameter after `ctx`, before `sandboxName`, matching `Forward` and `Listen`.

**Rationale**: PR #41 established workspace as a mandatory parameter on all RPCs. The position `(ctx, workspace, sandboxName, ...)` is consistent across all sub-client methods in the SDK.
