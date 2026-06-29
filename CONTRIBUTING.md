# Contributing to OpenShell SDK for Go

## Prerequisites

- **Go 1.23+**: Install from https://go.dev/dl/ or via mise
- **mise**: Install from https://mise.jdx.dev (manages Go and tool versions)

## Setup

```bash
# Clone the repository
git clone https://github.com/rhuss/openshell-sdk-go.git
cd openshell-sdk-go

# Install pinned tool versions (Go, golangci-lint)
mise install

# Verify setup
make test
make lint
```

## Build Commands

All commands use `make` which delegates to mise tasks. If mise is not
installed, `make` will print installation instructions.

| Command | Description |
|---------|-------------|
| `make test` | Run unit tests with coverage |
| `make test-integration` | Run integration tests (requires `//go:build integration` tag) |
| `make lint` | Run golangci-lint with all configured linters |
| `make fmt` | Format code with goimports and go fmt |
| `make build` | Build all packages |
| `make ci` | Run full CI pipeline (lint + build + test) |

## Spec-Driven Development

This project follows a spec-driven development (SDD) approach using [Speckit](https://github.com/speckit/speckit) and [cc-spex](https://github.com/rhuss/cc-spex) for larger features. For smaller changes (bug fixes, minor improvements, documentation), regular PRs and issues work just fine. The SDD workflow is meant for features that benefit from upfront design alignment.

**How it works:**

- Rough ideas and explorations go into `brainstorms/` or GitHub issues
- Refined specifications live in `specs/`, one directory per feature (e.g., `specs/003-core-sdk/`)
- Each spec contains a `spec.md`, `plan.md`, and `tasks.md` that document the design decisions, implementation plan, and work breakdown
- Specs should be kept up to date as the implementation evolves. The `/speckit-spex-evolve` command helps reconcile spec drift when code and spec diverge.

**Contribution workflow:**

1. Open a PR with the spec first (or update an existing spec)
2. Get alignment on the design before writing implementation code
3. Then implement against the agreed spec

The [cc-spex collaboration extension](https://github.com/rhuss/cc-spex) supports this workflow with commands for creating spec PRs, handling review feedback, and reconciling changes.

## Development Workflow

1. Fork the repository and create a feature branch
2. Start with a spec in `specs/` (see "Spec-Driven Development" above)
3. Make your changes
4. Add SPDX license headers to all new `.go` files:
   ```go
   // SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
   // SPDX-License-Identifier: Apache-2.0
   ```
5. Run `make ci` to verify lint, build, and tests pass
6. Submit a pull request

## Testing

- **Unit tests**: Use the standard Go testing package with
  [testify](https://github.com/stretchr/testify) assertions
- **Integration tests**: Use the `//go:build integration` build tag to
  separate from unit tests. Run with `make test-integration`.
- **Coverage**: `make test` produces a `coverage.out` file

## Code Style

- Follow standard Go conventions (gofmt, goimports)
- All exported types and functions must have doc comments
- Use `errors.New()` for simple errors, `fmt.Errorf()` with `%w` for wrapping

## License Headers

Every `.go` file must include SPDX license headers. The `goheader` linter
enforces this automatically during `make lint`. Missing headers will cause
CI to fail.

## Commit Conventions

- Use clear, descriptive commit messages
- Sign off commits with DCO: `git commit -s`
