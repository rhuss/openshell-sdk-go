# Research: Token Refresh with Coalesced Caching

## R1: Concurrency pattern for token caching

**Decision**: RWMutex double-checked locking (k8s client-go pattern)

**Rationale**: The `cachingTokenSource` in [k8s client-go transport/token_source.go](https://github.com/kubernetes/client-go/blob/master/transport/token_source.go) uses this exact pattern. Fast path reads the cached token under RLock (no contention). Slow path acquires a full Lock, re-checks (another goroutine may have refreshed while waiting), then calls the underlying source. This is simpler and more appropriate than `singleflight.Group` because we need to cache the result across calls, not just deduplicate in-flight calls.

**Alternatives considered**:
- `golang.org/x/sync/singleflight`: Deduplicates concurrent calls but doesn't cache the result. Every call after the flight completes would trigger a new flight. Would need additional caching logic on top, making it more complex than RWMutex.
- `sync.Once`: Too rigid. Once the token is fetched, it never refreshes. Not suitable for expiring tokens.
- Channel-based: More complex, harder to reason about, no advantage over RWMutex for this use case.

## R2: Token expiry check with leeway

**Decision**: Check `token.Expiry.Add(-leeway).Before(time.Now())`. Default leeway 10 seconds.

**Rationale**: Matches k8s client-go's approach. The leeway ensures tokens are refreshed slightly before actual expiry, preventing RPCs from hitting the server with tokens that expire during transit. 10 seconds is the k8s default and provides a reasonable buffer.

**Alternatives considered**:
- No leeway (refresh only on actual expiry): Risk of sending expired tokens due to clock skew or network latency.
- oauth2.Token.Valid() only: This method uses a 10-second expiry window internally, but it's not configurable. Using our own check allows the WithLeeway option.

## R3: Graceful degradation on refresh failure

**Decision**: Return stale cached token on refresh error, log warning. Error only if no cached token exists.

**Rationale**: k8s client-go does exactly this in `cachingTokenSource.Token()`: on error, if a previous token exists, it returns the stale token and logs via klog. This handles transient IdP outages gracefully. The server may still accept a recently-expired token, and even if it doesn't, the gRPC error from the server is more informative than a local "token refresh failed" error.

**Alternatives considered**:
- Always error on refresh failure: Too aggressive. Brief IdP blips would break all RPCs even when the cached token might still work.
- Retry with backoff: Adds complexity (goroutine management, timer). Better suited to a higher-level layer. The core caching wrapper should be simple.

## R4: golang.org/x/oauth2 dependency justification

**Decision**: Add `golang.org/x/oauth2` as a direct dependency.

**Rationale**: The `oauth2.TokenSource` interface is the de facto standard in Go for token management. It's used by k8s client-go, google-cloud-go, aws-sdk-go-v2, and every major Go SDK that handles OAuth2 tokens. The package is maintained by the Go team under the golang.org/x namespace. It provides `Token` and `TokenSource` types that our users already work with. Defining a custom interface would force users to write adapters.

**Alternatives considered**:
- Custom `TokenSource` interface in the SDK: Would duplicate the oauth2 interface and force adapter code. Users already have oauth2.TokenSource implementations from their IdP libraries.
- Accept `func() (string, time.Time, error)` instead: Simpler but loses type safety and interoperability with the oauth2 ecosystem.

## R5: RequireTransportSecurity behavior

**Decision**: `RequireTransportSecurity()` returns `true` on the refreshable auth provider.

**Rationale**: Bearer tokens should not be sent over plaintext connections. The existing `StaticToken` implementation returns `true`. The refreshable auth should match this behavior for security consistency. Users who need plaintext (local dev) can use `NoAuth()` or configure gRPC's `grpc.WithInsecure()` which bypasses this check.

**Alternatives considered**:
- Return `false` to allow plaintext: Security risk. Bearer tokens sent in the clear can be intercepted.
- Make it configurable: Unnecessary complexity. The gRPC framework already has mechanisms to override transport security requirements.
