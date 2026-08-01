# Implementation Plan: Global Policy Flag

**Branch**: `024-global-policy-flag` | **Date**: 2026-08-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/024-global-policy-flag/spec.md`

## Summary

Add `WithListGlobal` and `WithStatusGlobal` functional options to the policy client's `List` and `GetStatus` methods, enabling gateway-global policy queries. Extends existing option config structs with a `global` bool field and wires it into the proto request. Updates the fake client to support global mode for these two methods. Small, backward-compatible change touching ~6 files.

## Technical Context

**Language/Version**: Go 1.23 (per go.mod)
**Primary Dependencies**: google.golang.org/grpc, google.golang.org/protobuf
**Storage**: N/A (gRPC client)
**Testing**: `go test` with testify, `make ci`
**Target Platform**: Linux/macOS (SDK library)
**Project Type**: Library
**Performance Goals**: N/A (thin client wrapper)
**Constraints**: Backward compatible, no new dependencies
**Scale/Scope**: ~6 files modified, ~200 lines added

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Global flag wired in client, not exposed as proto type |
| II. Idiomatic Go | PASS | Functional options pattern, existing convention |
| III. Test-First | PASS | Tests written alongside implementation |
| V. Minimal Dependencies | PASS | No new dependencies |
| VII. Deep Copy at Boundaries | N/A | No new slice/map types crossing boundary |
| IX. Agent-Friendly Docs | PASS | Doc comments on WithListGlobal, WithStatusGlobal |
| X. Proto-SDK Naming Fidelity | PASS | `global` field maps to `Global` option |
| XI. Fake-Real Parity | PASS | Fake List/GetStatus mirror real validation |
| XIII. Documentation Accompanies | PASS | README and doc.go updated |

## Project Structure

### Documentation (this feature)

```text
specs/024-global-policy-flag/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
└── tasks.md
```

### Source Code (modified files)

```text
openshell/v1/
├── types/policy.go              # Add global field to listPolicyConfig, getStatusConfig
├── policy.go                    # Re-export WithListGlobal, WithStatusGlobal
├── policy_client.go             # Wire global into proto requests, adjust validation
├── policy_client_test.go        # Tests for global mode
├── fake/
│   ├── policy.go                # Implement List/GetStatus with global support
│   └── policy_test.go           # Fake tests for global mode
├── doc.go                       # Add global policy usage example
README.md (project root)             # Add global policy to feature list
```

**Structure Decision**: Modifies existing files only. No new files created except test files if needed.

## Complexity Tracking

No constitution violations. All principles pass.
