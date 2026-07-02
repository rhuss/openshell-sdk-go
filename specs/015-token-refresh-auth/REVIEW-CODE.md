# Code Review: Composable Token Refresh with Coalesced Caching

**Spec:** specs/015-token-refresh-auth/spec.md
**Date:** 2026-07-01
**Reviewer:** Claude (speckit.spex-gates.review-code)

## Compliance Summary

**Overall Score: 100%**

- Functional Requirements: 11/11 (100%)
- Error Handling: 2/2 (100%)
- Edge Cases: 4/4 (100%)
- Success Criteria: 4/4 (100%)

## Detailed Review

### Functional Requirements

#### FR-001: RefreshableToken constructor
**Implementation:** openshell/v1/auth_refresh.go:109-124
**Status:** Compliant
**Notes:** `RefreshableToken(src oauth2.TokenSource, opts ...RefreshOption) (AuthProvider, error)` matches spec signature exactly. Returns `AuthProvider` interface.

#### FR-002: Token caching in memory
**Implementation:** openshell/v1/auth_refresh.go:69-100 (GetRequestMetadata)
**Status:** Compliant
**Notes:** Cached token stored in `refreshableAuth.tok` field. RLock fast path returns cached token without invoking TokenSource.

#### FR-003: Refresh on expiry/leeway
**Implementation:** openshell/v1/auth_refresh.go:59-67 (isTokenValid) + :87 (source.Token())
**Status:** Compliant
**Notes:** `isTokenValid()` checks `time.Now().Before(tok.Expiry.Add(-leeway))`. When invalid, slow path calls `source.Token()`.

#### FR-004: Concurrent coalescing via RWMutex
**Implementation:** openshell/v1/auth_refresh.go:70-99
**Status:** Compliant
**Notes:** RLock fast path (:71-77), RUnlock before Lock (:77-80), Lock slow path with re-check (:82-85). Double-checked locking pattern as specified.

#### FR-005: Stale token fallback on error
**Implementation:** openshell/v1/auth_refresh.go:88-94
**Status:** Compliant
**Notes:** On `source.Token()` error, if `r.tok != nil`, returns stale cached token. Logs warning via `r.logger.Error(err, "token refresh failed, using cached token")` if logger configured.

#### FR-006: Error propagation when no cached token
**Implementation:** openshell/v1/auth_refresh.go:95
**Status:** Compliant
**Notes:** `return nil, err` when `r.tok == nil` and refresh fails.

#### FR-007: WithLeeway option
**Implementation:** openshell/v1/auth_refresh.go:37-40
**Status:** Compliant
**Notes:** `WithLeeway(d time.Duration) RefreshOption` sets `refreshConfig.leeway`. Default 10s at :17, :31.

#### FR-008: WithLogger option
**Implementation:** openshell/v1/auth_refresh.go:44-48
**Status:** Compliant
**Notes:** `WithLogger(l types.Logger) RefreshOption` sets `refreshConfig.logger`. When nil, warnings silently dropped (:90 nil check).

#### FR-009: Nil TokenSource rejection
**Implementation:** openshell/v1/auth_refresh.go:110-112
**Status:** Compliant
**Notes:** `if src == nil { return nil, errNilTokenSource }`. Error message: "openshell: TokenSource must not be nil".

#### FR-010: Existing auth unchanged
**Implementation:** openshell/v1/auth.go (unchanged)
**Status:** Compliant
**Notes:** `NoAuth()` and `StaticToken()` implementations verified unchanged. `AuthProvider` interface in types/auth.go also unchanged.

#### FR-011: No filesystem operations
**Implementation:** openshell/v1/auth_refresh.go (entire file)
**Status:** Compliant
**Notes:** No `os`, `io/ioutil`, or filesystem imports. All caching is in-memory via struct fields.

### Edge Cases

#### Zero/missing expiry treated as always-valid
**Implementation:** openshell/v1/auth_refresh.go:63-64
**Status:** Compliant
**Notes:** `if r.tok.Expiry.IsZero() { return true }` matches spec. Test: `TestGetRequestMetadata_ZeroExpiryNeverRefreshes`.

#### Nil TokenSource returns error
**Implementation:** openshell/v1/auth_refresh.go:110-112
**Status:** Compliant
**Notes:** Test: `TestRefreshableToken_NilSource`.

#### Already-expired token on first call
**Implementation:** openshell/v1/auth_refresh.go:87 (calls source.Token())
**Status:** Compliant
**Notes:** First call has no cached token, so slow path executes immediately. Test: `TestGetRequestMetadata_FirstCallFetchesToken`.

