# Data Model: Inference Route Client

## Entities

### InferenceRouteConfig (input to SetRoute)

Configuration for setting an inference route on a workspace.

| Field | Type | Description |
|-------|------|-------------|
| ProviderName | string | Provider record name for credentials and endpoint mapping. Required. |
| ModelID | string | Model identifier to force on generation calls. Required. |
| RouteName | string | Route name to target. Empty string = default user-facing route. |
| NoVerify | bool | When true, skip synchronous endpoint validation before persistence. |
| TimeoutSecs | uint64 | Per-route request timeout in seconds. 0 = default (60s). |

**Validation rules**:
- ProviderName must not be empty
- ModelID must not be empty
- Workspace is passed as a separate method parameter, not part of this type

### InferenceRoute (response from SetRoute and GetRoute)

Represents a configured inference route as returned by the gateway.

| Field | Type | Description |
|-------|------|-------------|
| ProviderName | string | Provider record name. |
| ModelID | string | Model identifier. |
| Version | uint64 | Server-assigned version for the route. |
| RouteName | string | Route name that was configured or queried. |
| TimeoutSecs | uint64 | Per-route request timeout in seconds. |
| Workspace | string | Workspace the route belongs to. |

### SetInferenceRouteResult (extended response for Set only)

Extends InferenceRoute with validation metadata returned only by the Set operation.

| Field | Type | Description |
|-------|------|-------------|
| InferenceRoute | (embedded) | All fields from InferenceRoute. |
| ValidationPerformed | bool | Whether endpoint verification ran during this request. |
| ValidatedEndpoints | []ValidatedEndpoint | Endpoints probed during validation, if any. |

### ValidatedEndpoint

An endpoint that was probed during route validation.

| Field | Type | Description |
|-------|------|-------------|
| URL | string | Endpoint URL that was validated. |
| Protocol | string | Protocol used (e.g., "openai", "vertex"). |

## Relationships

- InferenceRouteConfig is the **input** for SetRoute; InferenceRoute is the **output**.
- Routes are scoped to a workspace (workspace passed as method parameter).
- Routes are uniquely identified by (workspace, route_name).
- Version field enables optimistic concurrency checks by the caller if needed.

## Identity & Uniqueness

- A route is uniquely identified by the combination of workspace and route_name.
- The empty string route_name is a valid, distinct route (the default user-facing route).
- The fake client uses a composite key `workspace + "/" + routeName` for storage.
