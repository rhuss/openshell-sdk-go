# Tasks: Phase 2b-2 Policy, Logs, MergeOps, ErrorConflict

**Input**: Design documents from `specs/008-policy-logs/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Constitution III (Test-First) requires tests alongside implementation. Each task includes its test file.

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1-US7)

---

## Phase 1: Foundational (Domain Types, Converters, Error Code)

**Purpose**: All domain types, converters, and the ErrorConflict error code. MUST complete before any client implementation.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

### Error Code

- [X] T001 [P] Add ErrorConflict to ErrorCode enum, IsConflict helper, and String() method in openshell/v1/types/errors.go
- [X] T002 [P] Add ErrorConflict tests (IsConflict true/false/nil) in openshell/v1/errors_test.go
- [X] T003 [P] Map codes.Aborted to ErrorConflict in openshell/v1/internal/converter/errors.go and add test in errors_test.go

### Network Policy Types + Converters

- [X] T004 [P] Create NetworkPolicyRule, NetworkEndpoint, NetworkBinary, L7Rule, L7Allow, L7DenyRule, L7QueryMatcher, GraphqlOperation types in openshell/v1/types/network_policy.go
- [X] T005 [P] Create NetworkPolicyRule converters (FromProto + ToProto for all nested types) in openshell/v1/internal/converter/network_policy.go
- [X] T006 Create NetworkPolicyRule converter round-trip tests in openshell/v1/internal/converter/network_policy_test.go

### Policy Types + Converters

- [X] T007 [P] Create PolicyChunk, DraftPolicy, PolicyStatusResult, SandboxPolicyRevision, PolicyLoadStatus, ApproveResult, ApproveAllResult, UndoResult, ClearResult, DraftHistoryEntry types and functional option types (GetDraftOption, ApproveAllOption, GetStatusOption, ListPolicyOption) in openshell/v1/types/policy.go
- [X] T008 [P] Create policy converters (PolicyChunkFromProto, DraftPolicyFromProto, PolicyStatusResultFromProto, SandboxPolicyRevisionFromProto, DraftHistoryEntryFromProto, ApproveResultFromProto, ApproveAllResultFromProto, UndoResultFromProto, ClearResultFromProto) in openshell/v1/internal/converter/policy.go
- [X] T009 Create policy converter round-trip tests in openshell/v1/internal/converter/policy_test.go

### Log Types + Converters

- [X] T010 [P] Create LogLine, LogResult, LogOption types (WithLogLines, WithLogSince, WithLogSources, WithLogMinLevel) in openshell/v1/types/log.go
- [X] T011 [P] Create log converters (LogLineFromProto, LogResultFromProto) in openshell/v1/internal/converter/log.go
- [X] T012 Create log converter tests in openshell/v1/internal/converter/log_test.go

### MergeOperation Types + Converter Update

- [X] T013 Create PolicyMergeOperation, AddNetworkRule, RemoveNetworkEndpoint, RemoveNetworkRule, AddDenyRules, AddAllowRules, RemoveNetworkBinary types in openshell/v1/types/network_policy.go (append to existing file from T004; depends on T004)
- [X] T014 Change ConfigUpdate.MergeOperations from []byte to []PolicyMergeOperation in openshell/v1/types/setting.go
- [X] T015 Add PolicyMergeOperationToProto converter and update ConfigUpdateToProto to serialize MergeOperations in openshell/v1/internal/converter/setting.go
- [X] T016 Add MergeOperation converter tests and update existing MergeOperations test in openshell/v1/internal/converter/setting_test.go

### Type Re-exports

- [X] T017 Add ErrorConflict and IsConflict re-exports in openshell/v1/errors.go. Add policy type re-exports (PolicyChunk, DraftPolicy, PolicyStatusResult, SandboxPolicyRevision, PolicyLoadStatus, ApproveResult, ApproveAllResult, UndoResult, ClearResult, DraftHistoryEntry, and option types) in openshell/v1/policy.go alongside the interface. Add log type re-exports (LogLine, LogResult, LogOption) in openshell/v1/sandbox.go alongside GetLogs. Add network policy type re-exports (NetworkPolicyRule, NetworkEndpoint, NetworkBinary, L7Rule, L7Allow, L7DenyRule, L7QueryMatcher, GraphqlOperation, PolicyMergeOperation, AddNetworkRule, RemoveNetworkEndpoint, RemoveNetworkRule, AddDenyRules, AddAllowRules, RemoveNetworkBinary) in openshell/v1/types_reexport.go

**Checkpoint**: All types, converters, and error code ready. Client implementation can begin.

---

## Phase 2: User Story 1 — Operator Reviews and Approves Draft Policy Chunks (Priority: P1) 🎯 MVP

**Goal**: Core draft-review-approve workflow: GetDraft, ApproveDraftChunk, RejectDraftChunk

**Independent Test**: Get draft policy, approve a chunk (verify policy version returned), reject a chunk with reason. Test error paths (NotFound, invalid chunk ID).

- [X] T018 Define PolicyInterface with all 10 methods in openshell/v1/policy.go
- [X] T019 [US1] Implement policyClient struct, newPolicyClient, GetDraft (with WithStatusFilter option), ApproveDraftChunk, and RejectDraftChunk in openshell/v1/policy_client.go
- [X] T020 [US1] Add tests for GetDraft, ApproveDraftChunk, RejectDraftChunk (success + error paths, status filter option) in openshell/v1/policy_client_test.go

**Checkpoint**: Core approve/reject workflow testable.

---

## Phase 3: User Story 2 — Operator Bulk-Approves or Clears Draft Chunks (Priority: P1)

**Goal**: ApproveAllDraftChunks (with WithIncludeSecurityFlagged), ClearDraftChunks, GetDraftHistory

**Independent Test**: Approve all (verify counts), approve all with security-flagged, clear chunks, get history entries.

- [X] T021 [US2] Implement ApproveAllDraftChunks (with WithIncludeSecurityFlagged), ClearDraftChunks, GetDraftHistory in openshell/v1/policy_client.go
- [X] T022 [US2] Add tests for ApproveAllDraftChunks (default + security-flagged), ClearDraftChunks, GetDraftHistory in openshell/v1/policy_client_test.go

**Checkpoint**: Full draft management workflow (approve, reject, bulk approve, clear, history) testable.

---

## Phase 4: User Story 3 — Operator Inspects Policy Status and History (Priority: P2)

**Goal**: GetStatus (with WithVersion), List (with WithLimit/WithOffset), EditDraftChunk, UndoDraftChunk

**Independent Test**: Get status (verify revision + active version), list revisions, edit a chunk, undo an approved chunk.

- [X] T023 [US3] Implement GetStatus (with WithVersion), List (with WithLimit/WithOffset), EditDraftChunk, UndoDraftChunk in openshell/v1/policy_client.go
- [X] T024 [US3] Add tests for GetStatus, List (pagination options), EditDraftChunk, UndoDraftChunk in openshell/v1/policy_client_test.go

**Checkpoint**: All 10 PolicyInterface methods implemented and tested.

---

## Phase 5: User Story 4 — Developer Retrieves Sandbox Logs (Priority: P2)

**Goal**: GetLogs on SandboxInterface with name→id resolution and functional options

**Independent Test**: Get logs for a sandbox (verify lines + buffer total), test filtering options, test name→id resolution, test NotFound error.

- [X] T025 [US4] Add GetLogs to SandboxInterface definition in openshell/v1/sandbox.go
- [X] T026 [US4] Implement GetLogs on sandboxClient (name→id resolution via Get, functional options to proto request) in openshell/v1/sandbox_client.go
- [X] T027 [US4] Add tests for GetLogs (success, filtering options, name resolution, not-found) in openshell/v1/sandbox_client_test.go

**Checkpoint**: GetLogs testable on SandboxInterface.

---

## Phase 6: User Story 5 + 6 — Typed MergeOperations + ErrorConflict Wiring (Priority: P2/P3)

**Goal**: Remove MergeOperations runtime rejection from config client, wire typed converter. Verify ErrorConflict surfaces from Config().Update with stale resource version.

**Independent Test**: Config().Update with MergeOperations succeeds (no longer rejected). Config().Update with stale expected_resource_version returns IsConflict=true.

- [X] T028 [P] [US5] Remove MergeOperations runtime rejection and wire typed converter in openshell/v1/config_client.go
- [X] T029 [P] [US6] Update config_client_test.go: change MergeOperationsRejected test to MergeOperationsAccepted, add ErrorConflict test (codes.Aborted → IsConflict) in openshell/v1/config_client_test.go

**Checkpoint**: MergeOperations accepted, ErrorConflict detectable via IsConflict().

---

## Phase 7: User Story 7 — Fake Client Stubs (Priority: P3)

**Goal**: Fake PolicyClient, fake GetLogs, fake Config MergeOps acceptance, ClientInterface wiring

**Independent Test**: Fake Policy() methods return Unimplemented. Fake GetLogs returns Unimplemented. Fake Config().Update with MergeOps returns no error. Compile-time interface check passes.

- [X] T030 [P] [US7] Create fake PolicyClient (all 10 methods returning Unimplemented) in openshell/v1/fake/policy.go
- [X] T031 [P] [US7] Add fake PolicyClient tests in openshell/v1/fake/policy_test.go
- [X] T032 [US7] Add GetLogs stub to fake SandboxClient in openshell/v1/fake/sandbox.go and test in fake/sandbox_test.go
- [X] T033 [US7] Remove MergeOperations rejection from fake ConfigClient in openshell/v1/fake/config.go and update test in fake/config_test.go
- [X] T034 [US7] Add Policy() accessor and policy field to fake.Client, update compile-time interface check in openshell/v1/fake/fake.go
- [X] T035 [US7] Add Policy() accessor and policy field to Client, wire newPolicyClient in openshell/v1/client.go

**Checkpoint**: Full ClientInterface compiles with Policy(), GetLogs, typed MergeOps, ErrorConflict.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Doc comments, re-exports, CI validation

- [X] T036 [P] Add doc comments to PolicyInterface methods listing error codes (NotFound, Unimplemented, InvalidArgument, Conflict) in openshell/v1/policy.go
- [X] T037 [P] Add doc comments to GetLogs method on SandboxInterface in openshell/v1/sandbox.go
- [X] T038 Run mise run ci to verify lint + build + test pass with zero violations
- [X] T039 Run go test -race ./... to verify concurrency safety

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Foundational)**: No dependencies — start immediately
- **Phase 2 (US1)**: Depends on Phase 1 (types + converters)
- **Phase 3 (US2)**: Depends on Phase 2 (policyClient struct exists)
- **Phase 4 (US3)**: Depends on Phase 2 (policyClient struct exists), can run parallel with Phase 3
- **Phase 5 (US4)**: Depends on Phase 1 (log types + converters), independent of policy phases
- **Phase 6 (US5+US6)**: Depends on Phase 1 (merge types + error code)
- **Phase 7 (US7)**: Depends on Phases 2-6 (all interfaces finalized)
- **Phase 8 (Polish)**: Depends on Phase 7

### Parallel Opportunities

After Phase 1 completes:
- **Phase 2 + Phase 5 + Phase 6** can run in parallel (different files, no dependencies)
- **Phase 3 + Phase 4** can run in parallel after Phase 2 (both add methods to policy_client.go, but to different sections)

### Within Phase 1 (Maximum Parallelism)

```
T001 + T002 + T003  (error code, all [P])
T004 + T007 + T010  (types files, all [P])
T005 + T008 + T011  (converter files, all [P])
— then —
T006, T009, T012    (converter tests, depend on their converter)
T013                (merge types, append to T004's file)
T014                (setting.go change)
T015 + T016         (setting converter + test)
T017                (re-exports)
```

---

## Implementation Strategy

### MVP First (User Stories 1+2)

1. Complete Phase 1: Foundational types + converters + error code
2. Complete Phase 2: Core draft workflow (GetDraft, Approve, Reject)
3. Complete Phase 3: Bulk operations (ApproveAll, Clear, History)
4. **STOP and VALIDATE**: Core policy management testable

### Full Delivery

5. Complete Phase 4: Status + List + Edit + Undo
6. Complete Phase 5: GetLogs
7. Complete Phase 6: MergeOperations + ErrorConflict
8. Complete Phase 7: Fakes + ClientInterface
9. Complete Phase 8: Polish + CI

---

## Notes

- All tasks include test files (Constitution III: Test-First)
- Deep copy at boundaries in all converters (Constitution VII)
- SPDX Apache-2.0 headers on all new files
- Proto imports: `pb "...openshellv1"`, `sbv1 "...sandboxv1"`
- Existing TimeFromMillis/MillisFromTime helpers used for all timestamp conversions
- Existing CopyStringSlice/CopyStringMap helpers used for deep copy
