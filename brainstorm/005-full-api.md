# Brainstorm: Full API Coverage (Phase 2)

**Date:** 2026-06-27
**Status:** active
**Depends on:** [004-core-sdk](004-core-sdk.md) (Phase 1 must ship first)

## Problem Framing

Phase 1 covers the minimum viable SDK (sandbox, provider, exec, files,
health). The remaining ~24 RPCs from the OpenShell gateway API need to be
added to reach full coverage. These RPCs cover service exposure, provider
profiles, credential refresh, policy management, configuration, SSH
sessions, and TCP forwarding.

Not all consumers need all of these. Service exposure and provider profiles
are useful for operators. Policy management is useful for governance tooling.
SSH and TCP forwarding are advanced capabilities for direct sandbox access.

## Approaches Considered

### A: Two Increments (Chosen)

Split Phase 2 into two sub-phases based on consumer need:
- **Phase 2a:** Service exposure + provider profiles + credential refresh
  (operator-useful RPCs)
- **Phase 2b:** Policy, configuration, SSH, TCP forwarding (advanced RPCs)

- Pros: Delivers operator-useful capabilities first. Each increment is
  testable and shippable independently. Keeps spec size manageable.
- Cons: Two implementation cycles instead of one.

### B: Single Spec, All Remaining RPCs

One spec covering everything. Implement as one batch.

- Pros: One spec, one review, one merge.
- Cons: Large scope increases risk. Mixes high-priority (services) with
  low-priority (TCP forwarding) work. Harder to review.

### C: Per-Domain Increments

Each new sub-client gets its own spec.

- Pros: Most granular. Each piece can be reviewed independently.
- Cons: 5-6 tiny specs with high overhead. Some sub-clients are only
  2-3 methods.

## Decision

**Approach A: Two increments.** Phase 2a delivers the most-needed operator
capabilities. Phase 2b adds advanced features. Controller-runtime
integration (informers, CRD types, reconciler helpers) is deferred to a
separate Phase 3 brainstorm when there's a real operator to validate against.

## Phase 2a: Operator-Useful Extensions

### New Sub-Clients

**ServiceInterface** (4 RPCs):
```go
type ServiceInterface interface {
    Expose(ctx context.Context, sandbox string, spec *ServiceSpec, opts CreateOptions) (*ServiceEndpoint, error)
    Get(ctx context.Context, name string, opts GetOptions) (*ServiceEndpoint, error)
    List(ctx context.Context, sandbox string, opts ListOptions) (*ServiceEndpointList, error)
    Delete(ctx context.Context, name string, opts DeleteOptions) error
}
```

**ProviderProfileInterface** (extension to existing ProviderInterface, 5 RPCs):
```go
// Added to ProviderInterface or as a separate sub-client
type ProviderProfileInterface interface {
    ListProfiles(ctx context.Context, opts ListOptions) (*ProviderProfileList, error)
    GetProfile(ctx context.Context, name string, opts GetOptions) (*ProviderProfile, error)
    ImportProfiles(ctx context.Context, profiles []*ProviderProfile) error
    UpdateProfiles(ctx context.Context, profiles []*ProviderProfile) error
    DeleteProfile(ctx context.Context, name string, opts DeleteOptions) error
}
```

**ProviderRefreshInterface** (extension to ProviderInterface, 3 RPCs):
```go
type ProviderRefreshInterface interface {
    GetRefreshStatus(ctx context.Context, provider string) (*RefreshStatus, error)
    ConfigureRefresh(ctx context.Context, provider string, config *RefreshConfig) error
    RotateCredential(ctx context.Context, provider string) error
    DeleteRefresh(ctx context.Context, provider string) error
}
```

### Design Decisions for 2a

- Provider profiles and refresh may be sub-interfaces on the existing
  `ProviderInterface` (e.g., `client.Providers().Profiles().List(ctx, opts)`)
  or separate top-level sub-clients. Decision deferred to spec phase, but
  the nested approach matches client-go patterns for sub-resources.
- `ServiceSpec` needs port, protocol, and optional custom domain fields.
  Exact shape comes from the proto definition.

### Phase 2a Scope: ~12 RPCs, 2-3 new sub-clients.

