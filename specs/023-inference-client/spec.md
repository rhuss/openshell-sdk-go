# Feature Specification: Inference Route Client

**Feature Branch**: `023-inference-client`  
**Created**: 2026-07-31  
**Status**: Draft  
**Input**: Brainstorm 028-inference-client.md

## Overview

The inference route client adds workspace-scoped inference routing management to the OpenShell SDK. Workspace administrators can configure how inference requests are routed by setting a provider, model, and route name for a given workspace. This capability follows the SDK's established sub-client pattern (like Sandboxes, Workspaces, Config) and exposes three operations: set, get, and delete. A sandbox-internal bundle resolution RPC exists in the upstream proto but is explicitly excluded from the SDK's public surface.

## User Scenarios & Testing

### User Story 1 - Configure Inference Routing for a Workspace (Priority: P1)

A workspace administrator needs to set up how inference requests are routed for their workspace. They configure a route by specifying a provider, model, and route name so that workspace users can run inference through the configured path.

**Why this priority**: This is the primary write operation. Without the ability to configure routes, the entire inference routing feature is unusable.

**Independent Test**: Can be tested by creating a route configuration for a workspace and verifying the route is persisted and retrievable.

**Acceptance Scenarios**:

1. **Given** a workspace exists and the caller is a workspace admin, **When** they set an inference route with a provider, model, and route name, **Then** the route is created and the configured route details are returned.
2. **Given** a workspace already has a configured route, **When** the admin sets a new route with updated parameters, **Then** the existing route is replaced with the new configuration.
3. **Given** the caller is not a workspace admin, **When** they attempt to set an inference route, **Then** the operation fails with a permission error.

---

### User Story 2 - Retrieve Inference Route Configuration (Priority: P1)

A workspace user needs to view the current inference routing configuration for their workspace to understand how inference requests will be handled.

**Why this priority**: Reading route configuration is equally critical, as it enables visibility and is needed for display in dashboards and management UIs.

**Independent Test**: Can be tested by retrieving a previously configured route and verifying all fields are returned correctly.

**Acceptance Scenarios**:

1. **Given** a workspace has a configured inference route, **When** a workspace user retrieves the route by name, **Then** the full route configuration is returned.
2. **Given** a workspace has no configured route for the requested name, **When** a user retrieves the route, **Then** an appropriate error is returned indicating the route was not found.

---

### User Story 3 - Remove Inference Route (Priority: P2)

A workspace administrator needs to remove an inference route from their workspace, disabling inference routing through that path.

**Why this priority**: Deletion is important for lifecycle management but is used less frequently than create/read operations.

**Independent Test**: Can be tested by deleting an existing route and confirming subsequent retrieval returns a not-found error.

**Acceptance Scenarios**:

1. **Given** a workspace has a configured inference route, **When** an admin deletes the route by name, **Then** the route is removed and subsequent retrieval returns not found.
2. **Given** a workspace has no configured route for the requested name, **When** an admin attempts to delete it, **Then** the operation completes without error (idempotent).
3. **Given** the caller is not a workspace admin, **When** they attempt to delete an inference route, **Then** the operation fails with a permission error.

---

### User Story 4 - Test Inference Operations Without a Live Backend (Priority: P2)

A developer building an application that uses inference routing needs a way to test their integration without connecting to a live gateway, enabling fast iteration and CI testing.

**Why this priority**: Supports developer workflow and testing infrastructure. Required for any consumer of the SDK to write reliable tests.

**Independent Test**: Can be tested by using the fake implementation to perform set/get/delete operations and verifying correct behavior without network calls.

**Acceptance Scenarios**:

1. **Given** a developer uses the fake implementation, **When** they set an inference route, **Then** the route is stored in memory and can be retrieved.
2. **Given** a developer uses the fake implementation, **When** they delete a route, **Then** subsequent retrieval returns not found.

### Edge Cases

- Setting a route with empty or missing required fields (provider name, model ID) MUST return a validation error before contacting the gateway. The SDK validates required fields client-side.
- Retrieving or deleting a route with an empty workspace identifier MUST return a validation error. Workspace is required for all inference operations.
- Concurrent route modifications to the same workspace are handled server-side by the gateway (last-write-wins). The SDK does not implement client-side locking or conflict detection.
- An empty string route name is valid and represents the default user-facing route. This is not an error condition.

