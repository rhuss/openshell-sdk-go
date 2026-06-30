# Brainstorm: Proto Generation Pipeline

**Date:** 2026-06-27
**Status:** active
**Depends on:** [002-project-setup](002-project-setup.md) (build tooling must exist first)

## Problem Framing

The OpenShell gateway exposes a gRPC API defined in `.proto` files. The Go
SDK needs Go bindings generated from these protos to communicate with the
gateway. The upstream protos live in the OpenShell repo (Rust/Python), have
no `go_package` option, and are generated differently per language (Rust
uses tonic-prost via build.rs, Python uses grpc_tools.protoc via mise task).

Key challenges:
1. How do upstream proto files reach this repo?
2. How are Go bindings generated and maintained?
3. How are Go package paths assigned (no `go_package` in upstream protos)?
4. Which of the three upstream services need Go bindings?

## Approaches Considered

### A: Copy + Commit with Protoc Flag Overrides (Chosen)

Copy upstream protos into the SDK repo, commit them. Run `protoc` with
`protoc-gen-go` and `protoc-gen-go-grpc` locally, using `--go_opt=M` flags
to set Go package paths without modifying the proto files. Commit generated
`.pb.go` files.

- Pros: No submodule complexity. Users can `go get` without protoc. Version
  is pinned explicitly by the committed proto files. No upstream modification
  needed.
- Cons: Proto files can drift from upstream if sync task is not run. Generated
  code adds to repo size. Reviewers must trust that generated code matches
  the committed protos.

### B: Git Submodule + Build-Time Generation

Add OpenShell as a git submodule, generate Go code at build time.

- Pros: Protos always linked to a specific upstream commit. No copied files.
- Cons: Submodules are painful for contributors. Every build needs protoc
  and plugins installed. Cannot `go get` without build tools.

### C: Buf CLI with buf.gen.yaml

Use the Buf CLI to manage proto generation. `buf.gen.yaml` defines package
mapping and plugin configuration.

- Pros: Structured, declarative proto management. Buf handles dependency
  resolution. Linting and breaking change detection built in.
- Cons: Adds Buf as a dependency (upstream doesn't use it). Over-engineered
  for three proto files with no external proto dependencies beyond
  `google/protobuf/struct.proto`.

## Decision

**Approach A: Copy + Commit with Protoc Flag Overrides.** It's the simplest
approach that works. Three proto files don't justify Buf or submodule
complexity. The sync task makes upstream updates a deliberate action with a
clear diff.

## Key Requirements

### Proto File Management

**Upstream source:** `/proto/` directory in OpenShell repo
(currently at `../OpenShell/proto/` relative to this repo)

**Files to copy (the OpenShell service import chain):**
- `openshell.proto` (package `openshell.v1`, 44 RPCs)
- `datamodel.proto` (package `openshell.datamodel.v1`, imported by openshell.proto)
- `sandbox.proto` (package `openshell.sandbox.v1`, imported by openshell.proto)

**Files NOT copied (not needed for the gateway client SDK):**
- `compute_driver.proto` (provider-side contract, Rust only)
- `inference.proto` (internal inference routing)
- `test.proto` (test fixtures)

**Layout in SDK repo:**
```
proto/
  openshell.proto
  datamodel.proto
  sandbox.proto
```

**Sync task:** `mise run proto:sync` copies protos from a configured
upstream path (default: `../OpenShell/proto/`), preserving only the three
needed files. Optionally accepts a git ref or path override.

### Go Code Generation

**Tools:**
- `protoc` (protocol buffer compiler)
- `protoc-gen-go` (Go message code generation, `google.golang.org/protobuf`)
- `protoc-gen-go-grpc` (Go gRPC stub generation, `google.golang.org/grpc`)

**Tool versions pinned in `mise.toml`.**

**Generated Go packages:**
```
proto/
  openshellv1/       # openshell.v1 service + messages
    openshell.pb.go
    openshell_grpc.pb.go
  datamodelv1/       # openshell.datamodel.v1 messages
    datamodel.pb.go
  sandboxv1/         # openshell.sandbox.v1 messages
    sandbox.pb.go
```

**Package path mapping via `--go_opt=M` flags:**
```
--go_opt=Mopenshell.proto=github.com/rhuss/openshell-sdk-go/proto/openshellv1
--go_opt=Mdatamodel.proto=github.com/rhuss/openshell-sdk-go/proto/datamodelv1
--go_opt=Msandbox.proto=github.com/rhuss/openshell-sdk-go/proto/sandboxv1
```

Equivalent `--go-grpc_opt=M` flags for gRPC stubs.

**Generation task:** `mise run proto:gen` runs protoc with all flags.
Generated files are committed to the repo.

**Validation task:** `mise run proto:check` verifies that committed generated
code matches what protoc would produce (for CI to catch stale generated code).

### Mise Tasks

| Task | Description |
|------|-------------|
| `proto:sync` | Copy protos from upstream OpenShell repo |
| `proto:gen` | Run protoc to generate Go bindings |
| `proto:check` | Verify generated code is up to date (CI) |
| `proto:clean` | Remove generated `.pb.go` files |

### Proto Isolation (Constitution Principle)

Generated proto types are **internal to the SDK**. The public API in
`openshell/` defines its own domain types (`Client`, `SandboxRef`,
`ExecResult`, etc.) and converts to/from proto types internally.

Consumers never import `proto/openshellv1` directly. This allows:
- Proto version changes without public API breakage
- Idiomatic Go types (e.g., `time.Duration` instead of proto `int64` millis)
- Hiding proto fields that aren't relevant to SDK consumers

### CI Integration

The `proto:check` task should run in CI to ensure:
1. Committed `.pb.go` files match the committed `.proto` files
2. No manual edits to generated files
3. Proto sync + generation is reproducible

### Upstream Proto Version Tracking

Each sync should record the upstream commit hash in a metadata file:
```
proto/UPSTREAM_VERSION
```
Content: the git commit SHA from the OpenShell repo at sync time.
This makes it easy to see which upstream version the protos track.

## Open Questions

- Should `proto/` be a separate Go module (`github.com/rhuss/openshell-sdk-go/proto`)
  to allow independent versioning? Or keep it as packages within the main module?
  (Leaning toward single module for simplicity.)
- Should the generated code include the `Inference` service for future use,
  or add it only when needed? (Decided: skip for now, add later.)
- How should `google/protobuf/struct.proto` be resolved at generation time?
  (protoc's built-in well-known types should handle this automatically.)
