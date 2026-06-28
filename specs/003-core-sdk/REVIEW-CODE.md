# Code Review: Core SDK (Phase 1)

**Date**: 2026-06-27
**Branch**: 003-core-sdk
**Reviewer**: Automated Deep Review

## Spec Compliance

**Score**: 19/19 requirements implemented

| Requirement | Status | Evidence |
|-------------|--------|----------|
| FR-001 | PASS | `client.go:16-23` Config struct with Address, TLS, Auth, Timeout, RetryPolicy, Logger |
| FR-002 | PASS | `client.go:100-112` Sub-client accessors: Sandboxes(), Providers(), Exec(), Files(), Health() |
| FR-003 | PASS | `sandbox.go:75`, `provider.go:30`, `exec.go:38`, `file.go:9`, `health.go:15` all interfaces |
| FR-004 | PASS | All domain types use `time.Time`, `string`, `int`, `map[string]string`. No proto imports in public API |
| FR-005 | PASS | `errors.go:62-93` Is* helpers: IsNotFound, IsAlreadyExists, IsUnavailable, IsPermissionDenied, IsInvalidArgument, IsDeadlineExceeded, IsCancelled |
| FR-006 | PASS | `watch.go:12-22` WatchInterface[T] with ResultChan() and Stop(), Event[T] with EventType |
| FR-007 | PASS | `exec.go:38-42` ExecInterface with Run, Stream, Interactive |
| FR-008 | PASS | `file.go:9-12` FileInterface with Upload and Download |
| FR-009 | PASS | `health.go:15-17` HealthInterface with Check |
| FR-010 | PASS | `provider.go:36` ProviderInterface.Ensure |
| FR-011 | PASS | `sandbox.go:80-82` AttachProvider, DetachProvider, ListProviders |
| FR-012 | PASS | `sandbox.go:83` WaitReady with context-based timeout |
| FR-013 | PASS | `client.go:115-120` Close with sync.Once for safe multi-call |
| FR-014 | PASS | All interface methods accept context.Context as first parameter |
| FR-015 | PASS | `internal/converter/` package with unexported conversion functions |
| FR-016 | PASS | Package path `openshell/v1` enables future v2 coexistence |
| FR-017 | PASS | sync.Once on Close, gRPC ClientConn is inherently goroutine-safe |
| FR-018 | PASS | `client.go:80` NewClient calls NewConnection eagerly, returns error on failure |
| FR-019 | PASS | `logger.go` Logger interface, `client.go:22` Config.Logger optional field |

## Deep Review Report

### Correctness

- All sub-client interfaces match the spec requirements exactly
- SandboxPhase values (Provisioning, Ready, Error, Deleting, Unknown) correctly map from proto enum values
- Error mapping covers all specified gRPC codes (NotFound, AlreadyExists, Unavailable, PermissionDenied, InvalidArgument, DeadlineExceeded, Cancelled, Internal)
- WatchInterface uses Go generics correctly with proper channel lifecycle (Stop closes done channel, cancels context)
- ExecStream interface correctly separates Next() for chunk iteration from ExitCode() for final result
- InteractiveSession implements io.Reader and io.Writer semantics for bidirectional I/O
- No correctness issues found

### Architecture

- Clean separation: public types in `openshell/v1/`, internal details in `internal/converter/` and `internal/grpc/`
- Consistent sub-client pattern: interface definition (*.go) separate from implementation (*_client.go)
- ClientInterface aggregates all sub-client accessors, enabling full interface-based mocking
- Proto isolation is complete: no proto imports in any public API file
- Converter package correctly handles the proto-to-SDK type boundary
- No architecture issues found

### Security

- AuthProvider interface follows gRPC PerRPCCredentials pattern for credential injection
- StaticToken correctly sets RequireTransportSecurity() to true, preventing token transmission over plaintext
- TLSConfig supports CertFile/KeyFile/CAFile for mTLS, Insecure flag clearly labeled for localhost only
- Credentials are write-only in Provider (ProviderSpec.Credentials used for Create/Update, not returned in Get responses)
- No credential logging or leaking detected
- No security issues found

### Production Readiness

- Client.Close() uses sync.Once to prevent double-close panics on the gRPC connection
- All operations propagate context.Context for cancellation and deadline support
- Watch uses separate goroutine with proper cleanup via Stop() and context cancellation
- Error types implement the standard `error` interface with `errors.As` support
- gRPC connection shared across sub-clients (no per-operation connections)
- Advisory: RetryPolicy is defined in Config but not wired into the gRPC dial options (acceptable per spec: "retry implementation is an internal concern" deferred to later)
- Advisory: Logger interface is defined but not used in any sub-client implementation yet (acceptable per spec: "no logging occurs when no logger is configured")

### Test Quality

- Coverage: 81.9% for `openshell/v1`, 98.4% for `internal/converter`
- All sub-clients tested against in-process mock gRPC servers using bufconn
- Tests cover both happy paths (successful operations) and error paths (gRPC errors mapped to StatusError)
- Converter tests verify all proto-to-SDK type mappings including edge cases (nil values, empty collections)
- Integration test stubs created with `//go:build integration` tag for future gateway testing
- Advisory: `internal/grpc` package has 0% coverage (connection setup is integration-level, not unit-testable without a real server)

## CI Verification

```
make ci: PASS
- lint: 0 issues
- build: success
- test: all passing
- proto:check: generated files up to date
```

## Summary

**Gate**: PASS
**Compliance**: 19/19 (100%)
**Findings**: 2 advisory (non-blocking)
  - RetryPolicy defined but not wired (by design, deferred)
  - Logger defined but not consumed (by design, no default logging)
**Auto-fixed**: 0 findings (none needed)
**Coverage**: 81.9% v1 package, 98.4% converter package
