# Brainstorm: Core SDK (Phase 1)

**Date:** 2026-06-27
**Status:** active
**Depends on:** [002-project-setup](002-project-setup.md), [003-proto-generation](003-proto-generation.md)

## Problem Framing

With project scaffolding and proto generation in place, the SDK needs its
core client library: the public API that Go consumers import and use. This
is Phase 1 from the original SDK brainstorm (001), covering the minimum
viable surface for cc-deck and basic operator use cases.

The primary design constraint is: **this SDK should feel like home for
Kubernetes developers.** The API shape, error handling, sub-client pattern,
options structs, and watch interfaces should follow Kubernetes client-go
conventions so operator authors have zero learning curve.

## Approaches Considered

### A: Kubernetes client-go Pattern (Chosen)

Model the SDK after `k8s.io/client-go`:
- Interface-based sub-clients per resource domain
- Config struct for connection setup
- Typed error helpers (`IsNotFound`, `IsAlreadyExists`)
- Options structs for CRUD operations
- Watch returning `watch.Interface`-compatible channels

- Pros: Zero surprise for the primary audience (K8s operator developers).
  Proven pattern at scale. Interface-based design enables easy mocking.
- Cons: More boilerplate than a minimal gRPC wrapper. Options structs are
  more verbose than functional options.

### B: Flat Client with Functional Options (from 001 draft)

All methods on `*Client`, functional options for configuration. More
typical for standalone Go gRPC SDKs (e.g., Google Cloud client libraries).

- Pros: Less code, simpler for small APIs, idiomatic for gRPC libraries.
- Cons: Unfamiliar to K8s developers. Harder to mock (no interfaces).
  Doesn't scale to 44+ RPCs without becoming a wall of methods.

### C: Code-Generated Client from Proto

Use `protoc-gen-go-grpc` stubs directly, with thin wrappers for
connection and auth. Minimal hand-written code.

- Pros: Minimal maintenance. API tracks protos automatically.
- Cons: Violates proto isolation. Exposes gRPC internals. No idiomatic
  Go types. Terrible developer experience.

## Decision

**Approach A: Kubernetes client-go pattern.** The SDK's primary consumers
are Go developers building Kubernetes operators and controllers. Following
client-go conventions gives them a familiar, well-tested API shape with
zero learning curve.

## Key Requirements

### Governing Principle

**"Feel like client-go."** When in doubt about an API design choice, check
how client-go handles the equivalent pattern and follow it. This applies to:
- Client construction and configuration
- Sub-client interfaces and implementations
- Error types and helper functions
- Options structs for operations
- Watch/informer patterns
- Context usage

### API Version Namespace

The public API lives under `openshell/v1/`:

```go
import "github.com/rhuss/openshell-sdk-go/openshell/v1"

client, err := v1.NewClient(config)
```

When upstream introduces `openshell.v2` protos in the future, a new
`openshell/v2/` package is added alongside `v1/` without breaking
existing consumers.

### Config Struct (Connection Setup)

```go
package v1

type Config struct {
    // Address is the gateway host:port.
    Address string

    // TLS configuration. Nil means auto-discover.
    TLS *TLSConfig

    // Auth provides credentials for gateway authentication.
    Auth AuthProvider

    // Timeout for unary RPCs. Zero means no timeout.
    Timeout time.Duration

    // RetryPolicy for transient failures. Nil means no retries.
    RetryPolicy *RetryPolicy
}

type TLSConfig struct {
    CertFile string
    KeyFile  string
    CAFile   string
    Insecure bool  // skip TLS (localhost dev only)
}
```

`NewClient(config)` returns `(*Client, error)`. Similar to how
`kubernetes.NewForConfig(restConfig)` works.

### Sub-Client Interfaces

Following client-go's `CoreV1Interface` / `PodsGetter` pattern:

```go
// ClientInterface is the top-level interface for the OpenShell v1 SDK.
type ClientInterface interface {
    Sandboxes() SandboxInterface
    Providers() ProviderInterface
    Exec() ExecInterface
    Files() FileInterface
    Health() HealthInterface
}

// SandboxInterface defines operations on sandboxes.
type SandboxInterface interface {
    Create(ctx context.Context, spec *SandboxSpec, opts CreateOptions) (*Sandbox, error)
    Get(ctx context.Context, name string, opts GetOptions) (*Sandbox, error)
    List(ctx context.Context, opts ListOptions) (*SandboxList, error)
    Delete(ctx context.Context, name string, opts DeleteOptions) error
    Watch(ctx context.Context, opts WatchOptions) (WatchInterface[Sandbox], error)
    WaitReady(ctx context.Context, name string, opts WaitOptions) (*Sandbox, error)
    AttachProvider(ctx context.Context, sandbox, provider string) error
    DetachProvider(ctx context.Context, sandbox, provider string) error
    ListProviders(ctx context.Context, sandbox string) ([]string, error)
}

// ProviderInterface defines operations on providers.
type ProviderInterface interface {
    Create(ctx context.Context, provider *Provider, opts CreateOptions) (*Provider, error)
    Get(ctx context.Context, name string, opts GetOptions) (*Provider, error)
    List(ctx context.Context, opts ListOptions) (*ProviderList, error)
    Update(ctx context.Context, provider *Provider, opts UpdateOptions) (*Provider, error)
    Delete(ctx context.Context, name string, opts DeleteOptions) error
    Ensure(ctx context.Context, provider *Provider) (*Provider, error)
}

// ExecInterface defines command execution operations.
type ExecInterface interface {
    Run(ctx context.Context, sandbox string, cmd []string, opts ExecOptions) (*ExecResult, error)
    Stream(ctx context.Context, sandbox string, cmd []string, opts ExecOptions) (*ExecStream, error)
    Interactive(ctx context.Context, sandbox string, cmd []string) (*InteractiveSession, error)
}

// FileInterface defines file transfer operations.
type FileInterface interface {
    Upload(ctx context.Context, sandbox string, localPath, remotePath string) error
    Download(ctx context.Context, sandbox string, remotePath, localPath string) error
}

// HealthInterface defines health check operations.
type HealthInterface interface {
    Check(ctx context.Context) error
}
```

