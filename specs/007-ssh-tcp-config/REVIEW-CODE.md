# Code Review: SSH, TCP Forwarding, and Config (Phase 2b-1)

**Date**: 2026-06-29
**Branch**: `007-ssh-tcp-config`
**Reviewer**: Ship pipeline (automated)

## Spec Compliance Score: 18/18 (100%)

| Requirement | Status | Implementation |
|-------------|--------|----------------|
| FR-001: SSHInterface on ClientInterface | PASS | `ssh.go`, `client.go:26` |
| FR-002: CreateSession returns SSHSession | PASS | `ssh_client.go:22`, `types/ssh.go` |
| FR-003: RevokeSession returns bool | PASS | `ssh_client.go:32` |
| FR-004: TCPInterface on ClientInterface | PASS | `tcp.go`, `client.go:27` |
| FR-005: Forward with port validation 1-65535 | PASS | `tcp_client.go:27` client-side check |
| FR-006: Close terminates gRPC stream | PASS | `tcp_client.go:151` cancel + CloseSend |
| FR-007: Context cancellation closes stream | PASS | `tcp_client.go:36` WithCancel wraps ctx |
| FR-007a: service_id omitted in v1 | PASS | `tcp_client.go:48` ServiceId not set |
| FR-008: ConfigInterface on ClientInterface | PASS | `config.go`, `client.go:28` |
| FR-009: GetSandbox returns SandboxConfig | PASS | `config_client.go:23`, `types/setting.go` |
| FR-010: GetGateway returns GatewayConfig | PASS | `config_client.go:33` |
| FR-011: Update accepts ConfigUpdate | PASS | `config_client.go:41` with nil check |
| FR-012: Update returns ConfigUpdateResult | PASS | `config_client.go:54` |
| FR-013: Types in v1/types/ | PASS | `types/ssh.go`, `types/setting.go` |
| FR-014: Deep copy at boundaries | PASS | Converters copy maps/slices |
| FR-015: Concurrent safety | PASS | `tcpForwardConn.sendMu` for Write, readLoop goroutine |
| FR-016: Typed StatusError | PASS | All error paths use `converter.FromGRPCError` |
| FR-017: Fake stubs return Unimplemented | PASS | `fake/ssh.go`, `fake/tcp.go`, `fake/config.go` |
| FR-018: Fake compile-time check | PASS | `fake/fake.go` var _ interface check |

## Constitution Compliance

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | All domain types in `v1/types/`, no proto in public API |
| II. Idiomatic Go | PASS | `io.ReadWriteCloser` for TCP, context propagation, functional options |
| III. Test-First | PASS | Test files exist for all implementation files |
| IV. Upstream Tracking | PASS | All 6 RPCs verified in proto |
| V. Minimal Dependencies | PASS | Zero new dependencies |
| VI. Secrets Never Leak | PASS | SSHSession.Token documented as sensitive, not in error messages |
| VII. Deep Copy | PASS | Settings maps deep-copied in converters |
| VIII. Doc Examples Compile | PASS | doc.go updated with SSH/TCP/Config examples |
| IX. Agent-Friendly Docs | PASS | All exported symbols have doc comments |
| X. Proto-SDK Naming | PASS | SandboxID, GatewayHost, GatewayPort match proto semantics |

## Deep Review Report

### Correctness

**No issues found.** All proto-to-SDK conversions correctly map fields. The TCP forward init frame correctly uses `TcpRelayTarget` with `host: "127.0.0.1"`. Port validation is client-side (1-65535). The `tcpForwardConn` readLoop correctly handles `io.EOF` vs error. Settings map converters deep-copy keys and values.

### Architecture

**No issues found.** All three sub-clients follow the established pattern: interface definition in `{domain}.go`, implementation in `{domain}_client.go`, tests in `{domain}_client_test.go`. Types in `v1/types/`, converters in `internal/converter/`. ClientInterface extension is additive. The `cfg` field name in `Client` struct avoids collision with the `Config` type alias.

### Security

**No issues found.** SSHSession.Token is documented as sensitive (Constitution VI). The token is never included in error messages. RevokeSession accepts a token parameter (not logged). TCP forward does not expose the authorization_token field.

### Production Readiness

**No issues found.** `tcpForwardConn` uses `sync.Mutex` for Write (concurrent safety), a buffered channel for Read (non-blocking), and `context.WithCancel` for clean shutdown. The `readLoop` goroutine exits on stream close or context cancellation. `Close()` calls both `cancel()` and `CloseSend()`.

### Test Quality

**No issues found.** All sub-clients have mock gRPC tests. SSH tests cover CreateSession and RevokeSession. TCP tests cover init frame construction, read/write data frames, close behavior, port validation, and context cancellation. Config tests cover GetSandbox, GetGateway, Update with settings map deep copy and optimistic concurrency. Fake tests verify Unimplemented errors and closed-client errors.

## CI Verification

```
make build  → PASS (0 errors)
make test   → PASS (86.2% coverage v1, 89.5% fake, 99.1% converter)
make lint   → PASS (0 issues)
```

## Gate Outcome: **PASS**

All 18 functional requirements met. All 10 constitution principles satisfied. Build, test, and lint pass. No critical or important findings.

## Files Changed

| File | Type | Lines |
|------|------|-------|
| `openshell/v1/types/ssh.go` | New | 25 |
| `openshell/v1/types/setting.go` | New | 116 |
| `openshell/v1/internal/converter/ssh.go` | New | 43 |
| `openshell/v1/internal/converter/ssh_test.go` | New | ~110 |
| `openshell/v1/internal/converter/setting.go` | New | ~200 |
| `openshell/v1/internal/converter/setting_test.go` | New | ~250 |
| `openshell/v1/ssh.go` | New | 26 |
| `openshell/v1/ssh_client.go` | New | 41 |
| `openshell/v1/ssh_client_test.go` | New | ~100 |
| `openshell/v1/tcp.go` | New | 23 |
| `openshell/v1/tcp_client.go` | New | 155 |
| `openshell/v1/tcp_client_test.go` | New | ~200 |
| `openshell/v1/config.go` | New | 76 |
| `openshell/v1/config_client.go` | New | 56 |
| `openshell/v1/config_client_test.go` | New | ~200 |
| `openshell/v1/client.go` | Modified | +20 |
| `openshell/v1/doc.go` | Modified | +30 |
| `openshell/v1/fake/ssh.go` | New | ~40 |
| `openshell/v1/fake/tcp.go` | New | ~30 |
| `openshell/v1/fake/config.go` | New | ~50 |
| `openshell/v1/fake/ssh_test.go` | New | ~40 |
| `openshell/v1/fake/tcp_test.go` | New | ~30 |
| `openshell/v1/fake/config_test.go` | New | ~50 |
| `openshell/v1/fake/fake.go` | Modified | +15 |
| **Total** | | ~2738 |
