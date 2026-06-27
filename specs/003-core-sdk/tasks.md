# Tasks: Core SDK (Phase 1)

**Input**: Design documents from `/specs/003-core-sdk/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Included. Constitution principle III (Test-First) is NON-NEGOTIABLE: tests are written before or alongside implementation.

**Organization**: Tasks are grouped by user story, ordered by implementation dependency. Health (P3) and Provider (P2) come before Sandbox (P1) because they are simpler and establish the sub-client pattern that Sandbox follows.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create package structure and shared documentation

- [x] T001 Create v1 package directory structure: `openshell/v1/`, `openshell/v1/internal/grpc/`, `openshell/v1/internal/converter/`
- [x] T002 Create package doc.go with usage examples in `openshell/v1/doc.go`

---

## Phase 2: Foundational (US7 - Typed Error Handling + shared infrastructure)

**Purpose**: Core types, error handling, and client infrastructure that ALL sub-clients depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

### Tests for Foundation ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T003 [P] Test StatusError, ErrorCode, and Is* helpers in `openshell/v1/errors_test.go`
- [x] T004 [P] Test gRPC status code to StatusError mapping in `openshell/v1/internal/converter/errors_test.go`
- [x] T005 [P] Test timestamp conversions (ms epoch <-> time.Time) in `openshell/v1/internal/converter/time_test.go`
- [x] T006 [P] Test NewClient with valid config, empty address, and Close behavior in `openshell/v1/client_test.go`
- [x] T007 [P] Test NoAuth and StaticToken implementations in `openshell/v1/auth_test.go`

### Implementation for Foundation

- [x] T008 [P] Implement shared types: SandboxPhase, EventType, StreamType constants, RetryPolicy struct, and TLSConfig struct in `openshell/v1/types.go`
- [x] T009 [P] Implement Options structs (CreateOptions, GetOptions, ListOptions, DeleteOptions, UpdateOptions, WatchOptions, WaitOptions, ExecOptions) in `openshell/v1/options.go`
- [x] T010 [P] Implement Logger interface in `openshell/v1/logger.go`
- [x] T011 [P] Implement AuthProvider interface with NoAuth and StaticToken in `openshell/v1/auth.go`
- [x] T012 Implement StatusError, ErrorCode enum, and Is* helper functions (IsNotFound, IsAlreadyExists, IsUnavailable, IsPermissionDenied, IsInvalidArgument, IsDeadlineExceeded, IsCancelled) in `openshell/v1/errors.go`
- [x] T013 [P] Implement timestamp converter (ms epoch <-> time.Time) in `openshell/v1/internal/converter/time.go`
- [x] T014 [P] Implement gRPC error converter (gRPC status -> StatusError) in `openshell/v1/internal/converter/errors.go`
- [x] T015 Implement gRPC connection setup (dial options, TLS config, credentials) in `openshell/v1/internal/grpc/conn.go`
- [x] T016 Implement Client struct, NewClient, Close, and ClientInterface with sub-client accessor stubs in `openshell/v1/client.go`

**Checkpoint**: Foundation ready. Client can connect to gateway, errors are typed, all shared types exist. US7 acceptance scenarios can be verified.

---

## Phase 3: US5 - Health Checking (Priority: P3)

**Goal**: Validate gateway connectivity with a simple health check call

**Independent Test**: Call Health.Check against a mock server returning healthy/unhealthy status

### Tests for US5 ⚠️

- [x] T017 [US5] Test Health.Check success and unavailable error against mock gRPC server in `openshell/v1/health_client_test.go`

### Implementation for US5

- [x] T018 [US5] Define HealthInterface in `openshell/v1/health.go`
- [x] T019 [US5] Implement healthClient calling Health RPC in `openshell/v1/health_client.go`
- [x] T020 [US5] Wire Health() accessor on Client in `openshell/v1/client.go`

**Checkpoint**: `client.Health().Check(ctx)` works against a mock server. First sub-client pattern established.

---

## Phase 4: US3 - Provider Management (Priority: P2)

**Goal**: CRUD operations on providers plus idempotent Ensure

**Independent Test**: Create, Get, List, Update, Delete, and Ensure a provider against mock server

### Tests for US3 ⚠️

- [x] T021 [P] [US3] Test provider proto <-> SDK type conversion in `openshell/v1/internal/converter/provider_test.go`
- [x] T022 [US3] Test Provider CRUD (Create, Get, List, Update, Delete) and Ensure against mock gRPC server in `openshell/v1/provider_client_test.go`

### Implementation for US3

- [x] T023 [US3] Define Provider, ProviderSpec domain types and ProviderInterface in `openshell/v1/provider.go`
- [x] T024 [US3] Implement provider converter (proto Provider <-> SDK Provider) in `openshell/v1/internal/converter/provider.go`
- [x] T025 [US3] Implement providerClient with Create, Get, List, Update, Delete methods in `openshell/v1/provider_client.go`
- [x] T026 [US3] Implement Ensure (Get + Create/Update) on providerClient in `openshell/v1/provider_client.go`
- [x] T027 [US3] Wire Providers() accessor on Client in `openshell/v1/client.go`

**Checkpoint**: `client.Providers().Create/Get/List/Update/Delete/Ensure(ctx, ...)` all work against mock server.

---

## Phase 5: US1 - Sandbox Lifecycle Management (Priority: P1)

**Goal**: Full sandbox CRUD plus provider attachment/detachment and WaitReady

**Independent Test**: Create, Get, List, Delete sandboxes, attach/detach providers, WaitReady with mock server

### Tests for US1 ⚠️

- [x] T028 [P] [US1] Test sandbox proto <-> SDK type conversion (including SandboxPhase mapping) in `openshell/v1/internal/converter/sandbox_test.go`
- [x] T029 [US1] Test Sandbox CRUD (Create, Get, List, Delete) against mock gRPC server in `openshell/v1/sandbox_client_test.go`
- [x] T030 [US1] Test AttachProvider, DetachProvider, ListProviders against mock gRPC server in `openshell/v1/sandbox_client_test.go`
- [x] T031 [US1] Test WaitReady (success, timeout, sandbox-failed cases) against mock gRPC server in `openshell/v1/sandbox_client_test.go`

### Implementation for US1

- [x] T032 [US1] Define Sandbox, SandboxSpec, SandboxCondition domain types and SandboxInterface in `openshell/v1/sandbox.go`
- [x] T033 [US1] Implement sandbox converter (proto Sandbox <-> SDK Sandbox, SandboxPhase mapping) in `openshell/v1/internal/converter/sandbox.go`
- [x] T034 [US1] Implement sandboxClient with Create, Get, List, Delete methods in `openshell/v1/sandbox_client.go`
- [x] T035 [US1] Implement AttachProvider, DetachProvider, ListProviders on sandboxClient in `openshell/v1/sandbox_client.go`
- [x] T036 [US1] Implement WaitReady on sandboxClient (poll Get until Ready/Failed or context deadline) in `openshell/v1/sandbox_client.go`
- [x] T037 [US1] Wire Sandboxes() accessor on Client in `openshell/v1/client.go`

**Checkpoint**: Full sandbox lifecycle works against mock server. US1 acceptance scenarios verified.

---

## Phase 6: US6 - Watching Sandbox State Changes (Priority: P3)

**Goal**: Watch sandbox events via channel-based interface compatible with Kubernetes watch pattern

**Independent Test**: Start a watch, receive ADDED/MODIFIED/DELETED/ERROR events on channel

### Tests for US6 ⚠️

- [x] T038 [US6] Test WatchInterface event delivery, Stop, and error handling against mock streaming server in `openshell/v1/watch_test.go`
- [x] T039 [US6] Test Sandbox.Watch integration (WatchSandbox RPC -> Event[Sandbox] channel) in `openshell/v1/sandbox_client_test.go`

### Implementation for US6

- [x] T040 [US6] Implement WatchInterface[T], Event[T], and watcher goroutine in `openshell/v1/watch.go`
- [x] T041 [US6] Implement Watch method on sandboxClient (WatchSandbox RPC -> filter SandboxStreamEvent to status events -> Event[Sandbox]) in `openshell/v1/sandbox_client.go`

**Checkpoint**: `client.Sandboxes().Watch(ctx, opts)` delivers typed events on a channel.

---

## Phase 7: US2 - Command Execution (Priority: P1)

**Goal**: Run, Stream, and Interactive command execution modes

**Independent Test**: Run a command and get ExecResult, stream output chunks, interactive I/O

### Tests for US2 ⚠️

- [X] T042 [P] [US2] Test exec event conversion (ExecSandboxEvent -> ExecChunk/ExecResult) in `openshell/v1/internal/converter/exec_test.go`
- [X] T043 [US2] Test Run (collect stream into ExecResult) and Stream (iterate ExecChunks) against mock streaming server in `openshell/v1/exec_client_test.go`
- [X] T044 [US2] Test Interactive session (bidirectional stream, Read/Write/Resize/Close) against mock bidirectional server in `openshell/v1/exec_client_test.go`

### Implementation for US2

- [X] T045 [US2] Define ExecResult, ExecStream, ExecChunk, InteractiveSession, StreamType types and ExecInterface in `openshell/v1/exec.go`
- [X] T046 [US2] Implement exec converter (ExecSandboxEvent oneof -> ExecChunk) in `openshell/v1/internal/converter/exec.go`
- [X] T047 [US2] Implement execClient.Run (ExecSandbox RPC, collect all events into ExecResult) in `openshell/v1/exec_client.go`
- [X] T048 [US2] Implement execClient.Stream (ExecSandbox RPC, yield ExecChunks via Next(), track exit code) in `openshell/v1/exec_client.go`
- [X] T049 [US2] Implement execClient.Interactive (ExecSandboxInteractive bidirectional stream, io.Reader/Writer, Resize) in `openshell/v1/exec_client.go`
- [X] T050 [US2] Wire Exec() accessor on Client in `openshell/v1/client.go`

**Checkpoint**: All three exec modes work against mock server. US2 acceptance scenarios verified.

---

## Phase 8: US4 - File Transfer (Priority: P2)

**Goal**: Upload and Download files to/from sandboxes

**Independent Test**: Upload a file to sandbox, download it back, compare contents

### Tests for US4 ⚠️

- [X] T051 [US4] Test Upload and Download against mock SSH session server in `openshell/v1/file_client_test.go`
- [X] T052 [US4] Test Upload error cases (non-existent local file, transfer interruption) in `openshell/v1/file_client_test.go`

### Implementation for US4

- [X] T053 [US4] Define FileInterface in `openshell/v1/file.go`
- [X] T054 [US4] Implement fileClient using CreateSshSession RPC internally for Upload and Download in `openshell/v1/file_client.go`
- [X] T055 [US4] Wire Files() accessor on Client in `openshell/v1/client.go`

**Checkpoint**: `client.Files().Upload/Download(ctx, ...)` works against mock server. US4 acceptance scenarios verified.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Cleanup, documentation, and CI validation

- [X] T056 Remove old stub `openshell/client.go` and `openshell/client_test.go`
- [X] T057 Create integration test stubs with `//go:build integration` tag in `openshell/v1/integration_test.go`
- [X] T058 Verify `make ci` passes (lint + build + test)
- [X] T059 Verify all SPDX license headers are present on new .go files

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup - BLOCKS all user stories
- **US5 Health (Phase 3)**: Depends on Foundational - establishes sub-client pattern
- **US3 Provider (Phase 4)**: Depends on Foundational - independent of US5
- **US1 Sandbox (Phase 5)**: Depends on Foundational - independent of US3/US5
- **US6 Watch (Phase 6)**: Depends on US1 (sandbox types and sandbox_client.go)
- **US2 Exec (Phase 7)**: Depends on Foundational - independent of US1/US3/US5
- **US4 File (Phase 8)**: Depends on Foundational - independent of other stories
- **Polish (Phase 9)**: Depends on all user stories being complete

