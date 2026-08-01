# Tasks: Global Policy Flag

**Input**: Design documents from `specs/024-global-policy-flag/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Included per Constitution III (Test-First).

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No setup required. Proto already has `global` field on both request messages.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend option config structs and add functional option constructors

- [x] T001 Add `global` field to `listPolicyConfig` and `getStatusConfig`, add `Global() bool` accessors, add `WithListGlobal(bool)` and `WithStatusGlobal(bool)` constructors in `openshell/v1/types/policy.go`
- [x] T002 Re-export `WithListGlobal` and `WithStatusGlobal` from `openshell/v1/policy.go`

**Checkpoint**: New option types compile. `make build` passes.

---

## Phase 3: User Story 1+2 - Global List and GetStatus (Priority: P1)

**Goal**: Platform admins can list global policy revisions and query global policy status.

**Independent Test**: List policies with global flag and verify results without sandbox name; query global status by version.

### Implementation for User Story 1+2

- [x] T003 [US1] Write tests for List with global flag (global returns results, global skips workspace validation, non-global preserves existing behavior) in `openshell/v1/policy_client_test.go`
- [x] T004 [US2] Write tests for GetStatus with global flag (global returns status, global skips name/workspace validation, composes with WithVersion) in `openshell/v1/policy_client_test.go`
- [x] T005 [US1] Wire `global` field into `ListSandboxPoliciesRequest` in `Policy().List()` in `openshell/v1/policy_client.go`. Note: the current List method has no client-side validation; when `global=true`, workspace is ignored by the gateway so no validation bypass is needed client-side.
- [x] T006 [US2] Wire `global` field into `GetSandboxPolicyStatusRequest` in `Policy().GetStatus()` in `openshell/v1/policy_client.go`. Note: the current GetStatus method has no client-side validation; when `global=true`, name and workspace are ignored by the gateway so no validation bypass is needed client-side.

**Checkpoint**: Real client global mode works. `make test` passes.

---

## Phase 4: User Story 3 - Fake Client (Priority: P2)

**Goal**: Fake client supports global policy mode for List and GetStatus with in-memory storage.

**Independent Test**: Use fake client to store and retrieve global vs sandbox-scoped revisions independently.

### Implementation for User Story 3

- [x] T007 [US3] Write fake policy tests for global List and GetStatus (global isolation, sandbox isolation, closed client) in `openshell/v1/fake/policy_test.go`
- [x] T008 [US3] Implement real in-memory List and GetStatus in fake policy client (replacing Unimplemented stubs) in `openshell/v1/fake/policy.go`. Add `globalRevisions []types.SandboxPolicyRevision` and `sandboxRevisions map[string][]types.SandboxPolicyRevision` (keyed by "workspace/name") fields to `fakePolicyClient`. Add public `AddRevision(workspace, name string, rev types.SandboxPolicyRevision)` and `AddGlobalRevision(rev types.SandboxPolicyRevision)` methods for test seeding.

**Checkpoint**: Fake client passes all tests. `make test` passes.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and final validation

- [x] T009 [P] Update `README.md` (project root) with global policy feature entry
- [x] T010 [P] Add global policy example to doc.go in `openshell/v1/doc.go`
- [x] T011 Run `make ci` (lint + build + test) to validate everything passes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 2)**: BLOCKS all user stories. Option types must exist first.
- **US1+2 (Phase 3)**: Depends on Phase 2. Real client wiring.
- **US3 (Phase 4)**: Depends on Phase 2. Can run in parallel with Phase 3.
- **Polish (Phase 5)**: Depends on all user stories being complete.

### Parallel Opportunities

- T003 and T004 can run in parallel (different test sections, same file but non-overlapping)
- T005 and T006 can run in parallel (different methods in same file)
- T009 and T010 can run in parallel (different files)
- Phase 4 (fake) can start after Phase 2 (does not need real client)

---

## Implementation Strategy

### MVP First (User Stories 1+2)

1. Complete Phase 2: Option types
2. Complete Phase 3: Wire global into real client
3. **STOP and VALIDATE**: `make test` passes
4. Continue to Phase 4-5

---

## Notes

- Small feature: 11 tasks, ~200 lines of new code
- No new files created, only existing files modified (except possibly new test file for fake)
- Fake policy client currently returns Unimplemented for all methods; List and GetStatus get real implementations, rest stays as stubs
- Backward compatible: global defaults to false, all existing tests pass unchanged
