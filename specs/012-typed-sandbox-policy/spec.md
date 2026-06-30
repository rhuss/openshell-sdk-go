# Feature Specification: Typed SandboxPolicy Domain Type

**Feature Branch**: `012-typed-sandbox-policy`
**Created**: 2026-06-30
**Status**: Draft
**Input**: Brainstorm 013-typed-sandbox-policy, Issue #11

## User Scenarios & Testing

### User Story 1 - Set Initial Policy at Sandbox Creation (Priority: P1)

An SDK consumer creates a sandbox with an initial security policy that
controls filesystem access, process execution, and network rules. The
consumer constructs a typed `SandboxPolicy` struct, sets it on
`SandboxSpec.Policy`, and calls `Sandboxes().Create()`. The gateway
receives the policy and applies it when provisioning the sandbox.

**Why this priority**: This is the primary gap identified in issue #11.
Without this, SDK consumers cannot pass a policy at creation time and
must rely on separate runtime policy management after the sandbox starts.

**Independent Test**: Can be tested by constructing a `SandboxPolicy`
with all sub-types populated, setting it on `SandboxSpec`, converting
to proto, and verifying the proto `SandboxSpec.policy` field is fully
populated with matching values.

**Acceptance Scenarios**:

1. **Given** a `SandboxSpec` with a fully populated `Policy` field
   (all sub-types set), **When** the spec is converted to proto via
   `SandboxSpecToProto`, **Then** the proto `SandboxSpec.policy` field
   contains all values with correct mapping.
2. **Given** a `SandboxSpec` with `Policy` set to nil, **When** the
   spec is converted to proto, **Then** the proto `SandboxSpec.policy`
   field is nil (no policy sent to server).
3. **Given** a proto `SandboxSpec` response with a populated `policy`
   field, **When** converted via `sandboxSpecFromProto`, **Then** the
   SDK `SandboxSpec.Policy` is fully populated with matching values.

---

### User Story 2 - Replace Full Policy at Runtime (Priority: P1)

An SDK consumer replaces the entire active policy on a running sandbox
using `Config().Update()` with a typed `*SandboxPolicy` instead of
opaque `[]byte`. The consumer constructs the policy programmatically,
sets it on `ConfigUpdate.Policy`, and calls `Update()`.

**Why this priority**: The existing `ConfigUpdate.Policy` field is
`[]byte`, forcing consumers to manually serialize policy structs. This
is the second consumer of the same `SandboxPolicy` type.

**Independent Test**: Can be tested by constructing a `ConfigUpdate`
with a typed `*SandboxPolicy`, converting to proto, and verifying the
proto `UpdateConfigRequest.policy` field matches.

**Acceptance Scenarios**:

1. **Given** a `ConfigUpdate` with a typed `*SandboxPolicy`, **When**
   converted to proto, **Then** the proto `UpdateConfigRequest.policy`
   is fully populated.
2. **Given** a `ConfigUpdate` with `Policy` set to nil, **When**
   converted to proto, **Then** the proto policy field is nil
   (no policy replacement requested).

---

### User Story 3 - Read Policy from Revision History (Priority: P2)

An SDK consumer retrieves policy revision history via
`Policy().List()` and inspects the policy content of each revision as
a typed `*SandboxPolicy` instead of opaque `[]byte`. The consumer can
programmatically examine filesystem rules, process settings, and
network policies from historical revisions.

**Why this priority**: This completes the typed policy surface across
all four usage contexts (create, update, read-config, read-history).
Without this, reading a revision still requires manual deserialization.

**Independent Test**: Can be tested by constructing a proto
`SandboxPolicyRevision` with a populated `policy` field, converting
to SDK type, and verifying all sub-fields are accessible as typed
structs.

**Acceptance Scenarios**:

1. **Given** a proto `SandboxPolicyRevision` with a populated policy,
   **When** converted via `SandboxPolicyRevisionFromProto`, **Then**
   the SDK `SandboxPolicyRevision.Policy` is a `*SandboxPolicy` with
   all sub-fields populated.
2. **Given** a proto `SandboxPolicyRevision` with no policy set,
   **When** converted, **Then** `SandboxPolicyRevision.Policy` is nil.