### User Story Dependencies

- **US5 Health**: Independent (after Foundation)
- **US3 Provider**: Independent (after Foundation)
- **US1 Sandbox**: Independent (after Foundation)
- **US6 Watch**: Depends on US1 (sandbox types)
- **US2 Exec**: Independent (after Foundation)
- **US4 File**: Independent (after Foundation)

### Within Each User Story

- Tests MUST be written FIRST and FAIL before implementation
- Converters before client implementations
- Type definitions before implementations
- Client wiring (accessor on Client) after implementation

### Parallel Opportunities

- T003-T007: All foundation tests can run in parallel
- T008-T011: Shared type files can be written in parallel
- T013-T014: Converter utilities can be written in parallel
- US5, US3, US1, US2, US4: Can all start after Foundation (if team capacity allows)
- Within each story: converter tests and converter implementation can parallel with type definitions

---

## Parallel Example: Foundation Phase

```bash
# Write all foundation tests in parallel:
T003: Test StatusError in errors_test.go
T004: Test error converter in internal/converter/errors_test.go
T005: Test timestamp converter in internal/converter/time_test.go
T006: Test NewClient in client_test.go
T007: Test auth implementations in auth_test.go

# Write all shared type files in parallel:
T008: types.go (SandboxPhase, EventType, StreamType)
T009: options.go (Options structs)
T010: logger.go (Logger interface)
T011: auth.go (AuthProvider, NoAuth, StaticToken)
```

---

## Implementation Strategy

### MVP First (Foundation + Health + Provider)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (errors, types, client)
3. Complete Phase 3: Health (simplest sub-client, validates pattern)
4. Complete Phase 4: Provider (CRUD + Ensure, validates full pattern)
5. **STOP and VALIDATE**: `make ci` passes, pattern is solid
6. Continue with Sandbox, Watch, Exec, File

### Incremental Delivery

1. Foundation + Health → Gateway connectivity verified
2. + Provider → CRUD pattern established with Ensure
3. + Sandbox → Core resource lifecycle complete
4. + Watch → Operator-ready with event streaming
5. + Exec → Command execution in sandboxes
6. + File → File transfer capability
7. Polish → CI green, docs updated, stubs removed

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Tests are REQUIRED per constitution principle III (Test-First)
- Commit after each task or logical group
- Implementation order differs from priority order for practical reasons (simpler sub-clients first to establish patterns)
