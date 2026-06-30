# Data Model: Typed SandboxPolicy

## New Domain Types

### SandboxPolicy

Top-level security policy configuration for a sandbox.

| Field | Type | Description |
|---|---|---|
| Version | uint32 | Policy version number. Server may override on write. |
| Filesystem | *FilesystemPolicy | Filesystem access rules. Nil = no filesystem policy. |
| Landlock | *LandlockPolicy | Landlock LSM configuration. Nil = no landlock policy. |
| Process | *ProcessPolicy | Process identity rules. Nil = no process policy. |
| NetworkPolicies | map[string]NetworkPolicyRule | Named network access rules. Nil = no network policies. |

### FilesystemPolicy

Controls which directories the sandbox can access.

| Field | Type | Description |
|---|---|---|
| IncludeWorkdir | bool | Auto-include working directory as read-write. |
| ReadOnly | []string | Read-only directory allow list. |
| ReadWrite | []string | Read-write directory allow list. |

### LandlockPolicy

Linux Landlock LSM configuration.

| Field | Type | Description |
|---|---|---|
| Compatibility | string | Compatibility mode (e.g., "best_effort", "hard_requirement"). |

### ProcessPolicy

Process identity configuration.

| Field | Type | Description |
|---|---|---|
| RunAsUser | string | User name for sandboxed processes. |
| RunAsGroup | string | Group name for sandboxed processes. |

## Modified Types

### SandboxSpec (add field)

| Field | Change | New Type |
|---|---|---|
| Policy | added | *SandboxPolicy |

### ConfigUpdate (breaking change)

| Field | Change | Old Type | New Type |
|---|---|---|---|
| Policy | type change | []byte | *SandboxPolicy |

### SandboxPolicyRevision (breaking change)

| Field | Change | Old Type | New Type |
|---|---|---|---|
| Policy | type change | []byte | *SandboxPolicy |

### SandboxConfig (breaking change)

| Field | Change | Old Type | New Type |
|---|---|---|---|
| Policy | type change | []byte | *SandboxPolicy |

## Existing Type (reused, no changes)

### NetworkPolicyRule

Already fully typed in `openshell/v1/types/network_policy.go`.
Reused by reference in `SandboxPolicy.NetworkPolicies` map values.

## Nil Semantics

| Value | Meaning |
|---|---|
| `Policy == nil` | No policy specified (create) or no policy in response (read) |
| `Filesystem == nil` | No filesystem sub-policy |
| `NetworkPolicies == nil` | No network policies |
| `NetworkPolicies == map{}` (empty) | Explicitly empty network policies (distinct from nil) |
| `ReadOnly == nil` | No read-only directories |
| `ReadOnly == []string{}` (empty) | Explicitly empty read-only list (distinct from nil) |
