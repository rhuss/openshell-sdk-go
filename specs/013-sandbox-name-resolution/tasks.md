# Tasks: Sandbox Name Resolution

**Input**: Design documents from `specs/013-sandbox-name-resolution/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Foundational (Interface + Constructor Wiring)

**Purpose**: Update interfaces and wire SandboxInterface dependency into
sub-client constructors. All user story tasks depend on this phase.

- [X] T001 Rename `sandboxID` to `sandboxName` in ExecInterface methods (Run, Stream, Interactive) and update doc comments in openshell/v1/exec.go
- [X] T002 [P] Rename `sandboxID` to `sandboxName` in TCPInterface.Forward and update doc comment in openshell/v1/tcp.go
- [X] T003 [P] Rename `sandboxID` to `sandboxName` in FileInterface methods (Upload, Download) and update doc comments in openshell/v1/file.go
- [X] T004 [P] Rename `sandboxID` to `sandboxName` in ConfigInterface.GetSandbox and update doc comment in openshell/v1/config.go
- [X] T005 Add `sandboxes SandboxInterface` field to execClient, tcpClient, fileClient, configClient structs and update their constructors to accept SandboxInterface in openshell/v1/exec_client.go, openshell/v1/tcp_client.go, openshell/v1/file_client.go, openshell/v1/config_client.go
- [X] T006 Wire SandboxInterface into newExecClient, newTCPClient, newFileClient, newConfigClient calls in openshell/v1/client.go

**Checkpoint**: Interfaces updated, constructors wired. Code will not compile yet (implementations still use old signatures).

---

## Phase 2: User Story 1 - Execute Command by Sandbox Name (Priority: P1) MVP

**Goal**: Exec.Run/Stream/Interactive accept sandbox name, resolve internally.

**Independent Test**: Call Exec().Run() with a sandbox name, verify proto request has resolved sandbox_id.

### Implementation for User Story 1

- [X] T007 [US1] Add name-to-ID resolution in execClient.Run, execClient.Stream, and execClient.Interactive: call s.sandboxes.Get(ctx, sandboxName), use sb.ID in proto request, in openshell/v1/exec_client.go
- [X] T008 [US1] Update fake execClient method signatures from sandboxID to sandboxName in openshell/v1/fake/exec.go
- [X] T009 [US1] Update exec_client_test.go test call sites and add test verifying name-to-ID resolution in Exec.Run in openshell/v1/exec_client_test.go
- [X] T010 [US1] Update fake exec test call sites in openshell/v1/fake/exec_test.go

**Checkpoint**: Exec sub-client fully name-based and tested.

---

## Phase 3: User Story 2 - Upload/Download File by Sandbox Name (Priority: P1)

**Goal**: Files.Upload/Download accept sandbox name, resolve internally.

**Independent Test**: Call Files().Upload() with a sandbox name, verify SSH session request has resolved sandbox_id.

### Implementation for User Story 2

- [X] T011 [P] [US2] Add name-to-ID resolution in fileClient.Upload and fileClient.Download: call f.sandboxes.Get(ctx, sandboxName), use sb.ID in proto request, in openshell/v1/file_client.go
- [X] T012 [P] [US2] Update fake fileClient method signatures from sandboxID to sandboxName in openshell/v1/fake/file.go
- [X] T013 [US2] Update file_client_test.go test call sites and add test verifying name-to-ID resolution in Files.Upload in openshell/v1/file_client_test.go
- [X] T014 [P] [US2] Update fake file test call sites in openshell/v1/fake/file_test.go

**Checkpoint**: File sub-client fully name-based and tested.

---

## Phase 4: User Story 3 - Watch Sandbox Events Correctly (Priority: P1)

**Goal**: Watch resolves name to ID before WatchSandboxRequest (bug fix).

**Independent Test**: Call Watch() with a sandbox name, verify WatchSandboxRequest.Id has the resolved ID, not the name.

### Implementation for User Story 3

- [X] T015 [US3] Add name-to-ID resolution in sandboxClient.Watch: call s.Get(ctx, name) before constructing WatchSandboxRequest, pass sb.ID to Id field, in openshell/v1/sandbox_client.go
- [X] T016 [US3] Add test verifying Watch sends resolved sandbox ID (not name) in WatchSandboxRequest.Id, and test NotFound error for non-existent sandbox, in openshell/v1/sandbox_client_test.go

**Checkpoint**: Watch bug fixed and tested.

---

## Phase 5: User Story 4 - Forward TCP by Sandbox Name (Priority: P2)

**Goal**: TCP.Forward accepts sandbox name, resolves internally.

**Independent Test**: Call TCP().Forward() with a sandbox name, verify TcpForwardInit has resolved sandbox_id.

### Implementation for User Story 4

- [X] T017 [P] [US4] Add name-to-ID resolution in tcpClient.Forward: call t.sandboxes.Get(ctx, sandboxName), use sb.ID in proto request, in openshell/v1/tcp_client.go
- [X] T018 [P] [US4] Update fake tcpClient method signature from sandboxID to sandboxName in openshell/v1/fake/tcp.go
- [X] T019 [US4] Update tcp_client_test.go test call sites and add test verifying name-to-ID resolution in TCP.Forward in openshell/v1/tcp_client_test.go
- [X] T020 [P] [US4] Update fake tcp test call sites in openshell/v1/fake/tcp_test.go

**Checkpoint**: TCP sub-client fully name-based and tested.

---

## Phase 6: User Story 5 - Get Sandbox Config by Name (Priority: P2)

**Goal**: Config.GetSandbox accepts sandbox name, resolves internally.

**Independent Test**: Call Config().GetSandbox() with a sandbox name, verify GetSandboxConfigRequest has resolved sandbox_id.

### Implementation for User Story 5

- [X] T021 [P] [US5] Add name-to-ID resolution in configClient.GetSandbox: call c.sandboxes.Get(ctx, sandboxName), use sb.ID in proto request, in openshell/v1/config_client.go
- [X] T022 [P] [US5] Update fake configClient.GetSandbox signature from sandboxID to sandboxName in openshell/v1/fake/config.go
- [X] T023 [US5] Update config_client_test.go test call sites and add test verifying name-to-ID resolution in Config.GetSandbox in openshell/v1/config_client_test.go
- [X] T024 [P] [US5] Update fake config test call sites in openshell/v1/fake/config_test.go

**Checkpoint**: Config sub-client fully name-based and tested.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, SSH.CreateSession doc update, final validation.

- [X] T025 Add doc comment to SSHInterface.CreateSession recommending SSH().Tunnel() for name-based access in openshell/v1/ssh.go
- [X] T026 Verify doc.go examples in openshell/v1/doc.go use sandbox names (already the case; no changes expected). Note: the CreateSession example at line 141 passes a name string to an ID-based method; add a clarifying comment there alongside the T025 doc update
- [X] T027 Run `make ci` to verify lint, build, and all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies, start immediately
- **Phases 2-6 (User Stories)**: All depend on Phase 1 completion
- **Phase 7 (Polish)**: Depends on all user stories complete

### User Story Dependencies

- **US1 (Exec)**: After Phase 1, no other story dependencies
- **US2 (Files)**: After Phase 1, no other story dependencies
- **US3 (Watch)**: After Phase 1, no other story dependencies (different file from US1/US2)
- **US4 (TCP)**: After Phase 1, no other story dependencies
- **US5 (Config)**: After Phase 1, no other story dependencies

### Parallel Opportunities

- T002, T003, T004 can run in parallel (different interface files)
- After Phase 1, all five user story phases can run in parallel (different files)
- Within each user story, implementation and fake updates marked [P] can run in parallel

---

## Parallel Example: After Phase 1

```bash
# All user stories can start simultaneously after Phase 1:
Task: "T007 [US1] Exec name resolution in exec_client.go"
Task: "T011 [US2] File name resolution in file_client.go"
Task: "T015 [US3] Watch name resolution in sandbox_client.go"
Task: "T017 [US4] TCP name resolution in tcp_client.go"
Task: "T021 [US5] Config name resolution in config_client.go"
```

---

## Implementation Strategy

### MVP First (User Story 1: Exec)

1. Complete Phase 1: Interface renames + constructor wiring
2. Complete Phase 2: Exec name resolution + tests
3. **STOP and VALIDATE**: `make test` passes, Exec methods accept names
4. Continue with remaining user stories

### Sequential Delivery (Recommended)

1. Phase 1: Foundational (T001-T006)
2. Phase 2: Exec (T007-T010) - MVP
3. Phase 3: Files (T011-T014)
4. Phase 4: Watch bug fix (T015-T016)
5. Phase 5: TCP (T017-T020)
6. Phase 6: Config (T021-T024)
7. Phase 7: Polish (T025-T027)

Each phase is independently testable via `make test`.

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- Commit after each phase completion
- `make ci` after each user story for confidence
- Total tasks: 27
- Parallel tasks: 15 (marked [P])
