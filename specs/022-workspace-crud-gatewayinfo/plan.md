# Implementation Plan: Workspace CRUD, GatewayInfo & GetCurrentUser

**Branch**: `022-workspace-crud-gatewayinfo` | **Date**: 2026-07-31 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/022-workspace-crud-gatewayinfo/spec.md`

## Summary

Add a new `Workspaces()` sub-client with 7 workspace/member CRUD operations, and extend the existing `Health()` sub-client with `GetGatewayInfo` and `GetCurrentUser`. All 9 RPCs already exist in the generated proto bindings. Implementation follows the established sub-client pattern: interface definition, gRPC client, converter, types, fake, tests.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: gRPC, protobuf (existing), testify (test only)
**Storage**: N/A (gRPC client SDK, no local storage)
**Testing**: Go testing + testify (assert/require), `make ci`
**Target Platform**: Any platform supported by Go
**Project Type**: Library (Go SDK)
**Performance Goals**: N/A (thin gRPC wrapper, performance is gateway-dependent)
**Constraints**: Proto isolation, deep copy at boundaries, fake-real parity
**Scale/Scope**: 9 new RPC methods, ~15 new files, ~6 new SDK types

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | All new types in `types/` package; converters map proto to SDK types |
| II. Idiomatic Go | PASS | Sub-client pattern, error returns, context propagation, ListOptions |
| III. Test-First | PASS | Each new method gets tests before implementation |
| IV. Upstream Tracking | PASS | All 9 RPCs exist in current proto bindings |
| V. Minimal Dependencies | PASS | No new dependencies required |
| VI. Secrets Never Leak | PASS | No credential fields in workspace/member/gateway types |
| VII. Deep Copy | PASS | Labels, annotations, roles, scopes, compute drivers all deep copied |
| VIII. Doc Examples | PASS | quickstart.md has compilable examples |
| IX. Agent-Friendly Docs | PASS | All interfaces document error codes per method |
| X. Proto-SDK Naming | PASS | Field names match proto semantics (PrincipalSubject, IdentityProvider) |
| XI. Fake-Real Parity | PASS | Fake validates same inputs as real client |
| XII. Graceful Shutdown | N/A | No streaming or shutdown behavior in new operations |
| XIII. Docs Accompany | PASS | README, doc.go updates included in plan |

## Project Structure

### Documentation (this feature)

```text
specs/022-workspace-crud-gatewayinfo/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── public-api.go
└── tasks.md             # Phase 2 output (speckit-tasks)
```

### Source Code (repository root)

```text
openshell/v1/
├── workspace.go                        # WorkspaceInterface + type aliases
├── workspace_client.go                 # gRPC client implementation
├── workspace_test.go                   # Unit tests (interface contract)
├── health.go                           # Extended HealthInterface (+ GetGatewayInfo, GetCurrentUser)
├── health_client.go                    # Extended health gRPC client
├── health_test.go                      # Tests for new health methods
├── client.go                           # + Workspaces() accessor + workspaces field
├── types/
│   ├── workspace.go                    # Workspace, WorkspaceMember, WorkspacePhase, WorkspaceRole
│   └── health.go                       # + GatewayInfo, ComputeDriverInfo, ServiceStatus, CurrentUser
├── internal/converter/
│   ├── workspace.go                    # Workspace/Member proto<->SDK converters
│   ├── workspace_test.go              # Converter round-trip tests
│   ├── health.go                       # GatewayInfo/CurrentUser converters (new file)
│   └── health_test.go                 # Converter round-trip tests
├── fake/
│   ├── workspace.go                    # Fake workspace + member store
│   ├── workspace_test.go              # Fake workspace tests
│   ├── health.go                       # + GetGatewayInfo, GetCurrentUser fake methods
│   ├── health_test.go                 # Tests for extended fake health
│   └── fake.go                         # + workspaceStore, WithGatewayInfo, WithCurrentUser
└── doc.go                              # Updated package examples
```

**Structure Decision**: Follows existing single-package SDK layout. New files for workspace sub-client, extended files for health sub-client. No new packages needed.

## Implementation Strategy

### Phase 1: Types & Converters

Define SDK types in `types/` and converter functions. This is the foundation that all other code depends on.

1. Add workspace types: `Workspace`, `WorkspaceMember`, `WorkspacePhase`, `WorkspaceRole`
2. Add gateway types: `GatewayInfo`, `ComputeDriverInfo`, `ServiceStatus`, `CurrentUser`
3. Write converter functions with deep copy (maps, slices)
4. Write converter round-trip tests

### Phase 2: Workspace Sub-client

Build the new `WorkspaceInterface` and its gRPC implementation.

1. Define `WorkspaceInterface` in `workspace.go` with type aliases
2. Implement `workspaceClient` in `workspace_client.go`
3. Add input validation (non-empty names, valid roles)
4. Wire into `Client` struct and `NewClient` constructor
5. Add `Workspaces()` to `ClientInterface`
6. Write unit tests

### Phase 3: Health Extension (GetGatewayInfo & GetCurrentUser)

Extend the existing `HealthInterface` with two new methods. US3 and US4 are combined into a single phase because both modify the same files (health.go, health_client.go, health_test.go).

1. Add `GetGatewayInfo` and `GetCurrentUser` to `HealthInterface`
2. Implement both in `healthClient`
3. Write unit tests for both methods

### Phase 4: Fake Client

Add fake implementations matching real client validation.

1. Add workspace fake with in-memory `objectStore` (workspaces are top-level, use empty workspace key; members are workspace-scoped)
2. Add member management to workspace fake
3. Add `GetGatewayInfo` and `GetCurrentUser` to fake health
4. Add `WithGatewayInfo`, `WithCurrentUser`, `WithWorkspaces` client options
5. Write fake-specific tests ensuring validation parity

### Phase 5: Documentation & CI

1. Update `doc.go` with workspace and gateway examples
2. Update `README.md` feature list
3. Run `make ci` to verify everything passes

## Complexity Tracking

No constitution violations to justify. All changes follow established patterns.
