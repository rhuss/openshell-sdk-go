# Research: Workspace Scoping

## Proto Workspace Support

**Decision**: All 28 proto Request messages already include a `string workspace` field. No proto changes needed.

**Rationale**: The upstream OpenShell proto was designed with workspace scoping from the start. The SDK simply hasn't been populating these fields.

**Alternatives considered**: None needed; proto support is complete.

## AllWorkspaces Proto Support

**Decision**: The `all_workspaces` bool field exists on ListSandboxesRequest (field 5), ListProvidersRequest (field 5), and ListServicesRequest (field 4). The SDK's `ListOptions` struct needs an `AllWorkspaces bool` field mapped to these proto fields.

**Rationale**: The proto already models cross-workspace listing. The SDK just needs to expose it.

## Fake Store Workspace Isolation

**Decision**: Use composite key (workspace + name) in the existing `objectStore[T]` generic store.

**Rationale**: The current `objectStore` uses a flat `map[string]T` keyed by name. Workspace isolation requires either (a) nested maps `map[workspace]map[name]T` or (b) composite keys `workspace/name`. Composite keys are simpler and avoid changing the generic store's type signature. The `nameFunc` becomes a `keyFunc` that returns `workspace + "/" + name`. For AllWorkspaces listing, the store iterates all entries and strips the workspace prefix.

**Alternatives considered**:
- Nested maps (`map[string]map[string]T`): More type-safe but requires rewriting the generic store interface. Overkill for string-keyed lookups.
- Separate store instance per workspace: Clean isolation but wastes memory for empty workspaces and complicates AllWorkspaces listing.

## Parameter Position Convention

**Decision**: Workspace is the first parameter after `ctx context.Context` in all methods.

**Rationale**: Workspace is a mandatory scoping parameter, not an optional modifier. Placing it first (after ctx) follows Go convention for required parameters. It mirrors how Kubernetes client-go places namespace.

## Nested Sub-client Workspace Flow

**Decision**: Each method on ProfileInterface and RefreshInterface accepts workspace as an explicit parameter (per FR-001). The factory methods `Profiles()` and `Refresh()` on ProviderInterface remain parameterless.

**Rationale**: Approach A (explicit parameter) was chosen in the brainstorm. Making workspace implicit via factory methods would create hidden state, contradicting the design decision. Each call site explicitly passes workspace.

## Unscoped Operations

**Decision**: `Health().Check()` and `Config().GetGateway()` do NOT get workspace parameters.

**Rationale**: These RPCs are gateway-scoped, not workspace-scoped. Their proto request messages do not have workspace fields. Adding workspace would be misleading.
