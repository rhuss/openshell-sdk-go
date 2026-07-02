# Brainstorm: CLI Auth Convenience Layer

**Date:** 2026-07-02
**Status:** implemented
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/21

## Problem Framing

The Go SDK has a solid auth foundation: `NoAuth`, `StaticToken`, and
`RefreshableToken` (spec 015) in the core, with `WithExtraHeaders` and
`edge.CloudflareAccess` brainstormed (018) for edge proxies. But a Go
program that wants to connect to a gateway configured via
`openshell gateway add` must manually:

1. Locate the gateway's metadata file on disk
2. Parse the JSON metadata to determine endpoint and auth mode
3. Load the right token from disk (edge token, OIDC token bundle)
4. Construct the matching `AuthProvider` chain
5. Set up TLS/dial options based on the gateway's configuration

The Rust CLI handles all of this automatically via the `bootstrap` crate.
Both brainstorms 017 (sdk-core-auth) and 018 (edge-auth) explicitly
deferred this as "CLI convenience layer" for a separate brainstorm.

**Who needs this:** Go-based operators, automation tools, custom CLIs,
and integration tests that talk to gateways configured by the primary
Rust CLI.

## Approaches Considered

### A: Gateway config reader + convenience constructor (chosen)

A new `openshell/v1/gateway/` package that reads the Rust CLI's on-disk
gateway configs and constructs the right `AuthProvider` + dial options
automatically:

```go
client, err := gateway.NewClient("my-gateway")       // named gateway
client, err := gateway.NewClient("")                  // active gateway
cfg, err := gateway.LoadConfig("my-gateway")          // manual wiring
```

- Pros: high value, focused scope, immediate interop with Rust CLI,
  read-only (Rust CLI stays the primary gateway manager)
- Cons: coupled to Rust CLI's on-disk format, breaks if Rust changes
  the layout

### B: Full gateway management package

Like A but adds `Add()`, `Remove()`, `List()`, `SetActive()`, `Login()`.
Mirrors the Rust CLI's `gateway *` subcommands programmatically.

- Pros: complete, Go programs don't need the Rust CLI
- Cons: much larger scope, duplicates Rust CLI functionality, needs OIDC
  login flow (separate concern), YAGNI

### C: Abstract config loader

A `FromConfig(reader io.Reader)` that parses gateway metadata JSON from
any source, not tied to disk paths.

- Pros: flexible, testable, no filesystem assumptions
- Cons: loses the convenience, caller still locates files manually,
  doesn't solve the "give me a client for my-gateway" problem

## Decision

**Option A: Gateway config reader with convenience constructors in
`openshell/v1/gateway/` package.** Read-only interop with the Rust CLI's
on-disk gateway configs.

The core value is removing boilerplate: a Go program says "connect to
gateway X" and gets a working client. This mirrors how kubectl uses
kubeconfig, how docker reads config.json.

If abstract loading is needed later (embedded configs, env vars),
a `FromConfig(reader)` variant can be added without breaking anything.

## Key Requirements

### Config reading

- Read gateway metadata from
  `$XDG_CONFIG_HOME/openshell/gateways/<name>/metadata.json`
- Resolve active gateway from active_gateway marker file
- Validate gateway names (single path component, no traversal, matching
  the Rust `validated_gateway_name` behavior)
- XDG Base Directory compliant (fallback to `~/.config/` when
  `$XDG_CONFIG_HOME` is unset)
- Also check system gateway directory as fallback (installer-provided
  configs)

### Token loading

- Load edge tokens from `<gateway-dir>/edge_token` (plain text)
- Load OIDC token bundles from `<gateway-dir>/oidc_token.json` (JSON)
- Legacy fallback: `cf_token` to `edge_token` (matching Rust migration)
- Respect 0600 file permissions (read existing, don't modify)
- Tokens are loaded lazily on first use, not at config parse time

### Auth mode resolution

Map `auth_mode` field to the right AuthProvider chain:

| auth_mode | AuthProvider | Notes |
|-----------|-------------|-------|
| (unset/none) | `NoAuth()` | Local/trusted gateway |
| `plaintext` | `NoAuth()` | Plus insecure dial option |
| `mtls` | mTLS credentials | Separate concern, see open questions |
| `cloudflare_jwt` | `edge.CloudflareAccess(base, edgeToken)` | Depends on edge package (018) |
| `oidc` | `RefreshableToken(diskTokenSource)` | Disk-backed token source with OIDC refresh token |

### Convenience constructors

- `NewClient(name string, opts ...ClientOption) (*v1.Client, error)`:
  one-call gateway resolution, auth setup, and client creation
- `LoadConfig(name string) (*GatewayConfig, error)`: parse config without
  connecting, for manual wiring or inspection
- `ListGateways() ([]GatewayInfo, error)`: enumerate available gateways
  (for CLI tab-completion or UI)
- Empty name resolves to active gateway

### Package placement

- `openshell/v1/gateway/` as a separate package (not in core)
- Depends on core SDK auth providers + edge package
- Filesystem operations isolated to this package (core SDK stays
  filesystem-free per constitution)

## Resolved Questions

- **TLS scope:** `NewClient` resolves both auth and TLS (insecure, custom
  CA, mTLS certs) from gateway metadata. One call gives a fully configured
  client, matching the Rust bootstrap behavior. mTLS cert/key/CA loading
  is part of this package.
- **OIDC disk token source:** Full `oauth2.TokenSource` implementation.
  Reads token bundle from disk, uses refresh token to cycle access tokens,
  writes updated bundle back to disk. This makes OIDC gateways work for
  unattended workflows without re-login.
- **Active gateway:** Supported. Empty name resolves to the `active_gateway`
  marker file, matching the Rust CLI's `gateway use` behavior.
- **LoadConfig:** Returns a frozen snapshot (parsed once, returned as a
  value). Caller can re-call `LoadConfig` for fresh data. No live view,
  no TOCTOU issues.
- **Edge dependency:** Direct import of `openshell/v1/edge`. Edge auth is
  already shipped (PR #22), no indirection needed.
- **Gateway precedence:** User gateways override system gateways (user-first
  fallback to system), matching Rust behavior.
- **ListGateways source info:** Includes source (user vs system) per entry,
  matching the Rust `GatewayMetadataSource` pattern. Useful for CLI
  tab-completion and debugging.

## Open Questions

(none remaining)

## Future Brainstorms

- Full OIDC browser flow as `openshell/v1/oidc/` package (auth code +
  PKCE, client credentials, discovery, localhost callback server)
- Gateway management operations (Add, Remove, SetActive) if Go programs
  need to manage gateways without the Rust CLI
- Multi-gateway client support (connecting to multiple gateways from
  one process)
