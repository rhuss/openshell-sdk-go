# Quickstart: Inference Route Client

## Setting an Inference Route

```go
route, err := client.Inference().SetRoute(ctx, "my-workspace", &v1.InferenceRouteConfig{
    ProviderName: "openai",
    ModelID:      "gpt-4",
    RouteName:    "",       // empty = default user-facing route
    TimeoutSecs:  120,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Route set: provider=%s model=%s version=%d\n",
    route.ProviderName, route.ModelID, route.Version)
```

## Getting an Inference Route

```go
route, err := client.Inference().GetRoute(ctx, "my-workspace", "")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Current route: %s/%s (timeout %ds)\n",
    route.ProviderName, route.ModelID, route.TimeoutSecs)
```

## Deleting an Inference Route

```go
err := client.Inference().DeleteRoute(ctx, "my-workspace", "")
if err != nil {
    log.Fatal(err)
}
```

## Testing with the Fake Client

```go
fc := fake.NewClient()
defer fc.Close()

// Set a route (stored in memory)
_, err := fc.Inference().SetRoute(ctx, "test-ws", &v1.InferenceRouteConfig{
    ProviderName: "local-provider",
    ModelID:      "test-model",
})
// Get it back
route, err := fc.Inference().GetRoute(ctx, "test-ws", "")
// Delete it
err = fc.Inference().DeleteRoute(ctx, "test-ws", "")
```