## Clarifications

### Session 2026-07-31

- Q: What should happen when setting a route with empty or missing required fields? → A: Client-side validation error before RPC call
- Q: What should happen with an empty workspace identifier? → A: Validation error; workspace is required for all operations
- Q: How should concurrent route modifications be handled? → A: Defer to server-side (last-write-wins, no client-side locking)
- Q: What fields does the InferenceRoute response contain? → A: Mirrors the route config fields (provider name, model identifier, route name, timeout, TLS verification). Server-assigned metadata (if any) is determined by the proto definition at sync time; the SDK exposes whatever the proto response message contains

## Requirements

### Functional Requirements

- **FR-001**: The SDK MUST provide inference route management as a distinct capability, separate from other workspace configuration.
- **FR-002**: The SDK MUST support setting an inference route for a workspace, accepting route configuration details including provider name, model identifier, route name, timeout, and TLS verification preference.
- **FR-003**: The SDK MUST support retrieving an inference route for a workspace by route name.
- **FR-004**: The SDK MUST support deleting an inference route from a workspace by route name.
- **FR-005**: The SDK MUST NOT expose sandbox-internal operations (such as bundle resolution) that are not intended for end-user consumption.
- **FR-006**: The SDK MUST provide a fake/mock implementation of the inference client that operates entirely in memory for testing purposes.
- **FR-007**: The SDK MUST use route name as a plain string parameter, with an empty string representing the default user-facing route.
- **FR-008**: The SDK MUST accept workspace as a method parameter rather than embedding it in configuration objects, consistent with how other workspace-scoped operations work in the SDK.
- **FR-009**: Set and delete operations MUST require workspace admin authorization. Get operations MUST require workspace user authorization.
- **FR-010**: All inference client methods MUST accept context.Context as the first parameter, following the standard Go convention and the existing SDK pattern.
- **FR-011**: The inference client MUST include documentation updates: README.md feature list and usage example, package-level doc.go with a runnable example, and Go doc comments on all exported types and methods (per Constitution VIII, IX, XIII).

### Key Entities

- **InferenceRouteConfig**: The configuration for an inference route, including provider name (string), model identifier (string), route name (string), timeout (time.Duration), and TLS verification skip flag (bool).
- **InferenceRoute**: The response representing a configured inference route, containing the full route configuration fields as persisted by the gateway. The exact field set mirrors InferenceRouteConfig; any additional server-assigned fields are determined by the proto response message at sync time.

### Out of Scope

- **Bundle resolution RPC**: The sandbox-internal bundle resolution operation is not exposed in the SDK (see FR-005).
- **Client-side conflict detection**: Concurrent route modifications are handled server-side (last-write-wins). The SDK does not implement optimistic locking or version tracking.
- **Route listing**: There is no "list all routes" operation. Routes are retrieved individually by name.
- **Route validation beyond required fields**: The SDK validates that required fields are non-empty but does not validate provider/model existence. The gateway performs semantic validation.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Developers can configure, retrieve, and remove inference routes for any workspace using three dedicated operations.
- **SC-002**: All inference operations complete successfully against the fake implementation without requiring network access or a running gateway.
- **SC-003**: The inference client follows the same access pattern as other SDK capabilities (accessed via a top-level accessor on the main client).
- **SC-004**: The inference client correctly enforces authorization requirements, rejecting unauthorized callers with clear error messages.

## Assumptions

- The upstream gateway's inference service definition is stable and available for proto synchronization.
- Proto file synchronization (brainstorm #025) provides the mechanism to import the inference proto definition into the SDK.
- The SDK already has an established sub-client pattern (e.g., Sandboxes, Config, Workspaces) that the inference client will follow.
- Only the three user-facing RPCs (Set, Get, Delete) are in scope. The sandbox-internal bundle resolution RPC is explicitly excluded.
- Route names are plain strings. The empty string represents the default user-facing route, and "sandbox-system" is reserved for internal use by the gateway.
- Workspace is passed as a method parameter rather than embedded in configuration types, following the existing SDK convention for workspace-scoped operations.