---

### User Story 4 - Read Policy from Current Config (Priority: P2)

An SDK consumer retrieves the current sandbox configuration via
`Config().Get()` and inspects the active policy as a typed
`*SandboxPolicy` instead of opaque `[]byte`. The consumer can
programmatically examine filesystem rules, process settings, and
network policies from the live sandbox config.

**Why this priority**: `SandboxConfig.Policy` is the fourth `[]byte`
policy field in the codebase. Without this, reading the current config
still requires manual deserialization even though create, update, and
revision history all use the typed struct.

**Independent Test**: Can be tested by constructing a proto
`GetSandboxConfigResponse` with a populated `policy` field, converting
to SDK type, and verifying all sub-fields are accessible as typed
structs.

**Acceptance Scenarios**:

1. **Given** a proto `GetSandboxConfigResponse` with a populated policy,
   **When** converted via `SandboxConfigFromProto`, **Then** the SDK
   `SandboxConfig.Policy` is a `*SandboxPolicy` with all sub-fields
   populated.
2. **Given** a proto `GetSandboxConfigResponse` with no policy set,
   **When** converted, **Then** `SandboxConfig.Policy` is nil.

---

### User Story 5 - Fake Client Preserves Policy (Priority: P2)

An SDK consumer writes tests using the fake client. When they create
a sandbox with an initial policy, the fake client stores the policy.
When they retrieve the sandbox, the policy is available on the
returned `SandboxSpec`.

**Why this priority**: Test support is essential for SDK consumers to
validate policy-related code without a live gateway.

**Independent Test**: Can be tested by creating a sandbox via the fake
client with a `SandboxPolicy`, then getting the sandbox and verifying
the policy round-trips correctly.

**Acceptance Scenarios**:

1. **Given** a fake client, **When** a sandbox is created with
   `SandboxSpec.Policy` set, **Then** getting the sandbox returns the
   same policy values.
2. **Given** a fake client, **When** `Config().Update()` is called
   with a typed policy, **Then** the update succeeds without error.

---

### Edge Cases

- What happens when `SandboxPolicy` has an empty `NetworkPolicies` map
  (non-nil but zero entries)? The converter must preserve the
  distinction between nil map and empty map.
- What happens when `FilesystemPolicy` has empty `ReadOnly` and
  `ReadWrite` slices? The converter must not convert empty slices to
  nil or vice versa.
- What happens when only some sub-policies are set (e.g., filesystem
  but not landlock)? Nil sub-structs must be preserved through
  conversion, not replaced with zero-value structs.
- What happens when `SandboxPolicy.Version` is 0? This is a valid
  proto value and must round-trip correctly.

## Requirements

### Functional Requirements

- **FR-001**: SDK MUST provide a `SandboxPolicy` domain type with
  fields: `Version` (uint32), `Filesystem` (*FilesystemPolicy),
  `Landlock` (*LandlockPolicy), `Process` (*ProcessPolicy),
  `NetworkPolicies` (map[string]NetworkPolicyRule).
- **FR-002**: SDK MUST provide a `FilesystemPolicy` domain type with
  fields: `IncludeWorkdir` (bool), `ReadOnly` ([]string),
  `ReadWrite` ([]string).
- **FR-003**: SDK MUST provide a `LandlockPolicy` domain type with
  field: `Compatibility` (string).
- **FR-004**: SDK MUST provide a `ProcessPolicy` domain type with
  fields: `RunAsUser` (string), `RunAsGroup` (string).
- **FR-005**: `SandboxSpec` MUST include a `Policy *SandboxPolicy`
  field.
- **FR-006**: `ConfigUpdate.Policy` MUST change from `[]byte` to
  `*SandboxPolicy`.
- **FR-007**: `SandboxPolicyRevision.Policy` MUST change from `[]byte`
  to `*SandboxPolicy`.
- **FR-007a**: `SandboxConfig.Policy` MUST change from `[]byte` to
  `*SandboxPolicy`.
- **FR-008**: SDK MUST provide `SandboxPolicyFromProto` and
  `SandboxPolicyToProto` converter functions with sub-type converters
  for `FilesystemPolicy`, `LandlockPolicy`, and `ProcessPolicy`.
