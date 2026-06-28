# Research: Core SDK (Phase 1)

## R1: Client-Go Sub-Client Pattern

**Decision**: Follow the Kubernetes client-go pattern with interface-based sub-clients.

**Rationale**: The SDK's primary audience is Kubernetes operator developers. The client-go pattern (top-level ClientInterface with typed sub-client accessors like `Sandboxes()`, `Providers()`) is immediately recognizable. Each sub-client is an interface, enabling consumers to mock individual domains without a full client.

**Alternatives considered**:
- Flat client with all methods: Scales poorly with 20+ RPCs, hard to mock selectively.
- Code-generated client: Exposes proto internals, poor developer experience.

## R2: Proto-to-SDK Type Conversion Strategy

**Decision**: Internal `converter` package maps between proto types and SDK domain types. All conversion functions are unexported.

**Rationale**: Proto types use `int64` timestamps, nested `ObjectMeta`, and proto-specific patterns (oneof, repeated). SDK types use `time.Time`, flat structs, and Go idioms. The conversion layer absorbs upstream proto changes without affecting the public API.

**Key mappings**:
- `openshell.datamodel.v1.ObjectMeta` -> extracted fields (Name, ID, CreatedAt, Labels, ResourceVersion)
- `openshell.v1.SandboxPhase` enum -> `SandboxPhase` string type with constants
- `int64` timestamps (ms since epoch) -> `time.Time`
- `map<string,string>` credentials -> excluded from SDK Provider type (sensitive, not needed by consumers)

**Alternatives considered**:
- Exposing proto types directly: Violates constitution (Proto Isolation principle).
- Type aliases to proto types: Still leaks proto imports into consumer code.

## R3: gRPC Error Mapping

**Decision**: Map gRPC status codes to SDK ErrorCode constants. Each StatusError carries a code, message, and optional details map.

**Mapping table**:

| gRPC Code | SDK ErrorCode | Helper Function |
|-----------|--------------|-----------------|
| NotFound | ErrorNotFound | IsNotFound() |
| AlreadyExists | ErrorAlreadyExists | IsAlreadyExists() |
| Unavailable | ErrorUnavailable | IsUnavailable() |
| PermissionDenied | ErrorPermissionDenied | IsPermissionDenied() |
| InvalidArgument | ErrorInvalidArgument | IsInvalidArgument() |
| DeadlineExceeded | ErrorDeadlineExceeded | IsDeadlineExceeded() |
| Cancelled | ErrorCancelled | IsCancelled() |
| Internal | ErrorInternal | (no helper, catch-all) |

**Rationale**: Consumers should never import `google.golang.org/grpc/status` or `google.golang.org/grpc/codes`. The helper functions match the `apimachinery/pkg/api/errors` pattern (IsNotFound, etc.).

## R4: Watch Implementation

**Decision**: Use the `WatchSandbox` server-streaming RPC. The SDK wraps the gRPC stream in a `WatchInterface[Sandbox]` that delivers typed events via a Go channel.

**Proto mapping**:
- `WatchSandboxRequest` -> `WatchOptions` (timeout, label selector)
- `SandboxStreamEvent` -> dispatched to `Event[Sandbox]` on the result channel
- Event types: status snapshot -> MODIFIED, sandbox deleted -> DELETED
- Stream errors -> ERROR event on channel, then channel close

**Rationale**: The channel-based pattern matches `k8s.io/apimachinery/pkg/watch.Interface`. Operators can select on the channel alongside other event sources.

**Key detail**: The proto `WatchSandbox` returns `SandboxStreamEvent` which is a oneof of sandbox status, log lines, platform events, and warnings. The SDK filters to status snapshots and maps them to watch events, discarding log/platform events (those are Phase 2 concerns).

## R5: Exec Streaming Architecture

**Decision**: Three execution modes mapping to two proto RPCs.

| SDK Method | Proto RPC | Pattern |
|-----------|-----------|---------|
| `Run()` | `ExecSandbox` | Collect all `ExecSandboxEvent` into `ExecResult` |
| `Stream()` | `ExecSandbox` | Yield each `ExecSandboxEvent` as `ExecChunk` |
| `Interactive()` | `ExecSandboxInteractive` | Bidirectional stream |

**Proto events**:
- `ExecSandboxEvent` is oneof: stdout (bytes), stderr (bytes), exit (code)
- `ExecSandboxInput` is oneof: start (command+sandbox), stdin (bytes), resize (cols+rows)

**Rationale**: Run and Stream use the same unary-request/server-stream RPC but differ in how the client consumes events. Interactive uses the bidirectional streaming RPC.

## R6: File Transfer via SSH

**Decision**: File transfer is implemented via the SSH session mechanism (CreateSshSession + SFTP over the session). The SDK creates a temporary SSH session, transfers the file, then cleans up.

**Rationale**: The proto API does not have dedicated file transfer RPCs. The upstream CLI uses SSH sessions for file operations. The SDK wraps this in a simple Upload/Download API.

**Implementation note**: This means the File sub-client internally depends on the SSH session RPC (CreateSshSession), even though SSH session management is out of scope for Phase 1's public API. The internal SSH usage is hidden from consumers.

## R7: Connection Management

**Decision**: NewClient establishes a gRPC connection eagerly using `grpc.NewClient` (the modern API, not the deprecated `grpc.Dial`). The connection is shared across all sub-clients.

**Key details**:
- `grpc.NewClient` with `grpc.WithTransportCredentials` for TLS
- `grpc.WithPerRPCCredentials` for auth token injection
- Connection is stored in the Client struct and shared by all sub-clients
- `Close()` closes the underlying `grpc.ClientConn`
- Thread safety: `grpc.ClientConn` is already goroutine-safe, so the Client is inherently safe for concurrent use

## R8: Ensure (Idempotent Create-or-Update)

**Decision**: `Ensure` is a client-side convenience that combines Get + Create/Update. It is not a single proto RPC.

**Logic**:
1. Call GetProvider
2. If NotFound: call CreateProvider
3. If found and differs: call UpdateProvider
4. If found and identical: return existing

**Rationale**: The gateway does not have an upsert RPC. The operator pattern needs idempotent setup, and the SDK absorbs the get-check-create/update boilerplate.

## R9: Package Layout Decision

**Decision**: Public API in `openshell/v1/`, internal packages for gRPC and conversion.

```
openshell/
  v1/
    client.go           # Client, NewClient, Config, ClientInterface
    sandbox.go          # SandboxInterface, types
    provider.go         # ProviderInterface, types
    exec.go             # ExecInterface, types
    file.go             # FileInterface
    health.go           # HealthInterface
    errors.go           # StatusError, Is* helpers
    types.go            # Shared types (phases, event types)
    options.go          # Options structs
    watch.go            # WatchInterface, Event
    doc.go              # Package documentation
    internal/
      grpc/             # gRPC connection, client wrappers
      converter/        # Proto <-> SDK type conversions
```

**Rationale**: The `v1` package enables future `v2` coexistence. Internal packages prevent consumers from depending on conversion or gRPC details.

## R10: Existing Code Integration

**Decision**: Replace the current stub `openshell/client.go` (which has `Dial` and empty `Client`) with the new `openshell/v1/` package. The existing `openshell/` package becomes a re-export shim or is removed.

**Current state**:
- `openshell/client.go`: Stub Client with Dial(address) and Close()
- `openshell/client_test.go`: Basic tests for Dial and Close
- `proto/`: Generated proto bindings (openshellv1, datamodelv1, sandboxv1)

**Migration**: The stub was scaffolding from spec 001. The new v1 package replaces it entirely. The old Dial/Close tests are superseded by the new NewClient tests.
