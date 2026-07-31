# Research: Proto Sync from Upstream PR #2445

**Date**: 2026-07-31

## R1: Does buf.gen.yaml need changes for inference.proto?

**Decision**: Yes, buf.gen.yaml requires explicit updates.

**Rationale**: The current `buf.gen.yaml` uses `inputs.paths` to explicitly list each proto file. Buf does not auto-discover proto files from a directory when paths are specified. The new `inference.proto` must be added to the paths list, and both plugins need `Minference.proto` import mappings to route generated code to the correct Go package.

**Alternatives considered**:
- Use `directory: proto` without `paths` for auto-discovery. Rejected because this would also pick up any future internal protos accidentally. Explicit paths provide a clear allowlist.

## R2: Upstream proto files beyond the 5 listed

**Decision**: Only copy the 5 SDK-relevant protos. Exclude `compute_driver.proto`, `gateway_interceptor.proto`, `supervisor_middleware.proto`, and `test.proto`.

**Rationale**: The excluded protos are internal/operator-only (`compute_driver`, `gateway_interceptor`, `supervisor_middleware`) or test fixtures (`test.proto`). They define services and types that SDK clients never interact with. Including them would add unnecessary generated code and potentially expose internal implementation details.

**Alternatives considered**:
- Copy all upstream protos for completeness. Rejected per FR-008 and principle V (Minimal Dependencies).

## R3: datamodel.proto changes in PR #2445

**Decision**: Copy `datamodel.proto` from upstream regardless of whether it changed in PR #2445.

**Rationale**: The upstream file at `/Users/rhuss/Work/projects/OpenShell/proto/datamodel.proto` is 3.2KB, matching the SDK's current copy size. It may not have changed in PR #2445 specifically, but syncing it ensures consistency. The copy is idempotent if unchanged.

## R4: inference.proto package and import structure

**Decision**: The inference proto uses Go package `inferencev1` following the existing naming pattern.

**Rationale**: Existing protos use the pattern `proto/<name>v1/` (e.g., `openshellv1`, `sandboxv1`). The `inference.proto` likely declares `package inference;` or similar, and the Go import mapping in `buf.gen.yaml` routes it to `github.com/rhuss/openshell-sdk-go/proto/inferencev1`. This follows the established convention.

## R5: Proto:sync mise task

**Decision**: No `proto:sync` mise task currently exists. Proto copying is manual.

**Rationale**: The AGENTS.md mentions `mise run proto:sync` but the task is not defined in `mise.toml`. The brainstorm references `UPSTREAM_PATH` as the mechanism. For this feature, manual copying with explicit verification (byte-for-byte match, SC-001) is appropriate. A `proto:sync` task could be added but is out of scope for this PR.
