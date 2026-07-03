# Data Model: OIDC Login Package

**Date**: 2026-07-03 | **Spec**: [spec.md](spec.md)

## Entities

### ProviderConfig (OIDC Discovery Document)

Represents the parsed `.well-known/openid-configuration` response. Cached in memory per issuer URL for the process lifetime.

| Field | Type | Description |
|-------|------|-------------|
| Issuer | string | The issuer identifier URL |
| AuthorizationEndpoint | string | URL for authorization requests |
| TokenEndpoint | string | URL for token exchange |
| DeviceAuthorizationEndpoint | string | URL for device code requests (may be empty) |
| ScopesSupported | []string | Scopes the provider supports |
| CodeChallengeMethodsSupported | []string | PKCE methods supported (e.g., "S256") |

**Lifecycle**: Created on first discovery fetch per issuer URL, cached for process lifetime via `sync.Once` per issuer.

### loginConfig (Internal Options Struct)

Holds resolved configuration for a single login attempt. Built by applying `LoginOption` functions.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| issuer | string | "" | OIDC issuer URL (from option or gateway metadata) |
| clientID | string | "" | OIDC client ID (from option or gateway metadata) |
| clientSecret | string | "" | Client secret (client credentials only) |
| scopes | []string | ["openid", "profile", "email"] | Requested scopes |
| callbackPort | int | 0 (auto: try 8000, 18000) | Fixed callback port for redirect URI |
| timeout | time.Duration | 2m | Timeout for interactive flows |
| keyboardFlow | bool | false | Force keyboard flow (skip browser) |
| inMemory | bool | false | Skip disk persistence |
| displayFunc | func(verificationURL, userCode string) | nil | Custom device code display |

### oidcBundle (On-Disk Token Format)

The on-disk JSON format in `oidc_token.json`. Shared with the Rust CLI and the existing `gateway/token.go` `diskTokenSource`.

| Field | JSON Key | Type | Description |
|-------|----------|------|-------------|
| AccessToken | access_token | string | The OAuth2 access token |
| RefreshToken | refresh_token | string | The OAuth2 refresh token (may be empty) |
| Expiry | expiry | string | RFC 3339 timestamp of token expiry |
| ExpiresIn | expires_in | int64 | Seconds until expiry (alternative to expiry) |

**Note**: This struct already exists in `gateway/token.go` as an unexported type. The OIDC package will define its own internal copy for writing, keeping the same JSON schema for interop.

### metadataJSON Extension (Gateway Config)

Two new optional fields added to the existing `metadataJSON` struct in `gateway/config.go`:

| Field | JSON Key | Type | Description |
|-------|----------|------|-------------|
| OIDCIssuer | oidc_issuer | string | OIDC provider issuer URL |
| OIDCClientID | oidc_client_id | string | OIDC client ID for this gateway |

**Backward compatibility**: Both fields are optional. Existing gateways without OIDC config will have empty values. The `parseMetadata` function does not validate these fields; the OIDC package checks them only when gateway-aware login is attempted.

## Relationships

```
Gateway Config (metadata.json)
  ├── endpoint (used by gateway.NewClient)
  ├── auth_mode (used by gateway.resolveAuthProvider)
  ├── oidc_issuer ──────► ProviderConfig (fetched via discovery)
  └── oidc_client_id ───► loginConfig.clientID

ProviderConfig
  ├── AuthorizationEndpoint ──► Auth Code + PKCE flow
  ├── TokenEndpoint ──────────► All flows (token exchange)
  └── DeviceAuthorizationEndpoint ──► Device Code flow

loginConfig ──► oidcBundle (written to disk after successful login)
                    │
                    ▼
              diskTokenSource (read by gateway.NewClient on next use)
```

## State Transitions

### Login Flow States

```
[Start]
   │
   ├── Gateway name provided?
   │   ├── Yes: Load metadata.json → extract oidc_issuer, oidc_client_id
   │   │        Check existing token on disk → valid? → [Return cached token]
   │   └── No: Require WithIssuer + WithClientID options
   │
   ▼
[Discovery] ── fetch .well-known/openid-configuration
   │
   ▼
[Auth Code Flow]
   ├── Generate PKCE verifier + challenge
   ├── Generate state parameter
   ├── Start callback server (try ports 8000, 18000, or custom)
   │   ├── Success: Open browser → wait for callback
   │   └── Failure: Fall back to keyboard flow
   ├── [Callback received] → validate state → exchange code for tokens
   │   ├── Exchange success → [Persist + Return]
   │   └── Exchange failure → [Error]
   └── [Timeout] → shutdown server → [Error]

[Keyboard Flow]
   ├── Print authorization URL
   ├── Read authorization code from stdin
   └── Exchange code for tokens → [Persist + Return]

[Device Code Flow]
   ├── Request device code
   ├── Display verification URL + user code (or call displayFunc)
   ├── Poll token endpoint at provider interval
   │   ├── authorization_pending → continue polling
   │   ├── slow_down → increase interval → continue polling
   │   ├── expired_token → [Error]
   │   └── success → [Persist + Return]
   └── [Context cancelled] → [Error]

[Client Credentials Flow]
   ├── Exchange client_id + client_secret for tokens
   └── [Return] (no disk persistence by default)

[Persist + Return]
   ├── inMemory? → skip disk write
   └── Write oidcBundle to <gateway-dir>/oidc_token.json
   └── Return *oauth2.Token
```
