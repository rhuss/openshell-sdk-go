# Implementation Plan: Inference Route Client

**Branch**: `023-inference-client` | **Date**: 2026-07-31 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/023-inference-client/spec.md`

## Summary

Add a new `Inference()` sub-client to the OpenShell SDK providing workspace-scoped inference route management (set, get, delete). The implementation follows the established sub-client pattern used by Workspaces, Config, and other SDK capabilities. The inference proto and generated stubs already exist; this feature adds the SDK layer: types, converters, real client, fake client, tests, and documentation.

## Technical Context

**Language/Version**: Go 1.23 (per go.mod)
**Primary Dependencies**: google.golang.org/grpc, google.golang.org/protobuf
**Storage**: N/A (gRPC client, server handles persistence)
**Testing**: `go test` with testify assertions, `//go:build integration` for integration tests
**Target Platform**: Linux/macOS (SDK library)
**Project Type**: Library (Go SDK)
**Performance Goals**: N/A (thin client wrapper over gRPC)
**Constraints**: Minimal dependencies (Constitution V), proto isolation (Constitution I)
**Scale/Scope**: 3 new methods, ~15 new files (types, converters, client, fake, tests, docs)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Generated types in `proto/inferencev1/`, SDK types in `types/`, converters bridge them |
| II. Idiomatic Go | PASS | context.Context first param, error returns, interfaces for testability |
| III. Test-First | PASS | Tests written alongside implementation per task ordering |
| IV. Upstream Tracking | PASS | Proto already synced, generated stubs exist |
| V. Minimal Dependencies | PASS | No new dependencies needed |
| VI. Secrets Never Leak | PASS | No credential fields in inference route types |
| VII. Deep Copy at Boundaries | PASS | Converters will deep-copy slices (ValidatedEndpoints) |
| VIII. Doc Examples Compile | PASS | Will add compilable examples in doc.go |
| IX. Agent-Friendly Documentation | PASS | All exported types/methods get doc comments with error codes |
| X. Proto-SDK Naming Fidelity | PASS | SDK fields mirror proto field semantics (ProviderName, ModelID, etc.) |
| XI. Fake-Real Parity | PASS | Fake mirrors real client validation (empty workspace, empty provider) |
| XII. Graceful Shutdown Order | N/A | No streaming or long-lived connections in inference client |
| XIII. Documentation Accompanies Features | PASS | README, doc.go, and examples included in scope |

## Project Structure

### Documentation (this feature)

```text
specs/023-inference-client/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── public-api.go
└── tasks.md
```

### Source Code (repository root)

```text
openshell/v1/
├── inference.go                    # InferenceInterface + type aliases
├── inference_client.go             # gRPC implementation
├── inference_client_test.go        # Unit tests for real client
├── types/
│   └── inference.go                # InferenceRouteConfig, InferenceRoute, ValidatedEndpoint
├── internal/converter/
│   ├── inference.go                # Proto-to-SDK and SDK-to-proto converters
│   └── inference_test.go           # Round-trip converter tests
├── fake/
│   ├── inference.go                # In-memory fake implementation
│   └── inference_test.go           # Fake behavior tests
├── client.go                       # Add inference field + Inference() accessor
├── doc.go                          # Add inference usage example
└── example_fake_test.go            # Add inference fake example (or new file)
```

**Structure Decision**: Follows existing single-project SDK layout. Each sub-client has a consistent set of files: interface definition, gRPC client, types, converters, fake, and tests.

## Complexity Tracking

No constitution violations. All principles pass.
