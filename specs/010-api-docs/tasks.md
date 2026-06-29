# Tasks: API Documentation Site

**Input**: Design documents from `specs/010-api-docs/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (mdBook Scaffolding)

**Purpose**: Initialize mdBook project structure and CI pipeline

- [x] T001 Create mdBook configuration in docs/book.toml with title, language, build settings, and mdbook-admonish preprocessor
- [x] T002 Create navigation sidebar in docs/src/SUMMARY.md with all planned sections and pages
- [x] T003 [P] Create custom CSS theme in docs/theme/custom.css with Inter font (Google Fonts CDN), 18px base, 1.7 line-height, 800px max-width
- [x] T004 [P] Create GitHub Actions workflow in .github/workflows/docs.yml to build mdBook and deploy to GitHub Pages
- [x] T005 Create landing page in docs/src/introduction.md with project overview and links to getting started

**Checkpoint**: mdBook builds locally with `mdbook build docs/` and CI deploys placeholder site

---

## Phase 2: Foundational (No blocking prerequisites for this feature)

**Purpose**: This feature is documentation-only. No foundational phase needed since all content pages are independent markdown files.

---

## Phase 3: User Story 1 - Browse SDK Documentation (Priority: P1) + User Story 3 - Quick Start (Priority: P1)

**Goal**: Create the docs site infrastructure and getting started guide so developers can discover and onboard with the SDK.

**Independent Test**: Visit the deployed site, navigate all pages, verify search works, follow the getting started guide.

- [x] T006 [US1] [US3] Write getting started guide in docs/src/getting-started.md covering installation, gateway connection, sandbox creation, command execution, and cleanup

**Checkpoint**: A developer can follow the getting started guide end-to-end

---

## Phase 4: User Story 7 - Architecture Overview (Priority: P3)

**Goal**: Explain the SDK's internal structure so developers understand the sub-client pattern and proto isolation.

**Independent Test**: Navigate to architecture page, verify diagram and explanation of client hierarchy.

- [x] T007 [US7] Write architecture overview in docs/src/architecture.md with Client hierarchy diagram (text-based), sub-client pattern, and proto isolation explanation

**Checkpoint**: Architecture page explains the SDK design clearly

---

## Phase 5: User Story 2 - gRPC Mapping Hero Interfaces (Priority: P1)

**Goal**: Create detailed API reference pages for the three hero interfaces with side-by-side SDK/gRPC code blocks.

**Independent Test**: Each hero page shows every public method with SDK code and corresponding gRPC proto definition.

- [x] T008 [P] [US2] Write API overview page in docs/src/api/overview.md listing all 13 interfaces with accessor paths and brief descriptions
- [x] T009 [P] [US2] Write Sandboxes API reference in docs/src/api/sandboxes.md with side-by-side SDK/gRPC for Create, Get, List, Delete, WaitReady, Watch, GetLogs, AttachProvider, DetachProvider, ListProviders
- [x] T010 [P] [US2] Write Exec API reference in docs/src/api/exec.md with side-by-side SDK/gRPC for Run, Stream, Interactive
- [x] T011 [P] [US2] Write Providers API reference in docs/src/api/providers.md with side-by-side SDK/gRPC for Create, Get, List, Update, Delete, Ensure

**Checkpoint**: Three hero interface pages have complete side-by-side gRPC mapping

---

## Phase 6: User Story 2 - gRPC Mapping Standard Interfaces (Priority: P1)

**Goal**: Create API reference pages for the remaining 10 interfaces with reference tables.

**Independent Test**: Each page has a reference table mapping SDK methods to gRPC RPC names and proto files.

- [x] T012 [P] [US2] Write Profiles API reference in docs/src/api/profiles.md with reference table for List, Get, Import, Update, Lint, Delete
- [x] T013 [P] [US2] Write Refresh API reference in docs/src/api/refresh.md with reference table for GetStatus, Configure, Rotate, Delete
- [x] T014 [P] [US2] Write Services API reference in docs/src/api/services.md with reference table for Expose, Get, List, Delete
- [x] T015 [P] [US2] Write Files API reference in docs/src/api/files.md with reference table for Upload, Download
- [x] T016 [P] [US2] Write Health API reference in docs/src/api/health.md with reference table for Check
- [x] T017 [P] [US2] Write SSH API reference in docs/src/api/ssh.md with reference table for CreateSession, RevokeSession, Tunnel
- [x] T018 [P] [US2] Write TCP API reference in docs/src/api/tcp.md with reference table for Forward
- [x] T019 [P] [US2] Write Config API reference in docs/src/api/config.md with reference table for GetSandbox, GetGateway, Update
- [x] T020 [P] [US2] Write Policy API reference in docs/src/api/policy.md with reference table for GetDraft, ApproveDraftChunk, RejectDraftChunk, ApproveAllDraftChunks, ClearDraftChunks, EditDraftChunk, UndoDraftChunk, GetDraftHistory, GetStatus, List

**Checkpoint**: All 13 interfaces have dedicated API reference pages

---

## Phase 7: User Story 4 - Error Handling Guide (Priority: P2)

**Goal**: Document error handling patterns for production code.

**Independent Test**: Error handling page covers all typed error checks with compilable examples.

- [x] T021 [US4] Write error handling guide in docs/src/error-handling.md covering StatusError, IsNotFound, IsAlreadyExists, IsConflict, IsUnavailable, IsUnimplemented, and retry patterns with backoff

**Checkpoint**: Error handling guide is complete with all typed checks

---

## Phase 8: User Story 5 - Testing Guide (Priority: P2)

**Goal**: Document fake client usage for operator test suites.

**Independent Test**: Testing page shows fake client creation, fixture seeding, watch events, and health simulation.

- [x] T022 [US5] Write testing guide in docs/src/testing.md covering fake client creation, AddSandbox/AddProvider fixture seeding, watch event testing, StopOnTerminal, and WithHealthResult

**Checkpoint**: Testing guide shows complete fake client workflows

---

## Phase 9: User Story 6 - Example* Test Functions (Priority: P2)

**Goal**: Add compilable Example* functions that render on pkg.go.dev.

**Independent Test**: `go test ./openshell/v1/...` passes with all Example functions compiling.

- [x] T023 [P] [US6] Create Example* test file in openshell/v1/example_test.go with ExampleClient_Sandboxes, ExampleClient_Providers, ExampleClient_Health, ExampleIsNotFound, ExampleIsAlreadyExists, ExampleIsUnavailable
- [x] T024 [P] [US6] Create fake client Example* test file in openshell/v1/example_fake_test.go with ExampleNewClient_addSandbox, ExampleNewClient_addProvider, ExampleNewClient_withHealthResult, ExampleNewClient_watchEvents, ExampleNewClient_stopOnTerminal

**Checkpoint**: `go test ./openshell/v1/...` passes, Example functions appear in `go doc`

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Final quality pass across all documentation

- [x] T025 Verify all Go code examples in docs/ compile by extracting and running them
- [x] T026 Run `mdbook build docs/` and fix any build warnings or broken links (note: mdbook-admonish preprocessor fails on `:::code-group` directives in hero pages, pre-existing issue; build succeeds without admonish)
- [x] T027 Update README.md to add Documentation section linking to the GitHub Pages site

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **No Foundational phase**: All content is independent markdown
- **User Stories (Phase 3-9)**: All depend on Setup (Phase 1) for SUMMARY.md structure
  - Stories can proceed in parallel after Phase 1
  - Priority order recommended: P1 stories first, then P2, then P3
- **Polish (Phase 10)**: Depends on all content phases completing

### User Story Dependencies

- **US1 + US3 (Browse + Quick Start)**: Start after Phase 1. No dependencies on other stories.
- **US7 (Architecture)**: Start after Phase 1. No dependencies on other stories.
- **US2 (gRPC Mapping)**: Start after Phase 1. Phases 5 and 6 can run in parallel. No dependencies on other stories.
- **US4 (Error Handling)**: Start after Phase 1. No dependencies on other stories.
- **US5 (Testing)**: Start after Phase 1. No dependencies on other stories.
- **US6 (Example* Functions)**: Start after Phase 1. Independent of docs site content (Go test files).

### Parallel Opportunities

- T003, T004 can run in parallel with T001/T002
- All Phase 5 tasks (T008-T011) can run in parallel
- All Phase 6 tasks (T012-T020) can run in parallel
- T023 and T024 can run in parallel
- Phases 3-9 can all run in parallel after Phase 1 completes

---

## Parallel Example: Phase 6 (Standard Interfaces)

```
# All 9 standard interface pages can be written simultaneously:
Task T012: Profiles API reference
Task T013: Refresh API reference
Task T014: Services API reference
Task T015: Files API reference
Task T016: Health API reference
Task T017: SSH API reference
Task T018: TCP API reference
Task T019: Config API reference
Task T020: Policy API reference
```

---

## Implementation Strategy

### MVP First (Phase 1 + Phase 3)

1. Complete Phase 1: mdBook scaffolding and CI
2. Complete Phase 3: Getting started guide
3. **STOP and VALIDATE**: Site is deployed, searchable, and has a working guide
4. Deploy as initial docs site

### Incremental Delivery

1. Setup + Getting Started -> Minimal docs site live
2. Add Architecture -> SDK design explained
3. Add Hero API Reference -> Core interfaces documented with gRPC mapping
4. Add Standard API Reference -> All 13 interfaces documented
5. Add Error Handling + Testing guides -> Production usage documented
6. Add Example* functions -> pkg.go.dev enhanced
7. Polish -> Final quality pass

---

## Notes

- All docs content is markdown, no Go compilation needed for docs tasks
- Example* test functions (Phase 9) require Go compilation and use the fake client
- The `docs/go-sdk-proposal.md` file is preserved, not overwritten
- gRPC RPC names are extracted from `proto/openshell.proto`
- SDK method signatures are extracted from `openshell/v1/*.go` interface definitions
