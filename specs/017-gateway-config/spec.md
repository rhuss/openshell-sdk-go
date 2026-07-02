# Feature Specification: Gateway Config Convenience Layer

**Feature Branch**: `017-gateway-config`
**Created**: 2026-07-02
**Status**: Draft
**Input**: Brainstorm 019 - CLI Auth Convenience Layer

## User Scenarios & Testing

### User Story 1 - Connect to a Named Gateway (Priority: P1)

A Go developer building automation tooling wants to connect to a gateway
that was previously configured via `openshell gateway add`. They provide
the gateway name and get a fully configured, ready-to-use SDK client
without manually parsing config files, loading tokens, or wiring auth
providers.

**Why this priority**: This is the core value proposition. Without this,
every Go program must duplicate 20+ lines of boilerplate to connect to
a gateway.

**Independent Test**: Can be tested by configuring a gateway via the Rust
CLI, then calling the convenience constructor from Go and verifying
a working client is returned.

**Acceptance Scenarios**:

1. **Given** a gateway named "prod" was configured via `openshell gateway add`,
   **When** a Go program calls the convenience constructor with name "prod",
   **Then** it receives a connected SDK client with the correct auth provider
   for that gateway's auth mode.

2. **Given** a gateway named "prod" uses `cloudflare_jwt` auth mode,
   **When** a Go program calls the convenience constructor with name "prod",
   **Then** the returned client uses edge-token-based auth loaded from
   the gateway's token file on disk.

3. **Given** no gateway exists with name "nonexistent",
   **When** a Go program calls the convenience constructor with that name,
   **Then** it receives a clear error indicating the gateway was not found.

---

### User Story 2 - Connect to the Active Gateway (Priority: P1)

A Go developer wants to connect to whichever gateway is currently
marked as active (set via `openshell gateway use`), without knowing
or hardcoding its name.

**Why this priority**: Active gateway resolution removes the need to
pass gateway names through config or environment variables, matching
the Rust CLI's default behavior.

**Independent Test**: Can be tested by setting an active gateway, then
calling the constructor with an empty name and verifying it resolves
correctly.

**Acceptance Scenarios**:

1. **Given** gateway "staging" is set as the active gateway,
   **When** a Go program calls the convenience constructor with an empty name,
   **Then** it receives a client connected to the "staging" gateway.

2. **Given** no active gateway is set,
   **When** a Go program calls the convenience constructor with an empty name,
   **Then** it receives a clear error indicating no active gateway is configured.

---

### User Story 3 - Load Gateway Config for Manual Wiring (Priority: P2)

A Go developer wants to inspect a gateway's configuration (endpoint,
auth mode, token paths) without creating a client, so they can customize
the connection setup or use the config for non-standard purposes.

**Why this priority**: Some users need finer control than a one-call
constructor provides, but still want to avoid parsing config files
manually.

**Independent Test**: Can be tested by loading config for a known gateway
and verifying all fields (endpoint, auth mode, name) match the on-disk
metadata.

**Acceptance Scenarios**:

1. **Given** a gateway "dev" is configured with endpoint "localhost:8080"
   and auth mode "plaintext",
   **When** a Go program loads the config for "dev",
   **Then** it receives a config struct with the correct endpoint, auth mode,
   and gateway name.

2. **Given** a gateway "secure" uses OIDC auth,
   **When** a Go program loads the config for "secure",
   **Then** the config includes the OIDC token bundle path but does not
   eagerly load or validate the token contents.

---

### User Story 4 - List Available Gateways (Priority: P3)

A Go developer building a CLI or UI wants to enumerate all configured
gateways to present them to users for selection (tab-completion,
dropdown menu, status display).

**Why this priority**: Enumeration is a convenience feature for
interactive tools. Core connectivity (P1) and config inspection (P2)
deliver more immediate value.

**Independent Test**: Can be tested by configuring multiple gateways,
calling the list function, and verifying all are returned with correct
names and source information.

