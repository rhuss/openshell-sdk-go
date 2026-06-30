# Feature Specification: Sandbox Name Resolution

**Feature Branch**: `013-sandbox-name-resolution`
**Created**: 2026-06-30
**Status**: Draft
**Input**: Brainstorm `brainstorm/014-name-id-consistency.md`, Issue [#15](https://github.com/rhuss/openshell-sdk-go/issues/15)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Execute Command by Sandbox Name (Priority: P1)

An SDK consumer creates a sandbox by name, then wants to execute a
command in it. Today they must call `Sandboxes().Get(name)` to obtain
the ID, then pass `sandbox.ID` to `Exec().Run()`. With this change,
they pass the sandbox name directly to `Exec().Run()` and the SDK
resolves internally.

**Why this priority**: Exec is the most commonly used ID-based method.
Fixing it eliminates the most frequent source of confusion for new SDK
consumers.

**Independent Test**: Can be tested by calling `Exec().Run()` with a
sandbox name and verifying the correct proto request is sent with the
resolved sandbox ID.

**Acceptance Scenarios**:

1. **Given** an SDK client and a sandbox named "my-sandbox", **When** the
   consumer calls `Exec().Run(ctx, "my-sandbox", cmd)`, **Then** the SDK
   resolves the name to the sandbox's ID via `Sandboxes().Get()` and sends
   the correct `ExecSandboxRequest` with the resolved `sandbox_id`.
2. **Given** an SDK client and a non-existent sandbox name, **When** the
   consumer calls `Exec().Run(ctx, "no-such-sandbox", cmd)`, **Then** the
   SDK returns a `NotFound` error from the resolution step before
   attempting the exec RPC.

---

### User Story 2 - Upload File by Sandbox Name (Priority: P1)

An SDK consumer wants to upload a file to a sandbox using only its name.
Today `Files().Upload()` requires the sandbox ID. With this change, the
consumer passes the sandbox name directly.

**Why this priority**: File operations are the second most common
ID-based interaction after exec. Resolving names here completes the
"create sandbox, then use it" workflow without manual ID tracking.

**Independent Test**: Can be tested by calling `Files().Upload()` with a
sandbox name and verifying the underlying SSH session request uses the
resolved sandbox ID.

**Acceptance Scenarios**:

1. **Given** a sandbox named "dev-box", **When** the consumer calls
   `Files().Upload(ctx, "dev-box", content, "/path")`, **Then** the SDK
   resolves the name to ID and sends the correct `CreateSshSessionRequest`
   with the resolved `sandbox_id`.
2. **Given** a sandbox named "dev-box", **When** the consumer calls
   `Files().Download(ctx, "dev-box", "/remote/path", "/local/path")`,
   **Then** the SDK resolves the name to ID and sends the correct request
   with the resolved `sandbox_id`.

---

### User Story 3 - Watch Sandbox Events Correctly (Priority: P1)

An SDK consumer calls `Sandboxes().Watch(ctx, "my-sandbox")`. Today the
SDK passes the name directly into the `WatchSandboxRequest.Id` proto
field, which expects an ID. With this change, the SDK resolves the name
to ID before constructing the watch request.

**Why this priority**: This is a correctness bug. Passing a name into an
ID field may silently fail or return wrong results depending on server
behavior.

**Independent Test**: Can be tested by calling `Watch()` with a sandbox
name and verifying the `WatchSandboxRequest.Id` field contains the
resolved sandbox ID, not the name.

**Acceptance Scenarios**:

1. **Given** a sandbox named "my-sandbox" with ID "sb-abc-123", **When**
   the consumer calls `Watch(ctx, "my-sandbox")`, **Then** the SDK sends
   `WatchSandboxRequest{Id: "sb-abc-123"}`, not `{Id: "my-sandbox"}`.
2. **Given** a non-existent sandbox name, **When** the consumer calls
   `Watch(ctx, "ghost")`, **Then** the SDK returns a `NotFound` error
   before opening the watch stream.

---

### User Story 4 - Forward TCP by Sandbox Name (Priority: P2)

An SDK consumer wants to forward a TCP port using the sandbox name.
Today `TCP().Forward()` requires the sandbox ID. With this change, the
consumer passes the name.

**Why this priority**: TCP forwarding is less commonly used than exec or
files, but the inconsistency is confusing when it appears.

**Independent Test**: Can be tested by calling `TCP().Forward()` with a
sandbox name and verifying the `TcpForwardInit` uses the resolved ID.

**Acceptance Scenarios**:

1. **Given** a sandbox named "tunnel-box", **When** the consumer calls
   `TCP().Forward(ctx, "tunnel-box", 8080)`, **Then** the SDK resolves the
   name and sends the correct `TcpForwardInit` with the resolved
   `sandbox_id`.

---

### User Story 5 - Get Sandbox Config by Name (Priority: P2)

An SDK consumer wants to retrieve sandbox configuration using the sandbox
name. Today `Config().GetSandbox()` requires the sandbox ID. With this
change, the consumer passes the name.

**Why this priority**: Config retrieval is an infrequent operation but
the inconsistency breaks the "everything uses names" expectation.

**Independent Test**: Can be tested by calling `Config().GetSandbox()`
with a sandbox name and verifying the `GetSandboxConfigRequest` uses the
resolved ID.

**Acceptance Scenarios**:

1. **Given** a sandbox named "my-sandbox", **When** the consumer calls
   `Config().GetSandbox(ctx, "my-sandbox")`, **Then** the SDK resolves the
   name and sends the correct `GetSandboxConfigRequest` with the resolved
   `sandbox_id`.

---

### Edge Cases

- What happens when the sandbox name resolves successfully but the
  underlying ID-based RPC fails? The SDK returns the RPC error, not a
  resolution error.
- What happens when multiple rapid calls use the same sandbox name? Each
  call independently resolves the name (no caching). This is correct but
  may be slower for batch operations.
- What happens when a sandbox is deleted between resolution and the
  subsequent RPC call? The RPC returns a `NotFound` error, which the SDK
  surfaces as-is.
- How does `SSH.CreateSession` behave? It remains ID-based. Consumers
  should use `SSH().Tunnel()` which resolves names internally.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: `ExecInterface.Run()`, `ExecInterface.Stream()`, and
  `ExecInterface.Interactive()` MUST accept a sandbox name (not ID) and
  resolve it to a sandbox ID internally before sending the proto request.
- **FR-002**: `TCPInterface.Forward()` MUST accept a sandbox name and
  resolve it to a sandbox ID internally.
- **FR-003**: `FileInterface.Upload()` and `FileInterface.Download()`
  MUST accept a sandbox name and resolve it to a sandbox ID internally.
- **FR-004**: `ConfigInterface.GetSandbox()` MUST accept a sandbox name
  and resolve it to a sandbox ID internally.
- **FR-005**: `SandboxInterface.Watch()` MUST resolve the sandbox name
  to an ID before constructing the `WatchSandboxRequest`, passing the
  resolved ID (not the name) to the proto `Id` field. Note: unlike the
  other affected methods, `Watch` already uses a `name` parameter, so
  this is a resolution-logic addition, not a parameter rename.
- **FR-006**: Name resolution MUST use the existing
  `SandboxInterface.Get()` method to look up the sandbox by name and
  extract its ID.
- **FR-007**: If name resolution fails (sandbox not found), the SDK MUST
  return the resolution error without attempting the underlying RPC.
- **FR-008**: `SSHInterface.CreateSession()` MUST remain ID-based. Its
  doc comment MUST recommend using `SSH().Tunnel()` for name-based access.
- **FR-009**: Fake client implementations MUST update method signatures
  to match the new interfaces (sandbox name parameters) but are NOT
  required to perform actual name-to-ID resolution.
- **FR-010**: All sub-clients that perform name resolution MUST receive
  a `SandboxInterface` dependency at construction time, following the
  pattern established by `SSHClient`.

### Key Entities

- **Sandbox**: Has both a human-readable `Name` (set at creation) and an
  opaque `ID` (assigned by the server). The name is stable and unique
  within a gateway; the ID is the internal identifier used by low-level
  RPCs.
- **SandboxInterface**: The sub-client responsible for sandbox lifecycle.
  Used as a dependency by other sub-clients to resolve sandbox names to
  IDs via `Get()`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Every public SDK method (except `SSH.CreateSession`)
  accepts a sandbox name as its sandbox identifier parameter. No public
  method requires the caller to pass a sandbox ID.
- **SC-002**: All existing unit tests pass after the parameter rename,
  with updated test call sites.
- **SC-003**: New unit tests verify name-to-ID resolution for at least
  one method per affected sub-client (Exec, TCP, File, Config, Watch).
- **SC-004**: The `Watch` method sends the resolved sandbox ID in
  `WatchSandboxRequest.Id`, not the sandbox name (bug fix verified by
  test assertion).
- **SC-005**: `make ci` passes cleanly (lint, build, test).

## Assumptions

- The SDK is pre-1.0, so breaking changes to public interface signatures
  are acceptable without a deprecation cycle.
- The `SandboxInterface.Get()` RPC is lightweight and the extra round-trip
  per call is acceptable for correctness and API consistency.
- No name-to-ID caching is needed in this iteration. Consumers who need
  batch performance can hold a reference to the sandbox object returned
  by `Create()` or `Get()`.
- The upstream proto field naming inconsistency is a separate concern,
  tracked for a future upstream issue on NVIDIA/OpenShell.
