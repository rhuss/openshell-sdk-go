# Brainstorm: Policy Management and Logs (Phase 2b-2)

**Date:** 2026-06-29
**Status:** active
**Depends on:** [005-full-api](005-full-api.md) (Phase 2b scope defined there)

## Problem Framing

The second half of Phase 2b covers policy management and sandbox log
retrieval. The policy domain is the largest remaining API surface: 10
RPCs implementing a draft-review-approve workflow for sandbox security
policies. Log retrieval is a single RPC that fits naturally on the
existing SandboxInterface.

## Scope

11 RPCs: 10 new on PolicyInterface, 1 added to SandboxInterface.

| Sub-client | Methods | Proto RPCs |
|------------|---------|------------|
| `PolicyInterface` | `GetStatus`, `List`, `GetDraft`, `ApproveDraftChunk`, `RejectDraftChunk`, `ApproveAllDraftChunks`, `EditDraftChunk`, `UndoDraftChunk`, `ClearDraftChunks`, `GetDraftHistory` | `GetSandboxPolicyStatus`, `ListSandboxPolicies`, `GetDraftPolicy`, `ApproveDraftChunk`, `RejectDraftChunk`, `ApproveAllDraftChunks`, `EditDraftChunk`, `UndoDraftChunk`, `ClearDraftChunks`, `GetDraftHistory` |
| `SandboxInterface` (extended) | `GetLogs` | `GetSandboxLogs` |

Plus fake client stubs for PolicyInterface and the GetLogs extension.

## Design Decisions

### D1: Flat PolicyInterface

All 10 policy RPCs live on a single flat `PolicyInterface`. The draft
operations (approve, reject, edit, undo, clear, history) are not nested
under a separate `DraftPolicyInterface`.

**Why:** All operations are scoped to a sandbox. Nesting adds an accessor
without clarity gain. The interface is large (10 methods) but cohesive:
it represents the full policy management workflow.

### D2: Top-Level Policy Sub-Client

`PolicyInterface` is a top-level sub-client on `ClientInterface`:
`client.Policy().GetStatus(ctx, sandboxName)`.

**Why:** Consistent with the flat sub-client pattern. All methods take
sandbox name as a parameter.

### D3: GetLogs on SandboxInterface

`GetSandboxLogs` is added to the existing `SandboxInterface` as
`GetLogs(ctx, sandboxName)`.

**Why:** Logs are a property of the sandbox. Creating a single-method
LogsInterface would be unnecessary overhead. This also avoids adding
scope to the Policy spec.

## Proposed Interfaces

```go
type PolicyInterface interface {
    GetStatus(ctx context.Context, sandboxName string) (*PolicyStatus, error)
    List(ctx context.Context, sandboxName string, opts ...ListOptions) ([]*PolicyRevision, error)
    GetDraft(ctx context.Context, sandboxName string) (*DraftPolicy, error)
    ApproveDraftChunk(ctx context.Context, sandboxName string, chunkID string) error
    RejectDraftChunk(ctx context.Context, sandboxName string, chunkID string) error
    ApproveAllDraftChunks(ctx context.Context, sandboxName string) error
    EditDraftChunk(ctx context.Context, sandboxName string, chunkID string, edit *ChunkEdit) error
    UndoDraftChunk(ctx context.Context, sandboxName string, chunkID string) error
    ClearDraftChunks(ctx context.Context, sandboxName string) error
    GetDraftHistory(ctx context.Context, sandboxName string) (*DraftHistory, error)
}

// Added to existing SandboxInterface:
// GetLogs(ctx context.Context, sandboxName string, opts ...LogOptions) ([]*LogLine, error)
```

## Key Requirements

- PolicyInterface added to `ClientInterface` as top-level sub-client
- GetLogs added to existing `SandboxInterface`
- Domain types in `v1/types/` (PolicyStatus, PolicyRevision, DraftPolicy,
  PolicyChunk, ChunkEdit, DraftHistory, LogLine, LogOptions)
- Converters in `internal/converter/`
- Fake stubs for PolicyInterface (returning Unimplemented)
- Fake GetLogs on existing fake SandboxClient
- Thread-safe, deep copy at boundaries, SPDX headers on all files

## Open Questions

- NetworkPolicyRule and related sandbox.v1 types: inspect sandbox proto for exact SDK type shape (resolve during spec)
- ListSandboxPolicies has a `global` flag: should SDK expose this or keep sandbox-scoped only? (resolve during spec)

---

## Revisit: 2026-06-29

### Updated Problem Framing

Proto inspection and PR #5 review feedback revealed three additional
concerns beyond the original 11 RPCs:

1. **MergeOperations gap**: `ConfigUpdate.MergeOperations` is `[]byte`
   with a runtime rejection. The proto defines a full
   `PolicyMergeOperation` oneof with 6 operation types. This needs typed
   SDK structs, not raw bytes.
2. **Optimistic concurrency gap**: `codes.Aborted` (used for
   `expected_resource_version` mismatch) maps to `ErrorInternal`. Callers
   cannot distinguish concurrency conflicts programmatically.
