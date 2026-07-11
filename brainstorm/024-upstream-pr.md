# Brainstorm: Upstream PR Preparation

**Date:** 2026-07-11
**Status:** active
**Upstream Issue:** [NVIDIA/OpenShell#2044](https://github.com/NVIDIA/OpenShell/issues/2044)

## Problem Framing

The Go SDK has matured through 18 spec-driven iterations covering all 55
RPCs, auth (token refresh, OIDC, edge proxy), gateway config, SSH/TCP
tunneling, fake client, and comprehensive tests. It currently lives at
`github.com/rhuss/openshell-sdk-go` as a standalone repo. The goal is to
contribute it upstream to NVIDIA/OpenShell as a draft PR under `sdk/go/`,
matching the pattern where the Python SDK already lives in `python/`.

This requires several coordinated changes: module path migration, example
extraction, documentation integration with the Fern-based docs site, CI
automation for proto freshness, and scoping development artifacts so they
don't clutter the upstream repo.

## Approaches Considered

### A: Fork-first migration (chosen)

Prepare everything in the `rhuss/OpenShell` fork, then open a single draft
PR against `NVIDIA/OpenShell`:

1. Create `github.com/rhuss/openshell-examples`, move `examples/oshell/`
   TUI there with its own `go.mod`
2. In the OpenShell fork, create branch `sdk/go-sdk`, add `sdk/go/` with
   the full SDK (rewritten module path, all sub-clients, auth packages,
   fake client, tests, committed `.pb.go` files)
3. Include `specs/` as design documentation with a PR note asking upstream
   whether to keep or remove them
4. Write Fern MDX docs under `docs/sdks/go/` and wire into navigation
5. Add `go:proto` mise task + CI validation step
6. Open as draft PR referencing issue #2044

- Pros: single coherent PR, reviewers see the full picture, docs and CI
  show production readiness
- Cons: large PR, import path rewrite touches every file

### B: Minimal SDK first, docs and CI follow

Two sequential PRs: first the code, then docs and CI after feedback.

- Pros: smaller first PR
- Cons: incomplete picture, reviewers will ask about missing docs/CI

### C: RFC-first, code later

Open an issue proposing the SDK, wait for approval, then submit code.

- Pros: gets buy-in first
- Cons: slower, issue #2044 already exists as the proposal, code speaks
  louder

## Decision

Approach A: fork-first migration as a single draft PR.

The SDK is mature enough to present as a complete package. The draft
status gives reviewers space without merge pressure. Issue #2044 already
serves as the RFC, so a separate proposal step would be redundant.

## Key Requirements

### Module path migration
- Change `github.com/rhuss/openshell-sdk-go` to
  `github.com/NVIDIA/OpenShell/sdk/go` in go.mod and all import paths
- Consumer imports become
  `github.com/NVIDIA/OpenShell/sdk/go/openshell/v1`

### Example extraction
- Move `examples/oshell/` (4,649 LOC, 12 files) to
  `github.com/rhuss/openshell-examples` as a separate repo
- Own `go.mod` referencing the SDK via the new module path

### Documentation (Fern MDX)
- Write Go SDK docs as Fern MDX pages under `docs/sdks/go/`
  (getting-started, architecture, error-handling, authentication)
- Wire into `docs/index.yml` navigation alongside Python SDK docs
- Remove the local mdbook docs (`docs/`) from the PR

### Proto change detection
- Add `go:proto` mise task in `tasks/go.toml` that regenerates Go
  bindings from `proto/` using protoc
- Add CI step in `branch-checks.yml` that runs the task and diffs,
  failing if generated files are stale
- Mirrors the Python SDK pattern (`python:proto` mise task)
- Commit generated `.pb.go` files so consumers can `go get` without
  needing protoc

### Spec artifacts
- Include `specs/` directory in the PR as design documentation
- PR description mentions spec-driven development methodology and asks
  upstream whether specs should be retained or removed
- Exclude `brainstorm/`, `.specify/`, `.claude/`, `CLAUDE.md`, `AGENTS.md`

### PR format
- Draft PR against `NVIDIA/OpenShell` main branch
- References issue [#2044](https://github.com/NVIDIA/OpenShell/issues/2044)
- Full SDK shipped at once (all sub-clients, auth, OIDC, edge, tunnels,
  fake client, tests)
- Mentions that the SDK was developed using spec-driven development
  with the spec-kit toolchain

## Open Questions

- Should `sdk/go/` have its own CODEOWNERS entry for Go-specific review?
- Should the Go SDK CI steps run only when `sdk/go/` or `proto/` files
  change (path filter), or on every PR?
- Does the Fern docs build need Go SDK pages validated separately?
- Should the Go SDK have its own release tagging scheme (e.g.,
  `sdk/go/v0.1.0`) or follow the main repo's releases?
- How should `option go_package` be added to the upstream proto files
  without breaking existing consumers?
