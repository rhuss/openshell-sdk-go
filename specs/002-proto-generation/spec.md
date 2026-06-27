# Feature Specification: Proto Generation Pipeline

**Feature Branch**: `002-proto-generation`
**Created**: 2026-06-27
**Status**: Draft
**Input**: Brainstorm 003-proto-generation.md

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Generate Go Bindings from Proto Files (Priority: P1)

An SDK developer runs `mise run proto:gen` after syncing proto files from
upstream. The command invokes `protoc` with the correct flags and plugins,
producing Go source files in versioned packages under `proto/`. The
generated files compile without errors and are ready to commit.

**Why this priority**: Without generated Go bindings, no SDK code can
communicate with the OpenShell gateway. This is the foundation for all
client functionality.

**Independent Test**: Run `mise run proto:gen` in a clean checkout with
proto files present, then run `go build ./proto/...` to confirm the
generated code compiles.

**Acceptance Scenarios**:

1. **Given** proto files exist in `proto/` (openshell.proto, datamodel.proto,
   sandbox.proto), **When** the developer runs `mise run proto:gen`,
   **Then** Go source files are generated in `proto/openshellv1/`,
   `proto/datamodelv1/`, and `proto/sandboxv1/` with correct package paths.
2. **Given** generated Go files exist, **When** the developer runs
   `go build ./proto/...`, **Then** all packages compile without errors.
3. **Given** `openshell.proto` imports `datamodel.proto` and `sandbox.proto`,
   **When** Go code is generated, **Then** cross-package imports resolve
   correctly using the module path
   `github.com/rhuss/openshell-sdk-go/proto/...`.
4. **Given** `openshell.proto` imports `google/protobuf/struct.proto`,
   **When** Go code is generated, **Then** the well-known type resolves
   via protoc's built-in include path without additional configuration.

---

### User Story 2 - Sync Proto Files from Upstream (Priority: P1)

An SDK developer runs `mise run proto:sync` to copy the latest proto files
from the upstream OpenShell repository. The task copies only the three
needed proto files, records the upstream commit hash, and leaves the
developer ready to regenerate Go bindings.

**Why this priority**: Proto sync is the entry point for incorporating
upstream API changes. Without it, proto files cannot be updated.

**Independent Test**: Run `mise run proto:sync` with the upstream repo
at a known commit, verify the three expected files are copied and
`proto/UPSTREAM_VERSION` contains the correct commit SHA.

**Acceptance Scenarios**:

1. **Given** the upstream OpenShell repo exists at the default path
   (`../OpenShell/proto/`), **When** the developer runs
   `mise run proto:sync`, **Then** `openshell.proto`, `datamodel.proto`,
   and `sandbox.proto` are copied to `proto/`.
2. **Given** the upstream repo contains additional proto files
   (compute_driver.proto, inference.proto, test.proto), **When** sync
   runs, **Then** only the three specified files are copied.
3. **Given** a successful sync, **When** the developer checks
   `proto/UPSTREAM_VERSION`, **Then** it contains the git commit SHA of
   the upstream repo at sync time.
4. **Given** the developer provides a custom upstream path via argument,
   **When** sync runs, **Then** proto files are copied from the specified
   path instead of the default.

---

### User Story 3 - Validate Generated Code in CI (Priority: P2)

A CI pipeline runs `mise run proto:check` to verify that committed
`.pb.go` files match what `protoc` would generate from the committed
`.proto` files. If they diverge, CI fails with a clear message indicating
that regeneration is needed.

**Why this priority**: Prevents stale generated code from being merged.
Important for code integrity but not needed for initial development.

**Independent Test**: Modify a committed `.pb.go` file by hand, run
`mise run proto:check`, verify it fails. Restore the file, run again,
verify it passes.

**Acceptance Scenarios**:

1. **Given** committed `.pb.go` files match the committed `.proto` files,
   **When** CI runs `mise run proto:check`, **Then** the check passes
   with exit code 0.
2. **Given** a developer manually edited a `.pb.go` file, **When** CI
   runs `mise run proto:check`, **Then** the check fails with a message
   indicating which files are out of date.
3. **Given** a developer updated `.proto` files but forgot to regenerate,
   **When** CI runs `mise run proto:check`, **Then** the check fails.

---

### User Story 4 - Clean Generated Files (Priority: P3)

A developer runs `mise run proto:clean` to remove all generated `.pb.go`
files, returning the `proto/` directory to only the source `.proto` files
and metadata.

**Why this priority**: Convenience task for development workflow. Not
critical for functionality.

**Independent Test**: Run `mise run proto:clean`, verify all `.pb.go`
and `_grpc.pb.go` files are removed while `.proto` files remain.

