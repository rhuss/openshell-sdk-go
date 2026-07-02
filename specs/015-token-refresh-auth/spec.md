# Feature Specification: Composable Token Refresh with Coalesced Caching

**Feature Branch**: `015-token-refresh-auth`
**Created**: 2026-07-01
**Status**: Draft
**Input**: Add composable token refresh with single-flight caching to the Go SDK core auth layer, following the Kubernetes client-go pattern.

## User Scenarios & Testing

### User Story 1 - Automatic token refresh for SDK consumers (Priority: P1)

A developer building an application that connects to an OpenShell gateway with OIDC authentication needs their SDK client to automatically refresh expired tokens without manual intervention. They provide a `TokenSource` (from `golang.org/x/oauth2`) that knows how to obtain fresh tokens, and the SDK handles caching, expiry checking, and concurrent-safe refresh.

**Why this priority**: This is the core capability. Without it, SDK consumers with expiring tokens must implement their own refresh logic outside the SDK.

**Independent Test**: Can be tested by creating a mock `TokenSource` that returns tokens with short expiry, making multiple concurrent gRPC calls, and verifying that tokens are refreshed before expiry and only one refresh call happens under concurrent load.

**Acceptance Scenarios**:

1. **Given** a client configured with `RefreshableToken(src)` where the token source returns tokens expiring in 30 seconds, **When** the client makes an RPC call 25 seconds later (within leeway), **Then** the SDK refreshes the token before the call and attaches the new bearer token.
2. **Given** a client with a cached valid token, **When** the client makes an RPC call, **Then** the SDK uses the cached token without calling the underlying `TokenSource`.
3. **Given** 10 goroutines making concurrent RPC calls while the cached token is expired, **When** all 10 trigger a refresh simultaneously, **Then** only one call to the underlying `TokenSource` occurs and all 10 goroutines receive the same refreshed token.

---

### User Story 2 - Graceful degradation on refresh failure (Priority: P2)

When the token refresh mechanism fails (network error, IdP unavailable), the SDK returns the most recently cached token if one exists rather than immediately failing the RPC call. This gives transient failures a chance to resolve on the next call.

**Why this priority**: Production environments experience intermittent IdP outages. Failing every RPC during a brief outage degrades the user experience unnecessarily when the existing token may still be accepted by the server.

**Independent Test**: Can be tested by creating a `TokenSource` that fails after the first successful token, making an RPC call after expiry, and verifying the stale token is returned with a log warning.

**Acceptance Scenarios**:

1. **Given** a client with an expired cached token and a `TokenSource` that returns an error, **When** the client makes an RPC call, **Then** the SDK returns the stale cached token and logs a warning.
2. **Given** a client with no cached token and a `TokenSource` that returns an error, **When** the client makes an RPC call, **Then** the SDK returns the error (no fallback available).

---

### User Story 3 - Configurable refresh leeway (Priority: P3)

A developer wants to control how far in advance of token expiry the SDK triggers a proactive refresh. The default leeway is 10 seconds, but some environments need more or less buffer.

**Why this priority**: Useful for tuning, but the 10-second default (matching Kubernetes client-go) works for most cases.

**Independent Test**: Can be tested by creating a token with 60-second expiry, setting leeway to 30 seconds, and verifying refresh triggers at the 30-second mark.

**Acceptance Scenarios**:

1. **Given** a client configured with `RefreshableToken(src, WithLeeway(30*time.Second))` and a token expiring in 60 seconds, **When** the client makes a call at the 35-second mark, **Then** the SDK proactively refreshes the token.
2. **Given** a client configured with default leeway (10s) and a token expiring in 60 seconds, **When** the client makes a call at the 35-second mark, **Then** the SDK uses the cached token without refreshing.

---

### Edge Cases

- What happens when the `TokenSource` returns a token with zero/missing expiry? The SDK treats it as always-valid (never refreshes), matching `oauth2.Token.Valid()` behavior.
- What happens when `RefreshableToken` is called with a nil `TokenSource`? The constructor returns an error.
- What happens when the token has already expired by more than the leeway when the first RPC is made? The SDK calls the `TokenSource` immediately (no cached token to fall back on for the very first call).
- What happens when the `TokenSource` returns a token with an expiry in the past? The SDK calls the source again on the next RPC (the returned token is immediately stale).

