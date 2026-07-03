# Research: OIDC Login Package

**Date**: 2026-07-03 | **Spec**: [spec.md](spec.md)

## R1: Gateway OIDC Metadata Location

**Decision**: Extend the existing `metadataJSON` struct in `gateway/config.go` with `oidc_issuer` and `oidc_client_id` fields.

**Rationale**: The `metadataJSON` struct already uses `json.Unmarshal` with ignored unknown fields (forward compatibility). Adding two optional string fields is backward-compatible: existing gateways without OIDC config simply have empty values. The `parseMetadata` function does not need to require these fields; the OIDC package validates them only when gateway-aware login is attempted.

**Alternatives considered**:
- Separate `oidc_config.json` file: rejected because it fragments gateway config across files and requires additional file I/O
- Derive from gateway name convention: rejected because it's brittle and doesn't support custom OIDC providers

## R2: PKCE Implementation (RFC 7636)

**Decision**: Implement PKCE using stdlib `crypto/rand` for code verifier generation and `crypto/sha256` for S256 challenge.

**Rationale**: PKCE is straightforward: generate 32 random bytes, base64url-encode as the verifier, SHA-256 hash + base64url-encode as the challenge. No external library needed. The `oauth2` package doesn't provide PKCE helpers, so inline implementation is appropriate.

**Alternatives considered**:
- Use a third-party OIDC library (e.g., `coreos/go-oidc`): rejected per Constitution V (minimal dependencies). The OIDC discovery and token exchange are simple HTTP calls that don't justify adding a dependency.

## R3: Browser Opening Strategy

**Decision**: Use `os/exec` with platform-specific commands: `open` (macOS), `xdg-open` (Linux), `cmd /c start` (Windows). Detect platform via `runtime.GOOS`.

**Rationale**: This is the standard approach used by kubelogin and other Go CLI tools. No external dependency needed. If the command fails (no display, no browser), fall back to keyboard flow automatically per FR-006.

**Alternatives considered**:
- `github.com/pkg/browser`: rejected per Constitution V. The implementation is trivial (< 20 lines).

## R4: Localhost Callback Server Design

**Decision**: Create a temporary `net/http` server per login attempt. Try port 8000 first, then 18000, with `WithCallbackPort` override. The server handles exactly one callback request, extracts the code and state, responds with an HTML success page, then signals the waiting goroutine via a channel.

**Rationale**: Per-attempt servers (FR-020) avoid port conflicts between concurrent logins to different gateways. Using channels for synchronization is idiomatic Go. The server is shut down via `http.Server.Shutdown(ctx)` after receiving the callback or on context cancellation/timeout.

**Alternatives considered**:
- Shared long-running server: rejected because it complicates lifecycle management and creates port contention
- Random port selection: rejected because many OIDC providers require pre-registered redirect URIs with specific ports

## R5: Token Persistence Format

**Decision**: Write the `oidcBundle` JSON format: `{access_token, refresh_token, expiry (RFC 3339), expires_in (seconds)}`. Read via `gateway.diskTokenSource.Token()`.

**Rationale**: The existing `oidcBundle` struct in `gateway/token.go` defines the on-disk format. The OIDC package must write this exact format for interop with `gateway.NewClient` (which reads via `diskTokenSource`) and the Rust CLI (which reads/writes the same file). The `expiry` field uses RFC 3339, and `expires_in` provides a seconds-based fallback.

**Alternatives considered**:
- New token format with additional fields: rejected for interop reasons. The existing format is sufficient.

## R6: State Parameter for CSRF Protection

**Decision**: Generate a 16-byte random `state` parameter using `crypto/rand`, base64url-encode it, and validate it in the callback. Store it in memory (not disk) for the duration of the login attempt.

**Rationale**: The `state` parameter prevents CSRF attacks where a malicious site could redirect a user's browser to the callback URL with a forged authorization code. Per OIDC Core spec, the state must be opaque, unpredictable, and verified on callback receipt.

## R7: Context Propagation Design

**Decision**: All public functions take `context.Context` as first parameter. The context controls: HTTP client timeouts for discovery/token exchange, callback server shutdown, device code polling cancellation. The default 2-minute timeout (FR-013) is applied as a `context.WithTimeout` wrapper when the caller doesn't set a deadline.

**Rationale**: Context-first is idiomatic Go (Constitution II). It enables callers to cancel long-running interactive flows, set custom deadlines, and propagate cancellation across goroutines (e.g., shutting down the callback server when the parent context is cancelled).

## R8: Existing Token Reuse

**Decision**: When `Login` is called with a gateway name, check `<gateway-dir>/oidc_token.json` for a valid token before starting an interactive flow. "Valid" means: file exists, parses correctly, and `expiry` is in the future (with the same leeway used by `RefreshableToken`, 10 seconds). If valid, return the token without user interaction.

**Rationale**: FR-019 requires this. Avoids unnecessary re-authentication when tokens are still valid, improving UX for repeated CLI invocations. Uses the same `oidcBundle` parsing logic as `diskTokenSource.Token()`.

## R9: Error Type Design

**Decision**: Define package-level sentinel errors following the `gateway/errors.go` pattern: `ErrDiscovery`, `ErrAuthCode`, `ErrDeviceCode`, `ErrClientCredentials`, `ErrTimeout`, `ErrCallbackServer`, `ErrTokenPersist`, `ErrOIDCConfig`. All errors wrap these sentinels using `fmt.Errorf("%w: ...")` for `errors.Is()` matching.

**Rationale**: Typed errors enable callers to handle specific failure modes programmatically (e.g., retry on timeout, prompt for manual login on callback server failure). This follows the existing SDK error pattern in `gateway/errors.go`.
