# Feature Specification: Add Workspace Scoping to All RPCs

**Feature Branch**: `021-workspace-scoping`
**Created**: 2026-07-31
**Status**: Draft
**Input**: Brainstorm 026-workspace-scoping, Issue #33

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Workspace-scoped sandbox operations (Priority: P1)

A dashboard backend-for-frontend (BFF) developer uses the SDK to manage sandboxes within a specific workspace. Every sandbox CRUD call includes the workspace so the gateway can enforce tenant isolation. When no workspace is specified (empty string), the gateway treats it as the "default" workspace.

**Why this priority**: Sandboxes are the primary resource in OpenShell. Every other sub-client operation depends on sandboxes existing within a workspace. This is the most frequently called set of operations from the dashboard.

**Independent Test**: Can be fully tested by creating, listing, getting, and deleting sandboxes within a named workspace, and verifying they are isolated from sandboxes in a different workspace.

**Acceptance Scenarios**:

1. **Given** a connected SDK client, **When** creating a sandbox with workspace "team-alpha", **Then** the sandbox is created in workspace "team-alpha" and the proto request includes the workspace field.
2. **Given** sandboxes exist in workspaces "team-alpha" and "team-beta", **When** listing sandboxes with workspace "team-alpha", **Then** only sandboxes in "team-alpha" are returned.
3. **Given** no explicit workspace is provided (empty string), **When** performing any sandbox operation, **Then** the operation targets the "default" workspace.
4. **Given** the fake client is used in tests, **When** creating sandboxes in different workspaces, **Then** each workspace maintains its own isolated set of sandboxes.

---

### User Story 2 - Workspace-scoped provider and profile management (Priority: P1)

A platform engineer uses the SDK to configure providers and profiles within a workspace. Provider creation, update, and deletion calls include the workspace to ensure resources are scoped correctly.

**Why this priority**: Providers are required before sandboxes can be attached to infrastructure. They are a foundational resource with the same workspace scoping needs.

**Independent Test**: Can be tested by creating providers in different workspaces and verifying CRUD isolation.

**Acceptance Scenarios**:

1. **Given** a connected SDK client, **When** creating a provider with workspace "staging", **Then** the provider is created scoped to workspace "staging".
2. **Given** providers exist in workspace "staging" and "production", **When** listing providers with workspace "staging", **Then** only "staging" providers are returned.
3. **Given** profile operations are called with a workspace, **When** importing or listing profiles, **Then** profile operations are scoped to that workspace.

---

### User Story 3 - Cross-workspace listing for platform admins (Priority: P2)

A platform administrator needs to view all resources (sandboxes, providers, profiles, services, policies) across all workspaces for monitoring and auditing purposes. They use a special listing option to bypass workspace scoping and retrieve resources from every workspace.

**Why this priority**: This is a less common operation used by administrators, not regular workspace users. It is important for operational visibility but not part of the primary user workflow.

**Independent Test**: Can be tested by creating resources in multiple workspaces, then listing with the all-workspaces option and verifying all resources appear.

**Acceptance Scenarios**:

1. **Given** sandboxes exist in workspaces "alpha", "beta", and "default", **When** listing sandboxes with the all-workspaces option enabled, **Then** sandboxes from all three workspaces are returned.
2. **Given** the all-workspaces option is not set, **When** listing sandboxes, **Then** only sandboxes in the specified workspace are returned (standard scoping behavior).

---

### User Story 4 - Workspace-scoped sandbox interactions (Priority: P2)

A developer uses the SDK to execute commands, transfer files, manage services, and handle policies within workspace-scoped sandboxes. All interactive operations (exec, file transfer, service exposure, policy management, SSH, TCP forwarding) accept the workspace parameter.

**Why this priority**: These operations are used after sandboxes exist. They build on the P1 sandbox CRUD foundation and are needed for the dashboard BFF to fully replace its current gateway layer.

**Independent Test**: Can be tested by running exec commands or uploading files to a sandbox in a specific workspace and verifying the proto requests include the workspace field.

**Acceptance Scenarios**:

1. **Given** a running sandbox in workspace "dev", **When** executing a command via `Exec().Run()` with workspace "dev", **Then** the command runs in the correct sandbox and the workspace is included in the proto request.
2. **Given** a sandbox in workspace "dev", **When** uploading a file with workspace "dev", **Then** the file is uploaded to the correct sandbox.
3. **Given** a sandbox in workspace "dev", **When** managing services, policies, SSH sessions, or TCP forwards with workspace "dev", **Then** all operations target the correct workspace.

---

### Edge Cases

- What happens when a workspace name contains special characters (spaces, unicode)? The SDK passes the workspace string through to the gateway without validation; the gateway is responsible for rejecting invalid workspace names.
- What happens when a method is called with a workspace that does not exist? The gateway returns a NotFound error, which the SDK surfaces to the caller.
- What happens when `Health().Check()` or `Config().GetGateway()` is called? These are gateway-scoped operations and do not accept a workspace parameter.
- What happens when `Config().GetSandbox()` or `Config().Update()` is called? These are sandbox-scoped and do accept the workspace parameter, since sandbox names are unique only within a workspace.
- What happens when `AllWorkspaces` is set alongside a non-empty workspace parameter? The AllWorkspaces flag takes precedence and the workspace parameter is silently ignored.
- What happens when `SSHInterface.RevokeSession` is called with a workspace? The workspace is included in the proto request for consistency, even though the token is globally unique.

