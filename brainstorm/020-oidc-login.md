# Brainstorm: OIDC Login Package

**Date:** 2026-07-03
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/24

## Problem Framing

The Go SDK can read OIDC tokens from disk (`gateway.NewClient` with
`diskTokenSource`) and refresh them via `RefreshableToken`. But when
tokens expire and no valid refresh token exists, there's no way to
obtain fresh tokens from Go. The user must fall back to the Rust CLI
(`openshell gateway login`) or manually acquire tokens.

This gap blocks Go-based CLI tools, operators, and automation from
being fully self-contained. A Go program should be able to
authenticate a user (or service account) against the gateway's OIDC
provider without depending on the Rust CLI.

The Kubernetes ecosystem solved this with
[kubelogin](https://github.com/int128/kubelogin), a kubectl plugin
that handles the OIDC browser flow and caches tokens to disk. The
OpenShell Go SDK needs an equivalent.

## Approaches Considered

### A: Single `openshell/v1/oidc/` package (chosen)

One cohesive package handling OIDC discovery, all grant types, gateway
integration, and disk persistence.

```go
// Gateway-aware: discovers OIDC provider from gateway metadata
token, err := oidc.Login("my-gateway")

// Standalone: explicit provider
token, err := oidc.Login("",
    oidc.WithIssuer("https://auth.example.com"),
    oidc.WithClientID("my-client"),
)

// Client credentials (service accounts)
token, err := oidc.ClientCredentials(
    oidc.WithIssuer("https://auth.example.com"),
    oidc.WithClientID("my-service"),
    oidc.WithClientSecret(secret),
)

// Device code (headless/CI)
token, err := oidc.DeviceLogin(
    oidc.WithIssuer("https://auth.example.com"),
    oidc.WithClientID("my-device"),
)
```

- Pros: single import, cohesive API, gateway metadata drives the login
  flow automatically, matches `openshell gateway login` UX
- Cons: larger package, mixes interactive and non-interactive flows

### B: Split into `oidc/` (core) + `oidc/gateway/` (integration)

Core package for pure OIDC flows, separate sub-package for gateway
awareness and disk persistence.

- Pros: clean separation, core usable without gateway dependency
- Cons: two imports for the common case, extra wiring

### C: Extend `gateway/` package directly

Add `gateway.Login(name)` to the existing gateway package.

- Pros: simplest API for the common case
- Cons: bloats gateway with HTTP server, browser logic, OIDC protocol.
  Can't use OIDC flows outside gateway context.

## Decision

**Option A: Single `openshell/v1/oidc/` package.** The common case is
`oidc.Login("my-gateway")` and it should be one import. Standalone
flows (client_credentials, device code) share OIDC discovery and token
handling logic, so they belong together.

## Key Requirements

### OIDC Discovery

- Fetch `.well-known/openid-configuration` from the issuer URL
- Resolve authorization_endpoint, token_endpoint, device_authorization_endpoint
- Cache discovery documents (they rarely change)

### Auth Code + PKCE (interactive browser flow)

- Start a localhost HTTP callback server
- Port selection: try fixed ports [8000, 18000] by default, configurable
  via `WithCallbackPort(port)` for providers requiring pre-registered
  redirect URIs
- Generate PKCE code verifier + challenge (S256)
- Auto-open system browser (xdg-open/open/cmd start)
- Receive callback, exchange code for tokens
- Timeout after configurable duration (default: 2 minutes)

### Authcode-Keyboard (headless fallback)

- Following kubelogin's pattern: display the authorization URL, prompt
  the user to open it manually and paste the authorization code
- No localhost server needed
- Activated automatically when browser open fails, or via
  `WithKeyboardFlow()` option

### Device Code Flow (RFC 8628)

- For input-constrained devices and CI runners
- Display device code + verification URL
- Poll token endpoint at the provider's specified interval
- Accept a display callback (`WithDisplayFunc`) for custom UX

### Client Credentials Grant

- For service accounts and CI/CD pipelines
- `WithClientID` + `WithClientSecret`
- No user interaction required
- Returns `oauth2.Token` directly

### Gateway-Aware Login

- `Login(gatewayName)` reads gateway metadata to discover the OIDC
  provider (issuer URL, client ID from metadata.json or a new
  `oidc_config` field)
- After successful login, writes the token bundle to
  `<gateway-dir>/oidc_token.json`
- Empty gateway name resolves via active gateway (same as
  `gateway.NewClient("")`)
- Matches `openshell gateway login` behavior

### Token Persistence

- Default: write to gateway's `oidc_token.json` (interop with Rust CLI
  and `gateway.NewClient`)
- Option: `WithInMemory()` to skip disk writes for embedded scenarios
- Token format matches existing `oidc_token.json` schema
  (`access_token`, `refresh_token`, `token_type`, `expiry`)

### Integration with Existing SDK

- Returns `oauth2.TokenSource` compatible with `v1.RefreshableToken`
- Works alongside `gateway.NewClient` (login writes tokens,
  NewClient reads them)
- Follows SDK conventions: functional options, typed errors, secrets
  never leak, doc comments on all exports

## Open Questions

- Where does the OIDC client_id come from for gateway-aware login?
  The current `metadata.json` schema has `endpoint`, `auth_mode`,
  `name` but no OIDC-specific fields. Options: add `oidc_client_id`
  to metadata.json, use a separate `oidc_config.json`, or derive from
  the gateway name.
- Should the package support custom scopes beyond the OpenShell
  default, or hardcode the required scopes?
- Should token refresh (using the refresh_token) be handled by this
  package or left to `RefreshableToken`? The current design has
  `diskTokenSource` re-reading from disk, and `RefreshableToken`
  handles the oauth2 refresh cycle.
- How should the device code flow display the verification URL and
  code? A callback function, stdout, or a channel?
- Should the package validate the ID token (signature verification,
  audience check) or trust the token endpoint response?

## Future Brainstorms

- Token revocation endpoint support
- OIDC provider auto-configuration from gateway admin API
- Multi-tenant OIDC (different providers per gateway)
