# OpenShell Go SDK

## Problem

OpenShell ships a Python SDK (`python/openshell/sandbox.py`) and a Rust CLI, but no Go SDK.
Go consumers (cc-deck, potential Kubernetes operators, CI/CD integrations) must either:

1. Wrap the CLI binary (fragile, version coupling, loses structured errors)
2. Vendor protos and build their own client (what cc-deck did, leading to proto drift and API misalignment)
3. Use the Python SDK via subprocess (defeats the purpose of Go)

A first-party Go SDK would give Go consumers a stable, versioned, idiomatic client that tracks the gateway API as it evolves.

## Consumers

### cc-deck (TUI/CLI for OpenShell workspaces)

cc-deck manages OpenShell workspaces from the terminal. It needs:

- Sandbox lifecycle (create, get, list, delete, wait-for-ready)
- Command execution (streaming exec with stdout/stderr separation)
- Provider management (CRUD, ensure-idempotent, credential refresh)
- File transfer (upload/download to sandbox filesystem)
- Interactive attach (bidirectional PTY for terminal sessions)
- Connection management (mTLS auto-discovery, insecure localhost fallback)

cc-deck currently wraps the CLI binary. The gRPC client prototype (branch `074-openshell-grpc-client`) uncovered multiple alignment issues: proto version drift, sandbox ID vs. name confusion, reserved env var prefixes, phase enum mismatches.

### Kubernetes Operator (NVIDIA/OpenShell#1719)

A Kubernetes operator for OpenShell would need:

- **Deployment operator**: Gateway installation, upgrades, configuration, TLS/ingress wiring, health reconciliation
- **Sandbox CRD operator**: Declarative sandbox lifecycle, policy attachment, provider wiring, status conditions
- **Gateway API client**: All 55 RPCs for full control plane access from the reconciler

The operator needs the Go SDK as a library dependency, not a CLI wrapper.
The SDK must support long-lived connections, reconnection, and context cancellation for controller-runtime compatibility.

### CI/CD Integrations

Automated pipelines that provision sandboxes for testing:

- Create sandbox with specific image/policy
- Run test commands
- Collect results
- Tear down sandbox

These need a simple, stable API surface.

## Gateway API Surface (55 RPCs)

Grouped by domain:

### Health (1 RPC)

| RPC | Type | Description |
|-----|------|-------------|
| `Health` | Unary | Gateway health check |

### Sandbox Lifecycle (8 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `CreateSandbox` | Unary | Create a new sandbox from a spec |
| `GetSandbox` | Unary | Get sandbox by name |
| `ListSandboxes` | Unary | List all sandboxes |
| `DeleteSandbox` | Unary | Delete a sandbox |
| `WatchSandbox` | Server-stream | Watch sandbox state changes |
| `ListSandboxProviders` | Unary | List providers attached to a sandbox |
| `AttachSandboxProvider` | Unary | Attach a provider to a running sandbox |
| `DetachSandboxProvider` | Unary | Detach a provider from a sandbox |

### Command Execution (2 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `ExecSandbox` | Server-stream | Execute command, stream stdout/stderr/exit |
| `ExecSandboxInteractive` | Bidi-stream | Interactive PTY exec (stdin + stdout/stderr) |

### SSH/Tunnel (3 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `CreateSshSession` | Unary | Get SSH tunnel credentials |
| `RevokeSshSession` | Unary | Revoke an SSH session |
| `ForwardTcp` | Bidi-stream | TCP port forwarding through gateway |

### Service Exposure (4 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `ExposeService` | Unary | Expose a sandbox port as a service |
| `GetService` | Unary | Get service endpoint info |
| `ListServices` | Unary | List exposed services |
| `DeleteService` | Unary | Remove a service exposure |

### Provider Management (12 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `CreateProvider` | Unary | Create a provider |
| `GetProvider` | Unary | Get provider by name |
| `ListProviders` | Unary | List all providers |
| `UpdateProvider` | Unary | Update provider credentials/config |
| `DeleteProvider` | Unary | Delete a provider |
| `ListProviderProfiles` | Unary | List provider profiles |
| `GetProviderProfile` | Unary | Get a specific profile |
| `ImportProviderProfiles` | Unary | Import profiles |
| `UpdateProviderProfiles` | Unary | Update profiles |
| `LintProviderProfiles` | Unary | Validate profiles |
| `DeleteProviderProfile` | Unary | Delete a profile |
| `GetProviderRefreshStatus` | Unary | Get credential refresh status |
| `ConfigureProviderRefresh` | Unary | Configure automatic refresh |
| `RotateProviderCredential` | Unary | Trigger credential rotation |
| `DeleteProviderRefresh` | Unary | Remove refresh configuration |

