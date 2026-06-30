# Research: Typed SandboxPolicy Domain Type

## Existing Patterns

### Policy Bytes Pattern (current)

The codebase uses a consistent pattern for opaque policy handling:
- **Read path**: `proto.Marshal(sandboxPolicy)` to produce `[]byte`,
  then `CopyByteSlice` for deep-copy isolation.
  Used in: `SandboxPolicyRevisionFromProto`, `SandboxConfigFromProto`.
- **Write path**: `proto.Unmarshal(bytes, &sbv1.SandboxPolicy{})` to
  reconstruct proto, then set on request.
  Used in: `ConfigUpdateToProto`.

Decision: Replace both patterns with typed converter calls.
Rationale: Typed converters eliminate marshal/unmarshal overhead and
provide compile-time safety.

### Converter Organization

Each domain area has its own converter file:
- `network_policy.go` for NetworkPolicyRule (with 18-field endpoint)
- `policy.go` for PolicyChunk, DraftPolicy, SandboxPolicyRevision
- `setting.go` for ConfigUpdate, SandboxConfig, SettingValue
- `sandbox.go` for Sandbox, SandboxSpec, SandboxTemplate

Decision: Add `SandboxPolicyFromProto`/`SandboxPolicyToProto` and
sub-type converters to `policy.go` (policy domain). Update
`sandbox.go` to call it for SandboxSpec. Update `setting.go` to call
it for ConfigUpdate and SandboxConfig.
Rationale: Keeps the SandboxPolicy converter with other policy
converters. Cross-file calls follow existing pattern (e.g.,
`sandbox.go` already calls helpers from other files).

### Deep-Copy Helpers

`copy.go` provides: `CopyStringMap`, `CopyStringSlice`,
`CopyByteSlice`, `CopyBoolPtr`. The fake client has local
`copyStringMap`/`copyStringSlice` helpers.

Decision: Use `CopyStringSlice` for `ReadOnly`/`ReadWrite` slices.
For the `NetworkPolicies` map, write a dedicated deep-copy that
iterates entries and copies each `NetworkPolicyRule` value.
Rationale: Constitution VII requires deep-copy at boundaries. Maps
of structs need per-entry copying.

### Fake Client Pattern

The fake `copySandboxSpec` helper copies each field individually. The
fake's `Create` stores the copied spec on the in-memory sandbox object.

Decision: Extend `copySandboxSpec` to deep-copy the `Policy` field.
Add a `copySandboxPolicy` helper in the fake package.
Rationale: Consistent with existing fake deep-copy pattern.

### Test Patterns

Converter tests use:
1. Nil input test (returns nil/zero)
2. Round-trip test (SDK -> proto -> SDK, assert equality)
3. Deep-copy isolation test (convert, mutate source, verify copy)

Decision: Follow same 3-test pattern for all new converters.
Rationale: Consistency with existing test suite.

## Proto Field Mapping

| Proto Field | SDK Field | Notes |
|---|---|---|
| `SandboxPolicy.version` | `SandboxPolicy.Version` | uint32, pass-through |
| `SandboxPolicy.filesystem` | `SandboxPolicy.Filesystem` | *FilesystemPolicy |
| `SandboxPolicy.landlock` | `SandboxPolicy.Landlock` | *LandlockPolicy |
| `SandboxPolicy.process` | `SandboxPolicy.Process` | *ProcessPolicy |
| `SandboxPolicy.network_policies` | `SandboxPolicy.NetworkPolicies` | map[string]NetworkPolicyRule |
| `FilesystemPolicy.include_workdir` | `FilesystemPolicy.IncludeWorkdir` | bool |
| `FilesystemPolicy.read_only` | `FilesystemPolicy.ReadOnly` | []string |
| `FilesystemPolicy.read_write` | `FilesystemPolicy.ReadWrite` | []string |
| `LandlockPolicy.compatibility` | `LandlockPolicy.Compatibility` | string |
| `ProcessPolicy.run_as_user` | `ProcessPolicy.RunAsUser` | string |
| `ProcessPolicy.run_as_group` | `ProcessPolicy.RunAsGroup` | string |

## Affected Files

| File | Change Type |
|---|---|
| `openshell/v1/types/policy.go` | Add SandboxPolicy, FilesystemPolicy, LandlockPolicy, ProcessPolicy |
| `openshell/v1/types/sandbox.go` | Add Policy field to SandboxSpec |
| `openshell/v1/types/setting.go` | Change ConfigUpdate.Policy, SandboxConfig.Policy from []byte to *SandboxPolicy |
| `openshell/v1/internal/converter/policy.go` | Add SandboxPolicy/sub-type converters, update SandboxPolicyRevisionFromProto |
| `openshell/v1/internal/converter/sandbox.go` | Update sandboxSpecFromProto/SandboxSpecToProto for policy field |
| `openshell/v1/internal/converter/setting.go` | Update ConfigUpdateToProto, SandboxConfigFromProto for typed policy |
| `openshell/v1/internal/converter/policy_test.go` | Add SandboxPolicy converter tests, update revision test |
| `openshell/v1/internal/converter/sandbox_test.go` | Update SandboxSpec round-trip test for policy |
| `openshell/v1/internal/converter/setting_test.go` | Update ConfigUpdate and SandboxConfig tests |
| `openshell/v1/fake/sandbox.go` | Update copySandboxSpec for policy |
| `openshell/v1/fake/sandbox_test.go` | Add test for policy round-trip through fake |
| `openshell/v1/fake/config.go` | Verify config update accepts typed policy |
