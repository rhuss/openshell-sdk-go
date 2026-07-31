# Tasks: Workspace CRUD, GatewayInfo & GetCurrentUser

**Input**: Design documents from `specs/022-workspace-crud-gatewayinfo/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included (test-first is a non-negotiable constitution requirement).

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)

## Phase 1: Setup

**Purpose**: No project setup needed, SDK already exists. This phase is empty.

---

## Phase 2: Foundational (Types & Converters)

**Purpose**: SDK types and proto converters that all user stories depend on. MUST complete before any user story.

- [X] T001 [P] Add workspace types (Workspace, WorkspaceMember, WorkspacePhase, WorkspaceRole) in `openshell/v1/types/workspace.go`
- [X] T002 [P] Add GatewayInfo, ComputeDriverInfo, ServiceStatus, CurrentUser types in `openshell/v1/types/health.go`
- [X] T003 Write workspace and member converters (WorkspaceFromProto, WorkspaceMemberFromProto, WorkspacePhaseFromProto, WorkspaceRoleFromProto, WorkspaceRoleToProto) in `openshell/v1/internal/converter/workspace.go`
- [X] T004 [P] Write health converters (GatewayInfoFromProto, ServiceStatusFromProto, ComputeDriverInfoFromProto, CurrentUserFromProto) in `openshell/v1/internal/converter/health.go`
- [X] T005 Write converter round-trip tests for workspace types in `openshell/v1/internal/converter/workspace_test.go`
- [X] T006 [P] Write converter round-trip tests for health types in `openshell/v1/internal/converter/health_test.go`

**Checkpoint**: All SDK types and converters ready. User story implementation can begin.

---

## Phase 3: User Story 1 - Workspace Lifecycle Management (Priority: P1)

**Goal**: SDK consumers can create, get, list, and delete workspaces.

**Independent Test**: Create a workspace, verify it in list results, get by name, delete it.

- [X] T007 [US1] Define WorkspaceInterface (Create, Get, List, Delete, AddMember, RemoveMember, ListMembers) and type aliases in `openshell/v1/workspace.go`
- [X] T008 [US1] Implement workspaceClient Create method with input validation in `openshell/v1/workspace_client.go`
- [X] T009 [US1] Implement workspaceClient Get method with input validation in `openshell/v1/workspace_client.go`
- [X] T010 [US1] Implement workspaceClient List method with ListOptions support in `openshell/v1/workspace_client.go`
- [X] T011 [US1] Implement workspaceClient Delete method with input validation in `openshell/v1/workspace_client.go`
- [X] T012 [US1] Add Workspaces() accessor to ClientInterface and Client struct, wire in NewClient in `openshell/v1/client.go`
- [X] T013 [US1] Write tests for workspace CRUD operations (success + error paths) in `openshell/v1/workspace_test.go`

---

## Phase 4: User Story 2 - Workspace Member Management (Priority: P1)

**Goal**: Workspace admins can add, remove, and list members with roles.

**Independent Test**: Add a member to an existing workspace, list to verify, remove the member.

- [X] T014 [US2] Implement workspaceClient AddMember method with input validation (non-empty workspace, subject, valid role) in `openshell/v1/workspace_client.go`
- [X] T015 [US2] Implement workspaceClient RemoveMember method with input validation in `openshell/v1/workspace_client.go`
- [X] T016 [US2] Implement workspaceClient ListMembers method with ListOptions support in `openshell/v1/workspace_client.go`
- [X] T017 [US2] Write tests for member management operations (success + error paths) in `openshell/v1/workspace_test.go`

---

## Phase 5: User Stories 3 & 4 - Health Extension (GetGatewayInfo & GetCurrentUser)

**Goal**: Platform admins can query gateway metadata; SDK consumers can determine their identity.

**Independent Test**: Call GetGatewayInfo and verify response. Call GetCurrentUser and verify identity details.

**Note**: US3 and US4 are combined into one phase because both extend the same files (health.go, health_client.go, health_test.go). Parallel execution across these stories would cause file conflicts.

- [X] T018 [US3+US4] Add GetGatewayInfo and GetCurrentUser to HealthInterface in `openshell/v1/health.go`
- [X] T019 [US3+US4] Implement healthClient GetGatewayInfo and GetCurrentUser methods in `openshell/v1/health_client.go`
- [X] T020 [US3+US4] Write tests for GetGatewayInfo and GetCurrentUser (success + error paths) in `openshell/v1/health_test.go`. Error coverage: GetGatewayInfo_Error (PermissionDenied), GetCurrentUser_Unauthenticated.

---

## Phase 6: User Story 5 - Fake Client Support (Priority: P1)

**Goal**: Fake client implements all new operations with in-memory storage and matching validation.

**Independent Test**: Use fake client to create workspaces, add members, query gateway info.

**Note**: Workspaces are top-level resources (not workspace-scoped). The fake workspace store should use an empty string as the workspace key in `objectStore`, and use `ListAll()` for listing since workspaces span all namespaces. Members are workspace-scoped and use the workspace name as the store key normally.

- [X] T021 [US5] Add workspace objectStore and fakeWorkspaceClient with Create, Get, List, Delete in `openshell/v1/fake/workspace.go`
- [X] T022 [US5] Add member management (AddMember, RemoveMember, ListMembers) to fakeWorkspaceClient in `openshell/v1/fake/workspace.go`
- [X] T023 [US5] Add GetGatewayInfo and GetCurrentUser to fakeHealthClient in `openshell/v1/fake/health.go`
- [X] T024 [US5] Add Workspaces() accessor, workspaceStore field, WithGatewayInfo, WithCurrentUser, WithWorkspaces options to fake Client in `openshell/v1/fake/fake.go`
- [X] T025 [US5] Write fake workspace tests (CRUD + member + validation parity) in `openshell/v1/fake/workspace_test.go`
- [X] T026 [US5] Write fake health extension tests (GetGatewayInfo, GetCurrentUser) in `openshell/v1/fake/health_test.go`

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates and CI validation.

- [X] T027 [P] Add workspace and gateway examples to package doc in `openshell/v1/doc.go`
- [X] T028 [P] Update README.md feature list with workspace CRUD, GatewayInfo, GetCurrentUser
- [X] T029 Run `make ci` (lint + build + test) and fix any issues

---

## Dependencies

```text
T001-T002 (types) → T003-T004 (converters) → T007-T012 (workspace client)
                                             → T014-T016 (member client)
                                             → T018-T019 (health extension)
                                             → T021-T024 (fake client)

T005 depends on T003
T006 depends on T004
T013 depends on T007-T012
T017 depends on T014-T016
T020 depends on T018-T019
T025-T026 depends on T021-T024
T029 depends on all prior tasks
```

## Parallel Execution Opportunities

- **Phase 2**: T001 and T002 (types) are fully parallel (different files). T003 and T004 (converters) are parallel after types (different files). T005 and T006 (tests) parallel after their respective converters.
- **Phase 3-4**: Sequential (shared workspace_client.go file).
- **Phase 5**: Sequential (all tasks modify the same health files).
- **Phase 6**: T021-T022 sequential (same file), T023 parallel with T024 (different files). T025 and T026 parallel after implementation (different files).
- **Phase 7**: T027 and T028 are parallel. T029 is last.

## Implementation Strategy

**MVP**: Phase 2 + Phase 3 (workspace CRUD). Delivers core workspace management with tests.

**Incremental delivery**:
1. Types + Converters (Phase 2)
2. Workspace CRUD (Phase 3)
3. Member Management (Phase 4)
4. Health extensions (Phase 5)
5. Fake client (Phase 6)
6. Polish (Phase 7)