### Configuration (3 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `GetSandboxConfig` | Unary | Get sandbox configuration |
| `GetGatewayConfig` | Unary | Get gateway configuration |
| `UpdateConfig` | Unary | Update configuration |

### Policy Management (7 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `GetSandboxPolicyStatus` | Unary | Get policy enforcement status |
| `ListSandboxPolicies` | Unary | List policies |
| `ReportPolicyStatus` | Unary | Report policy status (from supervisor) |
| `SubmitPolicyAnalysis` | Unary | Submit policy analysis results |
| `GetDraftPolicy` | Unary | Get draft policy for review |
| `ApproveDraftChunk` | Unary | Approve a policy draft chunk |
| `RejectDraftChunk` | Unary | Reject a policy draft chunk |
| `ApproveAllDraftChunks` | Unary | Approve all draft chunks |
| `EditDraftChunk` | Unary | Edit a draft chunk |
| `UndoDraftChunk` | Unary | Undo a draft chunk edit |
| `ClearDraftChunks` | Unary | Clear all draft chunks |
| `GetDraftHistory` | Unary | Get draft edit history |

### Sandbox Environment (2 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `GetSandboxProviderEnvironment` | Unary | Get provider env vars for a sandbox |
| `IssueSandboxToken` | Unary | Issue auth token for sandbox |
| `RefreshSandboxToken` | Unary | Refresh sandbox auth token |

### Logging (2 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `GetSandboxLogs` | Unary | Get sandbox logs |
| `PushSandboxLogs` | Client-stream | Push logs from supervisor |

### Internal/Supervisor (2 RPCs)

| RPC | Type | Description |
|-----|------|-------------|
| `ConnectSupervisor` | Bidi-stream | Supervisor-gateway control channel |
| `RelayStream` | Bidi-stream | Raw byte relay for SSH/exec |

## SDK Design

### Module Structure

```
github.com/NVIDIA/openshell-sdk-go/
  openshell/          # Public API package
    client.go         # Client, ClientConfig, Dial()
    sandbox.go        # Sandbox, SandboxSpec, SandboxRef
    exec.go           # ExecResult, ExecStream, InteractiveSession
    provider.go       # Provider, ProviderProfile, ProviderRefresh
    service.go        # ServiceEndpoint
    policy.go         # Policy, DraftPolicy
    tunnel.go         # SSHTunnel, TcpForward (higher-level wrappers)
    errors.go         # Typed errors (NotFound, AlreadyExists, etc.)
    option.go         # Functional options (WithTLS, WithInsecure, etc.)
  proto/              # Generated proto code (internal or vendored)
    openshellv1/
    datamodelv1/
    sandboxv1/
```

### Client API (Draft)

