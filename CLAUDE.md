<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
at specs/001-project-setup/plan.md
<!-- SPECKIT END -->

# OpenShell SDK for Go

Go SDK for interacting with OpenShell servers.

## Build Commands

All commands use `make` which delegates to mise tasks:

```bash
make test              # Run unit tests with coverage
make test-integration  # Run integration tests
make lint              # Run golangci-lint
make fmt               # Format code (goimports + go fmt)
make build             # Build all packages
make ci                # Full CI pipeline (lint + build + test)
```

## Project Conventions

- **Go version**: 1.23+ (pinned in mise.toml)
- **Module path**: `github.com/rhuss/openshell-sdk-go`
- **Test framework**: Go testing + testify (assert/require)
- **Integration tests**: Use `//go:build integration` build tag
- **License headers**: All `.go` files require SPDX headers (enforced by goheader linter)
- **Linter**: golangci-lint v2 with govet, errcheck, staticcheck, unused, ineffassign, revive, goheader

## SPDX License Header

Every `.go` file must start with:

```go
// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
```

## Project Structure

```
openshell/          # SDK package (Client, Dial, Close)
mise.toml           # Tool versions and task definitions
Makefile            # Build shim delegating to mise
.golangci.yml       # Linter configuration
.github/workflows/  # CI pipeline (lint, test, build)
```
