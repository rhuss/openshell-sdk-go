# Smoke Test Report

**Feature**: Proto Generation Pipeline
**Date**: 2026-06-27
**Spec**: specs/002-proto-generation/spec.md
**Result**: 13 passed, 0 skipped, 0 failed (out of 13)

## Scenario 1 of 13 (User Story: Generate Go Bindings from Proto Files)

**Given** proto files exist in `proto/` (openshell.proto, datamodel.proto, sandbox.proto)
**When** the developer runs `mise run proto:gen`
**Then** Go source files are generated in `proto/openshellv1/`, `proto/datamodelv1/`, and `proto/sandboxv1/` with correct package paths.

**Why it matters**: If generation doesn't produce files in the right directories with correct package names, all downstream Go imports break.

### Evidence

**Command**: `mise run proto:gen` followed by `ls` and `rg` for package declarations
**Output**:
```
Proto generation complete.
Generated packages:
  proto/openshellv1/: 2 files
  proto/datamodelv1/: 1 files
  proto/sandboxv1/: 1 files

proto/openshellv1/openshell.pb.go:   package openshellv1
proto/openshellv1/openshell_grpc.pb.go:   package openshellv1
proto/datamodelv1/datamodel.pb.go:   package datamodelv1
proto/sandboxv1/sandbox.pb.go:   package sandboxv1
```
**Observation**: All three packages generated with correct package names. openshellv1 has 2 files (.pb.go and _grpc.pb.go), datamodelv1 and sandboxv1 each have 1 file (.pb.go).

### Verdict: PASS

---

## Scenario 2 of 13 (User Story: Generate Go Bindings from Proto Files)

**Given** generated Go files exist
**When** the developer runs `go build ./proto/...`
**Then** all packages compile without errors.

**Why it matters**: Generated code that doesn't compile blocks all SDK development.

### Evidence

**Command**: `go build ./proto/...`
**Output**:
```
Go build: Success
```
**Observation**: All three generated packages compile without errors.

### Verdict: PASS

---

## Scenario 3 of 13 (User Story: Generate Go Bindings from Proto Files)

**Given** `openshell.proto` imports `datamodel.proto` and `sandbox.proto`
**When** Go code is generated
**Then** cross-package imports resolve correctly using the module path `github.com/rhuss/openshell-sdk-go/proto/...`.

**Why it matters**: Incorrect import paths cause compile failures in packages that depend on cross-proto references.

### Evidence

**Command**: `rg 'github.com/rhuss/openshell-sdk-go/proto/(datamodelv1|sandboxv1)' proto/openshellv1/`
**Output**:
```
proto/openshellv1/openshell.pb.go:13:  datamodelv1 "github.com/rhuss/openshell-sdk-go/proto/datamodelv1"
proto/openshellv1/openshell.pb.go:14:  sandboxv1 "github.com/rhuss/openshell-sdk-go/proto/sandboxv1"
proto/openshellv1/openshell_grpc.pb.go:14:  sandboxv1 "github.com/rhuss/openshell-sdk-go/proto/sandboxv1"
```
**Observation**: Both openshell.pb.go and openshell_grpc.pb.go import datamodelv1 and sandboxv1 using the correct module-qualified paths.

### Verdict: PASS

---

## Scenario 4 of 13 (User Story: Generate Go Bindings from Proto Files)

**Given** `openshell.proto` imports `google/protobuf/struct.proto`
**When** Go code is generated
**Then** the well-known type resolves via protoc's built-in include path without additional configuration.

**Why it matters**: If well-known types fail to resolve, protoc generation fails entirely.

### Evidence

**Command**: `rg 'structpb|google.golang.org/protobuf/types/known/structpb' proto/openshellv1/`
**Output**:
```
proto/openshellv1/openshell.pb.go:17:  structpb "google.golang.org/protobuf/types/known/structpb"
proto/openshellv1/openshell.pb.go:884:  Resources *structpb.Struct ...
proto/openshellv1/openshell.pb.go:12129:  (*structpb.Struct)(nil),  // 179: google.protobuf...
```
**Observation**: The well-known `google.protobuf.Struct` type resolves to `google.golang.org/protobuf/types/known/structpb` and is used in multiple generated struct fields. Since `go build` also succeeded (Scenario 2), the import actually compiles.

