# Implementation Plan: OIDC Login Package

**Branch**: `018-oidc-login` | **Date**: 2026-07-03 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/018-oidc-login/spec.md`

## Summary

Add an `openshell/v1/oidc/` package that handles OIDC authentication flows (auth code + PKCE, keyboard fallback, device code, client credentials) with gateway-aware login. The package discovers OIDC providers from gateway metadata, persists tokens to disk in the existing `oidcBundle` format, and returns `*oauth2.Token` compatible with the SDK's `RefreshableToken` and `diskTokenSource`. No new runtime dependencies required.

## Technical Context

**Language/Version**: Go 1.25
**Primary Dependencies**: `golang.org/x/oauth2` (already in go.mod), `net/http` (stdlib), `crypto/rand` + `crypto/sha256` (stdlib for PKCE)
**Storage**: File-based (`oidc_token.json` in gateway config directory)
**Testing**: Go testing + testify (assert/require), `//go:build integration` for integration tests
**Target Platform**: macOS, Linux, Windows
**Project Type**: Library (Go SDK package)
**Performance Goals**: Client credentials exchange < 2s excluding network latency
**Constraints**: Zero new runtime dependencies (Constitution V), secrets never in errors (Constitution VI)
**Scale/Scope**: Single package `openshell/v1/oidc/` with ~10 exported functions/types

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | N/A | No proto types involved; pure HTTP/OIDC |
| II. Idiomatic Go | PASS | Functional options, context.Context first param, error returns |
| III. Test-First | PASS | Tests written before implementation per tasks ordering |
| IV. Upstream Tracking | N/A | New package, no upstream proto dependency |
| V. Minimal Dependencies | PASS | Zero new deps; uses stdlib (net/http, crypto) + existing oauth2 |
| VI. Secrets Never Leak | PASS | FR-014 enforces; tokens/codes redacted from all errors |
| VII. Deep Copy at Boundaries | PASS | Token structs are value types; no shared mutable references |
| VIII. Doc Examples Compile | PASS | doc.go with runnable examples required per tasks |
| IX. Agent-Friendly Docs | PASS | All exports get doc comments with error code lists |
| X. Proto-SDK Naming | N/A | No proto types to map |
| XI. Fake-Real Parity | PASS | httptest.Server mocks OIDC provider in tests |
| XII. Graceful Shutdown | PASS | Callback server shutdown before context cancel (FR-020) |
| XIII. Docs Accompany Features | PASS | README update + doc.go in same PR |

## Project Structure

### Documentation (this feature)

```text
specs/018-oidc-login/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── api.md           # Public API contract
└── tasks.md             # Phase 2 output (by /speckit-tasks)
```

### Source Code (repository root)

```text
openshell/v1/oidc/
├── doc.go               # Package docs + runnable examples
├── oidc.go              # Login, DeviceLogin, ClientCredentials entry points
├── options.go           # LoginOption type and With* functions
├── discovery.go         # OIDC discovery document fetch + cache
├── authcode.go          # Auth code + PKCE flow (browser + callback server)
├── keyboard.go          # Keyboard fallback flow
├── device.go            # Device code flow (RFC 8628)
├── credentials.go       # Client credentials grant
├── token.go             # Token persistence (read/write oidcBundle)
├── browser.go           # Platform browser opener (xdg-open/open/cmd start)
├── errors.go            # Typed error definitions
├── oidc_test.go         # Tests for Login entry points
├── options_test.go      # Tests for option application
├── discovery_test.go    # Tests for discovery fetch + cache
├── authcode_test.go     # Tests for auth code flow
├── keyboard_test.go     # Tests for keyboard flow
├── device_test.go       # Tests for device code flow
├── credentials_test.go  # Tests for client credentials
├── token_test.go        # Tests for token persistence
├── browser_test.go      # Tests for browser opener
└── errors_test.go       # Tests for error types

openshell/v1/gateway/
├── config.go            # Extended: add OIDCIssuer, OIDCClientID to metadataJSON and public Config struct
└── (existing files unchanged)
```

**Structure Decision**: Single flat package `openshell/v1/oidc/` following the existing SDK layout pattern (e.g., `openshell/v1/gateway/`). No sub-packages needed. Internal types are unexported. The gateway package gets a minor extension to `metadataJSON` for OIDC fields.

## Complexity Tracking

No constitution violations. Zero new dependencies. Single new package with clear boundaries.
