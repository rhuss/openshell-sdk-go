# Data Model: Project Setup and Build Infrastructure

This feature creates project scaffolding files, not runtime data structures.
The "entities" are configuration files and their relationships.

## Entities

### Go Module (go.mod)

| Field | Value |
|-------|-------|
| Module path | `github.com/rhuss/openshell-sdk-go` |
| Go version | 1.23 |
| Dependencies | `github.com/stretchr/testify` (test only) |

### Mise Configuration (mise.toml)

| Field | Description |
|-------|-------------|
| tools.go | Pinned Go version |
| tools.golangci-lint | Pinned linter version |
| tasks | test, test:integration, lint, fmt, build, ci |

### Linter Configuration (.golangci.yml)

| Field | Description |
|-------|-------------|
| linters.enable | govet, errcheck, staticcheck, unused, gosimple, ineffassign, revive, goimports, goheader |
| goheader.template | SPDX two-line header template |

### CI Workflow (.github/workflows/ci.yml)

| Field | Description |
|-------|-------------|
| triggers | pull_request (main), push (main) |
| jobs | lint, test, build |
| go-version | Matches mise.toml |

### Stub Package (openshell/)

| Type | Description |
|------|-------------|
| Client | Empty struct, placeholder for gRPC connection |
| Dial(address string) | Returns (*Client, error), validates non-empty address |
| Close() | Returns nil error |

## Relationships

```
mise.toml ──pins──> Go version ──used by──> go.mod, ci.yml
mise.toml ──defines──> tasks ──delegated by──> Makefile
.golangci.yml ──validates──> openshell/*.go (SPDX headers)
ci.yml ──runs──> mise tasks (lint, test, build)
```

## State Transitions

N/A. No runtime state in scaffolding files.