- **FR-009**: Converters MUST reuse existing `NetworkPolicyRuleFromProto`
  and `NetworkPolicyRuleToProto` for the `network_policies` map.
- **FR-010**: Converters MUST deep-copy the `NetworkPolicies` map and
  all slice fields (`ReadOnly`, `ReadWrite`) at proto/SDK boundaries
  per Constitution VII.
- **FR-011**: The `SandboxSpec` converter MUST map the `policy` field
  (proto field 7), which is currently skipped.
- **FR-012**: The `ConfigUpdate` converter MUST use the typed
  `SandboxPolicy` converter instead of raw byte passthrough.
- **FR-012a**: The `SandboxConfig` converter MUST use the typed
  `SandboxPolicyFromProto` function instead of `proto.Marshal` to
  opaque bytes.
- **FR-013**: The fake client MUST store and return `SandboxPolicy` on
  sandbox creation and retrieval.
- **FR-014**: The fake client MUST accept `*SandboxPolicy` on
  `ConfigUpdate` operations without error.
- **FR-015**: All new converter functions MUST have round-trip tests
  verifying proto-to-SDK-to-proto fidelity.
- **FR-016**: Nil sub-policies MUST be preserved through conversion
  (nil in, nil out; not replaced with zero-value structs).
- **FR-017**: All new domain types MUST have Go doc comments describing
  their purpose and field semantics per Constitution IX.

### Key Entities

- **SandboxPolicy**: Top-level security policy configuration for a
  sandbox, containing filesystem, landlock, process, and network
  sub-policies plus a version number.
- **FilesystemPolicy**: Controls which directories the sandbox can
  access in read-only or read-write mode, and whether the working
  directory is auto-included.
- **LandlockPolicy**: Linux Landlock LSM configuration controlling
  compatibility mode for filesystem restriction enforcement.
- **ProcessPolicy**: Controls the user and group identity under which
  sandboxed processes execute.

## Success Criteria

### Measurable Outcomes

- **SC-001**: All four new domain types (`SandboxPolicy`,
  `FilesystemPolicy`, `LandlockPolicy`, `ProcessPolicy`) are exported
  from the `openshell/v1/types` package.
- **SC-002**: `SandboxSpec`, `ConfigUpdate`, `SandboxConfig`, and
  `SandboxPolicyRevision` all use `*SandboxPolicy` instead of `[]byte`.
- **SC-003**: Round-trip converter tests pass for all new types with
  100% field coverage (every field is set, converted, and verified).
- **SC-004**: Existing tests continue to pass (no regressions from the
  breaking changes).
- **SC-005**: `make ci` passes with zero lint violations and all tests
  green.
- **SC-006**: All new exported types and functions have Go doc
  comments.

## Assumptions

- The SDK is pre-1.0 and breaking changes to `ConfigUpdate.Policy` and
  `SandboxPolicyRevision.Policy` are acceptable without a deprecation
  period.
- `SandboxPolicy.Version` is included for proto fidelity. The server
  may manage versioning internally, but the SDK exposes the field so
  callers can set it if needed and read it back from revisions.
- `ResourceRequirements` (proto field 9 on `SandboxSpec`) is out of
  scope and tracked separately.
- The existing `NetworkPolicyRule` and `PolicyNetworkEndpoint` types
  and their converters are correct and complete; they are reused
  without modification.
- The `SandboxPolicyRevision` converter currently passes policy as raw
  bytes from the proto `SandboxPolicy` message; the new converter will
  use the typed `SandboxPolicyFromProto` function instead.
- The `SandboxConfig` converter currently serializes the proto
  `SandboxPolicy` to opaque bytes via `proto.Marshal`; the new
  converter will use the typed `SandboxPolicyFromProto` function
  instead.

## Clarification Log

- **Q: SandboxPolicy.Version settable by caller?** A: Yes, include
  the field. Document that the server may override the value. Callers
  can set it for creation, and it is populated on read-back from
  revisions.
- **Q: ResourceRequirements in same pass?** A: No, out of scope.
  Separate issue.
- **Q: Raw bytes escape hatch for GitOps?** A: Not now. YAGNI. Can
  be added later if a real need emerges.
