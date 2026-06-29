# Tasks: SSH, TCP Forwarding, and Config (Phase 2b-1)

**Input**: Design documents from `specs/007-ssh-tcp-config/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Organization**: Tasks are grouped by user story. Tests are written first per Constitution Principle III (Test-First).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Domain Types)

**Purpose**: Define all new domain types in v1/types/ before any client code

- [X] T001 [P] Define SSHSession type in `openshell/v1/types/ssh.go` (SandboxID, Token, GatewayHost, GatewayPort, GatewayScheme, HostKeyFingerprint, ExpiresAtMs)
- [X] T002 [P] Define config types in `openshell/v1/types/setting.go` (SettingValueType string enum, SettingValue struct with typed fields, SettingScope string enum, PolicySource string enum, EffectiveSetting, SandboxConfig, GatewayConfig, ConfigUpdate, ConfigUpdateResult). Note: `types/config.go` is taken by the existing client Config struct. UpdateResult renamed to ConfigUpdateResult to avoid collision with profile.UpdateResult.

---

## Phase 2: Foundational (Converters)

**Purpose**: Proto↔SDK converters for all new types. Must complete before client implementations.

- [X] T003 [P] Test SSH converter (CreateSshSessionResponse → SSHSession, field mapping, deep copy) in `openshell/v1/internal/converter/ssh_test.go`
- [X] T004 [P] Test config converters (GetSandboxConfigResponse → SandboxConfig, GetGatewayConfigResponse → GatewayConfig, ConfigUpdate → UpdateConfigRequest, UpdateConfigResponse → UpdateResult, SettingValue oneof mapping, EffectiveSetting, SettingScope/PolicySource enum mapping) in `openshell/v1/internal/converter/setting_test.go`
- [X] T005 [P] Implement SSH converter in `openshell/v1/internal/converter/ssh.go`
- [X] T006 [P] Implement config converters in `openshell/v1/internal/converter/setting.go`

**Checkpoint**: Run `make test` to verify all converter tests pass

---

## Phase 3: User Story 1 — SSH Session Management (Priority: P1)

**Goal**: Developers can create and revoke SSH sessions for sandboxes

**Independent Test**: CreateSession → verify session fields → RevokeSession → verify revoked=true → Revoke again → verify revoked=false

- [X] T007 [P] [US1] Define SSHInterface (CreateSession, RevokeSession) in `openshell/v1/ssh.go`
- [X] T008 [P] [US1] Test sshClient CreateSession and RevokeSession against mock gRPC server in `openshell/v1/ssh_client_test.go`
- [X] T009 [US1] Implement sshClient using gRPC RPCs (CreateSshSession, RevokeSshSession) in `openshell/v1/ssh_client.go`

---

## Phase 4: User Story 2 — TCP Port Forwarding (Priority: P1)

**Goal**: Developers can forward TCP ports from sandboxes as io.ReadWriteCloser

**Independent Test**: Forward to sandbox port → verify ReadWriteCloser returned → write/read bytes → Close → verify subsequent ops fail

- [X] T010 [P] [US2] Define TCPInterface (Forward) in `openshell/v1/tcp.go`
- [X] T011 [P] [US2] Test tcpClient Forward against mock bidirectional gRPC stream in `openshell/v1/tcp_client_test.go` (init frame construction, read/write data frames, close terminates stream, port validation 1-65535, context cancellation)
- [X] T012 [US2] Implement tcpClient with tcpForwardConn wrapper (sends TcpForwardInit with TcpRelayTarget, wraps data frames as Read/Write, Close cancels stream context) in `openshell/v1/tcp_client.go`

---

## Phase 5: User Story 3 — Gateway and Sandbox Configuration (Priority: P2)

**Goal**: Operators can read sandbox config, read gateway config, and update configuration

**Independent Test**: GetSandbox → verify settings and policy returned → GetGateway → verify global settings → Update setting → verify new revision

- [X] T013 [P] [US3] Define ConfigInterface (GetSandbox, GetGateway, Update) in `openshell/v1/config.go`
- [X] T014 [P] [US3] Test configClient GetSandbox, GetGateway, Update against mock gRPC server in `openshell/v1/config_client_test.go` (settings map deep copy, opaque policy bytes, optimistic concurrency with expected_resource_version, global vs sandbox scope)
- [X] T015 [US3] Implement configClient using gRPC RPCs (GetSandboxConfig, GetGatewayConfig, UpdateConfig) in `openshell/v1/config_client.go`

---

## Phase 6: Interface Wiring

**Purpose**: Wire new sub-clients into ClientInterface and Client constructor

- [X] T016 Add SSH(), TCP(), Config() to ClientInterface in `openshell/v1/client.go` and wire sshClient, tcpClient, configClient in NewClient constructor
- [X] T017 Re-export new types from `openshell/v1/` via type aliases (SSHSession, SSHInterface, TCPInterface, ConfigInterface, SandboxConfig, GatewayConfig, ConfigUpdate, UpdateResult, SettingValue, SettingValueType, EffectiveSetting, SettingScope, PolicySource)

---

## Phase 7: User Story 4 — Fake Client Stubs (Priority: P3)

**Goal**: FakeClient compiles with updated interfaces, new stubs return Unimplemented

- [X] T018 [P] [US4] Test fakeSSHClient stubs (CreateSession, RevokeSession return Unimplemented) in `openshell/v1/fake/ssh_test.go`
- [X] T019 [P] [US4] Test fakeTCPClient stubs (Forward returns Unimplemented) in `openshell/v1/fake/tcp_test.go`
- [X] T020 [P] [US4] Test fakeConfigClient stubs (GetSandbox, GetGateway, Update return Unimplemented) in `openshell/v1/fake/config_test.go`
- [X] T021 [P] [US4] Implement fakeSSHClient stub in `openshell/v1/fake/ssh.go`
- [X] T022 [P] [US4] Implement fakeTCPClient stub in `openshell/v1/fake/tcp.go`
- [X] T023 [P] [US4] Implement fakeConfigClient stub in `openshell/v1/fake/config.go`
- [X] T024 [US4] Wire SSH(), TCP(), Config() on FakeClient in `openshell/v1/fake/fake.go`

---

## Phase 8: Polish & Cross-Cutting

**Purpose**: Documentation, type re-exports, lint compliance, CI verification

- [X] T025 [P] Add doc.go examples showing SSH().CreateSession, TCP().Forward, and Config().GetSandbox usage in `openshell/v1/doc.go`
- [X] T026 Verify all `.go` files in new paths have SPDX license headers
- [X] T027 Run `make ci` (lint + build + test) and fix any violations

---

## Dependencies

```
T001-T002 (types) ─────────┐
                            ├─► T003-T006 (converters)
                            │         │
                            │   T007-T009 (US1: SSH)
                            │   T010-T012 (US2: TCP)     ◄── parallel with US1
                            │   T013-T015 (US3: config)  ◄── parallel with US1/US2
                            │         │
                            │   T016-T017 (interface wiring)
                            │         │
                            │   T018-T024 (US4: fake stubs)
                            │         │
                            └─► T025-T027 (polish)
```

## Parallel Execution Opportunities

- **Phase 1**: T001, T002 parallel (different type files)
- **Phase 2**: T003-T004 parallel (test files), T005-T006 parallel (impl files)
- **Phase 3-5**: US1, US2, US3 can run in parallel after converters complete
- **Phase 7**: T018-T023 all parallel (independent stub files)

## Implementation Strategy

**MVP**: Phases 1-3 (types + converters + SSH) — delivers the most-requested connectivity capability.

**Full Scope**: All 8 phases, 27 tasks total.
- Types: 2 tasks
- Converters: 4 tasks
- US1 (SSH): 3 tasks
- US2 (TCP): 3 tasks
- US3 (Config): 3 tasks
- Interface wiring: 2 tasks
- US4 (Fake stubs): 7 tasks
- Polish: 3 tasks
