# Implementation Plan: Edge Auth

**Branch**: `016-edge-auth` | **Date**: 2026-07-02 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/016-edge-auth/spec.md`

## Summary

Add generic per-RPC header layering to the core SDK (`WithExtraHeaders`) and an optional `openshell/v1/edge/` package containing a Cloudflare Access convenience constructor and a WebSocket tunnel proxy for gRPC behind edge proxies that reject HTTP/2 POST.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: google.golang.org/grpc (existing), github.com/coder/websocket v2 (new, edge package only)
**Storage**: N/A
**Testing**: Go testing + testify (assert/require), `make test` / `make ci`
**Target Platform**: Linux/macOS (server-side SDK)
**Project Type**: Library
**Performance Goals**: Header wrapper adds negligible overhead (map merge). Tunnel proxy supports 10+ concurrent gRPC streams.
**Constraints**: Zero new dependencies in core SDK. Edge package dependency (github.com/coder/websocket) isolated.
**Scale/Scope**: 3 new source files, 1 new package, ~400-600 lines of production code + tests.

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | No proto types exposed. AuthProvider is an idiomatic Go interface. |
| II. Idiomatic Go | PASS | Functional options (TunnelOption), error returns, context propagation. |
| III. Test-First | PASS | Tests written before implementation per Red-Green-Refactor. |
| IV. Upstream Tracking | N/A | No proto changes needed. |
| V. Minimal Dependencies | PASS w/ justification | github.com/coder/websocket added to edge package only. Stdlib has no WebSocket support. Core SDK: zero new deps. |
| VI. Secrets Never Leak | PASS | Edge tokens never appear in error messages or logs. Write-only fields. |
| VII. Deep Copy at Boundaries | PASS | Extra headers map deep-copied at construction time. |
| VIII. Doc Examples Compile | PASS | doc.go examples for both packages. |
| IX. Agent-Friendly Docs | PASS | All exported types/functions have doc comments with error codes. |
| X. Proto-SDK Naming | N/A | No proto counterparts for edge types. |
| XI. Fake-Real Parity | PASS | WithExtraHeaders wrapping fakes works without fake changes. Tunnel: not faked (transport concern). |
| XII. Graceful Shutdown | PASS | TunnelProxy.Close: drain, then cancel, then wait. Matches constitution order. |
| XIII. Docs Accompany Features | PASS | README update, doc.go for edge package, quickstart.md. |

## Project Structure

### Documentation (this feature)

```text
specs/016-edge-auth/
├── plan.md
├── research.md
├── data-model.md
├── quickstart.md
├── contracts/
│   └── public-api.go
└── tasks.md
```

### Source Code (repository root)

```text
openshell/v1/
├── auth.go              # Existing: NoAuth, StaticToken
├── auth_extra.go        # NEW: WithExtraHeaders wrapper
├── auth_extra_test.go   # NEW: Tests for WithExtraHeaders
├── auth_refresh.go      # Existing: RefreshableToken
└── edge/                # NEW PACKAGE
    ├── doc.go           # Package documentation with examples
    ├── cloudflare.go    # CloudflareAccess convenience constructor
    ├── cloudflare_test.go
    ├── tunnel.go        # TunnelProxy WebSocket tunnel
    └── tunnel_test.go
```

**Structure Decision**: `WithExtraHeaders` lives in the core `openshell/v1/` package alongside existing auth providers because it is a generic decorator. The `edge/` sub-package contains vendor-specific (Cloudflare) and transport-specific (WebSocket tunnel) code. Since this is a single-module repo, `github.com/coder/websocket` appears in go.mod, but only the `openshell/v1/edge/` package imports it. Consumers who do not import the edge package take no transitive WebSocket dependency.

## Complexity Tracking

No constitution violations requiring justification. The new dependency (github.com/coder/websocket) is justified under Constitution V because the stdlib lacks WebSocket support and the dependency is isolated to the optional edge package.
