# Review Guide: Upstream PR Preparation

**Generated**: 2026-07-11 | **Spec**: [spec.md](spec.md)

## Why This Change

The OpenShell Go SDK has been developed in a standalone repository
(`rhuss/openshell-sdk-go`) providing typed gRPC clients for the gateway
and edge APIs, in-memory fakes for testing, and OIDC authentication.
Upstream issue [#2044](https://github.com/NVIDIA/OpenShell/issues/2044)
requests contributing this SDK into the `NVIDIA/OpenShell` monorepo under
`sdk/go/`. Without this contribution, Go developers must discover and
depend on a personal fork rather than the official project.

## What Changes

The Go SDK source is relocated into the upstream repo at `sdk/go/` with
its module path rewritten from `github.com/rhuss/openshell-sdk-go` to
`github.com/NVIDIA/OpenShell/sdk/go`. The 4,649-LOC oshell TUI example
is extracted to a separate repository (`rhuss/openshell-examples`) to
keep the PR focused on library code. Four concise Fern MDX documentation
pages are added under `docs/sdks/go/`. A Go CI job and proto freshness
check are added to the existing branch-checks workflow. The PR is
delivered as a single squashed draft commit referencing issue #2044.

## How It Works

The implementation proceeds in three phases:

1. **Module path migration**: Mechanical `sed` replacement of the module
   path in `go.mod`, all `.go` files, and the mise proto generation
   scripts. Proto bindings are regenerated under the new module. Build
   and test verification confirms zero regressions.

2. **Documentation and CI**: Fern MDX pages (getting-started, architecture,
   error-handling, authentication) are created under `docs/sdks/go/` and
   wired into the docs navigation. A `tasks/go.toml` mise task wraps
   proto generation for the monorepo context. A Go job is added to
   `branch-checks.yml` following the existing Rust/Python pattern
   (checkout, mise install, lint, build, test, proto:check).

3. **PR assembly**: Internal artifacts (brainstorms, spec-kit config,
   Claude Code config) are excluded. Design specs are included under
   `sdk/go/specs/` with an explicit question to upstream maintainers
   about whether to retain them. All changes are squashed into a single
   DCO-signed commit.

## When It Applies

**Applies when**:
- Contributing the Go SDK to the NVIDIA/OpenShell upstream repository
- Preparing a draft PR with module path rewrite, docs, and CI integration
- Establishing the `sdk/` directory pattern for future language SDKs

**Does not apply when**:
- Ongoing SDK feature development (that happens in the development repo)
- Python SDK or other language SDK contributions (separate scope)
- Changes to upstream proto definitions (the SDK adapts to upstream, not the reverse)

## Key Decisions

1. **SDK placed under `sdk/go/`, not `go/` (top-level)**. The `sdk/`
   namespace establishes a pattern for future language SDKs. Python
   already lives at `python/` but a new `sdk/` prefix provides cleaner
   organization. Alternative: flat `go/` directory. Rejected because
   `go/` is ambiguous and the issue description proposed `sdk/go/`.

2. **Module path rewrite via `sed`, not specialized tools**. The rewrite
   is a simple string replacement with no `replace` directives or
   vendor directory to complicate matters. `gomvpkg` and `gofmt -r`
   were considered but add unnecessary tooling for a one-time operation.

3. **Examples extracted to separate repo, not kept in-tree**. The
   4,649-LOC oshell TUI is application code, not library code. Including
   it would bloat the PR and blur the SDK scope. The examples repo
   depends on the upstream module path, validating end-to-end importability.

4. **Single squashed commit for the PR**. Avoids partial-state commits
   where intermediate steps (e.g., module path rewrite without updated
   imports) would not compile. Alternative: per-story commits. Rejected
   because reviewers examine the full diff anyway and intermediate
   commits add noise without independent value.

5. **Proto generation uses M_FLAGS, not proto file modification**.
   The `go_package` option is supplied via protoc command-line flags
   (`--go_opt=M...`) rather than editing upstream `.proto` files.
   This respects upstream ownership of proto definitions.

6. **Concise reference docs, not full tutorials**. Fern MDX pages are
   1-2 pages each with key concepts and short code snippets. Full
   tutorial walkthroughs can be added post-merge. This reduces PR
   review burden while demonstrating production readiness.

## Areas Needing Attention

- **Module path completeness**: The `sed` rewrite must catch every
  occurrence in `.go` files AND the mise.toml proto generation scripts.
  A missed reference will cause subtle import failures.

- **Proto M_FLAGS correctness**: The mapping between proto file names
  and Go package paths in the mise task must match the upstream proto
  directory structure. If upstream proto files are reorganized, these
  flags need updating.

- **Fern compatibility**: The Go SDK docs are the first SDK documentation
  in the upstream docs site. The "SDKs" navigation section is new and
  must integrate without breaking existing page routing.

- **CI container toolchain**: The Go job assumes the CI container
  (`ghcr.io/nvidia/openshell/ci:latest`) can install Go 1.25 and
  protoc via mise. If the container lacks prerequisites for mise-managed
  Go installation, the CI job will fail.

- **Examples repo `replace` directive**: The examples repo temporarily
  uses a `replace` directive since the upstream SDK module isn't
  published until the PR merges. This directive must be removed once
  the SDK is available at the upstream path.

## Open Questions

- Should the `specs/` directory be retained in the upstream repo?
  The PR description explicitly asks upstream maintainers to decide.
- Will the CI container support Go 1.25 installation via mise, or
  does a container image update need to be coordinated?

## Review Checklist

- [ ] Key decisions are justified
- [ ] Breaking changes are documented with migration guidance
- [ ] Scope matches the stated boundaries
- [ ] Success criteria are achievable
- [ ] No unstated assumptions
- [ ] Zero references to old module path (`github.com/rhuss/openshell-sdk-go`) in `sdk/go/`
- [ ] Fern MDX code examples use actual SDK function signatures (FR-016)
- [ ] SPDX license headers present on all `.go` files
- [ ] PR description references issue #2044 and includes spec retention question
- [ ] Internal artifacts (brainstorm, .specify, .claude) are excluded from PR

---

<!-- Code phase sections are appended below this line by the phase-manager command -->
