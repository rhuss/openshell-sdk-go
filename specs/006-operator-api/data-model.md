# Data Model: Operator API Extensions (Phase 2a)

## New Domain Types (in v1/types/)

### ServiceEndpoint

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Unique identifier from metadata |
| SandboxID | string | Sandbox object ID |
| SandboxName | string | Sandbox name |
| ServiceName | string | Service name within the sandbox |
| TargetPort | uint32 | Loopback TCP port inside the sandbox |
| Domain | bool | Whether browser-facing URL routing is enabled |
| URL | string | Browser-facing URL (empty when Domain=false) |

### ProviderProfile

| Field | Type | Description |
|-------|------|-------------|
| ID | string | Profile identifier |
| DisplayName | string | Human-readable name |
| Description | string | Profile description |
| Category | ProfileCategory | Profile category (Inference, Agent, etc.) |
| Credentials | []ProfileCredential | Credential definitions for this profile type |
| Endpoints | []NetworkEndpoint | Network endpoints provided by this profile |
| Binaries | []NetworkBinary | Binary artifacts provided by this profile |
| InferenceCapable | bool | Whether providers of this type support inference |
| Discovery | ProfileDiscovery | Local discovery configuration |
| ResourceVersion | uint64 | Optimistic concurrency version (0 for built-in) |

### ProfileCategory (string enum)

| Value | Description |
|-------|-------------|
| ProfileCategoryOther | Uncategorized |
| ProfileCategoryInference | Inference provider (LLM, embedding, etc.) |
| ProfileCategoryAgent | Agent framework |
| ProfileCategorySourceControl | Source control integration |
| ProfileCategoryMessaging | Messaging integration |
| ProfileCategoryData | Data integration |
| ProfileCategoryKnowledge | Knowledge base integration |

### ProfileCredential

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Credential name (e.g., "api_key") |
| Description | string | Human-readable description |
| Required | bool | Whether this credential is required |
| Secret | bool | Whether this credential is sensitive |

### NetworkEndpoint

| Field | Type | Description |
|-------|------|-------------|
| Host | string | Hostname or host glob pattern |
| Port | uint32 | TCP port |
| Protocol | string | Protocol (e.g., "rest", "websocket", "graphql") |

### NetworkBinary

| Field | Type | Description |
|-------|------|-------------|
| Path | string | Path within sandbox |

### ProfileDiscovery

| Field | Type | Description |
|-------|------|-------------|
| Credentials | []string | Credential names eligible for local discovery |

### ProfileImportItem

| Field | Type | Description |
|-------|------|-------------|
| Profile | ProviderProfile | Profile definition to import |
| Source | string | Origin identifier (e.g., file path) |

### ProfileDiagnostic

| Field | Type | Description |
|-------|------|-------------|
| Source | string | Import item source |
| ProfileID | string | Profile ID the diagnostic applies to |
| Field | string | Specific field with the issue |
| Message | string | Diagnostic message |
| Severity | string | Severity level |

### ImportResult

| Field | Type | Description |
|-------|------|-------------|
| Diagnostics | []ProfileDiagnostic | Validation diagnostics |
| Profiles | []ProviderProfile | Successfully imported profiles |
| Imported | bool | Whether any profiles were imported |

### UpdateResult

| Field | Type | Description |
|-------|------|-------------|
| Diagnostics | []ProfileDiagnostic | Validation diagnostics |
| Profile | *ProviderProfile | Updated profile (nil on failure) |
| Updated | bool | Whether the profile was updated |

### LintResult

| Field | Type | Description |
|-------|------|-------------|
| Diagnostics | []ProfileDiagnostic | Validation diagnostics |
| Valid | bool | Whether all profiles passed validation |

### RefreshStrategy (string enum)

| Value | Description |
|-------|-------------|
| RefreshStrategyStatic | Static credential (no automatic refresh) |
| RefreshStrategyExternal | External refresh mechanism |
| RefreshStrategyOAuth2RefreshToken | OAuth2 refresh token flow |
| RefreshStrategyOAuth2ClientCredentials | OAuth2 client credentials flow |
| RefreshStrategyGoogleServiceAccountJWT | Google service account JWT |

### RefreshStatus

| Field | Type | Description |
|-------|------|-------------|
| ProviderName | string | Provider name |
| ProviderID | string | Provider ID |
| CredentialKey | string | Credential key |
| Strategy | RefreshStrategy | Refresh strategy |
| Status | string | Current status (e.g., "active", "expired") |
| ExpiresAt | time.Time | When the credential expires |
| NextRefreshAt | time.Time | When the next refresh is scheduled |
| LastRefreshAt | time.Time | When the last refresh occurred |
| LastError | string | Last error message (empty if none) |

### RefreshConfig

| Field | Type | Description |
|-------|------|-------------|
| Provider | string | Provider name |
| CredentialKey | string | Credential key |
| Strategy | RefreshStrategy | Refresh strategy to use |
| Material | map[string]string | Refresh material key-value pairs |
| SecretMaterialKeys | []string | Keys in Material that are sensitive |
| ExpiresAt | *time.Time | Optional expiration time |

## Modified Existing Types

### WatchOptions (add field)

| Field | Type | Description |
|-------|------|-------------|
| StopOnTerminal | bool | Close watch when sandbox reaches Ready or Error |

## New Interfaces

### ServiceInterface (top-level)

```
client.Services() → ServiceInterface
  .Expose(ctx, sandbox, serviceName, targetPort, domain) → *ServiceEndpoint
  .Get(ctx, sandbox, serviceName) → *ServiceEndpoint
  .List(ctx, sandbox, opts ...ListOptions) → []*ServiceEndpoint
  .Delete(ctx, sandbox, serviceName) → error
```

### ProfileInterface (nested under Providers)

```
client.Providers().Profiles() → ProfileInterface
  .List(ctx) → []*ProviderProfile
  .Get(ctx, id) → *ProviderProfile
  .Import(ctx, items []ProfileImportItem) → *ImportResult
  .Update(ctx, item ProfileImportItem, id string, expectedVersion uint64) → *UpdateResult
  .Lint(ctx, items []ProfileImportItem) → *LintResult
  .Delete(ctx, id) → error
```

### RefreshInterface (nested under Providers)

```
client.Providers().Refresh() → RefreshInterface
  .GetStatus(ctx, provider, credentialKey) → *RefreshStatus
  .Configure(ctx, config RefreshConfig) → *RefreshStatus
  .Rotate(ctx, provider, credentialKey) → *RefreshStatus
  .Delete(ctx, provider, credentialKey) → error
```

## Relationships

```
ClientInterface
  ├── SandboxInterface (existing, WatchOptions gains StopOnTerminal)
  ├── ProviderInterface (existing, gains Profiles() and Refresh())
  │     ├── ProfileInterface (new, 6 methods)
  │     └── RefreshInterface (new, 4 methods)
  ├── ServiceInterface (new, 4 methods)
  ├── ExecInterface (existing)
  ├── FileInterface (existing)
  └── HealthInterface (existing)
```
