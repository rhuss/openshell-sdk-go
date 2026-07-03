# Code Review: OIDC Login Package (018-oidc-login)

**Date**: 2026-07-03
**Reviewer**: Automated pipeline (speckit-spex-ship)
**Branch**: `018-oidc-login`

## Compliance Score

**20/20 FRs implemented = 100%**

| FR | Description | Implementing File | Status |
|----|-------------|-------------------|--------|
| FR-001 | OIDC discovery fetch | discovery.go | PASS |
| FR-002 | Discovery cache | discovery.go | PASS |
| FR-003 | Auth code + PKCE + state | authcode.go | PASS |
| FR-004 | Callback server ports | authcode.go, options.go | PASS |
| FR-005 | Browser open | browser.go | PASS |
| FR-006 | Keyboard fallback | keyboard.go | PASS |
| FR-007 | Device code flow | device.go | PASS |
| FR-008 | Client credentials | credentials.go | PASS |
| FR-009 | Gateway metadata OIDC fields | gateway/config.go, oidc.go | PASS |
| FR-010 | Token persistence | token.go | PASS |
| FR-011 | In-memory mode | oidc.go, options.go | PASS |
| FR-012 | oauth2.Token compatibility | oidc.go, credentials.go | PASS |
| FR-013 | Configurable timeout | oidc.go, options.go | PASS |
| FR-014 | Secrets never leak | errors.go, all flow files | PASS |
| FR-015 | Functional options | options.go | PASS |
| FR-016 | Configurable scopes | options.go | PASS |
| FR-017 | Display callback | device.go, options.go | PASS |
| FR-018 | context.Context first param | oidc.go, device.go, credentials.go | PASS |
| FR-019 | Token reuse check | oidc.go | PASS |
| FR-020 | Server lifecycle scoped | authcode.go | PASS |

## Gate Outcome

**PASS**

- Lint: 0 issues in `openshell/v1/oidc/` and `openshell/v1/gateway/`
- Build: PASS
- Tests: 103 test functions, all passing (22.7s)
- Coverage: 13.5% of total statements (expected for new package with existing large codebase)

## Deep Review Report

### Correctness Review

All three public entry points (`Login`, `DeviceLogin`, `ClientCredentials`) follow the spec:
- `Login` resolves gateway config, checks existing tokens (FR-019), runs auth code + PKCE or keyboard fallback
- `DeviceLogin` implements RFC 8628 with `authorization_pending`, `slow_down`, `expired_token` handling
- `ClientCredentials` exchanges client ID + secret without user interaction

Token persistence writes the `oidcBundle` format (access_token, refresh_token, expiry, expires_in) matching `gateway/token.go:oidcBundle`. Interop with `diskTokenSource` confirmed by schema parity.

**Finding**: No correctness issues found.

### Architecture Review

Package structure follows SDK conventions:
- Single flat package `openshell/v1/oidc/` (consistent with `gateway/`, `edge/`, `fake/`)
- Functional options pattern via `LoginOption` (consistent with `gateway.ClientOption`, `v1.RefreshOption`)
- Sentinel errors with `fmt.Errorf("%w: ...")` wrapping (consistent with `gateway/errors.go`)
- Zero new runtime dependencies (Constitution V satisfied)

The gateway config extension adds `OIDCIssuer` and `OIDCClientID` to both `metadataJSON` and public `Config`, maintaining backward compatibility (empty strings for non-OIDC gateways).

**Finding**: No architectural issues found.

### Security Review

- PKCE (S256) generated via `crypto/rand` + `crypto/sha256` (cryptographically secure)
- State parameter generated via `crypto/rand` (16 bytes, base64url-encoded)
- State validated on callback receipt (CSRF protection)
- Error messages never include tokens, secrets, or authorization codes (FR-014)
- Client secret is write-only (set via `WithClientSecret`, never returned or logged)
- Callback server binds to `127.0.0.1` only (not `0.0.0.0`)

**Finding**: No security issues found.

### Production Readiness Review

- Context propagation: all public functions accept `context.Context` (FR-018)
- Graceful shutdown: callback server uses `http.Server.Shutdown(ctx)` before context cancel (Constitution XII)
- Timeout enforcement: default 2-minute timeout via `context.WithTimeout` (FR-013)
- Concurrent safety: discovery cache uses `sync.Once`; each login attempt is independent
- Error types: typed sentinels enable programmatic error handling via `errors.Is()`

**Finding**: No production readiness issues found.

### Test Quality Review

- 103 test functions covering all flows (auth code, keyboard, device code, client credentials)
- Tests use `httptest.Server` to mock OIDC provider (no network dependencies)
- Error paths tested (invalid credentials, timeout, discovery failure)
- Token persistence roundtrip tested (write then read)
- Gateway config backward compatibility tested (missing OIDC fields)
- PKCE generation and S256 challenge tested
- State validation tested (valid + mismatch cases)

**Finding**: No test quality issues found.

## Pre-existing Issues (Not From This Change)

- `docs:check` fails for `edge` and `fake` packages (missing docs pages, pre-existing)
- Lint issues in other packages (19 violations across errcheck, revive, staticcheck, pre-existing)

## Summary

The OIDC login package implementation is complete and spec-compliant. All 20 functional requirements are implemented, all 103 tests pass, zero lint issues in the new code, and no security concerns. The package follows all constitution principles (minimal deps, secrets never leak, test-first, graceful shutdown, doc comments on exports). Ready for merge.
