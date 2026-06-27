# Feature Specification: Project Setup and Build Infrastructure

**Feature Branch**: `001-project-setup`
**Created**: 2026-06-27
**Status**: Draft
**Input**: Brainstorm 002-project-setup.md

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Clone and Build (Priority: P1)

A Go developer clones the repository and runs `make test` to validate their
environment. The project compiles, tests pass, and linting succeeds without
any manual setup beyond having Go and mise installed.

**Why this priority**: Without a working build, no other work can proceed.
This is the foundation for all future SDK development.

**Independent Test**: Clone a fresh copy, run `make test` and `make lint`,
verify both exit 0 with meaningful output.

**Acceptance Scenarios**:

1. **Given** a fresh clone of the repository, **When** the developer runs
   `make test`, **Then** Go unit tests execute and pass with coverage output.
2. **Given** a fresh clone of the repository, **When** the developer runs
   `make lint`, **Then** golangci-lint runs and reports no issues.
3. **Given** a fresh clone without mise installed, **When** the developer
   runs `make test`, **Then** a helpful message explains how to install mise.

---

### User Story 2 - CI Validates PRs (Priority: P1)

When a contributor opens a pull request, GitHub Actions automatically runs
lint, test, and build jobs. The PR gets a green or red status check.

**Why this priority**: CI enforcement prevents broken code from merging.
Essential for project quality from day one.

**Independent Test**: Push a branch, open a PR, verify that CI runs lint,
test, and build jobs and reports status.

**Acceptance Scenarios**:

1. **Given** a PR with valid Go code, **When** CI runs, **Then** lint, test,
   and build jobs all pass and report green status.
2. **Given** a PR with a linting violation, **When** CI runs, **Then** the
   lint job fails and reports the specific violation.
3. **Given** a push to the main branch, **When** CI triggers, **Then** the
   same lint, test, and build jobs run.

---

### User Story 3 - Agentic Development (Priority: P2)

An AI coding agent reads CLAUDE.md and immediately knows how to build, test,
lint, and contribute to the project without additional context.

**Why this priority**: The project is developed with agentic workflows.
Clear machine-readable instructions accelerate development.

**Independent Test**: A new Claude Code session can read CLAUDE.md and
successfully run build, test, and lint commands.

**Acceptance Scenarios**:

1. **Given** CLAUDE.md exists with build commands, **When** an agent reads
   it, **Then** it finds working commands for test, lint, build, and CI.
2. **Given** a contributor reads CONTRIBUTING.md, **When** they follow the
   setup steps, **Then** they can build and test the project successfully.

---

### User Story 4 - License Compliance (Priority: P2)

Every Go source file includes SPDX license headers. The project includes
an Apache-2.0 LICENSE file. The linter enforces header presence.

**Why this priority**: License compliance is required for NVIDIA open source
projects and must be established before any code is written.

**Independent Test**: Run `make lint` on a file without SPDX headers and
verify the linter reports a violation.

**Acceptance Scenarios**:

1. **Given** a Go source file without SPDX headers, **When** golangci-lint
   runs, **Then** it reports a license header violation.
2. **Given** all Go source files have SPDX headers, **When** golangci-lint
   runs, **Then** no license header violations are reported.

---

### User Story 5 - Project Constitution (Priority: P3)

The project constitution documents the five governing principles (proto
isolation, idiomatic Go, test-first, upstream tracking, minimal dependencies)
so all contributors and agents align on design decisions.

**Why this priority**: The constitution guides all future SDK design decisions
but does not block immediate development.

**Independent Test**: Read the constitution file and verify all five
principles are documented with clear descriptions.

**Acceptance Scenarios**:

1. **Given** the constitution file exists, **When** a contributor reads it,
   **Then** they find all five principles with actionable descriptions.

---

### Edge Cases

- What happens when mise is not installed? The Makefile should detect this
  and print an installation hint instead of failing with a cryptic error.
- What happens when the Go version does not match mise.toml? The CI workflow
  should use the same Go version pinned in mise.toml.
