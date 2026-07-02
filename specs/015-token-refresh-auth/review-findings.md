# Deep Review Findings

**Date:** 2026-07-01
**Branch:** 015-token-refresh-auth
**Rounds:** 0
**Gate Outcome:** PASS
**Invocation:** superpowers

## Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 0 | 0 | 0 |
| Minor | 3 | - | 3 |
| Notable | 1 | - | 1 |
| **Total** | **4** | **0** | **4** |

**Agents completed:** 5/5 (+ 1 external tool)
**Agents failed:** 0

## Agent Summary

| Agent | Findings | Critical | Important | Minor | Notable |
|-------|----------|----------|-----------|-------|---------|
| Correctness | 1 | 0 | 0 | 1 | 0 |
| Architecture & Idioms | 0 | 0 | 0 | 0 | 0 |
| Security | 0 | 0 | 0 | 0 | 0 |
| Production Readiness | 1 | 0 | 0 | 0 | 1 |
| Test Quality | 0 | 0 | 0 | 0 | 0 |
| CodeRabbit (external) | 2 | 0 | 0 | 2 | 0 |

## Findings

### FINDING-1
- **Severity:** Minor
- **Confidence:** 60
- **File:** openshell/v1/auth_refresh.go:69
- **Category:** correctness
- **Source:** correctness-agent
- **Round found:** 1
- **Resolution:** deferred (by design)

**What is wrong:**
The `context.Context` parameter in `GetRequestMetadata` is accepted but not
forwarded to `r.source.Token()`. The underlying `oauth2.TokenSource` interface
does not accept a context, so the parameter cannot be propagated.

**Why this matters:**
Context cancellation cannot abort an in-flight token refresh. However, the
`oauth2.TokenSource` interface (`Token() (*Token, error)`) has no context
parameter, making propagation impossible without wrapping the source. The spec
(FR-003, FR-004) does not require context propagation.

**Recommendation:**
No action required. This is an inherent limitation of the `oauth2.TokenSource`
interface. A future enhancement could accept a context-aware token source, but
that is outside the current spec scope.

---

### FINDING-2
- **Severity:** Notable
- **Confidence:** 70
- **File:** openshell/v1/auth_refresh.go:80
- **Category:** production-readiness
- **Source:** production-readiness-agent
- **Round found:** 1
- **Resolution:** deferred (by design)

**What is wrong:**
When a token refresh is needed, all concurrent callers block on `r.mu.Lock()`
while one goroutine performs the refresh. If the `TokenSource.Token()` call is
slow (e.g., network timeout), all callers are blocked.

**Why this matters:**
This is inherent to the RWMutex double-checked locking pattern specified in the
design (modeled after Kubernetes client-go `cachingTokenSource`). The pattern
trades potential blocking for simplicity and correctness. The alternative
(`singleflight.Group`) adds complexity without meaningful benefit for this use
case since callers must wait for the token regardless.

**Recommendation:**
No action required. The blocking behavior is the expected tradeoff of the
specified pattern. Document this in operational notes if the SDK gains an SLA
on `GetRequestMetadata` latency.

---

### FINDING-3
- **Severity:** Minor
- **Confidence:** 75
- **File:** openshell/v1/auth_refresh.go:59-66
- **Category:** correctness
- **Source:** coderabbit
- **Round found:** 1
- **Resolution:** deferred (defensive improvement)

**What is wrong:**
`isTokenValid()` returns `true` for tokens with a zero expiry even if
`AccessToken` is empty. Similarly, the refresh path caches `newTok` without
verifying that `AccessToken` is non-empty.

**Why this matters:**
The `oauth2.TokenSource` contract guarantees that `Token()` returns a valid
token with a populated `AccessToken` on success. An empty `AccessToken` would
indicate a broken `TokenSource` implementation, which is the caller's
responsibility. The spec does not require empty-token validation.

**Recommendation:**
Consider adding an `AccessToken != ""` check as a defensive measure in a
future iteration. Not required for spec compliance.

---

### FINDING-4
- **Severity:** Minor
- **Confidence:** 75
- **File:** openshell/v1/auth_refresh.go:35-40
- **Category:** correctness
- **Source:** coderabbit
- **Round found:** 1
- **Resolution:** deferred (defensive improvement)

**What is wrong:**
`WithLeeway` accepts negative durations without validation. A negative leeway
would cause `Expiry.Add(-leeway)` to shift forward in time, making tokens
appear valid longer than their actual expiry.

**Why this matters:**
While a negative leeway is a caller error (not a security vulnerability since
the token still works), it produces counterintuitive behavior. The spec
(FR-009) defines leeway as "duration before token expiry" which implies a
non-negative value, but does not explicitly require validation.

**Recommendation:**
Consider clamping negative leeway values to zero in a future iteration. Not
required for spec compliance.