```go
package openshell

// Dial connects to an OpenShell gateway.
// Options control TLS, auth, timeouts, and retry behavior.
func Dial(addr string, opts ...Option) (*Client, error)

// Client wraps the OpenShell gateway gRPC connection.
type Client struct { ... }
func (c *Client) Close() error

// Sandbox operations
func (c *Client) CreateSandbox(ctx context.Context, spec *SandboxSpec) (*SandboxRef, error)
func (c *Client) GetSandbox(ctx context.Context, name string) (*SandboxRef, error)
func (c *Client) ListSandboxes(ctx context.Context) ([]*SandboxRef, error)
func (c *Client) DeleteSandbox(ctx context.Context, name string) error
func (c *Client) WaitReady(ctx context.Context, name string) (*SandboxRef, error)
func (c *Client) WatchSandbox(ctx context.Context, name string) (<-chan SandboxEvent, error)

// Sandbox provider attachment
func (c *Client) AttachProvider(ctx context.Context, sandbox, provider string) error
func (c *Client) DetachProvider(ctx context.Context, sandbox, provider string) error
func (c *Client) ListSandboxProviders(ctx context.Context, sandbox string) ([]string, error)

// Command execution
func (c *Client) Exec(ctx context.Context, sandboxID string, cmd []string, opts ...ExecOption) (*ExecResult, error)
func (c *Client) ExecStream(ctx context.Context, sandboxID string, cmd []string, opts ...ExecOption) (*ExecStream, error)
func (c *Client) ExecInteractive(ctx context.Context, sandboxID string, cmd []string) (*InteractiveSession, error)

// Provider management
func (c *Client) CreateProvider(ctx context.Context, p *Provider) error
func (c *Client) GetProvider(ctx context.Context, name string) (*Provider, error)
func (c *Client) ListProviders(ctx context.Context) ([]*Provider, error)
func (c *Client) UpdateProvider(ctx context.Context, p *Provider) error
func (c *Client) DeleteProvider(ctx context.Context, name string) error
func (c *Client) EnsureProvider(ctx context.Context, p *Provider) error

// Provider profiles
func (c *Client) ListProviderProfiles(ctx context.Context) ([]*ProviderProfile, error)
func (c *Client) ImportProviderProfiles(ctx context.Context, profiles []*ProviderProfile) error

// Service exposure
func (c *Client) ExposeService(ctx context.Context, sandbox string, port int, opts ...ServiceOption) (*ServiceEndpoint, error)
func (c *Client) GetService(ctx context.Context, name string) (*ServiceEndpoint, error)
func (c *Client) ListServices(ctx context.Context, sandbox string) ([]*ServiceEndpoint, error)
func (c *Client) DeleteService(ctx context.Context, name string) error

// SSH/Tunnel
func (c *Client) CreateSSHSession(ctx context.Context, sandboxID string) (*SSHSession, error)
func (c *Client) ForwardTCP(ctx context.Context, sandboxID string, remotePort int) (net.Conn, error)

// File transfer (convenience wrappers over SSH tunnel)
func (c *Client) Upload(ctx context.Context, sandboxID, localPath, remotePath string) error
func (c *Client) Download(ctx context.Context, sandboxID, remotePath, localPath string) error

// Policy
func (c *Client) GetPolicyStatus(ctx context.Context, sandbox string) (*PolicyStatus, error)
func (c *Client) ListPolicies(ctx context.Context) ([]*Policy, error)
func (c *Client) GetDraftPolicy(ctx context.Context) (*DraftPolicy, error)

// Configuration
func (c *Client) GetGatewayConfig(ctx context.Context) (*GatewayConfig, error)

// Health
func (c *Client) Health(ctx context.Context) error
```

### Connection Options

```go
// TLS auto-discovery (matches CLI behavior)
func WithAutoTLS() Option

// Explicit mTLS certificates
func WithTLS(certPath, keyPath, caPath string) Option

// Insecure (for localhost development)
func WithInsecure() Option

// Bearer token auth (for OIDC/OAuth flows)
func WithBearerToken(tokenProvider func() string) Option

// Timeouts
func WithTimeout(d time.Duration) Option

// Retry policy
func WithRetry(maxAttempts int, backoff time.Duration) Option

// For controller-runtime: reconnect on disconnect
func WithReconnect() Option
```

### Key Design Decisions

1. **SandboxRef vs. Sandbox name**: The SDK should handle the ID-vs-name distinction internally. Users pass names, the SDK resolves IDs when needed for RPCs that require them.

2. **Proto isolation**: Generated proto types should NOT leak through the public API. The SDK defines its own domain types and converts internally. This decouples consumers from proto version changes.

3. **Streaming exec**: `ExecStream` returns a typed stream with `Next()` that yields `ExecChunk` (stdout/stderr/exit), not raw proto events.

4. **File transfer**: Upload/Download are convenience methods built on SSH tunnel + tar, matching the CLI approach. They are not part of the proto API.

5. **Kubernetes operator compatibility**: The SDK must work with controller-runtime's reconciler pattern: context-aware, cancellable, idempotent operations. `WatchSandbox` returns a channel for informer-like patterns.

## Kubernetes Operator Considerations

