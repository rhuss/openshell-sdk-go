# Implementation Plan: Fake Client Package

**Branch**: `005-fake-client` | **Date**: 2026-06-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/005-fake-client/spec.md`

## Summary

Implement an in-memory fake client package at `openshell/v1/fake/` that implements all SDK interfaces (ClientInterface, SandboxInterface, ProviderInterface, HealthInterface, ExecInterface, FileInterface). The fake provides CRUD state tracking, watch event broadcasting, and lifecycle simulation for consumer test suites, following the client-go/kubernetes/fake pattern. A prerequisite step adds the missing `ErrorUnimplemented` error code to the types package.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: None beyond the SDK itself (`openshell/v1/types/`) and Go stdlib (`sync`, `context`)
**Storage**: In-memory maps (no persistence)
**Testing**: Go testing + testify (assert/require), `go test -race`
**Target Platform**: Go library (any OS)
**Project Type**: Library (test support package)
**Performance Goals**: N/A (test-time only)
**Constraints**: Zero external dependencies, thread-safe, deep copy at boundaries
**Scale/Scope**: ~500-700 lines of implementation + ~400-500 lines of tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Fake uses domain types from `v1/types/`, never proto types |
| II. Idiomatic Go | PASS | Interfaces for testability, functional options pattern not needed (simple constructor) |
| III. Test-First (NON-NEGOTIABLE) | PASS | Tests written before each implementation file |
| IV. Upstream Tracking | N/A | No proto changes needed |
| V. Minimal Dependencies | PASS | Zero new dependencies — stdlib only |
| VI. Secrets Never Leak | PASS | Provider.Spec.Credentials stored as-is in fake (same as real client behavior for returned objects) |
| VII. Deep Copy at Boundaries | PASS | ObjectStore deep-copies on insert and retrieval |
| VIII. Doc Examples Compile | PASS | Package doc.go will include compilable example |

## Project Structure

### Documentation (this feature)

```text
specs/005-fake-client/
├── plan.md              # This file
├── research.md          # Phase 0: research decisions
├── data-model.md        # Phase 1: entity model
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (created by /speckit.tasks)
```

### Source Code (repository root)

```text
openshell/v1/
├── types/
│   └── errors.go            # Add ErrorUnimplemented + IsUnimplemented (prerequisite)
└── fake/
    ├── doc.go               # Package documentation with compilable example
    ├── store.go             # Generic ObjectStore[T] with thread-safe CRUD
    ├── store_test.go        # ObjectStore unit tests
    ├── broadcaster.go       # WatchBroadcaster[T] for event distribution
    ├── broadcaster_test.go  # WatchBroadcaster unit tests
    ├── fake.go              # FakeClient, NewFakeClient, option types
    ├── fake_test.go         # FakeClient integration tests (lifecycle, Close)
    ├── sandbox.go           # fakeSandboxClient implementing SandboxInterface
    ├── sandbox_test.go      # Sandbox CRUD, WaitReady, Watch, AttachProvider tests
    ├── provider.go          # fakeProviderClient implementing ProviderInterface
    ├── provider_test.go     # Provider CRUD, Ensure, Update tests
    ├── health.go            # fakeHealthClient implementing HealthInterface
    ├── health_test.go       # Health check tests (default + configurable)
    ├── exec.go              # fakeExecClient stub (returns Unimplemented)
    ├── exec_test.go         # Exec stub tests
    ├── file.go              # fakeFileClient stub (returns Unimplemented)
    └── file_test.go         # File stub tests
```

**Structure Decision**: Single `fake/` package under `openshell/v1/` following the client-go convention (`k8s.io/client-go/kubernetes/fake`). Internal building blocks (ObjectStore, WatchBroadcaster) are unexported. All sub-client implementations are in separate files for navigability.

## Design Decisions

### D1: Constructor API

`NewFakeClient(opts ...FakeClientOption)` with functional options:
- `WithHealthResult(r *HealthResult)` — configure the health check response

No other options needed for v1. The constructor creates empty stores, initializes the broadcaster, and wires all sub-clients.

### D2: Pre-Seed API

`AddSandbox(sb *Sandbox)` and `AddProvider(p *Provider)` are methods on `FakeClient` (not on sub-clients). This follows the client-go pattern where `fake.NewSimpleClientset(objects...)` accepts pre-existing objects at the top level. Pre-seeded objects are deep-copied into the store without triggering watch events.

### D3: Deep Copy Strategy

Each type gets a `deepCopy` function in the fake package (unexported). These are passed to `ObjectStore` as the `copyFunc`. Maps and slices are copied element-by-element. Nested pointers (e.g., `SandboxTemplate`) are copied recursively.

### D4: Watch Name Filtering

`Watch(ctx, name, opts...)`: if `name` is non-empty, the watcher only receives events for that specific sandbox. If empty, all events are delivered. Filtering happens in the broadcaster's send loop, not in the watcher channel.

### D5: AttachProvider/DetachProvider

Provider-to-sandbox associations are tracked in a `map[string][]string` (sandbox name → provider names) on the sandbox client. AttachProvider/DetachProvider modify this map and update the sandbox's ResourceVersion. They also broadcast MODIFIED events.

### D6: Error Code Addition

Add `ErrorUnimplemented` to `openshell/v1/types/errors.go` as a prerequisite. This is a backward-compatible addition (new enum value, new helper function). The String() method and IsUnimplemented() helper follow the existing pattern.

## Global Constraints

*Copied verbatim from spec — every task implicitly inherits these.*

- **FR-012**: All fake operations MUST be safe for concurrent use from multiple goroutines.
- **FR-014**: The fake package MUST reside at `openshell/v1/fake/`.
- **FR-015**: The fake MUST have no external dependencies beyond the SDK itself and the Go standard library.
- **Constitution III**: Tests written before each implementation file (test-first).
- **SPDX headers**: Every `.go` file must start with the SPDX license header.
- **Deep copy at boundaries**: ObjectStore deep-copies on insert and retrieval.

## Implementation Order

The implementation follows a bottom-up approach: infrastructure first (store, broadcaster), then sub-clients in priority order, then the top-level FakeClient assembly.

1. **Prerequisite**: Add ErrorUnimplemented to types package
2. **Foundation**: ObjectStore[T] generic store
3. **Foundation**: WatchBroadcaster[T]
4. **Sub-client**: fakeSandboxClient (P1 — highest priority)
5. **Sub-client**: fakeProviderClient (P2)
6. **Sub-client**: fakeHealthClient (P3)
7. **Sub-client**: fakeExecClient stub (P3)
8. **Sub-client**: fakeFileClient stub (P3)
9. **Assembly**: FakeClient, NewFakeClient, options, pre-seed helpers, Close
10. **Polish**: doc.go with compilable example, race detector test
