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

## Invariants

- **Test-first**: Write a failing test before implementation. Every public function needs at least one test. `make ci` must pass before merge.
- **Proto isolation**: Generated proto types stay in `proto/` packages. The public API uses idiomatic Go types; conversion happens internally. Never expose proto types directly.
- **Deep copy at boundaries**: Slices, maps, and mutable references crossing the proto/SDK boundary must be deep-copied. A returned SDK struct must never share references with the proto message it was converted from.
- **Secrets never leak**: Tokens, credentials, and API keys must never appear in error messages, logs, or response types. Credential fields are write-only.
- **Fake-real parity**: Fake implementations must mirror the real client's input validation (nil checks, range validation, empty-string rejection). Stubs that only return Unimplemented hide production bugs.
- **Graceful shutdown order**: Protocol-level close (e.g., gRPC CloseSend) executes before context cancellation. Cancelling context first causes spurious context-cancelled errors.

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
