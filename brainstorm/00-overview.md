# Brainstorm Overview

Last updated: 2026-06-29

## Sessions

| # | Date | Topic | Status | Spec | Issue |
|---|------|-------|--------|------|-------|
| 001 | 2026-06-27 | go-sdk | active | - | - |
| 002 | 2026-06-27 | project-setup | active | - | - |
| 003 | 2026-06-27 | proto-generation | active | - | - |
| 004 | 2026-06-27 | core-sdk | spec-created | 003 | - |
| 005 | 2026-06-27 | full-api | active | - | - |
| 006 | 2026-06-28 | converter-refactor | done | 004 | - |
| 007 | 2026-06-28 | fake-client | active | - | - |
| 008 | 2026-06-29 | ssh-tcp-config | active | - | - |
| 009 | 2026-06-29 | policy-logs | active | - | - |

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
                      └─ 009-policy-logs (Phase 2b-2: policy, draft policy, logs)
```

## Open Threads
- Repo ownership: `rhuss` now, transfer to NVIDIA later (from #001, #002)
- SDK versioning: track gateway versions or independent semver? (from #001)
- Auth patterns: OIDC/OAuth flows deferred to dedicated brainstorm, separate from credential refresh (from #001)
- Multi-gateway client support (multi-cluster)? (from #004)
- TCP Forward: local port binding option or raw ReadWriteCloser sufficient for v1? (from #008)
- ApproveAllDraftChunks: should Force be a parameter for security-flagged chunks? (from #009)
- Doc comment coverage: bring all exported symbols to 80%+ per Constitution IX (from PR #4 review)
- Proto-to-SDK naming verification: add lint rule or converter test pattern (from PR #4 review, Constitution X)

## Resolved Threads
- Interface evolution: follow client-go pattern — accept interface growth, provide fake, use concrete `*Client` in production
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
- Phase 4 (operator building blocks): removed — informers/listers/reconciler helpers belong in the operator repo, not the SDK

## Parked Ideas
(none)
