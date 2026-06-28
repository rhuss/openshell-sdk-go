# Implementation Plan: Proto Generation Pipeline

**Branch**: `002-proto-generation` | **Date**: 2026-06-27 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/002-proto-generation/spec.md`

## Summary

Set up the proto generation pipeline for the OpenShell Go SDK: copy upstream
`.proto` files, generate Go bindings with `protoc` + plugins using `--go_opt=M`
flag overrides (no upstream modifications), pin tool versions in `mise.toml`,
and add CI validation for staleness detection. Four mise tasks: sync, gen,
check, clean.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: protoc 29.6, protoc-gen-go (google.golang.org/protobuf),
protoc-gen-go-grpc (google.golang.org/grpc/cmd/protoc-gen-go-grpc),
google.golang.org/grpc (runtime)
**Storage**: N/A
**Testing**: Go testing package + `go build ./proto/...` for compilation validation
**Target Platform**: Cross-platform (Linux, macOS, Windows)
**Project Type**: Library (Go SDK)
**Performance Goals**: Sync + gen under 30 seconds
**Constraints**: Zero modifications to upstream proto files; generated code committed to repo
**Scale/Scope**: 3 proto files in, 3 Go packages out (~5 generated files total)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| Proto Isolation | PASS | Generated packages under `proto/` are internal; public API in `openshell/` wraps them |
| Idiomatic Go | PASS | Standard Go package layout, versioned package names (openshellv1, datamodelv1, sandboxv1) |
| Test-First | PASS | proto:check validates generation correctness; go build validates compilation |
| Upstream Tracking | PASS | proto/UPSTREAM_VERSION records source commit SHA; proto:sync is the sync mechanism |
| Minimal Dependencies | PASS | Only adds gRPC runtime (required for proto stubs); protoc/plugins are dev-only tools |

No violations. All applicable principles satisfied.

## Project Structure

### Documentation (this feature)

```text
specs/002-proto-generation/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
└── tasks.md             # Phase 2 output (via /speckit-tasks)
```

### Source Code (repository root)

```text
proto/
├── openshell.proto          # Copied from upstream (openshell.v1)
├── datamodel.proto          # Copied from upstream (openshell.datamodel.v1)
├── sandbox.proto            # Copied from upstream (openshell.sandbox.v1)
├── UPSTREAM_VERSION         # Git commit SHA of upstream at sync time
├── openshellv1/             # Generated Go package
│   ├── openshell.pb.go
│   └── openshell_grpc.pb.go
├── datamodelv1/             # Generated Go package
│   └── datamodel.pb.go
└── sandboxv1/               # Generated Go package
    └── sandbox.pb.go
```

**Structure Decision**: Proto source files live at `proto/` root level.
Generated Go packages live in versioned subdirectories matching the proto
package names (minus dots). This keeps source and generated code co-located
while maintaining clear separation via directory structure.

## Implementation Approach

### Tool Version Pinning

Pin in `mise.toml`:
- `protoc = "29.6"` (matches upstream OpenShell)
- `go:google.golang.org/protobuf/cmd/protoc-gen-go` (latest stable)
- `go:google.golang.org/grpc/cmd/protoc-gen-go-grpc` (latest stable)

### Proto Sync (mise run proto:sync)

Shell script that:
1. Accepts optional `UPSTREAM_PATH` env var (default: `../OpenShell/proto/`)
2. Validates the source path exists
3. Copies exactly 3 files: openshell.proto, datamodel.proto, sandbox.proto
4. Reads `git -C "$UPSTREAM_PATH/.." rev-parse HEAD` for the commit SHA
5. Writes SHA to `proto/UPSTREAM_VERSION` (or "unknown" if not a git repo)

### Proto Generation (mise run proto:gen)

Shell script that:
1. Checks protoc, protoc-gen-go, protoc-gen-go-grpc are available
2. Creates output directories: `proto/openshellv1/`, `proto/datamodelv1/`, `proto/sandboxv1/`
3. Runs protoc with:
   - `--proto_path=proto/` (source root)
   - `--go_out=.` and `--go-grpc_out=.`
   - `--go_opt=M` flags mapping each proto to its Go package path
   - `--go-grpc_opt=M` flags (same mapping)
4. Only openshell.proto needs gRPC generation (it contains the service definition);
   datamodel.proto and sandbox.proto are message-only

### Proto Check (mise run proto:check)

Shell script that:
1. Generates to a temp directory using the same protoc flags
2. Diffs temp output against committed files
3. Exits 0 if identical, exits 1 with diff output if divergent

### Proto Clean (mise run proto:clean)

Shell script that:
1. Removes all `*.pb.go` files under `proto/` recursively
2. Removes empty generated subdirectories
3. Preserves `.proto` files and `UPSTREAM_VERSION`

### Go Module Dependencies

Add to `go.mod`:
- `google.golang.org/protobuf` (protobuf runtime, required by generated .pb.go)
- `google.golang.org/grpc` (gRPC runtime, required by generated _grpc.pb.go)

### Linter Configuration

Update `.golangci.yml` to exclude `proto/` subdirectories from the `goheader`
linter, since generated `.pb.go` files have their own headers from protoc.

### CI Integration

Add `proto:check` to the CI pipeline (GitHub Actions workflow) so it runs
alongside lint, build, and test.
