# Code Review: 017 - Gateway Config Convenience Layer

**Reviewer**: Claude Code (automated)
**Date**: 2026-07-02
**Spec**: specs/017-gateway-config/spec.md
**Package**: `openshell/v1/gateway/`

## Spec Compliance

| Requirement | Status | Evidence |
|---|---|---|
| FR-001: XDG metadata.json resolution | PASS | `paths.go:userConfigDir()` reads `XDG_CONFIG_HOME`, falls back to `~/.config/`. `config.go:parseMetadata()` reads `metadata.json`. Tests: `TestUserConfigDir_XDGSet`, `TestUserConfigDir_XDGUnset`. |
| FR-002: Active gateway resolution | PASS | `paths.go:resolveActiveGateway()` reads `active_gateway` plain text file, trims whitespace. Tests: `TestResolveActiveGateway_ValidName`, `_WhitespaceHandling`, `_FileMissing`, `_EmptyFile`. |
| FR-003: Gateway name validation | PASS | `paths.go:validateGatewayName()` rejects empty, path separators, dots, special characters, non-ASCII. Tests: `TestValidateGatewayName_ValidNames`, `_Empty`, `_PathSeparators`, `_Dots`, `_SpecialCharacters`. |
| FR-004: System directory fallback | PASS | `paths.go:resolveGatewayDir()` checks user dir first, then system dir. `systemConfigBase` is `/etc/openshell/gateways`. Tests: `TestResolveGatewayDir_UserDir`, `_UserPrecedenceOverSystem`. |
| FR-005: Edge token loading with cf_token fallback | PASS | `token.go:readEdgeToken()` reads `edge_token` then `cf_token`. Tests: `TestReadEdgeToken_PrimaryFile`, `_CfTokenFallback`, `_PrecedencePrimaryOverCf`, `_BothMissing`, `_EmptyFiles`. |
| FR-006: OIDC token bundle loading | PASS | `token.go:diskTokenSource.Token()` reads `oidc_token.json` with `access_token`, `refresh_token`, `expiry`, `expires_in`. Unknown fields ignored by `json.Unmarshal`. Tests: `TestDiskTokenSource_ValidBundle`, `_ExpiresIn`, `_ExpiryPrecedence`, `_MissingFile`, `_MalformedJSON`, `_MissingAccessToken`. |
| FR-007: Lazy token loading | PASS | Edge tokens use `lazyEdgeAuth` (implements `AuthProvider`, defers `loader.load()` to `GetRequestMetadata`). OIDC uses `diskTokenSource` (reads on `Token()` call). Both defer I/O past `NewClient` construction. Tests: `TestNewClient_MissingEdgeToken` (succeeds), `TestNewClient_MissingOIDCToken` (succeeds), `TestEdgeTokenLoader_LazyBehavior`. |
| FR-008: Auth mode mapping | PASS | `gateway.go:resolveAuthProvider()` maps none->NoAuth, plaintext->NoAuth, cloudflare_jwt->lazyEdgeAuth, oidc->RefreshableToken(diskTokenSource), mtls->ErrUnsupportedAuthMode. Tests: `TestResolveAuthProvider_None`, `_Plaintext`, `_CloudflareJWT`, `_OIDC`, `_MTLS`, `_UnknownMode`. |
| FR-009: One-call convenience constructor | PASS | `gateway.go:NewClient()` takes name + options, returns `*v1.Client`. Tests: `TestNewClient_AuthModeNone`, `_Plaintext`, `_CloudflareJWT`, `_OIDC`, `_ActiveGateway`. |
| FR-010: Config-only loader | PASS | `gateway.go:LoadConfig()` returns `*Config` without client creation. Tests: `TestLoadConfig_ValidConfig`, `_NotFound`, `_FrozenSnapshot`. |
| FR-011: Gateway enumeration | PASS | `gateway.go:ListGateways()` scans user+system dirs, deduplicates, marks active. Tests: `TestListGateways_MultipleGateways`, `_EmptyDirs`, `_ActiveStatus`. |
| FR-012: Client options | PASS | `options.go`: `WithLogger`, `WithTimeout`, `WithTLS`, `WithAuth`, `WithRetryPolicy`. Tests: `TestNewClient_WithOptions`, `TestNewClient_WithAuthOverride`. |
| FR-013: No credential leaks | PASS | Error messages use generic descriptions, never include token values. Tests: `TestNewClient_NoCredentialLeaks`, `TestResolveAuthProvider_NoTokenLeaks`. |
| FR-014: Filesystem isolation | PASS | All `os.ReadFile`, `os.ReadDir`, `os.Stat` calls are in `gateway/` package. Core SDK (`openshell/v1/`) has no filesystem operations. Verified by grep. |
| FR-015: Thread safety | PASS | `NewClient`, `LoadConfig`, `ListGateways` use no shared mutable state. `edgeTokenLoader` uses `sync.Once`. All functions documented as safe for concurrent use. |
| FR-016: Typed errors with errors.Is/As | PASS | `errors.go`: `ErrGatewayNotFound`, `ErrConfigParse`, `ErrTokenLoad`, `ErrUnsupportedAuthMode`, `ErrInvalidGatewayName`, `ErrNoActiveGateway`. All sentinel errors, wrapped with `%w`. Tests: `TestSentinelErrors_ErrorsIs`, `_WrappedErrors`, `_DoubleWrapped`. |
| FR-017: Unknown field tolerance | PASS | `config.go:parseMetadata()` uses `json.Unmarshal` into a struct, which silently ignores unknown JSON keys. Tests: `TestParseMetadata_UnknownFieldsIgnored`. |
| FR-018: System directory path | PASS | `paths.go:systemConfigBase = "/etc/openshell/gateways"`. Test: `TestSystemGatewayDir`. |