**Acceptance Scenarios**:

1. **Given** gateways "prod", "staging", and "dev" are configured,
   **When** a Go program lists available gateways,
   **Then** it receives all three with their names and active status.

2. **Given** no gateways are configured,
   **When** a Go program lists available gateways,
   **Then** it receives an empty list (not an error).

---

### Edge Cases

- When the gateway config directory exists but metadata.json is missing
  or malformed, a typed config-parse-failure error is returned.
- When the token file does not exist or has wrong permissions, loading
  succeeds (tokens are lazy) but a typed token-load-failure error is
  returned on first authentication attempt.
- When `XDG_CONFIG_HOME` points to a nonexistent directory, the system
  falls through to the system gateway directory. If neither exists,
  a gateway-not-found error is returned.
- When a gateway name contains path traversal characters (e.g.,
  `../etc/passwd`), validation rejects it immediately with an error
  before any filesystem access.
- When both a user gateway and a system gateway exist with the same
  name, the user gateway takes precedence (user-first fallback).
- When the legacy `cf_token` file exists but `edge_token` does not,
  the system reads `cf_token` as the edge token (backward compatible).
- When `metadata.json` contains an unrecognized `auth_mode` value
  (not none, plaintext, cloudflare_jwt, oidc, or mtls), a typed
  unsupported-auth-mode error is returned.

## Clarifications

### Session 2026-07-02

- Q: What is the format of the active_gateway marker file? → A: A plain text file containing the gateway name as a single line (no JSON, no symlink). Located at `$XDG_CONFIG_HOME/openshell/active_gateway`.
- Q: Should exported functions (NewClient, LoadConfig, ListGateways) be safe for concurrent use? → A: Yes, all exported functions must be safe for concurrent use from multiple goroutines (standard Go library expectation).
- Q: Should callers be able to distinguish different error categories (not found, parse error, token error)? → A: Yes, the package provides typed errors so callers can use errors.Is/As to distinguish gateway-not-found, config-parse-failure, token-load-failure, and unsupported-auth-mode.
- Q: Should unknown fields in metadata.json be ignored or cause errors? → A: Ignored silently for forward compatibility with newer Rust CLI versions that may add fields.
- Q: What is the exact system gateway directory path? → A: `/etc/openshell/gateways/` on Linux/macOS. Windows support is out of scope for v1.

## Requirements

### Functional Requirements

- **FR-001**: System MUST read gateway metadata from
  `$XDG_CONFIG_HOME/openshell/gateways/<name>/metadata.json`, falling
  back to `~/.config/` when `XDG_CONFIG_HOME` is unset.

- **FR-002**: System MUST resolve the active gateway by reading
  the `active_gateway` plain text file from the config directory
  (`$XDG_CONFIG_HOME/openshell/active_gateway`), which contains
  the gateway name as a single line.

- **FR-003**: System MUST validate gateway names as single path
  components, rejecting names containing path separators, `.`, `..`,
  or other traversal patterns.

- **FR-004**: System MUST check the system gateway directory as a
  fallback when a gateway is not found in the user directory, with
  user gateways taking precedence.

- **FR-005**: System MUST load edge tokens from
  `<gateway-dir>/edge_token` as plain text, with a legacy fallback
  to `cf_token` for backward compatibility.

- **FR-006**: System MUST load OIDC token bundles from
  `<gateway-dir>/oidc_token.json` as structured JSON containing
  at minimum `access_token`, `refresh_token`, `token_type`, and
  `expiry` fields. Unknown fields are ignored for forward
  compatibility.

- **FR-007**: Tokens MUST be loaded lazily on first use, not at
  config parse time.

- **FR-008**: System MUST map the `auth_mode` field to the correct
  auth provider: unset/none to NoAuth, plaintext to NoAuth with
  insecure dial, cloudflare_jwt to edge token auth, oidc to
  refreshable token with disk-backed source.

