# Feature Specification: Fake Client Package

**Feature Branch**: `005-fake-client`
**Created**: 2026-06-28
**Status**: Draft
**Input**: In-memory fake client package following the client-go/kubernetes/fake pattern for consumer test suites.

## User Scenarios & Testing

### User Story 1 - Consumer Tests Sandbox Lifecycle (Priority: P1)

A consumer writes a test that creates a sandbox, waits for it to become ready, then deletes it. Today, the consumer must build a custom mock that tracks state across Create, WaitReady, Get, and Delete calls. With the fake package, the consumer instantiates a FakeClient, calls Create (sandbox appears in store with Pending phase), calls WaitReady (sandbox transitions to Ready), calls Delete (sandbox is removed), and a subsequent Get returns NotFound.

**Why this priority**: Sandbox lifecycle is the most common consumer interaction. Covering Create/Get/List/Delete/WaitReady with realistic state transitions eliminates the majority of mock boilerplate.

**Independent Test**: Create a FakeClient, exercise the full sandbox lifecycle (Create → WaitReady → Get → Delete → Get-returns-NotFound), and verify each step returns the expected state.

**Acceptance Scenarios**:

1. **Given** a FakeClient with no pre-seeded data, **When** the consumer calls Sandbox.Create with name "test-sb", **Then** the returned Sandbox has phase Pending and the sandbox is retrievable via Get.
2. **Given** a sandbox "test-sb" exists with phase Pending, **When** the consumer calls WaitReady, **Then** the sandbox transitions to phase Ready and is returned.
3. **Given** a sandbox "test-sb" exists, **When** the consumer calls Delete, **Then** a subsequent Get returns a NotFound StatusError.
4. **Given** a FakeClient, **When** the consumer calls Create with a name that already exists, **Then** an AlreadyExists StatusError is returned.

---

### User Story 2 - Consumer Watches for Events (Priority: P1)

A consumer writes a test that watches for sandbox events. The consumer creates a Watch, then performs Create/Delete operations, and verifies that the watcher receives ADDED and DELETED events with the correct sandbox objects.

**Why this priority**: Watch is essential for consumers building controllers or reactive UIs. Without fake watch support, consumers must build their own channel-based event broadcaster.

**Independent Test**: Create a FakeClient, start a Watch, perform Create and Delete, and verify the watcher channel receives matching ADDED and DELETED events.

**Acceptance Scenarios**:

1. **Given** a FakeClient with an active sandbox Watch, **When** a sandbox is created, **Then** the watcher receives an ADDED event containing the created sandbox.
2. **Given** a FakeClient with an active sandbox Watch, **When** a sandbox is deleted, **Then** the watcher receives a DELETED event containing the deleted sandbox.
3. **Given** multiple active watchers on the same FakeClient, **When** a sandbox is created, **Then** all watchers receive the ADDED event.
4. **Given** a watcher, **When** the consumer calls Stop, **Then** the event channel is closed and no further events are delivered.

---

### User Story 3 - Consumer Pre-Seeds Test Fixtures (Priority: P2)

A consumer writes a test that starts with known sandbox and provider state. Instead of calling Create for each object, the consumer uses AddSandbox and AddProvider helpers to pre-populate the fake store before running the test logic.

**Why this priority**: Pre-seeding reduces test setup boilerplate. Most tests need specific objects to already exist before exercising the behavior under test.

**Independent Test**: Pre-seed a FakeClient with AddSandbox and AddProvider, then verify Get and List return the pre-seeded objects.

**Acceptance Scenarios**:

1. **Given** a FakeClient, **When** the consumer calls AddSandbox with a Sandbox object, **Then** Get returns that sandbox and List includes it.
2. **Given** a FakeClient, **When** the consumer calls AddProvider with a Provider object, **Then** Provider.Get returns that provider and Provider.List includes it.
3. **Given** a pre-seeded sandbox, **When** the consumer calls Delete on it, **Then** it is removed from the store.

---

### User Story 4 - Consumer Tests Provider CRUD (Priority: P2)

A consumer writes a test exercising provider Create, Get, List, Update, Delete, and Ensure operations. The fake implements the same CRUD semantics as the real client, including Ensure (create-or-update).

**Why this priority**: Provider management is the second most common sub-client. Ensure has specific semantics (create if missing, update if exists) that consumers need to verify.

**Independent Test**: Exercise provider Create, Get, Update, Ensure, List, and Delete against a FakeClient and verify each returns the expected state.

**Acceptance Scenarios**:

1. **Given** no providers exist, **When** the consumer calls Provider.Ensure, **Then** the provider is created and returned.
2. **Given** a provider exists, **When** the consumer calls Provider.Ensure with updated fields, **Then** the provider is updated and the updated version is returned.
3. **Given** a provider exists, **When** the consumer calls Provider.Update, **Then** the updated provider is returned and subsequent Get reflects the changes.
4. **Given** a provider exists, **When** the consumer calls Provider.Delete, **Then** a subsequent Get returns NotFound.

---

### User Story 5 - Consumer Tests Health Check (Priority: P3)

A consumer writes a test that checks gateway health. The fake returns a configurable HealthResult so consumers can test both healthy and unhealthy code paths.

**Why this priority**: Health checks are simple but necessary for completeness. Consumers need to test error handling when the gateway is unhealthy.

**Independent Test**: Create a FakeClient with default health (healthy), call Health.Check, verify success. Create a FakeClient with unhealthy status, call Health.Check, verify the unhealthy result.

**Acceptance Scenarios**:

1. **Given** a FakeClient with default configuration, **When** the consumer calls Health.Check, **Then** a healthy HealthResult is returned.
2. **Given** a FakeClient configured with an unhealthy status, **When** the consumer calls Health.Check, **Then** the configured unhealthy HealthResult is returned.

