# Code Review: Sandbox Name Resolution

**Spec:** specs/013-sandbox-name-resolution/spec.md
**Date:** 2026-06-30
**Reviewer:** Claude (speckit.spex-gates.review-code)

## Compliance Summary

**Overall Score: 100%**

- Functional Requirements: 10/10 (100%)
- Success Criteria: 5/5 (100%)

## Detailed Review

### Functional Requirements

#### FR-001: Exec methods accept sandbox name
**Implementation:** openshell/v1/exec.go (interface), openshell/v1/exec_client.go (resolution)
**Status:** Compliant
**Notes:** Run, Stream, Interactive all accept `sandboxName string` and resolve via `e.sandboxes.Get(ctx, sandboxName)` before using `sb.ID` in proto requests.

#### FR-002: TCP.Forward accepts sandbox name
**Implementation:** openshell/v1/tcp.go (interface), openshell/v1/tcp_client.go:26-38 (resolution)
**Status:** Compliant
**Notes:** Forward accepts `sandboxName string`, resolves via `t.sandboxes.Get(ctx, sandboxName)`, uses `sb.ID` in TcpForwardInit.SandboxId. Port validation correctly occurs before resolution.

#### FR-003: File Upload/Download accept sandbox name
**Implementation:** openshell/v1/file.go (interface), openshell/v1/file_client.go (resolution)
**Status:** Compliant
**Notes:** Both Upload and Download accept `sandboxName string`, resolve via `f.sandboxes.Get(ctx, sandboxName)`, use `sb.ID` in CreateSshSessionRequest.SandboxId.

#### FR-004: Config.GetSandbox accepts sandbox name
**Implementation:** openshell/v1/config.go (interface), openshell/v1/config_client.go:24-29 (resolution)
**Status:** Compliant
**Notes:** GetSandbox accepts `sandboxName string`, resolves via `c.sandboxes.Get(ctx, sandboxName)`, uses `sb.ID` in GetSandboxConfigRequest.SandboxId.

#### FR-005: Watch resolves name to ID before WatchSandboxRequest
**Implementation:** openshell/v1/sandbox_client.go:169-183
**Status:** Compliant
**Notes:** Watch calls `s.Get(ctx, name)` at line 176, then passes `sb.ID` (not name) to `WatchSandboxRequest{Id: sb.ID}` at line 183. Bug fix verified.

#### FR-006: Resolution uses SandboxInterface.Get()
**Implementation:** All sub-client files
**Status:** Compliant
**Notes:** Every sub-client calls `*.sandboxes.Get(ctx, sandboxName)` which is the `SandboxInterface.Get()` method.

#### FR-007: Resolution failure returns error without attempting RPC
**Implementation:** All sub-client files
**Status:** Compliant
**Notes:** All resolution calls return early on error. File tests verify `createCallCount==0` after resolution failure. Other tests verify NotFound error propagation.

#### FR-008: SSH.CreateSession remains ID-based with doc recommendation
**Implementation:** openshell/v1/ssh.go:34-41
**Status:** Compliant
**Notes:** CreateSession signature unchanged (`sandboxID string`). Doc comment at lines 38-41: "Note: CreateSession accepts a raw sandbox ID, not a name. For name-based access with automatic session lifecycle management, prefer [SSHInterface.Tunnel]..."

#### FR-009: Fake clients update signatures without resolution logic
**Implementation:** openshell/v1/fake/exec.go, fake/tcp.go, fake/file.go, fake/config.go
**Status:** Compliant
**Notes:** All fake method signatures use `_ string` for sandboxName parameter, matching updated interfaces without performing resolution.

#### FR-010: Sub-clients receive SandboxInterface at construction
**Implementation:** openshell/v1/client.go:96-101
**Status:** Compliant
**Notes:** Constructor wiring: `newExecClient(conn, c.sandboxes)`, `newFileClient(conn, c.sandboxes)`, `newSSHClient(conn, c.sandboxes)`, `newTCPClient(conn, c.sandboxes)`, `newConfigClient(conn, c.sandboxes)`.

### Success Criteria

#### SC-001: All public methods (except SSH.CreateSession) use sandbox name
**Status:** Compliant
**Notes:** Verified across ExecInterface, TCPInterface, FileInterface, ConfigInterface, SandboxInterface.Watch, SandboxInterface.GetLogs.

#### SC-002: Existing tests pass with updated call sites
**Status:** Compliant
**Notes:** `make ci` passes. All test setup functions use updated constructors with stubSandboxResolver.

