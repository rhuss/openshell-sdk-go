# Tasks: Inference Route Client

**Input**: Design documents from `specs/023-inference-client/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included per Constitution III (Test-First).

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: No setup required. Proto file (`proto/inference.proto`) and generated stubs (`proto/inferencev1/`) already exist and are wired into `buf.gen.yaml`.

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: SDK types and converters that all user stories depend on

- [x] T001 Create inference SDK types (InferenceRouteConfig, InferenceRoute, ValidatedEndpoint) in `openshell/v1/types/inference.go`
- [x] T002 [P] Create proto-to-SDK and SDK-to-proto converters in `openshell/v1/internal/converter/inference.go`
- [x] T003 [P] Create converter round-trip tests in `openshell/v1/internal/converter/inference_test.go`

**Checkpoint**: Types and converters ready. User story implementation can begin.

---

## Phase 3: User Story 1+2 - Configure and Retrieve Inference Routes (Priority: P1)

**Goal**: Workspace admins can set inference routes; workspace users can retrieve them.

**Independent Test**: Set a route for a workspace via gRPC stub, then retrieve it and verify all fields match.

### Implementation for User Story 1+2

- [x] T004 [US1] Define InferenceInterface (SetRoute, GetRoute, DeleteRoute) with type aliases in `openshell/v1/inference.go`
- [x] T005 [US1] Write unit tests for SetRoute and GetRoute (validation errors, successful calls) in `openshell/v1/inference_client_test.go`
- [x] T006 [US1] Implement inferenceClient with all three RPC methods using `inferencev1.NewInferenceClient(conn)` in `openshell/v1/inference_client.go`
- [x] T007 [US1] Wire Inference() accessor into ClientInterface, Client struct, and NewClient constructor in `openshell/v1/client.go`

**Checkpoint**: SetRoute and GetRoute work against a gRPC stub. `make test` passes.

---

## Phase 4: User Story 3 - Remove Inference Route (Priority: P2)

**Goal**: Workspace admins can delete inference routes. Operation is idempotent.

**Independent Test**: Delete an existing route and verify subsequent GetRoute returns not-found.

### Implementation for User Story 3

- [x] T008 [US3] Add DeleteRoute unit tests (validation, idempotent delete) to `openshell/v1/inference_client_test.go`

**Checkpoint**: All three real client methods tested. `make test` passes.

---

## Phase 5: User Story 4 - Fake Client (Priority: P2)

**Goal**: Developers can test inference operations without a live gateway using an in-memory fake.

**Independent Test**: Use fake.NewClient() to set, get, and delete routes, verifying correct in-memory behavior.

### Implementation for User Story 4

- [x] T009 [US4] Write fake inference client tests (set/get/delete, validation parity with real client) in `openshell/v1/fake/inference_test.go`
- [x] T010 [US4] Implement fakeInferenceClient with in-memory store keyed by workspace/routeName in `openshell/v1/fake/inference.go`
- [x] T011 [US4] Wire Inference() accessor into fake.Client struct and NewClient constructor in `openshell/v1/fake/fake.go`

**Checkpoint**: Fake client passes all tests. Real and fake share identical validation behavior (Constitution XI).

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation and final validation

- [x] T012 [P] Update README.md with inference client feature entry and usage example
- [x] T013 [P] Add inference example to package doc.go in `openshell/v1/doc.go` and add fake example in `openshell/v1/example_fake_test.go`
- [x] T014 Run `make ci` (lint + build + test) to validate everything passes

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: N/A (already complete)
- **Foundational (Phase 2)**: BLOCKS all user stories. Types and converters must exist first.
- **US1+2 (Phase 3)**: Depends on Phase 2. Defines interface and real client.
- **US3 (Phase 4)**: Depends on Phase 3 (tests extend existing test file, implementation already in Phase 3).
- **US4 (Phase 5)**: Depends on Phase 2 (types). Can run in parallel with Phase 3 since fake and real client are in different packages.
- **Polish (Phase 6)**: Depends on all user stories being complete.

### Within Each Phase

- Types (T001) before converters (T002, T003)
- Interface definition (T004) before tests (T005) before implementation (T006)
- Implementation (T006) before wiring (T007)
- Fake tests (T009) before fake implementation (T010) before wiring (T011)

### Parallel Opportunities

- T002 and T003 can run in parallel (different files, both depend only on T001)
- T012 and T013 can run in parallel (different files)
- Phase 5 (fake) can start after Phase 2 (does not need real client)

---

## Parallel Example: Phase 2

```text
# After T001 (types) completes, launch converters in parallel:
Task: "Create converters in openshell/v1/internal/converter/inference.go"
Task: "Create converter tests in openshell/v1/internal/converter/inference_test.go"
```

---

## Implementation Strategy

### MVP First (User Stories 1+2)

1. Complete Phase 2: Foundational (types + converters)
2. Complete Phase 3: Real client with SetRoute + GetRoute
3. **STOP and VALIDATE**: `make test` passes, SetRoute/GetRoute work
4. Continue to Phase 4-6

### Incremental Delivery

1. Phase 2 (Foundational) -> Types and converters ready
2. Phase 3 (US1+2) -> Set/Get route works -> Test independently (MVP!)
3. Phase 4 (US3) -> Delete route works -> Test independently
4. Phase 5 (US4) -> Fake client works -> Test independently
5. Phase 6 (Polish) -> Docs, lint, CI green

---

## Notes

- Proto stubs already exist; no code generation tasks needed
- All 3 RPC methods implemented in one file (Go interface requires all methods)
- US1+US2 combined because SetRoute and GetRoute are both P1 and tightly coupled
- DeleteRoute implementation is part of Phase 3 (same file), but its tests are Phase 4
- Fake must mirror real client validation per Constitution XI (Fake-Real Parity)