---

### User Story 6 - Consumer Tests Concurrent Access (Priority: P3)

A consumer runs parallel test cases that share a FakeClient. The fake is thread-safe: concurrent Create, Get, List, Delete, and Watch operations do not cause data races or panics.

**Why this priority**: Go tests run in parallel by default. A thread-unsafe fake would produce flaky tests that are difficult to diagnose.

**Independent Test**: Run multiple goroutines performing concurrent CRUD and Watch operations against a shared FakeClient, verify no race conditions with `go test -race`.

**Acceptance Scenarios**:

1. **Given** a shared FakeClient, **When** multiple goroutines concurrently call Create, Get, List, and Delete, **Then** no data race is detected by the Go race detector.
2. **Given** a shared FakeClient with active watchers, **When** multiple goroutines concurrently create and delete sandboxes, **Then** all watchers receive consistent events without panics.

---

### Edge Cases

- What happens when Get is called with an empty name? A NotFound StatusError is returned.
- What happens when Delete is called for a non-existent sandbox? The operation succeeds silently (idempotent delete), matching real client behavior.
- What happens when WaitReady is called with a context that is already cancelled? The context error is returned immediately.
- What happens when Watch is started and then Stop is called before any events? The channel is closed cleanly with no events delivered.
- What happens when Exec or File methods are called on the FakeClient? An Unimplemented StatusError is returned, directing consumers to mock ExecInterface/FileInterface directly.
- What happens when Close is called on the FakeClient? All active watchers are stopped and the client becomes unusable (subsequent calls return an error).

## Requirements

### Functional Requirements

- **FR-001**: The fake package MUST provide a FakeClient that implements ClientInterface and all sub-client interfaces (SandboxInterface, ProviderInterface, HealthInterface, ExecInterface, FileInterface).
- **FR-002**: The fake MUST maintain an in-memory store keyed by resource name for sandboxes and providers, supporting Create, Get, List, and Delete operations with correct state tracking.
- **FR-003**: Create operations MUST return an AlreadyExists StatusError when an object with the same name already exists.
- **FR-004**: Get operations MUST return a NotFound StatusError when no object with the given name exists.
- **FR-005**: Delete operations MUST be idempotent — deleting a non-existent object succeeds silently.
- **FR-006**: Sandbox.WaitReady MUST simulate phase transition from Pending to Ready, returning the updated sandbox. The transition MUST be immediate (no artificial delay) by default.
- **FR-007**: Watch MUST broadcast typed events (ADDED, MODIFIED, DELETED) to all active watchers when Create, Update, or Delete operations occur.
- **FR-008**: Watch MUST support Stop to close the event channel and stop receiving events. Watch accepts a name parameter: if non-empty, only events for the named sandbox are delivered; if empty, events for all sandboxes are delivered.
- **FR-009**: The fake MUST provide AddSandbox and AddProvider pre-seed helpers that insert objects into the store without triggering watch events.
- **FR-010**: Health.Check MUST return a configurable HealthResult (defaults to healthy).
- **FR-011**: ExecInterface and FileInterface methods MUST return an Unimplemented StatusError. If no `ErrorUnimplemented` constant exists in the SDK's ErrorCode enum, it MUST be added as a prerequisite.
- **FR-012**: All fake operations MUST be safe for concurrent use from multiple goroutines.
- **FR-013**: Close MUST stop all active watchers and prevent further operations. Subsequent calls to any method after Close MUST return an Unavailable StatusError.
- **FR-014**: The fake package MUST reside at `openshell/v1/fake/`.
- **FR-015**: The fake MUST have no external dependencies beyond the SDK itself and the Go standard library.
- **FR-016**: Sandbox.AttachProvider, DetachProvider, and ListProviders MUST be implemented with basic in-memory semantics, tracking provider-to-sandbox associations in the store.

### Key Entities

- **FakeClient**: The top-level fake implementing ClientInterface. Holds the in-memory stores and provides sub-client accessors.
- **ObjectStore**: A thread-safe, generic in-memory map storing objects by name. Used for both sandboxes and providers.
- **WatchBroadcaster**: A mechanism that distributes typed watch events to all registered watchers when store mutations occur.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Consumers can write a complete sandbox lifecycle test (create, wait-ready, get, delete, verify-not-found) in under 10 lines of test code using the fake package.
- **SC-002**: All fake operations pass the Go race detector (`go test -race`) under concurrent access from multiple goroutines.
- **SC-003**: The fake package compiles and passes all tests with zero external dependencies beyond the SDK and Go standard library.
- **SC-004**: Watch events are delivered to all active watchers within the same goroutine scheduling cycle as the triggering mutation (no artificial delays).
- **SC-005**: The fake produces the same StatusError codes (NotFound, AlreadyExists, Unimplemented) as the real client for equivalent error conditions.

## Assumptions

- The SDK is pre-1.0 with limited external consumers, so the fake API can evolve without strict backward-compatibility guarantees.
- Reactors (consumer-defined interceptors for error injection and custom behavior) are deferred to a follow-up feature, not included in this version.
- Label-based filtering in List is deferred — List returns all objects in the store.
- Sandbox.AttachProvider, DetachProvider, and ListProviders are included in the SandboxInterface and the fake implements them with basic in-memory semantics.
- The WaitReady phase transition is immediate by default. Configurable delays are deferred to a follow-up.
- Pre-seed helpers (AddSandbox, AddProvider) do not trigger watch events, matching the client-go convention where pre-seeded objects are treated as pre-existing state.
- The fake does not simulate network errors or timeouts — consumers who need those scenarios should use custom mocks.
