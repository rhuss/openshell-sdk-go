# Tasks: Reverse Port Forwarding (ssh -R)

**Input**: Design documents from `specs/025-reverse-port-forwarding/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Tests ARE required (Constitution III: Test-First is NON-NEGOTIABLE).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Add the new types and extend the TCPInterface

- [x] T001 Add `RemoteListenOption`, `remoteListenConfig`, `WithRemoteBindAddress`, and `WithRemoteListenServiceID` to `openshell/v1/tcp.go`
- [x] T002 Add `RemoteListen` method signature to `TCPInterface` in `openshell/v1/tcp.go`

**Checkpoint**: Interface extended. Code does NOT compile (neither tcpClient nor fakeTCPClient satisfy TCPInterface yet).

---

## Phase 2: User Story 1 - Expose Local Service to Sandbox (Priority: P1)

**Goal**: Core `RemoteListen` method works on both real and fake clients with input validation.

**Independent Test**: Call `RemoteListen` on both real and fake clients with valid and invalid inputs, verify correct error types.

### Interfaces (from Phase 1)

```go
// In openshell/v1/tcp.go

type remoteListenConfig struct {
    bindAddress string
    serviceID   string
}

type RemoteListenOption func(*remoteListenConfig)

func WithRemoteBindAddress(addr string) RemoteListenOption
func WithRemoteListenServiceID(id string) RemoteListenOption