- What happens when a contributor adds a Go file without SPDX headers? The
  linter should catch this in both local runs and CI.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Project MUST have a valid `go.mod` with module path
  `github.com/rhuss/openshell-sdk-go` and Go version >= 1.23.
- **FR-002**: Project MUST have a `mise.toml` that pins Go version,
  golangci-lint version, and defines tasks: `test`, `test:integration`,
  `lint`, `fmt`, `build`, `ci`.
- **FR-003**: Project MUST have a `Makefile` that delegates to mise tasks.
  Each Makefile target MUST check for mise availability and print an
  installation hint if missing.
- **FR-004**: Project MUST have a `.golangci.yml` enabling at minimum:
  govet, errcheck, staticcheck, unused, gosimple, ineffassign, revive,
  goimports, and go-header (for SPDX enforcement).
- **FR-005**: Project MUST have a `.github/workflows/ci.yml` that runs
  lint, test, and build on PRs and pushes to main.
- **FR-006**: Project MUST have an Apache-2.0 `LICENSE` file at the
  repository root.
- **FR-007**: Every `.go` file MUST include SPDX license headers:
  `SPDX-FileCopyrightText` and `SPDX-License-Identifier: Apache-2.0`.
- **FR-008**: Project MUST have a stub `openshell/client.go` with a `Client`
  struct, `Dial()` function returning `(*Client, error)`, and `Close()`
  method.
- **FR-009**: Project MUST have a stub `openshell/client_test.go` with at
  least one test that validates `Dial` with an empty address returns an error.
- **FR-010**: Integration tests MUST use the `//go:build integration` build
  tag to separate them from unit tests.
- **FR-011**: The `ci` mise task MUST run lint, build, and test in sequence.
- **FR-012**: Project MUST have a `.gitignore` covering Go binaries,
  coverage files, and IDE artifacts.
- **FR-013**: Project MUST have a `CLAUDE.md` with build, test, lint,
  and CI commands documented.
- **FR-014**: Project MUST have a `CONTRIBUTING.md` with prerequisites,
  setup instructions, and contribution workflow.
- **FR-015**: Project MUST have a `README.md` with project description and
  quick start instructions.
- **FR-016**: Project MUST have a filled-in constitution at
  `.specify/memory/constitution.md` with five principles: proto isolation,
  idiomatic Go, test-first, upstream tracking, minimal dependencies.
- **FR-017**: The `test` mise task MUST produce a coverage profile via
  `go test -coverprofile`.
- **FR-018**: The `test:integration` mise task MUST run only tests tagged
  with `//go:build integration`.

### Key Entities

- **Go Module**: The root `go.mod` defining the module path and Go version.
- **Mise Configuration**: `mise.toml` defining tool versions and task
  definitions.
- **Makefile**: Thin shim delegating to mise tasks with availability checks.
- **Linter Configuration**: `.golangci.yml` defining enabled linters and
  SPDX header rules.
- **CI Workflow**: `.github/workflows/ci.yml` defining GitHub Actions jobs.
- **Stub Package**: Minimal `openshell/` package with `Client` type and test.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: `make test` completes successfully in under 30 seconds on a
  fresh clone with Go and mise installed.
- **SC-002**: `make lint` completes successfully with zero violations on
  all committed Go files.
- **SC-003**: `make ci` runs the full lint, build, and test pipeline
  successfully.
- **SC-004**: CI workflow passes on a PR containing only the scaffolding
  files.
- **SC-005**: All committed `.go` files contain valid SPDX license headers,
  verified by the go-header linter.
- **SC-006**: A contributor can go from clone to passing tests in under
  5 minutes following CONTRIBUTING.md instructions.

## Assumptions

- Contributors have Go >= 1.23 installed or can install it via mise.
- Contributors have mise installed or are willing to install it (the
  Makefile provides guidance).
- The GitHub repository will be created at `github.com/rhuss/openshell-sdk-go`
  before CI can run, but CI configuration is written ahead of time.
- The stub package is intentionally minimal and will be replaced by real
  SDK code in Phase 1 (Core SDK).
- The `testify` assertion library is acceptable as the only non-stdlib
  test dependency.
- protoc is pinned in mise.toml for future use but not required for this
  setup phase.
