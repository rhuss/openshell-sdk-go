# Implementation Plan: Project Setup and Build Infrastructure

**Branch**: `001-project-setup` | **Date**: 2026-06-27 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/001-project-setup/spec.md`

## Summary

Bootstrap the openshell-sdk-go repository with Go module initialization,
mise-based build tooling (with Makefile shim), golangci-lint configuration
with SPDX header enforcement, GitHub Actions CI, Apache-2.0 licensing,
a minimal stub package for CI validation, project documentation, and a
filled-in project constitution.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: testify (assertions), golangci-lint (linting)
**Storage**: N/A
**Testing**: Go testing package + testify, golangci-lint, integration via build tags
**Target Platform**: Cross-platform (Linux, macOS, Windows)
**Project Type**: Library (Go SDK)
**Performance Goals**: N/A (scaffolding only)
**Constraints**: Minimal dependencies, Apache-2.0 license compliance
**Scale/Scope**: ~15 files created, no runtime code

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| Proto Isolation | N/A | No proto code in this phase |
| Idiomatic Go | PASS | Standard Go project layout, go.mod, testing conventions |
| Test-First | PASS | Stub test written alongside stub code, linter validates |
| Upstream Tracking | N/A | No upstream proto dependency in this phase |
| Minimal Dependencies | PASS | Only testify as test dependency, golangci-lint as dev tool |

No violations. All applicable principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/001-project-setup/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit-tasks)
```

### Source Code (repository root)

```text
.
├── go.mod                          # Go module definition
├── go.sum                          # Dependency checksums
├── mise.toml                       # Tool versions + task definitions
├── Makefile                        # Thin shim delegating to mise
├── .golangci.yml                   # Linter configuration
├── LICENSE                         # Apache-2.0
├── README.md                       # Project overview and quick start
├── CONTRIBUTING.md                 # Contributor guide
├── CLAUDE.md                       # Agentic development instructions
├── .gitignore                      # Go artifacts, IDE files, coverage
├── .github/
│   └── workflows/
│       └── ci.yml                  # GitHub Actions: lint, test, build
├── .specify/
│   └── memory/
│       └── constitution.md         # Project constitution (5 principles)
└── openshell/
    ├── client.go                   # Stub: Client, Dial(), Close()
    └── client_test.go              # Stub: TestDialEmptyAddress
```

**Structure Decision**: Flat Go package at `openshell/` (will expand to
`openshell/v1/` with sub-packages in Phase 1). No `cmd/`, `internal/`, or
`pkg/` directories needed for scaffolding.

## File Creation Order

Files should be created in dependency order:

1. **Foundation**: `go.mod`, `.gitignore`, `LICENSE`
2. **Build tooling**: `mise.toml`, `Makefile`, `.golangci.yml`
3. **Stub code**: `openshell/client.go`, `openshell/client_test.go`, `go.sum`
4. **CI**: `.github/workflows/ci.yml`
5. **Documentation**: `README.md`, `CONTRIBUTING.md`, `CLAUDE.md`
6. **Governance**: `.specify/memory/constitution.md`

## Key Implementation Details

### mise.toml Structure

```toml
[tools]
go = "1.23"
"go:github.com/golangci/golangci-lint/cmd/golangci-lint" = "1.64"

[tasks.test]
description = "Run unit tests with coverage"
run = "go test -coverprofile=coverage.out -race ./..."

[tasks."test:integration"]
description = "Run integration tests"
run = "go test -tags=integration -race ./..."

[tasks.lint]
description = "Run linter"
run = "golangci-lint run ./..."

[tasks.fmt]
description = "Format code"
run = "goimports -w . && go fmt ./..."

[tasks.build]
description = "Build all packages"
run = "go build ./..."

[tasks.ci]
description = "Run full CI pipeline"
depends = ["lint", "build", "test"]
```

### Makefile Pattern

Each target checks for mise and delegates:

```makefile
MISE := $(shell command -v mise 2>/dev/null)

.PHONY: test lint build ci fmt

test:
ifndef MISE
	$(error mise is not installed. Install from https://mise.jdx.dev)
endif
	mise run test
```

### golangci-lint SPDX Header Config

```yaml
linters-settings:
  goheader:
    template: |-
      SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
      SPDX-License-Identifier: Apache-2.0
```

### Stub Package

```go
// openshell/client.go
package openshell

import "errors"

type Client struct{}

func Dial(address string) (*Client, error) {
    if address == "" {
        return nil, errors.New("address must not be empty")
    }
    return &Client{}, nil
}

func (c *Client) Close() error { return nil }
```

### CI Workflow

- Uses `mise` to install Go and tools
- Three jobs: lint, test, build
- Triggers on PR and push to main
- Go version extracted from mise.toml for consistency

## Complexity Tracking

No constitution violations. No complexity justifications needed.