### Typed Errors

Following `apimachinery/pkg/api/errors` pattern:

```go
package v1

type StatusError struct {
    Code    ErrorCode
    Message string
    Details map[string]string
}

type ErrorCode int

const (
    ErrorNotFound       ErrorCode = iota
    ErrorAlreadyExists
    ErrorUnavailable
    ErrorPermissionDenied
    ErrorInvalidArgument
    ErrorDeadlineExceeded
    ErrorInternal
)

func IsNotFound(err error) bool
func IsAlreadyExists(err error) bool
func IsUnavailable(err error) bool
func IsPermissionDenied(err error) bool
```

gRPC status codes are mapped to SDK error codes internally. Consumers
never need to import `google.golang.org/grpc/status`.

### Domain Types

The SDK defines its own types, not proto-generated types:

```go
type Sandbox struct {
    Name      string
    Phase     SandboxPhase
    Image     string
    CreatedAt time.Time
    // ... fields relevant to consumers, not all proto fields
}

type SandboxSpec struct {
    Image    string
    Policy   *PolicyRef
    Env      map[string]string
    // ... creation-time fields
}

type SandboxPhase string

const (
    SandboxPending  SandboxPhase = "Pending"
    SandboxStarting SandboxPhase = "Starting"
    SandboxReady    SandboxPhase = "Ready"
    SandboxStopped  SandboxPhase = "Stopped"
    SandboxFailed   SandboxPhase = "Failed"
)
```

Conversion between SDK types and proto types happens in internal converter
functions, never exposed to consumers.

### Watch Interface

Generic watch pattern compatible with controller-runtime:

```go
type WatchInterface[T any] interface {
    ResultChan() <-chan Event[T]
    Stop()
}

type Event[T any] struct {
    Type   EventType
    Object T
}

type EventType string

const (
    EventAdded    EventType = "ADDED"
    EventModified EventType = "MODIFIED"
    EventDeleted  EventType = "DELETED"
    EventError    EventType = "ERROR"
)
```

### Exec Stream Types

```go
type ExecResult struct {
    ExitCode int
    Stdout   []byte
    Stderr   []byte
}

type ExecStream struct {
    // Next returns the next chunk of output.
    // Returns io.EOF when the command completes.
    Next() (*ExecChunk, error)
    // ExitCode blocks until completion and returns the exit code.
    ExitCode() (int, error)
    Close() error
}

type ExecChunk struct {
    Stream StreamType  // Stdout or Stderr
    Data   []byte
}
```

### Phase 1 Scope

| Sub-client | RPCs covered | Notes |
|------------|-------------|-------|
| Sandboxes | Create, Get, List, Delete, Watch, AttachProvider, DetachProvider, ListSandboxProviders | Core lifecycle |
| Providers | Create, Get, List, Update, Delete | + Ensure (idempotent create-or-update) |
| Exec | ExecSandbox, ExecSandboxInteractive | Run, Stream, Interactive |
| Files | (via SSH tunnel) | Upload, Download convenience methods |
| Health | Health | Simple health check |

**Total: ~20 RPCs wrapped, 5 sub-clients.**

### Out of Scope for Phase 1

- Service exposure (ExposeService, GetService, etc.)
- Provider profiles and credential refresh
- Policy management
- Configuration management
- SSH session management (raw API; file transfer uses it internally)
- TCP forwarding
- Logging RPCs
- Supervisor/internal RPCs

### Testing Strategy

- **Unit tests:** Each sub-client tested against a mock gRPC server
  (in-process, no network). Mock server implements the proto service
  interface with canned responses.
- **Integration tests:** `//go:build integration` tag. Require a running
  gateway. Test real gRPC calls, TLS, auth.
- **Interface mocks:** The `ClientInterface` and sub-client interfaces
  enable consumers to mock the SDK in their own tests without needing
  a gateway.

### Package Layout

```
openshell/
  v1/
    client.go           # Client, NewClient, Config
    sandbox.go          # SandboxInterface, Sandbox, SandboxSpec
    provider.go         # ProviderInterface, Provider
    exec.go             # ExecInterface, ExecResult, ExecStream
    file.go             # FileInterface
    health.go           # HealthInterface
    errors.go           # StatusError, IsNotFound, etc.
    types.go            # Shared types (SandboxPhase, EventType, etc.)
    options.go          # CreateOptions, ListOptions, WatchOptions, etc.
    watch.go            # WatchInterface, Event
    internal/
      converter/        # Proto <-> SDK type conversions
      grpc/             # gRPC connection management
```

## Open Questions

- Should the SDK provide a `fake` package (like `client-go/kubernetes/fake`)
  for consumer testing, or is interface-based mocking sufficient?
- How should file transfer handle large files? Streaming tar or chunked
  upload? (Implementation detail for the spec phase.)
- Should `Ensure` (idempotent create-or-update) be on every sub-client or
  just Provider? (cc-deck uses it heavily for providers.)
- Should the SDK support multiple gateway connections (multi-cluster) or
  is one client per gateway sufficient?
