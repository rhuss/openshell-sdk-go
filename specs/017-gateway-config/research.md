# Research: Gateway Config Convenience Layer

## R1: Rust CLI On-Disk Format

**Decision**: Match the Rust CLI's directory layout and file formats exactly.

**Findings**:
- Gateway configs live at `$XDG_CONFIG_HOME/openshell/gateways/<name>/metadata.json`
- System configs at `/etc/openshell/gateways/<name>/metadata.json`
- Active gateway marker: `$XDG_CONFIG_HOME/openshell/active_gateway` (plain text, single line)
- Token files sit alongside metadata: `edge_token` (plain text), `oidc_token.json` (JSON), `cf_token` (legacy)

**metadata.json schema** (from Rust CLI bootstrap):
```json
{
  "endpoint": "gateway.example.com:443",
  "auth_mode": "cloudflare_jwt",
  "name": "prod"
}
```

Fields: `endpoint` (string, required), `auth_mode` (string, optional), `name` (string, required).
Unknown fields are ignored for forward compatibility.

**oidc_token.json schema**:
```json
{
  "access_token": "eyJ...",
  "refresh_token": "dGhpcyBpcyBhIHJlZnJlc2g...",
  "token_type": "Bearer",
  "expiry": "2026-07-02T12:00:00Z"
}
```

**Rationale**: Read-only interop is the stated goal. Reproducing the format ensures any gateway configured via the Rust CLI works immediately from Go.

**Alternatives considered**: Abstract config format (rejected: loses convenience), protobuf config (rejected: Rust CLI uses JSON).

## R2: Auth Mode to AuthProvider Mapping

**Decision**: Direct mapping using existing SDK auth providers.

| auth_mode | AuthProvider | Token Source | TLS |
|-----------|-------------|-------------|-----|
| (unset/empty/none) | `v1.NoAuth()` | none | default |
| `plaintext` | `v1.NoAuth()` | none | insecure (TLS disabled) |
| `cloudflare_jwt` | `v1.StaticToken(edgeToken)` | `edge_token` file (lazy) | default |
| `oidc` | `v1.RefreshableToken(diskSrc)` | `oidc_token.json` (lazy) | default |
| `mtls` | error | n/a | n/a (out of scope) |
| (unknown) | error | n/a | n/a |

**Rationale**: Reuses existing, tested auth providers rather than introducing new ones. The gateway package is a wiring layer, not an auth layer.

**Alternatives considered**: New auth provider types (rejected: duplication), plugin registry (rejected: overengineered for 4 modes).

## R3: Naming Collision with v1.GatewayConfig

**Decision**: Use `gateway.Config` (package-qualified) for the on-disk metadata type.

**Finding**: `v1.GatewayConfig` already exists as a type alias for `types.GatewayConfig`, representing runtime gateway settings accessed via the Config API. This is semantically different from the on-disk gateway metadata.

**Rationale**: Go package scoping resolves this naturally. Users import `gateway` and use `gateway.Config`. No collision with `v1.GatewayConfig`.

## R4: Gateway Name Validation

**Decision**: Match the Rust CLI's `validated_gateway_name` behavior.

**Rules**:
- Must be non-empty
- Must not contain `/`, `\`, or null bytes
- Must not be `.` or `..`
- Must be a single path component (no separators)
- ASCII alphanumeric plus `-` and `_` only

**Rationale**: Prevents path traversal attacks. Matching Rust behavior ensures names valid in the CLI are valid in the Go SDK and vice versa.

## R5: XDG Base Directory Resolution

**Decision**: Follow XDG spec with `~/.config/` fallback.

**Resolution order**:
1. `$XDG_CONFIG_HOME/openshell/` if `XDG_CONFIG_HOME` is set and non-empty
2. `$HOME/.config/openshell/` as fallback
3. System directory `/etc/openshell/` as final fallback (for listing and lookup)

**Rationale**: Standard XDG behavior. The Go stdlib `os.UserConfigDir()` provides the right default on each platform.

## R6: Edge Package Dependency

**Decision**: Compile-time dependency on edge package. If edge package is not yet available, use `v1.StaticToken` directly for cloudflare_jwt mode.

**Finding**: The edge auth package (spec 018) may not be implemented when this spec ships. The `cloudflare_jwt` mapping only needs `StaticToken(edgeToken)` since the edge token is a static JWT read from disk. The `edge.CloudflareAccess` wrapper adds extra headers which can be handled via `WithExtraHeaders` if available, or deferred.

**Rationale**: No blocking dependency. StaticToken with the edge JWT is functionally correct for cloudflare_jwt gateways. The edge package adds convenience (extra headers) but is not required for basic auth.

## R7: Lazy Token Loading via oauth2.TokenSource

**Decision**: Implement `diskTokenSource` as an `oauth2.TokenSource` for OIDC mode.

**Finding**: `RefreshableToken` already accepts `oauth2.TokenSource`. A `diskTokenSource` that reads `oidc_token.json` on first `Token()` call satisfies both lazy loading and token refresh requirements.

For edge tokens, lazy loading is simpler: a sync.Once-guarded file read returning a static string for `StaticToken`.

**Rationale**: Reuses the existing `RefreshableToken` infrastructure. No new auth provider type needed.

## R8: Error Type Design

**Decision**: Package-level sentinel errors with `errors.Is` support, separate from the gRPC-based `StatusError`.

**Rationale**: Gateway errors are filesystem/config errors, not gRPC status codes. Using `StatusError` with `ErrorNotFound` would conflate "gateway config not on disk" with "gRPC resource not found". Distinct error types prevent this confusion.

**Error types**:
- `ErrGatewayNotFound`: no matching gateway directory in user or system paths
- `ErrConfigParse`: metadata.json is missing, malformed, or unreadable
- `ErrTokenLoad`: token file missing, unreadable, or malformed
- `ErrUnsupportedAuthMode`: auth_mode value not recognized
- `ErrInvalidGatewayName`: name fails validation (traversal, empty, etc.)
- `ErrNoActiveGateway`: active_gateway file missing or empty
