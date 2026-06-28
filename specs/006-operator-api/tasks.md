# Tasks: Operator API Extensions (Phase 2a)

**Input**: Design documents from `specs/006-operator-api/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Organization**: Tasks are grouped by user story. Tests are written first per Constitution Principle III (Test-First).

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Domain Types)

**Purpose**: Define all new domain types in v1/types/ before any client code

- [X] T001 [P] Define ServiceEndpoint type in `openshell/v1/types/service.go` (ID, SandboxID, SandboxName, ServiceName, TargetPort, Domain, URL)
- [X] T002 [P] Define ProviderProfile, ProfileCredential, ProfileCategory (string enum), ProfileDiscovery, NetworkEndpoint, NetworkBinary, ProfileImportItem, ProfileDiagnostic, ImportResult, UpdateResult, LintResult in `openshell/v1/types/profile.go`
- [X] T003 [P] Define RefreshStrategy (string enum), RefreshStatus, RefreshConfig in `openshell/v1/types/refresh.go`
- [X] T004 Add StopOnTerminal bool field to WatchOptions in `openshell/v1/types/options.go`

---

## Phase 2: Foundational (Converters)

**Purpose**: Proto↔SDK converters for all new types. Must complete before client implementations.

- [X] T005 [P] Test service endpoint converter (ServiceEndpointResponse → ServiceEndpoint, deep copy of fields) in `openshell/v1/internal/converter/service_test.go`
- [X] T006 [P] Test profile converters (ProviderProfile, ProfileCredential, ProfileImportItem, ProfileDiagnostic, ImportResult, UpdateResult, LintResult, NetworkEndpoint, NetworkBinary) in `openshell/v1/internal/converter/profile_test.go`
- [X] T007 [P] Test refresh converters (RefreshStatus, RefreshConfig ↔ proto, RefreshStrategy mapping) in `openshell/v1/internal/converter/refresh_test.go`
- [X] T008 [P] Implement service endpoint converter in `openshell/v1/internal/converter/service.go`
- [X] T009 [P] Implement profile converters in `openshell/v1/internal/converter/profile.go`
- [X] T010 [P] Implement refresh converters in `openshell/v1/internal/converter/refresh.go`

**Checkpoint**: Run `make test` to verify all converter tests pass

---

## Phase 3: User Story 1 — Service Exposure (Priority: P1)

**Goal**: Operators can expose, get, list, and delete sandbox service endpoints

**Independent Test**: Expose service → Get → List → Delete → Get returns NotFound

- [X] T011 [P] [US1] Define ServiceInterface (Expose, Get, List, Delete) in `openshell/v1/service.go`
- [X] T012 [P] [US1] Test serviceClient Expose, Get, List, Delete against mock gRPC server in `openshell/v1/service_client_test.go`
- [X] T013 [US1] Implement serviceClient using gRPC OpenShell service RPCs (ExposeService, GetService, ListServices, DeleteService) in `openshell/v1/service_client.go`
- [X] T014 [US1] Add `Services() ServiceInterface` to ClientInterface in `openshell/v1/client.go` and wire in Client constructor

---

## Phase 4: User Story 2 — Provider Profiles (Priority: P1)

**Goal**: Operators can list, get, import, update, lint, and delete provider profiles

**Independent Test**: List → Import → Get → Lint → Update → Delete → Get returns NotFound

- [X] T015 [P] [US2] Define ProfileInterface (List, Get, Import, Update, Lint, Delete) in `openshell/v1/profile.go`
- [X] T016 [P] [US2] Test profileClient List, Get, Import, Update, Lint, Delete against mock gRPC server in `openshell/v1/profile_client_test.go`
- [X] T017 [US2] Implement profileClient using gRPC RPCs (ListProviderProfiles, GetProviderProfile, ImportProviderProfiles, UpdateProviderProfiles, LintProviderProfiles, DeleteProviderProfile) in `openshell/v1/profile_client.go`
- [X] T018 [US2] Add `Profiles() ProfileInterface` to ProviderInterface in `openshell/v1/provider.go` and wire in providerClient

---

## Phase 5: User Story 3 — Credential Refresh (Priority: P2)

**Goal**: Operators can configure, monitor, trigger, and remove credential refresh

**Independent Test**: Configure → GetStatus → Rotate → Delete → GetStatus reflects removal

- [X] T019 [P] [US3] Define RefreshInterface (GetStatus, Configure, Rotate, Delete) in `openshell/v1/refresh.go`
- [X] T020 [P] [US3] Test refreshClient GetStatus, Configure, Rotate, Delete against mock gRPC server in `openshell/v1/refresh_client_test.go`
- [X] T021 [US3] Implement refreshClient using gRPC RPCs (GetProviderRefreshStatus, ConfigureProviderRefresh, RotateProviderCredential, DeleteProviderRefresh) in `openshell/v1/refresh_client.go`
- [X] T022 [US3] Add `Refresh() RefreshInterface` to ProviderInterface in `openshell/v1/provider.go` and wire in providerClient

---

## Phase 6: User Story 4 — StopOnTerminal Watch (Priority: P2)

**Goal**: Watch with StopOnTerminal=true auto-closes when sandbox reaches Ready or Error

**Independent Test**: Watch with StopOnTerminal=true, sandbox transitions to Ready, channel closes

- [X] T023 [P] [US4] Test StopOnTerminal=true closes watcher after Ready event in `openshell/v1/sandbox_client_test.go`
- [X] T024 [P] [US4] Test StopOnTerminal=true closes watcher after Error event in `openshell/v1/sandbox_client_test.go`
- [X] T025 [US4] Implement StopOnTerminal support in sandboxClient Watch (pass to server + client-side phase check) in `openshell/v1/sandbox_client.go`

---

## Phase 7: User Story 5 — Fake Client Updates (Priority: P3)

**Goal**: FakeClient compiles with updated interfaces, new stubs return Unimplemented

- [X] T026 [P] [US5] Test fakeServiceClient stubs (Expose, Get, List, Delete return Unimplemented) in `openshell/v1/fake/service_test.go`
- [X] T027 [P] [US5] Test fakeProfileClient stubs (List, Get, Import, Update, Lint, Delete return Unimplemented) in `openshell/v1/fake/profile_test.go`
- [X] T028 [P] [US5] Test fakeRefreshClient stubs (GetStatus, Configure, Rotate, Delete return Unimplemented) in `openshell/v1/fake/refresh_test.go`
- [X] T029 [P] [US5] Implement fakeServiceClient stub in `openshell/v1/fake/service.go`
- [X] T030 [P] [US5] Implement fakeProfileClient stub in `openshell/v1/fake/profile.go`
- [X] T031 [P] [US5] Implement fakeRefreshClient stub in `openshell/v1/fake/refresh.go`
- [X] T032 [US5] Wire Services() on FakeClient in `openshell/v1/fake/fake.go`, wire Profiles() and Refresh() on fakeProviderClient in `openshell/v1/fake/provider.go`
- [X] T033 [US5] Test StopOnTerminal support in fake Watch (auto-close on Ready/Error) in `openshell/v1/fake/sandbox_test.go`
- [X] T034 [US5] Implement StopOnTerminal in fake Watch (close watcher after terminal phase event) in `openshell/v1/fake/sandbox.go`

---

## Phase 8: Polish & Cross-Cutting

**Purpose**: Re-exports, documentation, lint compliance, CI verification

- [X] T035 [P] Re-export all new types and interfaces from `openshell/v1/` via type aliases (ServiceEndpoint, ServiceInterface, ProviderProfile, ProfileInterface, RefreshInterface, etc.)
- [X] T036 [P] Add doc.go examples showing Services().Expose and Providers().Profiles().List usage
- [X] T037 Verify all `.go` files in new paths have SPDX license headers
- [X] T038 Run `make ci` (lint + build + test) and fix any violations

---

## Dependencies

```
T001-T004 (types) ─────────┐
                            ├─► T005-T010 (converters)
                            │         │
                            │   T011-T014 (US1: services)
                            │   T015-T018 (US2: profiles)  ◄── parallel with US1
                            │   T019-T022 (US3: refresh)   ◄── parallel with US1/US2
                            │         │
                            │   T023-T025 (US4: StopOnTerminal)
                            │         │
                            │   T026-T034 (US5: fake stubs + StopOnTerminal)
                            │         │
                            └─► T035-T038 (polish)
```

## Parallel Execution Opportunities

- **Phase 1**: T001, T002, T003 all parallel (different type files)
- **Phase 2**: T005-T007 parallel (test files), T008-T010 parallel (impl files)
- **Phase 3-5**: US1, US2, US3 can run in parallel after converters complete
- **Phase 7**: T026-T031 all parallel (independent stub files)

## Implementation Strategy

**MVP**: Phases 1-3 (types + converters + services) — delivers the most-requested operator capability.

**Full Scope**: All 8 phases, 38 tasks total.
- Types: 4 tasks
- Converters: 6 tasks
- US1 (services): 4 tasks
- US2 (profiles): 4 tasks
- US3 (refresh): 4 tasks
- US4 (StopOnTerminal): 3 tasks
- US5 (fake stubs): 9 tasks
- Polish: 4 tasks
