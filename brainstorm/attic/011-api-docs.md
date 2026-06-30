# Brainstorm: API Documentation Site

**Date:** 2026-06-29
**Status:** active

## Problem Framing

The SDK has 14 public interfaces, 20+ source files, and a comprehensive
README with usage examples, but no dedicated documentation site. Developers
discover the SDK through two paths: pkg.go.dev (Go ecosystem standard) and
direct browsing (linked from README or OpenShell project). Neither path
currently offers searchable, structured API reference with gRPC mapping or
guided content beyond the doc.go examples.

The constitution mandates compiled doc examples (Principle VIII) and
agent-friendly documentation (Principle IX), but no `Example*` test
functions exist yet, and pkg.go.dev renders only the raw doc.go.

Goals:
- Beautiful, aesthetic documentation with modern typography
- Built-in search
- Real-world usage examples per interface
- Show the relationship between SDK methods and underlying gRPC RPCs
- No local build steps required (CI-only build is acceptable)
- Hosted from this repository via GitHub Pages

## Approaches Considered

### A: MkDocs Material (Python-based)

The most polished documentation framework available. Native content tabs
for side-by-side code comparison, excellent search (searches inside code
blocks), admonitions, dark/light mode, social cards. Used by FastAPI,
Pydantic.

- Pros: Best aesthetics, native tab widget for gRPC side-by-side, rich
  component library, massive community
- Cons: Requires Python in CI, heavier dependency chain, not Go-ecosystem
  native

### B: mdBook (Rust-based, single binary)

Clean, book-style documentation with left sidebar navigation, built-in
search, 6 color themes, keyboard navigation. Used by the Rust Book,
Comprehensive Rust (Google). Single binary with no dependencies.

- Pros: Simplest CI setup (one binary), minimal maintenance, clean
  aesthetic, fast builds, great code highlighting
- Cons: No native content tabs (side-by-side uses sequential labeled code
  blocks instead), fewer built-in components than MkDocs Material

### C: Hugo + Doks Theme (Go-native)

Go-native static site generator, extremely fast builds. The Doks theme
provides a documentation-focused layout. Used by Kubernetes, Istio.

- Pros: Go-native toolchain, very fast, powerful templating, large theme
  ecosystem
- Cons: Most complex configuration, theme quality varies, higher
  maintenance burden, Doks theme less polished than MkDocs Material

## Decision

**Chosen: Approach B (mdBook)** for the docs site, combined with
`Example*` test functions for pkg.go.dev.

Rationale:
- Simplest setup and lowest maintenance (single binary, markdown only)
- Clean aesthetic that fits a developer SDK
- Built-in search covers the discoverability requirement
- The lack of native tabs is acceptable: sequential labeled code blocks
  (SDK code, then gRPC proto) are clear and readable
- CI setup is trivial (one GitHub Action step)
- Custom CSS overrides handle typography requirements (modern sans-serif,
  larger font size, generous line-height)

The `Example*` test functions for pkg.go.dev are a separate, complementary
effort that improves the Go-ecosystem discovery path with zero tooling.

## Key Requirements

### Docs Site (mdBook)

**Content structure:**
- Getting started guide (install, connect, create first sandbox)
- Architecture overview (Client, sub-clients, gRPC, gateway diagram)
- API reference per interface (14 interfaces total):
  - Full side-by-side gRPC comparison for hero interfaces: Sandboxes, Exec,
    Providers
  - Reference tables (SDK method, RPC name, proto file) for remaining ~10
    interfaces
- Error handling guide (StatusError, typed checks like IsNotFound/IsConflict,
  retry patterns)
- Testing guide (fake client usage for operator/controller test suites)

**Typography and aesthetics:**
- Modern sans-serif font (Inter or system-ui stack)
- Larger base font size (18px+) for readability
- Generous line-height (1.6+)
- Custom CSS via mdBook's theme/ directory
- Clean, not cluttered

**Infrastructure:**
- Source markdown in `docs/` directory
- `book.toml` configuration at project root or `docs/`
- GitHub Actions workflow to build and deploy to GitHub Pages
- No local build steps required for contributors

### pkg.go.dev Enhancement

- Add `Example*` test functions for major operations across all interfaces
- Examples must compile (constitution Principle VIII)
- Cover: connect, create sandbox, exec run, exec stream, watch, providers,
  services, profiles, refresh, error handling, fake client setup

## Open Questions

- Should the docs site have a custom landing/index page with the
  architecture diagram, or jump straight into the getting started guide?
- Which mdBook theme should be the default? (Ayu and Navy are popular dark
  options; Light is the classic default)
- Should code examples in the mdBook site be extracted from the actual
  `Example*` test functions (single source of truth) or maintained
  separately?
- Is `mdbook-admonish` (callout boxes for tips, warnings, notes) worth
  adding as a preprocessor, or keep it pure markdown?
