# Research: Operator API Extensions (Phase 2a)

## R1: Sub-Client Nesting Pattern

**Decision**: Provider profiles and refresh are nested sub-interfaces accessed via `client.Providers().Profiles()` and `client.Providers().Refresh()`. ServiceInterface is a new top-level sub-client accessed via `client.Services()`.

**Rationale**: Profiles and refresh are sub-resources of providers (they operate on provider-specific data). The nested accessor pattern (`Providers().Profiles()`) mirrors the client-go sub-resource convention and keeps `ClientInterface` from growing unbounded. Services are sandbox-scoped, not provider-scoped, so they belong at the top level.

**Implementation**: `ProviderInterface` gains two new methods: `Profiles() ProfileInterface` and `Refresh() RefreshInterface`. The `providerClient` struct holds references to the profile and refresh sub-clients, initialized during `NewClient`. `ClientInterface` gains `Services() ServiceInterface`.

## R2: ServiceEndpoint Proto Mapping

**Decision**: Map `ServiceEndpointResponse` (proto) to `ServiceEndpoint` (SDK type). The proto separates the endpoint object and URL into a response wrapper; the SDK flattens them into a single type with an optional URL field.

**Rationale**: Consumers don't care about the proto's response wrapping. A single `ServiceEndpoint` with all fields (including URL) is more ergonomic.

**Fields**: ID (from metadata), SandboxID, SandboxName, ServiceName, TargetPort (uint32), Domain (bool), URL (string, empty when domain=false).

## R3: ProviderProfile Proto Mapping

**Decision**: Map `ProviderProfile` (proto) to `ProviderProfile` (SDK type) with nested types for credentials, endpoints, binaries, and discovery.

**Rationale**: The proto ProviderProfile is already well-structured. The SDK type mirrors it closely but converts proto-specific types (repeated fields → slices, enums → string constants).

**Nested types needed**:
- `ProfileCredential`: name, description, required, secret (from ProviderProfileCredential)
- `ProfileCategory`: string enum (Other, Inference, Agent, SourceControl, Messaging, Data, Knowledge)
- `ProfileDiscovery`: credential names eligible for local discovery
- `ProfileImportItem`: profile + source string
- `ProfileDiagnostic`: source, profile ID, field, message, severity

**Note**: NetworkEndpoint and NetworkBinary come from the sandbox.proto package. These need SDK types or pass-through handling.

## R4: NetworkEndpoint and NetworkBinary

**Decision**: Create lightweight SDK types for `NetworkEndpoint` and `NetworkBinary` from `sandbox.proto` since they are part of the ProviderProfile structure.

**Rationale**: Proto Isolation requires we never expose proto types directly. These are small types (3-5 fields each) that appear in ProviderProfile. Creating SDK types follows the existing pattern.

**Alternative considered**: Omit these fields from the SDK ProviderProfile → loses information that operators need for profile management.

## R5: RefreshStrategy Enum

**Decision**: Map `ProviderCredentialRefreshStrategy` proto enum to a string-typed Go constant set: `RefreshStrategyStatic`, `RefreshStrategyExternal`, `RefreshStrategyOAuth2RefreshToken`, `RefreshStrategyOAuth2ClientCredentials`, `RefreshStrategyGoogleServiceAccountJWT`.

**Rationale**: String-typed constants are more ergonomic in Go than iota-based enums for domain values. They print readably in logs and test output. This matches how SandboxPhase and EventType are handled in Phase 1.

## R6: StopOnTerminal Implementation

**Decision**: StopOnTerminal is implemented at both the SDK and server level. The SDK passes `stop_on_terminal=true` to the server in `WatchSandboxRequest`. Additionally, the SDK's watch client monitors incoming `SandboxStreamEvent` messages and closes the watcher when a terminal phase is detected (SandboxReady or SandboxError).

**Rationale**: Dual implementation provides defense-in-depth. The server may close the stream, but the SDK also handles it in case the server doesn't (e.g., older gateway versions). The SDK-side check is a simple phase comparison after each event, adding negligible overhead.

**Implementation**: Add `StopOnTerminal bool` to `WatchOptions`. In the sandbox client's Watch goroutine, after delivering each status event, check if `StopOnTerminal` is set and the phase is Ready or Error. If so, close the watcher.

## R7: Fake Client Extension Pattern

**Decision**: Add stub sub-clients to the fake package following the same pattern as `fakeExecClient` and `fakeFileClient`: unexported structs returning `ErrorUnimplemented` for all methods. Add `fakeServiceClient`, `fakeProfileClient`, `fakeRefreshClient`.

**Rationale**: The fake must compile against the updated `ClientInterface`. Stubs are trivial (~20 lines each) and direct consumers to mock the interfaces directly if they need realistic behavior.

**Implementation**: `fakeProviderClient` gains `Profiles()` and `Refresh()` accessors returning the new stubs. `FakeClient` gains `Services()` accessor. All stub methods return `&StatusError{Code: ErrorUnimplemented, Message: "..."}`.
