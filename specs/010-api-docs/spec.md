# Feature Specification: API Documentation Site

**Feature Branch**: `010-api-docs`
**Created**: 2026-06-29
**Status**: Draft
**Input**: API documentation site using mdBook with custom typography, gRPC mapping, usage examples, and pkg.go.dev Example* test functions.

## User Scenarios & Testing

### User Story 1 - Browse SDK Documentation Online (Priority: P1)

A Go developer discovers the OpenShell SDK and wants to understand its
capabilities, see usage examples, and learn how to get started. They visit
the documentation site (hosted on GitHub Pages) and find a searchable,
well-organized reference covering all SDK interfaces (11 top-level plus
2 provider sub-interfaces).

**Why this priority**: First impressions determine adoption. A polished
docs site is the primary discovery and onboarding surface.

**Independent Test**: Visit the deployed GitHub Pages URL, verify the site
loads with navigation sidebar, search works, and all content pages render
with correct typography and code highlighting.

**Acceptance Scenarios**:

1. **Given** a deployed docs site, **When** a developer opens the URL,
   **Then** they see a sidebar with all documentation sections and can
   navigate between pages.
2. **Given** the docs site, **When** a developer uses the search feature,
   **Then** results include matches from page content and code examples.
3. **Given** the docs site, **When** viewed on desktop, **Then** text uses
   a modern sans-serif font at 18px+ with generous line-height for
   readability.

---

### User Story 2 - Understand gRPC Mapping (Priority: P1)

A developer familiar with the OpenShell gRPC API wants to understand how
SDK methods map to underlying RPCs. For the three hero interfaces
(Sandboxes, Exec, Providers), they see side-by-side code showing the SDK
call alongside the corresponding gRPC RPC. For the remaining interfaces,
they find reference tables mapping each SDK method to its RPC name and
proto file.

**Why this priority**: The gRPC-to-SDK mapping is the primary
differentiator of this documentation. Developers migrating from raw gRPC
need to see exactly what the SDK abstracts.

**Independent Test**: Navigate to the Sandboxes API reference page, verify
side-by-side SDK and gRPC code blocks are present for each method. Navigate
to the Config API reference page, verify a reference table maps each method
to its RPC.

**Acceptance Scenarios**:

1. **Given** the Sandboxes reference page, **When** a developer reads the
   Create method section, **Then** they see a labeled SDK code block
   followed by a labeled gRPC proto block showing the corresponding RPC.
2. **Given** the Config reference page, **When** a developer reads the
   method table, **Then** each row shows the SDK method name, the gRPC RPC
   name, and the proto file path.
3. **Given** any API reference page, **When** a developer reads the code
   examples, **Then** all Go code compiles against the current SDK version.

---

### User Story 3 - Quick Start with the SDK (Priority: P1)

A new developer wants to go from zero to running their first sandbox in
under 5 minutes. The getting started guide walks them through installation,
connecting to a gateway, creating a sandbox, running a command, and
cleaning up.

**Why this priority**: Getting started guides are the highest-traffic pages
in any SDK documentation. A clear onboarding path directly drives adoption.

**Independent Test**: Follow the getting started guide from scratch, verify
each code snippet works and the guide is completable in under 5 minutes.

**Acceptance Scenarios**:

1. **Given** a developer with Go 1.23+ installed, **When** they follow the
   getting started guide, **Then** they can connect to a gateway and create
   a sandbox using the provided code snippets.
2. **Given** the getting started guide, **When** a developer copies a code
   snippet, **Then** it compiles without modification (assuming valid
   credentials).

---

### User Story 4 - Learn Error Handling Patterns (Priority: P2)

A developer building production code needs to understand how to handle
errors from the SDK. The error handling guide covers StatusError, typed
error checks (IsNotFound, IsAlreadyExists, IsConflict, IsUnavailable),
and practical retry patterns.

**Why this priority**: Error handling is critical for production code but
is often poorly documented. A dedicated guide reduces support burden.

