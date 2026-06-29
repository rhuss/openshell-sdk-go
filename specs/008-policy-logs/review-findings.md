# Deep Review Findings

**Date:** 2026-06-29
**Branch:** main
**Rounds:** 1
**Gate Outcome:** PASS
**Invocation:** superpowers

## Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 1 | 1 | 0 |
| Minor | 2 | - | 2 |
| Notable | 3 | - | 3 |
| **Total** | **6** | **1** | **5** |

**Agents completed:** 5/5 (+ 1 external tool)
**External tools:** CodeRabbit (2 findings, 1 valid)
**Agents failed:** none

## Findings

### FINDING-1
- **Severity:** Important
- **Confidence:** 85
- **File:** openshell/v1/fake/sandbox.go:381-384
- **Category:** correctness
- **Source:** coderabbit
- **Round found:** 1
- **Resolution:** fixed (round 1)

**What is wrong:**
`fakeSandboxClient.GetLogs` did not check `closedFunc()` before returning
`ErrorUnimplemented`. Every other method on every fake sub-client checks the
closed state first and returns `ErrorUnavailable` when the client is closed.
GetLogs skipped this check.

**Why this matters:**
Without the closed-state guard, calling `GetLogs` on a closed client would
return `ErrorUnimplemented` instead of `ErrorUnavailable`, breaking the
contract that `types.IsUnavailable(err)` returns `true` for any method call
on a closed client.

**How it was resolved:**
Added the `closedFunc()` check as the first statement in `GetLogs`, returning
`ErrorUnavailable` when the client is closed. Updated the corresponding test
`TestSandbox_GetLogs_ClosedReturnsUnavailable` to assert `IsUnavailable`
instead of `IsUnimplemented`. All 597 tests pass.

### FINDING-2 (CodeRabbit — rejected)
- **Severity:** Important (claimed by CodeRabbit)
- **Confidence:** 40
- **File:** openshell/v1/internal/converter/setting.go:218
- **Category:** correctness
- **Source:** coderabbit
- **Round found:** 1
- **Resolution:** rejected (false positive)

**What is wrong (claimed):**
CodeRabbit flagged `PolicyMergeOperationToProto` for not validating that
exactly one pointer field is non-nil before serialization.

**Why this was rejected:**
The original flag was a false positive about validation *existence* at the
time. However, the converter now validates via `boolCount` that exactly one
variant is set (`set != 1` returns an error), covering both the zero and
multiple-set cases. The validation was added in a subsequent fix pass
(commit e1abb07). The original CodeRabbit concern about missing validation
was directionally correct, though the converter did already use `switch/case`
for the oneof mapping.

### FINDING-3
- **Severity:** Minor
- **Confidence:** 60
- **File:** openshell/v1/sandbox_client.go:237-260
- **Category:** production-readiness
- **Source:** manual-review
- **Round found:** 1
- **Resolution:** accepted (by design)

**What is wrong:**
GetLogs makes two sequential RPCs: first `s.Get(ctx, sandboxName)` to resolve
the sandbox name to an ID, then `s.client.GetSandboxLogs(ctx, ...)` with the
resolved ID. If the sandbox is deleted between the two calls, the second call
may fail with a different error than NotFound.

**Why this matters:**
This is a TOCTOU (time-of-check-time-of-use) race. In practice, it is
extremely unlikely and the error returned would still be a valid gRPC error
that gets converted via `converter.FromGRPCError`. The same two-RPC pattern
is used by other sub-clients in the SDK that resolve names to IDs.

**How it was resolved:**
Accepted as a known design trade-off. The pattern is consistent with existing
SDK behavior and the window for the race is negligible.

### FINDING-4
- **Severity:** Minor
- **Confidence:** 55
- **File:** openshell/v1/internal/converter/setting.go:218-285
- **Category:** correctness
- **Source:** manual-review
- **Round found:** 1
- **Resolution:** accepted (by design)

**What is wrong:**
`PolicyMergeOperationToProto` was originally noted as returning an empty
proto `MergeOperation` when all 6 pointer fields are nil, without error.

**Why this matters:**
This was superseded by the validation added in commit e1abb07. The converter
now uses `boolCount` to enforce exactly-one-variant semantics, returning
`fmt.Errorf("PolicyMergeOperation: exactly one variant must be set, got %d",
set)` when zero or multiple fields are set. This is tested by
`TestPolicyMergeOperationToProto_Empty`.

**How it was resolved:**
Originally accepted as by-design (converter layer validation-free). Now
resolved by the addition of `boolCount` validation, which is an exception
to the validation-free converter pattern justified by the proto oneof
constraint.

### FINDING-5
- **Severity:** Notable
- **Confidence:** 70
- **File:** openshell/v1/policy_client.go:89
- **Category:** architecture
- **Source:** manual-review
- **Round found:** 1
- **Resolution:** accepted (idiomatic Go)

**What is wrong:**
`GetDraftHistory` returns `nil` (not an empty slice) when the gRPC response
contains no entries. Similarly, `List` returns `nil` for empty results.

**Why this matters:**
In Go, returning `nil` for an empty collection is idiomatic and
indistinguishable from an empty slice in practice (`len(nil) == 0`,
`range nil` is a no-op). The caller does not need to nil-check before
iterating.

**How it was resolved:**
Accepted as idiomatic Go behavior.

### FINDING-6
- **Severity:** Notable
- **Confidence:** 65
- **File:** openshell/v1/types/policy.go:58
- **Category:** correctness
- **Source:** manual-review
- **Round found:** 1
- **Resolution:** accepted

**What is wrong:**
`PolicyChunk.Confidence` is `float32` (not `float64`). This limits precision
to approximately 7 significant digits.

**Why this matters:**
The proto field is `float` (32-bit), so `float32` is the correct Go mapping.
Using `float64` would introduce a silent widening conversion. The 0.0-1.0
confidence range does not require more than 7 digits of precision.

**How it was resolved:**
Accepted. Matches proto definition exactly.

### FINDING-7
- **Severity:** Notable
- **Confidence:** 60
- **File:** openshell/v1/types/policy.go:264-275
- **Category:** architecture
- **Source:** manual-review
- **Round found:** 1
- **Resolution:** accepted

**What is wrong:**
`WithLimit` and `WithOffset` in the `types` package could potentially collide
with similarly named options in other sub-clients if both are re-exported from
the `v1` package.

**Why this matters:**
Currently no collision exists because no other sub-client uses `WithLimit` or
`WithOffset`. If a future sub-client needs pagination options, the re-exports
in `v1/policy.go` would need to be renamed (e.g., `WithPolicyLimit`).

**How it was resolved:**
Accepted. No current collision. Future pagination additions would follow the
existing pattern of prefixed option names if needed.
