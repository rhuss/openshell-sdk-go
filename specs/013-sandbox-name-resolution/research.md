# Research: Sandbox Name Resolution

## R1: Existing Name-to-ID Resolution Pattern

**Decision**: Replicate the SSH.Tunnel() pattern (SandboxInterface
injection + Get() call) across all ID-based sub-clients.

**Rationale**: The pattern is already proven, tested, and understood in
the codebase. SSHClient stores a `SandboxInterface` reference and calls
`sandboxes.Get(ctx, name)` to resolve before each proto request. This
adds one extra gRPC round-trip per call but keeps the code simple and
self-contained.

**Alternatives considered**:
- SandboxSession wrapper (Python SDK pattern): rejected because it
  introduces a new API concept and breaks the sub-client accessor
  pattern.
- Name-or-ID auto-detection: rejected because the heuristic is fragile
  and the API becomes ambiguous.
- Centralized resolver service: rejected as over-engineering for a
  simple `Get()` call.

## R2: Watch Bug Analysis

**Decision**: Fix Watch by adding a `Get()` call before constructing
`WatchSandboxRequest`, same as `GetLogs` does.

**Rationale**: `WatchSandboxRequest.Id` (proto field 1) is documented as
"Sandbox id" in the proto. The current SDK passes the user-provided
`name` parameter directly into this field. The server may or may not
accept names here (undetermined without integration test), but passing
the wrong type of identifier is incorrect regardless.

**Implementation**: `sandbox_client.go` line 176. Add `Get(ctx, name)`
call before the stream creation, use `sb.ID` in the request.

## R3: Fake Client Impact

**Decision**: Rename parameters only, no resolution logic in fakes.

**Rationale**: Fake clients use in-memory maps keyed by sandbox name.
Their operations (like fake exec) don't call real proto RPCs, so there
is no ID field to populate. The parameter rename ensures interface
compliance. Per Constitution XI (Fake-Real Parity), fakes should mirror
client-side validation but not resolution logic (which is a server
interaction concern).

## R4: SSH.CreateSession Exception

**Decision**: Keep `CreateSession` as ID-based.

**Rationale**: `CreateSession` is a low-level RPC wrapped by `Tunnel()`.
`Tunnel()` already resolves the name and passes the ID to
`CreateSession`. If `CreateSession` also resolved, it would double-
resolve. Consumers should use `Tunnel()` for name-based access. A doc
comment will make this explicit.

## R5: Breaking Change Assessment

**Decision**: Proceed with the break. Pre-1.0 SDK, no external consumers
known beyond cc-deck (which is also owned by us).

**Rationale**: The parameter rename from `sandboxID` to `sandboxName`
changes the public interface signatures. This is a compile-time break:
callers that pass `sandbox.ID` must change to `sandbox.Name`. The
migration is mechanical and searchable.
