# Proto Gap Fixes: Drop D Network Policy Types

**Date**: 2026-08-08
**Issues**: #35, #36, #37
**JIRA**: RHAIENG-6595
**Branch**: `fix/proto-gap-drop-d`

## Problem

Three sets of upstream proto fields have no corresponding SDK domain types or converter code. Some are incorrectly listed as "handled" in the coverage test despite having no implementation. This blocks Drop D upstream contribution.

## Scope

Single branch and PR covering all three issues. All changes follow the established 3-layer pattern: domain types, converters, coverage test, unit tests.

## Issue #37: SigV4 + JSON-RPC Fields on NetworkEndpoint

Four scalar fields added to `PolicyNetworkEndpoint`:

| Field | Go Type | Proto Source |
|-------|---------|-------------|
| `CredentialSigning` | `string` | `credential_signing` (line 166) |
| `SigningService` | `string` | `signing_service` (line 169) |
| `SigningRegion` | `string` | `signing_region` (line 172) |
| `JsonRpcMaxBodyBytes` | `uint32` | `json_rpc_max_body_bytes` (line 175) |

Converter: direct field assignment, no nil checks (scalars).

## Issue #35: MCP Policy Types

### New type: `McpOptions`

```go
type McpOptions struct {
    StrictToolNames         *bool
    AllowAllKnownMcpMethods *bool
}
```

Uses `*bool` because proto fields are `optional bool`. Converter uses `CopyBoolPtr()`.

### Modified types

- `PolicyNetworkEndpoint`: add `Mcp *McpOptions` field
- `L7Allow`: add `Params map[string]L7QueryMatcher` field
- `L7DenyRule`: add `Params map[string]L7QueryMatcher` field

Note: `L7QueryMatcher` needs to be defined if not already present. Check proto for the exact message shape during implementation.

## Issue #36: Network Middleware Types

### New type: `NetworkMiddlewareConfig`

```go
type NetworkMiddlewareConfig struct {
    Name       string
    Middleware  string
    Config     map[string]any
    OnError    string
    Endpoints  *MiddlewareEndpointSelector
    Order      int32
}
```

`Config` maps from `google.protobuf.Struct` (same pattern used elsewhere in the SDK).

### New type: `MiddlewareEndpointSelector`

```go
type MiddlewareEndpointSelector struct {
    Include []string
    Exclude []string
}
```

### New type: `SupervisorMiddlewareService`

```go
type SupervisorMiddlewareService struct {
    Name          string
    GrpcEndpoint  string
    MaxBodyBytes  uint64
    Timeout       string
}
```

### Modified type: `SandboxPolicy`

Add `NetworkMiddlewares map[string]NetworkMiddlewareConfig`. Move `network_middlewares` from `skipped` to `handled` in coverage test.

## Files Changed

| File | Changes |
|------|---------|
| `openshell/v1/types/network_policy.go` | `PolicyNetworkEndpoint` fields, `McpOptions` type |
| `openshell/v1/types/policy.go` | Middleware types, `SandboxPolicy.NetworkMiddlewares`, `Params` on L7 types |
| `openshell/v1/internal/converter/network_policy.go` | Converter for SigV4, MCP, Params fields |
| `openshell/v1/internal/converter/policy.go` | Converter for middleware types |
| `openshell/v1/internal/converter/coverage_test.go` | Move fields to handled, remove from skipped |
| `openshell/v1/internal/converter/network_policy_test.go` | Tests for endpoint/MCP conversions |
| `openshell/v1/internal/converter/policy_test.go` | Tests for middleware conversions |

## Testing Strategy

For each new type/field:
- `TestXxxFromProto`: build proto message, convert, assert each field
- `TestXxxRoundTrip`: SDK -> proto -> SDK, assert equality
- `TestXxxDeepCopy`: convert, mutate source, assert target isolation

Coverage test must pass with all new fields in `handled` set.

`mise run ci` must pass (lint, build, test, proto:check).

## Out of Scope

- Fake client changes (fakes don't need to know about these types)
- New sub-client methods (these are type/converter additions only)
- Proto file changes (we consume upstream proto as-is)
