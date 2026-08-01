# Quickstart: Global Policy Flag

## Listing Global Policy Revisions

```go
// List gateway-global policy revisions (no sandbox name needed)
revisions, err := client.Policy().List(ctx, "", v1.WithListGlobal(true))
if err != nil {
    log.Fatal(err)
}
for _, rev := range revisions {
    fmt.Printf("Version %d: %s\n", rev.Version, rev.Status)
}
```

## Getting Global Policy Status

```go
// Get status of a specific global policy version
status, err := client.Policy().GetStatus(ctx, "", "", v1.WithStatusGlobal(true), v1.WithVersion(3))
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Version %d status: %s\n", status.Revision.Version, status.Revision.LoadStatus)
```

## Combining with Pagination

```go
// List global revisions with pagination
revisions, err := client.Policy().List(ctx, "",
    v1.WithListGlobal(true),
    v1.WithLimit(10),
    v1.WithOffset(0),
)
```

## Testing with Fake Client

```go
fc := fake.NewClient()
defer fc.Close()

// List global revisions (fake supports global mode)
revisions, err := fc.Policy().List(ctx, "", v1.WithListGlobal(true))
```
