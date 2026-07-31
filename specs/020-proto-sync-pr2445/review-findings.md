# Deep Review Findings

**Date:** 2026-07-31
**Branch:** 020-proto-sync-pr2445
**Rounds:** 0
**Gate Outcome:** PASS
**Invocation:** superpowers

## Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 0 | 0 | 0 |
| Minor | 0 | - | 0 |
| Notable | 1 | - | 1 |
| **Total** | **1** | **0** | **1** |

**Agents completed:** 5/5 (+ 1 external tool)
**Agents failed:** none

## Findings

### FINDING-1
- **Severity:** Notable
- **Confidence:** 50
- **File:** proto/openshell.proto:241-259
- **Category:** external
- **Source:** coderabbit
- **Round found:** 1
- **Resolution:** not-actionable (upstream file copied verbatim per FR-001)

**What is wrong:**
CodeRabbit flags that `ImportProviderProfiles`, `UpdateProviderProfiles`, and `DeleteProviderProfile` RPCs use `workspace_role: "admin"` authorization, suggesting they might be gateway-global operations that should instead require `global_role: "platform_admin"`.

**Why this matters:**
If provider type profiles are truly gateway-scoped rather than workspace-scoped, these RPCs would grant workspace admins access to modify gateway-wide resources. This is a design consideration for the upstream OpenShell API.

**How it was resolved:**
Not actionable in this PR. Per FR-001, all proto files are copied verbatim from the upstream OpenShell repository. Authorization annotations are defined upstream and reflect upstream's design decisions. Any change to authorization scoping must be proposed and reviewed in the upstream repository, not in the SDK.

**External rationale (CodeRabbit):**
Verify whether provider type profiles are gateway-global or workspace-scoped by inspecting the related RPC semantics and implementation. If global, update the authorization options for ImportProviderProfiles, UpdateProviderProfiles, and DeleteProviderProfile to require global_role "platform_admin", matching gateway-wide RPCs; if workspace-scoped, retain workspace_role "admin".
