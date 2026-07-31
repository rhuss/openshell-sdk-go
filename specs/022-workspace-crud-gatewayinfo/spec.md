# Feature Specification: Workspace CRUD, GatewayInfo & GetCurrentUser

**Feature Branch**: `022-workspace-crud-gatewayinfo`
**Created**: 2026-07-31
**Status**: Draft
**Input**: Brainstorm document `brainstorm/027-workspace-crud-gatewayinfo.md`

## User Scenarios & Testing

### User Story 1 - Workspace Lifecycle Management (Priority: P1)

An SDK consumer managing a multi-tenant platform needs to create, retrieve, list, and delete workspaces programmatically. They create a workspace with a name and labels, retrieve it later to check its state, list all workspaces they have access to, and delete workspaces that are no longer needed.

**Why this priority**: Workspace CRUD is the foundation for multi-tenant operations. Without it, users cannot organize resources into logical groupings.

**Independent Test**: Can be fully tested by creating a workspace, verifying it appears in list results, retrieving it by name, and deleting it. Delivers workspace lifecycle value independently of member management.

**Acceptance Scenarios**:

1. **Given** a valid client with platform admin credentials, **When** `Create` is called with a name and labels, **Then** a Workspace is returned with the specified name, labels, and a creation timestamp.
2. **Given** a workspace exists, **When** `Get` is called with that workspace's name, **Then** the workspace's full details are returned.
3. **Given** multiple workspaces exist, **When** `List` is called with pagination options, **Then** only workspaces the caller is a member of are returned, respecting limit and offset.
4. **Given** a workspace exists with no active resources, **When** `Delete` is called with that workspace's name, **Then** the workspace is removed.
5. **Given** a non-existent workspace name, **When** `Get` is called, **Then** a not-found error is returned.
6. **Given** `List` is called with a label selector, **Then** only workspaces matching those labels are returned.

---

### User Story 2 - Workspace Member Management (Priority: P1)

A workspace administrator needs to control who has access to a workspace. They add members with specific roles (admin or user), list current members for auditing, and remove members who no longer need access.

**Why this priority**: Member management is tightly coupled with workspace CRUD; workspaces without access control are not useful in multi-tenant environments.

**Independent Test**: Can be tested by adding a member to an existing workspace, listing members to verify presence and role, and removing the member. Delivers access control value independently.

**Acceptance Scenarios**:

1. **Given** a workspace exists and the caller is a workspace admin, **When** `AddMember` is called with a principal subject and role, **Then** the member is added and a WorkspaceMember is returned.
2. **Given** a workspace has members, **When** `ListMembers` is called with pagination options, **Then** all members are returned with their roles.
3. **Given** a member exists in a workspace, **When** `RemoveMember` is called with that member's principal subject, **Then** the member is removed.
4. **Given** an attempt to add a member who already exists, **When** `AddMember` is called, **Then** an already-exists error is returned.
5. **Given** an attempt to remove a non-existent member, **When** `RemoveMember` is called, **Then** a not-found error is returned.

---

### User Story 3 - Gateway Metadata Retrieval (Priority: P2)

A platform administrator needs to query the gateway for operational metadata: its status, version, and available compute drivers. This supports monitoring dashboards, compatibility checks, and operational tooling.

**Why this priority**: Gateway metadata is useful for operational visibility but not required for core workspace operations. It serves a narrower audience (platform admins only).

**Independent Test**: Can be tested by calling `GetGatewayInfo` and verifying the response contains status, version, and a list of compute drivers.

**Acceptance Scenarios**:

1. **Given** a client with platform admin credentials, **When** `GetGatewayInfo` is called, **Then** the response includes gateway status, version string, and a list of compute drivers.
2. **Given** each compute driver info entry in the response, **Then** it includes a name and capabilities (driver-reported name and version).

---

### User Story 4 - Authenticated User Identity (Priority: P2)

An SDK consumer needs to determine who the current authenticated user is. This supports authorization checks, audit logging, and user-facing displays of the current identity within tooling built on the SDK.

