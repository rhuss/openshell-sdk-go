# Quickstart: Workspace CRUD, GatewayInfo & GetCurrentUser

## Workspace Management

```go
client, _ := v1.NewClient(v1.Config{Address: "gateway:443"})
defer client.Close()

// Create a workspace
ws, _ := client.Workspaces().Create(ctx, "my-project", map[string]string{
    "team": "platform",
})

// Add a member
member, _ := client.Workspaces().AddMember(ctx, "my-project", "subject-123", v1.WorkspaceRoleAdmin)

// List workspaces
workspaces, _ := client.Workspaces().List(ctx, v1.ListOptions{Limit: 10})

// List members
members, _ := client.Workspaces().ListMembers(ctx, "my-project")

// Remove a member
_ = client.Workspaces().RemoveMember(ctx, "my-project", "subject-123")

// Delete workspace
_ = client.Workspaces().Delete(ctx, "my-project")
```

## Gateway Info & Current User

```go
// Get gateway metadata
info, _ := client.Health().GetGatewayInfo(ctx)
fmt.Printf("Version: %s, Status: %s\n", info.Version, info.Status)

for _, driver := range info.ComputeDrivers {
    fmt.Printf("Driver: %s (v%s)\n", driver.Name, driver.DriverVersion)
}

// Get authenticated identity
user, _ := client.Health().GetCurrentUser(ctx)
fmt.Printf("Logged in as: %s (%s)\n", user.DisplayName, user.Subject)
```

## Testing with Fake Client

```go
fc := fake.NewClient(
    fake.WithGatewayInfo(&types.GatewayInfo{
        Status:  types.ServiceStatusHealthy,
        Version: "1.2.3",
    }),
    fake.WithCurrentUser(&types.CurrentUser{
        Subject:     "test-user",
        DisplayName: "Test User",
        Roles:       []string{"admin"},
    }),
)

// Use fc.Workspaces() for workspace operations (in-memory store)
ws, _ := fc.Workspaces().Create(ctx, "test-ws", nil)
```
