# Tasks: Add Workspace Scoping to All RPCs

**Input**: Design documents from `specs/021-workspace-scoping/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/interfaces.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Add shared types and infrastructure needed by all user stories

- [x] T001 Add `AllWorkspaces bool` field to `ListOptions` in `openshell/v1/types/options.go`
- [x] T002 Add `Workspace string` field to exported types that return resources (Sandbox, Provider, etc.) by updating converter functions to populate workspace from proto `ObjectMeta.workspace` in `openshell/v1/internal/converter/sandbox.go`, `converter/provider.go`, `converter/profile.go`, `converter/service.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Update the fake store to support workspace-scoped composite keys. MUST complete before any user story work.

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T003 Refactor `objectStore[T]` in `openshell/v1/fake/store.go` to use composite keys (`workspace/name`) for workspace isolation. Add a `listAll()` method that returns items across all workspaces. Update `keyFunc` signature to accept workspace context. Update `openshell/v1/fake/store_test.go` with workspace isolation tests.

**Checkpoint**: Foundation ready, user story implementation can begin

---

## Phase 3: User Story 1 - Workspace-scoped sandbox operations (Priority: P1) MVP

**Goal**: Every sandbox CRUD method accepts workspace as first param after ctx. Fake client isolates sandboxes per workspace.

**Independent Test**: Create sandboxes in different workspaces, verify listing in one workspace does not return sandboxes from another.

### Implementation for User Story 1

- [ ] T004 [US1] Update `SandboxInterface` in `openshell/v1/sandbox.go` to add `workspace string` param to all methods: Create, Get, List, Delete, AttachProvider, DetachProvider, ListProviders, WaitReady, Watch, GetLogs
- [ ] T005 [US1] Update `sandboxClient` in `openshell/v1/sandbox_client.go` to set `Workspace` field on all proto request messages (CreateSandboxRequest, GetSandboxRequest, ListSandboxesRequest, DeleteSandboxRequest, etc.) and set `AllWorkspaces` on ListSandboxesRequest from ListOptions
- [ ] T006 [US1] Update fake sandbox implementation in `openshell/v1/fake/sandbox.go` to use workspace-scoped store keys. All methods pass workspace to store operations. List respects AllWorkspaces.
- [ ] T007 [US1] Update `openshell/v1/sandbox_client_test.go` to pass workspace param to all sandbox method calls
- [ ] T008 [US1] Update `openshell/v1/fake/sandbox_test.go` to pass workspace param and add workspace isolation test (create in "ws-a", list in "ws-b" returns empty)

**Checkpoint**: Sandbox operations fully workspace-scoped and testable independently

---

## Phase 4: User Story 2 - Workspace-scoped provider and profile management (Priority: P1)

**Goal**: Provider, Profile, and Refresh methods accept workspace param. Fake client isolates per workspace.

**Independent Test**: Create providers in different workspaces, verify CRUD isolation.

### Implementation for User Story 2

- [ ] T009 [P] [US2] Update `ProviderInterface` in `openshell/v1/provider.go`, `providerClient` in `openshell/v1/provider_client.go`, and fake in `openshell/v1/fake/provider.go` to add workspace param. Fake List must respect `AllWorkspaces` in ListOptions by calling the store's `listAll()` when set. Update `openshell/v1/provider_client_test.go` and `openshell/v1/fake/provider_test.go`.
- [ ] T010 [P] [US2] Update `ProfileInterface` in `openshell/v1/profile.go`, `profileClient` in `openshell/v1/profile_client.go`, and fake in `openshell/v1/fake/profile.go` to add workspace param. Fake List must respect `AllWorkspaces` in ListOptions by calling the store's `listAll()` when set. Update `openshell/v1/profile_client_test.go` and `openshell/v1/fake/profile_test.go`.
- [ ] T011 [P] [US2] Update `RefreshInterface` in `openshell/v1/refresh.go`, `refreshClient` in `openshell/v1/refresh_client.go`, and fake in `openshell/v1/fake/refresh.go` to add workspace param. Update `openshell/v1/refresh_client_test.go` and `openshell/v1/fake/refresh_test.go`.

**Checkpoint**: Provider, Profile, and Refresh operations fully workspace-scoped

---

## Phase 5: User Story 3 - Cross-workspace listing for platform admins (Priority: P2)

**Goal**: When `AllWorkspaces: true` is set in ListOptions, List operations return resources from all workspaces. The workspace parameter is silently ignored.

**Independent Test**: Create resources in multiple workspaces, list with AllWorkspaces=true, verify all returned.

### Implementation for User Story 3

- [ ] T012 [US3] Add cross-workspace listing tests in `openshell/v1/fake/sandbox_test.go` (create sandboxes in "ws-a" and "ws-b", list with AllWorkspaces=true, verify both returned). Add similar tests in `openshell/v1/fake/provider_test.go`, `openshell/v1/fake/service_test.go`, `openshell/v1/fake/profile_test.go`, and `openshell/v1/fake/policy_test.go`. All five List operations that accept ListOptions must be tested for AllWorkspaces.
- [ ] T013 [US3] Verify `AllWorkspaces` is correctly mapped to proto `all_workspaces` field in `openshell/v1/sandbox_client.go` (ListSandboxesRequest), `openshell/v1/provider_client.go` (ListProvidersRequest), and `openshell/v1/service_client.go` (ListServicesRequest). Ensure workspace param is ignored when AllWorkspaces is set.

