# Research: Fake Client Package

## R1: ErrorUnimplemented Error Code

**Decision**: Add `ErrorUnimplemented` to the `ErrorCode` enum in `openshell/v1/types/errors.go` with a corresponding `IsUnimplemented` helper function.

**Rationale**: FR-011 requires Exec and File methods to return an Unimplemented StatusError. The existing `ErrorCode` enum has 8 codes (NotFound through Internal) but no Unimplemented. gRPC has a standard `UNIMPLEMENTED` status code, so this is a natural addition.

**Alternatives considered**:
- Reuse `ErrorInternal` with an "unimplemented" message → Lossy, consumers cannot distinguish unimplemented from internal errors.
- Return a plain `errors.New` → Breaks the StatusError contract that all SDK operations use.

## R2: SandboxPhase for "Pending" State

**Decision**: Use `SandboxProvisioning` (the existing constant) as the initial phase when a sandbox is created via the fake. The spec said "Pending" but the codebase uses `Provisioning`.

**Rationale**: The `types.go` file defines `SandboxProvisioning`, `SandboxReady`, `SandboxError`, `SandboxDeleting`, and `SandboxUnknown`. There is no "Pending" phase. Provisioning is the correct initial state.

**Alternatives considered**: None — the enum is authoritative.

## R3: Generic Object Store vs Per-Type Stores

**Decision**: Use Go generics (`ObjectStore[T any]`) with a name-extraction function for the generic store, plus type-specific thin wrappers for sandbox and provider sub-clients.

**Rationale**: Both sandbox and provider stores share identical CRUD semantics (keyed by name, AlreadyExists/NotFound errors, List-all, idempotent delete). A generic store avoids duplicating this logic. The name-extraction function (`func(T) string`) lets the store get the name from any type.

**Alternatives considered**:
- Two separate map implementations → Code duplication with identical logic.
- Interface-based store (`Nameable` interface) → Requires methods on pointer receivers of types we don't control structurally.

## R4: Watch Broadcaster Design

**Decision**: Use a `sync.RWMutex`-protected slice of watcher channels. Each watcher gets a buffered channel. Broadcast iterates the slice, non-blocking send to each channel (drop events if buffer full, matching k8s watch semantics). Stop removes the watcher from the slice and closes its channel.

**Rationale**: This matches the client-go `watch.Broadcaster` pattern. Buffered channels prevent slow consumers from blocking mutations. The buffer size (default 100) matches k8s conventions.

**Alternatives considered**:
- Unbuffered channels → Mutations block until all watchers consume, risking deadlocks in tests.
- Fan-out goroutines per watcher → Over-engineered for test use cases.

## R5: Deep Copy at Store Boundaries

**Decision**: Deep-copy Sandbox and Provider objects on store insert and retrieval. This prevents mutations to returned objects from corrupting store state, per Constitution Principle VII (Deep Copy at Boundaries).

**Rationale**: Without deep copy, a consumer could do `sb, _ := fake.Sandboxes().Get(ctx, "x"); sb.Name = "y"` and corrupt the store. The real gRPC client naturally provides isolation because each response is a fresh proto-to-Go conversion.

**Alternatives considered**:
- Shallow copy → Fails for map and slice fields (Labels, Environment, Conditions).
- Store proto objects instead of domain types → Violates Proto Isolation principle.

## R6: Provider-Sandbox Association Storage

**Decision**: Store provider-sandbox associations as a `map[string][]string` (sandbox name → provider names) in the sandbox store. AttachProvider appends, DetachProvider removes, ListProviders returns the slice.

**Rationale**: The SandboxInterface includes AttachProvider, DetachProvider, and ListProviders. These need state tracking. A simple map is sufficient since we don't need rich provider data for association queries — the provider objects themselves live in the provider store.

**Alternatives considered**:
- Track associations on the Sandbox.Spec.Providers field → Couples mutation with object store, complicates deep copy logic.
