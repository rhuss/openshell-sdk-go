# Brainstorm: Edge Auth (Extra Headers + WebSocket Tunnel)

**Date:** 2026-07-02
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/20

## Problem Framing

When a gateway sits behind a zero-trust reverse proxy (Cloudflare Access,
Google IAP, Zscaler, etc.), the client must send two independent sets of
credentials on every RPC:

1. **Edge auth**: headers that the reverse proxy validates before forwarding
   the request to the origin. Each vendor uses different header names and
   formats.
2. **Application auth**: the standard `authorization: Bearer` header that
   the gateway itself validates. Already covered by `StaticToken` and
   `RefreshableToken`.

The Go SDK currently has no mechanism to attach extra per-RPC headers
alongside the existing `AuthProvider`. The Rust CLI handles this via
`EdgeAuthInterceptor` (dual-mode: OIDC bearer or CF Access headers) and
`EdgeTunnelProxy` (WebSocket tunnel for gRPC behind edge proxies that
reject HTTP/2 POST).

**Scope:** This brainstorm covers the full edge story: generic header
layering in the core SDK plus a Cloudflare Access convenience constructor
and WebSocket tunnel proxy in a separate package. Implementation is
phased (headers first, tunnel second).

## Approaches Considered

### A: Cloudflare-specific AuthProvider

A `CloudflareAccessToken(token string) AuthProvider` that sends
`cf-access-jwt-assertion` + `cookie` headers. Simple and direct.

- Pros: minimal, does one thing
- Cons: not reusable for other edge proxies, embeds vendor name in the
  core SDK

### B: Single EdgeToken constructor (Rust mirror)

`EdgeToken(oidcToken, edgeToken string) AuthProvider` mirrors the Rust
`EdgeAuthInterceptor` with built-in precedence logic.

- Pros: 1:1 match with Rust, single entry point
- Cons: mixes two auth mechanisms, OIDC path duplicates `StaticToken`,
  un-Go-ish

### C: Generic WithExtraHeaders wrapper (chosen)

A decorator that wraps any `AuthProvider` and adds extra static headers:

```go
// Core SDK (openshell/v1/)
WithExtraHeaders(base AuthProvider, headers map[string]string) AuthProvider
```

Vendor-specific convenience constructors live in a separate package:

```go
// openshell/v1/edge/
edge.CloudflareAccess(baseAuth AuthProvider, edgeToken string) AuthProvider
```

- Pros: generic, composable, no new deps in core, works with any edge
  proxy, CF-specific logic is optional
- Cons: static headers only (edge token can't refresh independently)

### D: WithHeaderFunc callback

Like C but per-RPC callback: `WithHeaderFunc(base, fn func() (map[string]string, error))`.

- Pros: supports dynamic/rotating edge tokens
- Cons: callback indirection, harder to test, YAGNI for static edge
  tokens

### E: Middleware chain

`ChainAuth(providers ...AuthProvider)` merges headers from multiple
providers.

- Pros: maximum composability
- Cons: `RequireTransportSecurity()` semantics get awkward (AND? OR?),
  over-designed

## Decision

**Option C: Generic `WithExtraHeaders` wrapper in core SDK, with
Cloudflare Access convenience constructor and WebSocket tunnel proxy in
a separate `openshell/v1/edge/` package.**

The generic mechanism keeps the core SDK vendor-neutral. Edge-specific
logic (CF Access headers, WS tunnel) is isolated in an optional package
that consumers only import when needed, keeping the WebSocket dependency
out of the core module.

If dynamic edge tokens are needed later, `WithHeaderFunc` can be added
without breaking anything.

## Key Requirements

### Phase 1: Core SDK (`openshell/v1/`)

- `WithExtraHeaders(base AuthProvider, headers map[string]string) AuthProvider`
- Merges extra headers with the base provider's headers per RPC
- On key collision, extra headers take precedence over base headers
- `RequireTransportSecurity()` delegates to the base provider
- Works with `NoAuth`, `StaticToken`, `RefreshableToken`, and any future
  `AuthProvider`
- No new dependencies

### Phase 2: Edge package (`openshell/v1/edge/`)

- `CloudflareAccess(baseAuth AuthProvider, edgeToken string) AuthProvider`
  convenience constructor: formats `cf-access-jwt-assertion` and
  `cookie: CF_Authorization=<token>` headers, delegates to
  `WithExtraHeaders`
- `NewTunnelProxy(gatewayURL, edgeToken string, opts ...TunnelOption) (*TunnelProxy, error)`
  WebSocket tunnel proxy for gRPC behind edge proxies that reject HTTP/2
  POST
- Tunnel takes its own edge token parameter (separate from AuthProvider)
- Explicit `Close()` for tunnel lifecycle, goroutine cleanup
- WebSocket dependency isolated to this package

### Fake support

- `WithExtraHeaders` wrapping a fake auth is testable without changes to
  the fake package (it's just a decorator)
- Tunnel proxy: not faked (transport concern, not API concern)

## Open Questions

- WithExtraHeaders: should empty-string values be silently skipped or
  sent as empty headers?
- Tunnel proxy: connection pool or goroutine-per-connection? (Rust uses
  goroutine-per-connection)
- Tunnel proxy: should TunnelOption include WithTLS for the WebSocket
  connection itself (wss:// vs ws://)?
- Tunnel proxy: error logging strategy (use types.Logger like
  RefreshableToken, or a separate mechanism?)
- Should edge.CloudflareAccess validate that the edgeToken is non-empty,
  or leave validation to the server?
- Tunnel proxy: should Close() drain in-flight connections or force-close?

## Future Brainstorms

- Dynamic edge token refresh (WithHeaderFunc or edge-specific
  TokenSource) if edge tokens need independent refresh cycles
- Other edge proxy convenience constructors (Google IAP, Zscaler) when
  concrete use cases arise
- CLI convenience layer: gateway config loading, browser-based CF Access
  login, disk-based edge token storage
