# Code Review: Add Workspace Scoping to All RPCs

**Spec:** specs/021-workspace-scoping/spec.md
**Date:** 2026-07-31
**Reviewer:** Claude (speckit.spex-gates.review-code)

## Compliance Summary

**Overall Score: 100%**

- Functional Requirements: 10/10 (100%)
- Error Handling: Compliant
- Edge Cases: Compliant

### Functional Requirements

| ID | Requirement | Status | Location |
|----|------------|--------|----------|
| FR-001 | Workspace as first param after ctx | Compliant | All 11 sub-client interfaces |
| FR-002 | Proto request workspace field | Compliant | All *_client.go files set `req.Workspace` |
| FR-003 | Fake client workspace isolation | Compliant | `fake/store.go` composite keys |
| FR-004 | AllWorkspaces on ListOptions | Compliant | `types/options.go`, all List methods |
| FR-005 | Gateway config no workspace | Compliant | `config.go:GetGateway()` |
| FR-006 | Workspace on exported types | Compliant | `types/service.go`, converters |
| FR-007 | Deep copy at boundaries | Compliant | `fake/store.go` copyFunc pattern |
| FR-008 | Factory methods no workspace | Compliant | `provider.go:Profiles()/Refresh()` |
| FR-009 | SSH/TCP sandbox resolution | Compliant | `ssh_client.go`, `tcp_client.go` |
| FR-010 | Doc examples updated | Compliant | `doc.go` ~22 examples updated |

## Deep Review Report

**Gate Outcome: PASS**
**Fix Rounds: 0**
**Review Agents: 5/5 completed**
**External Tools: CodeRabbit (no findings)**

### Findings Summary

| Severity | Found | Fixed | Remaining |
|----------|-------|-------|-----------|
| Critical | 0 | 0 | 0 |
| Important | 0 | 0 | 0 |
| Minor | 5 | 0 | 5 |
| Notable | 0 | 0 | 0 |
| **Total** | **5** | **0** | **5** |

### Review Perspectives

1. **Correctness**: All workspace parameters correctly propagated to proto requests. SSH/TCP clients properly resolve sandboxes via `Get(ctx, workspace, sandboxName)`. No nil-pointer or missing-field issues found.

2. **Architecture & Idioms**: Clean sub-client pattern with consistent workspace-first parameter ordering. Minor finding: interface doc comments removed rather than updated (FR-010 tension with codebase "no comments" convention).

3. **Security**: Composite key delimiter "/" in fake store could theoretically allow key collisions if workspace/name contain "/". Mitigated by gateway validation. No credential leakage found.

4. **Production Readiness**: Thread-safe objectStore with proper mutex usage. Deep copy at all boundaries. Graceful shutdown order preserved in SSH tunnel lifecycle.

5. **Test Quality**: Store-level workspace isolation tests are comprehensive. Three minor gaps at the fake client level: no cross-workspace isolation test, no Workspace field assertion on Create, no AllWorkspaces test on List. All underlying mechanisms tested at store level.

### Findings Detail

See `specs/021-workspace-scoping/review-findings.md` for full findings with code locations, rationale, and resolution status.

### Test Suite

All packages passed:
- `openshell/v1`: 2.904s, 49.1% coverage
- `openshell/v1/fake`: 3.001s, 16.6% coverage
- `openshell/v1/edge`: 2.419s
- `openshell/v1/gateway`: 2.200s
- `openshell/v1/internal/converter`: 3.127s
- `openshell/v1/oidc`: 27.649s
