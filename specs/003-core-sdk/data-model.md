# Data Model: Core SDK (Phase 1)

## SDK Domain Types

These are the public types exposed to consumers. They are independent of proto-generated types.

### Config

| Field | Type | Description |
|-------|------|-------------|
| Address | string | Gateway host:port |
| TLS | *TLSConfig | TLS configuration. Nil = system defaults |
| Auth | AuthProvider | Credential provider interface |
| Timeout | time.Duration | Default timeout for unary RPCs. Zero = no timeout |
| RetryPolicy | *RetryPolicy | Retry configuration. Nil = no retries |
| Logger | Logger | Optional structured logger. Nil = no logging |

### TLSConfig

| Field | Type | Description |
|-------|------|-------------|
| CertFile | string | Client certificate path |
| KeyFile | string | Client key path |
| CAFile | string | CA certificate path |
| Insecure | bool | Skip TLS (localhost dev only) |

### Sandbox

| Field | Type | Proto Source |
|-------|------|-------------|
| Name | string | metadata.name |
| ID | string | metadata.id |
| Phase | SandboxPhase | status.phase (enum mapping) |
| Image | string | spec.template.image |
| Environment | map[string]string | spec.environment |
| Labels | map[string]string | metadata.labels |
| CreatedAt | time.Time | metadata.created_at_ms (ms -> time.Time) |
| ResourceVersion | uint64 | metadata.resource_version |
| Conditions | []SandboxCondition | status.conditions |

### SandboxPhase (string enum)

| SDK Constant | Proto Value |
|-------------|-------------|
| SandboxProvisioning | SANDBOX_PHASE_PROVISIONING |
| SandboxReady | SANDBOX_PHASE_READY |
| SandboxError | SANDBOX_PHASE_ERROR |
| SandboxDeleting | SANDBOX_PHASE_DELETING |
| SandboxUnknown | SANDBOX_PHASE_UNKNOWN |

### SandboxSpec (creation input)

| Field | Type | Description |
|-------|------|-------------|
| Image | string | Container/VM image |
| Environment | map[string]string | Environment variables |
| Providers | []string | Provider names to attach |
| Labels | map[string]string | Labels for the sandbox |

### SandboxCondition

| Field | Type | Description |
|-------|------|-------------|
| Type | string | Condition class |
| Status | string | True/False/Unknown |
| Reason | string | Machine-readable reason |
| Message | string | Human-readable message |

### Provider

| Field | Type | Proto Source |
|-------|------|-------------|
| Name | string | metadata.name |
| ID | string | metadata.id |
| Type | string | type (e.g., "claude", "gitlab") |
| Config | map[string]string | config (non-secret) |
| Labels | map[string]string | metadata.labels |
| CreatedAt | time.Time | metadata.created_at_ms |
| ResourceVersion | uint64 | metadata.resource_version |

Note: `credentials` and `credential_expires_at_ms` are excluded from the SDK Provider type. Credentials are write-only (set during Create/Update, never returned to the consumer).

### ProviderSpec (creation/update input)

| Field | Type | Description |
|-------|------|-------------|
| Name | string | Provider name |
| Type | string | Provider type slug |
| Credentials | map[string]string | Secret values (write-only) |
| Config | map[string]string | Non-secret configuration |
| Labels | map[string]string | Labels |

### ExecResult

| Field | Type | Description |
|-------|------|-------------|
| ExitCode | int | Process exit code |
| Stdout | []byte | Standard output |
| Stderr | []byte | Standard error |

### ExecStream

| Method | Signature | Description |
|--------|-----------|-------------|
| Next | () (*ExecChunk, error) | Returns next output chunk. io.EOF on completion |
| ExitCode | () (int, error) | Blocks until completion, returns exit code |
| Close | () error | Closes the stream |

### ExecChunk

| Field | Type | Description |
|-------|------|-------------|
| Stream | StreamType | Stdout or Stderr |
| Data | []byte | Output bytes |

### StreamType (string enum)

| Constant | Value |
|----------|-------|
| StreamStdout | "stdout" |
| StreamStderr | "stderr" |

### InteractiveSession

