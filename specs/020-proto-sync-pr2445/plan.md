# Implementation Plan: Proto Sync from Upstream PR #2445

**Branch**: `020-proto-sync-pr2445` | **Date**: 2026-07-31 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/020-proto-sync-pr2445/spec.md`

## Summary

Sync all 5 SDK-relevant proto files from upstream OpenShell (post-PR #2445) and regenerate Go bindings. This adds `inference.proto` as a new file, updates `openshell.proto`, `options.proto`, `datamodel.proto`, and `sandbox.proto` with authorization annotations and workspace scoping fields. The `buf.gen.yaml` must be updated to include the new `inference.proto` and its import mappings. No new client code is added.

## Technical Context

**Language/Version**: Go 1.25.0
**Primary Dependencies**: buf (code gen), protoc-gen-go 1.36.11, protoc-gen-go-grpc 1.6.2
**Storage**: N/A
**Testing**: Go testing + testify (assert/require), `make ci`
**Target Platform**: Library (Go module)
**Project Type**: SDK library
**Performance Goals**: N/A (build-time only change)
**Constraints**: No new public API surface, generated files committed to repo
**Scale/Scope**: 5 proto files, ~4 generated Go packages + 1 new package

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Generated types stay in `proto/` packages, no public API changes |
| II. Idiomatic Go | PASS | No hand-written Go code added |
| III. Test-First | PASS | Existing tests must pass; no new features requiring tests |
| IV. Upstream Tracking | PASS | This IS the upstream tracking work |
| V. Minimal Dependencies | PASS | No new dependencies |
| VI. Secrets Never Leak | PASS | No credential handling |
| VII. Deep Copy at Boundaries | PASS | No new converters |
| VIII. Doc Examples Compile | PASS | No API changes |
| IX. Agent-Friendly Documentation | PASS | No new public symbols requiring docs |
| X. Proto-SDK Naming Fidelity | N/A | No new SDK domain types |
| XI. Fake-Real Parity | PASS | No fake updates needed (out of scope) |
| XII. Graceful Shutdown Order | N/A | No shutdown logic |
| XIII. Documentation Accompanies Features | PASS | No new public API surface (FR-006) |

All gates pass. No violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/020-proto-sync-pr2445/
├── spec.md
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── checklists/
    └── requirements.md
```

### Source Code (repository root)

```text
proto/
├── openshell.proto      # Updated from upstream
├── inference.proto      # NEW - copied from upstream
├── options.proto        # Updated from upstream
├── datamodel.proto      # Updated (or unchanged) from upstream
├── sandbox.proto        # Updated (or unchanged) from upstream
├── openshellv1/         # Regenerated Go bindings
│   ├── openshell.pb.go
│   └── openshell_grpc.pb.go
├── optionsv1/           # Regenerated Go bindings
│   └── options.pb.go
├── datamodelv1/         # Regenerated Go bindings
│   └── datamodel.pb.go
├── sandboxv1/           # Regenerated Go bindings
│   └── sandbox.pb.go
└── inferencev1/         # NEW - generated Go bindings
    ├── inference.pb.go
    └── inference_grpc.pb.go
```

**Structure Decision**: Follows existing `proto/` layout convention. Each proto file generates into a corresponding `proto/<name>v1/` package directory. The new `inference.proto` follows the same pattern with `inferencev1/`.

## Implementation Approach

### Step 1: Copy Proto Files from Upstream

Copy the 5 SDK-relevant proto files from the upstream OpenShell repository at `/Users/rhuss/Work/projects/OpenShell/proto/` to the SDK's `proto/` directory, overwriting existing files:

- `openshell.proto`
- `inference.proto` (new)
- `options.proto`
- `datamodel.proto`
- `sandbox.proto`

Exclude internal/operator-only protos: `compute_driver.proto`, `gateway_interceptor.proto`, `supervisor_middleware.proto`, `test.proto`.

### Step 2: Update buf.gen.yaml

Add `inference.proto` to the `inputs.paths` list and add `Minference.proto` import mappings to both plugins. The inference proto uses the package `inference`, so the Go package path will be `github.com/rhuss/openshell-sdk-go/proto/inferencev1`.

Also update `buf.yaml` if needed to include the new proto in the module.

### Step 3: Create inferencev1 Package Directory

Create `proto/inferencev1/` directory to receive the generated Go bindings.

### Step 4: Run Code Generation

Run `mise run proto:gen` (which executes `buf generate`) to regenerate all Go bindings. The existing `proto:gen` task cleans all `*.pb.go` files first, then regenerates.

### Step 5: Fix Compilation Issues

If the new proto files introduce imports or types that cause compilation issues, resolve them. Likely issues:
- `inference.proto` may import `options.proto` for auth annotations
- New field types in `openshell.proto` may reference types from `options.proto`

### Step 6: Verify CI

Run `make ci` (lint + build + test) to verify:
- All generated code compiles
- All existing tests pass
- Lint checks pass (generated files are exempt from SPDX header checks)

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Upstream proto has breaking field changes | Low | High | Proto compatibility rules; revert if needed |
| buf.gen.yaml misconfiguration | Medium | Medium | Follow existing pattern for 4 other protos |
| New imports not resolvable | Low | Medium | All imports are within the same proto directory |
| Existing tests fail due to field additions | Low | Low | Proto field additions are backward-compatible |

## Complexity Tracking

No constitution violations. No complexity justifications needed.
