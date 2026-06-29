# Research: Phase 2b-2 Policy, Logs, MergeOps, ErrorConflict

## R1: Proto-to-SDK Mapping for PolicyChunk (18 fields)

**Decision**: Full fidelity, flat struct. All 18 proto fields mapped 1:1.

**Rationale**: The SDK is a thin mapping layer. PolicyChunk is a read-only type returned by GetDraft — consumers need all fields for governance UIs. Grouping analysis fields into a nested struct was rejected because it adds an accessor hop without clarity gain for a type that consumers will likely serialize to JSON as-is.

**Alternatives considered**:
- Nested ChunkAnalysis struct (rejected: adds complexity without reducing API surface)
- Trimmed to core fields only (rejected: hides server-returned data consumers may need)

## R2: NetworkPolicyRule Type Hierarchy Depth

**Decision**: Full fidelity mapping of the sandbox.v1 proto hierarchy. All types mapped: NetworkPolicyRule, NetworkEndpoint (18 fields), NetworkBinary, L7Rule, L7Allow, L7DenyRule, L7QueryMatcher, GraphqlOperation.

**Rationale**: MergeOperations (AddNetworkRule, EditDraftChunk) reference NetworkPolicyRule. Consumers need to construct these types to use the policy management APIs. Partial mapping would force consumers to drop to raw proto, violating Constitution I (Proto Isolation).

**Alternatives considered**:
- Opaque bytes for NetworkPolicyRule (rejected: forces proto dependency on consumers)
- Partial mapping (top-level fields only) (rejected: L7 rules are needed for meaningful policy edits)

## R3: GetLogs Name-to-ID Resolution

**Decision**: GetLogs accepts sandbox name and resolves to ID via Sandboxes().Get() internally.

**Rationale**: The proto GetSandboxLogsRequest uses sandbox_id, but every other SDK method uses sandbox name. Consistency trumps the extra RPC round-trip. The Sandbox.ID field already exists.

**Alternatives considered**:
- Accept ID directly (rejected: breaks SDK consistency, exposes proto detail)
- Accept both (rejected: two code paths, unclear which is canonical)

## R4: Error Code for Optimistic Concurrency

**Decision**: Add ErrorConflict for gRPC ABORTED only. Do not map FAILED_PRECONDITION.

**Rationale**: The proto explicitly documents ABORTED for expected_resource_version mismatch. FAILED_PRECONDITION has different gRPC semantics (state precondition, not version conflict). No evidence the gateway returns FAILED_PRECONDITION for anything callers need to distinguish.

**Alternatives considered**:
- Map both to ErrorConflict (rejected: loses semantic distinction)
- Two codes: ErrorAborted + ErrorPreconditionFailed (rejected: speculative, no gateway usage evidence)

## R5: MergeOperations Type Design

**Decision**: Struct-per-variant with pointer fields. One non-nil field indicates the active variant.

**Rationale**: Matches the proto oneof pattern. Type-safe at compile time. The converter validates exactly-one-non-nil and maps to the proto oneof. Go does not have sum types, so a struct with pointer fields is the idiomatic pattern (see Kubernetes API machinery).

**Alternatives considered**:
- Interface-based polymorphism (rejected: adds complexity for consumers constructing operations)
- Functional options (rejected: unconventional for batch data operations)

## R6: Functional Options for Filter/Query Parameters

**Decision**: Functional options for GetDraft (WithStatusFilter), ApproveAllDraftChunks (WithIncludeSecurityFlagged), GetStatus (WithVersion), List (WithLimit, WithOffset), and GetLogs (WithLogLines, WithLogSince, WithLogSources, WithLogMinLevel).

**Rationale**: Consistent with Constitution II (Idiomatic Go). Optional parameters via functional options keeps method signatures clean and extensible. Matches the existing pattern used by other SDK methods.

**Alternatives considered**:
- Params struct (rejected: less idiomatic, requires constructing a struct even for no-filter calls)
- Bare parameters (rejected: unreadable for bool flags like include_security_flagged)

## R7: Existing Converter Patterns

**Findings from codebase inspection**:
- `converter.TimeFromMillis` / `converter.MillisFromTime` already exist in `time.go` — reuse for all timestamp conversions
- `converter.CopyStringSlice` / `converter.CopyStringMap` exist in `copy.go` — reuse for deep copy
- Each converter file follows the pattern: `TypeFromProto(proto) *sdk` and `TypeToProto(sdk) *proto`
- Round-trip tests assert proto→SDK→proto equality
- Proto imports use aliases: `pb "...openshellv1"`, `sbv1 "...sandboxv1"`
