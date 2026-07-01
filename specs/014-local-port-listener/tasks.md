# Tasks: Local Port Listener

**Input**: Design documents from `specs/014-local-port-listener/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included (Constitution III: Test-First is non-negotiable).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Foundational (Interface + Options)

**Purpose**: Extend TCPInterface and define ListenOption types that all user stories depend on.

- [x] T001 Add `ListenOption` type, `listenConfig` struct, `WithBindAddress()`, `WithSSHTunnel()`, and `WithListenServiceID()` option constructors in `openshell/v1/tcp.go`
- [x] T002 Add `Listen` method signature to `TCPInterface` in `openshell/v1/tcp.go` with doc comment listing error codes (InvalidArgument, Unimplemented, Unavailable)

**Checkpoint**: Interface compiles; `make build` passes. Fake will fail compile-time check (expected, fixed in T010).

---

## Phase 2: User Story 1 - Bind a Local Port to a Sandbox Port (Priority: P1)

**Goal**: SDK consumer calls `Listen()` and gets a `net.Listener` that tunnels each accepted connection to a sandbox port via `Forward()`.

**Independent Test**: Create a listener, connect to it, verify data flows through the tunnel mock.

- [x] T003 [US1] Add test for `Listen()` returning a valid `net.Listener` bound to the specified local port in `openshell/v1/tcp_client_test.go`
- [x] T004 [US1] Implement `tunnelListener` struct with `inner net.Listener`, `wg sync.WaitGroup`, `closeOnce sync.Once`, context fields, and config in `openshell/v1/tcp_client.go`
- [x] T005 [US1] Implement `Listen()` on `tcpClient`: validate inputs (sandboxName non-empty, remotePort 1-65535), call `net.Listen("tcp", addr)`, return `tunnelListener` in `openshell/v1/tcp_client.go`
- [x] T006 [US1] Implement `Accept()` on `tunnelListener`: accept local conn, call `Forward()` (or `Tunnel()` if SSH mode), spawn bridge goroutines with `io.Copy`, handle tunnel setup failures by closing local conn and retrying in `openshell/v1/tcp_client.go`
- [x] T007 [US1] Implement `Addr()` on `tunnelListener` returning `inner.Addr()` in `openshell/v1/tcp_client.go`
- [x] T008 [US1] Add test for bidirectional data flow: connect to listener, write data, verify it reaches the Forward mock and vice versa in `openshell/v1/tcp_client_test.go`
- [x] T009 [US1] Add test for multiple concurrent connections: 10+ goroutines connect simultaneously (per SC-002), each transfers data independently in `openshell/v1/tcp_client_test.go`

**Checkpoint**: `make test` passes; listener accepts connections and tunnels data.

---

## Phase 3: User Story 2 - OS-Assigned Local Port (Priority: P1)

**Goal**: Passing localPort=0 binds to an ephemeral port; the consumer discovers the assigned port via `Addr()`.

**Independent Test**: Create listener with port 0, read assigned address, connect to it.

- [x] T010 [P] [US2] Add test for `Listen()` with localPort=0: verify `Addr()` returns a non-zero port and connections succeed in `openshell/v1/tcp_client_test.go`

**Checkpoint**: Ephemeral port test passes (implementation covered by T005 net.Listen behavior).

---

## Phase 4: User Story 3 - Graceful Shutdown (Priority: P1)

**Goal**: `Close()` stops accepting, tears down all active connections, and blocks until all bridge goroutines finish. Context cancellation triggers the same behavior.

**Independent Test**: Create listener, establish connections, close listener, verify all connections terminated and no goroutine leaks.

- [x] T011 [US3] Add test for `Close()`: establish 3 connections, call Close with a 5-second test deadline (per SC-003), verify all connections return errors on read and WaitGroup drains with zero goroutine leaks in `openshell/v1/tcp_client_test.go`
- [x] T012 [US3] Implement `Close()` on `tunnelListener`: close inner listener, cancel context, wait on WaitGroup, guard with sync.Once in `openshell/v1/tcp_client.go`
- [x] T013 [US3] Add test for context cancellation: pass a cancellable context to Listen, cancel it, verify listener closes and all connections terminate in `openshell/v1/tcp_client_test.go`
- [x] T014 [US3] Add context-watcher goroutine in Listen that calls Close when parent context is cancelled in `openshell/v1/tcp_client.go`
- [x] T015 [US3] Add test that Accept on a closed listener returns an error in `openshell/v1/tcp_client_test.go`

**Checkpoint**: `make test` passes; no goroutine leaks on close or context cancellation.

---

## Phase 5: User Story 4 - Custom Bind Address (Priority: P2)

**Goal**: `WithBindAddress("0.0.0.0")` overrides the default `127.0.0.1` binding.

**Independent Test**: Create listener with custom bind address, verify the bound address matches.

- [x] T016 [P] [US4] Add test for `WithBindAddress`: verify listener binds to the specified address instead of default 127.0.0.1 in `openshell/v1/tcp_client_test.go`

**Checkpoint**: Custom bind address test passes (implementation covered by T001 option + T005 address construction).

---

## Phase 6: User Story 5 - SSH Tunnel Transport (Priority: P2)

**Goal**: `WithSSHTunnel()` routes each accepted connection through `sshClient.Tunnel()` instead of `tcpClient.Forward()`.

**Independent Test**: Create listener with SSH tunnel option, verify Accept calls Tunnel instead of Forward.

- [x] T017 [US5] Add test for `WithSSHTunnel()`: verify accepted connections use SSH Tunnel path instead of TCP Forward in `openshell/v1/tcp_client_test.go`
- [x] T018 [US5] Wire SSH client into `tcpClient` (add `ssh SSHInterface` field) and pass it through `newTCPClient` in `openshell/v1/tcp_client.go` and `openshell/v1/client.go`. Update existing `newTCPClient` call sites in `openshell/v1/tcp_client_test.go` to pass nil for the SSH parameter
- [x] T019 [US5] Update `Accept()` bridge logic to call `ssh.Tunnel()` when `cfg.useSSHTunnel` is true in `openshell/v1/tcp_client.go`

**Checkpoint**: SSH tunnel transport test passes.

---

## Phase 7: User Story 6 - Fake Implementation (Priority: P2)

**Goal**: `fakeTCPClient.Listen()` validates inputs then returns Unimplemented.

**Independent Test**: Call Listen on fake, verify input validation and Unimplemented error.

- [x] T020 [P] [US6] Add test for fake Listen: verify empty sandboxName returns InvalidArgument, remotePort 0 returns InvalidArgument, valid inputs return Unimplemented in `openshell/v1/fake/tcp_test.go`
- [x] T021 [P] [US6] Add `Listen` method to `fakeTCPClient` with sandboxName and port validation, returning Unimplemented in `openshell/v1/fake/tcp.go`

**Checkpoint**: `make test` passes; fake satisfies compile-time interface check.

---

## Phase 8: Polish & Documentation

**Purpose**: Doc comments, examples, and final validation.

- [x] T022 Add doc.go example showing Listen usage with net.Listener pattern in `openshell/v1/example_test.go`
- [x] T023 Run `make ci` (lint + build + test) and fix any issues
- [x] T024 Verify all exported types and functions have doc comments per Constitution IX

**Checkpoint**: `make ci` passes clean; all doc comments present.

---

## Dependencies

```
T001, T002 (interface) ──→ T003-T009 (US1: core listener)
                        ──→ T010 (US2: ephemeral port) [P]
                        ──→ T020-T021 (US6: fake) [P]

T003-T009 (US1) ──→ T011-T015 (US3: shutdown)
                ──→ T016 (US4: bind address) [P]
                ──→ T017-T019 (US5: SSH tunnel)

All phases ──→ T022-T024 (polish)
```

## Parallel Execution Opportunities

After Phase 1 completes:
- US2 (T010), US4 (T016), US6 (T020-T021) can run in parallel with US1
- US3 and US5 depend on US1 completion

## Implementation Strategy

**MVP**: Phase 1 + Phase 2 (US1) delivers the core listener with TCP forward transport.
**Incremental**: Each subsequent phase adds one independently testable capability.
**Risk**: SSH tunnel wiring (T018) touches `client.go` constructor; keep the change minimal.
