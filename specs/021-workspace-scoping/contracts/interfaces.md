# Interface Contracts: Workspace Scoping

## Signature Change Pattern

Every workspace-scoped method gains `workspace string` as the first parameter after `ctx context.Context`.

### SandboxInterface

```go
type SandboxInterface interface {
    Create(ctx context.Context, workspace, name string, spec *SandboxSpec, labels map[string]string) (*Sandbox, error)
    Get(ctx context.Context, workspace, name string) (*Sandbox, error)
    List(ctx context.Context, workspace string, opts ...ListOptions) ([]*Sandbox, error)
    Delete(ctx context.Context, workspace, name string) error
    AttachProvider(ctx context.Context, workspace, sandboxName, providerName string, expectedResourceVersion uint64) (*AttachProviderResult, error)
    DetachProvider(ctx context.Context, workspace, sandboxName, providerName string, expectedResourceVersion uint64) (*DetachProviderResult, error)
    ListProviders(ctx context.Context, workspace, sandboxName string) ([]*Provider, error)
    WaitReady(ctx context.Context, workspace, name string, opts ...WaitOptions) (*Sandbox, error)
    Watch(ctx context.Context, workspace, name string, opts ...WatchOptions) (WatchInterface[*Sandbox], error)
    GetLogs(ctx context.Context, workspace, sandboxName string, opts ...LogOption) (*LogResult, error)
}
```

### ProviderInterface

```go
type ProviderInterface interface {
    Create(ctx context.Context, workspace string, provider *Provider) (*Provider, error)
    Get(ctx context.Context, workspace, name string) (*Provider, error)
    List(ctx context.Context, workspace string, opts ...ListOptions) ([]*Provider, error)
    Update(ctx context.Context, workspace string, provider *Provider) (*Provider, error)
    Delete(ctx context.Context, workspace, name string) error
    Ensure(ctx context.Context, workspace string, provider *Provider) (*Provider, error)
    Profiles() ProfileInterface
    Refresh() RefreshInterface
}
```

### ProfileInterface

```go
type ProfileInterface interface {
    List(ctx context.Context, workspace string, opts ...ListOptions) ([]*ProviderProfile, error)
    Get(ctx context.Context, workspace, id string) (*ProviderProfile, error)
    Import(ctx context.Context, workspace string, items []ProfileImportItem) (*ImportResult, error)
    Update(ctx context.Context, workspace, id string, expectedResourceVersion uint64, item ProfileImportItem) (*UpdateResult, error)
    Lint(ctx context.Context, workspace string, items []ProfileImportItem) (*LintResult, error)
    Delete(ctx context.Context, workspace, id string) (bool, error)
}
```

### RefreshInterface

```go
type RefreshInterface interface {
    GetStatus(ctx context.Context, workspace, provider, credentialKey string) ([]*RefreshStatus, error)
    Configure(ctx context.Context, workspace string, config *RefreshConfig) (*RefreshStatus, error)
    Rotate(ctx context.Context, workspace, provider, credentialKey string) (*RefreshStatus, error)
    Delete(ctx context.Context, workspace, provider, credentialKey string) (bool, error)
}
```

### ServiceInterface

```go
type ServiceInterface interface {
    Expose(ctx context.Context, workspace, sandboxName, serviceName string, targetPort uint32, domain bool) (*ServiceEndpoint, error)
    Get(ctx context.Context, workspace, sandboxName, serviceName string) (*ServiceEndpoint, error)
    List(ctx context.Context, workspace, sandboxName string, opts ...ListOptions) ([]*ServiceEndpoint, error)
    Delete(ctx context.Context, workspace, sandboxName, serviceName string) error
}
```

### ExecInterface

```go
type ExecInterface interface {
    Run(ctx context.Context, workspace, sandboxName string, command []string, opts ...ExecOptions) (*ExecResult, error)
    Stream(ctx context.Context, workspace, sandboxName string, command []string, opts ...ExecOptions) (ExecStream, error)
    Interactive(ctx context.Context, workspace, sandboxName string, command []string, cols, rows uint32, opts ...ExecOptions) (InteractiveSession, error)
}
```

### FileInterface

```go
type FileInterface interface {
    Upload(ctx context.Context, workspace, sandboxName string, localPath string, remotePath string) error
    Download(ctx context.Context, workspace, sandboxName string, remotePath string, localPath string) error
}
```

### ConfigInterface

```go
type ConfigInterface interface {
    GetSandbox(ctx context.Context, workspace, sandboxName string) (*SandboxConfig, error)
    GetGateway(ctx context.Context) (*GatewayConfig, error)  // unchanged - gateway-scoped
    Update(ctx context.Context, workspace string, update *ConfigUpdate) (*ConfigUpdateResult, error)
}
```

### PolicyInterface

```go
type PolicyInterface interface {
    GetDraft(ctx context.Context, workspace, sandboxName string, opts ...GetDraftOption) (*DraftPolicy, error)
    ApproveDraftChunk(ctx context.Context, workspace, sandboxName, chunkID string) (*DraftChunk, error)
    RejectDraftChunk(ctx context.Context, workspace, sandboxName, chunkID string) (*DraftChunk, error)
    ApproveAllDraftChunks(ctx context.Context, workspace, sandboxName string) ([]*DraftChunk, error)
    ClearDraftChunks(ctx context.Context, workspace, sandboxName string) error
    GetDraftHistory(ctx context.Context, workspace, sandboxName string) ([]*DraftChunk, error)
    GetStatus(ctx context.Context, workspace, sandboxName string) (*PolicyStatus, error)
    List(ctx context.Context, workspace string, opts ...ListOptions) ([]*PolicySummary, error)
    EditDraftChunk(ctx context.Context, workspace, sandboxName, chunkID string, edit *DraftChunkEdit) (*DraftChunk, error)
    UndoDraftChunk(ctx context.Context, workspace, sandboxName, chunkID string) (*DraftChunk, error)
}
```

### SSHInterface

```go
type SSHInterface interface {
    CreateSession(ctx context.Context, workspace, sandboxID string) (*SSHSession, error)
    RevokeSession(ctx context.Context, workspace, token string) (bool, error)
    Tunnel(ctx context.Context, workspace, sandboxName string, port uint32, opts ...TunnelOption) (*SSHTunnel, error)
}
```

### TCPInterface

```go
type TCPInterface interface {
    Forward(ctx context.Context, workspace, sandboxName string, port uint32, opts ...ForwardOption) (io.ReadWriteCloser, error)
}
```

### Unchanged Interfaces

```go
type HealthInterface interface {
    Check(ctx context.Context) (*HealthResult, error)  // gateway-scoped, no workspace
}
```

## ListOptions

```go
type ListOptions struct {
    Limit         int
    Offset        int
    LabelSelector string
    AllWorkspaces bool  // NEW: list across all workspaces
}
```
