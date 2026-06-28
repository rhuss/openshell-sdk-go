# Implementation Plan: Converter Code Deduplication

**Branch**: `004-converter-dedup` | **Date**: 2026-06-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/004-converter-dedup/spec.md`

## Summary

Extract all domain types from `openshell/v1/` into a new `openshell/v1/types/` package to break the circular import between `v1/` and `v1/internal/converter/`. Once the cycle is broken, the `*_client.go` files can import the converter package directly, eliminating 13 duplicated conversion functions. Type aliases in `v1/` ensure backward compatibility.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: gRPC, protobuf (no new dependencies added)
**Storage**: N/A
**Testing**: Go testing + testify (assert/require), `make test`
**Target Platform**: Library (Go module)
**Project Type**: Library
**Performance Goals**: N/A (compile-time refactoring, no runtime changes)
**Constraints**: Zero breaking changes to public API; all existing tests pass
**Scale/Scope**: ~35 type definitions move; 13 duplicated functions removed; ~15 files modified

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Proto types stay in `proto/`. SDK types move to `v1/types/`, not mixed with proto. |
| II. Idiomatic Go | PASS | Type aliases for backward compat is idiomatic Go (used by golang.org/x/ packages). |
| III. Test-First | PASS | All existing tests must pass (FR-005/FR-006). No new features requiring new tests. |
| IV. Upstream Tracking | PASS | Proto definitions unchanged. Converter layer preserved. |
| V. Minimal Dependencies | PASS | No new dependencies. Pure internal refactoring. |
| VI. Secrets Never Leak | PASS | No changes to credential handling. |
| VII. Deep Copy at Boundaries | PASS | Copy helpers (`copyStringMap`, etc.) move to converter, preserving boundary semantics. |
| VIII. Doc Examples Compile | PASS | Type aliases ensure all doc examples compile unchanged. |

## Project Structure

### Documentation (this feature)

```text
specs/004-converter-dedup/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
openshell/
├── v1/
│   ├── types/               # NEW: domain types package
│   │   ├── doc.go           # Package documentation
│   │   ├── sandbox.go       # Sandbox, SandboxSpec, SandboxStatus, etc.
│   │   ├── provider.go      # Provider, ProviderSpec
│   │   ├── exec.go          # ExecResult, ExecChunk, ExecOptions
│   │   ├── health.go        # HealthResult
│   │   ├── errors.go        # ErrorCode, StatusError, IsStatusError
│   │   ├── options.go       # ListOptions, WaitOptions, etc.
│   │   ├── config.go        # Config, TLSConfig, RetryPolicy
│   │   ├── watch.go         # Event[T], WatchInterface[T], EventType
│   │   ├── logger.go        # Logger interface
│   │   └── auth.go          # AuthProvider interface
│   ├── internal/
│   │   ├── converter/       # MODIFIED: imports types/ instead of v1/
│   │   │   ├── sandbox.go
│   │   │   ├── provider.go
│   │   │   ├── exec.go
│   │   │   ├── errors.go
│   │   │   ├── time.go
│   │   │   └── copy.go      # NEW: deep-copy helpers moved from sandbox_client.go
│   │   └── grpc/
│   │       └── conn.go      # MODIFIED: imports types/ for TLSConfig
│   ├── sandbox.go           # MODIFIED: type aliases re-exporting from types/
│   ├── provider.go          # MODIFIED: type aliases
│   ├── exec.go              # MODIFIED: type aliases
│   ├── health.go            # MODIFIED: type aliases
│   ├── errors.go            # MODIFIED: type aliases
│   ├── options.go           # MODIFIED: type aliases
│   ├── types.go             # MODIFIED: type aliases
│   ├── watch.go             # MODIFIED: type aliases
│   ├── logger.go            # MODIFIED: type aliases
│   ├── auth.go              # MODIFIED: type aliases (keep noAuth, staticToken)
│   ├── client.go            # MODIFIED: type alias for Config, keep Client/ClientInterface
│   ├── sandbox_client.go    # MODIFIED: remove duplicated converters, import converter pkg
│   ├── provider_client.go   # MODIFIED: remove duplicated converters, import converter pkg
│   └── exec_client.go       # MODIFIED: remove duplicated converters, import converter pkg
```

**Structure Decision**: The types package mirrors the existing file organization in `v1/` (one file per domain area) for easy navigation. Client operation interfaces stay in `v1/` since they define client-specific method signatures.

## Implementation Approach

### Phase 1: Create types package

Create `openshell/v1/types/` with all domain types moved from their current locations. Each file mirrors the source file (e.g., types from `v1/sandbox.go` go to `v1/types/sandbox.go`). Include SPDX headers, package doc, and all constants/enums.

### Phase 2: Update converter package

Change all converter files to import `types` instead of `v1`. Replace `v1.Sandbox` with `types.Sandbox` etc. Move `copyStringMap`, `copyBoolPtr`, `copyStringSlice` helpers from `sandbox_client.go` to a new `converter/copy.go`. Update all converter tests.

### Phase 3: Add type aliases to v1/

Replace type definitions in `v1/*.go` files with type aliases pointing to `types/`. For structs: `type Sandbox = types.Sandbox`. For constants: `const SandboxPhaseRunning = types.SandboxPhaseRunning`. For functions: wrap or re-export. Keep client-specific types (Client, interfaces, implementations) in their original form.

### Phase 4: Wire clients to converter

Remove all 13 duplicated conversion functions from `*_client.go` files. Import the converter package and call `converter.SandboxFromProto()` instead of the local `sandboxFromProto()`. Update all call sites.

### Phase 5: Verify and clean up

Run `make ci` (lint + build + test). Verify no circular imports. Verify no duplicated conversion functions remain. Check that `go doc` renders correctly for re-exported types.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
