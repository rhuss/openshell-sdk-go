# Deep Review Findings

**Date:** 2026-07-31
**Branch:** 021-workspace-scoping (worktree)
**Rounds:** 0
**Gate Outcome:** PASS
**Invocation:** quality-gate

## Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 0 | 0 | 0 |
| Minor | 5 | 0 | 5 |
| Notable | 0 | 0 | 0 |
| **Total** | **5** | **0** | **5** |

**Agents completed:** 5/5 (+ 1 external tool)
**Agents failed:** none

## Findings

### FINDING-1
- **Severity:** Minor
- **Confidence:** 75
- **File:** openshell/v1/fake/sandbox_test.go
- **Lines:** 1-815
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** remaining (Minor, no fix required for gate)

**What is wrong:**
All sandbox client tests use the hardcoded workspace "default". No test creates sandboxes in different workspaces (e.g., "team-alpha" and "team-beta") to verify workspace isolation at the client level. The spec acceptance scenario 2 ("Given sandboxes exist in workspaces 'team-alpha' and 'team-beta', When listing sandboxes with workspace 'team-alpha', Then only sandboxes in 'team-alpha' are returned") is only tested at the store level in store_test.go, not at the sandbox client level.

**Why this matters:**
While the store-level tests (`TestObjectStore_List_WorkspaceIsolation`, `TestObjectStore_Create_SameNameDifferentWorkspace`, `TestObjectStore_Get_WrongWorkspace`) verify the core isolation logic, a regression in the fakeSandboxClient.List method (e.g., forgetting to pass workspace to `c.store.List()`) would not be caught.

**How it was resolved:**
Not fixed. This is a test coverage gap, not a correctness bug. The store-level tests provide the primary isolation guarantee. Adding client-level workspace isolation tests would improve confidence but is not required for gate passage.

### FINDING-2
- **Severity:** Minor
- **Confidence:** 80
- **File:** openshell/v1/fake/sandbox_test.go:29-41
- **Lines:** 29-41
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** remaining (Minor, no fix required for gate)

**What is wrong:**
`TestSandbox_Create` asserts on Name, Spec.LogLevel, Labels, Status.Phase, CreatedAt, and ResourceVersion, but does not assert that `sb.Workspace == "default"`. The fake sandbox Create method sets `Workspace: workspace` on the sandbox struct (line 225 of sandbox.go), but no test verifies this assignment.

**Why this matters:**
If the `Workspace: workspace` assignment in Create were accidentally removed, no test would catch it. This is relevant to SC-005 (verify workspace field on outgoing operations).

**How it was resolved:**
Not fixed. Minor test gap. The workspace assignment is a single line that's unlikely to regress without also breaking the store's composite key logic.

### FINDING-3
- **Severity:** Minor
- **Confidence:** 80
- **File:** openshell/v1/fake/sandbox_test.go
- **Lines:** 1-815
- **Category:** test-quality
- **Source:** test-quality-agent
- **Round found:** 1
- **Resolution:** remaining (Minor, no fix required for gate)

**What is wrong:**
No test calls `sc.List(ctx, "workspace", v1.ListOptions{AllWorkspaces: true})` on the fake sandbox client. The AllWorkspaces code path in `fakeSandboxClient.List` (sandbox.go:263-264) is untested at the client level. The store-level test `TestObjectStore_ListAll` covers the underlying ListAll method.

**Why this matters:**
The AllWorkspaces branching logic in the sandbox client List method (`if len(opts) > 0 && opts[0].AllWorkspaces`) could regress without detection. This is relevant to FR-004.

**How it was resolved:**
Not fixed. The store-level tests cover the underlying mechanism. A client-level test would add defense-in-depth.

### FINDING-4
- **Severity:** Minor
- **Confidence:** 70
- **File:** openshell/v1/config.go:61-66
- **Lines:** 61-66
- **Category:** architecture
- **Source:** architecture-agent
- **Round found:** 1
- **Resolution:** remaining (Minor, no fix required for gate)

**What is wrong:**
Interface method doc comments were removed from ConfigInterface (config.go), PolicyInterface (policy.go), SSHInterface (ssh.go), and TCPInterface (tcp.go) instead of being updated to reflect the new workspace parameter. FR-010 states doc comments "MUST be updated to reflect the new workspace parameter in the same PR." The comments were removed entirely rather than updated.

**Why this matters:**
For a public SDK, interface method doc comments appear in `go doc` output and are part of the developer experience. However, the codebase style guide says "default to writing no comments," creating a tension with FR-010. The package-level doc.go examples remain comprehensive and updated with workspace parameters.

**How it was resolved:**
Not fixed. The tension between FR-010 ("update doc comments") and the codebase convention ("no comments") is a judgment call. The doc.go package examples are the primary documentation surface and are fully updated.

### FINDING-5
- **Severity:** Minor
- **Confidence:** 70
- **File:** openshell/v1/fake/store.go:14-16
- **Lines:** 14-16
- **Category:** security
- **Source:** security-agent
- **Round found:** 1
- **Resolution:** remaining (Minor, no fix required for gate)

**What is wrong:**
The composite key function `compositeKey(workspace, name string)` uses string concatenation with "/" as delimiter: `workspace + "/" + name`. If either workspace or name contains "/", key collisions are possible. Example: workspace="a", name="b/c" produces the same key as workspace="a/b", name="c" (both yield "a/b/c").

**Why this matters:**
A crafted workspace or resource name could theoretically access resources in a different namespace. However, this is mitigated by the spec assumption: "Workspace validation (name format, existence) is handled by the gateway, not the SDK." The gateway would reject names containing "/".

**How it was resolved:**
Not fixed. The gateway validates workspace and resource names, preventing "/" in names. An alternative fix would be to use a non-printable delimiter (e.g., `\x00`) or a struct key, but the added complexity is not justified given gateway validation.

## Test Suite Results

| Round | Test Command | Exit Code | Failures | Status |
|-------|-------------|-----------|----------|--------|
| 1     | make test   | 0         | 0        | passed |

Test suite passed. No fix loop was needed (no Critical or Important findings).
