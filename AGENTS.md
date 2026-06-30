# OpenShell SDK for Go

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

## Conventions

- **Test framework**: Go testing + testify (assert/require)
- **Integration tests**: Use `//go:build integration` build tag
- **Linter**: golangci-lint v2 with govet, errcheck, staticcheck, unused, ineffassign, revive, goheader

## SPDX License Header

Every `.go` file must start with:

```go
// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
```

## Proto Generation

```bash
mise run proto:sync    # Copy proto files from upstream OpenShell repo
mise run proto:gen     # Generate Go bindings with protoc
mise run proto:check   # Verify generated files are up to date (CI)
mise run proto:clean   # Remove all generated .pb.go files
```

- Set `UPSTREAM_PATH` env var to override the default upstream path (`../OpenShell/proto/`)
- Generated `.pb.go` files are committed to the repo
- Generated files are exempt from SPDX license header checks
