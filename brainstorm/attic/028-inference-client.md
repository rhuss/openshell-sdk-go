# Brainstorm: Inference Route Client

**Date:** 2026-07-31
**Status:** active

## Problem Framing

The upstream OpenShell gateway has an `openshell.inference.v1.Inference` service with 4 RPCs. The SDK has no inference client, and doesn't even have `inference.proto` in its proto directory. The dashboard needs inference route CRUD (set/get/delete) for workspace-scoped inference management.

### Upstream RPCs

| RPC | Auth | Role | What it does |
|-----|------|------|-------------|
| `SetInferenceRoute` | bearer | workspace admin | Configure how `inference.local` routes for a workspace |
| `GetInferenceRoute` | bearer | workspace user | Fetch the configured route for a workspace |
| `DeleteInferenceRoute` | bearer | workspace admin | Remove a route from a workspace |
| `GetInferenceBundle` | sandbox | n/a | Return resolved route bundle for sandbox-local execution (sandbox-only, not user-facing) |

Only the first 3 are user-facing. `GetInferenceBundle` is sandbox-internal and should be skipped.

## Approaches Considered

### A: New top-level Inference() sub-client

Add `client.Inference()` returning an `InferenceInterface`:

```go
type InferenceInterface interface {
    SetRoute(ctx context.Context, workspace string, config *InferenceRouteConfig) (*InferenceRoute, error)
    GetRoute(ctx context.Context, workspace, routeName string) (*InferenceRoute, error)
    DeleteRoute(ctx context.Context, workspace, routeName string) error
}
```

- Pros: follows the sub-client pattern, inference is a distinct service in the proto
- Cons: small interface (3 methods), may feel over-engineered

### B: Methods on Config() sub-client

Add inference route methods to the existing `Config()` sub-client since inference routing is a form of workspace configuration.

- Pros: fewer interfaces, inference is conceptually a config setting
- Cons: inference has its own proto service, mixing concepts

## Decision

**Approach A: New `Inference()` sub-client.** Inference is a separate gRPC service (`openshell.inference.v1.Inference`) with its own proto file. Keeping it as a separate sub-client matches the proto structure and is consistent with how the dashboard's gateway layer already separates it (separate `inferencev1.InferenceClient`).

## Key Requirements

1. Add `inference.proto` to the SDK's proto directory (part of proto sync, brainstorm #025)
2. Wire inference stubs into `buf.gen.yaml`
3. Create `InferenceInterface` with 3 methods (skip `GetInferenceBundle`)
4. New types: `InferenceRouteConfig` (provider_name, model_id, route_name, timeout, no_verify), `InferenceRoute` (the response)
5. Converters: `InferenceRouteConfigToProto`, `InferenceRouteFromProto`
6. Fake client implementation
7. Add `Inference()` accessor to `ClientInterface` and `Client`

## Open Questions

- Route naming convention: the dashboard uses `routeName` "" for user-facing and "sandbox-system" for system route. Should the SDK expose this as a string, or define constants?
- The `InferenceRouteConfig` message has `provider_name`, `model_id`, `route_name`, `no_verify`, `timeout_secs`, `workspace`. Should the SDK type mirror this 1:1 or simplify (e.g., embed workspace in the method signature)?
