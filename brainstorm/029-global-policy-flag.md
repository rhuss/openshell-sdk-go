# Brainstorm: Global Policy Flag for List and GetStatus

**Date:** 2026-08-01
**Status:** active
**GitHub:** https://github.com/rhuss/openshell-sdk-go/issues/44

## Problem Framing

The `ListSandboxPolicies` and `GetSandboxPolicyStatus` RPCs accept a `global` boolean flag that switches them from sandbox-scoped to gateway-global policy mode. The SDK's `Policy().List()` and `Policy().GetStatus()` methods don't expose this flag.

The openshell-dashboard BFF needs to fetch global policy revisions for the Platform Admin "Global Policy" page. Without the `global` flag, the SDK sends `name=""` which the gateway rejects with `InvalidArgument: name is required`.

## Affected RPCs

| RPC | Proto field | SDK method |
|-----|-----------|------------|
| `ListSandboxPolicies` | `global` (bool) | `Policy().List()` |
| `GetSandboxPolicyStatus` | `global` (bool) | `Policy().GetStatus()` |

## Approach

Add a `Global` option to the existing functional options pattern used by List and GetStatus:

```go
// List with global flag
revisions, err := client.Policy().List(ctx, "", openshell.WithGlobal(true))

// GetStatus with global flag
status, err := client.Policy().GetStatus(ctx, "", "", openshell.WithGlobal(true))
```

The workspace and name parameters would be ignored when `global=true`, matching the gateway's behavior. This follows the existing SDK pattern for optional parameters via functional options.

## Key Requirements

1. Add `Global` field to the options struct used by List and GetStatus
2. Add `WithGlobal(bool)` functional option constructor
3. Wire the `global` field into the proto request in the client implementation
4. Update the fake client to support global policy mode
5. Update converters if needed (the global field is on the request, not the response)
6. Skip validation of `name` when `global=true` (global mode doesn't need a sandbox name)

## Scope

This is a small, focused change. No new interfaces, no new sub-clients, no new types. Just adding an option flag and wiring it through to the proto request.
