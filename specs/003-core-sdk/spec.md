# Feature Specification: Core SDK (Phase 1)

**Feature Branch**: `003-core-sdk`
**Created**: 2026-06-27
**Status**: Draft
**Input**: Core SDK Phase 1 - Kubernetes client-go style SDK with sub-client interfaces for Sandbox, Provider, Exec, File, and Health. Context from brainstorm/004-core-sdk.md

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Sandbox Lifecycle Management (Priority: P1)

A Go developer building a Kubernetes operator creates, inspects, lists, and tears down sandboxes through the SDK. They construct a client with gateway address and credentials, then call sandbox operations that return familiar Go types. The developer never imports proto-generated packages or deals with gRPC internals directly.

**Why this priority**: Sandbox lifecycle is the most fundamental capability. Every consumer of the SDK needs to create and manage sandboxes. Without this, no other operation is possible.

**Independent Test**: Can be fully tested by creating a client, performing CRUD operations on sandboxes, and verifying return types are idiomatic Go structs with standard field types (string, time.Time, etc.).

**Acceptance Scenarios**:

1. **Given** a valid gateway address and credentials, **When** the developer calls NewClient with a Config struct, **Then** a client is returned without error and is ready for use.
2. **Given** a connected client, **When** the developer creates a sandbox with a name and image, **Then** a Sandbox struct is returned containing the name, phase, image, and creation timestamp.
3. **Given** an existing sandbox, **When** the developer retrieves it by name, **Then** the same Sandbox struct is returned with current phase status.
4. **Given** multiple sandboxes exist, **When** the developer lists sandboxes, **Then** a typed list is returned containing all sandbox entries.
5. **Given** an existing sandbox, **When** the developer deletes it by name, **Then** the sandbox is removed and subsequent Get calls return a "not found" error.
6. **Given** a newly created sandbox in "Pending" phase, **When** the developer calls WaitReady with a context timeout, **Then** the call blocks until the sandbox reaches "Ready" phase or the context deadline is exceeded.

---

### User Story 2 - Command Execution in Sandboxes (Priority: P1)

A Go developer runs shell commands inside a sandbox. For simple commands, they call a run method that returns exit code, stdout, and stderr as a single result. For long-running commands, they stream output chunk by chunk. For interactive sessions (like a shell), they get a bidirectional connection for stdin/stdout.

**Why this priority**: Command execution is the core use case for sandboxes. Operators and tools like cc-deck need to run commands as their primary interaction with sandboxes.

**Independent Test**: Can be fully tested by creating a sandbox, running a simple command (e.g., `echo hello`), and verifying the exit code is 0 and stdout contains "hello".

**Acceptance Scenarios**:

1. **Given** a ready sandbox, **When** the developer runs a command with Run, **Then** an ExecResult is returned containing the integer exit code, stdout bytes, and stderr bytes.
2. **Given** a ready sandbox, **When** the developer runs a long-running command with Stream, **Then** they receive output chunks incrementally, each tagged with stdout or stderr, and can retrieve the final exit code after the command completes.
3. **Given** a ready sandbox, **When** the developer starts an interactive session, **Then** they can send input and receive output bidirectionally until the session is closed.
4. **Given** a non-existent sandbox name, **When** the developer attempts to run a command, **Then** a "not found" typed error is returned.

---

### User Story 3 - Provider Management (Priority: P2)

A Go developer configures compute providers (cloud accounts, container runtimes) that back sandboxes. They create, update, list, and delete providers. For operator use cases, they use an idempotent "ensure" operation that creates a provider if it does not exist or updates it if it does, avoiding create-then-check-then-update boilerplate.

**Why this priority**: Providers are required before sandboxes can be created. Operators need provider management for automated cluster setup, but manual sandbox users may use pre-configured providers.

**Independent Test**: Can be fully tested by creating a provider, retrieving it, ensuring it (idempotent update), and deleting it.

**Acceptance Scenarios**:

1. **Given** a connected client, **When** the developer creates a provider with a name and configuration, **Then** a Provider struct is returned with the provider details.
2. **Given** an existing provider, **When** the developer calls Ensure with the same name but updated configuration, **Then** the provider is updated in place and the updated Provider struct is returned.
3. **Given** an existing provider, **When** the developer calls Ensure with the same name and identical configuration, **Then** the existing provider is returned without modification.
4. **Given** a provider not yet linked to a sandbox, **When** the developer attaches the provider to a sandbox, **Then** the sandbox can use that provider's compute resources.

