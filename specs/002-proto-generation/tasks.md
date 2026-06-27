# Tasks: Proto Generation Pipeline

**Input**: Design documents from `specs/002-proto-generation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

**Tests**: Test-First is a constitution principle. Each task includes validation
steps. The `proto:check` CI validation task (US3) serves as the primary
automated test for generation correctness.

**Organization**: Tasks grouped by user story. US1 and US2 are both P1 and
foundational (sync must exist before gen). US3 and US4 build on top.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3, US4)
- Exact file paths included in descriptions

---

## Phase 1: Setup (Tool Pinning & Dependencies)

**Purpose**: Pin protoc and plugin versions, add Go module dependencies

- [ ] T001 Add protoc, protoc-gen-go, and protoc-gen-go-grpc to mise.toml tools section
- [ ] T002 Add google.golang.org/protobuf and google.golang.org/grpc dependencies to go.mod via `go get`
- [ ] T003 Update .golangci.yml to exclude proto/ subdirectories from goheader linter

---

## Phase 2: User Story 2 - Sync Proto Files from Upstream (Priority: P1)

**Goal**: Copy the three needed proto files from upstream and record the version

**Independent Test**: Run `mise run proto:sync`, verify 3 files copied and
UPSTREAM_VERSION contains a commit SHA

### Implementation for User Story 2

- [ ] T004 [US2] Create proto/ directory and add proto:sync mise task in mise.toml
- [ ] T005 [US2] Implement proto:sync shell script that copies openshell.proto, datamodel.proto, sandbox.proto from upstream path (default: ../OpenShell/proto/) to proto/
- [ ] T006 [US2] Add upstream commit SHA recording to proto:sync (writes to proto/UPSTREAM_VERSION, or "unknown" if not a git repo)
- [ ] T007 [US2] Add configurable upstream path support via UPSTREAM_PATH env var to proto:sync
- [ ] T008 [US2] Add error handling to proto:sync for missing upstream path
- [ ] T009 [US2] Run proto:sync and verify all 3 proto files are copied with correct content, and UPSTREAM_VERSION contains a valid SHA

**Checkpoint**: `mise run proto:sync` works, proto/ contains 3 proto files and UPSTREAM_VERSION

---

## Phase 3: User Story 1 - Generate Go Bindings (Priority: P1)

**Goal**: Generate compilable Go packages from the synced proto files

**Independent Test**: Run `mise run proto:gen` then `go build ./proto/...`

**Depends on**: Phase 2 (proto files must exist before generation)

### Implementation for User Story 1

- [ ] T010 [US1] Add proto:gen mise task in mise.toml
- [ ] T011 [US1] Implement proto:gen shell script that runs protoc with --go_out, --go-grpc_out, and all --go_opt=M / --go-grpc_opt=M flags per data-model.md mapping
- [ ] T012 [US1] Add output directory creation (proto/openshellv1/, proto/datamodelv1/, proto/sandboxv1/) to proto:gen
- [ ] T013 [US1] Add tool availability checks to proto:gen (protoc, protoc-gen-go, protoc-gen-go-grpc) with helpful error messages
- [ ] T014 [US1] Run proto:gen and verify generated .pb.go files exist in correct packages
- [ ] T015 [US1] Verify `go build ./proto/...` compiles all generated packages without errors
- [ ] T016 [US1] Verify cross-package imports resolve correctly (openshellv1 imports datamodelv1 and sandboxv1)

**Checkpoint**: `mise run proto:gen` produces compilable Go packages, `go build ./proto/...` passes

---

## Phase 4: User Story 3 - CI Validation (Priority: P2)

**Goal**: Detect stale or manually edited generated code in CI

**Independent Test**: Edit a .pb.go file by hand, run proto:check, verify it fails

**Depends on**: Phase 3 (generated files must exist)

### Implementation for User Story 3

- [ ] T017 [US3] Add proto:check mise task in mise.toml
- [ ] T018 [US3] Implement proto:check shell script that generates to a temp directory and diffs against committed files
- [ ] T019 [US3] Add proto:check to CI workflow in .github/workflows/ci.yml
- [ ] T020 [US3] Verify proto:check passes with unmodified generated files (exit 0)
- [ ] T021 [US3] Verify proto:check fails when a .pb.go file is manually edited (exit 1 with diff)

**Checkpoint**: `mise run proto:check` detects staleness, CI runs it alongside lint/build/test

---

## Phase 5: User Story 4 - Clean Generated Files (Priority: P3)

**Goal**: Remove all generated .pb.go files while preserving sources

**Independent Test**: Run proto:clean, verify .pb.go files removed and .proto files preserved

**Depends on**: Phase 3 (generated files must exist to clean)

### Implementation for User Story 4

- [ ] T022 [US4] Add proto:clean mise task in mise.toml
- [ ] T023 [US4] Implement proto:clean shell script that removes *.pb.go files recursively under proto/ and removes empty generated subdirectories
- [ ] T024 [US4] Verify proto:clean preserves .proto files and UPSTREAM_VERSION
- [ ] T025 [US4] Verify proto:clean removes all .pb.go and _grpc.pb.go files

**Checkpoint**: `mise run proto:clean` cleans generated files without affecting sources

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and documentation

- [ ] T026 [P] Update CLAUDE.md project structure section to include proto/ layout
- [ ] T027 [P] Commit all generated .pb.go files and proto source files
- [ ] T028 Run `make ci` to verify full CI pipeline passes (lint + build + test + proto:check)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies, start immediately
- **Phase 2 (US2: Sync)**: Depends on Phase 1 (tools must be pinned)
- **Phase 3 (US1: Gen)**: Depends on Phase 2 (proto files must exist)
- **Phase 4 (US3: Check)**: Depends on Phase 3 (generated files must exist)
- **Phase 5 (US4: Clean)**: Depends on Phase 3 (generated files must exist)
- **Phase 6 (Polish)**: Depends on Phases 3-5

### User Story Dependencies

- **US2 (Sync)**: First, no story dependencies. Proto files must exist before anything else.
- **US1 (Gen)**: Depends on US2. Cannot generate without proto source files.
- **US3 (Check)**: Depends on US1. Validates generation output.
- **US4 (Clean)**: Depends on US1. Removes generation output. Can run parallel with US3.

### Parallel Opportunities

- T001, T002, T003 (Setup) can run in parallel
- T026, T027 (Polish) can run in parallel
- US3 and US4 can run in parallel after US1 completes

---

## Implementation Strategy

### MVP First (US2 + US1)

1. Complete Phase 1: Setup (tool pins, deps, linter config)
2. Complete Phase 2: US2 Sync (copy protos, record version)
3. Complete Phase 3: US1 Gen (generate Go bindings, verify compilation)
4. **STOP and VALIDATE**: `go build ./proto/...` passes
5. This is the minimum needed for the core SDK (brainstorm 004) to proceed

### Incremental Delivery

1. Setup + US2 (Sync) + US1 (Gen) -> Proto bindings available for SDK development
2. Add US3 (Check) -> CI catches stale generated code
3. Add US4 (Clean) -> Developer convenience
4. Polish -> Documentation and full CI integration

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story
- US2 (Sync) is sequenced before US1 (Gen) despite both being P1, because gen requires proto files to exist
- Tool version pinning (T001) is critical for reproducibility (Constitution: Minimal Dependencies)
- Generated .pb.go files are committed to the repo per spec decision
- The goheader linter exclusion (T003) prevents false failures on generated files that have protoc-generated headers instead of SPDX headers
