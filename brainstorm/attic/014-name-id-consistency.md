# Brainstorm: Sandbox Name-vs-ID Consistency

**Date:** 2026-06-30
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/15

## Problem Framing

The OpenShell gateway proto API uses sandbox **name** for most RPCs
(Create, Get, Delete, all Policy/Draft/Service/Config RPCs) but sandbox
**ID** for a cluster of low-level RPCs: ExecSandbox, CreateSshSession,
ForwardTcp, WatchSandbox, GetSandboxLogs, and GetSandboxConfig.

The Go SDK surfaces this inconsistency directly. 8 public methods
require the caller to already have the opaque sandbox ID, while 20+
methods accept the human-readable sandbox name. Two methods (GetLogs
and SSH.Tunnel) already resolve name to ID internally via an extra
`Get()` call, but the remaining 6 push this burden onto the caller.

One method, `Watch`, passes a name directly into the proto
`WatchSandboxRequest.Id` field. This is either a bug or only works
because the server happens to accept both (the proto comment says
"Sandbox id").

Additionally, the proto itself uses three different field names for
name-based identifiers (`name`, `sandbox_name`, `sandbox`) and two
for ID-based identifiers (`sandbox_id`, `id`).

**Origin:** idea-inbox entry `log-id-vs-name-mismatch` from brainstorm
#009 (Phase 2b-2 revisit), expanded to cover all RPCs after audit.

### Audit Results

| SDK Method | Current Param | Proto Field | Resolution? |
|---|---|---|---|
| Exec.Run | sandboxID | sandbox_id | none (caller) |
| Exec.Stream | sandboxID | sandbox_id | none (caller) |
| Exec.Interactive | sandboxID | sandbox_id | none (caller) |
| TCP.Forward | sandboxID | sandbox_id | none (caller) |
| File.Upload | sandboxID | sandbox_id | none (caller) |
| File.Download | sandboxID | sandbox_id | none (caller) |
| SSH.CreateSession | sandboxID | sandbox_id | none (caller) |
| Config.GetSandbox | sandboxID | sandbox_id | none (caller) |
| SSH.Tunnel | sandboxName | sandbox_id | **yes** (injected SandboxInterface) |
| Sandbox.GetLogs | sandboxName | sandbox_id | **yes** (inline Get call) |
| Sandbox.Watch | name | id | **none (passes name to ID field)** |

The Python SDK sidesteps this with a `SandboxSession` wrapper that
holds the ID after an initial `get()`, so session methods always have
the ID available. The Go SDK has no equivalent session concept.

## Approaches Considered

### A: Inject SandboxInterface into ID-based sub-clients (chosen)

Follow the pattern already established by SSHClient, which receives
`SandboxInterface` at construction to resolve names for `Tunnel()`.
Extend this to ExecClient, TCPClient, FileClient, and ConfigClient.

All public methods switch from `sandboxID string` to `sandboxName
string` and resolve internally via `sandboxes.Get(ctx, name)`.

- Pros: Uniform API surface (all methods accept names), follows
  established SDK pattern, no new abstractions, each sub-client
  is self-contained
- Cons: Extra gRPC round-trip per call for name resolution, breaking
  change to public interface signatures

### B: Add a SandboxSession type (Python SDK pattern)

Create a `SandboxSession` returned by `Sandboxes().Get()` that holds
the resolved ID. Methods hang off the session: `session.Exec(...)`,
`session.Upload(...)`, etc.

- Pros: Resolution happens once, subsequent calls are free, mirrors
  Python SDK pattern
- Cons: Large API surface change, new type to maintain, breaks the
  sub-client accessor pattern (`client.Exec()` vs `session.Exec()`),
  existing consumers would need significant rewrite

### C: Accept both name and ID everywhere

Change method signatures to accept a generic "sandbox reference" that
could be either a name or an ID. Let the SDK detect which it is (IDs
have a known format) and skip resolution for IDs.

- Pros: Maximum flexibility for callers
- Cons: Ambiguous API, detection heuristic could be fragile, hides
  whether resolution happened, harder to test

## Decision

**Approach A: Inject SandboxInterface into ID-based sub-clients.**

The extra gRPC round-trip per call is acceptable. The gateway's
`GetSandbox` RPC is lightweight, and for the typical flow (create a
sandbox, then exec/upload/forward), the caller already has the sandbox
object from the create response. The name-to-ID resolution is a safety
net, not a hot path.

This is a breaking change (pre-1.0 SDK, acceptable). Callers replace
`sandbox.ID` with `sandbox.Name` in their calls.

**Exception:** `SSH.CreateSession` stays ID-based. It's a low-level RPC
that `Tunnel()` wraps with name resolution. Consumers should use
`Tunnel()`, not `CreateSession()` directly. A doc comment will note this.

## Key Requirements

1. All public SDK methods (except SSH.CreateSession) accept sandbox
   **name**, not ID
2. Name-to-ID resolution uses the existing `SandboxInterface.Get()`
   pattern from SSHClient
3. `Watch` must resolve name to ID before constructing
   `WatchSandboxRequest` (fix the bug)
4. Fake client implementations update method signatures (no resolution
   needed in fakes, they use names internally)
5. File a tracking issue on `rhuss/openshell-sdk-go` (not upstream yet)
   covering the proto field naming inconsistency and the potential for
   server-side name resolution

## Open Questions

- Does WatchSandboxRequest.id actually reject names server-side, or does
  the gateway resolve both? Answering this requires a running gateway
  (integration test). The SDK fix resolves regardless.
- Should the SDK cache the name-to-ID mapping to avoid repeated Get calls
  when the same sandbox name is used across multiple method calls? For
  now, no caching. Callers who care about performance can resolve once
  and pass the name (which the SDK will re-resolve, but that's a small
  overhead).
- Upstream proto cleanup: file on NVIDIA/OpenShell later once the SDK
  changes are merged and we can reference them.