**Why this priority**: Identity resolution is valuable for any authenticated workflow but does not block workspace or gateway operations.

**Independent Test**: Can be tested by calling `GetCurrentUser` with a valid token and verifying the response contains the caller's identity details.

**Acceptance Scenarios**:

1. **Given** a client authenticated with a valid bearer token, **When** `GetCurrentUser` is called, **Then** the response includes the caller's subject identifier, display name, roles, scopes, and identity provider.
2. **Given** a client with an expired or invalid token, **When** `GetCurrentUser` is called, **Then** an authentication error is returned.

---

### User Story 5 - Fake Client Support for Testing (Priority: P1)

SDK consumers writing tests for their own applications need a fake client that supports workspace, member, gateway info, and current user operations without requiring a live gateway. The fake must validate inputs the same way the real client does.

**Why this priority**: Without fake client support, consumers cannot write reliable unit tests. This is a core SDK invariant that applies to all new functionality.

**Independent Test**: Can be tested by using the fake client to create workspaces, add members, and retrieve gateway info, verifying the same validation and error behavior as documented for the real client.

**Acceptance Scenarios**:

1. **Given** a fake client, **When** workspace CRUD operations are called, **Then** they behave identically to the real client for input validation and error cases, using in-memory storage.
2. **Given** a fake client, **When** member operations are called, **Then** they enforce the same validation (non-empty workspace, non-empty subject, valid role).
3. **Given** a fake client with configurable gateway info, **When** `GetGatewayInfo` is called, **Then** the configured response is returned.
4. **Given** a fake client with configurable current user, **When** `GetCurrentUser` is called, **Then** the configured identity is returned.

---

### Edge Cases

- What happens when `Create` is called with a workspace name that already exists? An already-exists error is returned.
- What happens when `Delete` is called on a workspace with active members? The gateway handles this server-side; the SDK passes through whatever the gateway returns.
- What happens when `AddMember` is called with an invalid role value? A validation error is returned before the RPC is sent.
- What happens when `List` or `ListMembers` is called with limit=0? The SDK treats 0 as "use server default" (no client-side error).
- What happens when `GetGatewayInfo` is called by a non-admin user? The gateway returns a permission-denied error, which the SDK passes through.
- How does a user change a member's role? There is no UpdateMember operation; users must remove and re-add the member with the new role.

## Clarifications

### Session 2026-07-31

- Q: Should the sub-client accessor be named `Health()` or `Gateway()`? → A: Keep `Health()` to avoid breaking changes. The existing `HealthInterface` is extended with `GetGatewayInfo` and `GetCurrentUser`.
- Q: Should the Workspace type include all ObjectMeta fields (annotations, workspace, deletion_timestamp) added in spec 021? → A: Yes, include all ObjectMeta fields for consistency with existing SDK types like Sandbox.
- Q: Should `Delete` return the deleted Workspace or just an error? → A: Return only `error`, following the standard SDK delete pattern.

## Requirements

### Functional Requirements

- **FR-001**: The SDK MUST provide a workspace sub-client accessible from the main client, following the existing sub-client accessor pattern.
- **FR-002**: The workspace sub-client MUST support creating a workspace with a name and optional labels.
- **FR-003**: The workspace sub-client MUST support retrieving a workspace by name.
- **FR-004**: The workspace sub-client MUST support listing workspaces with pagination (limit, offset) and label selector filtering.
- **FR-005**: The workspace sub-client MUST support deleting a workspace by name.
- **FR-006**: The workspace sub-client MUST support adding a member to a workspace with a principal subject and role (admin or user).
- **FR-007**: The workspace sub-client MUST support removing a member from a workspace by principal subject.
- **FR-008**: The workspace sub-client MUST support listing members of a workspace with pagination.
- **FR-009**: The existing `Health()` sub-client MUST be extended with a `GetGatewayInfo` method that returns gateway status, version, and compute drivers.
- **FR-010**: The existing `Health()` sub-client MUST be extended with a `GetCurrentUser` method that returns the authenticated caller's identity (subject, display name, roles, scopes, identity provider).
- **FR-011**: All new operations MUST validate required inputs (non-empty names, valid roles) before making remote calls, returning clear validation errors.
- **FR-012**: All new SDK types MUST be idiomatic Go types, not exposing upstream protocol types directly.
- **FR-013**: The fake client MUST implement all new operations with in-memory storage and the same input validation as the real client.
- **FR-014**: The workspace role MUST be represented as a typed enum with values for Admin and User (not untyped strings).
- **FR-015**: All new types returned from operations MUST use deep copies, ensuring no shared mutable references with internal representations.