## Phase 2b: Advanced Features

### New Sub-Clients

**PolicyInterface** (up to 12 RPCs):
```go
type PolicyInterface interface {
    GetStatus(ctx context.Context, sandbox string) (*PolicyStatus, error)
    List(ctx context.Context, opts ListOptions) (*PolicyList, error)
    GetDraft(ctx context.Context) (*DraftPolicy, error)
    ApproveDraftChunk(ctx context.Context, chunkID string) error
    RejectDraftChunk(ctx context.Context, chunkID string) error
    ApproveAllDraftChunks(ctx context.Context) error
    EditDraftChunk(ctx context.Context, chunkID string, edit *ChunkEdit) error
    UndoDraftChunk(ctx context.Context, chunkID string) error
    ClearDraftChunks(ctx context.Context) error
    GetDraftHistory(ctx context.Context) (*DraftHistory, error)
}
```

**ConfigInterface** (3 RPCs):
```go
type ConfigInterface interface {
    GetSandboxConfig(ctx context.Context) (*SandboxConfig, error)
    GetGatewayConfig(ctx context.Context) (*GatewayConfig, error)
    Update(ctx context.Context, config *ConfigUpdate) error
}
```

**SSHInterface** (2 RPCs + TCP forwarding):
```go
type SSHInterface interface {
    CreateSession(ctx context.Context, sandbox string) (*SSHSession, error)
    RevokeSession(ctx context.Context, sessionID string) error
    ForwardTCP(ctx context.Context, sandbox string, remotePort int) (net.Conn, error)
}
```

### Design Decisions for 2b

- **Policy sub-client** is large (12 RPCs) but cohesive. The draft policy
  workflow (get, approve/reject/edit chunks, history) is a single user flow.
- **SSH and TCP forwarding** are low-level. `ForwardTCP` returns a `net.Conn`
  for maximum flexibility. File transfer (Phase 1) uses SSH internally but
  that's an implementation detail.
- **Config** may not need its own sub-client for only 3 RPCs. Could be
  methods on `Client` directly. Decide in spec phase.

### Phase 2b Scope: ~17 RPCs, 2-3 new sub-clients.

## RPCs NOT Covered by the SDK

These RPCs are internal to the gateway-supervisor protocol and should NOT
be exposed in the public SDK:

| RPC | Reason |
|-----|--------|
| ConnectSupervisor | Internal: supervisor-gateway control channel |
| RelayStream | Internal: raw byte relay for SSH/exec routing |
| PushSandboxLogs | Internal: supervisor pushes logs to gateway |
| ReportPolicyStatus | Internal: supervisor reports policy state |
| SubmitPolicyAnalysis | Internal: policy engine submits analysis |
| GetSandboxProviderEnvironment | Internal: supervisor fetches env vars |
| IssueSandboxToken | Internal: token issuance for sandbox auth |
| RefreshSandboxToken | Internal: token refresh |

These are supervisor-facing RPCs, not client-facing. The SDK is a gateway
client, not a supervisor implementation.

## Roadmap Summary

| Phase | Scope | Sub-clients | RPCs |
|-------|-------|-------------|------|
| 1 (004) | Core SDK | Sandbox, Provider, Exec, File, Health | ~20 |
| 2a | Operator extensions | Service, ProviderProfile, ProviderRefresh | ~12 |
| 2b | Advanced features | Policy, Config, SSH/TCP | ~17 |
| 3 (future) | Operator support | controller-runtime integration, CRD types | N/A |

**Total client-facing RPCs: ~49 out of 55.** The remaining ~6 are
supervisor-internal and excluded by design.

## Open Questions

- Should Phase 2a include `WatchSandbox` enhancements (e.g., watch all
  sandboxes, filtered watch)? Phase 1 has basic single-sandbox watch.
- Should policy draft operations use a transaction-like pattern
  (begin/commit/rollback) instead of individual chunk operations?
- Is `net.Conn` the right return type for TCP forwarding, or should it
  be wrapped in a higher-level type with reconnection support?
- When should Phase 3 (operator support) be brainstormed? After a real
  operator prototype validates the SDK, or proactively?
