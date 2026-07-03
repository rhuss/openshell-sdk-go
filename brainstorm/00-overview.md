# Brainstorm Overview

Last updated: 2026-07-03 (020 brainstormed)

## Active Sessions

| # | Date | Topic | Status | Spec | Issue |
|---|------|-------|--------|------|-------|
| 001 | 2026-06-27 | go-sdk | active | - | - |
| 015 | 2026-06-30 | local-port-listener | active | - | - |
| 016 | 2026-07-01 | reverse-forwarding | active | - | [#18](https://github.com/rhuss/openshell-sdk-go/issues/18) |
| 018 | 2026-07-02 | edge-auth | spec-created | 016 | [#20](https://github.com/rhuss/openshell-sdk-go/issues/20) |
| 020 | 2026-07-03 | oidc-login | active | - | [#24](https://github.com/rhuss/openshell-sdk-go/issues/24) |

## Attic (implemented)

| # | Date | Topic | Spec |
|---|------|-------|------|
| 002 | 2026-06-27 | project-setup | 001 |
| 003 | 2026-06-27 | proto-generation | 002 |
| 004 | 2026-06-27 | core-sdk | 003 |
| 005 | 2026-06-27 | full-api | 006 |
| 006 | 2026-06-28 | converter-refactor | 004 |
| 007 | 2026-06-28 | fake-client | 005 |
| 008 | 2026-06-29 | ssh-tcp-config | 007 |
| 009 | 2026-06-29 | policy-logs | 008 |
| 010 | 2026-06-29 | ssh-tunnel-forward-opts | 009 |
| 011 | 2026-06-29 | api-docs | 010 |
| 012 | 2026-06-30 | context-cancel-cleanup | 011 |
| 013 | 2026-06-30 | typed-sandbox-policy | 012 |
| 014 | 2026-06-30 | name-id-consistency | 013 |
| 017 | 2026-07-01 | sdk-core-auth | 015 |
| 019 | 2026-07-02 | cli-auth-convenience | 017 |

## Dependency Chain

```
001-go-sdk (vision)
  └─ 002-project-setup (scaffolding)
       └─ 003-proto-generation (proto pipeline)
            └─ 004-core-sdk (Phase 1: sandbox, provider, exec, files, health)
                 ├─ 006-converter-refactor (dedup converters)
                 ├─ 007-fake-client (testing support)
                 └─ 005-full-api (Phase 2a: services, profiles, refresh)
                      ├─ 008-ssh-tcp-config (Phase 2b-1: SSH, TCP, config)
                      │    └─ 010-ssh-tunnel-forward-opts (Phase 2b-3: SSH tunnel, forward opts)
                      └─ 009-policy-logs (Phase 2b-2: policy, draft policy, logs)
```

## Open Threads
- SandboxPolicy.Version: server-assigned or caller-settable? (from #013)
- ResourceRequirements (proto field 9) also missing from SDK SandboxSpec, add in same pass or separate? (from #013)
- WatchSandboxRequest.id: does the server actually reject names, or resolve both? (from #014)
- Name-to-ID caching: should the SDK cache name→ID mappings? Deferred, no caching for now. (from #014)
- Upstream proto cleanup: file on NVIDIA/OpenShell after SDK changes merge (from #014)
- Repo ownership: `rhuss` now, transfer to NVIDIA later (from #001, #002)
- SDK versioning: track gateway versions or independent semver? (from #001)
- Auth patterns: OIDC/OAuth flows deferred to dedicated brainstorm (from #001, core token refresh implemented in spec 015)
- Multi-gateway client support (multi-cluster)? (from #004)
- NetworkPolicyRule and related sandbox.v1 types: inspect sandbox proto for exact SDK type shape (from #009)
- ListSandboxPolicies `global` flag: should SDK expose this or keep sandbox-scoped only? (from #009)
- Doc comment coverage: bring all exported symbols to 80%+ per Constitution IX (from PR #4 review)
- Proto-to-SDK naming verification: add lint rule or converter test pattern (from PR #4 review, Constitution X)
- WithServiceID shared vs per-interface option type? (from #010)
- Listen error propagation: should per-connection Forward failures be returned from Accept(), silently retried, or routed to a callback? (from #015)
- WithOnError callback: should Listen accept an error handler for connection-level failures? (from #015)
- Connection limit: WithMaxConnections() deferred but internal design should accommodate it (from #015)
- Proto model for reverse forwarding: which approach does upstream prefer, client-polls or gateway-push? (from #016)
- WithOnError callback for RemoteListen: should per-connection errors surface via callback or just logging? (from #016)
- Sandbox-side bind semantics: port allocation, conflict handling inside sandbox (from #016)
- Reverse forwarding auth model: reuse existing credentials or separate token? (from #016)
- WithSSHTunnel for reverse direction: needed or always direct TCP? (from #016)
- Connection limit for RemoteListen: part of v1 or deferred? (from #016)
- CLI convenience layer: gateway config loading, disk-aware fileTokenSource, FromGatewayConfig(name) (from #017, separate brainstorm)
- Full OIDC browser flow: now brainstormed as #020, active (from #017, #019)
- OIDC client_id source for gateway-aware login: metadata.json field, separate oidc_config.json, or derived? (from #020)
- OIDC custom scopes: hardcode required scopes or allow custom? (from #020)
- Token refresh ownership: oidc package or RefreshableToken? (from #020)
- Device code flow display: callback, stdout, or channel? (from #020)
- ID token validation: verify signature/audience or trust token endpoint? (from #020)
- Dynamic edge token refresh (WithHeaderFunc) if edge tokens need independent refresh cycles (from #018, future brainstorm)
- Other edge proxy convenience constructors (Google IAP, Zscaler) when concrete use cases arise (from #018, future brainstorm)
- Gateway management operations (Add, Remove, SetActive) if Go programs need them (from #019, future brainstorm)
- Multi-gateway client support (connecting to multiple gateways from one process) (from #019, future brainstorm)

## Resolved Threads
- Interface evolution: follow client-go pattern, accept interface growth, provide fake, use concrete `*Client` in production
- Fake package path: `openshell/v1/fake/` (from #007, resolved by implementation in spec 005)
- Fake reactors: deferred to future brainstorm (from #007)
- Fake Exec configurable result: deferred, consumers mock ExecInterface directly (from #007)
- Proto generation: dedicated mise task, not `go generate` (from #002, resolved in #003)
- API shape: client-go sub-client pattern with interfaces (from #001, resolved in #004)
- API versioning: `openshell/v1/` namespace from day one (from #001, resolved in #004)
- Proto as separate Go module or packages in main module? (from #003, resolved: packages in main module)
- Minimum Go version to support (from #002, resolved: Go 1.23+)
- Fake package: yes, in-memory object store following client-go pattern (from #001, #004, resolved in #007)
- Ensure: Provider only, not on every sub-client. Services are ephemeral. (from #004)
- Types package: `openshell/v1/types/` following k8s.io/api pattern, consumers import directly, no re-exports (from #006)
- Auth/credential refresh: refresh mechanism in Phase 2a, OIDC/OAuth flows deferred to separate brainstorm (from #001)
- Phase 2b split: SSH/TCP/Config (008) and Policy/Logs (009) as separate specs (from #005)
- SSH/TCP placement: top-level sub-clients, not nested under Sandboxes (from #008)
- SSH vs TCP: separate sub-clients, different interface shapes (from #008)
- TCP Forward return type: `io.ReadWriteCloser`, not `net.Conn` (from #008)
- Config sub-client: own top-level interface for consistency (from #008)
- Policy interface: flat, not nested DraftPolicyInterface (from #009)
- GetLogs: added to SandboxInterface, not a standalone LogsInterface (from #009)
- Phase 4 (operator building blocks): removed, informers/listers/reconciler helpers belong in the operator repo, not the SDK
- TCP Forward: local port binding option or raw ReadWriteCloser sufficient for v1? (from #008, resolved: ReadWriteCloser for v1, Listen() convenience layer in #015)
- Listen transport: TCP Forward by default, WithSSHTunnel() option for per-connection SSH auth (from #015)
- Listen API surface: on TCPInterface, not standalone type (from #015)
- Listen fake: returns Unimplemented error, same as fake Tunnel() (from #015)
- ApproveAllDraftChunks: `include_security_flagged` via functional option, not Force param (from #009)
- ErrorConflict: map `codes.Aborted` only, leave `FAILED_PRECONDITION` as ErrorInternal (from #009)
- MergeOperations: typed struct-per-variant, not raw bytes or interface polymorphism (from #009)
- PolicyChunk: full fidelity (18 fields), flat struct (from #009)
- GetLogs ID resolution: accept sandbox name, resolve to ID internally (from #009)
- GetDraft status filter: functional option pattern (from #009)
- SSH tunnel API: SSH().Tunnel() with auto-lifecycle, not Forward() options (from #010)
- Tunnel return type: io.ReadWriteCloser, not richer TunnelConn struct (from #010)
- Tunnel close behavior: auto-revoke session on Close() (from #010)
- Forward() gap fix: WithServiceID() only, SSH target options deferred (from #010)
- WithServiceID on Tunnel: yes, for symmetry with Forward() (from #010)
- Tunnel() sandbox name vs ID: resolved by spec 013, Tunnel accepts name and resolves internally (from #010)
- WithExtraHeaders: empty values silently skipped, goroutine-per-connection, WithTLS supported, types.Logger for logging, validate non-empty edgeToken, drain then force-close (from #018, resolved in spec 016)
- NewClient resolves both auth and TLS from gateway metadata (from #019, resolved in brainstorm expansion)
- Full oauth2.TokenSource for OIDC disk tokens with refresh cycling (from #019, resolved in brainstorm expansion)
- Active gateway supported via marker file (from #019, resolved in brainstorm expansion)
- LoadConfig returns frozen snapshot (from #019, resolved in brainstorm expansion)
- Direct edge package dependency, no callback indirection (from #019, resolved in brainstorm expansion)
- User-first gateway precedence with source info in ListGateways (from #019, resolved in brainstorm expansion)
- mTLS cert loading is part of gateway package (from #019, resolved in brainstorm expansion)
- oauth2.Token: callers import golang.org/x/oauth2 directly; SDK does not re-export (from #017, resolved in spec 015)
- RefreshableToken leeway default: 10s, matching k8s client-go (from #017, resolved in spec 015)
- 401-triggered token reset: not implemented, overkill for gRPC (from #017, resolved in spec 015)

## Parked Ideas
(none)
