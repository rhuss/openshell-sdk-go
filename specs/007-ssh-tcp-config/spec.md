# Feature Specification: SSH, TCP Forwarding, and Config (Phase 2b-1)

**Feature Branch**: `007-ssh-tcp-config`
**Created**: 2026-06-29
**Status**: Draft
**Input**: Phase 2b-1 SDK extension: SSH session management, TCP port forwarding, and gateway/sandbox configuration sub-clients.

## User Scenarios & Testing

### User Story 1 - Developer Creates SSH Session to Sandbox (Priority: P1)

A developer or operator needs to establish an SSH connection to a running sandbox for interactive debugging or file inspection. They create an SSH session through the SDK, which returns connection details (gateway host, port, session token) needed to configure their SSH client's ProxyCommand. When done, they revoke the session to invalidate the token.

**Why this priority**: SSH access is the most direct way to interact with a sandbox. Without it, users are limited to the Exec interface which lacks persistent sessions, port forwarding, and standard SSH tooling.

**Independent Test**: Create an SSH session for a sandbox, verify the returned session contains gateway connection details and a token. Revoke the session, verify it returns revoked=true. Revoke the same session again, verify it returns revoked=false (idempotent).

**Acceptance Scenarios**:

1. **Given** a running sandbox "debug-env", **When** the developer calls SSH().CreateSession with sandbox ID "debug-env", **Then** an SSHSession is returned containing the sandbox ID, a non-empty token, gateway host, gateway port, gateway scheme, and an expiry timestamp.
2. **Given** an active SSH session with a token, **When** the developer calls SSH().RevokeSession with that token, **Then** the call succeeds and returns revoked=true.
3. **Given** an already-revoked session token, **When** the developer calls SSH().RevokeSession with that token again, **Then** the call succeeds and returns revoked=false.
4. **Given** no sandbox with ID "missing" exists, **When** the developer calls SSH().CreateSession with sandbox ID "missing", **Then** a NotFound error is returned.

---

### User Story 2 - Developer Forwards TCP Port from Sandbox (Priority: P1)

A developer running a web server or database inside a sandbox needs to access it from their local machine. They open a TCP forward through the SDK, which returns an io.ReadWriteCloser connected to the specified port inside the sandbox. They read and write bytes through this connection until they close it.

**Why this priority**: TCP forwarding enables access to any network service running inside a sandbox (HTTP APIs, databases, debugging ports) without exposing services publicly. This is essential for local development workflows.

**Independent Test**: Forward TCP to a sandbox on port 8080, verify the returned ReadWriteCloser is non-nil. Write bytes, read response bytes. Close the connection, verify subsequent reads return an error. Attempt to forward to a non-existent sandbox, verify NotFound error.

**Acceptance Scenarios**:

1. **Given** a running sandbox "web-app" with a service on port 8080, **When** the developer calls TCP().Forward with sandbox ID "web-app" and port 8080, **Then** an io.ReadWriteCloser is returned that can send and receive bytes to/from the sandbox port.
2. **Given** an open TCP forward connection, **When** the developer writes bytes to it, **Then** the bytes are delivered to the sandbox service, and response bytes can be read back.
3. **Given** an open TCP forward connection, **When** the developer calls Close(), **Then** the underlying gRPC stream is terminated and subsequent Read/Write calls return an error.
4. **Given** no sandbox with ID "missing" exists, **When** the developer calls TCP().Forward with sandbox ID "missing", **Then** a NotFound error is returned.
5. **Given** the context is cancelled, **When** a TCP forward is active, **Then** the connection is closed and pending Read/Write calls return a context error.

---

### User Story 3 - Operator Reads and Updates Configuration (Priority: P2)

An operator managing an OpenShell gateway needs to inspect sandbox-specific settings, read gateway-global settings, and update configuration (settings or policy) at either scope. They retrieve the current configuration, modify a setting, and apply the update.

**Why this priority**: Configuration management is needed for governance and operational control, but most users can work with default settings initially. Operators who manage fleets of sandboxes need this for customization and policy enforcement.

**Independent Test**: Get sandbox config for a sandbox, verify settings and policy are returned. Get gateway config, verify global settings are returned. Update a setting, verify the response contains a new settings revision. Attempt to get config for a non-existent sandbox, verify NotFound error.

**Acceptance Scenarios**:

