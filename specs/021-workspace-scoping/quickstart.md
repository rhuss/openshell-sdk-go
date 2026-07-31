# Quickstart: Workspace Scoping

## Before (current API)

```go
client, _ := openshell.NewClient(ctx, endpoint)

sandbox, _ := client.Sandboxes().Create(ctx, "my-sandbox", spec, labels)
sandboxes, _ := client.Sandboxes().List(ctx)
provider, _ := client.Providers().Get(ctx, "my-provider")
```

## After (with workspace scoping)

```go
client, _ := openshell.NewClient(ctx, endpoint)

// All operations now require a workspace (first param after ctx)
sandbox, _ := client.Sandboxes().Create(ctx, "team-alpha", "my-sandbox", spec, labels)
sandboxes, _ := client.Sandboxes().List(ctx, "team-alpha")
provider, _ := client.Providers().Get(ctx, "team-alpha", "my-provider")

// Empty string = "default" workspace
sandbox, _ = client.Sandboxes().Get(ctx, "", "my-sandbox")

// Cross-workspace listing for admins
allSandboxes, _ := client.Sandboxes().List(ctx, "", openshell.ListOptions{AllWorkspaces: true})
```

## Testing with the Fake Client

```go
client := fake.NewClient()

// Resources are isolated per workspace
client.Sandboxes().Create(ctx, "ws-a", "sandbox-1", spec, nil)
client.Sandboxes().Create(ctx, "ws-b", "sandbox-2", spec, nil)

listA, _ := client.Sandboxes().List(ctx, "ws-a")  // returns only sandbox-1
listB, _ := client.Sandboxes().List(ctx, "ws-b")  // returns only sandbox-2

// AllWorkspaces returns both
listAll, _ := client.Sandboxes().List(ctx, "", openshell.ListOptions{AllWorkspaces: true})
// returns sandbox-1 and sandbox-2
```