### Key Entities

- **Workspace**: Represents a logical grouping of resources. Attributes: ObjectMeta (name, labels, annotations, creation timestamp, workspace, deletion timestamp, resource version), status (containing lifecycle phase). The lifecycle phase is a typed enum with values Active and Terminating. All ObjectMeta fields are included for consistency with existing SDK types.
- **WorkspaceMember**: Represents a user's membership in a workspace. Attributes: ObjectMeta (id, name, labels, timestamps, resource version), principal subject identifier (OIDC subject claim), role (admin or user).
- **WorkspaceRole**: Enumeration of member roles within a workspace: Admin, User.
- **WorkspacePhase**: Enumeration of workspace lifecycle phases: Active, Terminating.
- **ServiceStatus**: Enumeration of gateway service status: Healthy, Degraded, Unhealthy.
- **GatewayInfo**: Gateway operational metadata. Attributes: status (typed enum: Healthy, Degraded, Unhealthy), gateway version string, list of compute driver info entries.
- **ComputeDriverInfo**: A compute backend available on the gateway. Attributes: name (gateway-selected driver routing key), capabilities (containing driver-reported name and driver-reported version).
- **CurrentUser**: The authenticated caller's identity. Attributes: subject identifier (OIDC sub claim), display name, roles (list of strings), scopes (list of OAuth2 scope strings), identity provider name.

## Success Criteria

### Measurable Outcomes

- **SC-001**: All 7 workspace operations (create, get, list, delete, add member, remove member, list members) are callable from the SDK and produce correct results.
- **SC-002**: `GetGatewayInfo` and `GetCurrentUser` are callable from the existing health/gateway sub-client.
- **SC-003**: Every new public function has at least one unit test covering the success path and one covering an error path.
- **SC-004**: The fake client passes the same test suite as the real client for input validation and error behavior.
- **SC-005**: No upstream protocol types are exposed in the public API; all conversions happen internally.
- **SC-006**: All new operations validate inputs client-side and return descriptive errors for invalid arguments (empty names, invalid roles).

## Out of Scope

- **UpdateWorkspace**: No update RPC exists in the gateway proto.
- **UpdateMember / role changes**: No UpdateMember RPC exists; role changes require remove + re-add (documented in edge cases).
- **Workspace-scoped resource enumeration**: Listing sandboxes/providers within a workspace is handled by existing sub-clients with workspace scoping (spec 021), not by the workspace sub-client.
- **Gateway config operations**: GetGatewayConfig and UpdateConfig are existing operations, not part of this feature.

## Assumptions

- The upstream proto definitions for all RPCs (CreateWorkspace, GetWorkspace, ListWorkspaces, DeleteWorkspace, AddWorkspaceMember, RemoveWorkspaceMember, ListWorkspaceMembers, GetGatewayInfo, GetCurrentUser) already exist and are available in the generated proto bindings.
- The SDK is pre-1.0, so extending the existing health/gateway interface is acceptable without a deprecation period.
- There is no UpdateMember RPC in the gateway; role changes require remove + re-add. This is a gateway constraint, not an SDK design choice.
- Label selectors for `ListWorkspaces` use the same format as the existing SDK list options pattern.
- The `ListWorkspaces` RPC returns only workspaces the caller is a member of (server-side filtering). The SDK does not need to provide an "all workspaces" option.
