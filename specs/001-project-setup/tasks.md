# Tasks: Project Setup and Build Infrastructure

**Input**: Design documents from `specs/001-project-setup/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, module definition, and licensing

- [ ] T001 Initialize Go module with `go mod init github.com/rhuss/openshell-sdk-go` in go.mod
- [ ] T002 [P] Create Apache-2.0 LICENSE file at repository root
- [ ] T003 [P] Create .gitignore with Go binaries, coverage files, IDE artifacts, and .specify ignores

---

## Phase 2: Foundational (Build Tooling)

**Purpose**: Build infrastructure that MUST be complete before any code tasks

**CRITICAL**: No user story work can begin until this phase is complete

- [ ] T004 Create mise.toml with Go 1.23, golangci-lint 1.64 tool pins and task definitions (test, test:integration, lint, fmt, build, ci)
- [ ] T005 [P] Create Makefile as thin shim delegating to mise tasks with availability check and install hint
- [ ] T006 [P] Create .golangci.yml with govet, errcheck, staticcheck, unused, gosimple, ineffassign, revive, goimports, and goheader linters configured with SPDX template

**Checkpoint**: Build tooling ready. `make lint` and `make test` targets exist (tests will fail until stub code is created).

---

## Phase 3: User Story 1 - Clone and Build (Priority: P1) MVP

**Goal**: A developer clones the repo, runs `make test` and `make lint`, both pass.

**Independent Test**: Run `make test` and `make lint` from a fresh clone, verify exit 0.

### Implementation for User Story 1

- [ ] T007 [US1] Create openshell/client.go with Client struct, Dial(address string) function returning (*Client, error), and Close() method with SPDX headers
- [ ] T008 [US1] Create openshell/client_test.go with TestDialEmptyAddress test validating Dial("") returns error, with SPDX headers
- [ ] T009 [US1] Run `go mod tidy` to add testify dependency and generate go.sum
- [ ] T010 [US1] Verify `make test` passes with coverage output
- [ ] T011 [US1] Verify `make lint` passes with zero violations

**Checkpoint**: User Story 1 complete. `make test` and `make lint` both pass.

---

## Phase 4: User Story 2 - CI Validates PRs (Priority: P1)

**Goal**: GitHub Actions runs lint, test, and build on PRs and pushes to main.

**Independent Test**: Push a branch, open a PR, verify CI runs and reports status.

### Implementation for User Story 2

- [ ] T012 [US2] Create .github/workflows/ci.yml with lint, test, and build jobs triggered on PR and push to main, using mise for Go version consistency

**Checkpoint**: CI workflow ready. Will validate on first PR push.

---

## Phase 5: User Story 3 - Agentic Development (Priority: P2)

**Goal**: CLAUDE.md contains working build commands, CONTRIBUTING.md guides contributors.

**Independent Test**: Read CLAUDE.md, verify commands match actual mise tasks.

### Implementation for User Story 3

- [ ] T013 [P] [US3] Create README.md with project description, status, quick start, and license info
- [ ] T014 [P] [US3] Create CONTRIBUTING.md with prerequisites (Go, mise), setup steps, build commands, test commands, commit conventions, and DCO sign-off
- [ ] T015 [US3] Update CLAUDE.md with build, test, lint, ci commands and project conventions

**Checkpoint**: Documentation complete. Contributors can onboard from CONTRIBUTING.md.

---

## Phase 6: User Story 5 - Project Constitution (Priority: P3)

**Goal**: Constitution documents five governing principles with actionable descriptions.

**Independent Test**: Read constitution file, verify all five principles are present and concrete.

### Implementation for User Story 5

- [ ] T016 [US5] Fill in .specify/memory/constitution.md with five principles: proto isolation, idiomatic Go, test-first, upstream tracking, minimal dependencies

**Checkpoint**: Constitution ready. All governance documentation in place.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all user stories

- [ ] T017 Verify all .go files have SPDX license headers by running `make lint`
- [ ] T018 Verify `make ci` runs full pipeline (lint, build, test) successfully
- [ ] T019 Verify .gitignore covers all generated artifacts (coverage.out, binaries)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (go.mod must exist for mise tasks)
- **US1 (Phase 3)**: Depends on Phase 2 (mise tasks and linter config must exist)
- **US2 (Phase 4)**: Depends on Phase 3 (CI needs code to validate)
- **US3 (Phase 5)**: Depends on Phase 2 (needs to reference real build commands)
- **US5 (Phase 6)**: No code dependencies, can run after Phase 1
- **Polish (Phase 7)**: Depends on all prior phases

### User Story Dependencies

- **US1 (P1)**: Blocked by Foundational. Core MVP.
- **US2 (P1)**: Blocked by US1 (CI needs passing tests to validate).
- **US3 (P2)**: Can start after Foundational. Parallel with US1.
- **US4 (P2)**: Integrated into US1 (SPDX headers on stub files, linter config in Phase 2). No separate phase needed.
- **US5 (P3)**: Independent. Can run in parallel with anything after Phase 1.

### Parallel Opportunities

- T002 and T003 (LICENSE, .gitignore) can run in parallel
- T005 and T006 (Makefile, .golangci.yml) can run in parallel
- T013 and T014 (README.md, CONTRIBUTING.md) can run in parallel
- US5 (Phase 6) can run in parallel with US1/US2/US3

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (go.mod, LICENSE, .gitignore)
2. Complete Phase 2: Foundational (mise.toml, Makefile, .golangci.yml)
3. Complete Phase 3: User Story 1 (stub code + tests)
4. **STOP and VALIDATE**: `make test` and `make lint` pass
5. Foundation is usable

### Incremental Delivery

1. Setup + Foundational -> Build tooling ready
2. US1 -> `make test` passes (MVP)
3. US2 -> CI validates PRs
4. US3 -> Documentation complete
5. US5 -> Constitution in place
6. Polish -> All cross-cutting checks pass

---

## Notes

- US4 (License Compliance) is handled across Phases 1-3: LICENSE file in Phase 1, go-header linter in Phase 2, SPDX headers on stub files in Phase 3
- Total tasks: 19
- Parallel opportunities: T002+T003, T005+T006, T013+T014, Phase 6 parallel with Phases 3-5
- All tasks include exact file paths for direct implementation
