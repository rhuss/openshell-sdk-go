# Brainstorm: Add Workspace Scoping to All RPCs

**Date:** 2026-07-31
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/33

## Problem Framing

Every existing SDK sub-client (Sandboxes, Providers, Services, Exec, Files, Config, Policy) operates without workspace context. The upstream OpenShell gateway scopes all user-facing RPCs to a workspace (empty string = "default" workspace). The dashboard BFF passes `workspace` on every call. Without workspace scoping, the SDK cannot be used as a drop-in replacement for the dashboard's current gateway layer.

This is the highest-priority SDK gap: issues #32 (Workspace CRUD) and #34 (GatewayInfo) add new sub-clients, but #33 affects every existing sub-client.

## Approaches Considered

### A: Workspace parameter on every method

Add `workspace string` as the first parameter to every sub-client method:

```go
client.Sandboxes().Create(ctx, "my-workspace", "sandbox-name", spec, labels)
client.Sandboxes().List(ctx, "my-workspace", opts...)
client.Providers().Get(ctx, "my-workspace", "provider-name")
```

- Pros: explicit, no hidden state, matches how the dashboard BFF works
- Cons: breaks all existing callers, verbose for CLI tools that set workspace once

### B: Workspace in ListOptions / per-call options

Add workspace to the variadic options pattern:

```go
client.Sandboxes().List(ctx, openshell.WithWorkspace("ws"), openshell.WithLimit(10))
client.Sandboxes().Create(ctx, "name", spec, labels, openshell.WithWorkspace("ws"))
```

- Pros: backward-compatible (empty workspace = "default"), consistent with existing options pattern
- Cons: workspace is required for most calls but looks optional, easy to forget

### C: Workspace-scoped sub-client factory

Add a `Workspace(name)` method that returns a scoped view:

```go
ws := client.Workspace("my-workspace")
ws.Sandboxes().Create(ctx, "sandbox-name", spec, labels)
ws.Sandboxes().List(ctx, opts...)
```

- Pros: set workspace once, clean API, matches Kubernetes client-go namespace pattern
- Cons: more structural change, need to decide if unscoped client still works for "default" workspace

## Decision

**Approach A: Explicit workspace parameter.** It's the simplest, most explicit approach. The dashboard BFF always knows the workspace from the URL path, so passing it explicitly is natural. Empty string means "default" workspace, matching the gateway convention. Breaking existing callers is acceptable since the SDK is pre-1.0 and the workspace scoping is a fundamental API change.

For methods that take `sandbox_id` (UUID) rather than name (GetSandboxLogs, ExecSandbox, WatchSandbox), the workspace parameter may not be needed on the proto level, but including it for consistency is fine since the gateway accepts it.

## Key Requirements

1. Add `workspace string` parameter to all sub-client interface methods
2. Update all converter functions to set workspace fields on proto requests
3. Update fake client implementations to respect workspace scoping (store per workspace)
4. Update all existing tests
5. Update `ListOptions` with an `AllWorkspaces bool` field for list RPCs that support cross-workspace listing (platform admin only)

### Affected Sub-clients

| Sub-client | Methods to update |
|------------|-------------------|
| SandboxInterface | Create, Get, List, Delete, AttachProvider, DetachProvider, ListProviders, WaitReady, Watch, GetLogs |
| ProviderInterface | Create, Get, List, Update, Delete, Ensure |
| ProfileInterface | List, Get, Import, Update, Lint, Delete |
| RefreshInterface | GetStatus, Configure, Rotate, Delete |
| ServiceInterface | Expose, Get, List, Delete |
| ExecInterface | Run, Stream, Interactive |
| FileInterface | Upload, Download |
| ConfigInterface | GetSandbox, Update |
| PolicyInterface | GetDraft, ApproveDraftChunk, RejectDraftChunk, ApproveAllDraftChunks, ClearDraftChunks, GetDraftHistory, GetStatus, List, EditDraftChunk, UndoDraftChunk |

## Open Questions

- Should `Health().Check()` and `Config().GetGateway()` get workspace params? They're unscoped in the proto. Probably not.
- How should the fake client handle workspace isolation? Separate in-memory stores per workspace, or a flat store with workspace as a key prefix?
- The `oshell` TUI example needs updating too. Can that be a follow-up?
