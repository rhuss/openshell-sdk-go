# Research: Global Policy Flag

## Option Naming

- **Decision**: Use `WithListGlobal(bool)` returning `ListPolicyOption` and `WithStatusGlobal(bool)` returning `GetStatusOption`.
- **Rationale**: Go doesn't support function overloading. Each option type needs its own constructor. The prefix matches the pattern: `WithLimit`, `WithOffset` for List; `WithVersion` for GetStatus.
- **Alternatives considered**: A single `WithGlobal` function (impossible without overloading), a shared option interface (over-engineered for a bool flag).

## Config Struct Extension

- **Decision**: Add a `global bool` field to both `listPolicyConfig` and `getStatusConfig` in `openshell/v1/types/policy.go`. Add corresponding `Global() bool` accessor methods.
- **Rationale**: Follows the exact same pattern as `limit`/`offset` on `listPolicyConfig` and `version` on `getStatusConfig`. The accessor is needed because the config structs are unexported.
- **Alternatives considered**: None, this is the established pattern.

## Validation Bypass

- **Decision**: When `global=true`, skip the empty-workspace validation in `List` and skip both empty-workspace and empty-name validation in `GetStatus`. The proto comments explicitly state these fields are "ignored when global is true".
- **Rationale**: The gateway ignores name/workspace when global=true. Client-side validation that rejects empty values would prevent valid global queries.
- **Alternatives considered**: Requiring callers to pass dummy values (poor UX, misleading).

## Fake Client Approach

- **Decision**: Implement real in-memory `List` and `GetStatus` in the fake, using a separate `globalRevisions` slice alongside the existing store. Other policy methods remain Unimplemented.
- **Rationale**: FR-006 requires the fake to support global mode. The current fake returns Unimplemented for all policy methods. Only List and GetStatus need real implementations; the rest can stay as stubs.
- **Alternatives considered**: Making all policy methods real (out of scope, much larger effort).
