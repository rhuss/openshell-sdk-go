# Brainstorm: Fake Client Package

**Date:** 2026-06-28
**Status:** active

## Problem Framing

The SDK provides interfaces on every sub-client, enabling consumers to
write their own mocks. But writing a mock that simulates realistic
sandbox lifecycle (create returns Pending, WaitReady transitions to Ready,
Watch emits events, delete then get returns NotFound) requires substantial
boilerplate in every consumer test suite.

A `fake` package following the `client-go/kubernetes/fake` pattern would
give consumers a ready-made in-memory implementation that handles CRUD
state, watch event broadcasting, and lifecycle simulation out of the box.

## Approaches Considered

### A: In-Memory Object Store (Chosen)

Follow the `client-go/kubernetes/fake` pattern. An in-memory store
implements all SDK interfaces with real CRUD semantics:

- Create adds objects to store, returns them with generated metadata
- Get retrieves by name from store, returns NotFound if missing
- List returns all objects matching optional label selector
- Delete removes from store, subsequent Get returns NotFound
- Watch broadcasts typed events (ADDED/MODIFIED/DELETED) to all watchers
- WaitReady simulates phase transitions (configurable delay)
- Pre-seed helpers (AddSandbox, AddProvider) for simple setup
- Optional reactors for error injection and custom behavior

- Pros: Full lifecycle simulation. Watch works naturally. Familiar
  to Kubernetes developers. Simple cases are still simple. Stateful
  tests verify side effects. No external dependencies.
- Cons: ~300-400 lines of code. Thread-safe map + watch broadcaster.
  Reactors add complexity.

### B: Function-Field Fake

A FakeClient struct where each method is a configurable function field
that the consumer sets per test.

- Pros: Dead simple. Maximum flexibility. Zero magic.
- Cons: No state between calls. Every test manually wires responses.
  Watch simulation is painful. Tests read like mock setup.

### C: Generated Mocks (gomock/mockery)

Provide go:generate directives to produce mock implementations from
interfaces automatically.

- Pros: Zero maintenance. Always in sync. Familiar to Go developers.
- Cons: Adds code generation dependency. Verbose setup. No lifecycle
  semantics. Not the client-go pattern.

## Decision

**Approach A: In-memory object store.** The client-go fake pattern is
familiar to the SDK's primary audience and handles both simple and
complex test scenarios without consumers writing their own state
management. The in-memory store covers Sandbox, Provider, and Health
sub-clients. Exec and File interfaces are left to consumer mocking
(too complex to simulate meaningfully).

## Key Requirements

- Package path: `openshell/v1/fake/` (or `openshell/fake/`)
- Implements ClientInterface and all sub-client interfaces
- Thread-safe: safe for concurrent use in parallel test cases
- CRUD state: in-memory map keyed by resource name
- Watch: channel-based event broadcaster, triggers on Create/Update/Delete
- WaitReady: configurable phase transition (immediate or delayed)
- Pre-seed: AddSandbox, AddProvider helpers for test fixtures
- Typed errors: returns StatusError (NotFound, AlreadyExists) matching
  real client behavior
- Reactors (optional): allow consumers to intercept and modify operations
- No external dependencies beyond the SDK itself
- Follows constitution: Test-First, no new dependencies

## Open Questions

- Should reactors be included in v1 or deferred to a follow-up?
- Should the fake support label-based filtering in List, or just return all?
- Should Exec fake return a configurable ExecResult, or should consumers
  mock ExecInterface directly?
- Package path: `openshell/v1/fake/` (scoped to v1) or `openshell/fake/`
  (shared across versions)?
