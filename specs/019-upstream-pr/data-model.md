# Data Model: Upstream PR Preparation

**Date**: 2026-07-11 | **Feature**: 019-upstream-pr

This feature does not introduce new runtime entities. The "data model"
describes the file tree mapping between the development repo and the
upstream PR target.

## Entity: File Tree Mapping

### Source (development repo: `rhuss/openshell-sdk-go`)

| Source Path | Destination in PR | Transform |
|-------------|------------------|-----------|
| `openshell/` | `sdk/go/openshell/` | Module path rewrite in all `.go` files |
| `proto/*.proto` | `sdk/go/proto/` | Copy as-is |
| `proto/*v1/` | `sdk/go/proto/` | Regenerate with new module path |
| `go.mod` | `sdk/go/go.mod` | Rewrite module declaration |
| `go.sum` | `sdk/go/go.sum` | Regenerate via `go mod tidy` |
| `Makefile` | `sdk/go/Makefile` | Copy as-is |
| `mise.toml` | `sdk/go/mise.toml` | Update MODULE variable in proto tasks |
| `specs/` | `sdk/go/specs/` | Copy as-is (retention TBD by upstream) |
| `examples/oshell/` | (excluded) | Moved to `rhuss/openshell-examples` |
| `brainstorm/` | (excluded) | Internal only |
| `.specify/` | (excluded) | Internal only |
| `.claude/` | (excluded) | Internal only |
| `CLAUDE.md` | (excluded) | Internal only |
| `AGENTS.md` | (excluded) | Internal only |
| `docs/` | (excluded) | mdbook docs replaced by Fern MDX |

### New Files (created in upstream fork)

| Path | Purpose |
|------|---------|
| `docs/sdks/go/getting-started.mdx` | SDK installation and first API call |
| `docs/sdks/go/architecture.mdx` | Module structure and gRPC transport |
| `docs/sdks/go/error-handling.mdx` | Error types and gRPC status codes |
| `docs/sdks/go/authentication.mdx` | OIDC and gateway authentication |
| `tasks/go.toml` | Mise task for `go:proto` in upstream CI |

### Modified Files (existing in upstream fork)

| Path | Change |
|------|--------|
| `docs/index.yml` | Add "SDKs" section with Go subfolder |
| `.github/workflows/branch-checks.yml` | Add Go CI job |

## Entity: Module Path

| Field | Before | After |
|-------|--------|-------|
| go.mod module | `github.com/rhuss/openshell-sdk-go` | `github.com/NVIDIA/OpenShell/sdk/go` |
| Import prefix | `github.com/rhuss/openshell-sdk-go/openshell/v1` | `github.com/NVIDIA/OpenShell/sdk/go/openshell/v1` |
| Proto M_FLAGS | `$MODULE/proto/openshellv1` | Same pattern, new module |

## Entity: Examples Repository

| Field | Value |
|-------|-------|
| Name | `rhuss/openshell-examples` |
| Visibility | Public |
| go.mod module | `github.com/rhuss/openshell-examples` |
| SDK dependency | `github.com/NVIDIA/OpenShell/sdk/go` |
| Initial content | `examples/oshell/` (connection.go, demo.go, 4,649 LOC) |

## Validation Rules

- Every `.go` file under `sdk/go/` must contain zero references to
  `github.com/rhuss/openshell-sdk-go` after rewrite
- `sdk/go/go.mod` must declare module `github.com/NVIDIA/OpenShell/sdk/go`
- Proto M_FLAGS in `sdk/go/mise.toml` must reference the new module path
- No files from the exclusion list appear in the PR diff
- All unit tests pass under the new module identity