**Compliance Score: 18/18 (100%)**

All success criteria (SC-001 through SC-006) are met.

## Deep Review Report

### Correctness

**Severity: No issues remaining.**

One critical bug was found and fixed during this review:

- **[FIXED] FR-007 violation: Eager edge token loading in cloudflare_jwt mode.**
  In `resolveAuthProvider()`, the `AuthModeCloudflareJWT` case called `loader.load()` eagerly and passed the result to `v1.StaticToken()`. This meant `NewClient()` failed when the token file was missing, violating FR-007 (lazy loading). Fixed by introducing `lazyEdgeAuth` (a `types.AuthProvider` implementation in `token.go`) that defers `loader.load()` to `GetRequestMetadata()`. The corresponding test `TestNewClient_MissingEdgeToken` was updated to expect success (matching OIDC behavior).

All other correctness aspects are solid:
- `parseMetadata()` correctly validates required fields (endpoint) and maps auth modes.
- `resolveGatewayDir()` correctly implements user-first, system-fallback search order.
- `validateGatewayName()` rejects all path traversal vectors including dots, separators, and special characters.
- `resolveActiveGateway()` correctly trims whitespace and handles missing/empty files.
- `diskTokenSource.Token()` re-reads from disk on every call, enabling CLI-refreshed token pickup.
- `LoadConfig` returns an immutable snapshot (verified by `TestLoadConfig_FrozenSnapshot`).

### Architecture

**Severity: No issues.**

The package follows idiomatic Go patterns:

- **Functional options** (`ClientOption`) for extensible API without breaking changes.
- **Sentinel errors** with `%w` wrapping for proper `errors.Is`/`errors.As` support.
- **Interface compliance**: `diskTokenSource` implements `oauth2.TokenSource`; `lazyEdgeAuth` implements `types.AuthProvider`.
- **Clean separation**: All filesystem I/O is isolated in the `gateway/` package. The core SDK (`openshell/v1/`) remains filesystem-free (FR-014).
- **Shared internal logic**: `loadConfigInternal()` is used by both `NewClient` and `LoadConfig`, avoiding code duplication.
- **Package documentation** (`doc.go`) includes Quick Start examples for all three entry points.

Minor observations (not blocking):
- The `Config.Dir` field exposes the resolved directory path. This is useful for manual token loading but could be considered an implementation detail. Acceptable for a convenience layer targeting advanced users.
- `ListGateways` silently swallows `userConfigDir()` errors (line 124). This is correct behavior (system dir may still have gateways) but means a misconfigured `XDG_CONFIG_HOME` is silently ignored.

### Security

**Severity: No issues.**

- **Path traversal prevention**: `validateGatewayName()` rejects all names containing `/`, `\`, `.`, and non-alphanumeric characters beyond `-` and `_`. This runs before any filesystem access.
- **No credential leaks**: Error messages never include token values. Tests explicitly verify this (`TestNewClient_NoCredentialLeaks`, `TestResolveAuthProvider_NoTokenLeaks`).
- **Transport security**: `lazyEdgeAuth.RequireTransportSecurity()` returns `true`, preventing tokens from being sent over plaintext connections. `AuthModePlaintext` correctly uses `NoAuth()` with insecure TLS.
- **No directory listing exposure**: Error messages for missing gateways do not reveal the full filesystem path.

### Production Readiness

**Severity: No issues.**

- **Thread safety**: All exported functions are stateless or use `sync.Once` for caching. Documented in godoc comments.
- **Error handling**: Every error path returns a wrapped sentinel error, enabling programmatic error handling by callers.
- **Graceful degradation**: `ListGateways` returns an empty slice (not an error) when no gateways exist. User dir errors fall through to system dir.
- **Forward compatibility**: Unknown JSON fields are ignored, so newer Rust CLI versions adding fields won't break the Go package.
- **mTLS guidance**: The unsupported `mtls` auth mode returns a clear error message directing callers to use `WithAuth()` for custom auth providers.

### Test Quality

**Severity: No issues.**

- **Coverage**: 5 test files with 45+ test functions covering all code paths.
- **Error path coverage**: Tests verify error types (`errors.Is`), error messages (`Contains`), and error wrapping depth (double-wrapped sentinels).
- **Edge cases covered**: Empty files, missing files, whitespace handling, malformed JSON, unknown fields, case sensitivity, path traversal vectors, special characters, Unicode names.
- **Security tests**: Dedicated tests verify no credential leakage in error messages and string representations.
- **Lazy loading tests**: Both edge token (`TestNewClient_MissingEdgeToken`) and OIDC (`TestNewClient_MissingOIDCToken`) verify that `NewClient` succeeds with missing token files.
- **Snapshot isolation**: `TestLoadConfig_FrozenSnapshot` verifies that on-disk changes after `LoadConfig` don't affect the returned config.
- **Test helpers**: `setupGateway` and `setupGatewayWithTokens` keep tests DRY and readable.

## Gate Outcome

| Check | Result |
|---|---|
| Spec compliance | 18/18 (100%) |
| `make test` | PASS (all tests pass, including race detector) |
| `golangci-lint` | PASS (zero issues) |
| Critical issues | 1 found, 1 fixed (FR-007 lazy loading) |
| Important issues | 0 |

**GATE: PASS**
