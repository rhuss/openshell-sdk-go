# Research: Inference Route Client

## Proto Service Structure

- **Decision**: Use `inferencev1.NewInferenceClient(conn)` for the gRPC stub, not `openshellv1.OpenShellClient`.
- **Rationale**: The inference service is defined as a separate gRPC service (`openshell.inference.v1.Inference`) in `proto/inference.proto`, with generated code in `proto/inferencev1/`. Unlike workspaces (which use the main `OpenShell` service), inference has its own service definition.
- **Alternatives considered**: Using the OpenShell service was not viable since inference RPCs are not part of it.

## Proto & Code Generation Status

- **Decision**: No proto sync or buf generation work needed.
- **Rationale**: `inference.proto` already exists in `proto/` and is already listed in `buf.gen.yaml` inputs. Generated stubs exist at `proto/inferencev1/inference.pb.go` and `proto/inferencev1/inference_grpc.pb.go`.
- **Alternatives considered**: None; this was already done in a prior PR.

## SDK Type Design

- **Decision**: Workspace is a method parameter, not part of InferenceRouteConfig.
- **Rationale**: The existing SDK pattern (workspace_client.go) passes workspace as a method parameter. The brainstorm explicitly decided this. Keeping workspace out of the config struct avoids redundancy and matches the rest of the SDK.
- **Alternatives considered**: Embedding workspace in InferenceRouteConfig (rejected for inconsistency with SDK patterns).

## SetInferenceRoute Request Fields

- **Decision**: The SDK config type exposes `ProviderName`, `ModelID`, `RouteName`, `NoVerify`, and `TimeoutSecs`. The proto `verify` field (field 4) is NOT exposed; `no_verify` (field 5) is sufficient.
- **Rationale**: The proto has both `verify` and `no_verify` booleans. The SDK simplifies to `NoVerify` only, since the gateway treats `no_verify=true` as skipping verification and `verify=true` as forcing it, but the default (both false) already verifies. Exposing only `NoVerify` is simpler.
- **Alternatives considered**: Exposing both fields; rejected to avoid confusing combinations.

## SetInferenceRoute Response Fields

- **Decision**: The SDK response type mirrors the response proto: `ProviderName`, `ModelID`, `Version`, `RouteName`, `ValidationPerformed`, `ValidatedEndpoints`, `TimeoutSecs`, `Workspace`.
- **Rationale**: The response contains useful metadata (version for optimistic concurrency, validation results) that callers may need.
- **Alternatives considered**: Stripping validation fields; rejected because dashboards need to display validation status.

## Fake Client Store Design

- **Decision**: Use a simple map keyed by `workspace + "/" + routeName` for in-memory storage.
- **Rationale**: Routes are uniquely identified by workspace + route name. The existing fake uses `objectStore[T]` but that keys by a single name field. A composite key is simpler than a nested map.
- **Alternatives considered**: Nested `map[workspace]map[routeName]` (more complex, not needed for the fake).

## Error Handling

- **Decision**: Follow existing `StatusError` pattern with `ErrorInvalidArgument` for validation, `converter.FromGRPCError` for RPC errors.
- **Rationale**: Consistent with every other sub-client (workspace_client.go, config_client.go, etc.).
- **Alternatives considered**: None; this is the established pattern.
