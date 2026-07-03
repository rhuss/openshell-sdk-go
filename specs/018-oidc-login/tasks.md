# Tasks: OIDC Login Package

**Input**: Design documents from `specs/018-oidc-login/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Included per Constitution III (Test-First, NON-NEGOTIABLE).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Create package structure, error types, and options framework

- [X] T001 Create package directory and doc.go stub in openshell/v1/oidc/doc.go
- [X] T002 [P] Define error sentinel variables in openshell/v1/oidc/errors.go
- [X] T003 [P] Define LoginOption type and all With* option functions in openshell/v1/oidc/options.go
- [X] T004 [P] Write tests for option application in openshell/v1/oidc/options_test.go
- [X] T005 [P] Write tests for error sentinel types in openshell/v1/oidc/errors_test.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [X] T006 Write tests for OIDC discovery fetch and cache in openshell/v1/oidc/discovery_test.go
- [X] T007 Implement OIDC discovery document fetch and in-memory cache in openshell/v1/oidc/discovery.go
- [X] T008 [P] Write tests for token persistence (read/write oidcBundle) in openshell/v1/oidc/token_test.go
- [X] T009 [P] Implement token read/write (oidcBundle JSON format) in openshell/v1/oidc/token.go
- [X] T010 Write tests for gateway OIDC config fields in openshell/v1/gateway/config_test.go: verify that parseMetadata populates Config.OIDCIssuer and Config.OIDCClientID from metadata.json, and that missing/empty OIDC fields are handled gracefully (backward compatibility with existing gateways)
- [X] T011 Extend gateway config with OIDC fields in openshell/v1/gateway/config.go: add OIDCIssuer and OIDCClientID to both the internal metadataJSON struct (JSON tags: oidc_issuer, oidc_client_id) and the public Config struct, and update parseMetadata to copy the new fields into the returned Config. Interfaces consumed by T022: gateway.LoadConfig(name string) (*gateway.Config, error), Config.Dir string, Config.OIDCIssuer string, Config.OIDCClientID string

**Checkpoint**: Foundation ready. Discovery, token persistence, and gateway metadata extension all working.

---

## Phase 3: User Story 1 - Gateway-Aware Interactive Login (Priority: P1) MVP

**Goal**: A Go program calls `Login(ctx, "my-gateway")`, completes browser-based OIDC auth, and tokens are persisted to disk for use by `gateway.NewClient`.

**Independent Test**: Call Login with a gateway name against a test OIDC provider (httptest.Server), complete the auth code exchange, verify oidc_token.json is written correctly.

### Tests for User Story 1

- [X] T012 [P] [US1] Write tests for PKCE verifier/challenge generation in openshell/v1/oidc/authcode_test.go
- [X] T013 [P] [US1] Write tests for auth code flow (callback server, state validation, code exchange) in openshell/v1/oidc/authcode_test.go
- [X] T014 [P] [US1] Write tests for keyboard fallback flow in openshell/v1/oidc/keyboard_test.go
- [X] T015 [P] [US1] Write tests for browser opener (platform detection, failure fallback) in openshell/v1/oidc/browser_test.go

### Implementation for User Story 1

- [X] T016 [P] [US1] Implement PKCE code verifier and S256 challenge generation in openshell/v1/oidc/authcode.go
- [X] T017 [P] [US1] Implement platform-aware browser opener in openshell/v1/oidc/browser.go
- [X] T018 [US1] Implement localhost callback server (port selection, state validation, code extraction) in openshell/v1/oidc/authcode.go
- [X] T019 [US1] Implement auth code token exchange with PKCE in openshell/v1/oidc/authcode.go. Check discovery document's CodeChallengeMethodsSupported; if S256 is not listed, proceed without PKCE (omit code_challenge and code_verifier parameters) per spec edge case "provider does not support PKCE"
- [X] T020 [US1] Implement keyboard fallback flow (display URL, read pasted code) in openshell/v1/oidc/keyboard.go
- [X] T021 [US1] Write tests for Login entry point (gateway resolution, token reuse, flow orchestration) in openshell/v1/oidc/oidc_test.go
- [X] T022 [US1] Implement Login function with gateway-aware resolution, existing token check, and flow orchestration in openshell/v1/oidc/oidc.go

**Checkpoint**: Gateway-aware interactive login works end-to-end. Tokens persist to disk. Browser fallback to keyboard works.

---

## Phase 4: User Story 2 - Client Credentials for Service Accounts (Priority: P2)

**Goal**: An operator calls `ClientCredentials(ctx, opts...)` with client ID and secret, receives tokens without user interaction.

**Independent Test**: Call ClientCredentials against a test OIDC provider, verify access token returned, verify secrets never appear in errors.

### Tests for User Story 2

- [X] T023 [P] [US2] Write tests for client credentials exchange (success, invalid creds, secret redaction) in openshell/v1/oidc/credentials_test.go

### Implementation for User Story 2

- [X] T024 [US2] Implement ClientCredentials function in openshell/v1/oidc/credentials.go

**Checkpoint**: Client credentials flow works. Secrets never leak in errors.

---

## Phase 5: User Story 3 - Device Code Flow (Priority: P3)

**Goal**: A developer calls `DeviceLogin(ctx, opts...)`, the system displays a verification URL and user code, polls until authorized.

**Independent Test**: Call DeviceLogin against a test OIDC provider, verify display output, simulate authorization, verify tokens returned after polling.

### Tests for User Story 3

- [X] T025 [P] [US3] Write tests for device code flow (request, polling, slow_down, expiry, custom display) in openshell/v1/oidc/device_test.go

### Implementation for User Story 3

- [X] T026 [US3] Implement DeviceLogin function with device code request, polling loop, and display callback in openshell/v1/oidc/device.go

**Checkpoint**: Device code flow works. Polling respects provider interval. Custom display callback invoked.

---

## Phase 6: User Story 4 - Standalone Login Without Gateway (Priority: P3)

**Goal**: A developer calls `Login(ctx, "", WithIssuer(...), WithClientID(...))` without any gateway dependency, tokens returned in-memory.

**Independent Test**: Call Login with explicit issuer/clientID and WithInMemory, verify tokens returned, verify no disk write.

### Tests for User Story 4

- [X] T027 [P] [US4] Write tests for standalone login (explicit issuer/clientID, in-memory mode) in openshell/v1/oidc/oidc_test.go

### Implementation for User Story 4

- [X] T028 [US4] Add standalone flow path and in-memory mode to Login function in openshell/v1/oidc/oidc.go

**Checkpoint**: Standalone login works without gateway. In-memory mode skips disk persistence.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Documentation, examples, and final quality checks

- [X] T029 [P] Write runnable examples in doc.go (Login, DeviceLogin, ClientCredentials) in openshell/v1/oidc/doc.go
- [X] T030 [P] Update README.md with OIDC login feature description and usage examples
- [X] T031 Run make ci to verify lint, build, and all tests pass

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, can start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (error types, options)
- **US1 (Phase 3)**: Depends on Phase 2 (discovery, token persistence, gateway config)
- **US2 (Phase 4)**: Depends on Phase 2 (discovery). Independent of US1.
- **US3 (Phase 5)**: Depends on Phase 2 (discovery). Independent of US1, US2.
- **US4 (Phase 6)**: Depends on Phase 3 (Login function exists). Extends Login with standalone path.
- **Polish (Phase 7)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (P1)**: After Phase 2. No dependencies on other stories.
- **US2 (P2)**: After Phase 2. Independent of US1.
- **US3 (P3)**: After Phase 2. Independent of US1, US2.
- **US4 (P3)**: After US1 (extends the Login function). Independent of US2, US3.

### Within Each User Story

- Tests MUST be written and FAIL before implementation (Constitution III)
- Foundation types before flow logic
- Core implementation before integration
- Story complete before moving to next priority

### Parallel Opportunities

- T002, T003, T004, T005 can all run in parallel (Phase 1)
- T008/T009 can run in parallel with T006/T007 (Phase 2, different files)
- T012, T013, T014, T015 can all run in parallel (US1 tests)
- T016, T017 can run in parallel (US1 PKCE + browser, different files)
- US2 (Phase 4) and US3 (Phase 5) can run in parallel after Phase 2

---

## Parallel Example: User Story 1

```bash
# Launch all US1 tests together:
Task: "Write tests for PKCE in openshell/v1/oidc/authcode_test.go"
Task: "Write tests for auth code flow in openshell/v1/oidc/authcode_test.go"
Task: "Write tests for keyboard fallback in openshell/v1/oidc/keyboard_test.go"
Task: "Write tests for browser opener in openshell/v1/oidc/browser_test.go"

# Launch parallel implementation:
Task: "Implement PKCE generation in openshell/v1/oidc/authcode.go"
Task: "Implement browser opener in openshell/v1/oidc/browser.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (package structure, errors, options)
2. Complete Phase 2: Foundational (discovery, token persistence, gateway config)
3. Complete Phase 3: User Story 1 (gateway-aware interactive login)
4. **STOP and VALIDATE**: Test Login against a real OIDC provider
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational: Foundation ready
2. Add US1: Gateway interactive login works (MVP)
3. Add US2: Service accounts can authenticate (CI/CD unblocked)
4. Add US3: Device code flow works (headless environments)
5. Add US4: Standalone login without gateway (reusability)
6. Polish: Documentation, examples, final CI pass

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently completable and testable
- Tests written first per Constitution III (Test-First)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Total tasks: 31