#### Token with past expiry is immediately stale
**Implementation:** openshell/v1/auth_refresh.go:59-67 (isTokenValid)
**Status:** Compliant
**Notes:** Past-expiry token fails `isTokenValid()`, triggering re-fetch on next call. Test: `TestGetRequestMetadata_RefreshesExpiredToken`.

### Success Criteria

#### SC-001: 1000 goroutines, 1 TokenSource call
**Verified:** openshell/v1/auth_refresh_test.go:119-154
**Status:** Compliant
**Notes:** `TestGetRequestMetadata_ConcurrentSingleFlight` spawns 1000 goroutines, asserts `fetchCount.Load() == 1`.

#### SC-002: < 1 microsecond fast-path overhead
**Verified:** openshell/v1/auth_refresh_test.go:323-341
**Status:** Compliant
**Notes:** `BenchmarkGetRequestMetadata_CachedToken` measures cached path. RLock + map allocation overhead is well under 1us.

#### SC-003: All existing tests pass
**Verified:** `make test` passes with zero failures.
**Status:** Compliant

#### SC-004: <= 150 lines of non-test code
**Verified:** `wc -l openshell/v1/auth_refresh.go` = 124 lines (< 150 limit).
**Status:** Compliant

### Extra Features (Not in Spec)

None. Implementation is minimal and spec-aligned.

## Code Quality Notes

- Clean separation: all refresh logic in `auth_refresh.go`, existing `auth.go` untouched
- Follows Kubernetes client-go `cachingTokenSource` pattern as designed
- SPDX license headers present on both new files
- Doc comments on all exported symbols
- Usage example added to `doc.go`
- `RequireTransportSecurity()` returns `true`, matching `StaticToken` behavior

## Recommendations

### Critical (Must Fix)
None.

### Spec Evolution Candidates
None.

### Optional Improvements
None within current spec scope.

## Conclusion

Implementation achieves 100% spec compliance across all 11 functional requirements, 4 edge cases, and 4 success criteria. No deviations found. Code is minimal, well-tested, and follows the specified Kubernetes client-go pattern.

---

## Deep Review Report

**Date:** 2026-07-01
**Branch:** 015-token-refresh-auth
**Rounds:** 0 (no fix rounds needed)
**Gate Outcome:** PASS
**Invocation:** superpowers

### Review Coverage

| Agent | Status | Critical | Important | Minor | Notable |
|-------|--------|----------|-----------|-------|---------|
| Correctness | completed | 0 | 0 | 1 | 0 |
| Architecture & Idioms | completed | 0 | 0 | 0 | 0 |
| Security | completed | 0 | 0 | 0 | 0 |
| Production Readiness | completed | 0 | 0 | 0 | 1 |
| Test Quality | completed | 0 | 0 | 0 | 0 |
| CodeRabbit (external) | completed | 0 | 0 | 2 | 0 |
| **Total** | **6/6** | **0** | **0** | **3** | **1** |

### Gate Decision

**Critical + Important = 0. GATE PASS.**

No fix loop rounds were needed. All findings are Minor or Notable severity (informational only, excluded from gate check).

### Findings Summary

**FINDING-1 (Minor, correctness):** `context.Context` parameter in `GetRequestMetadata` is not forwarded to `TokenSource.Token()`. This is an inherent limitation of the `oauth2.TokenSource` interface which has no context parameter. Spec does not require context propagation.

**FINDING-2 (Notable, production-readiness):** Concurrent callers block on `mu.Lock()` during refresh. This is the expected tradeoff of the RWMutex double-checked locking pattern specified in the design, matching Kubernetes client-go behavior.

**FINDING-3 (Minor, correctness, CodeRabbit):** `isTokenValid()` returns true for zero-expiry tokens even if `AccessToken` is empty. The `oauth2.TokenSource` contract guarantees non-empty tokens on success. Spec does not require empty-token validation.

**FINDING-4 (Minor, correctness, CodeRabbit):** `WithLeeway` accepts negative durations without validation. A negative leeway is a caller error but not a security vulnerability. Spec does not explicitly require validation.

### External Tool Results

- **CodeRabbit:** completed, 2 Minor findings (FINDING-3, FINDING-4). No Critical/Important.
- **Copilot:** not invoked (not installed).

### Fix Loop

Not needed (Critical + Important = 0).

### Post-Fix Spec Compliance

Not needed (no code removed during review).

### Spec Compliance Score

**100%** (19/19 requirements verified: 11 FRs + 4 edge cases + 4 success criteria)

Full findings with rationale documented in [review-findings.md](review-findings.md).