| Method | Signature | Description |
|--------|-----------|-------------|
| Read | (p []byte) (int, error) | Read stdout (implements io.Reader) |
| Write | (p []byte) (int, error) | Write stdin (implements io.Writer) |
| Resize | (cols, rows uint32) error | Send terminal resize |
| ExitCode | () (int, error) | Blocks until completion |
| Close | () error | Closes the session |

### WatchInterface[T]

| Method | Signature | Description |
|--------|-----------|-------------|
| ResultChan | () <-chan Event[T] | Channel delivering typed events |
| Stop | () | Stops the watch and closes channel |

### Event[T]

| Field | Type | Description |
|-------|------|-------------|
| Type | EventType | Added, Modified, Deleted, Error |
| Object | T | The resource that changed |

### EventType (string enum)

| Constant | Value |
|----------|-------|
| EventAdded | "ADDED" |
| EventModified | "MODIFIED" |
| EventDeleted | "DELETED" |
| EventError | "ERROR" |

### StatusError

| Field | Type | Description |
|-------|------|-------------|
| Code | ErrorCode | SDK error code |
| Message | string | Human-readable message |
| Details | map[string]string | Optional error details |

### ErrorCode (int enum)

| Constant | gRPC Source |
|----------|------------|
| ErrorNotFound | codes.NotFound |
| ErrorAlreadyExists | codes.AlreadyExists |
| ErrorUnavailable | codes.Unavailable |
| ErrorPermissionDenied | codes.PermissionDenied |
| ErrorInvalidArgument | codes.InvalidArgument |
| ErrorDeadlineExceeded | codes.DeadlineExceeded |
| ErrorCancelled | codes.Canceled |
| ErrorInternal | codes.Internal (+ all unmapped) |

### Options Structs

| Struct | Fields | Used By |
|--------|--------|---------|
| CreateOptions | (empty, future extensibility) | Sandbox.Create, Provider.Create |
| GetOptions | (empty) | Sandbox.Get, Provider.Get |
| ListOptions | Limit int, Offset int, LabelSelector string | Sandbox.List, Provider.List |
| DeleteOptions | (empty) | Sandbox.Delete, Provider.Delete |
| UpdateOptions | (empty) | Provider.Update |
| WatchOptions | TimeoutSeconds int64, LabelSelector string | Sandbox.Watch |
| WaitOptions | (empty, uses context for timeout) | Sandbox.WaitReady |
| ExecOptions | Env map[string]string, WorkDir string | Exec.Run, Exec.Stream |

## Interfaces

### AuthProvider

| Method | Signature | Description |
|--------|-----------|-------------|
| GetRequestMetadata | (ctx, uri ...string) (map[string]string, error) | Returns auth headers |
| RequireTransportSecurity | () bool | Whether TLS is required |

Note: Implements `credentials.PerRPCCredentials` from gRPC. Two built-in implementations: `NoAuth()` and `StaticToken(token string)`.

### Logger

| Method | Signature | Description |
|--------|-----------|-------------|
| Debug | (msg string, keysAndValues ...any) | Debug-level output |
| Info | (msg string, keysAndValues ...any) | Info-level output |
| Error | (err error, msg string, keysAndValues ...any) | Error-level output |

Note: Compatible with `logr.Logger` and `slog.Logger` adapters.

## Proto-to-SDK Conversion Map

| Proto Package | Proto Type | SDK Type | Conversion Notes |
|--------------|-----------|----------|-----------------|
| datamodelv1 | ObjectMeta | (fields extracted) | name, id, created_at_ms, labels, resource_version |
| openshellv1 | Sandbox | Sandbox | Flatten metadata + status + spec.template |
| openshellv1 | SandboxSpec | SandboxSpec | Subset of fields relevant to creation |
| openshellv1 | SandboxPhase | SandboxPhase | Enum int -> string constant |
| openshellv1 | SandboxResponse | Sandbox | Unwrap response wrapper |
| datamodelv1 | Provider | Provider | Flatten metadata, exclude credentials |
| openshellv1 | ProviderResponse | Provider | Unwrap response wrapper |
| openshellv1 | ExecSandboxEvent | ExecChunk/ExecResult | Oneof dispatch |
| openshellv1 | SandboxStreamEvent | Event[Sandbox] | Filter to status events only |
| openshellv1 | HealthResponse | (error or nil) | Status check |
