# Brainstorm: Project Setup and Build Infrastructure

**Date:** 2026-06-27
**Status:** active
**Depends on:** [001-go-sdk](001-go-sdk.md) (overall SDK vision)

## Problem Framing

Before any SDK code can be written, the `openshell-sdk-go` repository needs
proper Go project scaffolding: module initialization, build tooling, test
infrastructure, linting, CI, licensing, and project governance. Without this
foundation, contributors (human or agentic) have no reliable way to build,
test, or validate changes.

The upstream OpenShell project uses Rust/Cargo with mise as its task runner.
This Go SDK needs its own build conventions that feel idiomatic to Go
developers while maintaining compatibility with the upstream project's
tooling philosophy.

## Approaches Considered

### A: Full Skeleton with Stub Package (Chosen)

Set up all tooling and include a minimal `openshell` package with a stub
`Client` type and a test for it. CI validates real Go code from day one.

- Pros: CI is green from first commit. `make test` works immediately.
  Contributors can clone and validate their environment.
- Cons: The stub code is throwaway (replaced by Phase 1 SDK implementation).

### B: Tooling Only, No Go Code

All build infrastructure but no Go source files. `go.mod` exists but the
`openshell/` package does not.

- Pros: No throwaway code. Clean separation.
- Cons: CI has nothing to validate. `make test` is a no-op.

### C: Full Skeleton with Proto Generation

Like A, but also set up the protobuf generation pipeline from upstream
`.proto` files.

- Pros: Validates proto toolchain early.
- Cons: Premature complexity. Proto strategy (vendored vs. submodule vs.
  separate module) is still an open question from the SDK brainstorm.

## Decision

**Approach A: Full Skeleton with Stub Package.** It gives a working project
with green CI without pulling in proto complexity that belongs in Phase 1.
The stub code is intentionally minimal and gets naturally replaced.

## Key Requirements

### Module and Ownership

- Go module path: `github.com/rhuss/openshell-sdk-go`
- Temporary ownership under `rhuss`, to be transferred to NVIDIA org later
- Module path change will happen at transfer time (Go module retract + new path)

### Build Tooling

- **mise** as the primary task runner and tool version manager
  - Pins Go version, golangci-lint, protoc (for future use)
  - Defines tasks: `test`, `test:integration`, `lint`, `fmt`, `build`, `ci`
- **Makefile** as a thin shim that delegates to mise
  - `make test` calls `mise run test`
  - `make lint` calls `mise run lint`
  - Allows contributors without mise to see what commands are available
  - Includes a check/install hint if mise is not found

### Testing Infrastructure

- Go's built-in `testing` package + `testify` for assertions
- Unit tests: `*_test.go` files alongside source, no build tags
- Integration tests: `//go:build integration` build tag, separate mise task
- `make test` runs unit tests only
- `make test-integration` runs integration tests (requires a running gateway)
- Coverage reporting via `go test -coverprofile`

### Linting

- `golangci-lint` with a `.golangci.yml` configuration
- Linters to enable: `govet`, `errcheck`, `staticcheck`, `unused`,
  `gosimple`, `ineffassign`, `revive`, `goimports`
- SPDX license header check (custom or via `go-header` linter)

### CI (GitHub Actions)

- Workflow: `.github/workflows/ci.yml`
- Triggers: PR to main, push to main
- Jobs: lint, test (unit), build
- Go version matrix: match the version pinned in mise.toml
- Integration tests: separate workflow or manual trigger (needs gateway)

### Licensing

- Apache-2.0 (matches upstream OpenShell)
- `LICENSE` file at repo root
- SPDX header on every `.go` file:
  ```
  // SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  // SPDX-License-Identifier: Apache-2.0
  ```

### Project Constitution

Five core principles:

1. **Proto Isolation** - Generated proto types never leak through the public
   API. The SDK defines its own domain types and converts internally. This
   decouples consumers from proto version changes.

2. **Idiomatic Go** - Follow Go conventions: functional options for
   configuration, `context.Context` as first parameter, error returns (not
   panics), `io.Reader`/`io.Writer` for streaming, channel-based async.

3. **Test-First** - Tests written before implementation. Unit tests for
   business logic, integration tests for gRPC interactions. No untested
   public API surface.

4. **Upstream Tracking** - The SDK tracks the OpenShell gateway API. Proto
   updates flow from upstream. Breaking API changes in the gateway produce
   deliberate SDK version bumps, not silent breakage.

5. **Minimal Dependencies** - Only depend on what's necessary: gRPC, protobuf,
   testify. No utility frameworks, no logging libraries (use `log/slog`
   from stdlib), no dependency injection containers.

### Stub Package

- `openshell/client.go`: `Client` struct, `Dial()` function returning
  `(*Client, error)`, `Close()` method
- `openshell/client_test.go`: basic test that `Dial` with no address returns
  an error
- Just enough to make `go test ./...` and `golangci-lint run` meaningful

### Project Files

- `.gitignore`: expanded for Go binaries, coverage files, IDE files
- `CLAUDE.md`: updated with build/test commands for agentic development
- `CONTRIBUTING.md`: contributor guide with prerequisites and workflow
- `README.md`: project description, quick start, status badge

## Open Questions

- Should `go generate` be used for proto compilation, or a dedicated mise
  task? (Deferred to Phase 1 / proto brainstorm)
- What minimum Go version to support? (Suggest: latest two releases, currently
  1.23 and 1.24)
- Should the SDK publish a mock server or test fixtures for consumer testing?
  (Deferred to later phase)