// Added to TCPInterface:
RemoteListen(ctx context.Context, workspace, sandboxName string, remotePort uint32, localTarget string, opts ...RemoteListenOption) error
```

### Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [x] T003 [P] [US1] Test real client `RemoteListen` returns `Unimplemented` for valid inputs in `openshell/v1/tcp_client_test.go`
- [x] T004 [P] [US1] Test real client `RemoteListen` returns `InvalidArgument` for empty sandbox name in `openshell/v1/tcp_client_test.go`
- [x] T005 [P] [US1] Test real client `RemoteListen` port validation (0, 65536, boundary ports 1 and 65535) in `openshell/v1/tcp_client_test.go`
- [x] T006 [P] [US1] Test real client `RemoteListen` returns `InvalidArgument` for malformed `localTarget` (missing port, no host) in `openshell/v1/tcp_client_test.go`
- [x] T007 [P] [US1] Test real client `RemoteListen` accepts valid `localTarget` formats including IPv6 (`[::1]:8080`) in `openshell/v1/tcp_client_test.go`
- [x] T008 [P] [US1] Test fake client `RemoteListen` returns `Unimplemented` for valid inputs in `openshell/v1/fake/tcp_test.go`
- [x] T009 [P] [US1] Test fake client `RemoteListen` validation parity (empty name, bad port, malformed target, closed client) in `openshell/v1/fake/tcp_test.go`

### Implementation for User Story 1

- [x] T010 [US1] Implement `RemoteListen` on `fakeTCPClient` with input validation in `openshell/v1/fake/tcp.go`
- [x] T011 [US1] Implement `RemoteListen` on `tcpClient` with input validation and `Unimplemented` return in `openshell/v1/tcp_client.go`
- [x] T012 [US1] Run `make test` to verify all new and existing tests pass

**Checkpoint**: Core RemoteListen compiles and all validation tests pass on both real and fake clients.

---

## Phase 3: User Story 2 - Custom Bind Address and Service Identification (Priority: P2)

**Goal**: Options (`WithRemoteBindAddress`, `WithRemoteListenServiceID`) are accepted and config values accessible.

**Independent Test**: Call `RemoteListen` with options on fake client, verify no option-related errors.

### Tests for User Story 2

- [x] T013 [P] [US2] Test fake client `RemoteListen` accepts `WithRemoteBindAddress` and `WithRemoteListenServiceID` options in `openshell/v1/fake/tcp_test.go`
- [x] T014 [P] [US2] Test real client `RemoteListen` accepts options without error (still returns `Unimplemented`) in `openshell/v1/tcp_client_test.go`

### Implementation for User Story 2

Options were already defined in T001. Fake and real client already accept variadic options in T010/T011. These tests verify that options don't cause errors.

- [x] T015 [US2] Run `make test` to verify option acceptance tests pass

**Checkpoint**: Options work on both clients without errors.

---

## Phase 4: User Story 3 - Graceful Error Handling (Priority: P2)

**Goal**: Error handling behavior is correctly specified in the stub (Unavailable for closed client).

**Independent Test**: Call `RemoteListen` on a closed client, verify `Unavailable` is returned.

### Tests for User Story 3

- [x] T016 [P] [US3] Test fake client `RemoteListen` on closed client returns `Unavailable` in `openshell/v1/fake/tcp_test.go`
- [x] T017 [P] [US3] Test real client `RemoteListen` on closed client returns `Unavailable` in `openshell/v1/tcp_client_test.go`

### Implementation for User Story 3

Closed-client check is already implemented in T010/T011. Context cancellation testing (FR-009) is deferred to the real gRPC implementation phase, since the stub returns Unimplemented immediately without blocking.

- [x] T018 [US3] Run `make test` to verify error handling tests pass

**Checkpoint**: All error paths tested and passing.

---

## Phase 5: User Story 4 - Fake Client Support (Priority: P3)

**Goal**: Fake client matches real client validation exactly (parity check).

**Independent Test**: Run identical input validation test cases against both real and fake, verify same error types.

### Tests for User Story 4

- [x] T019 [US4] Test validation parity between real and fake clients using table-driven test with shared test cases in `openshell/v1/tcp_client_test.go`

### Implementation for User Story 4

Parity was implemented in T010/T011. This test cross-validates both implementations.

- [x] T020 [US4] Run `make ci` to verify full pipeline passes (lint + build + test)

**Checkpoint**: All validation paths identical between real and fake clients.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and final validation

- [x] T021 [P] Add godoc comments for `RemoteListen`, `RemoteListenOption`, `WithRemoteBindAddress`, `WithRemoteListenServiceID` in `openshell/v1/tcp.go`
- [x] T022 [P] Add `RemoteListen` example to `openshell/v1/doc.go`
- [x] T023 Run `make ci` to verify full pipeline passes (lint + build + test)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies, starts immediately
- **Phase 2 (US1)**: Depends on Phase 1 (interface must exist)
- **Phase 3 (US2)**: Depends on Phase 2 (options tested against working method)
- **Phase 4 (US3)**: Depends on Phase 2 (error handling tested against working method)
- **Phase 5 (US4)**: Depends on Phases 2-4 (parity requires both implementations)
- **Phase 6 (Polish)**: Depends on all user stories

### Within Each User Story

- Tests MUST be written and FAIL before implementation
- Implementation makes tests pass
- `make test` checkpoint after each story

### Parallel Opportunities

- T003-T009 (all US1 tests) can run in parallel
- T013-T014 (US2 tests) can run in parallel
- T016-T017 (US3 tests) can run in parallel
- T021-T022 (docs) can run in parallel
- Phases 3 and 4 can run in parallel (different concerns, no file conflicts)

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Types and interface
2. Complete Phase 2: Core RemoteListen with validation
3. **STOP and VALIDATE**: `make test` passes, both clients compile

### Incremental Delivery

1. Phase 1 (Setup) + Phase 2 (US1) = Working stub with validation
2. Phase 3 (US2) = Options verified
3. Phase 4 (US3) = Error paths verified
4. Phase 5 (US4) = Parity cross-validated
5. Phase 6 (Polish) = Docs and final CI

## Notes

- All changes are within existing `openshell/v1/` package, no new packages
- Total: 23 tasks across 6 phases
- Files modified: `tcp.go`, `tcp_client.go`, `tcp_client_test.go`, `fake/tcp.go`, `fake/tcp_test.go`, `doc.go`
- The real client stub will be replaced with actual gRPC implementation when upstream proto support lands