**Acceptance Scenarios**:

1. **Given** generated `.pb.go` files exist in `proto/` subdirectories,
   **When** the developer runs `mise run proto:clean`, **Then** all
   `.pb.go` and `_grpc.pb.go` files are removed.
2. **Given** proto source files exist in `proto/`, **When** clean runs,
   **Then** `.proto` files and `UPSTREAM_VERSION` are preserved.

---

### Edge Cases

- What happens when the upstream repo path does not exist? The sync task
  fails with a clear error message indicating the path is not found.
- What happens when `protoc` or plugins are not installed? The gen task
  fails with a message listing the missing tool and how to install it
  (via `mise install`).
- What happens when a proto file has syntax errors? The gen task fails
  with protoc's error output, which includes file and line information.
- What happens when the upstream repo has no `.git` directory (e.g.,
  downloaded tarball)? The sync task writes "unknown" to
  `proto/UPSTREAM_VERSION` and proceeds.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST copy exactly three proto files (openshell.proto,
  datamodel.proto, sandbox.proto) from the upstream source during sync.
- **FR-002**: System MUST NOT copy proto files that are not needed for the
  gateway client SDK (compute_driver.proto, inference.proto, test.proto).
- **FR-003**: System MUST record the upstream git commit SHA in
  `proto/UPSTREAM_VERSION` after each sync.
- **FR-004**: System MUST generate Go source files using `protoc` with
  `protoc-gen-go` and `protoc-gen-go-grpc` plugins.
- **FR-005**: System MUST use `--go_opt=M` flags to map proto packages to
  Go import paths without modifying the upstream proto files.
- **FR-006**: System MUST generate files into versioned Go packages:
  `proto/openshellv1/`, `proto/datamodelv1/`, `proto/sandboxv1/`.
- **FR-007**: System MUST provide a validation task that compares committed
  generated files against freshly generated output to detect staleness.
- **FR-008**: System MUST provide a clean task that removes all generated
  `.pb.go` files while preserving source `.proto` files and metadata.
- **FR-009**: System MUST pin tool versions (`protoc`, `protoc-gen-go`,
  `protoc-gen-go-grpc`) in `mise.toml` for reproducible builds.
- **FR-010**: System MUST allow the upstream source path to be configured,
  defaulting to `../OpenShell/proto/`.
- **FR-011**: Generated proto packages MUST NOT be part of the SDK's public
  API surface. They are internal implementation details.

### Key Entities

- **Proto Source File**: An upstream `.proto` file defining gRPC services
  and message types. Three files are in scope: openshell.proto (service
  definition with 55 RPCs), datamodel.proto (shared data types),
  sandbox.proto (sandbox-related types).
- **Generated Go Package**: A directory containing `.pb.go` and
  `_grpc.pb.go` files produced by protoc. Three packages: openshellv1,
  datamodelv1, sandboxv1.
- **Upstream Version Marker**: A file (`proto/UPSTREAM_VERSION`) recording
  which upstream commit the proto files were synced from.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Developers can sync upstream protos and generate Go bindings
  in under 30 seconds total (sync + gen).
- **SC-002**: Generated code compiles successfully on the first attempt
  after running the generation task.
- **SC-003**: CI detects 100% of cases where committed generated code
  diverges from the committed proto source files.
- **SC-004**: A new contributor can set up the proto toolchain by running
  `mise install` with no additional manual steps.
- **SC-005**: The proto generation pipeline requires zero modifications
  to upstream proto files.

## Scope

### In Scope

- Mise tasks for sync, gen, check, and clean
- Tool version pinning in mise.toml
- Proto file layout and Go package structure
- CI integration for staleness detection
- Upstream version tracking

### Out of Scope

- Public Go API types (handled by the core SDK spec, brainstorm 004)
- Proto-to-Go type conversion layer (handled by the core SDK spec)
- Additional proto services (inference, compute_driver) beyond the three
  gateway client protos
- Buf CLI integration (decided against in brainstorm)
- Separate Go module for proto packages (single module for simplicity)

## Assumptions

- The upstream OpenShell repo is available locally at `../OpenShell/` during
  development. CI environments may not have it, so `proto:sync` is a
  developer-only task, not a CI step.
- `mise install` handles installation of `protoc`, `protoc-gen-go`, and
  `protoc-gen-go-grpc` to the correct versions.
- The `google/protobuf/struct.proto` well-known type is resolved by protoc's
  built-in include path and requires no additional configuration.
- Proto files are committed to the repository so that `go get` works without
  requiring protoc on the consumer's machine.
- Generated `.pb.go` files are committed to the repository to avoid build-time
  proto compilation requirements for SDK consumers.
