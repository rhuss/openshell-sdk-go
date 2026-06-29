# Implementation Plan: API Documentation Site

**Branch**: `010-api-docs` | **Date**: 2026-06-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/010-api-docs/spec.md`

## Summary

Create an mdBook-based documentation site with custom typography, full-text
search, and gRPC mapping for all 13 SDK interfaces. Deploy via GitHub
Actions to GitHub Pages. Add Example* test functions for pkg.go.dev
rendering. The site uses markdown source in `docs/src/`, custom CSS for
modern sans-serif fonts at 18px+, and a SUMMARY.md navigation sidebar.

## Technical Context

**Language/Version**: Go 1.23+ (Example* tests), Markdown (docs content)
**Primary Dependencies**: mdBook (Rust binary, CI-only), mdbook-admonish (preprocessor for callout boxes)
**Storage**: N/A (static site, no database)
**Testing**: `go test` for Example* functions, `mdbook build` for site validation
**Target Platform**: GitHub Pages (static hosting), pkg.go.dev (Go package registry)
**Project Type**: Documentation (static site + Go test examples)
**Performance Goals**: Site build under 5 minutes in CI, search results under 1 second
**Constraints**: No local build tools required for contributors, markdown-only authoring
**Scale/Scope**: 13 API reference pages, 4 guide pages, 10+ Example* test functions

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Documentation shows SDK types only; proto shown separately in gRPC mapping sections |
| II. Idiomatic Go | PASS | Example* functions follow Go testing conventions |
| III. Test-First | PASS | Example* test functions ARE tests; they compile-check all code examples |
| IV. Upstream Tracking | N/A | Documentation feature, not proto-dependent |
| V. Minimal Dependencies | PASS | mdBook is CI-only, not a runtime dependency |
| VI. Secrets Never Leak | PASS | Code examples use placeholder credentials ("my-token") |
| VII. Deep Copy at Boundaries | N/A | No proto/SDK boundary crossing in docs |
| VIII. Doc Examples Compile | PASS | Core requirement: Example* functions enforce this |
| IX. Agent-Friendly Documentation | PASS | This feature directly implements Principle IX |
| X. Proto-SDK Naming Fidelity | PASS | gRPC mapping tables verify naming correspondence |
| XI. Fake-Real Parity | PASS | Testing guide demonstrates fake client with real validation |
| XII. Graceful Shutdown Order | N/A | Documentation feature |

## Project Structure

### Documentation (this feature)

```text
specs/010-api-docs/
├── plan.md              # This file
├── spec.md              # Feature specification
├── research.md          # Phase 0: research findings
├── data-model.md        # Phase 1: content model
└── tasks.md             # Phase 2: task breakdown
```

### Source Code (repository root)

```text
docs/
├── book.toml                # mdBook configuration
├── src/
│   ├── SUMMARY.md           # Navigation sidebar structure
│   ├── introduction.md      # Landing page (links to getting started)
│   ├── getting-started.md   # Quick start guide
│   ├── architecture.md      # Client hierarchy + proto isolation
│   ├── error-handling.md    # StatusError, typed checks, retry patterns
│   ├── testing.md           # Fake client guide
│   └── api/
│       ├── overview.md      # API surface overview
│       ├── sandboxes.md     # Hero: side-by-side gRPC mapping
│       ├── exec.md          # Hero: side-by-side gRPC mapping
│       ├── providers.md     # Hero: side-by-side gRPC mapping
│       ├── profiles.md      # Reference table
│       ├── refresh.md       # Reference table
│       ├── services.md      # Reference table
│       ├── files.md         # Reference table
│       ├── health.md        # Reference table
│       ├── ssh.md           # Reference table
│       ├── tcp.md           # Reference table
│       ├── config.md        # Reference table
│       └── policy.md        # Reference table
├── theme/
│   └── custom.css           # Typography overrides (Inter, 18px, 1.6 line-height)
└── go-sdk-proposal.md       # Existing file (preserved)

openshell/v1/
├── example_test.go          # Example* functions for pkg.go.dev
└── example_fake_test.go     # Example* functions for fake client

.github/workflows/
├── ci.yml                   # Existing CI workflow
└── docs.yml                 # New: mdBook build + GitHub Pages deploy
```

**Structure Decision**: mdBook source lives in `docs/src/` with configuration
at `docs/book.toml`. API reference pages are grouped under `docs/src/api/`.
Custom CSS is in `docs/theme/`. Example test files are in the existing
`openshell/v1/` package alongside the source they document. The existing
`docs/go-sdk-proposal.md` is preserved at its current location.

## Implementation Approach

### Phase 1: mdBook scaffolding and CI

Set up the mdBook project structure, book.toml configuration, SUMMARY.md
navigation, custom CSS theme, and GitHub Actions workflow. Validate that
the site builds and deploys successfully with placeholder content.

### Phase 2: Guide content

Write the getting started guide, architecture overview, error handling
guide, and testing guide. These are narrative pages with code examples
drawn from the existing README and doc.go.

### Phase 3: API reference (hero interfaces)

Create the three hero interface pages (Sandboxes, Exec, Providers) with
full side-by-side SDK/gRPC code blocks for every method. Extract method
signatures from the Go source and RPC definitions from the proto files.

### Phase 4: API reference (remaining interfaces)

Create the 10 remaining interface pages with reference tables mapping
each SDK method to its gRPC RPC name and proto file path.

### Phase 5: Example* test functions

Write Example* test functions in `openshell/v1/example_test.go` and
`openshell/v1/example_fake_test.go`. These must compile against the
current SDK using the fake client (no live gateway needed).

### gRPC Mapping Strategy

For hero interfaces (Sandboxes, Exec, Providers), each method section
shows two labeled code blocks:

```markdown
**SDK (Go):**
` ` `go
sandbox, err := client.Sandboxes().Create(ctx, "my-sb", spec, nil)
` ` `

**gRPC (Proto):**
` ` `protobuf
rpc CreateSandbox(CreateSandboxRequest) returns (SandboxResponse);
` ` `
```

For non-hero interfaces, a reference table per interface:

```markdown
| SDK Method | gRPC RPC | Proto File |
|------------|----------|------------|
| `GetSandbox(ctx, name)` | `GetSandboxConfig` | `openshell.proto` |
```

### Typography Implementation

Custom CSS in `docs/theme/custom.css`:

- Font: Inter (loaded from Google Fonts CDN) with system-ui fallback
- Base font size: 18px
- Line-height: 1.7
- Code blocks: 15px with comfortable padding
- Headings: scaled proportionally from 18px base
- Max content width: 800px for optimal reading line length

### CI/CD Pipeline

New `.github/workflows/docs.yml`:

- Trigger: push to main branch, changes in `docs/` or `openshell/v1/example_*`
- Steps: install mdBook + mdbook-admonish, build, deploy to GitHub Pages
- Uses `peaceiris/actions-gh-pages` or equivalent for deployment
- Build time target: under 5 minutes
