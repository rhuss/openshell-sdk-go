# Feature Specification: Upstream PR Preparation

**Feature Branch**: `019-upstream-pr`
**Created**: 2026-07-11
**Status**: Draft
**Input**: Brainstorm 024 - Prepare draft PR against NVIDIA/OpenShell contributing the Go SDK under sdk/go/
**Upstream Issue**: [NVIDIA/OpenShell#2044](https://github.com/NVIDIA/OpenShell/issues/2044)

## User Scenarios & Testing

### User Story 1 - Module Path Migration (Priority: P1)

A contributor prepares the Go SDK source code for inclusion in the
NVIDIA/OpenShell monorepo by rewriting the Go module path from the
development repo to the upstream path, ensuring all internal imports
compile and tests pass under the new module identity.

**Why this priority**: Without correct module paths, no consumer can
import the SDK and no CI can compile it. This is the foundation for
every other story.

**Independent Test**: After rewriting, `go build ./...` and `go test ./...`
succeed under the new module path with zero import errors.

**Acceptance Scenarios**:

1. **Given** the SDK source at `github.com/rhuss/openshell-sdk-go`,
   **When** the module path is rewritten to `github.com/NVIDIA/OpenShell/sdk/go`,
   **Then** `go.mod` declares the new module, all internal import statements
   reference the new path, and `go build ./...` succeeds with no errors.

2. **Given** the rewritten module,
   **When** `go test ./...` is run,
   **Then** all existing unit tests pass with the same results as before
   the rewrite.

3. **Given** the rewritten module,
   **When** a consumer creates a new Go project and runs
   `go get github.com/NVIDIA/OpenShell/sdk/go/openshell/v1`,
   **Then** the SDK packages resolve and compile correctly.

---

### User Story 2 - Example Extraction (Priority: P1)

A contributor extracts the oshell TUI example from the SDK repo into a
separate repository so the upstream PR contains only SDK library code,
not application-level examples.

**Why this priority**: The 4,649-LOC TUI example adds significant bulk
to the PR and is not part of the SDK library. Extracting it keeps the
upstream contribution focused.

**Independent Test**: The examples repo builds and runs independently,
importing the SDK via the new module path.

**Acceptance Scenarios**:

1. **Given** the `examples/oshell/` directory in the SDK repo,
   **When** it is moved to `github.com/rhuss/openshell-examples`,
   **Then** the examples repo has its own `go.mod` with a dependency on
   `github.com/NVIDIA/OpenShell/sdk/go` and compiles successfully.

2. **Given** the examples have been extracted,
   **When** the SDK repo is examined,
   **Then** the `examples/` directory no longer exists.

---

### User Story 3 - Fern Documentation Integration (Priority: P2)

A contributor writes Go SDK documentation as Fern MDX pages that
integrate into the existing OpenShell docs site, appearing alongside
the Python SDK documentation.

**Why this priority**: Documentation demonstrates that the SDK is
production-ready and helps reviewers understand the API surface.
Ranked P2 because the SDK is functional without docs.

**Independent Test**: The Fern docs build succeeds with the new Go SDK
pages and they appear in the navigation.

**Acceptance Scenarios**:

1. **Given** the OpenShell Fern docs structure,
   **When** Go SDK MDX pages are added under `docs/sdks/go/`,
   **Then** the Fern build (`fern check`) succeeds and the pages appear
   in the rendered navigation under an "SDKs > Go" section.

2. **Given** the Go SDK docs pages,
   **When** a reader navigates to the docs site,
   **Then** they find getting-started, architecture, error-handling, and
   authentication guides with code examples that match the actual SDK API.

3. **Given** the existing Python SDK documentation,
   **When** Go SDK docs are added,
   **Then** the Python SDK docs remain unchanged and accessible.

---

### User Story 4 - Proto Generation Automation (Priority: P2)

A contributor adds a mise task and CI validation step that regenerates
Go protobuf bindings from the repo's `proto/` directory and fails the
build if the committed `.pb.go` files are stale.

**Why this priority**: Ensures the Go SDK stays in sync with proto
changes. Ranked P2 because manual regeneration works as a fallback.

**Independent Test**: Modify a proto file, run CI, and observe that it
detects the stale generated files and fails.

**Acceptance Scenarios**:

1. **Given** the proto files in `proto/` and the mise task `go:proto`,
   **When** a contributor runs `mise run go:proto`,
   **Then** Go protobuf bindings are regenerated in `sdk/go/proto/` from
   the current proto definitions.

2. **Given** a PR that modifies a proto file without regenerating Go bindings,
   **When** CI runs the proto freshness check,
   **Then** the check fails with a clear message indicating which files
   are stale and how to regenerate them.

3. **Given** a PR where proto files and generated Go bindings are in sync,
   **When** CI runs the proto freshness check,
   **Then** the check passes.

---

### User Story 5 - Spec Artifact Scoping (Priority: P3)

A contributor includes the `specs/` directory in the upstream PR as
design documentation while excluding internal development artifacts
(brainstorms, spec-kit config, Claude Code config).

**Why this priority**: The specs provide design rationale but are not
required for the SDK to function. The PR description asks upstream
maintainers whether to keep or remove them.

**Independent Test**: The PR diff includes `sdk/go/specs/` but does not
include `brainstorm/`, `.specify/`, `.claude/`, `CLAUDE.md`, or `AGENTS.md`.

**Acceptance Scenarios**:

1. **Given** the SDK repo with specs and brainstorms,
   **When** the PR is prepared,
   **Then** `sdk/go/specs/` is included in the PR and `brainstorm/`,
   `.specify/`, `.claude/`, `CLAUDE.md`, `AGENTS.md` are excluded.

2. **Given** the draft PR description,
   **When** a reviewer reads it,
   **Then** they see a note explaining that the specs were created using
   spec-driven development and asking whether they should be retained
   or removed from the repository.

---

### User Story 6 - Draft PR Creation (Priority: P3)

A contributor opens a draft PR against NVIDIA/OpenShell that presents
the complete Go SDK, referencing the existing upstream issue.

**Why this priority**: The PR is the delivery vehicle but depends on
all prior stories being complete.

**Independent Test**: The draft PR exists on GitHub, references issue
#2044, and contains the expected file tree under `sdk/go/`.

**Acceptance Scenarios**:

1. **Given** all SDK code, docs, and CI changes are ready in the fork,
   **When** a draft PR is opened against `NVIDIA/OpenShell`,
   **Then** the PR title references the Go SDK, the description references
   issue [#2044](https://github.com/NVIDIA/OpenShell/issues/2044), and
   the PR is marked as draft.

2. **Given** the draft PR,
   **When** a reviewer examines the file tree,
   **Then** they see `sdk/go/` with Go source, `sdk/go/proto/` with
   committed `.pb.go` files, `sdk/go/specs/` with design documentation,
   `docs/sdks/go/` with Fern MDX pages, and `tasks/go.toml` with the
   mise task.

---

### Edge Cases

- What happens when a proto file adds a new service or message type?
  The CI freshness check detects the drift and fails, prompting
  regeneration.
- What if the upstream repo already has an `sdk/` directory by the time
  the PR is ready? The PR adapts to the existing directory structure.
- What if the Fern docs build uses a version of Fern that is
  incompatible with the added MDX pages? The contributor runs
  `fern check` locally before submitting.
- What if `option go_package` is missing from upstream proto files?
  The proto generation task supplies the package option via protoc
  flags rather than modifying the proto source files.

## Clarifications

### Session 2026-07-11

- Q: Should creating the actual GitHub repo for examples be in scope, or only prepare the extraction? → A: Create the GitHub repo now and push the extracted examples.
- Q: How deep should the Fern MDX documentation pages be? → A: Concise reference pages (1-2 pages each) with key concepts and short code snippets.
- Q: Should the draft PR use a single squashed commit or separate commits per story? → A: Single squashed commit with a comprehensive commit message.

## Requirements

### Functional Requirements

- **FR-001**: The Go module path MUST be rewritten from
  `github.com/rhuss/openshell-sdk-go` to
  `github.com/NVIDIA/OpenShell/sdk/go` in `go.mod` and all `.go` files.
- **FR-002**: All unit tests MUST pass after the module path rewrite
  with no behavior changes.
- **FR-003**: The `examples/oshell/` directory MUST be moved to
  `github.com/rhuss/openshell-examples` as a separate repository with
  its own `go.mod`. The GitHub repo MUST be created and the extracted
  code pushed as part of this feature (not deferred).
- **FR-004**: The examples repository MUST depend on the SDK via the new
  upstream module path.
- **FR-005**: Fern MDX documentation pages MUST be created under
  `docs/sdks/go/` covering getting-started, architecture, error-handling,
  and authentication topics. Each page SHOULD be concise (1-2 pages) with
  key concepts and short code snippets, not full tutorial walkthroughs.
- **FR-006**: The MDX pages MUST be wired into `docs/index.yml` so they
  appear in the Fern docs site navigation.
- **FR-007**: A `go:proto` mise task MUST be created in `tasks/go.toml`
  that regenerates Go protobuf bindings from the repo's `proto/` directory.
- **FR-008**: A CI step MUST be added to `branch-checks.yml` that runs
  the proto generation task and fails if committed `.pb.go` files diverge
  from what the task produces.
- **FR-009**: Generated `.pb.go` files MUST be committed to the repository
  so consumers can `go get` the SDK without needing protoc installed.
- **FR-010**: The `specs/` directory MUST be included in the PR under
  `sdk/go/specs/`.
- **FR-011**: The following MUST be excluded from the PR: `brainstorm/`,
  `.specify/`, `.claude/`, `CLAUDE.md`, `AGENTS.md`, `docs/` (mdbook).
- **FR-012**: The PR MUST be opened as a draft against the
  `NVIDIA/OpenShell` main branch with a single squashed commit
  containing all changes.
- **FR-013**: The PR description MUST reference issue
  [#2044](https://github.com/NVIDIA/OpenShell/issues/2044).
- **FR-014**: The PR description MUST mention the spec-driven development
  methodology and ask upstream maintainers whether to retain or remove
  the `specs/` directory.
- **FR-015**: The proto generation MUST work without modifying upstream
  proto files (supply `go_package` via protoc command-line flags).
- **FR-016**: Code examples in Fern MDX documentation pages MUST use
  actual SDK function signatures, type names, and argument counts so
  they compile against the current SDK API.

### Key Entities

- **SDK source tree**: The Go packages under `sdk/go/openshell/` and
  `sdk/go/proto/` that comprise the library.
- **Examples repository**: A standalone GitHub repo at
  `github.com/rhuss/openshell-examples` holding the oshell TUI and
  future examples.
- **Fern MDX pages**: Documentation files under `docs/sdks/go/` that
  integrate into the OpenShell docs site.
- **Mise task**: The `go:proto` task definition in `tasks/go.toml` that
  drives proto-to-Go code generation.
- **Draft PR**: The GitHub pull request against NVIDIA/OpenShell
  delivering all changes.

## Success Criteria

### Measurable Outcomes

- **SC-001**: `go build ./...` and `go test ./...` succeed under the new
  module path `github.com/NVIDIA/OpenShell/sdk/go` with zero failures.
- **SC-002**: The examples repository at `github.com/rhuss/openshell-examples`
  compiles and runs the oshell TUI against a gateway using the upstream
  SDK module path.
- **SC-003**: `fern check` passes with the Go SDK documentation pages
  included and they appear in the rendered navigation.
- **SC-004**: The proto freshness CI step detects intentionally stale
  `.pb.go` files and fails within the normal CI pipeline.
- **SC-005**: The draft PR file tree contains exactly the expected
  directories (`sdk/go/`, `docs/sdks/go/`, `tasks/go.toml` changes) and
  excludes all internal development artifacts.
- **SC-006**: A reviewer can read the PR description, understand the SDK
  scope, find the referenced upstream issue, and see the question about
  spec retention without additional context.

## Assumptions

- The `rhuss/OpenShell` fork is up to date with `NVIDIA/OpenShell` main
  branch at the time the PR branch is created.
- The upstream repo does not already have an `sdk/` directory. If it
  does, the directory structure adapts accordingly.
- The Fern docs framework version used by upstream supports the MDX
  features needed for Go SDK pages (code blocks, tabs, callouts).
- `protoc` and the Go gRPC plugin are available in the CI environment
  or can be installed via mise.
- The existing SPDX license headers (Apache-2.0) are compatible with
  the upstream repo's licensing requirements.
- The module path rewrite is a mechanical find-and-replace operation
  that does not change any SDK behavior or API surface.
