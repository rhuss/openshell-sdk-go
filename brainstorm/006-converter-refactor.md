# Brainstorm: Converter Code Deduplication

**Date:** 2026-06-28
**Status:** done (implemented as spec 004-converter-dedup)

## Problem Framing

The SDK has duplicated proto-to-SDK conversion logic in two places:
`internal/converter/` (exported, independently tested) and the `*_client.go`
files in `v1/` (unexported, duplicated). This exists because of a circular
import: `internal/converter/` imports `v1` for domain types, so `v1/` cannot
import `internal/converter/` back.

The duplication was flagged by Devin during the PR #1 review. It creates
maintenance risk: changes to conversion logic must be applied in two places,
and the two copies can drift.

## Approaches Considered

### A: Extract Types to Separate Package (Chosen)

Move all domain types (Sandbox, Provider, ExecResult, StatusError,
SandboxPhase, EventType, StreamType, Options structs, AuthProvider,
Logger, Config, TLSConfig, RetryPolicy) to a dedicated `openshell/types/`
package. Both `v1/` (client code) and `internal/converter/` import
`types/` instead of `v1/`.

- Pros: Clean dependency graph with no cycles. Single source of truth
  for converters. Follows client-go pattern (`k8s.io/api/` for types,
  `k8s.io/client-go/` for clients). Independent converter testing
  preserved.
- Cons: Breaking API change (import paths change for consumers). More
  packages to navigate. Requires deciding whether `v1/` re-exports
  types or consumers import `types/` directly.

### B: Delete internal/converter, Keep Everything in v1/

Remove the `internal/converter/` package entirely. The duplicated
functions in `*_client.go` become the only copy.

- Pros: Simplest change, no new packages, no API change.
- Cons: Loses independent converter testing. The `v1/` package grows
  large with mixed concerns (client logic + conversion logic). Harder
  to verify conversion correctness in isolation.

### C: Dependency Inversion via Interfaces

Define converter interfaces in `v1/`, implement in `internal/converter/`,
inject via NewClient.

- Pros: No duplication, no new packages.
- Cons: Over-engineered for simple type mapping. Adds runtime
  indirection with no benefit. Obscures what should be static functions.

## Decision

**Approach A: Extract types to a separate package.** The package
structure follows the client-go convention. The re-export question
should be resolved in the spec phase: either `v1/` re-exports all
types (no consumer import change needed) or consumers import
`types/` directly (cleaner but breaking).

## Key Requirements

- All domain types move to `openshell/types/` (or similar path)
- `v1/` package contains only client logic (Client, sub-client implementations)
- `internal/converter/` imports `types/` (not `v1/`) to break the cycle
- `v1/` imports both `types/` and `internal/converter/`
- Duplicated converter functions in `*_client.go` are deleted
- All existing tests continue to pass
- Consider re-exporting types from `v1/` to minimize consumer impact

## Open Questions

- Should the types package be `openshell/types/` or `openshell/v1/types/`?
  The former is version-agnostic (types shared across v1/v2), the latter
  is scoped to v1.
- Should `v1/` re-export types via type aliases (`type Sandbox = types.Sandbox`)
  to avoid breaking existing consumers?
- Should the `internal/converter/` tests import `types/` directly or
  continue to use `v1`?