3. **Proto details answered open questions**: PolicyChunk has 18 fields,
   GetDraftPolicy has `status_filter`, RejectDraftChunk has `reason`,
   ApproveAllDraftChunks has `include_security_flagged`, GetSandboxLogs
   uses `sandbox_id` (not name) with filtering by sources/level/since/lines.

### Expanded Scope

Single spec covering all of Phase 2b-2:

| Area | Details |
|------|---------|
| **PolicyInterface** | 10 methods: GetStatus, List, GetDraft, ApproveDraftChunk, RejectDraftChunk, ApproveAllDraftChunks, EditDraftChunk, UndoDraftChunk, ClearDraftChunks, GetDraftHistory |
| **GetLogs** | Added to SandboxInterface. Takes sandbox name, resolves name→id internally. Functional options for filtering. |
| **MergeOperations** | Replace `[]byte` with `[]PolicyMergeOperation` (typed structs for 6 oneof variants). Update ConfigUpdate, converter, and fake. |
| **ErrorConflict** | New error code + `IsConflict()` helper. Map `codes.Aborted` → `ErrorConflict`. |
| **Domain types** | PolicyChunk (18 fields, flat), DraftPolicy, PolicyStatus, SandboxPolicyRevision, DraftHistoryEntry, LogLine, LogResult, ApproveAllResult, plus MergeOperation variant structs |
| **Fake stubs** | PolicyInterface fake, GetLogs fake, MergeOperations in fake ConfigClient |
| **Functional options** | WithIncludeSecurityFlagged, WithStatusFilter, WithLogLines/Since/Sources/MinLevel |

### New Design Decisions

#### D4: ErrorConflict for ABORTED Only

Add `ErrorConflict` to `ErrorCode` enum. Map `codes.Aborted` →
`types.ErrorConflict`. Add `IsConflict()` helper. Leave
`FAILED_PRECONDITION` mapping to `ErrorInternal` — no evidence the
gateway returns it for anything callers need to distinguish.

**Why:** The proto explicitly documents `ABORTED` for resource_version
mismatch. Callers need this for read-modify-write retry loops.
`FAILED_PRECONDITION` has different gRPC semantics (state precondition,
not version conflict).

#### D5: Typed MergeOperations (Struct-per-Variant)

Replace `MergeOperations []byte` with `[]PolicyMergeOperation` where
each variant is a concrete struct:

```go
type PolicyMergeOperation struct {
    AddRule        *AddNetworkRule
    RemoveEndpoint *RemoveNetworkEndpoint
    RemoveRule     *RemoveNetworkRule
    AddDenyRules   *AddDenyRules
    AddAllowRules  *AddAllowRules
    RemoveBinary   *RemoveNetworkBinary
}
```

**Why:** Type-safe, matches the proto oneof pattern. The converter
handles mapping. Consumers get compile-time checking instead of
opaque bytes.

#### D6: PolicyChunk — Full Fidelity, Flat Struct

All 18 proto fields mapped 1:1. No nested metadata struct.

**Why:** The SDK is a thin mapping layer. Hiding fields means consumers
can't access data the API returns. Governance UIs may need any of
the analysis fields (confidence, hit_count, validation_result, etc.).

#### D7: GetLogs — Name Parameter, Internal ID Resolution

`GetLogs(ctx, sandboxName, opts ...LogOption)` on SandboxInterface.
Internally resolves name → ID via `Get(name)` before calling
`GetSandboxLogs` RPC.

**Why:** Consistent with the rest of the SDK which uses sandbox name
everywhere. The `Sandbox.ID` field already exists.

#### D8: ApproveAllDraftChunks — Functional Option for Force

`WithIncludeSecurityFlagged()` functional option. Default is safe
(skips security-flagged chunks).

**Why:** Consistent with the functional options pattern used for
GetDraft and GetLogs. Bare bool parameters are less readable.

#### D9: GetDraft — StatusFilter via Functional Option

`WithStatusFilter(status)` functional option on GetDraft. Default
returns all chunks.

### Out of Scope

- SubmitPolicyAnalysis (supervisor-internal RPC)
- ReportPolicyStatus (supervisor-internal RPC)
- PushSandboxLogs (supervisor-internal RPC)
- FAILED_PRECONDITION error mapping
- Fake reactors
- NetworkPolicyRule type hierarchy details (resolved during spec from sandbox proto)

### Resolved Open Questions

- **DraftPolicy type**: PolicyChunk with 18 fields, chunks as `[]PolicyChunk` (from proto)
- **ChunkEdit**: EditDraftChunk takes a replacement `NetworkPolicyRule` for `proposed_rule` (from proto)
- **ApproveAllDraftChunks Force**: `include_security_flagged` bool exposed as functional option
- **LogOptions**: lines, since (time.Time), sources ([]string), min_level (string) — all via functional options
- **PolicyRevision vs PolicyStatus**: distinct types — `SandboxPolicyRevision` (versioned record) vs `PolicyStatus` enum (load state)
