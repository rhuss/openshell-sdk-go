# Code Review: Global Policy Flag

**Branch**: `024-global-policy-flag`
**Date**: 2026-08-01
**Reviewer**: Claude Code (automated pipeline)

## Spec Compliance Check

**Score: 8/8 (100%)**

| Requirement | Status | Evidence |
|-------------|--------|----------|
| FR-001: Global mode for List | PASS | `Global: cfg.Global()` wired into `ListSandboxPoliciesRequest` at `policy_client.go:126` |
| FR-002: Global mode for GetStatus | PASS | `Global: cfg.Global()` wired into `GetSandboxPolicyStatusRequest` at `policy_client.go:112` |
| FR-003: Skip name validation on GetStatus when global | PASS | No client-side name validation exists; gateway handles this |
| FR-004: Skip workspace validation when global | PASS | No client-side workspace validation in List/GetStatus; gateway handles this |
| FR-005: Functional options pattern | PASS | `WithListGlobal(bool)` returns `ListPolicyOption`, `WithStatusGlobal(bool)` returns `GetStatusOption` |
| FR-006: Fake client with global support | PASS | Real in-memory List/GetStatus with `globalRevisions`/`sandboxRevisions` storage, deep copy, mutex |
| FR-007: Backward compatible | PASS | `global` defaults to `false`, all existing tests pass unchanged |
| FR-008: Documentation updates | PASS | Doc comments on WithListGlobal/WithStatusGlobal, global policy section in doc.go, README updated |

## Constitution Compliance

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | Global flag wired in client layer, not exposed as proto type |
| II. Idiomatic Go | PASS | Functional options pattern, consistent with existing code |
| III. Test-First | PASS | 19 new tests (8 real client, 11 fake client) |
| V. Minimal Dependencies | PASS | No new dependencies |
| VII. Deep Copy at Boundaries | PASS | `copySandboxPolicyRevision` deep-copies Policy pointer and NetworkPolicies map |
| IX. Agent-Friendly Docs | PASS | Doc comments on all new exported symbols |
| X. Proto-SDK Naming Fidelity | PASS | `global` proto field maps to `Global` option |
| XI. Fake-Real Parity | PASS | Fake List/GetStatus validate closed state, support global isolation |
| XIII. Documentation Accompanies | PASS | README and doc.go updated |

## Deep Review Report

### Review Dimensions

5 independent review agents assessed the implementation across correctness, architecture, security, production readiness, and test quality.

### Findings Summary

| Severity | Count | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 3 | 3 | 0 |
| Notable | 1 | 0 | 1 |
| Minor | 2 | 1 | 1 |

### Fixed Findings

1. **FIXED (Important, production-agent)**: `fake/policy.go` - Missing `sync.RWMutex` on `fakePolicyClient`. Added mutex with `Lock()` on AddGlobalRevision/AddRevision and `RLock()` on List/GetStatus, matching the pattern used by all other fake sub-clients.

2. **FIXED (Important, architecture-agent)**: `fake/policy.go:179-180` - Shallow copy of `SandboxPolicyRevision` violated deep-copy-at-boundaries invariant. Added `copySandboxPolicyRevision` helper that deep-copies `*SandboxPolicy` and its `NetworkPolicies` map. Applied in AddGlobalRevision, AddRevision, List return path, and GetStatus return paths.

3. **FIXED (Important, test-quality-agent)**: No test verified `global=true` with non-empty sandbox name/workspace (the "ignores" semantic). Added `TestPolicyGetStatus_WithGlobalIgnoresNonEmptyName` and `TestPolicyList_WithGlobalIgnoresWorkspace` proving parameters are forwarded but global flag takes precedence. Also added `TestFakePolicy_GetStatus_Sandbox` for sandbox-scoped GetStatus happy path.

### Remaining Findings (Non-blocking)

1. **Notable (security-agent)**: `fake/policy.go:40,104` - Composite key `workspace + "/" + name` has theoretical collision potential if names contain "/". Pre-existing pattern across the fake package (`objectStore`, inference fake). Test-only code, not a production concern.

2. **Minor (architecture-agent)**: `WithListGlobal(bool)` takes a bool parameter while `WithIncludeSecurityFlagged()` takes none. API style inconsistency but the bool parameter is more flexible. Not blocking.

### Agent Assessments

| Agent | Result | Critical | Important | Notable | Minor |
|-------|--------|----------|-----------|---------|-------|
| Correctness | PASS | 0 | 0 | 0 | 0 |
| Architecture | PASS | 0 | 1 | 0 | 1 |
| Security | PASS | 0 | 0 | 1 | 0 |
| Production | PASS | 0 | 1 | 0 | 0 |
| Tests | PASS | 0 | 1 | 0 | 0 |

### Gate Outcome

**PASS** - All 3 Important findings fixed. No critical or unresolved important findings remain.

### Test Results

```text
ok  github.com/rhuss/openshell-sdk-go/openshell/v1           coverage: 49.4%
ok  github.com/rhuss/openshell-sdk-go/openshell/v1/fake       coverage: 21.2%
```

All tests pass. 19 new tests added (8 real client, 11 fake client).
