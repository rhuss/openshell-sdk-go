# Research: Project Setup and Build Infrastructure

## Go Module Path

**Decision**: `github.com/rhuss/openshell-sdk-go`
**Rationale**: Temporary ownership under user account. Will transfer to
NVIDIA org later with Go module retract + new path.
**Alternatives considered**: `github.com/NVIDIA/openshell-sdk-go` (deferred
until org transfer).

## Go Version

**Decision**: Go 1.23 as minimum version
**Rationale**: Latest stable release with full `log/slog` support, improved
generic type inference, and range-over-func. Supports the "latest two
releases" policy (1.23 and 1.24).
**Alternatives considered**: Go 1.22 (wider compatibility but lacks some
stdlib improvements used by the SDK).

## Build Tooling: mise

**Decision**: mise as primary task runner with Makefile shim
**Rationale**: mise pins tool versions (Go, golangci-lint, protoc) ensuring
reproducible builds. Makefile provides familiar entry point for Go developers
without mise. Upstream OpenShell uses mise, maintaining tooling consistency.
**Alternatives considered**: Makefile only (no version pinning), mise only
(unfamiliar to Go ecosystem).

## Linter Configuration

**Decision**: golangci-lint with go-header for SPDX enforcement
**Rationale**: golangci-lint is the standard Go meta-linter. go-header linter
can enforce SPDX headers via template matching, catching missing headers in
both local development and CI.
**Alternatives considered**: Custom shell script for header checking (fragile),
addlicense tool (separate step, not integrated with lint pipeline).

## Test Framework

**Decision**: Go testing + testify for assertions
**Rationale**: testify is the most widely used assertion library in Go. Its
`require` and `assert` packages provide readable test code without adding
framework overhead. Compatible with standard `go test`.
**Alternatives considered**: stdlib only (verbose assertion code), gomega
(BDD style, heavier dependency).

## Integration Test Separation

**Decision**: `//go:build integration` build tag
**Rationale**: Standard Go pattern for separating slow/external tests.
`go test ./...` runs only unit tests by default. Integration tests require
explicit `-tags=integration` flag.
**Alternatives considered**: Separate `_integration_test.go` suffix
convention (not enforceable by Go toolchain), separate test directory
(breaks Go package test conventions).

## CI Platform

**Decision**: GitHub Actions
**Rationale**: Repository will be hosted on GitHub. Native integration,
free for open source, widely understood workflow syntax.
**Alternatives considered**: None (GitHub is the target platform).

## SPDX Header Format

**Decision**: Two-line header matching upstream OpenShell convention
**Rationale**: Consistent with NVIDIA open source standards and upstream
OpenShell project. go-header linter validates exact template match.

```
// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0
```
