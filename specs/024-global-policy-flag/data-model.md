# Data Model: Global Policy Flag

## Modified Entities

### listPolicyConfig (existing, extended)

| Field | Type | New? | Description |
|-------|------|------|-------------|
| limit | uint32 | No | Max revisions to return |
| offset | uint32 | No | Pagination offset |
| global | bool | **Yes** | When true, list gateway-global revisions instead of sandbox-scoped |

### getStatusConfig (existing, extended)

| Field | Type | New? | Description |
|-------|------|------|-------------|
| version | uint32 | No | Specific policy version to query (0 = latest) |
| global | bool | **Yes** | When true, query global policy status instead of sandbox-scoped |

## New Public Symbols

| Symbol | Type | Description |
|--------|------|-------------|
| WithListGlobal | func(bool) ListPolicyOption | Functional option to enable global mode on List |
| WithStatusGlobal | func(bool) GetStatusOption | Functional option to enable global mode on GetStatus |

## Validation Rules

- When `global=true` on List: workspace is not validated (may be empty)
- When `global=true` on GetStatus: workspace and sandboxName are not validated (may be empty)
- When `global=false` (default): all existing validation rules apply unchanged
