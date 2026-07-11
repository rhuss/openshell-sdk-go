# Implementation Plan: Upstream PR Preparation

**Branch**: `019-upstream-pr` | **Date**: 2026-07-11 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/019-upstream-pr/spec.md`

## Summary

Prepare and submit a draft PR contributing the Go SDK to `NVIDIA/OpenShell`
under `sdk/go/`. This involves rewriting the Go module path, extracting
examples into a separate repo, creating Fern MDX documentation, adding a
Go proto CI job to branch-checks.yml, and opening the PR as a single
squashed commit referencing upstream issue #2044.

## Technical Context

**Language/Version**: Go 1.25.0
**Primary Dependencies**: gRPC 1.81.1, protobuf 1.36.11, testify 1.11.1, websocket 1.8.15, oauth2 0.36.0
**Storage**: N/A
**Testing**: go test + testify (assert/require), `//go:build integration` tag for integration tests
**Target Platform**: Cross-platform Go library (consumers run Go 1.25+)
**Project Type**: Library (Go SDK) + CI automation + Fern documentation
**Performance Goals**: N/A (one-time PR preparation, not a runtime feature)
**Constraints**: Single squashed commit, no upstream proto file modifications, `go_package` supplied via protoc M_FLAGS
**Scale/Scope**: ~192 Go files, 3 proto files, 4 Fern MDX pages, 1 CI job, 1 mise task file

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Proto stays in `sdk/go/proto/`, never exposed publicly |
| II. Idiomatic Go | PASS | Module path follows Go conventions (`github.com/NVIDIA/OpenShell/sdk/go`) |
| III. Test-First | N/A | No new SDK code; existing tests validate post-rewrite correctness |
| IV. Upstream Tracking | PASS | Proto sync task + CI freshness check maintain tracking |
| V. Minimal Dependencies | PASS | No new dependencies added |
| VI. Secrets Never Leak | CHECK | Must verify no credentials leak into PR diff (`.env`, tokens, config) |
| VII. Deep Copy at Boundaries | N/A | No new boundary code |
| VIII. Doc Examples Compile | PASS | FR-016 requires Fern code examples match actual SDK API |
| IX. Agent-Friendly Docs | N/A | Applies to godoc, not Fern MDX |
| X. Proto-SDK Naming | N/A | No new types |
| XI. Fake-Real Parity | N/A | No fake changes |
| XII. Graceful Shutdown | N/A | No shutdown code |
| XIII. Docs Accompany Features | PASS | Story 3 delivers Fern MDX documentation |

**Gate result**: PASS (all applicable principles satisfied)

## Project Structure

### Documentation (this feature)

```text
specs/019-upstream-pr/
├── plan.md              # This file
├── research.md          # Phase 0: upstream structure analysis
├── data-model.md        # Phase 1: file tree layout for PR
├── quickstart.md        # Phase 1: step-by-step execution guide
└── tasks.md             # Phase 2: task breakdown (created by /speckit.tasks)
```

### Source Code (repository root)

The PR creates the following structure in the `rhuss/OpenShell` fork:

```text
sdk/
└── go/
    ├── go.mod                  # Module: github.com/NVIDIA/OpenShell/sdk/go
    ├── go.sum
    ├── Makefile
    ├── mise.toml               # Go-specific tool versions + tasks (build, test, lint)
    ├── openshell/
    │   └── v1/
    │       ├── edge/           # Edge API client
    │       ├── fake/           # In-memory fakes for testing
    │       ├── gateway/        # Gateway gRPC client
    │       ├── internal/       # Internal helpers
    │       ├── oidc/           # OIDC authentication
    │       └── types/          # Domain types
    ├── proto/
    │   ├── openshell.proto     # Synced from upstream proto/
    │   ├── datamodel.proto
    │   ├── sandbox.proto
    │   ├── openshellv1/        # Generated .pb.go files
    │   ├── datamodelv1/
    │   └── sandboxv1/
    └── specs/                  # Design documentation (retention TBD by upstream)
        ├── 001-project-setup/
        ├── 002-proto-generation/
        ├── 003-core-sdk/
        └── ...

docs/
└── sdks/
    └── go/
        ├── getting-started.mdx
        ├── architecture.mdx
        ├── error-handling.mdx
        └── authentication.mdx

tasks/
└── go.toml                    # Mise task: go:proto (references sdk/go/ paths)
```

Files modified in the fork (not new):
- `docs/index.yml` - Add SDK navigation section
- `.github/workflows/branch-checks.yml` - Add Go CI job

**Structure Decision**: The SDK lives under `sdk/go/` as a self-contained Go
module. This mirrors the existing `python/` top-level directory pattern but
places the Go SDK under `sdk/` to establish the namespace for future
language-specific SDKs. Fern docs go under `docs/sdks/go/` following the
existing `docs/<section>/` folder pattern. The mise task file at
`tasks/go.toml` follows the existing `tasks/<language>.toml` pattern.

## Complexity Tracking

No constitution violations requiring justification.

## Implementation Strategy

### Phase 1: Module Path Rewrite + Example Extraction (P1 stories)

1. Create a working branch in the `rhuss/OpenShell` fork based on upstream main
2. Copy SDK source into `sdk/go/`, rewriting module path from
   `github.com/rhuss/openshell-sdk-go` to `github.com/NVIDIA/OpenShell/sdk/go`
3. Update all internal imports, go.mod, and proto generation scripts
4. Verify `go build ./...` and `go test ./...` pass under new module
5. Create `github.com/rhuss/openshell-examples` repo, move `examples/oshell/`
6. Verify examples compile against upstream module path

### Phase 2: Documentation + CI (P2 stories)

7. Create 4 concise Fern MDX pages under `docs/sdks/go/`
8. Wire into `docs/index.yml` navigation
9. Create `tasks/go.toml` with `go:proto` task (adapted from current mise.toml)
10. Add Go CI job to `.github/workflows/branch-checks.yml`

### Phase 3: PR Assembly (P3 stories)

11. Exclude internal artifacts (brainstorm, .specify, .claude, CLAUDE.md, AGENTS.md, docs/)
12. Squash into single commit with comprehensive message
13. Open draft PR referencing issue #2044 with spec retention question
