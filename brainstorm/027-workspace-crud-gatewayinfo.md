# Brainstorm: Workspace CRUD + GatewayInfo + GetCurrentUser

**Date:** 2026-07-31
**Status:** active

## Problem Framing

The SDK has no way to manage workspaces or members, and no way to fetch gateway metadata or the authenticated user's identity. These are tracked as separate GitHub issues but are closely related: they all add new sub-clients/methods for RPCs that already exist in the upstream proto.

### Missing RPCs

**Workspace CRUD** (Issue [#32](https://github.com/rhuss/openshell-sdk-go/issues/32)):
- `CreateWorkspace(name, labels)` - platform admin
- `GetWorkspace(name)` - workspace member
- `ListWorkspaces(limit, offset, labelSelector)` - any user (filtered by membership)
- `DeleteWorkspace(name)` - platform admin
- `AddWorkspaceMember(workspace, principalSubject, role)` - workspace admin
- `RemoveWorkspaceMember(workspace, principalSubject)` - workspace admin
- `ListWorkspaceMembers(workspace, limit, offset)` - workspace member

**GatewayInfo** (Issue [#34](https://github.com/rhuss/openshell-sdk-go/issues/34)):
- `GetGatewayInfo()` - returns status, version, compute drivers (platform admin only)

**GetCurrentUser** (no issue):
- `GetCurrentUser()` - returns authenticated caller identity (any bearer token holder)

## Approaches Considered

### A: New Workspaces() sub-client + extend Health() sub-client

Add `client.Workspaces()` returning a `WorkspaceInterface` for all 7 workspace RPCs. Add `GetGatewayInfo()` and `GetCurrentUser()` to the existing `Health()` sub-client (rename to `Gateway()` or keep as `Health()`).

- Pros: follows existing pattern, workspace ops are naturally grouped
- Cons: `Health()` sub-client name doesn't fit for GatewayInfo/GetCurrentUser

### B: New Workspaces() + new Gateway() sub-client

Add `client.Workspaces()` for workspace/member CRUD. Add `client.Gateway()` for Health, GatewayInfo, and GetCurrentUser (deprecate `Health()`).

- Pros: clean naming, each sub-client has a coherent scope
- Cons: breaking change on `Health()` accessor

### C: New Workspaces() + top-level methods

Add `client.Workspaces()` for workspace CRUD. Add `client.GetGatewayInfo()` and `client.GetCurrentUser()` as top-level methods on the Client struct (like `client.Close()`).

- Pros: simple, GatewayInfo/GetCurrentUser are inherently client-scoped not resource-scoped
- Cons: breaks the sub-client pattern, mixes accessor patterns

## Decision

**Approach A with a rename: new `Workspaces()` sub-client, extend `Health()` to include GatewayInfo and GetCurrentUser.** The `HealthInterface` name still works since it's about gateway-level queries (health, info, identity). If it gets too crowded later, refactor to `Gateway()`. This avoids breaking changes while adding the needed functionality.

Alternatively, if renaming to `Gateway()` is preferred (since it's pre-1.0), that's fine too, but keep `Health()` as a deprecated alias for one release.

## Key Requirements

### WorkspaceInterface

```go
type WorkspaceInterface interface {
    Create(ctx context.Context, name string, labels map[string]string) (*Workspace, error)
    Get(ctx context.Context, name string) (*Workspace, error)
    List(ctx context.Context, opts ...ListOptions) ([]*Workspace, error)
    Delete(ctx context.Context, name string) error
    AddMember(ctx context.Context, workspace, principalSubject string, role WorkspaceRole) (*WorkspaceMember, error)
    RemoveMember(ctx context.Context, workspace, principalSubject string) error
    ListMembers(ctx context.Context, workspace string, opts ...ListOptions) ([]*WorkspaceMember, error)
}
```

### New types needed
- `Workspace` (from `datamodel.v1.Workspace`)
- `WorkspaceMember` (from `openshell.v1.WorkspaceMember`)
- `WorkspaceRole` enum (Admin, User)
- `GatewayInfo` (status, version, compute drivers)
- `ComputeDriver` (from `openshell.v1.ComputeDriver`)
- `CurrentUser` (subject, roles, provider)

### Fake client additions
- `fake.Client` must implement `Workspaces()` with in-memory workspace + member store
- GatewayInfo/GetCurrentUser fakes with configurable responses

## Open Questions

- `WorkspaceRole` as a Go type: use `string` constants (like SandboxPhase) or a proper `type WorkspaceRole int` enum? The proto uses `enum WorkspaceRole { WORKSPACE_ROLE_UNSPECIFIED, WORKSPACE_ROLE_USER, WORKSPACE_ROLE_ADMIN }`.
- Member role change: the gateway has no UpdateMember RPC. Document that role change = RemoveMember + AddMember in the SDK docs.
- Should `ListWorkspaces` expose an `AllWorkspaces` option? The gateway returns only workspaces the user is a member of (platform admins see all). This is server-side filtering, not a client option.
