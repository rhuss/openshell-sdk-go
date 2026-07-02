# Tasks: Edge Auth (Extra Headers + WebSocket Tunnel)

**Input**: Design documents from `specs/016-edge-auth/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included per Constitution III (Test-First, NON-NEGOTIABLE).

**Organization**: Tasks grouped by user story for independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create the edge package directory and prepare the project structure.

- [x] T001 Create edge package directory at `openshell/v1/edge/`

---

## Phase 2: User Story 1 - Attach Extra Headers to Any Auth Provider (Priority: P1) MVP

**Goal**: Provide a generic `WithExtraHeaders` decorator that wraps any `AuthProvider` with additional static per-RPC headers.

**Independent Test**: Create a wrapped auth provider with custom headers and verify both base and extra headers appear on outgoing RPCs, with extra headers winning on key collision.

- [x] T002 [US1] Write tests for `WithExtraHeaders` in `openshell/v1/auth_extra_test.go`: nil base error, nil/empty headers error, header merge with base provider, extra headers precedence on key collision (case-insensitive), empty-string values silently skipped, `RequireTransportSecurity` delegation, composition with `NoAuth`/`StaticToken`/`RefreshableToken`
- [x] T003 [US1] Implement `WithExtraHeaders(base AuthProvider, headers map[string]string) (AuthProvider, error)` in `openshell/v1/auth_extra.go`: lowercase-normalize keys, deep-copy headers map, filter empty values, merge in `GetRequestMetadata` with extra-wins precedence, delegate `RequireTransportSecurity` to base
- [x] T004 [US1] Add `WithExtraHeaders` example to `openshell/v1/doc.go`

**Checkpoint**: `make ci` passes. `WithExtraHeaders` wraps any auth provider with extra headers. US2 and US3 can now begin.

---

## Phase 3: User Story 2 - Cloudflare Access Convenience (Priority: P2)

**Goal**: Provide a `CloudflareAccess` convenience constructor that formats CF-specific headers from a single edge token parameter.

**Independent Test**: Create a Cloudflare Access auth provider and verify `cf-access-jwt-assertion` header and `CF_Authorization` cookie are set correctly.

- [x] T005 [P] [US2] Write tests for `CloudflareAccess` in `openshell/v1/edge/cloudflare_test.go`: valid token produces correct headers (`cf-access-jwt-assertion` and `cookie: CF_Authorization=<token>`), empty token returns error, base auth headers are preserved alongside CF headers
- [x] T006 [US2] Implement `CloudflareAccess(baseAuth AuthProvider, edgeToken string) (AuthProvider, error)` in `openshell/v1/edge/cloudflare.go`: validate non-empty token, delegate to `WithExtraHeaders` with CF-specific header names
- [x] T007 [US2] Create edge package documentation `openshell/v1/edge/doc.go` with `CloudflareAccess` example

**Checkpoint**: `make ci` passes. Cloudflare Access integration works with any base auth provider.

---

## Phase 4: User Story 3 - WebSocket Tunnel for gRPC Behind Edge Proxies (Priority: P3)

**Goal**: Provide a `TunnelProxy` that bridges gRPC connections over WebSocket for edge proxies that reject HTTP/2 POST.

**Independent Test**: Create a tunnel proxy pointed at a test WebSocket server and verify gRPC traffic is forwarded through the WebSocket connection.

- [x] T008 [US3] Add `nhooyr.io/websocket` dependency via `go get nhooyr.io/websocket@latest`
- [x] T009 [US3] Write tests for `TunnelProxy` in `openshell/v1/edge/tunnel_test.go`: creation with valid URL, invalid/empty URL error, `Addr()` returns dialable address, `Close()` on unused proxy returns nil, `Close()` drains in-flight connections, `Close()` force-closes after timeout, concurrent `Close()` from multiple goroutines is safe (second call returns immediately), goroutine cleanup verified with runtime.NumGoroutine, concurrent streams (at least 10), TLS option configures wss://, logger option receives log events
- [x] T010 [US3] Implement `TunnelProxy` and `TunnelOption` types in `openshell/v1/edge/tunnel.go`: local `net.Listener`, accept loop spawning goroutine-per-connection with WebSocket dial to gateway, bidirectional `io.Copy` bridge, `sync.WaitGroup` tracking, `Close()` with configurable drain timeout (default 5s) then force-close, `WithTunnelLogger`/`WithTunnelTLS`/`WithCloseTimeout` functional options
- [x] T011 [US3] Add `TunnelProxy` example to `openshell/v1/edge/doc.go`

**Checkpoint**: `make ci` passes. WebSocket tunnel proxy bridges gRPC over WebSocket with graceful shutdown.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Documentation updates and final quality checks.

- [x] T012 [P] Update `README.md` with edge auth section: `WithExtraHeaders` usage, Cloudflare Access example, WebSocket tunnel example
- [x] T013 Run `make ci` and fix any lint, vet, or test issues

---

## Dependencies

```
T001 (setup) ─────┐
                   ▼
T002→T003→T004 (US1: WithExtraHeaders)
                   │
         ┌─────────┼─────────┐
         ▼                   ▼
T005→T006→T007         T008→T009→T010→T011
(US2: CF Access)       (US3: WS Tunnel)
         │                   │
         └─────────┬─────────┘
                   ▼
            T012, T013 (Polish)
```

**Parallel opportunities**:
- T005 (US2 tests) can start in parallel with T008-T009 (US3 setup/tests) after US1 completes
- T012 (README) can run in parallel with T013 (CI check)

## Implementation Strategy

**MVP**: Phase 1 + Phase 2 (Setup + US1: WithExtraHeaders). Delivers the core header layering mechanism that all edge auth scenarios depend on. Can be merged independently.

**Incremental delivery**: Each phase is an independently testable increment. US2 (Cloudflare Access) adds convenience on top of US1. US3 (WebSocket Tunnel) adds transport tunneling. Both can be delivered separately if needed.