---

### User Story 4 - File Transfer (Priority: P2)

A Go developer uploads files from the local machine to a sandbox and downloads files from a sandbox to the local machine. The API is simple: source path, destination path, done. Large files are handled transparently without the developer managing chunking or SSH tunnels.

**Why this priority**: File transfer is essential for tools like cc-deck that need to move configuration files, scripts, and artifacts in and out of sandboxes. It is not needed by all consumers.

**Independent Test**: Can be fully tested by uploading a file to a sandbox, downloading it back, and comparing contents.

**Acceptance Scenarios**:

1. **Given** a ready sandbox and a local file, **When** the developer calls Upload with local and remote paths, **Then** the file is transferred to the sandbox at the specified path.
2. **Given** a ready sandbox with an existing file, **When** the developer calls Download with remote and local paths, **Then** the file is retrieved and written to the local path.
3. **Given** a non-existent local file path, **When** the developer calls Upload, **Then** an appropriate error is returned without contacting the gateway.

---

### User Story 5 - Health Checking (Priority: P3)

A Go developer checks whether the gateway is reachable and healthy before performing operations. This is used in readiness probes, startup checks, and operator reconciliation loops.

**Why this priority**: Health checking is simple but important for production reliability. It is low complexity but enables operators to gate their logic on gateway availability.

**Independent Test**: Can be fully tested by calling the health check method against a running gateway and verifying no error is returned.

**Acceptance Scenarios**:

1. **Given** a connected client with a reachable gateway, **When** the developer calls Health Check, **Then** no error is returned.
2. **Given** a connected client with an unreachable gateway, **When** the developer calls Health Check, **Then** an "unavailable" typed error is returned.

---

### User Story 6 - Watching Sandbox State Changes (Priority: P3)

A Go developer watches for sandbox state changes in real time, receiving events when sandboxes are created, modified, or deleted. The watch returns a channel of typed events, following the same pattern as Kubernetes watch interfaces, so operator authors can integrate it directly into their reconciliation loops.

**Why this priority**: Watch is essential for operators that need to react to sandbox state changes, but is not needed for simple CLI tools or one-shot scripts.

**Independent Test**: Can be fully tested by starting a watch, creating a sandbox in a separate goroutine, and verifying that an "ADDED" event is received on the watch channel.

**Acceptance Scenarios**:

1. **Given** a connected client, **When** the developer starts a watch on sandboxes, **Then** a watch handle is returned with a channel that emits typed events.
2. **Given** an active watch, **When** a sandbox is created, **Then** an event with type "ADDED" and the sandbox data is received on the channel.
3. **Given** an active watch, **When** the developer calls Stop on the watch handle, **Then** the channel is closed and no further events are delivered.
4. **Given** an active watch, **When** the gateway connection is interrupted, **Then** an event with type "ERROR" is received on the channel.

---

### User Story 7 - Typed Error Handling (Priority: P1)

A Go developer handles errors from SDK operations using typed error helpers (IsNotFound, IsAlreadyExists, IsUnavailable, IsPermissionDenied) rather than parsing error strings or importing gRPC status packages. Error handling follows the same pattern as Kubernetes apimachinery errors.

**Why this priority**: Proper error handling is critical for all consumers. Without typed errors, developers must resort to string matching or import gRPC internals, both of which are fragile and violate the SDK's abstraction boundary.

**Independent Test**: Can be fully tested by triggering known error conditions (get non-existent resource, create duplicate) and verifying the typed error helpers return true.

**Acceptance Scenarios**:

1. **Given** a Get call for a non-existent sandbox, **When** the error is checked with IsNotFound, **Then** it returns true.
2. **Given** a Create call for an already-existing provider, **When** the error is checked with IsAlreadyExists, **Then** it returns true.
3. **Given** a call to an unreachable gateway, **When** the error is checked with IsUnavailable, **Then** it returns true.
4. **Given** a call with insufficient credentials, **When** the error is checked with IsPermissionDenied, **Then** it returns true.
5. **Given** any SDK error, **When** it is printed with Error(), **Then** it produces a human-readable message including the error code and details.

