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

- DraftPolicy type: how are chunks represented? (resolve during spec from proto inspection)
- ChunkEdit: what fields can be edited in-place? (resolve during spec from proto inspection)
- ApproveAllDraftChunks: the proto mentions "skips security-flagged unless forced" -- should Force be a parameter?
- LogOptions: filtering by time range, severity, or line count? (resolve during spec from proto inspection)
- PolicyRevision vs PolicyStatus: are these the same type or distinct? (resolve during spec from proto inspection)