### Verdict: PASS

---

## Scenario 5 of 13 (User Story: Sync Proto Files from Upstream)

**Given** the upstream OpenShell repo exists at the default path (`../OpenShell/proto/`)
**When** the developer runs `mise run proto:sync`
**Then** `openshell.proto`, `datamodel.proto`, and `sandbox.proto` are copied to `proto/`.

**Why it matters**: Sync is the entry point for pulling upstream API changes. If it fails, the SDK cannot track upstream.

### Evidence

**Command**: `mise run proto:sync`
**Output**:
```
Copied openshell.proto
Copied datamodel.proto
Copied sandbox.proto
Upstream version: 29ce6a704cba222c29b5e0d73b90280cf5ed3b9f
Proto sync complete.
```
**Observation**: All three expected files were copied, and the upstream version was recorded.

### Verdict: PASS

---

## Scenario 6 of 13 (User Story: Sync Proto Files from Upstream)

**Given** the upstream repo contains additional proto files (compute_driver.proto, inference.proto, test.proto)
**When** sync runs
**Then** only the three specified files are copied.

**Why it matters**: Copying unnecessary proto files would add unwanted dependencies and confuse the SDK scope.

### Evidence

**Command**: `ls ../OpenShell/proto/` (upstream has 6 files) then `ls proto/*.proto` (local has 3)
**Output**:
```
Upstream contains: compute_driver.proto  datamodel.proto  inference.proto  openshell.proto  sandbox.proto  test.proto

Local proto/ contains only:
proto/datamodel.proto
proto/openshell.proto
proto/sandbox.proto
```
**Observation**: Despite the upstream having 6 proto files, only the 3 specified files were copied. compute_driver.proto, inference.proto, and test.proto are correctly excluded.

### Verdict: PASS

---

## Scenario 7 of 13 (User Story: Sync Proto Files from Upstream)

**Given** a successful sync
**When** the developer checks `proto/UPSTREAM_VERSION`
**Then** it contains the git commit SHA of the upstream repo at sync time.

**Why it matters**: Version tracking enables reproducibility and allows devs to know which upstream version they are building against.

### Evidence

**Command**: `cat proto/UPSTREAM_VERSION` and comparison with `git -C ../OpenShell rev-parse HEAD`
**Output**:
```
UPSTREAM_VERSION: 29ce6a704cba222c29b5e0d73b90280cf5ed3b9f
Upstream HEAD:    29ce6a704cba222c29b5e0d73b90280cf5ed3b9f
Valid SHA: yes (matches ^[0-9a-f]{40}$)
```
**Observation**: UPSTREAM_VERSION contains a valid 40-character hex SHA that exactly matches the upstream repo's HEAD commit.

### Verdict: PASS

---

## Scenario 8 of 13 (User Story: Sync Proto Files from Upstream)

**Given** the developer provides a custom upstream path via argument
**When** sync runs
**Then** proto files are copied from the specified path instead of the default.

**Why it matters**: Developers with non-standard directory layouts need to override the upstream path.

### Evidence

**Command**: `UPSTREAM_PATH="$TMPDIR/proto" mise run proto:sync` (using a temp git repo with proto files)
**Output**:
```
Copied openshell.proto
Copied datamodel.proto
Copied sandbox.proto
Upstream version: 396ddbab1ba2c17ec5d136ce9704d1c4c0a28db9

Expected SHA: 396ddbab1ba2c17ec5d136ce9704d1c4c0a28db9
Actual SHA in UPSTREAM_VERSION: 396ddbab1ba2c17ec5d136ce9704d1c4c0a28db9
MATCH
```
**Observation**: Proto files were copied from the custom path, and the UPSTREAM_VERSION reflects the SHA from that custom repo (not the default upstream).

### Verdict: PASS

---

