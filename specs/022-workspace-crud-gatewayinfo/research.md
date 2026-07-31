# Research: Workspace CRUD, GatewayInfo & GetCurrentUser

## R1: Proto Availability

**Decision**: All 9 RPCs and their request/response types exist in the generated proto bindings.

**Findings**:
- Workspace CRUD: `CreateWorkspace`, `GetWorkspace`, `ListWorkspaces`, `DeleteWorkspace` in `openshellv1`
- Member management: `AddWorkspaceMember`, `RemoveWorkspaceMember`, `ListWorkspaceMembers` in `openshellv1`
- Gateway: `GetGatewayInfo`, `GetCurrentUser` already in `openshellv1`
- Workspace data model: `datamodelv1.Workspace` with `ObjectMeta` + `WorkspaceStatus`
- `WorkspaceMember` has its own `ObjectMeta` + `PrincipalSubject` + `WorkspaceRole`

**No proto sync needed.** All types are current.

## R2: WorkspaceRole Representation

**Decision**: Use `type WorkspaceRole string` with string constants (matching `SandboxPhase` pattern).

**Rationale**: The existing SDK uses `type SandboxPhase string` with const values like `SandboxReady`, `SandboxError`. This is idiomatic Go for the SDK. The proto uses `int32` enum, but the converter layer maps between them.

**Alternatives considered**:
- `type WorkspaceRole int`: More type-safe but inconsistent with `SandboxPhase` pattern. Would require custom String() method.

## R3: Sub-client Pattern

**Decision**: Follow the existing sub-client pattern exactly.

**Findings**:
- Interface defined in `openshell/v1/<name>.go` (e.g., `sandbox.go`)
- Implementation in `openshell/v1/<name>_client.go` (e.g., `sandbox_client.go`)
- Types in `openshell/v1/types/<name>.go`
- Type aliases re-exported in `openshell/v1/<name>.go` or `types_reexport.go`
- Converter in `openshell/v1/internal/converter/<name>.go`
- Fake in `openshell/v1/fake/<name>.go`
- Client struct field + accessor in `openshell/v1/client.go`

## R4: Health Sub-client Extension

**Decision**: Add `GetGatewayInfo` and `GetCurrentUser` to the existing `HealthInterface`.

**Findings**:
- Current `HealthInterface` has only `Check(ctx) (*HealthResult, error)`
- Adding 2 more methods is a breaking change for anyone implementing the interface, but this is acceptable pre-1.0
- The `healthClient` struct already wraps `pb.OpenShellClient` which has both RPCs available

## R5: List Options Pattern

**Decision**: Workspace `List` and `ListMembers` use the existing `ListOptions` struct from `types/options.go`.

**Findings**:
- `ListOptions` already has `Limit`, `Offset`, `LabelSelector`, and `AllWorkspaces` fields
- `ListWorkspaces` uses `Limit`, `Offset`, `LabelSelector` (same as other list RPCs)
- `ListWorkspaceMembers` uses `Limit`, `Offset` only (no label filtering on members)
- The variadic `opts ...ListOptions` pattern is used by `SandboxInterface.List`

## R6: Fake Client Pattern

**Decision**: Follow `fake/sandbox.go` pattern with `objectStore` for workspace + member stores.

**Findings**:
- Fake uses `objectStore[T]` generic store with name-based key function
- `ClientOption` functions configure fake behavior (e.g., `WithHealthResult`)
- New options needed: `WithGatewayInfo(*types.GatewayInfo)`, `WithCurrentUser(*types.CurrentUser)`
- Workspace fake needs its own `objectStore[*types.Workspace]`
- Member fake needs a store keyed by `workspace:principalSubject`
