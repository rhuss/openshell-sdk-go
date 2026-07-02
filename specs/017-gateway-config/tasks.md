# Tasks: Gateway Config Convenience Layer

**Input**: Design documents from `specs/017-gateway-config/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: Included (constitution mandates test-first for every public function).

**Organization**: Tasks grouped by user story for independent implementation.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1, US2, US3, US4)
- Paths relative to repository root

## Phase 1: Setup

**Purpose**: Package scaffolding and shared types

- [x] T001 Create package directory and doc.go with package documentation and runnable example in `openshell/v1/gateway/doc.go`
- [x] T002 [P] Define AuthMode, ConfigSource types and constants in `openshell/v1/gateway/config.go`
- [x] T003 [P] Define sentinel error variables (ErrGatewayNotFound, ErrConfigParse, ErrTokenLoad, ErrUnsupportedAuthMode, ErrInvalidGatewayName, ErrNoActiveGateway) in `openshell/v1/gateway/errors.go`
- [x] T004 [P] Define ClientOption type and option constructors (WithLogger, WithTimeout, WithTLS, WithAuth, WithRetryPolicy) in `openshell/v1/gateway/options.go`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Path resolution and gateway name validation that all user stories depend on

- [x] T005 Implement XDG config directory resolution (user dir with $XDG_CONFIG_HOME fallback, system dir /etc/openshell/gateways/) in `openshell/v1/gateway/paths.go`
- [x] T006 Implement gateway name validation (reject empty, path separators, dots, non-ASCII-alnum-dash-underscore) in `openshell/v1/gateway/paths.go`
- [x] T007 Write tests for XDG resolution with custom $XDG_CONFIG_HOME, unset fallback, and nonexistent directories in `openshell/v1/gateway/paths_test.go`
- [x] T008 Write tests for name validation covering valid names, path traversal, dots, empty, special characters in `openshell/v1/gateway/paths_test.go`
- [x] T009 Implement metadata.json parsing into Config struct (endpoint, auth_mode, name fields, ignore unknown fields) in `openshell/v1/gateway/config.go`
- [x] T010 Write tests for metadata.json parsing: valid config, missing fields, malformed JSON, unknown fields ignored in `openshell/v1/gateway/config_test.go`
- [x] T011 Write tests for error types verifying errors.Is works for all sentinel errors in `openshell/v1/gateway/errors_test.go`

**Checkpoint**: Foundation ready. Gateway directory resolution, name validation, and config parsing all tested.

**Internal Interfaces** (consumed by Phase 3+ tasks):

```go
// paths.go
func userConfigDir() (string, error)           // Returns $XDG_CONFIG_HOME/openshell or ~/.config/openshell
func systemGatewayDir() string                  // Returns /etc/openshell/gateways
func resolveGatewayDir(name string) (dir string, source ConfigSource, err error)  // Searches user then system
func validateGatewayName(name string) error     // Returns ErrInvalidGatewayName or nil

// config.go
func parseMetadata(dir string) (*Config, error) // Reads metadata.json from dir, returns parsed Config
```

---

## Phase 3: User Story 1 - Connect to Named Gateway (Priority: P1)

**Goal**: A Go program calls NewClient("my-gateway") and gets a working SDK client.

**Independent Test**: Create temp gateway config directory with metadata.json, call NewClient, verify returned client has correct auth provider.

- [ ] T012 [US1] Implement lazy edge token loading (read edge_token file on first call, cf_token fallback, sync.Once guard) in `openshell/v1/gateway/token.go`
- [ ] T013 [US1] Implement diskTokenSource (oauth2.TokenSource reading oidc_token.json, parsing access_token/refresh_token/expiry) in `openshell/v1/gateway/token.go`
- [ ] T014 [US1] Write tests for edge token loading: file exists, cf_token fallback, missing file, lazy behavior in `openshell/v1/gateway/token_test.go`
- [ ] T015 [US1] Write tests for diskTokenSource: valid bundle, missing file, malformed JSON, expiry parsing in `openshell/v1/gateway/token_test.go`
- [ ] T016 [US1] Implement auth mode resolution (map AuthMode to v1.AuthProvider using token loaders, return error for mtls/unknown) in `openshell/v1/gateway/gateway.go`
- [ ] T017 [US1] Implement NewClient function (resolve gateway dir, parse config, resolve auth, apply ClientOptions, call v1.NewClient) in `openshell/v1/gateway/gateway.go`
- [ ] T018 [US1] Write tests for NewClient with named gateway: each auth mode (none, plaintext, cloudflare_jwt, oidc), not-found error, invalid name error in `openshell/v1/gateway/gateway_test.go`
- [ ] T019 [US1] Write test verifying tokens/credentials never appear in error messages from NewClient in `openshell/v1/gateway/gateway_test.go`

**Checkpoint**: Named gateway connectivity works for all auth modes. MVP complete.

**Internal Interfaces** (consumed by Phase 4+ tasks):

```go
// token.go
func loadEdgeToken(dir string) (string, error)          // Reads edge_token (cf_token fallback), lazy via sync.Once
func newDiskTokenSource(dir string) oauth2.TokenSource   // Reads oidc_token.json on first Token() call

