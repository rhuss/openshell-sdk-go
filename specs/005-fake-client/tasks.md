# Tasks: Fake Client Package

**Input**: Design documents from `specs/005-fake-client/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story. Tests are written first per Constitution Principle III (Test-First).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Prerequisites)

**Purpose**: Add missing error code and create package structure

- [x] T001 Add `ErrorUnimplemented` constant and `IsUnimplemented` helper to `openshell/v1/types/errors.go`, following the existing pattern for ErrorNotFound/IsNotFound

  **Interfaces produced**:
  - `ErrorUnimplemented ErrorCode` — new constant (value `ErrorInternal + 1`, i.e. `iota` continuation)
  - `IsUnimplemented(err error) bool` — helper function following `IsNotFound` pattern
  - `String()` case for `ErrorUnimplemented` returning `"Unimplemented"`

- [x] T002 [P] Re-export `ErrorUnimplemented` and `IsUnimplemented` via type alias in `openshell/v1/errors.go`
- [x] T003 [P] Create `openshell/v1/fake/` directory and `openshell/v1/fake/doc.go` with package documentation and SPDX header (example code added in final phase)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Generic ObjectStore and WatchBroadcaster that all sub-clients depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T004 [P] Test ObjectStore[T] CRUD operations (Create/Get/List/Delete, AlreadyExists, NotFound, idempotent delete, deep copy isolation) in `openshell/v1/fake/store_test.go`
- [x] T005 [P] Test WatchBroadcaster[T] (register watcher, broadcast events, name filtering, Stop closes channel, multiple watchers, StopAll) in `openshell/v1/fake/broadcaster_test.go`
- [x] T006 Implement generic `ObjectStore[T]` with `sync.RWMutex`, name extraction function, deep copy function, and CRUD methods in `openshell/v1/fake/store.go`

  **Interfaces produced** (consumed by T010, T014, T016, T027):
  - `type ObjectStore[T any] struct` — unexported, created via `newObjectStore`
  - `newObjectStore[T any](nameFunc func(T) string, copyFunc func(T) T) *ObjectStore[T]`
  - `(s *ObjectStore[T]) Create(obj T) (T, error)` — returns `AlreadyExists` StatusError if name exists
  - `(s *ObjectStore[T]) Get(name string) (T, error)` — returns `NotFound` StatusError if missing
  - `(s *ObjectStore[T]) List() []T` — returns deep-copied slice of all objects
  - `(s *ObjectStore[T]) Update(obj T) (T, error)` — returns `NotFound` StatusError if missing
  - `(s *ObjectStore[T]) Delete(name string)` — idempotent, no error on missing
  - `(s *ObjectStore[T]) Insert(obj T)` — direct insert without error (for pre-seeding)

- [x] T007 Implement `WatchBroadcaster[T]` with watcher registration, buffered channels (capacity 100), name-filtered broadcast, Stop, and StopAll in `openshell/v1/fake/broadcaster.go`

  **Interfaces produced** (consumed by T010, T012, T027):
  - `type WatchBroadcaster[T any] struct` — unexported, created via `newWatchBroadcaster`
  - `newWatchBroadcaster[T any]() *WatchBroadcaster[T]`
  - `(b *WatchBroadcaster[T]) Watch(name string) WatchInterface[T]` — register a watcher; if name non-empty, filters events
  - `(b *WatchBroadcaster[T]) Broadcast(event Event[T], name string)` — send event to all matching watchers
  - `(b *WatchBroadcaster[T]) StopAll()` — close all active watchers

**Checkpoint**: Foundation ready — run `make test` to verify store and broadcaster tests pass

---

## Phase 3: User Story 1 — Consumer Tests Sandbox Lifecycle (Priority: P1)

**Goal**: Full sandbox CRUD with WaitReady phase transition and correct StatusError codes

**Independent Test**: Create → WaitReady → Get → Delete → Get-returns-NotFound lifecycle in under 10 lines

- [x] T008 [P] [US1] Test sandbox Create (returns Provisioning phase), Get, List, Delete (idempotent), AlreadyExists, NotFound, deep copy isolation in `openshell/v1/fake/sandbox_test.go`
- [x] T009 [P] [US1] Test WaitReady (transitions Provisioning→Ready, updates store, context cancellation) in `openshell/v1/fake/sandbox_test.go`
- [x] T010 [US1] Implement `fakeSandboxClient` with Create, Get, List, Delete, WaitReady using ObjectStore and WatchBroadcaster in `openshell/v1/fake/sandbox.go`. Include deep copy helpers for Sandbox (maps, slices, nested SandboxTemplate pointer).

  **Interfaces produced** (consumed by T012, T027, T028):
  - `type fakeSandboxClient struct` — unexported, implements `SandboxInterface`
  - `newFakeSandboxClient(store *ObjectStore[*Sandbox], broadcaster *WatchBroadcaster[*Sandbox]) *fakeSandboxClient`
  - Constructor accepts a `closedFunc func() bool` for post-Close guard checks

---

## Phase 4: User Story 2 — Consumer Watches for Events (Priority: P1)

**Goal**: Watch broadcasts ADDED/MODIFIED/DELETED events to all active watchers on sandbox mutations

**Independent Test**: Start Watch, Create sandbox, verify ADDED event received; Delete sandbox, verify DELETED event; Stop watcher, verify channel closed

- [x] T011 [P] [US2] Test sandbox Watch (ADDED on Create, DELETED on Delete, MODIFIED on WaitReady, multiple watchers, name filtering, Stop) in `openshell/v1/fake/sandbox_test.go`
- [x] T012 [US2] Wire Watch method on `fakeSandboxClient` to register watchers via WatchBroadcaster and broadcast events from Create/Delete/WaitReady in `openshell/v1/fake/sandbox.go`

---

## Phase 5: User Story 3 — Consumer Pre-Seeds Test Fixtures (Priority: P2)

**Goal**: AddSandbox/AddProvider helpers insert objects into the store without triggering watch events

**Independent Test**: AddSandbox, then Get returns the sandbox and List includes it

- [x] T013 [P] [US3] Test AddSandbox and AddProvider pre-seed helpers (inserted objects retrievable via Get/List, no watch events triggered, deep copy isolation) in `openshell/v1/fake/fake_test.go`
- [x] T014 [US3] Implement AddSandbox and AddProvider methods on FakeClient that insert directly into ObjectStore bypassing watch broadcast in `openshell/v1/fake/fake.go`

---

## Phase 6: User Story 4 — Consumer Tests Provider CRUD (Priority: P2)

**Goal**: Full provider CRUD with Ensure (create-or-update) semantics

**Independent Test**: Create → Get → Update → Get-reflects-update → Ensure-creates → Ensure-updates → Delete → Get-NotFound

- [x] T015 [P] [US4] Test provider Create, Get, List, Delete, Update, Ensure (create-if-missing, update-if-exists), AlreadyExists, NotFound, deep copy isolation in `openshell/v1/fake/provider_test.go`
- [x] T016 [US4] Implement `fakeProviderClient` with Create, Get, List, Delete, Update, Ensure using ObjectStore in `openshell/v1/fake/provider.go`. Include deep copy helpers for Provider (maps).

  **Interfaces produced** (consumed by T027):
  - `type fakeProviderClient struct` — unexported, implements `ProviderInterface`
  - `newFakeProviderClient(store *ObjectStore[*Provider]) *fakeProviderClient`
  - Constructor accepts a `closedFunc func() bool` for post-Close guard checks

---

## Phase 7: User Story 5 — Consumer Tests Health Check (Priority: P3)

**Goal**: Configurable HealthResult (defaults to healthy)

**Independent Test**: Default FakeClient returns healthy; configured FakeClient returns custom unhealthy result

- [x] T017 [P] [US5] Test health Check (default healthy, configurable unhealthy result) in `openshell/v1/fake/health_test.go`
- [x] T018 [US5] Implement `fakeHealthClient` with configurable HealthResult in `openshell/v1/fake/health.go`

  **Interfaces produced** (consumed by T027):
  - `type fakeHealthClient struct` — unexported, implements `HealthInterface`
  - `newFakeHealthClient(result *HealthResult) *fakeHealthClient`
  - Constructor accepts a `closedFunc func() bool` for post-Close guard checks

---

## Phase 8: User Story 6 — Consumer Tests Concurrent Access (Priority: P3)

**Goal**: All operations pass `go test -race` under concurrent access

**Independent Test**: Multiple goroutines performing concurrent CRUD and Watch on shared FakeClient, no race detector violations

- [x] T019 [US6] Test concurrent sandbox Create/Get/List/Delete/Watch from multiple goroutines with race detector in `openshell/v1/fake/sandbox_test.go`
- [x] T020 [P] [US6] Test concurrent provider Create/Get/List/Delete from multiple goroutines with race detector in `openshell/v1/fake/provider_test.go`

---

## Phase 9: Stubs and Assembly

**Purpose**: Exec/File stubs, FakeClient assembly, Close behavior

- [x] T021 [P] Test exec stub methods (Run, Stream, Interactive) return Unimplemented StatusError in `openshell/v1/fake/exec_test.go`
- [x] T022 [P] Test file stub methods (Upload, Download) return Unimplemented StatusError in `openshell/v1/fake/file_test.go`
- [x] T023 [P] Implement `fakeExecClient` stub returning Unimplemented for all methods in `openshell/v1/fake/exec.go`

  **Interfaces produced** (consumed by T027):
  - `type fakeExecClient struct` — unexported, implements `ExecInterface`
  - `newFakeExecClient() *fakeExecClient`
  - Constructor accepts a `closedFunc func() bool` for post-Close guard checks

- [x] T024 [P] Implement `fakeFileClient` stub returning Unimplemented for all methods in `openshell/v1/fake/file.go`

  **Interfaces produced** (consumed by T027):
  - `type fakeFileClient struct` — unexported, implements `FileInterface`
  - `newFakeFileClient() *fakeFileClient`
  - Constructor accepts a `closedFunc func() bool` for post-Close guard checks
- [x] T025 [P] Test FakeClient Close (stops watchers, subsequent calls return Unavailable) in `openshell/v1/fake/fake_test.go`
- [x] T026 [P] Test AttachProvider, DetachProvider, ListProviders (basic in-memory association tracking) in `openshell/v1/fake/sandbox_test.go`
- [x] T027 Implement NewFakeClient constructor with WithHealthResult option, wire all sub-clients, implement Close with Unavailable guard in `openshell/v1/fake/fake.go`

  **Interfaces produced** (consumed by T029, all consumer test code):
  - `type FakeClient struct` — exported, implements `ClientInterface`
  - `type FakeClientOption func(*FakeClient)` — functional option type
  - `func NewFakeClient(opts ...FakeClientOption) *FakeClient`
  - `func WithHealthResult(r *HealthResult) FakeClientOption`
  - `func (c *FakeClient) AddSandbox(sb *Sandbox)` — pre-seed helper, no watch events
  - `func (c *FakeClient) AddProvider(p *Provider)` — pre-seed helper, no watch events
  - `func (c *FakeClient) Sandboxes() SandboxInterface`
  - `func (c *FakeClient) Providers() ProviderInterface`
  - `func (c *FakeClient) Exec() ExecInterface`
  - `func (c *FakeClient) Files() FileInterface`
  - `func (c *FakeClient) Health() HealthInterface`
  - `func (c *FakeClient) Close() error` — stops all watchers, sets closed flag
- [x] T028 Implement AttachProvider, DetachProvider, ListProviders on fakeSandboxClient with association map in `openshell/v1/fake/sandbox.go`

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, lint compliance, final validation

- [x] T029 Add compilable usage example to `openshell/v1/fake/doc.go` demonstrating sandbox lifecycle in ~10 lines
- [x] T030 Verify all `.go` files in `openshell/v1/fake/` have SPDX license headers
- [x] T031 Run `make ci` (lint + build + test) and fix any violations

---

## Dependencies

```
T001 (ErrorUnimplemented) ─┐
T002 (re-export)           ├─► T006, T007 (ObjectStore, Broadcaster)
T003 (doc.go)             ─┘         │
                                     ▼
                              T008-T012 (US1+US2: sandbox + watch)
                                     │
                              T013-T014 (US3: pre-seed) ◄── depends on FakeClient shell
                              T015-T016 (US4: provider) ◄── parallel with US3
                              T017-T018 (US5: health)   ◄── parallel with US3/US4
                                     │
                              T019-T020 (US6: concurrency) ◄── needs all CRUD implemented
                              T021-T028 (stubs + assembly)
                                     │
                              T029-T031 (polish)
```

## Parallel Execution Opportunities

- **Phase 1**: T002, T003 parallel after T001
- **Phase 2**: T004, T005 parallel (test files for store and broadcaster)
- **Phase 5-7**: US3, US4, US5 can run in parallel (independent sub-clients)
- **Phase 9**: T021-T026 all parallel (independent test files)

## Implementation Strategy

**MVP**: Phases 1-4 (Setup + Foundation + US1 + US2) — delivers sandbox lifecycle with watch events. This alone eliminates the majority of consumer mock boilerplate.

**Full Scope**: All 10 phases, 31 tasks total.
- US1+US2: 5 tasks (sandbox lifecycle + watch)
- US3: 2 tasks (pre-seed)
- US4: 2 tasks (provider CRUD)
- US5: 2 tasks (health)
- US6: 2 tasks (concurrency)
- Stubs + assembly: 8 tasks
- Polish: 3 tasks
