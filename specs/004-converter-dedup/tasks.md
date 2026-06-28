# Tasks: Converter Code Deduplication

**Input**: Design documents from `specs/004-converter-dedup/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create the types package skeleton

- [X] T001 Create openshell/v1/types/doc.go with package documentation and SPDX license header

---

## Phase 2: Foundational (Types Package)

**Purpose**: Populate the types package with all domain types moved from v1/

**CRITICAL**: No user story work can begin until this phase is complete

- [X] T002 [P] Create openshell/v1/types/sandbox.go with Sandbox, SandboxSpec, SandboxTemplate, SandboxStatus, SandboxCondition, AttachProviderResult, DetachProviderResult structs (copied from openshell/v1/sandbox.go)
- [X] T003 [P] Create openshell/v1/types/provider.go with Provider, ProviderSpec structs (copied from openshell/v1/provider.go)
- [X] T004 [P] Create openshell/v1/types/exec.go with ExecResult, ExecChunk structs (copied from openshell/v1/exec.go)
- [X] T005 [P] Create openshell/v1/types/health.go with HealthResult struct (copied from openshell/v1/health.go)
- [X] T006 [P] Create openshell/v1/types/errors.go with ErrorCode type, StatusError struct, IsStatusError function, and all error code constants (copied from openshell/v1/errors.go)
- [X] T007 [P] Create openshell/v1/types/types.go with SandboxPhase, EventType, StreamType enums, TLSConfig, RetryPolicy structs, and all associated constants (copied from openshell/v1/types.go)
- [X] T008 [P] Create openshell/v1/types/options.go with ExecOptions, ListOptions, WaitOptions, WatchOptions, CreateOptions, DeleteOptions, GetOptions, UpdateOptions structs (copied from openshell/v1/options.go)
- [X] T009 [P] Create openshell/v1/types/config.go with Config struct (copied from openshell/v1/client.go)
- [X] T010 [P] Create openshell/v1/types/watch.go with Event[T] struct and WatchInterface[T] interface (copied from openshell/v1/watch.go)
- [X] T011 [P] Create openshell/v1/types/logger.go with Logger interface (copied from openshell/v1/logger.go)
- [X] T012 [P] Create openshell/v1/types/auth.go with AuthProvider interface (copied from openshell/v1/auth.go)

**Checkpoint**: Types package compiles independently with `go build ./openshell/v1/types/...`

---

## Phase 3: User Story 2 - Backward Compatible Type Aliases (Priority: P1)

**Goal**: Replace type definitions in v1/ with type aliases pointing to types/, ensuring all existing consumers compile without changes

**Independent Test**: Run `make test` — all existing tests pass with aliases in place

- [X] T013 [US2] Replace type definitions in openshell/v1/sandbox.go with type aliases to types/ (Sandbox, SandboxSpec, SandboxTemplate, SandboxStatus, SandboxCondition, AttachProviderResult, DetachProviderResult)
- [X] T014 [P] [US2] Replace type definitions in openshell/v1/provider.go with type aliases to types/ (Provider, ProviderSpec)
- [X] T015 [P] [US2] Replace type definitions in openshell/v1/exec.go with type aliases to types/ (ExecResult, ExecChunk)
- [X] T016 [P] [US2] Replace type definition in openshell/v1/health.go with type alias to types/ (HealthResult)
- [X] T017 [P] [US2] Replace type definitions in openshell/v1/errors.go with type aliases and constant re-exports to types/ (ErrorCode, StatusError, IsStatusError, error code constants)
- [X] T018 [P] [US2] Replace type definitions in openshell/v1/types.go with type aliases and constant re-exports to types/ (SandboxPhase, EventType, StreamType, TLSConfig, RetryPolicy, all constants)
- [X] T019 [P] [US2] Replace type definitions in openshell/v1/options.go with type aliases to types/ (all option structs)
- [X] T020 [P] [US2] Replace Config definition in openshell/v1/client.go with type alias to types/ (Config only; keep Client, ClientInterface)
- [X] T021 [P] [US2] Replace type definitions in openshell/v1/watch.go with type aliases to types/ (Event[T], WatchInterface[T], keep watcher[T] implementation)
- [X] T022 [P] [US2] Replace type definitions in openshell/v1/logger.go with type alias to types/ (Logger)
- [X] T023 [P] [US2] Replace type definitions in openshell/v1/auth.go with type alias to types/ (AuthProvider; keep noAuth, staticToken implementations)

**Checkpoint**: `make test` passes — all existing tests work with type aliases

---

## Phase 4: User Story 1 - Single Source of Truth for Converters (Priority: P1)

**Goal**: Update converter package to import types/ instead of v1/, wire clients to use converter, remove all duplicated conversion functions

**Independent Test**: grep for unexported conversion functions in *_client.go and grpc_errors.go returns zero matches

- [X] T024 [US1] Update converter imports in openshell/v1/internal/converter/sandbox.go to use types/ instead of v1/
- [X] T025 [P] [US1] Update converter imports in openshell/v1/internal/converter/provider.go to use types/ instead of v1/
- [X] T026 [P] [US1] Update converter imports in openshell/v1/internal/converter/exec.go to use types/ instead of v1/
- [X] T027 [P] [US1] Update converter imports in openshell/v1/internal/converter/errors.go to use types/ instead of v1/
- [X] T028 [US1] Create openshell/v1/internal/converter/copy.go with deep-copy helpers (copyStringMap, copyBoolPtr, copyStringSlice) moved from openshell/v1/sandbox_client.go
- [X] T029 [US1] Remove duplicated conversion functions from openshell/v1/sandbox_client.go (sandboxFromProto, sandboxSpecFromProto, sandboxStatusFromProto, sandboxPhaseFromProto, sandboxSpecToProto, copyStringMap, copyBoolPtr, copyStringSlice) and replace call sites with converter package calls
- [X] T030 [P] [US1] Remove duplicated conversion functions from openshell/v1/provider_client.go (providerFromProto, providerToProto, timeFromMillis, millisFromTime) and replace call sites with converter package calls
- [X] T031 [P] [US1] Remove duplicated conversion functions from openshell/v1/exec_client.go (execChunkFromEvent, execRequestToProto, execInteractiveRequestToProto, execResultFromEvents) and replace call sites with converter package calls. NOTE: local execChunkFromEvent returns (chunk, exitCode, isExit, err) while converter.ExecChunkFromEvent returns (*ExecChunk, int, error) — call sites must adapt to the different return signature (the isExit bool is not returned by the converter; instead check for a non-nil exit code or use ExecResultFromEvents for batch processing).
- [X] T032a [US1] Remove duplicated fromGRPCError function and grpcToSDK map from openshell/v1/grpc_errors.go and replace call sites with converter.FromGRPCError. The converter package already has an exported FromGRPCError with identical logic.

**Checkpoint**: Zero duplicated conversion functions. `make test` passes.

---

## Phase 5: User Story 3 - Independent Converter Testing (Priority: P2)

**Goal**: Converter tests import types/ directly, proving the circular dependency is broken

**Independent Test**: `go test ./openshell/v1/internal/converter/...` passes and `go list -f '{{.Imports}}' ./openshell/v1/internal/converter/` does not include `openshell/v1`

- [X] T032 [US3] Update converter test imports in openshell/v1/internal/converter/sandbox_test.go to use types/ instead of v1/
- [X] T033 [P] [US3] Update converter test imports in openshell/v1/internal/converter/provider_test.go to use types/ instead of v1/
- [X] T034 [P] [US3] Update converter test imports in openshell/v1/internal/converter/exec_test.go to use types/ instead of v1/
- [X] T035 [P] [US3] Update converter test imports in openshell/v1/internal/converter/errors_test.go to use types/ instead of v1/

**Checkpoint**: Converter tests pass independently. No v1/ import in converter package.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification and cleanup

- [X] T036 Run `make ci` (lint + build + test) and fix any remaining issues
- [X] T037 Verify no circular imports exist with `go build ./...`
- [X] T038 Verify zero duplicated conversion functions with `rg '^func [a-z].*(From|To)(Proto|Events)|^func fromGRPCError' openshell/v1/*_client.go openshell/v1/grpc_errors.go`
- [X] T039 Update openshell/v1/internal/grpc/conn.go if it references types that moved (TLSConfig → types.TLSConfig or via v1/ alias)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **Foundational (Phase 2)**: Depends on Setup — BLOCKS all user stories
- **US2 (Phase 3)**: Depends on Foundational — type aliases require types/ to exist
- **US1 (Phase 4)**: Depends on US2 — converter needs types/ aliases in place before v1/ can import converter
- **US3 (Phase 5)**: Depends on US1 — converter tests update after converter source is updated
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

- **US2 (backward compat)**: Must complete first — type aliases enable the subsequent dependency break
- **US1 (single source of truth)**: Depends on US2 — requires cycle-free import graph
- **US3 (independent testing)**: Depends on US1 — converter source must be updated before tests

### Within Each Phase

- Tasks marked [P] within a phase can run in parallel
- Non-[P] tasks within a phase should run sequentially

### Parallel Opportunities

- Phase 2: All T002-T012 can run in parallel (each creates a separate file)
- Phase 3: T014-T023 can run in parallel after T013 (each modifies a separate v1/ file)
- Phase 4: T025-T027 in parallel, T030-T031 in parallel
- Phase 5: T033-T035 in parallel after T032

---

## Implementation Strategy

### MVP First (US2 + US1)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (types package)
3. Complete Phase 3: US2 (type aliases — backward compat verified)
4. Complete Phase 4: US1 (converter wiring — duplication eliminated)
5. **STOP and VALIDATE**: `make test` passes, zero duplicated functions

### Full Delivery

1. Complete MVP above
2. Complete Phase 5: US3 (converter test independence)
3. Complete Phase 6: Polish (final verification)

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- This is a pure refactoring: no new features, no new tests needed (existing tests validate correctness)
- Type aliases preserve full backward compatibility including struct literals and type assertions
- Constants use `const X = types.X` re-export pattern (not type aliases)