**Checkpoint**: Cross-workspace listing works for sandbox, provider, and service List operations

---

## Phase 6: User Story 4 - Workspace-scoped sandbox interactions (Priority: P2)

**Goal**: Service, Exec, File, Config, Policy, SSH, and TCP methods all accept workspace param.

**Independent Test**: Execute a command or upload a file to a sandbox in a specific workspace, verify proto request includes workspace.

### Implementation for User Story 4

- [ ] T014 [P] [US4] Update `ServiceInterface` in `openshell/v1/service.go`, `serviceClient` in `openshell/v1/service_client.go`, and fake in `openshell/v1/fake/service.go` to add workspace param. Fake List must respect `AllWorkspaces` in ListOptions by calling the store's `listAll()` when set. Update `openshell/v1/service_client_test.go` and `openshell/v1/fake/service_test.go`.
- [ ] T015 [P] [US4] Update `ExecInterface` in `openshell/v1/exec.go`, `execClient` in `openshell/v1/exec_client.go`, and fake in `openshell/v1/fake/exec.go` to add workspace param. Update `openshell/v1/exec_client_test.go` and `openshell/v1/fake/exec_test.go`.
- [ ] T016 [P] [US4] Update `FileInterface` in `openshell/v1/file.go`, `fileClient` in `openshell/v1/file_client.go`, and fake in `openshell/v1/fake/file.go` to add workspace param. Update `openshell/v1/file_client_test.go` and `openshell/v1/fake/file_test.go`.
- [ ] T017 [P] [US4] Update `ConfigInterface` in `openshell/v1/config.go` (GetSandbox and Update only, NOT GetGateway), `configClient` in `openshell/v1/config_client.go`, and fake in `openshell/v1/fake/config.go`. Update `openshell/v1/config_client_test.go` and `openshell/v1/fake/config_test.go`.
- [ ] T018 [P] [US4] Update `PolicyInterface` in `openshell/v1/policy.go`, `policyClient` in `openshell/v1/policy_client.go`, and fake in `openshell/v1/fake/policy.go` to add workspace param to all 10 methods. Fake List must respect `AllWorkspaces` in ListOptions by calling the store's `listAll()` when set. Update `openshell/v1/policy_client_test.go` and `openshell/v1/fake/policy_test.go`.
- [ ] T019 [P] [US4] Update `SSHInterface` in `openshell/v1/ssh.go`, `sshClient` in `openshell/v1/ssh_client.go`, and fake in `openshell/v1/fake/ssh.go` to add workspace param. Update `openshell/v1/ssh_client_test.go` and `openshell/v1/fake/ssh_test.go`.
- [ ] T020 [P] [US4] Update `TCPInterface` in `openshell/v1/tcp.go`, `tcpClient` in `openshell/v1/tcp_client.go`, and fake in `openshell/v1/fake/tcp.go` to add workspace param. Update `openshell/v1/tcp_client_test.go` and `openshell/v1/fake/tcp_test.go`.

**Checkpoint**: All sandbox interaction sub-clients fully workspace-scoped

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and final validation

- [ ] T021 [P] Update `openshell/v1/doc.go` and `README.md` with workspace-scoped usage examples per Constitution VIII and XIII and FR-010. Update package-level examples to show workspace parameter.
- [ ] T022 [P] Update `openshell/v1/example_test.go` and `openshell/v1/example_fake_test.go` to include workspace parameter in all example function calls
- [ ] T023 Run `make ci` (lint + build + test) to verify all tests pass and no lint violations exist

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (ListOptions needed for store refactor)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (needs workspace-aware store)
- **User Story 2 (Phase 4)**: Depends on Phase 2 (can run in parallel with US1)
- **User Story 3 (Phase 5)**: Depends on US1 and US2 (needs workspace-scoped resources to test cross-workspace listing)
- **User Story 4 (Phase 6)**: Depends on Phase 2 (can run in parallel with US1/US2)
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: After Phase 2, no other story dependencies
- **US2 (P1)**: After Phase 2, no other story dependencies. Can run in parallel with US1.
- **US3 (P2)**: After US1 and US2 (needs resources created in workspaces to verify cross-workspace listing)
- **US4 (P2)**: After Phase 2, no other story dependencies. Can run in parallel with US1/US2.

### Parallel Opportunities

- T009, T010, T011 within US2 can all run in parallel (different sub-clients)
- T014-T020 within US4 can all run in parallel (different sub-clients, different files)
- T021, T022 in Polish can run in parallel
- US1, US2, US4 can all run in parallel after Phase 2

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: Foundational (T003)
3. Complete Phase 3: User Story 1 (T004-T008)
4. **STOP and VALIDATE**: `make test` passes, sandbox workspace isolation works
5. Continue to remaining stories

### Incremental Delivery

1. Setup + Foundational (T001-T003) -> Foundation ready
2. US1: Sandbox scoping (T004-T008) -> MVP testable
3. US2: Provider/Profile/Refresh scoping (T009-T011) -> Core resources scoped
4. US3: Cross-workspace listing (T012-T013) -> Admin features
5. US4: Interaction sub-clients (T014-T020) -> Full coverage
6. Polish (T021-T023) -> Documentation and CI validation

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable after foundational phase
- Constitution III (Test-First) requires tests alongside implementation, bundled per task
- Commit after each task or logical group
- Total: 23 tasks across 7 phases
