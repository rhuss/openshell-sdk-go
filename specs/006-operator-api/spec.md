# Feature Specification: Operator API Extensions (Phase 2a)

**Feature Branch**: `006-operator-api`
**Created**: 2026-06-28
**Status**: Draft
**Input**: Phase 2a operator-useful API extensions: service exposure, provider profiles, credential refresh, and WatchOptions enhancement.

## User Scenarios & Testing

### User Story 1 - Operator Exposes Sandbox Services (Priority: P1)

An operator building automation tooling needs to expose HTTP services running inside a sandbox to external consumers. The operator creates a service endpoint specifying the sandbox, service name, target port, and whether a browser-facing URL should be generated. They can then retrieve, list, and delete service endpoints for any sandbox.

**Why this priority**: Service exposure is the most requested operator capability. Without it, operators must use lower-level mechanisms to route traffic to sandbox services.

**Independent Test**: Create a sandbox, expose a service on port 8080 with domain enabled, verify the returned endpoint contains a URL. List services for the sandbox, verify the exposed service appears. Delete the service, verify subsequent Get returns NotFound.

**Acceptance Scenarios**:

1. **Given** a running sandbox "web-app", **When** the operator calls Services().Expose with service name "api", target port 8080, and domain true, **Then** a ServiceEndpoint is returned containing the sandbox name, service name, target port, and a URL.
2. **Given** an exposed service "api" on sandbox "web-app", **When** the operator calls Services().Get with sandbox "web-app" and service "api", **Then** the matching ServiceEndpoint is returned.
3. **Given** multiple exposed services on sandbox "web-app", **When** the operator calls Services().List with sandbox "web-app", **Then** all service endpoints for that sandbox are returned.
4. **Given** an exposed service "api" on sandbox "web-app", **When** the operator calls Services().Delete, **Then** a subsequent Get returns a NotFound error.
5. **Given** no sandbox named "missing" exists, **When** the operator calls Services().Expose, **Then** a NotFound error is returned.

---

### User Story 2 - Operator Manages Provider Profiles (Priority: P1)

An operator managing a fleet of providers needs to list available provider type profiles, import custom profiles, validate profiles before import, update existing custom profiles, and delete profiles that are no longer needed. Profiles define provider type templates (credentials schema, endpoints, binaries) that providers are instantiated from.

**Why this priority**: Provider profiles are the foundation for provider management. Operators need to customize and validate profiles before deploying providers at scale.

**Independent Test**: List built-in profiles, import a custom profile, verify it appears in the list. Get the profile by ID. Lint a profile before import to check for errors. Update the profile with a new version. Delete the profile, verify subsequent Get returns NotFound.

**Acceptance Scenarios**:

1. **Given** built-in profiles exist on the gateway, **When** the operator calls Providers().Profiles().List, **Then** a list of provider profiles is returned including built-in profiles.
2. **Given** a profile ID "custom-llm", **When** the operator calls Providers().Profiles().Get with that ID, **Then** the matching profile is returned with all fields (display name, description, category, credentials, endpoints).
3. **Given** a valid profile import item, **When** the operator calls Providers().Profiles().Import, **Then** the import result contains the imported profile and imported=true.
4. **Given** an invalid profile import item, **When** the operator calls Providers().Profiles().Lint, **Then** the lint result contains diagnostics describing the validation errors and valid=false.
5. **Given** an existing custom profile "custom-llm" with resource version 1, **When** the operator calls Providers().Profiles().Update with expected resource version 1, **Then** the profile is updated and the new resource version is returned.
6. **Given** a stale resource version, **When** the operator calls Providers().Profiles().Update with an outdated expected resource version, **Then** a conflict or precondition-failed error is returned.
7. **Given** a custom profile "custom-llm", **When** the operator calls Providers().Profiles().Delete, **Then** a subsequent Get returns NotFound.

---

### User Story 3 - Operator Configures Credential Refresh (Priority: P2)

An operator needs to set up automatic credential refresh for a provider so that expiring credentials are rotated without manual intervention. The operator configures refresh parameters, checks refresh status, triggers manual rotation, and can remove the refresh configuration.

**Why this priority**: Credential expiration causes service outages. Automated refresh prevents operators from needing to manually rotate credentials on a schedule.

**Independent Test**: Configure refresh for a provider, verify the configuration is accepted. Check refresh status, verify it reflects the configuration. Trigger a manual rotation, verify success. Delete the refresh configuration, verify subsequent GetStatus reflects removal.

**Acceptance Scenarios**:

