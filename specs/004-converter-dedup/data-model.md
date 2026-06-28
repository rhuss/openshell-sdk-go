# Data Model: Converter Code Deduplication

## Package Dependency Graph (After Refactoring)

```
openshell/v1/types/          ← NEW: domain types (no imports from v1/ or converter/)
    ↑                ↑
    |                |
openshell/v1/        openshell/v1/internal/converter/
(client logic)       (proto ↔ SDK conversion)
    ↑
    |
openshell/v1/internal/grpc/
(connection management)
```

## Types Package Contents

### From `sandbox.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `Sandbox` | struct | ID, Name, CreatedAt, Labels, ResourceVersion, Spec, Status |
| `SandboxSpec` | struct | LogLevel, Environment, Template, Providers, GPUCount |
| `SandboxTemplate` | struct | Image, RuntimeClassName, AgentSocket, Labels, Annotations, Environment, UserNamespaces |
| `SandboxStatus` | struct | SandboxName, AgentPod, AgentFd, SandboxFd, Phase, Conditions, CurrentPolicyVersion |
| `SandboxCondition` | struct | Type, Status, Reason, Message, LastTransitionTime |
| `AttachProviderResult` | struct | Sandbox, Attached |
| `DetachProviderResult` | struct | Sandbox, Detached |

### From `provider.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `Provider` | struct | ID, Name, Type, CreatedAt, Labels, ResourceVersion, Spec |
| `ProviderSpec` | struct | Credentials, Config, CredentialExpiresAt |

### From `exec.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `ExecResult` | struct | ExitCode, Stdout, Stderr |
| `ExecChunk` | struct | Data, Stream |

### From `health.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `HealthResult` | struct | Status |

### From `types.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `SandboxPhase` | string enum | SandboxProvisioning, SandboxReady, SandboxError, SandboxDeleting, SandboxUnknown |
| `EventType` | string enum | EventAdded, EventModified, EventDeleted, EventError |
| `StreamType` | string enum | Stdout, Stderr |
| `TLSConfig` | struct | CertFile, KeyFile, CAFile, InsecureSkipVerify |
| `RetryPolicy` | struct | MaxRetries, InitialBackoff, MaxBackoff, BackoffMultiplier |

### From `errors.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `ErrorCode` | int enum | NotFound, AlreadyExists, Unavailable, PermissionDenied, InvalidArgument, DeadlineExceeded, Cancelled, Internal, Unknown |
| `StatusError` | struct | Code, Message |

### From `options.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `ExecOptions` | struct | Env, WorkDir |
| `ListOptions` | struct | Limit, Offset, LabelSelector |
| `WaitOptions` | struct | PollInterval |
| `WatchOptions` | struct | TimeoutSeconds, LabelSelector |
| `CreateOptions` | struct | (empty) |
| `DeleteOptions` | struct | (empty) |
| `GetOptions` | struct | (empty) |
| `UpdateOptions` | struct | (empty) |

### From `watch.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `Event[T]` | generic struct | Type (EventType), Object (T) |
| `WatchInterface[T]` | generic interface | ResultChan, Stop |

### From `logger.go` and `auth.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `Logger` | interface | Info, Error, Debug |
| `AuthProvider` | interface | GetRequestMetadata, RequireTransportSecurity |

### From `client.go`

| Type | Kind | Fields/Values |
|------|------|---------------|
| `Config` | struct | Address, TLS, Auth, Logger, RetryPolicy |

## Types Remaining in `v1/`

These are client operation types that stay in `v1/`:

| Type | Kind | Reason |
|------|------|--------|
| `Client` | struct | Client implementation |
| `ClientInterface` | interface | Client API surface |
| `SandboxInterface` | interface | Client operation methods |
| `ExecInterface` | interface | Client operation methods |
| `ProviderInterface` | interface | Client operation methods |
| `HealthInterface` | interface | Client operation methods |
| `FileInterface` | interface | Client operation methods |
| `ExecStream` | interface | Client streaming result |
| `InteractiveSession` | interface | Client interactive result |