## Requirements

### Functional Requirements

- **FR-001**: SDK MUST provide a `RefreshableToken(src oauth2.TokenSource, opts ...RefreshOption) (AuthProvider, error)` constructor that wraps a `TokenSource` into an `AuthProvider`.
- **FR-002**: The returned `AuthProvider` MUST cache the most recent valid token in memory and return it for `GetRequestMetadata` calls without invoking the underlying `TokenSource`.
- **FR-003**: When the cached token is expired or within the leeway window, the `AuthProvider` MUST call the underlying `TokenSource.Token()` to obtain a fresh token.
- **FR-004**: Concurrent calls to `GetRequestMetadata` MUST coalesce into a single `TokenSource.Token()` call using RWMutex double-checked locking (read-lock fast path, write-lock slow path with re-check).
- **FR-005**: When the `TokenSource` returns an error and a previously cached token exists, the `AuthProvider` MUST return the stale cached token and log a warning.
- **FR-006**: When the `TokenSource` returns an error and no cached token exists, the `AuthProvider` MUST return the error.
- **FR-007**: The `RefreshableToken` constructor MUST accept a `WithLeeway(d time.Duration)` option. Default leeway is 10 seconds. Negative values MUST be clamped to 0.
- **FR-008**: The `RefreshableToken` constructor MUST accept a `WithLogger(l types.Logger)` option. When set, the logger is used for stale-token fallback warnings (FR-005). When not set, warnings are silently dropped.
- **FR-009**: The `RefreshableToken` constructor MUST return an error if the provided `TokenSource` is nil.
- **FR-010**: The existing `AuthProvider` interface, `NoAuth()`, and `StaticToken()` implementations MUST remain unchanged.
- **FR-011**: The core SDK MUST NOT perform any filesystem operations. Token persistence and disk-based token sources are the responsibility of external packages.

### Key Entities

- **AuthProvider**: Existing interface wrapping gRPC `PerRPCCredentials`. Unchanged.
- **RefreshOption**: Functional option type for configuring the refreshable token wrapper (leeway, logger, future options).
- **refreshableAuth**: Internal struct implementing `AuthProvider` with cached token, RWMutex, underlying `TokenSource`, leeway duration, and optional logger.

## Success Criteria

### Measurable Outcomes

- **SC-001**: A test with 1000 concurrent goroutines calling `GetRequestMetadata` on an expired token results in exactly 1 call to the underlying `TokenSource`.
- **SC-002**: The caching wrapper adds less than 1 microsecond overhead on the fast path (cached valid token).
- **SC-003**: All existing SDK tests continue to pass without modification (no breaking changes).
- **SC-004**: The implementation adds no more than 150 lines of non-test Go code to the core SDK.

## Assumptions

- This feature adds `golang.org/x/oauth2` as a new runtime dependency. This is justified under Constitution principle V (Minimal Dependencies) because: (1) `oauth2.TokenSource` is the standard Go interface for token acquisition, used by every major Go SDK (Kubernetes client-go, Google Cloud, AWS); (2) re-defining equivalent types would force callers to write adapters; (3) the package is maintained by the Go team under `golang.org/x` with strong stability guarantees.
- Callers import `golang.org/x/oauth2` directly for `Token` and `TokenSource` types. The SDK does not re-export or vendor these types.
- The `oauth2.Token.Valid()` method is the authority on whether a token is expired. The SDK applies leeway by checking `token.Expiry.Add(-leeway).Before(time.Now())`.
- Logging uses the SDK's existing `Logger` interface, injected via `WithLogger(l types.Logger)`. Since `RefreshableToken` is a standalone constructor (not a `Client` method), the logger cannot be inherited from `types.Config`. If no logger is configured, warnings on stale-token fallback are silently dropped.
- 401-triggered token reset (as in k8s client-go's `ResetTokenOlderThan`) is out of scope. gRPC uses status codes differently from HTTP, and this can be added later if needed.
- The `RequireTransportSecurity()` method on the returned `AuthProvider` returns `true`, matching existing `StaticToken` behavior. Bearer tokens must not be sent over plaintext channels.