- **FR-009**: System MUST provide a one-call convenience constructor
  that takes a gateway name (or empty for active) and returns a
  fully configured SDK client.

- **FR-010**: System MUST provide a config-only loader that returns
  parsed gateway configuration without creating a client connection.

- **FR-011**: System MUST provide a function to enumerate all
  available gateways with their names, active status, and source
  (user vs system).

- **FR-012**: System MUST support client options to override default
  behavior (custom dial options, logger, additional auth providers).

- **FR-013**: System MUST NOT expose tokens, credentials, or
  sensitive config values in error messages or string representations.

- **FR-014**: The gateway package MUST NOT introduce filesystem
  operations into the core SDK package; all filesystem access is
  isolated to the gateway package.

- **FR-015**: All exported functions (NewClient, LoadConfig,
  ListGateways) MUST be safe for concurrent use from multiple
  goroutines.

- **FR-016**: System MUST provide typed errors so callers can
  distinguish failure categories using errors.Is/As: gateway not
  found, config parse failure, token load failure, and unsupported
  auth mode.

- **FR-017**: System MUST silently ignore unknown fields in
  metadata.json for forward compatibility with newer Rust CLI
  versions. The expected fields are: `endpoint` (string, gateway
  address), `auth_mode` (string, one of none/plaintext/cloudflare_jwt/
  oidc/mtls), and `name` (string, gateway identifier).

- **FR-018**: The system gateway directory MUST be
  `/etc/openshell/gateways/` on Linux/macOS. Windows support is
  out of scope for v1.

### Key Entities

- **GatewayConfig**: Parsed representation of a gateway's on-disk
  metadata, including endpoint, auth mode, name, and source location.

- **GatewayInfo**: Lightweight summary of a gateway for listing
  purposes, including name, active status, and source (user/system).

- **ClientOption**: Functional option type for customizing client
  creation behavior (dial options, logger, auth overrides).

## Success Criteria

### Measurable Outcomes

- **SC-001**: A Go program can obtain a working gateway client in
  a single function call, reducing connection setup from 20+ lines
  to 1 line of code.

- **SC-002**: All auth modes supported by the Rust CLI's gateway
  configuration (none, plaintext, cloudflare_jwt, oidc) are
  correctly resolved by the Go convenience layer.

- **SC-003**: Gateway name validation rejects 100% of path traversal
  attempts, matching the Rust CLI's validation behavior.

- **SC-004**: Token loading is lazy: creating a client for a gateway
  with missing token files does not fail until the token is actually
  needed for authentication.

- **SC-005**: Gateway enumeration returns all configured gateways
  from both user and system directories without blocking on network
  I/O (directory scan only, no connections opened).

- **SC-006**: The gateway package works correctly with both
  `XDG_CONFIG_HOME` set to a custom path and unset (defaulting
  to `~/.config/`).

## Assumptions

- The Rust CLI's on-disk gateway config format (metadata.json structure,
  directory layout, token file naming) is treated as a stable interface.
  Changes to this format would require updates to the Go gateway package.

- mTLS certificate loading is out of scope for this specification.
  The auth mode mapping returns an error for `mtls` with guidance to
  use manual client configuration. A future spec will address mTLS.

- The `diskTokenSource` for OIDC auth reads token bundles from disk
  and uses the refresh token for renewal. It does not implement the
  full OIDC browser login flow (that is a separate future concern).

- The gateway package depends on the core SDK auth providers (NoAuth,
  RefreshableToken) and implements its own lazy edge token auth for
  cloudflare_jwt mode. No dependency on the edge auth package is
  required.

- `LoadConfig` returns a frozen snapshot of the gateway configuration
  at the time of the call. It does not re-read on subsequent access.

- The system gateway directory path follows platform conventions
  (e.g., `/etc/openshell/gateways/` on Linux/macOS).