## Clarifications

### Session 2026-07-31

- Q: Should SSHInterface.RevokeSession (which takes a token, not a sandbox name) accept a workspace parameter? → A: Yes, for consistency. All workspace-scoped sub-client methods get workspace per FR-001. The gateway accepts it on all workspace-scoped RPCs.
- Q: When AllWorkspaces is set on ListOptions, should a non-empty workspace parameter cause an error or be silently ignored? → A: Silently ignored. AllWorkspaces overrides workspace scoping, matching Kubernetes namespace listing conventions.
- Q: Should the fake client's AllWorkspaces listing merge results from all per-workspace stores? → A: Yes, merge from all workspace stores to faithfully replicate cross-workspace listing behavior.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: All sub-client methods that operate on workspace-scoped resources MUST accept a workspace parameter as the first argument after context.
- **FR-002**: An empty workspace string MUST be treated as the "default" workspace, matching the upstream gateway convention.
- **FR-003**: The fake (in-memory) client MUST enforce workspace isolation by maintaining separate resource stores per workspace.
- **FR-004**: List operations MUST support an all-workspaces option that retrieves resources across all workspaces, intended for platform administrators. When AllWorkspaces is set, any workspace parameter is silently ignored.
- **FR-005**: Gateway-scoped operations (`Health().Check()`, `Config().GetGateway()`) MUST NOT accept a workspace parameter.
- **FR-006**: All converter functions that build proto requests from SDK types MUST populate the workspace field on the outgoing proto message.
- **FR-007**: All existing tests MUST be updated to pass the workspace parameter, ensuring backward compatibility at the test level.
- **FR-008**: Nested sub-client methods (e.g., those on `ProfileInterface` and `RefreshInterface` accessed via `Providers().Profiles()` and `Providers().Refresh()`) MUST each accept the workspace parameter per FR-001. The factory methods themselves (`Profiles()`, `Refresh()`) remain parameterless; workspace is passed explicitly on every method call, not captured as implicit state on the sub-client.
- **FR-009**: The SSH `CreateSession` method (which accepts a sandbox ID, not a name) MUST still accept a workspace parameter for consistency, since the gateway accepts it on all workspace-scoped RPCs.
- **FR-010**: All exported type, function, and method doc comments, README examples, and package-level `doc.go` examples MUST be updated to reflect the new workspace parameter in the same PR, per Constitution XIII (Documentation Accompanies Features).

### Affected Sub-clients

- **SandboxInterface**: Create, Get, List, Delete, AttachProvider, DetachProvider, ListProviders, WaitReady, Watch, GetLogs
- **ProviderInterface**: Create, Get, List, Update, Delete, Ensure
- **ProfileInterface**: List, Get, Import, Update, Lint, Delete
- **RefreshInterface**: GetStatus, Configure, Rotate, Delete
- **ServiceInterface**: Expose, Get, List, Delete
- **ExecInterface**: Run, Stream, Interactive
- **FileInterface**: Upload, Download
- **ConfigInterface**: GetSandbox, Update (NOT GetGateway)
- **PolicyInterface**: GetDraft, ApproveDraftChunk, RejectDraftChunk, ApproveAllDraftChunks, ClearDraftChunks, GetDraftHistory, GetStatus, List, EditDraftChunk, UndoDraftChunk
- **SSHInterface**: CreateSession, RevokeSession, Tunnel
- **TCPInterface**: Forward

### Unaffected Sub-clients

- **HealthInterface**: Check (gateway-scoped, no workspace)
- **ConfigInterface.GetGateway**: gateway-scoped, no workspace

### Key Entities

- **Workspace**: A tenant-level scope that isolates resources. Represented as a plain string (name). Empty string maps to "default".
- **ListOptions**: Extended with an AllWorkspaces boolean field for cross-workspace listing.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every sub-client method that targets a workspace-scoped resource accepts and propagates the workspace parameter.
- **SC-002**: The fake client isolates resources by workspace: creating a sandbox in workspace "A" and listing in workspace "B" returns no results.
- **SC-003**: Cross-workspace listing returns resources from all workspaces when the all-workspaces option is set.
- **SC-004**: All existing unit tests pass after updating them to include the workspace parameter.
- **SC-005**: Every proto request message built by a converter for a workspace-scoped RPC includes a non-empty workspace field (verifiable by inspecting converter output in unit tests).

## Assumptions

- The SDK is pre-1.0, so breaking all existing callers by adding a mandatory workspace parameter is acceptable.
- The upstream OpenShell gateway already supports the workspace field on all relevant proto messages.
- Workspace validation (name format, existence) is handled by the gateway, not the SDK.
- The `oshell` TUI example update is out of scope for this feature and will be handled as a follow-up.
- Profile and Refresh operations are workspace-scoped because they are nested under Provider, which is workspace-scoped.