1. **Given** a provider "openai-prod", **When** the operator calls Providers().Refresh().Configure with a refresh configuration, **Then** the configuration is accepted without error.
2. **Given** a configured refresh for "openai-prod", **When** the operator calls Providers().Refresh().GetStatus, **Then** the refresh status reflects the active configuration.
3. **Given** a configured refresh for "openai-prod", **When** the operator calls Providers().Refresh().Rotate, **Then** a credential rotation is triggered successfully.
4. **Given** a configured refresh for "openai-prod", **When** the operator calls Providers().Refresh().Delete, **Then** the refresh configuration is removed and subsequent GetStatus reflects no active refresh.
5. **Given** no refresh configured for "new-provider", **When** the operator calls Providers().Refresh().GetStatus, **Then** the status indicates no refresh is configured (not an error).

---

### User Story 4 - Consumer Uses StopOnTerminal Watch Option (Priority: P2)

A consumer watching a sandbox lifecycle wants the watch to automatically stop when the sandbox reaches a terminal phase (Ready or Error) instead of continuing to receive events indefinitely. This simplifies the common "wait for ready or error" pattern.

**Why this priority**: Most watch consumers only care about reaching a terminal state. Without StopOnTerminal, they must implement their own phase-checking logic to decide when to stop watching.

**Independent Test**: Start a Watch with StopOnTerminal=true, create a sandbox, wait for it to reach Ready. Verify the watch channel closes automatically after the Ready event.

**Acceptance Scenarios**:

1. **Given** a Watch started with StopOnTerminal=true, **When** the sandbox transitions to Ready, **Then** the watch delivers the Ready event and then closes the event channel.
2. **Given** a Watch started with StopOnTerminal=true, **When** the sandbox transitions to Error, **Then** the watch delivers the Error event and then closes the event channel.
3. **Given** a Watch started with StopOnTerminal=false (default), **When** the sandbox transitions to Ready, **Then** the watch delivers the Ready event but the channel remains open for further events.

---

### User Story 5 - Consumer Uses Updated Fake Client (Priority: P3)

A consumer writing tests against the SDK uses the fake client. After the Phase 2a update, the fake client compiles and satisfies the updated ClientInterface. New sub-clients (Services, Profiles, Refresh) return Unimplemented errors, directing consumers to mock these interfaces directly if needed.

**Why this priority**: The fake client must remain compilable after adding new interfaces to ClientInterface. Without stubs, the fake breaks all consumers on upgrade.

**Independent Test**: Instantiate a FakeClient, call Services().Expose, verify Unimplemented error. Call Providers().Profiles().List, verify Unimplemented error. Call Providers().Refresh().GetStatus, verify Unimplemented error.

**Acceptance Scenarios**:

1. **Given** a FakeClient, **When** the consumer calls any Services() method, **Then** an Unimplemented error is returned.
2. **Given** a FakeClient, **When** the consumer calls any Providers().Profiles() method, **Then** an Unimplemented error is returned.
3. **Given** a FakeClient, **When** the consumer calls any Providers().Refresh() method, **Then** an Unimplemented error is returned.
4. **Given** the updated SDK with new interfaces, **When** the FakeClient is compiled, **Then** it satisfies the updated ClientInterface without compilation errors.

---

### Edge Cases

- What happens when Services().Expose is called with an empty sandbox name? A validation error (InvalidArgument) is returned.
- What happens when Services().Expose is called with target port 0? A validation error is returned.
- What happens when Profiles().Import is called with an empty profile list? The operation succeeds with imported=false and no profiles returned.
- What happens when Profiles().Update is called with a built-in profile ID? An error is returned indicating built-in profiles cannot be modified.
- What happens when Refresh().Configure is called with an empty provider name? A validation error is returned.
- What happens when Refresh().Rotate is called for a provider with no refresh configured? An error indicating no refresh configuration exists is returned.
- What happens when the gateway is unavailable during any operation? An Unavailable error is returned, consistent with Phase 1 behavior.

## Requirements

### Functional Requirements

**Service Exposure**

- **FR-001**: The SDK MUST provide a ServiceInterface accessible via `client.Services()` with Expose, Get, List, and Delete operations for sandbox HTTP service endpoints.
- **FR-002**: Services().Expose MUST accept a sandbox name, service name, target port, and domain flag, and return a ServiceEndpoint containing the endpoint details and URL (when domain is enabled).
- **FR-003**: Services().Get MUST retrieve a service endpoint by sandbox name and service name, returning NotFound if it does not exist.
- **FR-004**: Services().List MUST return all service endpoints for a given sandbox. An empty sandbox name returns endpoints across all sandboxes.
- **FR-005**: Services().Delete MUST remove a service endpoint by sandbox and service name, returning NotFound if it does not exist.

**Provider Profiles**

- **FR-006**: The SDK MUST provide a ProfileInterface accessible via `client.Providers().Profiles()` with List, Get, Import, Update, Lint, and Delete operations.
- **FR-007**: Profiles().List MUST return all available provider profiles, including both built-in and custom profiles.
- **FR-008**: Profiles().Get MUST retrieve a single provider profile by ID, returning NotFound if it does not exist.
- **FR-009**: Profiles().Import MUST accept a list of profile import items and return an import result containing diagnostics, imported profiles, and an imported flag.
- **FR-010**: Profiles().Update MUST accept a profile import item, profile ID, and expected resource version for optimistic concurrency, returning the updated profile or a concurrency error if the version is stale.
- **FR-011**: Profiles().Lint MUST validate profile import items without registering them, returning diagnostics and a valid flag.
- **FR-012**: Profiles().Delete MUST remove a custom provider profile by ID, returning NotFound if it does not exist.