**Independent Test**: Navigate to the error handling guide, verify it
covers all typed error checks with compilable examples.

**Acceptance Scenarios**:

1. **Given** the error handling guide, **When** a developer reads it,
   **Then** they find examples for each typed error check function
   (IsNotFound, IsAlreadyExists, IsConflict, IsUnavailable, IsUnimplemented).
2. **Given** the error handling guide, **When** a developer looks for retry
   patterns, **Then** they find a practical example showing retry with
   backoff for transient errors.

---

### User Story 5 - Test with the Fake Client (Priority: P2)

A developer building a Kubernetes operator needs to write tests without a
real OpenShell gateway. The testing guide shows how to use the fake client
package, seed test fixtures, and verify watch events.

**Why this priority**: The fake client is a key differentiator of this SDK.
A testing guide helps operator developers adopt it quickly.

**Independent Test**: Navigate to the testing guide, verify it shows fake
client creation, fixture seeding, and watch event testing with compilable
examples.

**Acceptance Scenarios**:

1. **Given** the testing guide, **When** a developer reads it, **Then**
   they find a complete example of creating a fake client, seeding fixtures,
   and running assertions.
2. **Given** the testing guide, **When** a developer reads the watch events
   section, **Then** they find an example showing how mutations emit events
   to active watchers.

---

### User Story 6 - Discover SDK via pkg.go.dev (Priority: P2)

A Go developer finds the SDK on pkg.go.dev and sees runnable Example*
functions demonstrating key operations. These examples appear in the
standard Go documentation with Share/Format/Run buttons.

**Why this priority**: pkg.go.dev is the canonical discovery path for Go
packages. Example functions are compile-checked and appear automatically.

**Independent Test**: Run `go test` on the example files and verify all
Example functions compile and produce expected output.

**Acceptance Scenarios**:

1. **Given** the SDK package, **When** a developer views it on pkg.go.dev,
   **Then** they see Example functions for at least: NewClient, Sandbox
   Create, Exec Run, Watch, Provider operations, error handling, and fake
   client.
2. **Given** any Example function, **When** `go test` runs it, **Then** it
   compiles and produces the expected output.

---

### User Story 7 - View Architecture Overview (Priority: P3)

A developer wants to understand how the SDK is structured internally
before diving into specific interfaces. The architecture overview shows
the Client, sub-client hierarchy, gRPC layer, and gateway relationship.

**Why this priority**: Understanding the architecture helps developers
navigate the API surface. Lower priority because most developers go
straight to specific interfaces.

**Independent Test**: Navigate to the architecture page, verify it shows
the client hierarchy diagram and explains the sub-client pattern.

**Acceptance Scenarios**:

1. **Given** the architecture page, **When** a developer reads it,
   **Then** they see a diagram showing Client and its sub-client accessors.
2. **Given** the architecture page, **When** a developer reads it,
   **Then** it explains the proto isolation principle (SDK types vs wire
   types).

---

### Edge Cases

- What happens when a new SDK interface is added? The documentation
  structure must accommodate new interfaces without restructuring.
- How are code examples kept in sync with API changes? The Example* test
  functions are compile-checked, but mdBook code snippets need a manual
  or CI-based sync process.
- What if the site is accessed on mobile? mdBook's default responsive
  layout handles narrow viewports, but custom CSS must not break mobile
  rendering.

## Requirements

### Functional Requirements

- **FR-001**: The documentation site MUST be built with mdBook and
  deployed to GitHub Pages via GitHub Actions.
- **FR-002**: The site MUST include a getting started guide covering
  installation, connection, sandbox creation, command execution, and
  cleanup.
- **FR-003**: The site MUST include an architecture overview page with a
  diagram showing the Client and sub-client hierarchy.
