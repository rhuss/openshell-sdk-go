# Deep Review Findings

**Date:** 2026-07-31
**Branch:** 022-workspace-crud-gatewayinfo
**Rounds:** 1
**Gate Outcome:** PASS
**Invocation:** manual

## Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 3 | 3 | 0 |
| Minor | 7 | 3 | 4 |
| Notable | 2 | - | 2 |
| **Total** | **12** | **6** | **6** |

**Agents completed:** 5/5 (+ 1 external tool)
**Agents failed:** none

## Findings

### FINDING-1
- **Severity:** Important
- **Confidence:** 95
- **File:** openshell/v1/internal/converter/errors.go:13-24, openshell/v1/types/errors.go, openshell/v1/health_client_test.go:254-262
- **Category:** correctness + test-quality
- **Source:** test-quality-agent (escalated to correctness during fix analysis)
- **Round found:** 1
- **Resolution:** fixed (round 1)

**What is wrong:**
`TestGetCurrentUser_Unauthenticated` only asserted `require.Error(t, err)` without checking the error type. The root cause was deeper: `codes.Unauthenticated` was missing from the `grpcToSDK` mapping in `converter/errors.go`, causing gRPC authentication errors to silently map to `ErrorInternal`. The SDK had no `ErrorUnauthenticated` code, no `IsUnauthenticated` helper, and no mapping.

**Why this matters:**
The spec (FR-010, User Story 4 scenario 2) requires "an authentication error is returned" when credentials are invalid. SDK consumers would receive `ErrorInternal` instead, making it impossible to distinguish auth failures from server errors.

**How it was resolved:**
Added `ErrorUnauthenticated` to the error code enum, added `IsUnauthenticated` helpers, added `codes.Unauthenticated` gRPC mapping, and strengthened the test assertion.

### FINDING-2
- **Severity:** Important
- **Confidence:** 85
- **File:** openshell/v1/workspace_test.go:305-319
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** fixed (round 1)

**What is wrong:**
`TestAddMember_Success` did not verify that the correct workspace name and principalSubject were sent in the gRPC request. The mock server did not capture `lastAddMemberReq`.

**Why this matters:**
If the implementation swapped or omitted fields in the proto request, the test would still pass.

**How it was resolved:**
Added `lastAddMemberReq` field to `mockWorkspaceServer`, captured the request in the handler, and added assertions for `GetWorkspace()` and `GetPrincipalSubject()`.

### FINDING-3
- **Severity:** Important
- **Confidence:** 82
- **File:** openshell/v1/internal/converter/workspace.go:76-85
- **Category:** architecture
- **Source:** architecture-agent (also reported by: correctness-agent, security-agent)
- **Round found:** 1
- **Resolution:** fixed (round 1)

**What is wrong:**
`WorkspaceRoleFromProto` defaulted unknown proto roles (including `WORKSPACE_ROLE_UNSPECIFIED`) to `WorkspaceRoleUser`. Every other enum converter in the codebase defaults to an "Unknown" variant (`WorkspacePhaseFromProto` to `WorkspaceUnknown`, `ServiceStatusFromProto` to `ServiceStatusUnknown`).

**Why this matters:**
Silent privilege assignment: if the server sends `UNSPECIFIED` due to a data issue or a future proto version adds new roles, they would silently map to User. The "Unknown" pattern exists specifically to surface unrecognized values rather than hiding them.

**How it was resolved:**
Added `WorkspaceRoleUnknown` constant to `types/workspace.go`, re-exported it in `workspace.go`, and updated the converter default case. Updated tests to match new behavior.

### FINDING-4
- **Severity:** Minor
- **Confidence:** 82
- **File:** openshell/v1/fake/fake.go:53-65
- **Category:** correctness
- **Source:** correctness-agent
- **Round found:** 1
- **Resolution:** fixed (round 1)

**What is wrong:**
`WithGatewayInfo` and `WithCurrentUser` option functions stored the caller's pointer directly without deep-copying at the input boundary.

**Why this matters:**
Mutating the passed object after `NewClient` construction would corrupt the fake's internal state, violating the deep-copy-at-boundaries invariant.

**How it was resolved:**
Changed both option functions to use `copyGatewayInfo` and `copyCurrentUser` helpers.

### FINDING-5
- **Severity:** Minor
- **Confidence:** 90
- **File:** openshell/v1/fake/workspace.go:93-98, 158-166
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** accepted (design limitation)

**What is wrong:**
The fake `List` and `ListMembers` methods ignore `ListOptions` (limit, offset, labelSelector).

**Why this matters:**
Users testing pagination against the fake get all results regardless of limit/offset. However, the spec requires fake clients to mirror "input validation and error cases," and the real client also doesn't implement pagination locally (it passes params to the server).

### FINDING-6
- **Severity:** Minor
- **Confidence:** 85
- **File:** openshell/v1/workspace_test.go:173-186
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** remaining

**What is wrong:**
`TestWorkspaceGet_Success` only asserts the workspace name, not Phase, Labels, or other fields.

### FINDING-7
- **Severity:** Minor
- **Confidence:** 80
- **File:** openshell/v1/workspace_test.go:214-229
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** remaining

**What is wrong:**
`TestWorkspaceList_Success` only asserts list length and first item's name.

### FINDING-8
- **Severity:** Minor
- **Confidence:** 80
- **File:** openshell/v1/fake/workspace_test.go:209-217
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** remaining

**What is wrong:**
`TestFakeWorkspace_ListMembers` checks count but does not assert member fields.

### FINDING-9
- **Severity:** Minor
- **Confidence:** 75
- **File:** openshell/v1/fake/workspace.go
- **Category:** architecture
- **Source:** architecture-agent
- **Round found:** 1
- **Resolution:** fixed (round 1)

**What is wrong:**
Missing documentation comments on fake List/ListMembers methods.

**How it was resolved:**
Added doc comments matching the pattern used by the sandbox fake.

### FINDING-10
- **Severity:** Minor
- **Confidence:** 75
- **File:** openshell/v1/fake/fake.go
- **Category:** architecture
- **Source:** architecture-agent
- **Round found:** 1
- **Resolution:** fixed (round 1)

**What is wrong:**
Missing `AddMember` seeding helper on `fake.Client`, breaking the pattern established by `AddSandbox`, `AddProvider`, `AddWorkspace`.

**How it was resolved:**
Added `AddMember(workspace string, m *types.WorkspaceMember)` to complete the seeding pattern.

## Notable Observations

### NOTABLE-1
- **File:** openshell/v1/fake/workspace.go (Delete method)
- **Category:** architecture
- **Source:** architecture-agent
- **Description:** Intentional non-idempotent Delete behavior in workspace fake: deleting a non-existent workspace returns NotFound rather than succeeding silently.
- **Rationale:** This matches the real client behavior where the gateway returns NotFound for non-existent resources. The pattern is consistent across all sub-clients.

### NOTABLE-2
- **File:** openshell/v1/fake/workspace.go:93-98
- **Category:** test-quality
- **Source:** test-quality-agent
- **Description:** Fake client's List methods return all stored items regardless of pagination parameters. This is a design decision, not a bug.
- **Rationale:** The real client passes pagination to the server. The fake has no server. Implementing in-memory pagination would add complexity with minimal testing value. Documented as a known limitation.

## Test Suite Results

| Round | Test Command | Exit Code | Failures | Status |
|-------|-------------|-----------|----------|--------|
| 1     | make test   | 0         | 0        | passed |

Test suite passed in all fix rounds.
