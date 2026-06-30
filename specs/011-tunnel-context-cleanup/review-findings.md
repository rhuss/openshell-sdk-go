# Deep Review Findings

**Date:** 2026-06-30
**Branch:** 011-tunnel-context-cleanup
**Rounds:** 0
**Gate Outcome:** PASS
**Invocation:** quality-gate

## Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 0 | 0 | 0 |
| Minor | 2 | - | 2 |
| Notable | 0 | - | 0 |
| **Total** | **2** | **0** | **2** |

**Agents completed:** 5/5 (+ 0 external tools)
**Agents failed:** none

## Findings

### FINDING-1
- **Severity:** Minor
- **Confidence:** 60
- **File:** openshell/v1/ssh_client.go:123-125
- **Category:** production-readiness
- **Source:** production-readiness-agent
- **Round found:** 1
- **Resolution:** remaining (minor, no fix required)

**What is wrong:**
The revocation RPC in `revokeFunc` uses `context.Background()` without an
explicit timeout. If the server is unreachable, the cleanup goroutine will
block until gRPC's connection-level timeout fires.

**Why this matters:**
In most deployments this is acceptable since gRPC has its own connection
timeouts. However, an explicit timeout (e.g., 30s) would bound the worst
case explicitly and make the behavior more predictable in degraded network
conditions.

**Recommendation:**
Consider using `context.WithTimeout(context.Background(), 30*time.Second)`
instead of plain `context.Background()` for the revocation call. This is a
defense-in-depth improvement, not a correctness issue.

---

### FINDING-2
- **Severity:** Minor
- **Confidence:** 70
- **File:** openshell/v1/ssh_client_test.go:590
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** remaining (minor, no fix required)

**What is wrong:**
`TestSSHTunnel_CloseBeforeContextCancel` uses `time.Sleep(50 * time.Millisecond)`
to wait for the cleanup goroutine, while the other two new tests use
`require.Eventually` for the same purpose.

**Why this matters:**
The `time.Sleep` approach is less robust on slow CI environments and
inconsistent with the pattern established by the other tests. In practice,
50ms is more than sufficient since the cleanup goroutine runs immediately
after `conn.done` closes (which happens during `Close()`), but
`require.Eventually` would be more resilient.

**Recommendation:**
Replace `time.Sleep(50ms)` with `require.Eventually` checking
`revokeCount == 1`, matching the pattern in the other two tests.