1. **Given** a running sandbox "worker-1", **When** the operator calls Config().GetSandbox with sandbox ID "worker-1", **Then** a SandboxConfig is returned containing the policy, policy version, policy hash, effective settings, and config revision.
2. **Given** a running gateway, **When** the operator calls Config().GetGateway, **Then** a GatewayConfig is returned containing the global settings map and settings revision.
3. **Given** a running sandbox "worker-1", **When** the operator calls Config().Update with a sandbox-scoped setting change (using sandbox name "worker-1"), **Then** an UpdateResult is returned containing the new version, policy hash, and settings revision.
4. **Given** a running gateway, **When** the operator calls Config().Update with a global-scoped setting change (Global=true), **Then** an UpdateResult is returned reflecting the global settings revision.
5. **Given** no sandbox with ID "missing" exists, **When** the operator calls Config().GetSandbox with sandbox ID "missing", **Then** a NotFound error is returned.

---

### User Story 4 - Fake Client Stubs for Testing (Priority: P3)

A developer writing tests for code that uses the SSH, TCP, or Config sub-clients needs fake implementations that compile against the updated ClientInterface. The fake stubs return Unimplemented errors, consistent with the Phase 2a pattern for new sub-clients.

**Why this priority**: The fake client must compile with the updated interfaces, but full fake behavior (in-memory SSH session store, TCP stream simulation) is deferred. Stubs returning Unimplemented are sufficient for consumers who mock these interfaces directly in their tests.

**Independent Test**: Create a fake client, call SSH().CreateSession, verify it returns an Unimplemented error. Same for TCP().Forward and Config().GetGateway. Verify the fake client still compiles against ClientInterface.

**Acceptance Scenarios**:

1. **Given** a fake client, **When** the developer calls SSH().CreateSession, **Then** an Unimplemented error is returned.
2. **Given** a fake client, **When** the developer calls TCP().Forward, **Then** an Unimplemented error is returned.
3. **Given** a fake client, **When** the developer calls Config().GetSandbox, Config().GetGateway, or Config().Update, **Then** an Unimplemented error is returned for each.
4. **Given** the updated ClientInterface with SSH(), TCP(), and Config() accessors, **When** a fake.Client is assigned to a ClientInterface variable, **Then** it compiles without errors (compile-time interface check).

---

### Edge Cases

- What happens when an SSH session token expires before RevokeSession is called? The revoke call should succeed with revoked=false (already expired).
- What happens when the gRPC stream underlying a TCP forward is interrupted by a network error? Read/Write should return an appropriate error (Unavailable or the underlying gRPC error).
- What happens when Config().Update is called with an outdated expected_resource_version? A Conflict or FailedPrecondition error should be returned.
- What happens when TCP().Forward is called with port 0? An InvalidArgument error should be returned.
- What happens when the client is closed while a TCP forward is active? The ReadWriteCloser should return an Unavailable error on subsequent operations.

## Requirements

### Functional Requirements

- **FR-001**: SDK MUST provide an SSHInterface with CreateSession and RevokeSession methods as a top-level sub-client on ClientInterface.
- **FR-002**: SSHInterface.CreateSession MUST accept a context and sandbox ID and return an SSHSession containing sandbox ID, token, gateway host, gateway port, gateway scheme, optional host key fingerprint, and expiry timestamp.
- **FR-003**: SSHInterface.RevokeSession MUST accept a context and session token and return whether the session was revoked (bool).
- **FR-004**: SDK MUST provide a TCPInterface with a Forward method as a top-level sub-client on ClientInterface.
- **FR-005**: TCPInterface.Forward MUST accept a context, sandbox ID, and remote port (1-65535), and return an io.ReadWriteCloser wrapping the bidirectional gRPC stream. Port values outside 1-65535 MUST be rejected client-side with an InvalidArgument error before opening the gRPC stream.
- **FR-006**: The TCP forward ReadWriteCloser MUST close the underlying gRPC stream when Close() is called.
- **FR-007**: The TCP forward MUST respect context cancellation and close the stream when the context is cancelled.
- **FR-007a**: The TcpForwardInit proto has an optional `service_id` field for audit/correlation. The SDK MUST NOT expose this field in v1; it is set to empty string internally. If needed later, it can be added as a functional option.
- **FR-008**: SDK MUST provide a ConfigInterface with GetSandbox, GetGateway, and Update methods as a top-level sub-client on ClientInterface.
- **FR-009**: ConfigInterface.GetSandbox MUST accept a context and sandbox ID and return a SandboxConfig containing the policy, policy version, policy hash, effective settings (as a map of setting name to EffectiveSetting), config revision, policy source, global policy version, and provider environment revision.
- **FR-010**: ConfigInterface.GetGateway MUST return a GatewayConfig containing the global settings map (setting name to SettingValue) and settings revision.
- **FR-011**: ConfigInterface.Update MUST accept a ConfigUpdate struct supporting sandbox-scoped updates (with sandbox name as the canonical lookup key, policy, single setting upsert/delete, merge operations, and expected resource version) and global-scoped updates (with Global=true). Note: UpdateConfigRequest uses `name` (sandbox name), not `sandbox_id`, unlike the other RPCs in this spec.
- **FR-012**: ConfigInterface.Update MUST return an UpdateResult containing the assigned version, policy hash, settings revision, and a deleted flag.
- **FR-013**: All new domain types MUST be defined in `v1/types/` with no proto dependency.
- **FR-014**: All proto-to-SDK conversions MUST deep-copy maps and slices at boundaries.
- **FR-015**: All operations MUST be safe for concurrent use.
- **FR-016**: All operations MUST return typed StatusError with appropriate ErrorCode for error paths.
- **FR-017**: Fake client MUST implement SSH(), TCP(), and Config() accessors returning stubs that produce Unimplemented errors.
- **FR-018**: Fake client MUST pass the compile-time interface check (`var _ v1.ClientInterface = (*Client)(nil)`).