**Provider Credential Refresh**

- **FR-013**: The SDK MUST provide a RefreshInterface accessible via `client.Providers().Refresh()` with GetStatus, Configure, Rotate, and Delete operations.
- **FR-014**: Refresh().GetStatus MUST return the refresh status for a provider, indicating whether refresh is configured and its current state.
- **FR-015**: Refresh().Configure MUST set up gateway-owned refresh configuration for a provider credential.
- **FR-016**: Refresh().Rotate MUST trigger an immediate credential rotation for a provider.
- **FR-017**: Refresh().Delete MUST remove the refresh configuration for a provider.

**Watch Enhancement**

- **FR-018**: WatchOptions MUST support a StopOnTerminal field. When true, the watch stream MUST close automatically after the sandbox reaches a terminal phase (Ready or Error).
- **FR-019**: The default value for StopOnTerminal MUST be false, preserving backward compatibility with existing Watch consumers.

**Fake Client Updates**

- **FR-020**: The FakeClient MUST be updated to satisfy the expanded ClientInterface, providing Services(), and extending Providers() with Profiles() and Refresh() accessors.
- **FR-021**: All new fake sub-client methods MUST return an Unimplemented error, consistent with the existing pattern for Exec and File stubs.

**Cross-Cutting**

- **FR-022**: All new domain types MUST be defined in the types package, following the existing pattern where clients import types and converters import types (not clients).
- **FR-023**: All new operations MUST return typed errors (StatusError with appropriate ErrorCode) consistent with Phase 1 error handling.
- **FR-024**: All new operations MUST be safe for concurrent use from multiple goroutines.

### Key Entities

- **ServiceEndpoint**: Represents an exposed HTTP service on a sandbox. Key attributes: sandbox name, service name, target port, domain flag, URL.
- **ProviderProfile**: A provider type template defining the credentials schema, network endpoints, binaries, and discovery configuration for a class of providers. Key attributes: ID, display name, description, category, credentials, endpoints, binaries, inference capability, discovery, resource version.
- **ProfileImportItem**: An item submitted for profile import or lint validation, containing the profile definition file content.
- **ProfileDiagnostic**: A validation finding from Import or Lint, indicating an issue with a profile definition.
- **RefreshStatus**: The current state of credential refresh for a provider, indicating whether refresh is configured and its operational status.
- **RefreshConfig**: Configuration parameters for gateway-owned credential refresh on a provider.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Operators can expose, retrieve, list, and delete sandbox service endpoints through the SDK with the same reliability as direct gateway communication.
- **SC-002**: Operators can manage the full provider profile lifecycle (list, get, import, update, lint, delete) through a single nested sub-client accessor.
- **SC-003**: Operators can configure, monitor, trigger, and remove credential refresh for any provider through the SDK.
- **SC-004**: Consumers using Watch with StopOnTerminal=true receive automatic channel closure when a sandbox reaches a terminal phase, eliminating manual phase-checking code.
- **SC-005**: The FakeClient compiles and satisfies the updated ClientInterface without breaking existing consumer tests.
- **SC-006**: All new operations produce the same typed error codes (NotFound, AlreadyExists, InvalidArgument, Unavailable, Unimplemented) as Phase 1 operations for equivalent error conditions.
- **SC-007**: All new operations pass the race detector under concurrent access from multiple goroutines.

## Assumptions

- The gateway already implements all 14 Phase 2a RPCs as defined in the proto. The SDK wraps these RPCs without adding business logic.
- Phase 1 SDK infrastructure (Client, gRPC connection, converter pattern, error handling, types package) is stable and does not require changes beyond additive extensions.
- Provider profiles are a read-heavy workload (List/Get frequent, Import/Update/Delete rare). No caching is needed in the SDK — the gateway handles caching.
- Optimistic concurrency for profile updates uses the resource_version field returned by the gateway. The SDK passes the version through without enforcing it locally.
- The RefreshConfig structure mirrors the proto's ConfigureProviderRefreshRequest fields. The exact fields are determined during planning when the proto messages are analyzed in detail.
- StopOnTerminal is implemented in the SDK's Watch client (not on the server side), by monitoring incoming events and closing the channel when a terminal phase is detected. The proto's stop_on_terminal field is also passed to the server for server-side optimization.
- Phase 2b (policy, config, SSH, TCP) is explicitly out of scope for this specification.
- Enhanced watch capabilities (log streaming, event streaming, filtering) are out of scope — deferred to brainstorm 008.