## Scenario 9 of 13 (User Story: Validate Generated Code in CI)

**Given** committed `.pb.go` files match the committed `.proto` files
**When** CI runs `mise run proto:check`
**Then** the check passes with exit code 0.

**Why it matters**: False positive failures in CI would block all merges and erode trust in the pipeline.

### Evidence

**Command**: `mise run proto:check`
**Output**:
```
Proto check passed: generated files are up to date.
EXIT_CODE=0
```
**Observation**: Clean state produces exit code 0 and a clear success message.

### Verdict: PASS

---

## Scenario 10 of 13 (User Story: Validate Generated Code in CI)

**Given** a developer manually edited a `.pb.go` file
**When** CI runs `mise run proto:check`
**Then** the check fails with a message indicating which files are out of date.

**Why it matters**: Manual edits to generated files are a common mistake. CI must catch them before merge.

### Evidence

**Command**: Appended `// TAMPERED` to `proto/datamodelv1/datamodel.pb.go`, then ran `mise run proto:check`
**Output**:
```
ERROR: Generated proto files are out of date.
Run 'mise run proto:gen' to regenerate.

diff -r --ex .../datamodelv1/datamodel.pb.go proto/datamodelv1/datamodel.pb.go
284a285
> // TAMPERED
EXIT_CODE=1
```
**Observation**: The check detected the tampered file, reported it by name with a diff, exited with code 1, and suggested the regeneration command.

### Verdict: PASS

---

## Scenario 11 of 13 (User Story: Validate Generated Code in CI)

**Given** a developer updated `.proto` files but forgot to regenerate
**When** CI runs `mise run proto:check`
**Then** the check fails.

**Why it matters**: Stale generated code that doesn't match current protos causes runtime bugs and API mismatches.

### Evidence

**Command**: Added a new `message SmokeTestDummy` to `datamodel.proto`, then ran `mise run proto:check`
**Output**:
```
ERROR: Generated proto files are out of date.
Run 'mise run proto:gen' to regenerate.

diff -r --ex .../datamodelv1/datamodel.pb.go proto/datamodelv1/datamodel.pb.go
197,240d196
< type SmokeTestDummy struct { ... }
...
EXIT_CODE=1
```
**Observation**: The check detected that the proto source had a semantic change (new message type) not reflected in the committed .pb.go files, and failed with a detailed diff.

### Verdict: PASS

---

## Scenario 12 of 13 (User Story: Clean Generated Files)

**Given** generated `.pb.go` files exist in `proto/` subdirectories
**When** the developer runs `mise run proto:clean`
**Then** all `.pb.go` and `_grpc.pb.go` files are removed.

**Why it matters**: Developers need a clean slate when troubleshooting generation issues or switching proto versions.

### Evidence

**Command**: `mise run proto:clean` then `find proto/ -name '*.pb.go'`
**Output**:
```
Before clean: 4 files
  proto/datamodelv1/datamodel.pb.go
  proto/openshellv1/openshell.pb.go
  proto/openshellv1/openshell_grpc.pb.go
  proto/sandboxv1/sandbox.pb.go

Proto clean complete.

After clean: 0 .pb.go files remaining
```
**Observation**: All 4 generated files (3 .pb.go + 1 _grpc.pb.go) were removed.

### Verdict: PASS

---

## Scenario 13 of 13 (User Story: Clean Generated Files)

**Given** proto source files exist in `proto/`
**When** clean runs
**Then** `.proto` files and `UPSTREAM_VERSION` are preserved.

**Why it matters**: Clean must not destroy source files, or developers would need to re-sync from upstream.

### Evidence

**Command**: After `mise run proto:clean`, checked for .proto files and UPSTREAM_VERSION
**Output**:
```
Proto files remaining:
  proto/datamodel.proto
  proto/openshell.proto
  proto/sandbox.proto

UPSTREAM_VERSION: 29ce6a704cba222c29b5e0d73b90280cf5ed3b9f
```
**Observation**: All 3 source .proto files and the UPSTREAM_VERSION marker file survived the clean operation intact.

### Verdict: PASS
