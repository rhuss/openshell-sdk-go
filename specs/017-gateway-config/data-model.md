# Data Model: Gateway Config Convenience Layer

## Entities

### Config (gateway.Config)

Parsed representation of a gateway's on-disk metadata.json. Immutable
after construction (frozen snapshot).

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Name | string | metadata.json `name` | Validated gateway name |
| Endpoint | string | metadata.json `endpoint` | Host:port of the gateway |
| AuthMode | AuthMode | metadata.json `auth_mode` | Enum: None, Plaintext, CloudflareJWT, OIDC, MTLS |
| Source | ConfigSource | derived | User or System directory origin |
| Dir | string | derived | Absolute path to the gateway config directory |

### Info (gateway.Info)

Lightweight summary for listing. Does not load tokens or validate
config completeness.

| Field | Type | Source | Notes |
|-------|------|--------|-------|
| Name | string | directory name | Gateway name from directory listing |
| Active | bool | active_gateway file | Whether this is the active gateway |
| Source | ConfigSource | derived | User or System |

### AuthMode (gateway.AuthMode)

String enum for auth_mode values.

| Value | Constant | Maps to |
|-------|----------|---------|
| "" or "none" | AuthModeNone | v1.NoAuth() |
| "plaintext" | AuthModePlaintext | v1.NoAuth() + insecure TLS |
| "cloudflare_jwt" | AuthModeCloudflareJWT | Lazy edge token auth (lazyEdgeAuth) |
| "oidc" | AuthModeOIDC | v1.RefreshableToken(diskSrc) |
| "mtls" | AuthModeMTLS | error (out of scope) |

### ConfigSource (gateway.ConfigSource)

| Value | Constant | Meaning |
|-------|----------|---------|
| "user" | SourceUser | From $XDG_CONFIG_HOME/openshell/gateways/ |
| "system" | SourceSystem | From /etc/openshell/gateways/ |

### ClientOption (gateway.ClientOption)

Functional option for NewClient. Applied after gateway config is
resolved but before the v1.Client is created.

| Option Constructor | Effect |
|-------------------|--------|
| WithLogger(logger) | Sets logger on the v1.Config |
| WithTimeout(d) | Sets connection timeout |
| WithTLS(tlsCfg) | Overrides TLS settings from gateway config |
| WithAuth(provider) | Overrides the auto-resolved auth provider |
| WithRetryPolicy(p) | Sets retry policy on the v1.Config |

## Error Types

All errors implement the `error` interface. Callers use `errors.Is`
for classification.

| Error Variable | When Returned |
|---------------|---------------|
| ErrGatewayNotFound | No gateway directory found in user or system paths |
| ErrConfigParse | metadata.json missing, unreadable, or invalid JSON |
| ErrTokenLoad | Token file missing, unreadable, or malformed |
| ErrUnsupportedAuthMode | auth_mode not in known set |
| ErrInvalidGatewayName | Name fails validation (empty, traversal, bad chars) |
| ErrNoActiveGateway | active_gateway file missing or empty |

## On-Disk Layout (read-only)

```
$XDG_CONFIG_HOME/openshell/
├── active_gateway              # Plain text: gateway name
└── gateways/
    └── <name>/
        ├── metadata.json       # {"endpoint":"...","auth_mode":"...","name":"..."}
        ├── edge_token           # Plain text JWT (cloudflare_jwt mode)
        ├── cf_token             # Legacy name for edge_token
        └── oidc_token.json      # {"access_token":"...","refresh_token":"...","token_type":"...","expiry":"..."}

/etc/openshell/
└── gateways/                   # System-wide gateway configs (same structure)
    └── <name>/
        └── ...
```

## Relationships

```
Info (listing)           Config (detailed)           v1.Config (SDK)
  Name    ──────────────>  Name    ──────────────>     Address = Endpoint
  Active                   Endpoint                    Auth = resolved provider
  Source                   AuthMode ─── resolves ──>   TLS = resolved from mode
                           Source                      Logger = from ClientOption
                           Dir ─── token paths ──>     Timeout = from ClientOption
```

## State Transitions

None. All types are immutable snapshots. Config is read once from disk
and does not track changes. Token loading is lazy (deferred to first
auth call) but the Config itself does not change state.
