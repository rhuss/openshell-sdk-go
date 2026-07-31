# Tasks: Proto Sync from Upstream PR #2445

**Input**: Design documents from `specs/020-proto-sync-pr2445/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md

**Tests**: Not explicitly requested. Existing tests serve as regression verification.

**Organization**: Tasks follow the proto sync pipeline: copy, configure, generate, verify.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Prepare the build configuration for the new inference.proto

- [x] T001 [US1] Update `buf.gen.yaml` to add `inference.proto` to inputs.paths and add `Minference.proto` import mappings to both plugins

---

## Phase 2: Proto File Sync (US1 - Sync Proto Definitions, Priority: P1)

**Goal**: Copy all 5 SDK-relevant proto files from upstream to match byte-for-byte

**Independent Test**: Each proto file in `proto/` matches its upstream counterpart at `/Users/rhuss/Work/projects/OpenShell/proto/`

- [x] T002 [P] [US1] Copy `openshell.proto` from upstream `/Users/rhuss/Work/projects/OpenShell/proto/openshell.proto` to `proto/openshell.proto`
- [x] T003 [P] [US1] Copy `inference.proto` from upstream `/Users/rhuss/Work/projects/OpenShell/proto/inference.proto` to `proto/inference.proto`
- [x] T004 [P] [US1] Copy `options.proto` from upstream `/Users/rhuss/Work/projects/OpenShell/proto/options.proto` to `proto/options.proto`
- [x] T005 [P] [US1] Copy `datamodel.proto` from upstream `/Users/rhuss/Work/projects/OpenShell/proto/datamodel.proto` to `proto/datamodel.proto`
- [x] T006 [P] [US1] Copy `sandbox.proto` from upstream `/Users/rhuss/Work/projects/OpenShell/proto/sandbox.proto` to `proto/sandbox.proto`

---

## Phase 3: Code Generation (US1 + US3 - Generate Bindings, Priority: P1)

**Goal**: Regenerate all Go bindings including new inference package

**Independent Test**: `go build ./proto/...` compiles with zero errors

- [x] T007 [US1] Run `mise run proto:gen` to regenerate all Go bindings (cleans old files first, then runs `buf generate`)
- [x] T008 [US1] Verify generated `proto/inferencev1/` directory exists with `inference.pb.go` and `inference_grpc.pb.go`

---

## Phase 4: Verification (US2 - Preserve Existing Functionality, Priority: P1)

**Goal**: Confirm all existing tests pass and CI is green

**Independent Test**: `make ci` passes with zero failures

- [x] T009 [US2] Run `make build` to verify all packages compile (including new inferencev1)
- [x] T010 [US2] Run `make lint` to verify lint checks pass
- [x] T011 [US2] Run `make test` to verify all existing unit tests pass
- [x] T012 [US2] Run `make ci` as final full pipeline verification (docs:check failure is pre-existing on main, not introduced by proto sync; build+lint+test+proto:check all pass)

---

## Dependencies

```text
T001 (buf config) --> T007 (proto:gen)
T002-T006 (proto copy, parallel) --> T007 (proto:gen)
T007 (proto:gen) --> T008 (verify inference pkg)
T007 (proto:gen) --> T009-T012 (verification)
```

## Parallel Execution

- **T002-T006**: All proto file copies can run in parallel (independent files)
- **T009-T010**: Build and lint can run in parallel after generation

## Implementation Strategy

**MVP**: T001-T009 (config + copy + generate + build passes)
**Complete**: T001-T012 (full CI pipeline green)

All user stories are tightly coupled in this feature since they share the same proto sync + generation pipeline. US3 (inference stubs) is automatically satisfied by US1's generation step.

## Summary

- **Total tasks**: 12
- **US1 (Sync Proto Definitions)**: 8 tasks (T001-T008)
- **US2 (Preserve Existing Functionality)**: 4 tasks (T009-T012)
- **US3 (Generate Inference Stubs)**: Satisfied by T007-T008
- **Parallel opportunities**: T002-T006 (5 file copies)
