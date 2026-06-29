# Refresh

Accessor: `client.Providers().Refresh()`

Manage credential refresh schedules for provider profiles. Configure automatic
rotation of API keys and monitor refresh status.

## Methods

| SDK Method | gRPC RPC | Proto File | Description |
|------------|----------|------------|-------------|
| `GetStatus(ctx, provider, credentialKey string)` | `GetProviderRefreshStatus` | `openshell.proto` | Get refresh status for a credential |
| `Configure(ctx, config *RefreshConfig)` | `ConfigureProviderRefresh` | `openshell.proto` | Set up automatic credential refresh |
| `Rotate(ctx, provider, credentialKey string)` | `RotateProviderCredential` | `openshell.proto` | Manually trigger a credential rotation |
| `Delete(ctx, provider, credentialKey string)` | `DeleteProviderRefresh` | `openshell.proto` | Remove a refresh configuration |

## GetStatus

Check the refresh status for a specific provider credential.

```go
statuses, err := client.Providers().Refresh().GetStatus(ctx, "openai", "default")
if err != nil {
    log.Fatal(err)
}
for _, s := range statuses {
    fmt.Printf("Key: %s, Last refresh: %s, Next: %s\n",
        s.CredentialKey, s.LastRefresh, s.NextRefresh)
}
```

## Configure

Set up automatic credential refresh with a defined schedule.

```go
status, err := client.Providers().Refresh().Configure(ctx, &v1.RefreshConfig{
    Provider:      "openai",
    CredentialKey:  "default",
    IntervalHours:  24,
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Refresh configured, next rotation: %s\n", status.NextRefresh)
```

## Rotate

Manually trigger an immediate credential rotation.

```go
status, err := client.Providers().Refresh().Rotate(ctx, "openai", "default")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Rotated successfully at %s\n", status.LastRefresh)
```

See also: [Error Handling](../error-handling.md), [Profiles](profiles.md)
