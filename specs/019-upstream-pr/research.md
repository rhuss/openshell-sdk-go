# Research: Upstream PR Preparation

**Date**: 2026-07-11 | **Feature**: 019-upstream-pr

## R1: Upstream Repository Structure

**Decision**: Place SDK under `sdk/go/` as a self-contained Go module.

**Rationale**: The upstream repo has no `sdk/` directory yet. Python code
lives at `python/` (top-level). Creating `sdk/go/` establishes a
namespace for future language SDKs and separates the Go SDK cleanly from
the Rust crates (`crates/`), Python (`python/`), and other top-level
directories. The Go SDK is the first under `sdk/`, setting the pattern.

**Alternatives considered**:
- `go/` (top-level): Mirrors `python/` but Go is a reserved keyword in
  some contexts and `go/` is ambiguous. Rejected.
- `sdks/go/`: Plural `sdks` is unconventional. Rejected.
- `python/` pattern (flat `go/`): Inconsistent with the issue description
  which proposes `sdk/go/`. Rejected.

## R2: Module Path Rewrite Mechanics

**Decision**: Use `sed` for mechanical find-and-replace of the module path
in go.mod and all `.go` files, followed by `go mod tidy`.

**Rationale**: The rewrite is a simple string replacement:
`github.com/rhuss/openshell-sdk-go` to `github.com/NVIDIA/OpenShell/sdk/go`.
No `replace` directives exist in go.mod. No vendor directory exists.
The `go.sum` regenerates from `go mod tidy`. `sed` is sufficient and
avoids adding tooling dependencies.

**Alternatives considered**:
- `gomvpkg`: Overkill for a module-level rename with no package splits.
- `gofmt -r`: Cannot rewrite import paths.
- Manual editing: Error-prone with 192 Go files.

**Implementation detail**: Proto generation scripts (`mise.toml`) also
contain the module path in M_FLAGS. These must be updated simultaneously.

## R3: Fern Documentation Integration

**Decision**: Create a new `SDKs` navigation section in `docs/index.yml`
with a `go/` subfolder containing 4 MDX pages.

**Rationale**: The upstream docs have no SDK section yet (no Python SDK
docs either despite `python/` existing). The Go SDK docs will be the
first SDK documentation. The navigation structure uses `folder` entries
pointing to directories with `index.yml` files, or inline `page` entries.
A new top-level section "SDKs" with a Go subfolder follows the existing
pattern (e.g., `providers/`, `reference/`).

**Fern version**: 5.40.0 (supports MDX, tabs, callouts, code blocks).

**Docs layout**:
```
docs/sdks/go/
├── getting-started.mdx    # SDK installation, basic connection, first API call
├── architecture.mdx       # gRPC transport, module structure, proto layer
├── error-handling.mdx     # Error types, gRPC status codes, retry patterns
└── authentication.mdx     # OIDC, token refresh, gateway auth
```

Navigation entry in `docs/index.yml`:
```yaml
- section: "SDKs"
  slug: sdks
  contents:
  - folder: sdks/go
    title: "Go SDK"
```

## R4: CI Job Integration

**Decision**: Add a `go` job to `.github/workflows/branch-checks.yml`
following the existing Rust/Python job pattern.

**Rationale**: The branch-checks workflow already runs Rust and Python
checks in the same CI container (`ghcr.io/nvidia/openshell/ci:latest`).
Adding a Go job follows the same pattern: checkout, install tools via
mise, run lint/build/test/proto-check. The CI container needs Go and
protoc, both installable via mise.

**Job structure** (mirrors Rust/Python jobs):
```yaml
go:
  name: Go
  needs: pr_metadata
  if: needs.pr_metadata.outputs.should_run == 'true'
  runs-on: linux-amd64-cpu8
  container:
    image: ghcr.io/nvidia/openshell/ci:latest
  steps:
    - checkout
    - install tools (mise install --locked)
    - lint (cd sdk/go && mise run lint)
    - build (cd sdk/go && mise run build)
    - test (cd sdk/go && mise run test)
    - proto check (cd sdk/go && mise run proto:check)
```

**Proto task adaptation**: The current `proto:gen` and `proto:check` tasks
in `mise.toml` use `MODULE="github.com/rhuss/openshell-sdk-go"`. After
rewrite, this becomes `MODULE="github.com/NVIDIA/OpenShell/sdk/go"`.
The `tasks/go.toml` file provides two tasks for the monorepo CI:
- `go:proto` wraps `cd sdk/go && mise run proto:gen` (regeneration)
- `go:proto:check` wraps `cd sdk/go && mise run proto:check` (freshness validation)
The CI job calls `go:proto:check`; contributors run `go:proto` to regenerate.

## R5: Examples Repository

**Decision**: Create `github.com/rhuss/openshell-examples` as a public
repository with the oshell TUI extracted.

**Rationale**: The examples repo must be public for `go get` to resolve
it. It gets its own `go.mod` depending on
`github.com/NVIDIA/OpenShell/sdk/go`. Since the upstream SDK won't be
published until the PR is merged, the examples `go.mod` will initially
use a `replace` directive pointing to the local SDK during development,
switched to the real upstream path once published.

**Bootstrap**:
```bash
gh repo create rhuss/openshell-examples --public --description "Examples for the OpenShell Go SDK"
```

## R6: File Exclusion Strategy

**Decision**: Use a curated file list for the PR rather than copying
everything and excluding.

**Rationale**: The SDK repo contains many internal-only artifacts.
Instead of rsync with exclusions (fragile), build the PR branch by
selectively copying the needed directories:
- `openshell/` (SDK source)
- `proto/` (proto files + generated code)
- `specs/` (design documentation, retention TBD)
- `go.mod`, `go.sum`, `Makefile`, `mise.toml` (build configuration)

Explicitly excluded (never copied):
- `brainstorm/` (internal ideation)
- `.specify/` (spec-kit tooling)
- `.claude/` (Claude Code config)
- `CLAUDE.md`, `AGENTS.md` (agent instructions)
- `docs/` (mdbook, replaced by Fern docs in upstream)
- `examples/` (extracted to separate repo)
- `.github/` (development repo CI, not upstream CI)

## R7: Commit Message Structure

**Decision**: Single squashed commit with structured message following
upstream's DCO convention.

**Rationale**: Upstream requires DCO sign-off (`Signed-off-by:` trailer).
The commit message should describe the full SDK contribution scope.

**Template**:
```
feat(sdk): add Go SDK for OpenShell

Add a Go SDK that provides typed clients for the OpenShell gateway
and edge APIs. The SDK wraps the gRPC transport layer and exposes
idiomatic Go types with functional option configuration.

Includes:
- Gateway and Edge API clients with full CRUD operations
- In-memory fakes for testing (matching real client validation)
- OIDC authentication with automatic token refresh
- Proto generation automation (mise task + CI check)
- Fern documentation (getting-started, architecture, auth, errors)
- Design specifications from spec-driven development

Resolves: #2044
Signed-off-by: Roland Huß <roland@jolokia.org>
```
