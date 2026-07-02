# OpenShell SDK for Go Constitution

## Core Principles

### I. Proto Isolation

The SDK separates generated protobuf code from hand-written Go code.
Generated types live in dedicated packages and are never exposed directly
in the public API. Users interact with idiomatic Go types; conversion
between proto and Go types happens internally. This allows the proto
schema to evolve without breaking the SDK's public surface.

### II. Idiomatic Go

The SDK follows established Go conventions: standard project layout,
`error` returns instead of exceptions, context propagation, functional
options for configuration, and interfaces for testability. Code should
feel natural to a Go developer familiar with the standard library and
packages like `net/http` or `google.golang.org/grpc`.

### III. Test-First (NON-NEGOTIABLE)

Tests are written before or alongside implementation code. Every public
function has at least one test. The Red-Green-Refactor cycle is enforced:
write a failing test, make it pass, then clean up. Integration tests use
the `//go:build integration` build tag to separate from unit tests.
`make ci` must pass before any merge.

### IV. Upstream Tracking

The SDK tracks the upstream OpenShell project's proto definitions and
API changes. When upstream protos change, the SDK updates generated code
and adapts the Go API layer. Breaking upstream changes are absorbed in
the conversion layer, not propagated to SDK users. Version compatibility
with specific OpenShell releases is documented.

### V. Minimal Dependencies

The SDK minimizes external dependencies to reduce supply chain risk and
binary size. The only non-stdlib runtime dependency is gRPC and its
required packages. Test dependencies (testify) are acceptable. Every new
dependency requires justification. Prefer stdlib solutions over third-party
packages when the stdlib solution is adequate.

### VI. Secrets Never Leak

Sensitive values (tokens, credentials, API keys, session secrets) must
never appear in error messages, log output, or public response types.
Credential fields in domain types are write-only: set during Create or
Update, never returned by Get or List. Error messages must omit secrets
even in debug/development builds. This principle applies to both SDK
code and test fixtures.

### VII. Deep Copy at Boundaries

Maps, slices, and other mutable reference types crossing the proto/SDK
boundary must be deep-copied. The converter layer must not share
references between internal proto state and public SDK types. Mutations
to a returned SDK struct must never affect the proto message it was
converted from, and vice versa.

### VIII. Doc Examples Compile

Code examples in package documentation (doc.go, godoc comments) must
use actual function signatures, type names, and argument counts. When
a public API signature changes, all documentation examples referencing
it must be updated in the same commit. Stale examples that do not
compile are treated as bugs.

### IX. Agent-Friendly Documentation

Every exported type, function, interface, and method MUST have a Go
doc comment. Doc comments MUST describe what the symbol does, not how
it is implemented. Interface method comments MUST list the error codes
the method can return (NotFound, AlreadyExists, InvalidArgument, etc.)
so that agents can write correct error handling without reading the
implementation. Non-obvious struct fields MUST have inline comments
explaining their purpose and valid values. Package-level `doc.go`
files MUST include runnable examples demonstrating primary use cases.
These conventions enable AI agents to understand the SDK via `go doc`,
LSP hover, and symbol search without reading source files.

### X. Proto-SDK Naming Fidelity

SDK domain type field names must reflect the semantic meaning of the
corresponding proto fields they map from. When a proto field is named
`host` (a hostname pattern), the SDK field must not be named `Name`.
Converter round-trip tests must assert that SDK field names match the
proto field semantics, not just that values are preserved. Misnamed
fields mislead consumers even when the converter works correctly.
This principle applies to all types in `v1/types/` that have proto
counterparts.

## Development Standards

- All `.go` files carry SPDX Apache-2.0 license headers
- golangci-lint enforces code quality with govet, errcheck, staticcheck,
  unused, ineffassign, revive, and goheader linters
- mise pins tool versions for reproducible builds
- CI validates lint, build, and test on every pull request

## Quality Gates

- `make lint` must pass with zero violations
- `make test` must pass with coverage output
- `make ci` runs the full pipeline (lint + build + test)
- All exported types and functions must have doc comments

## Governance

### XI. Fake-Real Parity

Fake implementations MUST mirror the real client's client-side
validation (nil checks, range validation, empty-string rejection) so
that tests using fakes catch the same caller bugs that production would
reject. When a real client method validates input before making an RPC
call, the corresponding fake method must perform the same validation
and return the same error type. Stubs that only return Unimplemented
without validation hide bugs that surface only in production.

### XII. Graceful Shutdown Order

When closing resources that combine context cancellation with
protocol-level close operations (e.g., gRPC CloseSend), the
protocol-level close MUST execute before context cancellation. Cancelling
the context first can cause the protocol-level close to fail with a
spurious context-cancelled error, which then propagates to the caller as
the Close() result. The correct order is: graceful close, then cancel,
then wait for background goroutines.

### XIII. Documentation Accompanies Features

Features that add or change public API surface MUST update documentation
in the same PR. This includes: README.md feature list and examples,
package-level doc.go with usage examples per Constitution VIII, and any
generated documentation site content. A feature PR that adds exported
symbols without corresponding README and doc.go updates is incomplete.
Trivial internal changes (refactors, bug fixes, test-only changes) are
exempt.

This constitution governs all design decisions for the OpenShell SDK
for Go. Amendments require documentation and a clear migration plan for
any breaking changes. All pull requests must verify compliance with these
principles.

**Version**: 1.5.0 | **Ratified**: 2026-06-27 | **Last Amended**: 2026-07-01
