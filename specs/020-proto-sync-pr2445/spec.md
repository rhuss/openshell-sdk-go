# Feature Specification: Proto Sync from Upstream PR #2445

**Feature Branch**: `020-proto-sync-pr2445`
**Created**: 2026-07-31
**Status**: Draft
**Input**: Sync SDK proto files with upstream OpenShell after PR #2445 merged

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Sync Proto Definitions (Priority: P1)

As an SDK maintainer, I need the SDK's proto definitions to match the upstream OpenShell repository so that generated Go bindings accurately reflect the current API surface, including authorization annotations, workspace scoping fields, and the new inference service.

**Why this priority**: Without synced protos, the SDK cannot implement workspace scoping (issues #32, #33), GatewayInfo (#34), or the inference client. This is the foundational prerequisite for all upcoming feature work.

**Independent Test**: Run `make build` and `make test` after proto sync and code generation. All existing tests pass, and the generated Go bindings compile without errors.

**Acceptance Scenarios**:

1. **Given** the SDK proto directory contains outdated proto files, **When** all 5 proto files are copied from upstream and code generation runs, **Then** the generated Go bindings compile successfully with no errors.
2. **Given** the SDK is missing `inference.proto` entirely, **When** `inference.proto` is added from upstream and code generation runs, **Then** a new `inferencev1/` Go package is generated with valid bindings for all 4 inference RPCs.
3. **Given** existing converter tests reference proto fields, **When** new fields are added to proto messages, **Then** all existing tests continue to pass (new fields have zero-value defaults).

---

### User Story 2 - Preserve Existing SDK Functionality (Priority: P1)

As an SDK consumer, I need existing functionality to remain unbroken after the proto sync so that my applications continue to work without code changes.

**Why this priority**: Breaking existing consumers would block adoption. The proto sync must be purely additive from the SDK user's perspective.

**Independent Test**: Run the full test suite (`make ci`) and verify zero regressions.

**Acceptance Scenarios**:

1. **Given** the SDK has existing unit tests for converters and client code, **When** the proto sync is complete, **Then** `make ci` passes with no test failures.
2. **Given** the SDK's public API types are defined independently from proto types, **When** proto messages gain new fields, **Then** the public API types remain unchanged (new fields are not yet exposed).

---

### User Story 3 - Generate Inference Service Stubs (Priority: P2)

As an SDK developer planning the inference client, I need the generated Go stubs for the inference service available in the codebase so that inference client implementation can begin in a subsequent PR.

**Why this priority**: The inference client is a planned feature. Having stubs available now unblocks that work, but no client code is written in this PR.

**Independent Test**: Verify that `inferencev1/` package exists, contains generated `.pb.go` and `_grpc.pb.go` files, and compiles successfully.

**Acceptance Scenarios**:

1. **Given** `inference.proto` is copied from upstream, **When** code generation runs, **Then** the `inferencev1/` package contains at least `inference.pb.go` and `inference_grpc.pb.go`.
2. **Given** the generated inference stubs exist, **When** `go build ./...` runs, **Then** the inference package compiles with no import errors.

### Edge Cases

- What happens if upstream proto files reference new imports not present in the SDK's proto directory? The buf configuration must resolve all imports, or generation fails at build time (caught by `make build`).
- What happens if upstream changed a field type or number on an existing message? This would be a breaking proto change. The assumption is upstream follows proto compatibility rules, so this should not occur.
- What happens if `buf.gen.yaml` does not discover the new `inference.proto` automatically? The configuration must be updated to include the new proto file or its directory.

### Error Handling

- If `mise run proto:sync` fails because the upstream path is not found, the task exits with a clear error. The developer must set `UPSTREAM_PATH` or ensure the default path (`../OpenShell/proto/`) is accessible.
- If code generation (`mise run proto:gen`) fails due to unresolved imports or syntax errors in upstream protos, the build fails at the generation step. The developer must resolve the upstream dependency before proceeding.
- If generated code fails to compile, `make build` catches this. No partial state is committed.

### Out of Scope

- Inference client implementation (planned for a subsequent PR)
- Converter updates for newly added proto fields (workspace, annotations, deletion_timestamp)
- Public API type changes to expose new fields
- Updates to fake implementations for new fields
- Documentation changes (no new public API surface is introduced)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All 5 upstream proto files MUST be copied verbatim into the SDK proto directory: `openshell.proto`, `inference.proto`, `options.proto`, `datamodel.proto`, `sandbox.proto`
- **FR-002**: The `inference.proto` file MUST be integrated into the build system so that code generation produces Go bindings in a new `inferencev1/` package
- **FR-003**: Code generation MUST produce valid, compilable Go bindings for all proto files
- **FR-004**: All existing unit tests MUST continue to pass after the proto sync
- **FR-005**: The full CI pipeline (lint + build + test) MUST pass with no failures
- **FR-006**: No new client code, converters, or public API types MUST be added in this change; the scope is strictly proto files and generated stubs
- **FR-007**: Generated `.pb.go` files MUST be committed to the repository (following existing project conventions)
- **FR-008**: Internal/operator-only upstream protos (`compute_driver.proto`, `gateway_interceptor.proto`, `supervisor_middleware.proto`) MUST NOT be included in the SDK

### Key Entities

- **Proto File**: A Protocol Buffer definition file (`.proto`) defining service RPCs and message types. Source of truth is the upstream OpenShell repository.
- **Generated Stub**: A Go source file (`.pb.go`, `_grpc.pb.go`) produced by the code generation tool from a proto file. These are committed artifacts, not gitignored.
- **Inference Service**: A new upstream service with 4 RPCs for model inference operations. Entirely new to the SDK.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 5 proto files in the SDK match their upstream counterparts byte-for-byte
- **SC-002**: Generated Go bindings compile with zero errors across all packages
- **SC-003**: The existing test suite passes with 100% of previously-passing tests still passing
- **SC-004**: The `inferencev1/` package exists and contains valid generated Go code
- **SC-005**: The full CI pipeline completes successfully (lint, build, test all green)
- **SC-006**: No new public API surface is introduced beyond the generated proto stubs

## Clarifications

### Session 2026-07-31

- No critical ambiguities detected. Spec scope is well-bounded and all categories assessed as Clear or N/A.

## Assumptions

- Upstream proto files follow Protocol Buffer compatibility rules (no breaking changes to existing field numbers or types)
- The SDK's existing build configuration handles proto discovery within the `proto/` directory, or can be trivially updated to do so
- The `UPSTREAM_PATH` environment variable or default path (`../OpenShell/proto/`) provides access to the upstream proto files
- Authorization annotations added by PR #2445 are harmless metadata that do not affect client-side code generation
- `datamodel.proto` and `sandbox.proto` may have also been updated upstream and should be synced regardless of whether they changed in PR #2445 specifically