### Key Entities

- **SSHSession**: Represents an SSH session created for a sandbox. Contains sandbox ID, session token, gateway connection details (host, port, scheme), optional host key fingerprint, and expiry timestamp.
- **SandboxConfig**: Represents the full configuration state of a sandbox. Contains policy, policy version, policy hash, effective settings, config revision, policy source, global policy version, and provider environment revision.
- **GatewayConfig**: Represents gateway-global settings. Contains a settings map and settings revision.
- **ConfigUpdate**: Represents a configuration mutation. Supports sandbox-scoped updates (name, policy, setting key/value, delete flag, merge operations, expected resource version) and global-scoped updates (Global flag).
- **UpdateResult**: Represents the result of a configuration update. Contains assigned version, policy hash, settings revision, and deleted flag.
- **SettingValue**: A typed setting value supporting string, bool, int64, and bytes variants.
- **EffectiveSetting**: A setting value with its resolved scope (sandbox or global).
- **SettingScope**: Enum indicating whether a setting is controlled at sandbox or global level (SANDBOX or GLOBAL).
- **PolicySource**: Enum indicating the source of the policy payload in a SandboxConfig response (SANDBOX or GLOBAL).

## Success Criteria

### Measurable Outcomes

- **SC-001**: All 6 new methods (CreateSession, RevokeSession, Forward, GetSandbox, GetGateway, Update) are callable through the SDK with correct parameter passing and response mapping.
- **SC-002**: All tests pass with the race detector enabled (`go test -race`).
- **SC-003**: The fake client compiles against the updated ClientInterface without modification to existing fake sub-clients.
- **SC-004**: TCP forward Read/Write operations deliver bytes end-to-end through the gRPC stream with no data loss or corruption.
- **SC-005**: `make ci` (lint + build + test) passes with zero violations.

## Clarifications

### Session 2026-06-29

- Q: How should SandboxPolicy be represented in SDK types? → A: Opaque bytes (`[]byte`) for v1. The proto type is large and complex; consumers who need to inspect policy content can unmarshal the raw bytes themselves.
- Q: Should TCP Forward validate port client-side or rely on server-side validation? → A: Client-side validation. Port must be 1-65535, checked before opening the gRPC stream. Returns InvalidArgument for out-of-range values.
- Q: How should ConfigUpdate merge operations (PolicyMergeOperation) be represented? → A: Opaque for v1. Pass merge operations as raw proto bytes. The full PolicyMergeOperation type hierarchy is policy-domain-specific and better addressed in the Phase 2b-2 policy spec.

## Assumptions

- The OpenShell gateway implements all 6 RPCs covered by this spec (CreateSshSession, RevokeSshSession, ForwardTcp, GetSandboxConfig, GetGatewayConfig, UpdateConfig).
- The TcpForwardInit message uses the TcpRelayTarget variant (not SshRelayTarget) for TCP forwarding. The SDK constructs the init frame internally.
- SandboxPolicy is represented as opaque bytes (`[]byte`) in v1/types/. The proto type is large and complex; full SDK domain mapping is deferred. Consumers who need to inspect policy content can unmarshal the raw bytes.
- ConfigUpdate merge operations (PolicyMergeOperation) are represented as raw proto bytes for v1. The full type hierarchy (AddNetworkRule, RemoveNetworkEndpoint, etc.) is addressed in the Phase 2b-2 policy spec.
- The existing Phase 1 and Phase 2a sub-clients, types, converters, and fake stubs remain unchanged.
- No new external dependencies are required. The SDK continues to use only gRPC and Go stdlib.
