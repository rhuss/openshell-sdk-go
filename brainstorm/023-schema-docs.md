# Brainstorm: Schema and Data Type Documentation

**Date:** 2026-07-03
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/29

## Problem Framing

The SDK documentation has 16 API reference pages organized by gRPC
sub-client (Sandboxes, Exec, Providers, etc.) but no dedicated
documentation for the data types themselves. Types like `SandboxSpec`,
`PolicyRule`, `ProviderProfile`, and `NetworkPolicyEndpoint` appear
inline in endpoint pages without field descriptions, relationships,
or usage context. The `openshell/v1/types/` package contains ~96 types
across 20 files, but only ~30 are referenced in docs.

For developers working with the SDK, understanding the data model is
more important than knowing the CRUD methods. A developer creating a
sandbox needs to understand `SandboxSpec`, `SandboxTemplate`,
`SandboxStatus`, and their relationships. Currently they have to read
Go source code to learn this.

## Approaches Considered

### A: Domain-Concept Schema Pages (chosen)

New `docs/src/schema/` directory with pages grouped by domain concept
(sandbox, policy, provider, etc.). Each page contains field tables,
type relationships, and usage examples. All 16 existing API pages get
inline cross-links from type names to the schema pages.

- Pros: types are documented where developers think about them (by
  domain, not by Go file), cross-links make the docs navigable,
  examples show types in context
- Cons: more pages to maintain, requires updating all 16 API pages
  for cross-links, manual sync with Go source when types change

### B: Single Types Reference Page

One large `docs/src/schema/types.md` page listing all types with
anchor links, grouped by domain.

- Pros: single page to search, easy to maintain
- Cons: very long page (~96 types), poor navigation, doesn't scale

### C: Auto-Generated from Go Source

Generate type docs from Go doc comments using a tool like `gomarkdoc`
or a custom generator.

- Pros: always in sync with source, zero maintenance
- Cons: output format is hard to customize, loses the narrative
  structure (relationships, examples, cross-domain context), requires
  build tooling

## Decision

**Option A: Domain-concept schema pages with inline cross-links.**
Manual pages give full control over presentation, relationships, and
examples. Auto-generation (Option C) can be explored later as a
complement but not a replacement, since the value is in the curated
relationships and examples, not just field listings.

## Key Requirements

- New `docs/src/schema/` directory
- Pages organized by domain concept, not Go package structure
- Proposed pages: sandbox, policy, network-policy, provider, profile,
  exec, service, config-settings, auth, health, watch-events
- Each page includes: field tables with types and descriptions, type
  relationships (containment, references), usage examples
- Inline cross-links from all 16 existing API pages (first mention of
  each type becomes a link to the schema page anchor)
- New "Schema" section in SUMMARY.md between "API Reference" and "Guides"
- Types derived from Go source in `openshell/v1/types/`
- Only SDK types documented (proto types excluded per Constitution I)

## Open Questions

- Should the schema section appear before or after the API Reference
  section in SUMMARY.md? Before gives types prominence, after follows
  the natural reading flow (learn the API, then dive into types).
- Should enum-like constants (e.g., `SandboxPhase` values) be
  documented on the schema pages or kept on the endpoint pages where
  they're used?
- Should there be a schema overview page with a visual domain model
  diagram showing how the major types relate?