// gateway.go
func resolveAuthProvider(cfg *Config) (v1.AuthProvider, error)  // Maps AuthMode to provider using token loaders
```

---

## Phase 4: User Story 2 - Connect to Active Gateway (Priority: P1)

**Goal**: A Go program calls NewClient("") and connects to the active gateway.

**Independent Test**: Create active_gateway file pointing to a configured gateway, call NewClient(""), verify connection.

- [ ] T020 [US2] Implement active gateway resolution (read active_gateway file, trim whitespace, validate name, resolve to config dir) in `openshell/v1/gateway/paths.go`
- [ ] T021 [US2] Write tests for active gateway: file exists with valid name, file missing (ErrNoActiveGateway), empty file, whitespace handling in `openshell/v1/gateway/paths_test.go`
- [ ] T022 [US2] Wire empty-name handling in NewClient and LoadConfig to use active gateway resolution in `openshell/v1/gateway/gateway.go`
- [ ] T023 [US2] Write tests for NewClient("") and LoadConfig("") resolving active gateway correctly in `openshell/v1/gateway/gateway_test.go`

**Checkpoint**: Active gateway resolution works. Both P1 stories complete.

**Internal Interfaces** (consumed by Phase 5+ tasks):

```go
// paths.go
func resolveActiveGateway() (string, error)  // Reads active_gateway file, returns validated name or ErrNoActiveGateway
```

---

## Phase 5: User Story 3 - Load Config for Manual Wiring (Priority: P2)

**Goal**: A Go program calls LoadConfig("my-gateway") and gets a Config struct for custom wiring.

**Independent Test**: Create gateway config, call LoadConfig, verify all fields match on-disk metadata.

- [ ] T024 [US3] Implement LoadConfig function (resolve dir, parse metadata, return Config snapshot without creating client) in `openshell/v1/gateway/gateway.go`
- [ ] T025 [US3] Write tests for LoadConfig: valid config fields, source (user vs system), frozen snapshot behavior, active gateway resolution in `openshell/v1/gateway/gateway_test.go`

**Checkpoint**: Config inspection works independently.

---

## Phase 6: User Story 4 - List Available Gateways (Priority: P3)

**Goal**: A Go program calls ListGateways() and gets all configured gateways.

**Independent Test**: Create multiple gateway dirs in user and system paths, call ListGateways, verify all returned.

- [ ] T026 [US4] Implement ListGateways function (scan user and system dirs, deduplicate by name with user precedence, resolve active status) in `openshell/v1/gateway/gateway.go`
- [ ] T027 [US4] Write tests for ListGateways: multiple gateways, empty dirs, user/system precedence, active status, source field in `openshell/v1/gateway/gateway_test.go`

**Checkpoint**: Gateway enumeration works.

---

## Phase 7: Polish & Cross-Cutting

**Purpose**: Documentation, thread safety verification, and CI integration

- [ ] T028 [P] Add doc comments to all exported types and functions, verify go doc output in `openshell/v1/gateway/*.go`
- [ ] T029 [P] Add runnable Example functions (ExampleNewClient, ExampleLoadConfig, ExampleListGateways) in `openshell/v1/gateway/doc.go`
- [ ] T030 [P] Write concurrent access test (multiple goroutines calling NewClient, LoadConfig, ListGateways simultaneously) in `openshell/v1/gateway/gateway_test.go`
- [ ] T031 Update README.md with gateway package section showing usage examples
- [ ] T032 [P] Write benchmark test (BenchmarkLoadConfig, BenchmarkListGateways) verifying config loading completes under 10ms with 10 gateways in `openshell/v1/gateway/gateway_test.go`
- [ ] T033 Run `make ci` to verify lint, build, and all tests pass

---

## Dependencies

```
Phase 1 (Setup) ─────► Phase 2 (Foundation) ─────► Phase 3 (US1: Named Gateway)
                                                          │
                                                          ├──► Phase 4 (US2: Active Gateway)
                                                          │
                                                          ├──► Phase 5 (US3: Load Config)
                                                          │
                                                          └──► Phase 6 (US4: List Gateways)
                                                                      │
                                                                      ▼
                                                              Phase 7 (Polish)
```

- US1 (Phase 3) blocks US2, US3, US4 because it implements the core auth resolution and NewClient
- US2, US3, US4 (Phases 4, 5, 6) can run in parallel after US1 completes
- Phase 7 runs after all user stories complete

## Parallel Execution Opportunities

- T002, T003, T004 (Phase 1): Independent type files, no dependencies
- T007 + T008 (Phase 2): Test files for independent functions
- After Phase 3 completes: Phases 4, 5, 6 can run in parallel
- T028, T029, T030 (Phase 7): Independent polish tasks

## Implementation Strategy

**MVP**: Phase 1 + Phase 2 + Phase 3 (US1) = Named gateway connectivity with all auth modes. Delivers the core value proposition in 19 tasks.

**Full delivery**: All 33 tasks across 7 phases. US2-US4 are incremental additions after MVP.
