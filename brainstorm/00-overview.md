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
| 006 | 2026-06-28 | converter-refactor | active | - | - |

## Dependency Chain

```
001-go-sdk (vision)
  └─ 002-project-setup (scaffolding)
       └─ 003-proto-generation (proto pipeline)
            └─ 004-core-sdk (Phase 1: sandbox, provider, exec, files, health)
                 ├─ 006-converter-refactor (dedup converters)
                 └─ 005-full-api (Phase 2a: services, profiles, refresh)
                      └─ 005-full-api (Phase 2b: policy, config, SSH, TCP)
                           └─ (future) Phase 3: operator support
```

## Open Threads
- Should the SDK provide a `fake` package (like client-go/kubernetes/fake) for consumer testing? (from #001, #004)
- Repo ownership: `rhuss` now, transfer to NVIDIA later (from #001, #002)
- SDK versioning: track gateway versions or independent semver? (from #001)
- Auth patterns: what OIDC/OAuth flows out of the box? (from #001)
- Should `Ensure` (idempotent create-or-update) be on every sub-client or just Provider? (from #004)
- Multi-gateway client support (multi-cluster)? (from #004)
- Policy draft operations: transaction pattern vs. individual chunk ops? (from #005)
- When to brainstorm Phase 3 (operator support)? (from #005)
- Types package path: `openshell/types/` vs `openshell/v1/types/`? (from #006)
- Should v1/ re-export types via aliases to avoid breaking consumers? (from #006)

## Resolved Threads
- File transfer: included in Phase 1 SDK (from #001, resolved in #004)
- Proto generation: dedicated mise task, not `go generate` (from #002, resolved in #003)
- API shape: client-go sub-client pattern with interfaces (from #001, resolved in #004)
- API versioning: `openshell/v1/` namespace from day one (from #001, resolved in #004)
- Proto as separate Go module or packages in main module? (from #003, resolved: packages in main module)
- Minimum Go version to support (from #002, resolved: Go 1.23+)

## Parked Ideas
(none)
