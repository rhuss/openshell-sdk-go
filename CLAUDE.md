<!-- SPECKIT START -->
For additional context about technologies to be used, project structure,
shell commands, and other important information, read the current plan
at specs/007-ssh-tcp-config/plan.md
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

## Proto Generation

Proto tasks for syncing upstream `.proto` files and generating Go bindings:

```bash
mise run proto:sync    # Copy proto files from upstream OpenShell repo
mise run proto:gen     # Generate Go bindings with protoc
mise run proto:check   # Verify generated files are up to date (CI)
mise run proto:clean   # Remove all generated .pb.go files
```

- Set `UPSTREAM_PATH` env var to override the default upstream path (`../OpenShell/proto/`)
- Generated `.pb.go` files are committed to the repo
- Generated files are exempt from SPDX license header checks

## Project Structure

```
openshell/          # SDK package (Client, Dial, Close)
proto/              # Proto source and generated Go packages
  *.proto           # Upstream proto files (synced via proto:sync)
  UPSTREAM_VERSION  # Git SHA of upstream at sync time
  openshellv1/      # Generated: gRPC service + messages
  datamodelv1/      # Generated: shared data types
  sandboxv1/        # Generated: sandbox types
mise.toml           # Tool versions and task definitions
Makefile            # Build shim delegating to mise
.golangci.yml       # Linter configuration
.github/workflows/  # CI pipeline (lint, test, build, proto:check)
```
