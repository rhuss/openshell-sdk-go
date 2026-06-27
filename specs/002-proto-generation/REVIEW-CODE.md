# Code Review: Proto Generation Pipeline

**Spec:** specs/002-proto-generation/spec.md
**Date:** 2026-06-27
**Reviewer:** Claude (speckit.spex-gates.review-code)

## Compliance Summary

**Overall Score: 100%**

- Functional Requirements: 11/11 (100%)
- Error Handling: 4/4 (100%)
- Edge Cases: 4/4 (100%)
- Non-Functional: 5/5 (100%)

## Detailed Review

### Functional Requirements

#### FR-001: Copy exactly three proto files
**Implementation:** mise.toml:178 (`PROTO_FILES="openshell.proto datamodel.proto sandbox.proto"`)
**Status:** Compliant

#### FR-002: Must NOT copy unneeded protos
**Implementation:** mise.toml:178-186 (explicit file list, individual copy loop)
**Status:** Compliant

#### FR-003: Record upstream git commit SHA
**Implementation:** mise.toml:190-196 (git rev-parse HEAD, write to UPSTREAM_VERSION)
**Status:** Compliant

#### FR-004: Generate with protoc + plugins
**Implementation:** mise.toml:64-73 (protoc with --go_out, --go-grpc_out)
**Status:** Compliant

#### FR-005: Use --go_opt=M flags
**Implementation:** mise.toml:53-59 (6 M flags for 3 proto files)
**Status:** Compliant

#### FR-006: Versioned Go packages
**Implementation:** mise.toml:50 (mkdir openshellv1, datamodelv1, sandboxv1)
**Status:** Compliant

#### FR-007: Validation task (staleness detection)
**Implementation:** mise.toml:82-140 (proto:check generates to temp dir, diffs)
**Status:** Compliant

#### FR-008: Clean task
**Implementation:** mise.toml:142-159 (find -name "*.pb.go" -delete)
**Status:** Compliant

#### FR-009: Pin tool versions in mise.toml
**Implementation:** mise.toml:2-6 (protoc 29.6, protoc-gen-go 1.36.11, protoc-gen-go-grpc 1.6.2)
**Status:** Compliant

#### FR-010: Configurable upstream path
**Implementation:** mise.toml:167 (`UPSTREAM_PATH` env var with default)
**Status:** Compliant

#### FR-011: Proto packages not public API
**Implementation:** By design (proto/ packages not re-exported from openshell/)
**Status:** Compliant

### Error Handling

| Error Case | Location | Status |
|---|---|---|
| Upstream path not found | mise.toml:169-172 | Compliant |
| protoc/plugins missing | mise.toml:39-44 | Compliant |
| Proto syntax errors | Delegated to protoc | Compliant |
| No .git in upstream | mise.toml:191-195 | Compliant |

### Extra Features (Not in Spec)

None found.

## Deep Review Report

### Review Method

5 specialized review agents dispatched sequentially (model: opus), each
reviewing the implementation from a single perspective with no implementation
context. Findings consolidated and deduplicated below.

### Findings by Dimension

#### 1. Correctness (Agent 1/5)

- **[Important] FIXED: protoc plugins pinned to `latest`** (mise.toml:4-5)
  Pinned to explicit versions: protoc-gen-go 1.36.11, protoc-gen-go-grpc 1.6.2.

- **[Important] FIXED: `TMPDIR` trap quoting fragile** (mise.toml:99)
  Renamed to `WORK_DIR`, trap uses single quotes with inner double quotes.

- **[Minor] Unquoted `$M_FLAGS` expansion** (mise.toml:70, 120)
  Intentional word splitting for multi-argument expansion. Acceptable.

- **[Minor] `proto:sync` resolves parent via `cd "$UPSTREAM_PATH/.."**
  Follows physical path through symlinks. Correct for finding `.git`.

#### 2. Architecture (Agent 2/5)

- **[Important] ACKNOWLEDGED: DRY violation in M_FLAGS** (mise.toml:53-59, 105-111)
  M_FLAGS and protoc invocation duplicated between proto:gen and proto:check.
  Accepted for now: extracting to a shared script adds complexity for 3 proto
  files. Will refactor if/when a 4th proto is added.

- **[Minor] Proto file list hardcoded in 3 places**
  proto:gen, proto:check, and proto:sync each list the 3 filenames independently.
  Accepted for same reason as above.

- **[Info] Proto layout follows Go conventions correctly.**
- **[Info] Proto isolation (constitution) holds.**
- **[Info] CI structure is sound (4 parallel jobs).**

#### 3. Security (Agent 3/5)

- **[Important] FIXED: Supply chain, plugins pinned to `latest`**
  Same as correctness finding 1. Now pinned to explicit versions.

- **[Important] FIXED: Trap quoting**
  Same as correctness finding 2. Now uses proper quoting.

- **[Minor] Path traversal via UPSTREAM_PATH**
  Developer-only task, not run in CI. Low risk. Accepted.

- **[Info] CI permissions follow least privilege (`contents: read`).**
- **[Info] proto:sync correctly excluded from CI workflows.**

#### 4. Production Readiness (Agent 4/5)

- **[Important] FIXED: Non-reproducible plugin versions**
  Same as correctness finding 1. Now pinned.

- **[Important] ACKNOWLEDGED: proto:check diff could flag non-generated files**
  The `--exclude` flags handle `.proto` and `UPSTREAM_VERSION`. Any other
  non-generated file in proto/ subdirs would cause a spurious failure.
  Accepted: this is the desired behavior (proto/ subdirs should only contain
  generated files).

- **[Minor] FIXED: TMPDIR variable shadows system env**
  Renamed to WORK_DIR.

- **[Minor] Cross-platform: bash scripts won't work on Windows**
  Acceptable per spec (target: Linux, macOS). Windows not in scope.

#### 5. Test Quality (Agent 5/5)

- **[Important] ACKNOWLEDGED: No Go tests for generated proto code**
  The spec focuses on the generation pipeline, not on testing generated types.
  Go-level tests (instantiate messages, verify stubs) belong in the core SDK
  phase (brainstorm 004). proto:check validates generation correctness.

- **[Important] ACKNOWLEDGED: proto:check duplicates proto:gen logic**
  Same as architecture DRY finding. Accepted for now.

- **[Minor] Edge case paths (missing tools, syntax errors) untested in CI**
  These are developer-facing error paths. Testing would require removing
  tools from PATH, which is fragile in CI. Accepted.

- **[Info] proto:check in CI task deps is well-designed.**

### Fix Summary

| Finding | Severity | Action | Status |
|---------|----------|--------|--------|
| Plugin versions `latest` | Important | Pinned to 1.36.11 / 1.6.2 | FIXED |
| Trap quoting fragile | Important | Single-quoted trap, renamed var | FIXED |
| TMPDIR shadows system var | Minor | Renamed to WORK_DIR | FIXED |
| DRY: M_FLAGS duplicated | Important | Accepted (3 files, refactor later) | ACKNOWLEDGED |
| No Go tests for generated code | Important | Out of scope (core SDK phase) | ACKNOWLEDGED |
| proto:check diff false positives | Important | Desired behavior | ACKNOWLEDGED |

### Gate Outcome: **PASS**

0 Critical findings. 3 Important findings fixed, 3 Important acknowledged with
justification. All Minor findings documented. `make ci` passes after fixes.

### Verification

```
$ make ci
[lint] 0 issues.
[build] Finished in 435ms
[test] ok (coverage: 100.0% openshell)
[proto:check] Proto check passed: generated files are up to date.
Finished in 1.10s
```
