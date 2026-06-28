# Research: Proto Generation Pipeline

**Date**: 2026-06-27
**Branch**: `002-proto-generation`

## R1: Protoc Go Package Mapping Without go_package Option

**Decision**: Use `--go_opt=M` flags to set Go package paths at generation time.

**Rationale**: The upstream proto files have no `go_package` option and we
cannot modify them (Constitution: Upstream Tracking). The `--go_opt=M<proto>=<go_path>`
flag overrides the package path per proto file. This is the standard protoc
mechanism for this situation, documented in the protobuf-go FAQ.

**Alternatives considered**:
- Add `go_package` to upstream protos: Rejected, requires upstream changes.
- Use `option go_package` in a separate descriptor override file: No such
  mechanism exists in protoc.
- Use Buf with managed mode: Rejected in brainstorm (over-engineered for 3 files).

## R2: Protoc Version Selection

**Decision**: Pin protoc 29.6 (matches upstream OpenShell mise.toml).

**Rationale**: Using the same protoc version as upstream ensures proto file
compatibility. Version 29.x is the current stable release line (protobuf v5).

**Alternatives considered**:
- Latest protoc (30.x if available): Risk of incompatibility with upstream
  proto syntax. Match upstream for safety.
- No version pin: Non-reproducible builds violate Constitution (Minimal Dependencies
  principle extends to reproducibility).

## R3: Go Plugin Versions

**Decision**: Use latest stable `protoc-gen-go` and `protoc-gen-go-grpc`,
installed via `mise` as Go tools.

**Rationale**: These are Go modules installed via `go install`. Mise supports
pinning Go tools with the `go:` prefix syntax. Latest stable versions are
appropriate since they generate code compatible with the protobuf/gRPC Go
runtime libraries we'll depend on.

**Alternatives considered**:
- Pin specific versions: Acceptable but unnecessary complexity. The generated
  code API is stable. We pin via mise.toml regardless.

## R4: google/protobuf/struct.proto Resolution

**Decision**: Rely on protoc's built-in well-known types include path.

**Rationale**: `protoc` ships with well-known types (struct.proto, timestamp.proto,
etc.) and includes them automatically. The `--proto_path` flag for our source
directory is sufficient. No additional `--proto_path` needed for well-known types
when using a standard protoc installation (via mise).

**Alternatives considered**:
- Bundle google/protobuf/*.proto in the repo: Unnecessary, protoc handles this.
- Add explicit `--proto_path` to protoc's include dir: Only needed for
  non-standard installations.

## R5: gRPC Generation Scope

**Decision**: Generate gRPC stubs only for `openshell.proto` (the service file).
Generate message-only code for `datamodel.proto` and `sandbox.proto`.

**Rationale**: Only `openshell.proto` contains the `service OpenShell { ... }`
definition with RPCs. The other two files define message types only and have
no service definitions, so `protoc-gen-go-grpc` produces no output for them.
Running `--go-grpc_out` on all three files is harmless (no-op for message-only
files), so for simplicity we apply the same flags to all files.

**Alternatives considered**:
- Separate protoc invocations per file: More complex, no benefit.

## R6: Proto Check Implementation Strategy

**Decision**: Generate to a temp directory, then diff against committed files.

**Rationale**: This is the standard approach used by projects like grpc-go and
kubernetes. It catches both stale generated code and manual edits. The diff
output clearly shows what changed.

**Alternatives considered**:
- Hash comparison: Faster but doesn't show what diverged.
- `buf breaking`: Overkill; we're checking generation freshness, not API compatibility.

## R7: Go Module Structure for Proto Packages

**Decision**: Keep proto packages in the main module (`github.com/rhuss/openshell-sdk-go`),
not a separate Go module.

**Rationale**: Single module simplifies dependency management. Proto packages
are internal to the SDK (Constitution: Proto Isolation), so independent
versioning adds no value. A separate module would require managing replace
directives during development.

**Alternatives considered**:
- Separate `github.com/rhuss/openshell-sdk-go/proto` module: Adds complexity
  for no user-facing benefit since proto packages are not public API.
