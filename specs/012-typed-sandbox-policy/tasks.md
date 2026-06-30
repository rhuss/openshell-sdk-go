# Tasks: Typed SandboxPolicy Domain Type

**Input**: Design documents from `specs/012-typed-sandbox-policy/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Included (Constitution III: Test-First, FR-015 requires round-trip tests).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Foundational (Domain Types + Core Converters)

**Purpose**: New domain types and SandboxPolicy converter that ALL user stories depend on

**Warning**: No user story work can begin until this phase is complete

- [X] T001 [P] Add SandboxPolicy, FilesystemPolicy, LandlockPolicy, ProcessPolicy structs with doc comments in openshell/v1/types/policy.go
- [X] T002 [P] Add Policy *SandboxPolicy field to SandboxSpec in openshell/v1/types/sandbox.go
- [X] T003 [P] Change ConfigUpdate.Policy from []byte to *SandboxPolicy and SandboxConfig.Policy from []byte to *SandboxPolicy in openshell/v1/types/setting.go
- [X] T004 Change SandboxPolicyRevision.Policy from []byte to *SandboxPolicy in openshell/v1/types/policy.go
- [X] T005 Add SandboxPolicyFromProto, SandboxPolicyToProto (exported) and filesystemPolicyFromProto/ToProto, landlockPolicyFromProto/ToProto, processPolicyFromProto/ToProto (unexported) in openshell/v1/internal/converter/policy.go. Deep-copy NetworkPolicies map entries via existing NetworkPolicyRuleFromProto/ToProto. Deep-copy ReadOnly/ReadWrite slices via CopyStringSlice.
- [X] T006 Add TestSandboxPolicyFromProtoNil, TestSandboxPolicyRoundTrip (all fields populated), TestSandboxPolicyDeepCopy (mutate source, verify isolation), TestSandboxPolicyPartialSubPolicies (only some sub-policies set), TestFilesystemPolicyRoundTrip, TestLandlockPolicyRoundTrip, TestProcessPolicyRoundTrip in openshell/v1/internal/converter/policy_test.go

**Checkpoint**: All new types compile, core SandboxPolicy converters pass round-trip and deep-copy tests

---

## Phase 2: US1 + US2 - Create and Update Policy (Priority: P1) MVP

**Goal**: SDK consumers can set initial policy on SandboxSpec at creation time (US1) and replace full policy via ConfigUpdate at runtime (US2)

**Independent Test**: Construct SandboxSpec with Policy, convert to/from proto, verify round-trip. Construct ConfigUpdate with typed Policy, convert to proto, verify UpdateConfigRequest.policy is populated.

- [X] T007 [US1] Update sandboxSpecFromProto to call SandboxPolicyFromProto(s.GetPolicy()) and SandboxSpecToProto to call SandboxPolicyToProto(spec.Policy) in openshell/v1/internal/converter/sandbox.go
- [X] T008 [US1] Update TestSandboxSpecRoundTrip to include a fully populated Policy field in the test fixture in openshell/v1/internal/converter/sandbox_test.go
- [X] T009 [US2] Update ConfigUpdateToProto to call SandboxPolicyToProto(cu.Policy) instead of proto.Unmarshal(cu.Policy), remove invalid-bytes error path in openshell/v1/internal/converter/setting.go
- [X] T010 [US2] Update TestConfigUpdateToProto to use typed *SandboxPolicy instead of []byte in openshell/v1/internal/converter/setting_test.go

**Checkpoint**: SandboxSpec and ConfigUpdate round-trip with typed policy, make test passes

---

## Phase 3: US3 + US4 - Read Policy from History and Config (Priority: P2)

**Goal**: SDK consumers can inspect policy content from revision history (US3) and current config (US4) as typed structs instead of opaque bytes

**Independent Test**: Construct proto SandboxPolicyRevision/GetSandboxConfigResponse with populated policy, convert to SDK types, verify typed policy fields are accessible.

- [X] T011 [P] [US3] Update SandboxPolicyRevisionFromProto to call SandboxPolicyFromProto(r.GetPolicy()) instead of proto.Marshal + CopyByteSlice in openshell/v1/internal/converter/policy.go
- [X] T012 [P] [US4] Update SandboxConfigFromProto to call SandboxPolicyFromProto(resp.GetPolicy()) instead of proto.Marshal in openshell/v1/internal/converter/setting.go
- [X] T013 [P] [US3] Update TestSandboxPolicyRevisionFromProto to verify typed *SandboxPolicy instead of []byte in openshell/v1/internal/converter/policy_test.go
- [X] T014 [P] [US4] Update TestSandboxConfigFromProto to verify typed *SandboxPolicy instead of []byte in openshell/v1/internal/converter/setting_test.go

**Checkpoint**: All four policy surfaces (create, update, revision, config) use typed SandboxPolicy

---

## Phase 4: US5 - Fake Client Policy Support (Priority: P2)

**Goal**: Fake client stores and returns policy on sandbox create/get, accepts typed policy on config update

**Independent Test**: Create sandbox via fake with SandboxPolicy, get sandbox back, verify policy round-trips correctly.

- [X] T015 [US5] Add copySandboxPolicy helper and update copySandboxSpec to deep-copy Policy field in openshell/v1/fake/sandbox.go
- [X] T016 [US5] Add TestFakeSandboxCreateWithPolicy (create with policy, get back, verify all fields match) in openshell/v1/fake/sandbox_test.go
- [X] T017 [US5] Verify fake config Update method compiles and works with typed *SandboxPolicy in openshell/v1/fake/config.go (may need no changes if it stores ConfigUpdate directly)

**Checkpoint**: Fake client passes policy round-trip test, make test passes

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [X] T018 Verify all new exported types and functions have Go doc comments per Constitution IX
- [X] T019 Run make ci (lint + build + test) and fix any issues

---

## Dependencies & Execution Order

### Phase Dependencies

- **Foundational (Phase 1)**: No dependencies, start immediately. BLOCKS all user stories.
- **US1+US2 (Phase 2)**: Depends on Phase 1 completion (needs types + core converters)
- **US3+US4 (Phase 3)**: Depends on Phase 1 completion. Can run in parallel with Phase 2.
- **US5 (Phase 4)**: Depends on Phase 1 completion (needs types). Can run in parallel with Phases 2-3.
- **Polish (Phase 5)**: Depends on all phases complete.

### Parallel Opportunities

Within Phase 1:
- T001, T002, T003 can run in parallel (different files)
- T004 depends on T001 (same file: types/policy.go)
- T005 depends on T001-T004 (needs types to compile)
- T006 depends on T005

Within Phase 3:
- T011, T012, T013, T014 can all run in parallel (different files)

Cross-phase:
- Phase 2, Phase 3, and Phase 4 can all run in parallel after Phase 1 completes

---

## Implementation Strategy

### MVP First (US1 + US2)

1. Complete Phase 1: Foundational (types + core converters)
2. Complete Phase 2: US1 + US2 (create + update)
3. **STOP and VALIDATE**: `make test` passes, SandboxSpec and ConfigUpdate work with typed policy
4. Continue to Phase 3-5

### Sequential Delivery

1. Phase 1 (Foundational) -> Phase 2 (US1+US2) -> Phase 3 (US3+US4) -> Phase 4 (US5) -> Phase 5 (Polish)

---

## Notes

- Constitution III (Test-First): Write tests alongside implementation in each phase
- Constitution VII (Deep Copy): All map and slice fields must be deep-copied in converters and fake
- [P] tasks = different files, no dependencies
- Commit after each phase checkpoint
- Total tasks: 19
