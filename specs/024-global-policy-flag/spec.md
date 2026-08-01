# Feature Specification: Global Policy Flag

**Feature Branch**: `024-global-policy-flag`  
**Created**: 2026-08-01  
**Status**: Draft  
**Input**: GitHub issue #44

## Overview

The SDK's policy client currently only supports sandbox-scoped policy queries. The upstream gateway also supports a "global" policy mode where policies apply gateway-wide rather than to a specific sandbox. The SDK needs to expose this mode so that platform administrators can query global policy revisions and status through the same SDK methods they use for sandbox-scoped queries.

## User Scenarios & Testing

### User Story 1 - List Global Policy Revisions (Priority: P1)

A platform administrator needs to view the global policy revision history to understand what gateway-wide policies are in effect and track changes over time, without specifying a sandbox name.

**Why this priority**: This is the primary use case driving the feature. The dashboard's "Global Policy" page cannot display policy revisions without this capability.

**Independent Test**: Can be tested by listing global policy revisions and verifying results are returned without requiring a sandbox name.

**Acceptance Scenarios**:

1. **Given** global policies exist on the gateway, **When** a platform admin lists policies with the global flag enabled, **Then** the global policy revision history is returned.
2. **Given** the global flag is enabled, **When** a sandbox name is also provided, **Then** the sandbox name is ignored and global revisions are returned.
3. **Given** the global flag is not set, **When** a user lists policies without a sandbox name, **Then** the operation fails with a validation error (existing behavior preserved).

---

### User Story 2 - Get Global Policy Status (Priority: P1)

A platform administrator needs to check the load status of a specific global policy version to verify whether it has been successfully applied gateway-wide.

**Why this priority**: Equally critical as listing, needed for the same dashboard page to show per-version status.

**Independent Test**: Can be tested by querying global policy status and verifying the response contains the correct revision and load status.

**Acceptance Scenarios**:

1. **Given** a global policy version exists, **When** a platform admin queries its status with the global flag enabled, **Then** the status result is returned for that global version.
2. **Given** the global flag is enabled, **When** a specific version is also requested, **Then** the status for that specific global version is returned.

---

### User Story 3 - Test Global Policy Operations Without a Live Backend (Priority: P2)

A developer building a management UI needs the fake client to support global policy mode so they can write tests without a live gateway.

**Why this priority**: Testing support, needed for any consumer of the SDK to write reliable tests for global policy features.

**Independent Test**: Can be tested by using the fake client to store and retrieve global policy revisions separately from sandbox-scoped ones.

**Acceptance Scenarios**:

1. **Given** a developer uses the fake client, **When** they list policies with global flag, **Then** only globally-scoped revisions are returned.
2. **Given** sandbox-scoped and global revisions both exist in the fake, **When** querying without global flag, **Then** only sandbox-scoped revisions are returned (no cross-contamination).

### Edge Cases

- When `global=true`, the sandbox name parameter MUST be ignored (not validated), matching gateway behavior.
- When `global=true`, the workspace parameter MUST be ignored, matching gateway behavior.
- When `global=false` (default), existing validation behavior MUST be preserved unchanged.
- The global option MUST compose correctly with existing options (e.g., `WithVersion` on GetStatus, limit/offset on List).

## Clarifications

### Session 2026-08-01

- Q: Does ListPolicyOption already exist or needs to be created? → A: Already exists with WithLimit and WithOffset. The Global field extends the existing listPolicyConfig struct.
- Q: Does GetStatusOption already exist? → A: Already exists with WithVersion. The Global field extends the existing getStatusConfig struct.
- Q: Should WithGlobal be a single option reusable across both List and GetStatus? → A: No, each has its own option type (ListPolicyOption, GetStatusOption). Create separate WithGlobal constructors returning each type, named `WithListGlobal` and `WithStatusGlobal`. Go does not support function overloading, so distinct names are required.

## Out of Scope

- Adding global support to other policy methods (GetDraft, ApproveDraftChunk, etc.) since those RPCs don't have a global flag in the proto.
- Modifying the PolicyInterface signature. The change uses existing functional options.
- Adding new error types or response types.

## Requirements

### Functional Requirements

- **FR-001**: The SDK MUST support a global mode for listing policy revisions that retrieves gateway-wide policies instead of sandbox-scoped ones.
- **FR-002**: The SDK MUST support a global mode for querying policy status that retrieves gateway-wide status instead of sandbox-scoped status.
- **FR-003**: When global mode is enabled on `GetStatus`, the SDK MUST skip validation of the sandbox name parameter (name may be empty). The `List` method does not have a sandbox name parameter, so this requirement applies only to `GetStatus`.
- **FR-004**: When global mode is enabled, the SDK MUST skip validation of the workspace parameter (workspace may be empty). This applies to both `List` and `GetStatus`.
- **FR-005**: The global option MUST be exposed through the existing functional options pattern, consistent with other SDK options like WithVersion and WithStatusFilter.
- **FR-006**: The fake client MUST support global policy mode for `List` and `GetStatus` with in-memory storage that separates global from sandbox-scoped revisions. The current fake returns Unimplemented for all policy methods, so this requires implementing real in-memory `List` and `GetStatus` methods (not just adding a global flag to existing stubs).
- **FR-007**: Existing behavior for sandbox-scoped policy queries MUST remain unchanged when the global flag is not set (backward compatible).
- **FR-008**: New public API symbols (`WithListGlobal`, `WithStatusGlobal`) MUST have Go doc comments and be included in documentation updates (README feature list, package doc.go examples) per Constitution XIII.

## Error Handling

- When global mode is not set and no sandbox name is provided to `GetStatus`, the SDK MUST return an `InvalidArgument` error (preserving existing validation behavior).
- When global mode is enabled but the gateway returns an error (e.g., no global policies exist), the SDK MUST propagate the gRPC error through the existing `converter.FromGRPCError` path, consistent with all other policy methods.
- No new error types are introduced. All errors use the existing `StatusError` type with standard error codes (`ErrorNotFound`, `ErrorInvalidArgument`, etc.).

## Success Criteria

### Measurable Outcomes

- **SC-001**: Developers can list global policy revisions using the SDK without providing a sandbox name or workspace.
- **SC-002**: Developers can query global policy version status using the SDK without providing a sandbox name.
- **SC-003**: All existing policy tests continue to pass unchanged (backward compatibility).
- **SC-004**: The global option composes correctly with existing options (version filtering, pagination).

## Assumptions

- The upstream gateway's `global` field on `ListSandboxPoliciesRequest` and `GetSandboxPolicyStatusRequest` is stable and behaves as documented in the proto comments.
- The SDK already has an established functional options pattern for policy methods (GetDraftOption, GetStatusOption, ApproveAllOption) that the global option will follow.
- Only `List` and `GetStatus` need the global flag. Other policy RPCs do not have a `global` field in the proto.
- The `ListPolicyOption` type already exists with `WithLimit` and `WithOffset` constructors. The `listPolicyConfig` struct needs a `global` field added, with a new `WithListGlobal` constructor.
