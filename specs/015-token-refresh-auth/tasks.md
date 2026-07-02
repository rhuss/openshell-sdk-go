# Tasks: Composable Token Refresh with Coalesced Caching

**Input**: Design documents from `specs/015-token-refresh-auth/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Included per Constitution III (Test-First, NON-NEGOTIABLE).

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Add oauth2 dependency and create the new source files

- [x] T001 Add `golang.org/x/oauth2` dependency via `go get golang.org/x/oauth2`
- [x] T002 [P] Create `openshell/v1/auth_refresh.go` with SPDX header, package declaration, and imports
- [x] T003 [P] Create `openshell/v1/auth_refresh_test.go` with SPDX header, package declaration, and imports

---

## Phase 2: Foundational (RefreshOption type and constructor validation)

**Purpose**: Define the functional option types and constructor with nil validation before implementing caching logic

- [x] T004 Define `RefreshOption` type and `refreshConfig` struct with defaults (leeway: 10s, logger: nil) in `openshell/v1/auth_refresh.go`
- [x] T005 Implement `WithLeeway(d time.Duration) RefreshOption` in `openshell/v1/auth_refresh.go`
- [x] T006 Implement `WithLogger(l types.Logger) RefreshOption` in `openshell/v1/auth_refresh.go`
- [x] T007 Implement `RefreshableToken(src oauth2.TokenSource, opts ...RefreshOption) (AuthProvider, error)` constructor with nil-source validation in `openshell/v1/auth_refresh.go`
- [x] T008 Write test: `RefreshableToken` returns error when `TokenSource` is nil in `openshell/v1/auth_refresh_test.go`

**Checkpoint**: Constructor compiles and rejects nil sources. Options resolve correctly.

---

## Phase 3: User Story 1 - Automatic token refresh (Priority: P1) 🎯 MVP

**Goal**: SDK consumers get automatic token caching and concurrent-safe refresh via `RefreshableToken`.

**Independent Test**: Create a mock `TokenSource`, configure `RefreshableToken`, make concurrent calls, verify single-flight behavior and token caching.

### Tests for User Story 1

- [x] T009 [US1] Write test: `GetRequestMetadata` calls `TokenSource.Token()` on first invocation (no cached token) in `openshell/v1/auth_refresh_test.go`
- [x] T010 [US1] Write test: `GetRequestMetadata` returns cached token on second call (no additional `Token()` call) in `openshell/v1/auth_refresh_test.go`
- [x] T011 [US1] Write test: `GetRequestMetadata` refreshes token when cached token is expired in `openshell/v1/auth_refresh_test.go`
- [x] T012 [US1] Write test: 1000 concurrent goroutines calling `GetRequestMetadata` with expired token result in exactly 1 `Token()` call (SC-001) in `openshell/v1/auth_refresh_test.go`
- [x] T013 [US1] Write test: `RequireTransportSecurity()` returns `true` in `openshell/v1/auth_refresh_test.go`

### Implementation for User Story 1

- [x] T014 [US1] Implement `refreshableAuth` struct with `oauth2.TokenSource`, `sync.RWMutex`, cached `*oauth2.Token`, leeway, and logger fields in `openshell/v1/auth_refresh.go`
- [x] T015 [US1] Implement `refreshableAuth.GetRequestMetadata` with RWMutex double-checked locking: RLock fast path (cached valid token), Lock slow path (re-check + fetch from source) in `openshell/v1/auth_refresh.go`
- [x] T016 [US1] Implement `refreshableAuth.RequireTransportSecurity` returning `true` in `openshell/v1/auth_refresh.go`
- [x] T017 [US1] Wire `RefreshableToken` constructor to return `*refreshableAuth` with applied options in `openshell/v1/auth_refresh.go`
- [x] T018 [US1] Run `make test` and verify all US1 tests pass

**Checkpoint**: Core token refresh with single-flight caching works. SC-001 verified.

---

## Phase 4: User Story 2 - Graceful degradation on refresh failure (Priority: P2)

**Goal**: On refresh failure, return stale cached token with a logged warning instead of erroring.

**Independent Test**: Mock a `TokenSource` that succeeds once then fails, verify stale token returned and warning logged.

### Tests for User Story 2

- [x] T019 [US2] Write test: refresh failure with existing cached token returns stale token in `openshell/v1/auth_refresh_test.go`
- [x] T020 [US2] Write test: refresh failure with existing cached token logs warning via configured logger in `openshell/v1/auth_refresh_test.go`
- [x] T021 [US2] Write test: refresh failure with no cached token returns error in `openshell/v1/auth_refresh_test.go`
- [x] T022 [US2] Write test: refresh failure with no logger configured does not panic in `openshell/v1/auth_refresh_test.go`

### Implementation for User Story 2

- [x] T023 [US2] Add error handling to `refreshableAuth.GetRequestMetadata` slow path: on `Token()` error, if cached token exists return it and log warning; if no cached token, return error in `openshell/v1/auth_refresh.go`
- [x] T024 [US2] Run `make test` and verify all US1+US2 tests pass

**Checkpoint**: Graceful degradation works. Stale tokens returned on transient failures.

---

## Phase 5: User Story 3 - Configurable refresh leeway (Priority: P3)

**Goal**: Developers can tune refresh timing via `WithLeeway`.

**Independent Test**: Create token with 60s expiry, set 30s leeway, verify refresh triggers at 35s mark.

### Tests for User Story 3

- [x] T025 [US3] Write test: default leeway (10s) triggers refresh within 10s of expiry in `openshell/v1/auth_refresh_test.go`
- [x] T026 [US3] Write test: custom leeway via `WithLeeway(30s)` triggers refresh 30s before expiry in `openshell/v1/auth_refresh_test.go`
- [x] T027 [US3] Write test: token with zero expiry is treated as always-valid (never refreshes) in `openshell/v1/auth_refresh_test.go`

### Implementation for User Story 3

- [x] T028 [US3] Ensure leeway check in `isTokenValid` helper uses `token.Expiry.Add(-leeway).Before(time.Now())` and handles zero-expiry case in `openshell/v1/auth_refresh.go`
- [x] T029 [US3] Run `make test` and verify all US1+US2+US3 tests pass

**Checkpoint**: Leeway configuration works. Zero-expiry edge case handled.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, benchmarks, and CI validation

- [x] T030 [P] Write benchmark: `BenchmarkGetRequestMetadata_CachedToken` to verify <1μs fast-path (SC-002) in `openshell/v1/auth_refresh_test.go`
- [x] T031 [P] Add doc comments for all exported symbols: `RefreshableToken`, `RefreshOption`, `WithLeeway`, `WithLogger` in `openshell/v1/auth_refresh.go`
- [x] T032 Add `RefreshableToken` usage example to `openshell/v1/doc.go`
- [x] T033 Run `make ci` (lint + build + test) to verify full pipeline passes (SC-003)
- [x] T034 Verify implementation is ≤150 lines of non-test code (SC-004)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (files exist, dependency added)
- **User Story 1 (Phase 3)**: Depends on Phase 2 (types and constructor defined)
- **User Story 2 (Phase 4)**: Depends on Phase 3 (core caching logic exists to add error handling)
- **User Story 3 (Phase 5)**: Can run after Phase 2 (leeway is independent of error handling), but sequential with Phase 3 since they touch same file/function
- **Polish (Phase 6)**: Depends on all user stories complete

### Within Each User Story

- Tests written first (must fail before implementation)
- Implementation makes tests pass
- `make test` checkpoint after each story

### Parallel Opportunities

- T002, T003: File creation in parallel
- T030, T031: Benchmark and doc comments in parallel
- US2 and US3 are conceptually independent but touch the same function, so sequential is safer

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Phase 1: Setup (add dependency, create files)
2. Phase 2: Foundational (option types, constructor)
3. Phase 3: US1 (caching + single-flight)
4. **STOP and VALIDATE**: `make test` passes, SC-001 verified
5. Core value delivered: automatic token refresh works

### Incremental Delivery

1. Setup + Foundational + US1 → Core refresh works (MVP)
2. Add US2 → Graceful degradation on failures
3. Add US3 → Configurable leeway
4. Polish → Docs, benchmarks, CI validation

---

## Notes

- All tasks touch only 2 files: `auth_refresh.go` and `auth_refresh_test.go` (plus `doc.go` for T032)
- Constitution III mandates test-first: write failing tests before implementation
- `make ci` is the final validation gate
- Token values must never appear in error messages or log output (Constitution VI)