- **FR-004**: The site MUST include API reference pages for all SDK
  interfaces: the 11 top-level interfaces accessible from ClientInterface
  (SandboxInterface, ExecInterface, FileInterface, HealthInterface,
  ProviderInterface, ServiceInterface, SSHInterface, TCPInterface,
  ConfigInterface, PolicyInterface, and ClientInterface itself), plus the
  2 sub-interfaces accessible from ProviderInterface (ProfileInterface,
  RefreshInterface). ProfileInterface and RefreshInterface pages MUST
  clearly indicate they are accessed via `client.Providers().Profiles()`
  and `client.Providers().Refresh()` respectively.
- **FR-005**: The API reference for hero interfaces (SandboxInterface,
  ExecInterface, ProviderInterface) MUST include side-by-side code blocks
  showing the SDK call and the corresponding gRPC RPC definition.
- **FR-006**: The API reference for non-hero interfaces MUST include a
  reference table mapping each SDK method to its gRPC RPC name and proto
  file.
- **FR-007**: The site MUST include an error handling guide covering
  StatusError, all typed error check functions, and retry patterns.
- **FR-008**: The site MUST include a testing guide showing fake client
  usage for sandbox lifecycle, watch events, and health simulation.
- **FR-009**: The site MUST use custom CSS to apply a modern sans-serif
  font (Inter or system-ui fallback), base font size of 18px or larger,
  and line-height of 1.6 or greater.
- **FR-010**: The site MUST have built-in full-text search across all
  pages and code examples.
- **FR-011**: The GitHub Actions workflow MUST build and deploy the site
  on every push to the main branch without requiring any local build
  steps.
- **FR-012**: The SDK MUST include Example* test functions in the
  openshell/v1 package for pkg.go.dev rendering, covering at minimum:
  NewClient, Sandbox Create, Exec Run, Watch, Provider operations, error
  handling, and fake client setup.
- **FR-013**: All Example* test functions MUST compile and pass `go test`.
- **FR-014**: The site MUST include a SUMMARY.md file defining the
  navigation sidebar structure.

### Key Entities

- **mdBook project**: Configuration (book.toml), source directory (docs/src/),
  custom theme CSS, and SUMMARY.md navigation file.
- **API Reference Page**: One per SDK interface, containing method
  documentation, code examples, and gRPC mapping (side-by-side or table).
- **Example Test Function**: Go test file(s) containing Example* functions
  that render on pkg.go.dev.
- **GitHub Actions Workflow**: CI pipeline that builds mdBook and deploys
  to GitHub Pages.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A new developer can complete the getting started guide and
  have a working sandbox in under 5 minutes (excluding credential setup).
- **SC-002**: All 13 SDK interfaces (11 top-level + 2 provider
  sub-interfaces) have dedicated API reference pages with complete method
  documentation.
- **SC-003**: The three hero interfaces (Sandboxes, Exec, Providers) each
  have side-by-side SDK/gRPC code blocks for every public method.
- **SC-004**: All code examples in the documentation compile against the
  current SDK version.
- **SC-005**: The site search returns relevant results for any SDK method
  name or concept within 1 second.
- **SC-006**: At least 10 Example* test functions exist and pass
  `go test` in the openshell/v1 package.
- **SC-007**: The GitHub Actions deployment workflow completes in under
  5 minutes.

## Assumptions

- The repository uses GitHub Pages for hosting (not a custom domain or
  external hosting service).
- Contributors do not need mdBook installed locally; they edit markdown
  files and CI handles the build.
- The mdBook binary is available via GitHub Actions (the
  peaceiris/actions-mdbook action or similar).
- Code examples reference the current SDK API signatures; when signatures
  change, examples are updated in the same commit.
- The docs/src/ directory is used for mdBook source files. The existing
  docs/go-sdk-proposal.md will coexist alongside the mdBook source.
- The Example* test functions use the fake client for compilation
  without requiring a live gateway connection.
- Inter or a similar modern sans-serif font is loaded via CSS from a CDN
  or bundled in the theme directory.