#### SC-003: New resolution tests per sub-client
**Status:** Compliant
**Notes:** Each sub-client has `*_ResolvesNameToID` and `*_ResolutionError` tests:
- exec_client_test.go: TestExecRun_ResolvesNameToID, TestExecRun_ResolutionError
- tcp_client_test.go: TestTCPForward_ResolvesNameToID, TestTCPForward_ResolutionError
- file_client_test.go: TestFileUpload_ResolvesNameToID, TestFileUpload_ResolutionError
- config_client_test.go: TestConfigGetSandbox_ResolvesNameToID, TestConfigGetSandbox_ResolutionError
- sandbox_client_test.go: TestSandboxWatch_ResolvesNameToID, TestSandboxWatch_ResolutionError

#### SC-004: Watch sends resolved ID, not name
**Status:** Compliant
**Notes:** TestSandboxWatch_ResolvesNameToID asserts `req.GetId() == "resolved-id-123"` (the sandbox's actual ID from mock), not the name string.

#### SC-005: make ci passes cleanly
**Status:** Compliant
**Notes:** Lint: 0 issues. Build: success. Proto check: up to date. All tests pass with coverage.

### Edge Cases

#### Resolution succeeds but RPC fails
**Status:** Handled
**Notes:** RPC errors are returned as-is via `converter.FromGRPCError(err)`, not wrapped as resolution errors.

#### Multiple rapid calls (no caching)
**Status:** Handled
**Notes:** Each call independently resolves. No caching implemented, matching spec assumption.

#### Sandbox deleted between resolution and RPC
**Status:** Handled
**Notes:** RPC returns NotFound, surfaced as-is. No special handling needed.

#### SSH.CreateSession remains ID-based
**Status:** Handled
**Notes:** Doc comment recommends SSH().Tunnel() for name-based access.

### Extra Features (Not in Spec)

#### GetLogs name resolution
**Location:** openshell/v1/sandbox_client.go:243-248
**Description:** GetLogs also resolves sandbox name to ID before the proto request.
**Assessment:** Consistent extension of the pattern. GetLogs was not explicitly mentioned in the spec but follows the same SandboxInterface dependency pattern.
**Recommendation:** Add to spec via evolution (minor).

## Code Quality Notes

- Consistent resolution pattern across all sub-clients: resolve, check error, use sb.ID
- Resolution comments ("Resolve sandbox name to ID") are present in every resolution site
- Port validation in TCP.Forward correctly precedes resolution (fail fast)
- stubSandboxResolver test helper is well-designed with error injection via getErr field

## Recommendations

### Spec Evolution Candidates
- [ ] Add GetLogs to spec as FR-011 (already implemented, consistent with pattern)

### Optional Improvements
- [ ] Consider shared resolution helper to reduce boilerplate (4 lines per call site)

## Deep Review Report

**Review Date:** 2026-06-30
**Agents Dispatched:** 5 (correctness, architecture, security, production-readiness, test-quality)

### Correctness

- **Status:** PASS
- All name-to-ID resolution follows the established SSH.Tunnel pattern
- Watch bug fix confirmed: `sb.ID` passed to `WatchSandboxRequest.Id`, not raw name
- Resolution errors propagate without attempting the downstream RPC (fail-fast)
- No data races: resolution is synchronous, per-call, no shared mutable state

### Architecture

- **Status:** PASS
- Dependency injection pattern (SandboxInterface) is consistent with existing SSHClient
- Constructor wiring in client.go is clean and follows existing patterns
- No new abstractions introduced, no unnecessary complexity
- Breaking change is pre-1.0 acceptable

### Security

- **Status:** PASS (N/A for most concerns)
- No credentials or secrets involved in name resolution
- No new attack surface: resolution uses existing authenticated gRPC channel
- No user input passed unsanitized to proto fields

### Production Readiness

- **Status:** PASS
- Extra Get() round-trip per call is acceptable (documented in spec assumptions)
- No caching, so no cache invalidation bugs possible
- Error handling is consistent: resolution NotFound propagated cleanly
- No resource leaks: resolution is a simple unary RPC

### Test Quality

- **Status:** PASS
- Each sub-client has dedicated resolution success + error tests
- Tests use stubSandboxResolver with error injection
- File tests verify createCallCount==0 on resolution failure
- Watch test asserts resolved ID in proto request (not name)
- All 36 tests pass, 0 lint issues

### Findings Summary

| Severity | Count | Status |
|----------|-------|--------|
| Critical | 0 | - |
| Important | 0 | - |
| Minor | 1 | Noted |
| Informational | 1 | Noted |

**Minor-001:** GetLogs already implements name resolution but is not in the spec. Consistent with the pattern but technically an undocumented feature. Recommend spec evolution.

**Info-001:** Optional improvement: a shared `resolveSandboxID(ctx, sandboxes, name)` helper could reduce the 4-line resolution boilerplate. Low priority, current approach is clear and explicit.

### Gate Decision: **PASS**

No Critical or Important findings. Implementation is spec-compliant and production-ready.

## Conclusion

Implementation is 100% compliant with the specification. All 10 functional requirements and 5 success criteria are met. The Watch bug fix (FR-005/SC-004) is correctly implemented and tested. CI passes cleanly.
