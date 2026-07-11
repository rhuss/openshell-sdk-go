# Tasks: Upstream PR Preparation

**Input**: Design documents from `/specs/019-upstream-pr/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, quickstart.md

**Tests**: No test tasks generated (existing tests validate post-rewrite correctness; no new SDK code is written).

**Organization**: Tasks grouped by user story. US1 (Module Path Migration) is the MVP and must complete before other stories. US3/US4/US5 can execute in parallel after US1 completes. US6 depends on all other stories.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US6)
- Paths are relative to the `rhuss/OpenShell` fork unless noted otherwise

---

## Global Constraints

These constraints apply to ALL tasks and are inherited implicitly:

- **Module path**: Target module is `github.com/NVIDIA/OpenShell/sdk/go`. Old module is `github.com/rhuss/openshell-sdk-go`.
- **SPDX headers**: Every `.go` file must have `SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.` and `SPDX-License-Identifier: Apache-2.0`.
- **DCO sign-off**: The final commit must include `Signed-off-by:` trailer.
- **No proto file modifications**: `go_package` is supplied via protoc M_FLAGS, not by editing `.proto` files.
- **Work location**: All file operations happen in the `rhuss/OpenShell` fork, not in the SDK development repo.

---

## Phase 1: Setup

**Purpose**: Prepare the working branch in the upstream fork

- [ ] T001 Sync rhuss/OpenShell fork with NVIDIA/OpenShell main branch
- [ ] T002 Create `go-sdk` branch in rhuss/OpenShell fork based on upstream main
- [ ] T003 Create `sdk/go/` directory structure in the fork working branch

**Checkpoint**: Clean working branch exists in fork, ready for SDK content

---

## Phase 2: User Story 1 - Module Path Migration (Priority: P1) MVP

**Goal**: Rewrite Go module path from development repo to upstream path so all imports compile under the new module identity.

**Independent Test**: `go build ./...` and `go test ./...` succeed in `sdk/go/` with zero references to the old module path.

### Implementation for User Story 1

- [ ] T004 [US1] Copy SDK source to fork: `openshell/` to `sdk/go/openshell/`, `proto/` to `sdk/go/proto/`, and root build files (`go.mod`, `go.sum`, `Makefile`, `mise.toml`) to `sdk/go/`. Exclude `examples/`, `brainstorm/`, `.specify/`, `.claude/`, `docs/`, `CLAUDE.md`, `AGENTS.md`.
- [ ] T005 [US1] Rewrite module path from `github.com/rhuss/openshell-sdk-go` to `github.com/NVIDIA/OpenShell/sdk/go` in `sdk/go/go.mod` and all `.go` files under `sdk/go/`
- [ ] T006 [US1] Update MODULE variable in `sdk/go/mise.toml` proto generation tasks (`proto:gen`, `proto:check`) to use new module path
- [ ] T007 [US1] Regenerate proto bindings in `sdk/go/` using updated mise proto:gen task
- [ ] T008 [US1] Run `go mod tidy` in `sdk/go/` to regenerate `go.sum`
- [ ] T009 [US1] Verify `go build ./...` passes in `sdk/go/`
- [ ] T010 [US1] Verify `go test ./...` passes in `sdk/go/`
- [ ] T011 [US1] Verify zero remaining references to `github.com/rhuss/openshell-sdk-go` in `sdk/go/` using grep

**Checkpoint**: SDK compiles and all tests pass under new module path. MVP complete.

**Interfaces for downstream phases**: After US1, the following paths exist and are stable:
- `sdk/go/go.mod` (module `github.com/NVIDIA/OpenShell/sdk/go`)
- `sdk/go/openshell/v1/` (all SDK packages with rewritten imports)
- `sdk/go/proto/` (regenerated `.pb.go` files)
- `sdk/go/mise.toml` (proto tasks with updated MODULE variable)

---

## Phase 3: User Story 2 - Example Extraction (Priority: P1)

**Goal**: Extract the oshell TUI example into a standalone public repository so the upstream PR contains only library code.

**Independent Test**: The examples repo compiles and its `go.mod` depends on the upstream SDK module path.

### Implementation for User Story 2

- [ ] T012 [US2] Create public GitHub repository `rhuss/openshell-examples` with description "Examples for the OpenShell Go SDK"
- [ ] T013 [US2] Initialize examples repo with `go.mod` declaring module `github.com/rhuss/openshell-examples` and dependency on `github.com/NVIDIA/OpenShell/sdk/go`
- [ ] T014 [US2] Copy `examples/oshell/` source files (connection.go, demo.go, README.md) from SDK dev repo to examples repo root
- [ ] T015 [US2] Rewrite import paths in examples repo from `github.com/rhuss/openshell-sdk-go` to `github.com/NVIDIA/OpenShell/sdk/go`
- [ ] T016 [US2] Add temporary `replace` directive in examples `go.mod` pointing to local SDK copy for pre-merge development
- [ ] T017 [US2] Verify examples repo compiles with `go build ./...`
- [ ] T018 [US2] Push examples repo to GitHub

**Checkpoint**: Examples repo exists at github.com/rhuss/openshell-examples and compiles independently.

---

## Phase 4: User Story 3 - Fern Documentation Integration (Priority: P2)

**Goal**: Create concise Go SDK documentation as Fern MDX pages integrated into the OpenShell docs site navigation.

**Independent Test**: `fern check` passes (or `mise run docs`) and Go SDK pages appear in navigation.

### Implementation for User Story 3

- [ ] T019 [P] [US3] Create `docs/sdks/go/getting-started.mdx` covering SDK installation, basic gateway connection, and first API call with code snippets using actual SDK function signatures
- [ ] T020 [P] [US3] Create `docs/sdks/go/architecture.mdx` covering module structure, gRPC transport layer, proto isolation pattern, and package hierarchy
- [ ] T021 [P] [US3] Create `docs/sdks/go/error-handling.mdx` covering SDK error types, gRPC status code mapping, and retry patterns with code snippets
- [ ] T022 [P] [US3] Create `docs/sdks/go/authentication.mdx` covering OIDC flow, token refresh, gateway authentication options, and configuration examples
- [ ] T023 [US3] Add "SDKs" navigation section with Go subfolder to `docs/index.yml` following existing section pattern (see research.md R3 for YAML structure)
- [ ] T024 [US3] Verify Fern docs build passes with Go SDK pages included (run `mise run docs` or `fern check` in fork)

**Checkpoint**: Go SDK docs pages render in Fern navigation under "SDKs > Go SDK".

**Interfaces for US6**: `docs/sdks/go/*.mdx` (4 files) and `docs/index.yml` (modified with SDKs section).

---

## Phase 5: User Story 4 - Proto Generation Automation (Priority: P2)

**Goal**: Add a mise task and CI validation step that regenerates Go protobuf bindings and fails if committed files are stale.

**Independent Test**: Modify a proto file, run the CI check, and observe it detects stale generated files.

### Implementation for User Story 4

- [ ] T025 [P] [US4] Create `tasks/go.toml` with `go:proto` task that runs proto generation from `sdk/go/` directory (delegate to `sdk/go/mise.toml` proto:gen task, adapting paths for monorepo context)
- [ ] T026 [P] [US4] Add `go` job to `.github/workflows/branch-checks.yml` following Rust/Python job pattern: checkout, mise install, lint, build, test, proto:check (see research.md R4 for job structure)
- [ ] T027 [US4] Verify proto freshness check detects intentionally stale `.pb.go` files by modifying a proto file and running the check

**Checkpoint**: Go CI job runs lint, build, test, and proto check in branch-checks workflow.

**Interfaces for US6**: `tasks/go.toml` (new file) and `.github/workflows/branch-checks.yml` (modified with Go job).

---

## Phase 6: User Story 5 - Spec Artifact Scoping (Priority: P3)

**Goal**: Include design specs in the PR while excluding internal development artifacts.

**Independent Test**: `sdk/go/specs/` exists in the PR diff; `brainstorm/`, `.specify/`, `.claude/`, `CLAUDE.md`, `AGENTS.md` do not.

### Implementation for User Story 5

- [ ] T028 [US5] Copy `specs/` directory to `sdk/go/specs/` in fork, including all numbered spec directories with their plan.md, spec.md, and research.md files
- [ ] T029 [US5] Verify no internal artifacts exist in the fork working branch: grep for `.specify/`, `.claude/`, `CLAUDE.md`, `AGENTS.md`, `brainstorm/` in the file tree

**Checkpoint**: Spec artifacts are included under `sdk/go/specs/`, all internal artifacts excluded.

---

## Phase 7: User Story 6 - Draft PR Creation (Priority: P3)

**Goal**: Open a draft PR against NVIDIA/OpenShell presenting the complete Go SDK contribution.

**Independent Test**: Draft PR exists on GitHub, references issue #2044, contains expected file tree under `sdk/go/`.

### Implementation for User Story 6

- [ ] T030 [US6] Write PR description in a local file: include SDK overview, feature list, file tree summary, reference to issue [#2044](https://github.com/NVIDIA/OpenShell/issues/2044), spec-driven development methodology note, and question about `specs/` directory retention (see research.md R7 for commit message template)
- [ ] T031 [US6] Squash all commits on the `go-sdk` branch into a single commit with DCO sign-off (`Signed-off-by: Roland Huss <roland@jolokia.org>`) and comprehensive commit message
- [ ] T032 [US6] Open draft PR against `NVIDIA/OpenShell` main branch using `gh pr create --draft`
- [ ] T033 [US6] Verify PR file tree contains `sdk/go/` (source + proto + specs), `docs/sdks/go/` (MDX pages), `tasks/go.toml`, and modified `docs/index.yml` and `.github/workflows/branch-checks.yml`

**Checkpoint**: Draft PR is live on GitHub with correct references, file tree, and description.

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final validation across all stories

- [ ] T034 Run quickstart.md verification checklist end-to-end
- [ ] T035 Final review of PR description for completeness and clarity
- [ ] T036 Verify SPDX license headers present on all new `.go` files in fork

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies, start immediately
- **US1 (Phase 2)**: Depends on Setup. BLOCKS all other user stories.
- **US2 (Phase 3)**: Depends on US1 (needs new module path for examples repo dependency)
- **US3 (Phase 4)**: Depends on US1 (needs SDK API for code examples). Can parallel with US2, US4, US5.
- **US4 (Phase 5)**: Depends on US1 (needs proto tasks with new module path). Can parallel with US2, US3, US5.
- **US5 (Phase 6)**: Depends on US1 (specs copied after SDK source). Can parallel with US2, US3, US4.
- **US6 (Phase 7)**: Depends on US1, US2, US3, US4, US5 (all must complete)
- **Polish (Phase 8)**: Depends on US6

### User Story Dependencies

```text
Setup ──► US1 (MVP) ──┬──► US2 ──────────────┐
                       ├──► US3 (parallel) ──►│
                       ├──► US4 (parallel) ──►├──► US6 ──► Polish
                       └──► US5 (parallel) ──►│
```

### Within Each User Story

- Copy/create before rewrite
- Rewrite before regenerate/tidy
- Regenerate before verify
- All verification tasks run last in their phase

### Parallel Opportunities

After US1 completes, three stories can execute in parallel:
- **US3** (Fern docs): Touches `docs/sdks/go/` and `docs/index.yml`
- **US4** (Proto CI): Touches `tasks/go.toml` and `.github/workflows/branch-checks.yml`
- **US5** (Spec scoping): Touches `sdk/go/specs/`

No file conflicts between these three stories.

Within US3, all four MDX pages (T019-T022) can be written in parallel since they are independent files.

Within US4, the task file (T025) and CI workflow (T026) can be written in parallel.

---

## Parallel Example: Post-US1 Fan-Out

```bash
# After US1 checkpoint passes, launch three stories in parallel:

# Story 3 (docs):
Task: "Create docs/sdks/go/getting-started.mdx"
Task: "Create docs/sdks/go/architecture.mdx"
Task: "Create docs/sdks/go/error-handling.mdx"
Task: "Create docs/sdks/go/authentication.mdx"

# Story 4 (CI):
Task: "Create tasks/go.toml with go:proto task"
Task: "Add Go job to .github/workflows/branch-checks.yml"

# Story 5 (specs):
Task: "Copy specs/ to sdk/go/specs/"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: US1 Module Path Migration (T004-T011)
3. **STOP and VALIDATE**: `go build ./...` and `go test ./...` pass, zero old module references
4. SDK is compilable under upstream path

### Incremental Delivery

1. Setup + US1 = SDK compiles under upstream path (MVP)
2. Add US2 = Examples extracted to separate repo
3. Add US3 + US4 + US5 in parallel = Docs, CI, specs ready
4. US6 = Draft PR opened on GitHub
5. Polish = Final validation

### Parallel Team Strategy

With multiple developers after US1:
- Developer A: US2 (Example Extraction) + US5 (Spec Scoping)
- Developer B: US3 (Fern Documentation)
- Developer C: US4 (Proto CI Automation)
- Everyone: US6 (PR assembly after all stories merge)

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- US1 is both the first user story AND the foundational blocker for all others
- US3, US4, US5 have zero file conflicts and can run fully in parallel
- All work happens in the rhuss/OpenShell fork, not the SDK development repo
- The SDK development repo (`openshell-sdk-go`) is read-only during this process (source of truth for copy)
