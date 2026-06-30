# Brainstorm: Typed SandboxPolicy Domain Type

**Date:** 2026-06-30
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/11

## Problem Framing

The proto `SandboxSpec` has a `policy` field (field 7, type
`openshell.sandbox.v1.SandboxPolicy`) that supports setting an initial
network policy at sandbox creation time. The SDK's Go `SandboxSpec`
struct omits this field entirely, and the converter skips it. SDK
consumers cannot pass a policy when creating a sandbox.

Beyond creation, the same `SandboxPolicy` proto type appears in
`UpdateConfigRequest.policy` for full policy replacement at runtime.
The SDK currently represents this as opaque `[]byte` in
`ConfigUpdate.Policy` and `SandboxPolicyRevision.Policy`. This forces
consumers to manually marshal/unmarshal policy YAML when working
programmatically.

The proto `SandboxPolicy` message contains five fields: `version`,
`FilesystemPolicy` (3 fields), `LandlockPolicy` (1 field),
`ProcessPolicy` (2 fields), and `map<string, NetworkPolicyRule>`
(already fully typed in the SDK with 18-field endpoints).

## Approaches Considered

### A: Typed SandboxPolicy with Full Replacement (chosen)

Introduce a typed `SandboxPolicy` struct with `FilesystemPolicy`,
`LandlockPolicy`, `ProcessPolicy` sub-types. Replace `[]byte` with
`*SandboxPolicy` in all three locations: `SandboxSpec.Policy` (new),
`ConfigUpdate.Policy` (breaking), `SandboxPolicyRevision.Policy`
(breaking). Add converters for all new types, reuse existing
`NetworkPolicyRule` converters.

- Pros: Consistent typed API everywhere. Good programmatic UX. Mirrors
  proto structure faithfully. New sub-types are small (6 fields total
  across 3 structs). One domain type serves creation, update, and read.
- Cons: Breaking change on `ConfigUpdate` and `SandboxPolicyRevision`.
  Larger scope than just "add a field."

### B: SandboxSpec Field Only, Leave Existing []byte

Add the typed `SandboxPolicy` to `SandboxSpec` only. Leave
`ConfigUpdate.Policy` and `SandboxPolicyRevision.Policy` as `[]byte`.
Migrate those in a follow-up.

- Pros: Smaller scope. No breaking changes on existing types.
- Cons: Inconsistency between creation (typed) and update/read (bytes).
  Users must marshal/unmarshal when moving between contexts.

### C: Typed + Raw Bytes Dual Path

Add typed `*SandboxPolicy` everywhere AND keep a `PolicyRaw []byte`
escape hatch for GitOps YAML import scenarios.

- Pros: Maximum flexibility for different consumption patterns.
- Cons: Two ways to do the same thing. Precedence rules needed when
  both are set. More testing surface. YAGNI until a real need emerges.

## Decision

Approach A: Full typed replacement. The SDK is pre-1.0, so breaking
changes on `ConfigUpdate` and `SandboxPolicyRevision` are acceptable.
Typed structs provide better programmatic UX, and consistency across
creation/update/read is worth the slightly larger scope. A raw bytes
escape hatch can be added later if GitOps scenarios demand it.

## Key Requirements

- New domain types: `SandboxPolicy`, `FilesystemPolicy`,
  `LandlockPolicy`, `ProcessPolicy`
- New field: `SandboxSpec.Policy *SandboxPolicy`
- Breaking change: `ConfigUpdate.Policy` from `[]byte` to
  `*SandboxPolicy`
- Breaking change: `SandboxPolicyRevision.Policy` from `[]byte` to
  `*SandboxPolicy`
- Converters: `SandboxPolicyFromProto`/`SandboxPolicyToProto` plus
  sub-type converters; reuse existing `NetworkPolicyRule` converters
- Update `SandboxSpec` converter to map the policy field (currently
  skipped)
- Update `ConfigUpdate` converter to use typed policy instead of raw
  bytes
- Fake client: support policy field in sandbox create and config update
- Round-trip converter tests for all new types
- Deep-copy at boundaries (Constitution VII) for the network_policies
  map and filesystem path slices

## Open Questions

- Should `SandboxPolicy.Version` be settable by the caller, or is it
  server-assigned? The proto has it as a field, but the server may
  manage versioning internally.
- `ResourceRequirements` (proto field 9 on SandboxSpec) is also missing
  from the SDK. Should it be added in the same pass or kept separate?
