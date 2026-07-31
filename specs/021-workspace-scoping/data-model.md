# Data Model: Workspace Scoping

## Entities

### Workspace (string)

A workspace is a tenant-level isolation boundary. It is represented as a plain string (the workspace name), not a struct. Empty string normalizes to "default" at the gateway.

- No dedicated SDK type needed (it's just a `string` parameter)
- Validation is handled by the gateway, not the SDK

### ListOptions (modified)

Existing struct in `openshell/v1/types/options.go`:

| Field | Type | Description |
|-------|------|-------------|
| Limit | int | Max results per page |
| Offset | int | Pagination offset |
| LabelSelector | string | Label filter expression |
| **AllWorkspaces** | **bool** | **NEW: When true, list across all workspaces. Overrides the workspace parameter.** |

### ObjectMeta (existing, already has workspace)

The `ObjectMeta` type in proto (`datamodel.proto` field 7) already has a `workspace` field. The SDK's `types.ObjectMeta` should expose it as a read-only field on returned resources so callers can see which workspace a resource belongs to.

## Relationships

```
Workspace (string)
  └── scopes: Sandbox, Provider, Profile, RefreshConfig, Service, Policy
      └── Sandbox scopes: Exec, File, SSH, TCP, Config, Logs
```

All resources are scoped to exactly one workspace. Cross-workspace listing is opt-in via `AllWorkspaces` on List operations.

## State Transitions

No new state transitions. Workspace is an immutable attribute set at resource creation time.

## Fake Store Key Strategy

Current: `map[name]T`
New: `map["workspace/name"]T`

The composite key format `workspace + "/" + name` provides workspace isolation without changing the generic `objectStore[T]` type signature. The `nameFunc` parameter becomes a `keyFunc` that receives the workspace context.