---

### Edge Cases

- What happens when the gateway disconnects mid-operation? The SDK must return a typed "unavailable" error, not panic or hang indefinitely.
- What happens when context is cancelled during a streaming exec? The stream must close cleanly and return a context cancellation error.
- What happens when WaitReady is called on a sandbox that transitions to "Failed" instead of "Ready"? The call must return an error indicating the sandbox failed, not block forever.
- What happens when Upload is called with a path that exists as a directory in the sandbox? An appropriate error must be returned.
- What happens when a file transfer is interrupted mid-upload or mid-download (e.g., network failure)? The SDK must return an error and must not leave a partial file at the destination path.
- What happens when Watch receives a malformed event from the gateway? An ERROR event must be emitted on the channel rather than crashing the consumer.
- What happens when the client is used after Close is called? All subsequent calls must return an error indicating the client is closed.

## Clarifications

### Session 2026-06-27

- Q: Is the Client safe for concurrent use from multiple goroutines? → A: Yes, the Client and all sub-clients MUST be safe for concurrent use. This is standard Go convention for client libraries.
- Q: What is explicitly out of scope for Phase 1? → A: Service exposure, provider profiles, credential refresh, policy management, configuration management, SSH session management, TCP forwarding, logging RPCs, and all supervisor/internal RPCs. These are deferred to Phase 2a/2b.
- Q: Does the SDK validate resource names client-side? → A: No. The SDK passes names through to the gateway, which enforces naming rules. Invalid names result in a typed InvalidArgument error from the gateway.
- Q: Does NewClient connect eagerly or lazily? → A: Eagerly. NewClient establishes the gRPC connection during construction and returns an error if the gateway is unreachable. Consumers know immediately whether the connection is valid.
- Q: Should the SDK provide logging or tracing hooks? → A: The SDK accepts an optional structured logger interface for debug/trace output. No logging occurs by default. Detailed observability is deferred to a future phase.

## Out of Scope

The following capabilities are explicitly excluded from Phase 1 and deferred to later phases:

- Service exposure (ExposeService, GetService, ListServices, DeleteService)
- Provider profiles (ListProfiles, GetProfile, ImportProfiles, UpdateProfiles, DeleteProfile)
- Provider credential refresh (GetRefreshStatus, ConfigureRefresh, RotateCredential, DeleteRefresh)
- Policy management (GetStatus, List, GetDraft, draft chunk operations, history)
- Configuration management (GetSandboxConfig, GetGatewayConfig, Update)
- SSH session management (CreateSession, RevokeSession)
- TCP forwarding (ForwardTCP)
- Logging RPCs
- Supervisor/internal RPCs (ConnectSupervisor, RelayStream, PushSandboxLogs, etc.)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: SDK MUST provide a single entry point for creating a client from a configuration struct containing gateway address, TLS settings, authentication, timeout, and retry policy.
- **FR-002**: SDK MUST organize operations into domain-specific sub-clients (Sandbox, Provider, Exec, File, Health) accessible via accessor methods on the main client.
- **FR-003**: All sub-clients MUST be defined as Go interfaces so consumers can create mock implementations for testing without a running gateway.
- **FR-004**: SDK MUST define its own domain types (Sandbox, Provider, ExecResult, etc.) that are idiomatic Go structs using standard library types (string, time.Time, int). Proto-generated types MUST NOT appear in the public API.
- **FR-005**: SDK MUST map gRPC status codes to SDK-specific typed errors with helper functions (IsNotFound, IsAlreadyExists, IsUnavailable, IsPermissionDenied, IsDeadlineExceeded) so consumers never need to import gRPC packages.
- **FR-006**: SDK MUST provide a Watch operation that returns a channel of typed events (Added, Modified, Deleted, Error) compatible with the Kubernetes watch pattern.
- **FR-007**: SDK MUST support three command execution modes: synchronous (Run returns full result), streaming (Stream returns chunks), and interactive (bidirectional I/O).
- **FR-008**: SDK MUST provide file Upload and Download operations that accept local and remote file paths and handle transfer details internally.
- **FR-009**: SDK MUST provide a Health Check operation that validates gateway connectivity.
- **FR-010**: SDK MUST support provider Ensure (idempotent create-or-update) as a first-class operation.
- **FR-011**: SDK MUST support sandbox-provider attachment and detachment (AttachProvider, DetachProvider, ListProviders).
- **FR-012**: SDK MUST support WaitReady for sandboxes that blocks until the sandbox reaches a ready state or the context deadline is exceeded.
- **FR-013**: SDK MUST provide a Close method on the client that releases all resources and causes subsequent calls to return an error.
- **FR-014**: All operations MUST accept a context.Context as the first parameter for cancellation and deadline propagation.
- **FR-015**: SDK MUST convert between internal proto types and public domain types using internal converter functions that are not exported.
- **FR-016**: SDK MUST namespace its public API under a version path (v1) so future API versions can coexist without breaking existing consumers.
- **FR-017**: The Client and all sub-clients MUST be safe for concurrent use from multiple goroutines.
- **FR-018**: NewClient MUST establish the gateway connection eagerly and return an error if the gateway is unreachable, so consumers get immediate feedback on connectivity.
- **FR-019**: SDK MUST accept an optional structured logger interface for debug and trace output. No logging occurs when no logger is configured.

