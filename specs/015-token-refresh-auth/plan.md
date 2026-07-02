# Implementation Plan: Composable Token Refresh with Coalesced Caching

**Branch**: `015-token-refresh-auth` | **Date**: 2026-07-01 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/015-token-refresh-auth/spec.md`

## Summary

Add a `RefreshableToken` constructor to the Go SDK that wraps an `oauth2.TokenSource` into an `AuthProvider` with RWMutex-based coalesced caching. The implementation follows the Kubernetes client-go `cachingTokenSource` pattern: fast-path read under RLock, slow-path refresh under Lock with double-checked locking, configurable leeway, and graceful degradation on refresh failure.

## Technical Context

**Language/Version**: Go 1.25+
**Primary Dependencies**: `google.golang.org/grpc`, `golang.org/x/oauth2` (new, justified: standard Go token interface used by k8s client-go, GCP, AWS SDKs; maintained by Go team)
**Storage**: N/A (memory-only caching, no filesystem)
**Testing**: Go testing + testify (assert/require), `make test`
**Target Platform**: Any platform supported by Go (cross-compiles without cgo)
**Project Type**: Library (Go SDK)
**Performance Goals**: <1μs fast-path overhead, single TokenSource call under concurrent load
**Constraints**: No filesystem operations, no new non-stdlib dependencies beyond oauth2
**Scale/Scope**: ~100-150 lines of non-test code, ~200-300 lines of tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | No proto types involved; AuthProvider is a pure Go interface |
| II. Idiomatic Go | PASS | Functional options, standard oauth2.TokenSource, RWMutex pattern |
| III. Test-First | PASS | Tests written alongside implementation per spec |
| IV. Upstream Tracking | N/A | No proto changes |
| V. Minimal Dependencies | PASS (with justification) | `golang.org/x/oauth2` is the standard Go interface for token management, used by k8s client-go and all major cloud SDKs. Maintained by the Go team. No alternative exists in stdlib. |
| VI. Secrets Never Leak | PASS | Tokens never appear in error messages or log output. FR-005 warning logs mention "token refresh failed", not the token value. |
| VII. Deep Copy at Boundaries | N/A | No proto/SDK boundary crossing |
| VIII. Doc Examples Compile | PASS | doc.go examples will use actual signatures |
| IX. Agent-Friendly Documentation | PASS | All exported symbols get doc comments |
| X. Proto-SDK Naming Fidelity | N/A | No proto type mapping |
| XI. Fake-Real Parity | N/A | RefreshableToken is a constructor, not an interface method; no fake needed |
| XII. Graceful Shutdown Order | N/A | No background goroutines or streams |

## Project Structure

### Documentation (this feature)

```text
specs/015-token-refresh-auth/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Quality checklist
└── tasks.md             # Phase 2 output (created by /speckit-tasks)
```

### Source Code (repository root)

```text
openshell/v1/
├── auth.go              # EXISTING: NoAuth, StaticToken (unchanged)
├── auth_refresh.go      # NEW: RefreshableToken constructor, refreshableAuth struct, RefreshOption type
├── auth_refresh_test.go # NEW: unit tests for refresh, concurrency, leeway, graceful degradation
├── types/
│   ├── auth.go          # EXISTING: AuthProvider interface (unchanged)
│   └── logger.go        # EXISTING: Logger interface (unchanged)
└── doc.go               # EXISTING: update with RefreshableToken example
```

**Structure Decision**: Two new files (`auth_refresh.go`, `auth_refresh_test.go`) in the existing `openshell/v1/` package. No new packages or directories needed. The refreshable auth is a peer of the existing `auth.go` implementations.
