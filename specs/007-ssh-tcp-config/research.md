# Research: SSH, TCP Forwarding, and Config (Phase 2b-1)

**Date**: 2026-06-29
**Spec**: [spec.md](spec.md)

## R1: SandboxPolicy Representation

**Decision**: Opaque `[]byte` in `v1/types/`.

**Rationale**: SandboxPolicy is a deeply nested proto message (FilesystemPolicy, LandlockPolicy, ProcessPolicy, NetworkPolicyRule with its own sub-types). Mapping every field to SDK domain types would be a large effort with uncertain value — Config is P2 priority and most consumers just pass policy through. Opaque bytes let consumers who need to inspect policy unmarshal with the generated proto types directly.

**Alternatives considered**:
- Full SDK type hierarchy: Too much surface area for a P2 feature. Can be added later as a non-breaking extension.
- `proto.Message` interface: Violates Constitution I (proto isolation).

## R2: TCP Forward Stream Wrapper

**Decision**: Custom `tcpForwardConn` struct implementing `io.ReadWriteCloser` that wraps the bidirectional gRPC stream.

**Rationale**: The `ForwardTcp` RPC uses `stream TcpForwardFrame` in both directions. Each frame is either a `TcpForwardInit` (first client message) or `bytes data`. The SDK must:
1. Send `TcpForwardInit` with sandbox_id and `TcpRelayTarget{host, port}` as the first frame
2. Wrap subsequent `data` frames as `Read`/`Write` calls
3. Cancel the stream context on `Close()`

The internal `tcpForwardConn` handles this transparently. Consumers see `io.ReadWriteCloser`.

**Alternatives considered**:
- `net.Conn`: Requires synthesizing `LocalAddr`/`RemoteAddr` and wiring `SetDeadline` through gRPC context. Overkill for v1.
- Raw gRPC stream exposure: Violates proto isolation.

## R3: ConfigUpdate Merge Operations

**Decision**: `MergeOperations []byte` field in `ConfigUpdate` (raw proto-encoded bytes).

**Rationale**: The `PolicyMergeOperation` oneof has 6 variants (AddNetworkRule, RemoveNetworkEndpoint, RemoveNetworkRule, AddDenyRules, AddAllowRules, RemoveNetworkBinary). These are policy-domain-specific types that belong in the Phase 2b-2 policy spec. For Phase 2b-1, the Config sub-client passes them through as raw bytes. Consumers who need merge operations construct them using the generated proto types and serialize to bytes.

**Alternatives considered**:
- Full SDK types for each variant: 6+ new types, policy-domain knowledge, better suited for Phase 2b-2.
- Omit merge operations entirely: Would make UpdateConfig less useful for policy mutations.

## R4: SettingValue Type Design

**Decision**: Use a struct with typed getter methods, similar to how `database/sql.NullString` works.

**Rationale**: The proto `SettingValue` is a oneof with `string_value`, `bool_value`, `int_value`, `bytes_value`. In Go, the idiomatic approach is a struct with a `Type` field (string enum) and typed accessors. This avoids interface{}/any and keeps the API type-safe.

```go
type SettingValue struct {
    Type       SettingValueType
    StringVal  string
    BoolVal    bool
    IntVal     int64
    BytesVal   []byte
}
```

**Alternatives considered**:
- `any` field: Loses type safety, requires type assertions.
- Separate types per variant: Too fragmented for a settings map value.

## R5: SSH Session Token Sensitivity

**Decision**: The SSHSession.Token field is returned by CreateSession but MUST NOT be logged or included in error messages. This aligns with Constitution VI (Secrets Never Leak).

**Rationale**: The session token grants SSH access to a sandbox. It is a short-lived credential that should be treated with the same care as API keys.

## R6: Client-Side Port Validation for TCP Forward

**Decision**: Validate port 1-65535 client-side before opening the gRPC stream. Return `InvalidArgument` error.

**Rationale**: Consistent with existing client-side validation pattern (e.g., empty address check in `NewClient`, empty sandbox name in `ExecClient`). Avoids unnecessary network round-trips for obviously invalid inputs.
