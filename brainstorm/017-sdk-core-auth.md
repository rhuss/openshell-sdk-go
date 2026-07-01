# Brainstorm: SDK core auth (token refresh + single-flight)

**Date:** 2026-07-01
**Status:** active

## Problem Framing

The Go SDK currently supports static bearer tokens and TLS (plaintext, CA-only,
mTLS, insecure) via the `AuthProvider` interface wrapping gRPC's
`PerRPCCredentials`. The Rust CLI supports OIDC token refresh with single-flight
coalescing, disk-based token re-read, OIDC discovery, and the Cloudflare Access
tunnel. The gap matters both for practical SDK use and for the RFC-0008 discussion
where we're arguing that native SDKs can cover the transport/auth layer without a
shared Rust core.

This brainstorm scopes only the **core SDK** auth layer. CLI convenience
(gateway config loading, disk-aware token re-read) and full OIDC browser flow
are deferred to separate brainstorms.

## Approaches Considered

### A: Extend AuthProvider interface

Add `Refresh()` and `IsExpired()` methods to the existing `AuthProvider`
interface. Implementations that support refresh implement them; others return
no-ops.

- Pros: single interface, no new types
- Cons: breaks the small-interface Go convention, forces all implementations to
  carry refresh stubs, not composable

### B: Separate composable TokenRefresher wrapper

Keep `AuthProvider` as-is. Add a `RefreshableToken(tokenSource)` constructor
that wraps a `TokenSource` into an `AuthProvider` with built-in single-flight
caching.

- Pros: composable, follows k8s client-go pattern, existing AuthProvider
  implementations unchanged
- Cons: two concepts (AuthProvider + TokenSource) instead of one

### C: Use Go's standard oauth2.TokenSource interface

Wrap `golang.org/x/oauth2.TokenSource` with single-flight caching inside the
SDK, expose the result as an `AuthProvider`. This is option B but using the
standard ecosystem interface rather than inventing a custom one.

- Pros: most idiomatic Go, familiar to anyone who's used oauth2, composable,
  matches k8s client-go's exact pattern
- Cons: adds `golang.org/x/oauth2` dependency to the core module

## Decision

**Option C: Use `oauth2.TokenSource` with composable single-flight caching.**

Matches the Kubernetes client-go design exactly:

- `cachingTokenSource` wraps any `TokenSource` with RWMutex double-checked
  locking (fast path: read cached token, slow path: single-flight refresh)
- Configurable leeway (refresh N seconds before actual expiry)
- Core wrapper is memory-only, no filesystem access
- Disk re-read and OIDC refresh are separate `TokenSource` implementations
  that plug into the caching wrapper (deferred to CLI convenience layer)

Estimated effort: ~50-80 lines for the caching wrapper + constructor, plus
tests.

## Key Requirements

- `RefreshableToken(src oauth2.TokenSource, opts ...RefreshOption) AuthProvider`
  constructor in the core SDK
- Single-flight coalescing via RWMutex (k8s client-go pattern, not
  `singleflight` package, since we need to cache the result)
- Configurable leeway via `WithLeeway(d time.Duration)` option
- Graceful degradation: on refresh failure, return stale token if available
  (matches k8s client-go behavior)
- Core SDK stays filesystem-free; `TokenSource` implementations that read
  from disk belong in a separate package

## Open Questions

- Should we vendor or re-export `oauth2.Token` from the SDK's types package,
  or let callers import `golang.org/x/oauth2` directly?
- Exact leeway default (k8s uses 10s, Python SDK uses expiry-based check with
  no explicit leeway)
- Whether to support 401-triggered token reset (k8s client-go does this via
  `ResetTokenOlderThan`; may be overkill for gRPC where status codes differ)

## Future Brainstorms

- **CLI convenience layer**: gateway config loading from
  `~/.config/openshell/gateways/`, disk-aware `fileTokenSource`,
  `FromGatewayConfig(name)` constructor
- **Full OIDC browser flow**: OIDC discovery, authorization code + PKCE,
  localhost callback server, client-credentials flow, as an optional
  `openshell/v1/oidc` package