Based on [NVIDIA/OpenShell#1719](https://github.com/NVIDIA/OpenShell/issues/1719):

### Operator as Gateway API Client

The Go SDK enables writing a Kubernetes operator that:

1. **Manages sandbox lifecycle** through CRDs that map to gateway API calls
2. **Reconciles provider state** between K8s Secrets and gateway providers
3. **Exposes sandbox status** via CRD `.status` conditions
4. **Handles policy** through K8s-native policy CRDs or ConfigMaps

### CRD Shape (Sketch)

```yaml
apiVersion: openshell.nvidia.com/v1alpha1
kind: Sandbox
metadata:
  name: my-sandbox
spec:
  image: my-app:latest
  policy:
    configMapRef:
      name: my-policy
  providers:
    - name: vertex-ai
      secretRef:
        name: vertex-credentials
  resources:
    gpu: true
status:
  phase: Ready
  sandboxID: "abc-123"
  gatewayAddress: "gateway.openshell.svc:17670"
  conditions:
    - type: Ready
      status: "True"
    - type: ProvidersAttached
      status: "True"
```

### SDK Methods the Operator Needs

| Operator Responsibility | SDK Methods Used |
|------------------------|-----------------|
| Create sandbox from CRD | `CreateSandbox`, `WaitReady` |
| Attach providers | `AttachProvider`, `EnsureProvider` |
| Monitor sandbox health | `WatchSandbox`, `GetSandbox` |
| Cleanup on CRD deletion | `DeleteSandbox`, `DeleteProvider` |
| Policy management | `GetPolicyStatus`, `ListPolicies` |
| Service exposure | `ExposeService`, `DeleteService` |
| Gateway deployment | `Health`, `GetGatewayConfig` |

### Gateway vs. Operator Boundary

The SDK approach naturally clarifies the boundary question from issue #1719:

- **Gateway owns**: Sandbox runtime, exec routing, SSH relay, credential injection, policy enforcement
- **Operator owns**: K8s-native lifecycle (CRD reconciliation, status conditions, finalizers), K8s secret-to-provider sync, deployment/upgrade of the gateway itself
- **SDK bridges**: The operator talks to the gateway through the SDK, never directly to sandbox pods

## Implementation Phases

### Phase 1: Core SDK (cc-deck unblocking)

- Connection management (Dial, TLS, auth)
- Sandbox CRUD + WaitReady
- Exec (streaming, non-streaming)
- Provider CRUD + EnsureProvider
- File transfer (upload/download via SSH tunnel)
- Updated proto from upstream source

### Phase 2: Full API Coverage

- Service exposure
- Provider profiles and credential refresh
- Interactive exec (bidi streaming)
- TCP forwarding
- Policy management
- Watch/stream operations

### Phase 3: Operator Support

- controller-runtime integration patterns
- Reconnection and retry policies
- Informer-compatible watch channels
- CRD type definitions
- Example operator scaffold

## cc-deck Thin Wrapper (Interim)

While the full SDK is being developed, cc-deck uses a thin wrapper package (`internal/openshell/sdk/`) that:

1. Provides the same API surface as Phase 1 of the full SDK
2. Is isolated in its own package so it can be swapped for the real SDK later
3. Uses the upstream protos (copied from OpenShell source)
4. Handles ID-vs-name resolution internally
5. Implements TLS auto-discovery matching the CLI/Python SDK behavior

When the full SDK ships, cc-deck replaces `internal/openshell/sdk/` with `github.com/NVIDIA/openshell-sdk-go/openshell` and removes the vendored protos.

## Open Questions

1. **Repo ownership**: Should this live under `NVIDIA/openshell-sdk-go` or as a subdirectory in the main OpenShell repo?
2. **Proto publishing**: Should the Go proto bindings be published as a separate module (`openshell-api-go`) or bundled with the SDK?
3. **Versioning**: Should the SDK version track gateway versions, or have independent semver?
4. **Auth patterns**: What OIDC/OAuth flows should the SDK support out of the box? The Python SDK has `from_active_cluster` with OIDC refresh.
5. **File transfer**: Should file transfer be part of the SDK, or remain CLI-only? cc-deck needs it, but the Python SDK skips it.
6. **Testing**: Should the SDK ship a mock server (like testcontainers) for consumer integration tests?