### Key Entities

- **Client**: The top-level entry point. Holds connection state and provides access to sub-clients. Constructed from a Config struct.
- **Sandbox**: A compute environment with a name, phase (Pending, Starting, Ready, Stopped, Failed), image, and creation timestamp. Supports CRUD, watch, and wait-ready operations.
- **Provider**: A compute backend (cloud account, container runtime) with a name and configuration. Supports CRUD and idempotent ensure.
- **ExecResult**: The outcome of a synchronous command execution, containing exit code, stdout, and stderr.
- **ExecStream**: An incremental output stream for long-running commands, delivering chunks tagged by stream type (stdout/stderr) with a final exit code.
- **WatchEvent**: A typed event (Added, Modified, Deleted, Error) carrying a resource object, delivered via a Go channel.
- **StatusError**: A typed error carrying an error code, message, and details map, with helper functions for common conditions.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A Go developer with Kubernetes client-go experience can create a client, manage a sandbox lifecycle (create, wait-ready, exec, delete), and handle errors using typed helpers in under 30 lines of code.
- **SC-002**: All sub-client interfaces are mockable: a consumer can write a complete unit test suite for their operator without a running gateway by implementing the SDK interfaces.
- **SC-003**: No proto-generated types or gRPC packages appear in any import path that a consumer needs. Consumer code never directly imports gRPC or proto packages to use the SDK's public API.
- **SC-004**: 100% of the ~20 Phase 1 RPCs (Sandbox CRUD + Watch + WaitReady, Provider CRUD + Ensure, Exec Run + Stream + Interactive, File Upload + Download, Health Check, Sandbox-Provider attachment) are covered by the SDK.
- **SC-005**: Every sub-client operation is tested against an in-process mock gRPC server with at least one positive and one negative (error) test case.
- **SC-006**: The Watch interface delivers events within 1 second of the gateway emitting them under normal network conditions.
- **SC-007**: A developer familiar with client-go patterns (sub-client accessors, Options structs, typed errors, Watch channels) recognizes the API shape immediately without reading documentation.

## Assumptions

- The upstream OpenShell proto definitions are stable for the RPCs covered in Phase 1. Proto changes during implementation will be handled by re-running proto:sync and proto:gen.
- A running OpenShell gateway is available for integration tests but not required for unit tests. Unit tests use in-process mock gRPC servers.
- File transfer uses SSH tunnels internally, but this is an implementation detail hidden from SDK consumers. The public API only exposes Upload and Download with file paths.
- TLS configuration defaults to auto-discovery. When TLS fields are nil, the SDK attempts to connect with system-default TLS. An explicit Insecure flag is provided for localhost development only.
- Authentication is pluggable via an AuthProvider interface, but the specific auth implementations (OIDC, mTLS, token-based) are out of scope for Phase 1. Phase 1 ships with a no-auth option and a static-token option.
- The SDK targets Go 1.23+ as the minimum supported version.
- The module path is `github.com/rhuss/openshell-sdk-go` and the public API package is `openshell/v1`.
- Retry policy is opt-in. When RetryPolicy is nil in Config, no retries are attempted. The retry implementation is an internal concern.
