# Brainstorm Overview

Last updated: 2026-06-28

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

## Dependency Chain

```
001-go-sdk (vision)
  └─ 002-project-setup (scaffolding)
       └─ 003-proto-generation (proto pipeline)
            └─ 004-core-sdk (Phase 1: sandbox, provider, exec, files, health)
                 ├─ 006-converter-refactor (dedup converters)
                 ├─ 007-fake-client (testing support)
                 └─ 005-full-api (Phase 2a: services, profiles, refresh)
                      └─ 005-full-api (Phase 2b: policy, config, SSH, TCP)
                           └─ (future) Phase 3: operator support
```

## Open Threads
- Repo ownership: `rhuss` now, transfer to NVIDIA later (from #001, #002)
- SDK versioning: track gateway versions or independent semver? (from #001)
- Auth patterns: OIDC/OAuth flows deferred to dedicated brainstorm, separate from credential refresh (from #001)
- Multi-gateway client support (multi-cluster)? (from #004)
- Policy draft operations: transaction pattern vs. individual chunk ops? (from #005)
- When to brainstorm Phase 3 (operator support)? (from #005)
- File transfer: API exists but defaultSSHTransport is a stub. Add golang.org/x/crypto/ssh or defer? (from #004, reopened after PR #1 review)
- Should reactors be included in fake v1 or deferred? (from #007)
- Should fake Exec return configurable ExecResult or leave to consumer mocking? (from #007)
- Fake package path: `openshell/v1/fake/` vs `openshell/fake/`? (from #007)

## Resolved Threads
- Proto generation: dedicated mise task, not `go generate` (from #002, resolved in #003)
- API shape: client-go sub-client pattern with interfaces (from #001, resolved in #004)
- API versioning: `openshell/v1/` namespace from day one (from #001, resolved in #004)
- Proto as separate Go module or packages in main module? (from #003, resolved: packages in main module)
- Minimum Go version to support (from #002, resolved: Go 1.23+)
- Fake package: yes, in-memory object store following client-go pattern (from #001, #004, resolved in #007)
- Ensure: Provider only, not on every sub-client. Services are ephemeral. (from #004)
- Types package: `openshell/v1/types/` following k8s.io/api pattern, consumers import directly, no re-exports (from #006)
- Auth/credential refresh: refresh mechanism in Phase 2a, OIDC/OAuth flows deferred to separate brainstorm (from #001)

## Parked Ideas
(none)
