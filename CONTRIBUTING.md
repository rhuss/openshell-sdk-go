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

## Development Workflow

1. Fork the repository and create a feature branch
2. Make your changes
3. Add SPDX license headers to all new `.go` files:
   ```go
   // SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
   // SPDX-License-Identifier: Apache-2.0
   ```
4. Run `make ci` to verify lint, build, and tests pass
5. Submit a pull request

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
