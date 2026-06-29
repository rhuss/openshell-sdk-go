# Tasks: SSH Tunneling and TCP Forward Options

**Input**: Design documents from `specs/009-ssh-tunnel-forward-opts/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Tests are written alongside implementation per Constitution III (Test-First).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Types and interface changes that all user stories depend on

- [X] T001 Add TunnelOption type, tunnelConfig struct, and WithTunnelServiceID constructor in openshell/v1/ssh.go
- [X] T002 [P] Add ForwardOption type, forwardConfig struct, and WithForwardServiceID constructor in openshell/v1/tcp.go
- [X] T003 Add Tunnel method signature to SSHInterface in openshell/v1/ssh.go (depends on T001)
- [X] T004 [P] Update Forward method signature on TCPInterface to accept variadic ForwardOption in openshell/v1/tcp.go (depends on T002)
**Checkpoint**: Interfaces updated, project should compile (fakes will fail compile-time checks until Phase 2 and 3 update them)

---

## Phase 2: User Story 2 - Service ID on TCP Forward (Priority: P2)

**Goal**: Add service ID option to TCP Forward method so the field appears in the initial protocol frame.

**Independent Test**: Call Forward with WithForwardServiceID and verify the init frame includes service_id.

**Why P2 before P1**: The TCP forward options change is simpler and establishes the functional options pattern. The SSH tunnel (P1) builds on this pattern and needs Forward to work with init frame customization internally.

### Implementation for User Story 2

- [X] T006 [US2] Update tcpClient.Forward to accept variadic ForwardOption, build forwardConfig, and set ServiceId on TcpForwardInit in openshell/v1/tcp_client.go
- [X] T007 [US2] Add tests for Forward with WithForwardServiceID (service_id in init frame) and without options (backward compat) in openshell/v1/tcp_client_test.go
- [X] T008 [P] [US2] Update fakeTCPClient.Forward to accept variadic ForwardOption in openshell/v1/fake/tcp.go
- [X] T009 [P] [US2] Add test for fake Forward with ForwardOption in openshell/v1/fake/tcp_test.go

**Checkpoint**: TCP Forward accepts service ID option; existing callers unaffected; fake compiles

---

## Phase 3: User Story 1 - SSH Tunnel in One Call (Priority: P1) 🎯 MVP

**Goal**: A single Tunnel() call that combines CreateSession + ForwardTcp(SshRelayTarget) with auto-cleanup.

**Independent Test**: Call Tunnel with a sandbox name and port, verify stream works, verify session revoked on Close.

### Implementation for User Story 1

- [X] T010 [US1] Update newSSHClient to accept SandboxInterface parameter (conn is already passed); update client.go constructor to pass c.sandboxes to newSSHClient in openshell/v1/ssh_client.go and openshell/v1/client.go
- [X] T011 [US1] Implement sshTunnel struct and sshClient.Tunnel method in openshell/v1/ssh_client.go. The sshTunnel struct wraps tcpForwardConn and adds session revocation on Close via sync.Once. The Tunnel method: validates port (1-65535), resolves sandbox name via sandboxes.Get() (returns *Sandbox with ID field), calls CreateSession, calls t.client.ForwardTcp() gRPC method directly (not public Forward()) to build TcpForwardInit with SshRelayTarget + authorization_token + optional service_id, uses defer-based cleanup to revoke session on failure, returns sshTunnel wrapping the resulting tcpForwardConn.
- [X] T012 [US1] Add tests for Tunnel: success path, invalid port rejection, sandbox not found, session leak prevention on forward failure, double-close safety, context cancellation cleanup in openshell/v1/ssh_client_test.go

**Checkpoint**: SSH Tunnel works end-to-end; session always cleaned up; all edge cases tested

---

## Phase 4: User Story 3 - Fake Client Parity (Priority: P3)

**Goal**: Fake SSH client implements Tunnel with port validation before Unimplemented error.

**Independent Test**: Call fake Tunnel with valid/invalid ports, verify correct error types.

### Implementation for User Story 3

- [X] T013 [US3] Add Tunnel method to fakeSSHClient with port validation (per FR-010, Constitution XI) in openshell/v1/fake/ssh.go
- [X] T014 [US3] Add tests for fake Tunnel: Unimplemented on valid input, InvalidArgument on bad port, Unavailable when closed in openshell/v1/fake/ssh_test.go

**Checkpoint**: Fake client compiles against updated SSHInterface; all interface contracts satisfied

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and final validation

- [X] T015 [P] Update doc.go package documentation with Tunnel and ForwardOption examples in openshell/v1/doc.go
- [X] T016 Run make ci to validate lint, build, and all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies, start immediately
- **Phase 2 (US2 - Forward Options)**: Depends on T002, T004 from Phase 1
- **Phase 3 (US1 - SSH Tunnel)**: Depends on T001, T003 from Phase 1; benefits from Phase 2 establishing the options pattern
- **Phase 4 (US3 - Fake Parity)**: Depends on T003 (SSHInterface with Tunnel method signature), T004 (TCPInterface with ForwardOption)
- **Phase 5 (Polish)**: Depends on all prior phases

### User Story Dependencies

- **US2 (Forward Options)**: Independent after Phase 1
- **US1 (SSH Tunnel)**: Independent after Phase 1; does not depend on US2 (tunnel builds its own init frame)
- **US3 (Fake Parity)**: Depends on interface changes from US1 and US2 (needs final signatures)

### Within Each User Story

- Interface types before implementation
- Implementation before tests (tests written alongside per Constitution III)
- Core logic before edge case handling

### Parallel Opportunities

- T001 and T002 can run in parallel (different files: ssh.go vs tcp.go)
- T008 and T009 can run in parallel (different files: fake/tcp.go vs fake/tcp_test.go)
- T013 and T014 could overlap (write fake, then immediately test)
- US1 and US2 can run in parallel after Phase 1

---

## Implementation Strategy

### MVP First (User Story 2 + User Story 1)

1. Complete Phase 1: Foundation (types and interfaces)
2. Complete Phase 2: US2 (Forward options, simpler change)
3. Complete Phase 3: US1 (SSH Tunnel, core value)
4. **STOP and VALIDATE**: `make ci` should pass
5. Complete Phase 4: US3 (Fake parity)
6. Complete Phase 5: Polish

### Incremental Delivery

1. Phase 1 + Phase 2 = Forward options working (audit trail enabled)
2. + Phase 3 = SSH Tunnel working (core value delivered)
3. + Phase 4 = Fake client parity (test support complete)
4. + Phase 5 = Documentation and CI clean

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Constitution III requires test-first: write failing tests, then implement
- Constitution XII governs Close order: protocol close before context cancel
- All new .go files need SPDX Apache-2.0 headers
- US2 is implemented before US1 despite lower priority because it establishes the options pattern and is simpler
