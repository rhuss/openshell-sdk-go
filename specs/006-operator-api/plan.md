# Implementation Plan: Operator API Extensions (Phase 2a)

**Branch**: `006-operator-api` | **Date**: 2026-06-28 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/006-operator-api/spec.md`

## Summary

Extend the SDK with 14 new RPCs covering service exposure (4), provider profiles (6), and credential refresh (4), plus a StopOnTerminal watch enhancement and fake client stubs. ServiceInterface is a new top-level sub-client; ProfileInterface and RefreshInterface are nested under the existing ProviderInterface.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: None new — SDK itself (`v1/types/`), gRPC, Go stdlib
**Storage**: N/A (SDK wraps gateway RPCs)
**Testing**: Go testing + testify (assert/require), `go test -race`
**Target Platform**: Go library (any OS)
**Project Type**: Library (SDK extension)
**Constraints**: Zero new dependencies, thread-safe, deep copy at boundaries, proto isolation
**Scale/Scope**: ~1500-2000 lines implementation + ~1000-1500 lines tests

## Constitution Check

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | All new domain types in `v1/types/`, converters in `internal/converter/` |
| II. Idiomatic Go | PASS | Interfaces, functional options pattern where appropriate, context propagation |
| III. Test-First (NON-NEGOTIABLE) | PASS | Tests written before each implementation file |
| IV. Upstream Tracking | PASS | All 14 RPCs verified in proto/openshell.proto |
| V. Minimal Dependencies | PASS | Zero new dependencies |
| VI. Secrets Never Leak | PASS | RefreshConfig.Material with secret keys — passed through, never logged |
| VII. Deep Copy at Boundaries | PASS | All converter and fake operations deep-copy |
| VIII. Doc Examples Compile | PASS | Updated doc.go examples for new sub-clients |

## Project Structure

### Documentation (this feature)

```text
specs/006-operator-api/
├── plan.md              # This file
├── research.md          # Research decisions
├── data-model.md        # Entity model
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Task breakdown (created by /speckit.tasks)
```

### Source Code (repository root)

```text
openshell/v1/
├── types/
│   ├── service.go           # ServiceEndpoint type
│   ├── profile.go           # ProviderProfile, ProfileCredential, ProfileCategory,
│   │                        # ProfileDiscovery, ProfileImportItem, ProfileDiagnostic,
│   │                        # ImportResult, UpdateResult, LintResult,
│   │                        # NetworkEndpoint, NetworkBinary
│   ├── refresh.go           # RefreshStrategy, RefreshStatus, RefreshConfig
│   └── options.go           # Add StopOnTerminal to WatchOptions (existing file)
├── internal/
│   ├── converter/
│   │   ├── service.go       # ServiceEndpoint proto↔SDK converter
│   │   ├── service_test.go
│   │   ├── profile.go       # ProviderProfile proto↔SDK converter
│   │   ├── profile_test.go
│   │   ├── refresh.go       # RefreshStatus/Config proto↔SDK converter
│   │   └── refresh_test.go
│   └── grpc/
│       ├── service.go       # gRPC service client wrapper
│       ├── profile.go       # gRPC profile client wrapper
│       └── refresh.go       # gRPC refresh client wrapper
├── service.go               # ServiceInterface definition
├── service_client.go        # serviceClient implementation
├── service_client_test.go   # Service client tests (mock gRPC)
├── profile.go               # ProfileInterface definition
├── profile_client.go        # profileClient implementation
├── profile_client_test.go   # Profile client tests (mock gRPC)
├── refresh.go               # RefreshInterface definition
├── refresh_client.go        # refreshClient implementation
├── refresh_client_test.go   # Refresh client tests (mock gRPC)
├── client.go                # Add Services() to ClientInterface (existing)
├── provider.go              # Add Profiles(), Refresh() to ProviderInterface (existing)
├── provider_client.go       # Wire profile/refresh sub-clients (existing)
└── sandbox_client.go        # Add StopOnTerminal support to Watch (existing)

openshell/v1/fake/
├── service.go               # fakeServiceClient stub
├── service_test.go
├── profile.go               # fakeProfileClient stub
├── profile_test.go
├── refresh.go               # fakeRefreshClient stub
├── refresh_test.go
├── fake.go                  # Add Services() accessor (existing)
├── provider.go              # Add Profiles(), Refresh() accessors (existing)
└── sandbox.go               # Add StopOnTerminal support to fake Watch (existing)
```

## Design Decisions

### D1: Interface Extension Pattern

`ClientInterface` gains `Services() ServiceInterface`. `ProviderInterface` gains `Profiles() ProfileInterface` and `Refresh() RefreshInterface`. These are additive — no existing methods change signatures.

### D2: ServiceEndpoint Flattening

The proto uses `ServiceEndpointResponse{endpoint, url}` wrapping `ServiceEndpoint{...}`. The SDK flattens both into a single `ServiceEndpoint` type with all fields including URL. The converter handles the unwrapping.

### D3: Profile Category as String Enum

`ProfileCategory` uses string constants (`ProfileCategoryInference`, etc.) matching the Phase 1 pattern for `SandboxPhase` and `EventType`. More ergonomic than iota-based int enums.

### D4: RefreshConfig Accepts All Fields Directly

Rather than a builder pattern, `RefreshConfig` is a plain struct with all fields. The Configure method accepts it directly. This matches the Phase 1 pattern for `CreateOptions` and other config structs.

### D5: StopOnTerminal Dual Implementation

The SDK passes `stop_on_terminal` to the server AND monitors events client-side. If the server closes the stream, the SDK handles it normally. If the server doesn't (older gateway), the SDK closes the watcher after detecting a terminal phase. Defense-in-depth with no added complexity.

### D6: NetworkEndpoint/NetworkBinary SDK Types

These sandbox.proto types appear in ProviderProfile. Rather than importing proto types (violates Proto Isolation), we create lightweight SDK types. They're small (3-5 fields) and unlikely to change frequently.

## Global Constraints

- **FR-024**: All operations MUST be safe for concurrent use.
- **FR-022**: All domain types in `v1/types/`, converters import types not clients.
- **FR-023**: Typed StatusError with appropriate ErrorCode for all error paths.
- **Constitution III**: Tests written before each implementation file.
- **SPDX headers**: Every `.go` file must start with the SPDX license header.
- **Deep copy at boundaries**: All converter operations deep-copy maps and slices.

## Implementation Order

Bottom-up: types → converters → gRPC wrappers → client implementations → interface wiring → fake stubs → watch enhancement.

1. **Types**: New domain types in `v1/types/` (service, profile, refresh)
2. **Converters**: Proto↔SDK converters for all new types
3. **Service sub-client**: ServiceInterface + serviceClient + gRPC wrapper
4. **Profile sub-client**: ProfileInterface + profileClient + gRPC wrapper
5. **Refresh sub-client**: RefreshInterface + refreshClient + gRPC wrapper
6. **Interface wiring**: Add Services() to ClientInterface, Profiles()/Refresh() to ProviderInterface
7. **Watch enhancement**: StopOnTerminal in WatchOptions + sandbox client
8. **Fake stubs**: Service, profile, refresh stubs + fake wiring
9. **Polish**: doc.go updates, lint, CI verification
