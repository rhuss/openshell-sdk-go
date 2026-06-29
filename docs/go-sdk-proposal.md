# Proposal: Go SDK for OpenShell (client-go style)

> GitHub issue draft for NVIDIA/OpenShell.

## Why

[#1719](https://github.com/NVIDIA/OpenShell/issues/1719) puts a Kubernetes operator on the roadmap. You can write an operator in any language, but Go is the natural fit for the Kubernetes ecosystem, and a Go client for the OpenShell gRPC API would make that path much smoother.

The Kubernetes world runs on Go. Operators, controllers, admission webhooks, CLI tools, platform integrations: they're overwhelmingly written in Go and they all expect client libraries that follow the `k8s.io/client-go` conventions. Typed sub-clients per resource, domain types in a dedicated package, watch primitives with channels, functional options, and a fake client for testing. These patterns exist because they solve real problems: typed watchers keep controllers reliable, fake clients make tests fast and deterministic, and domain types separated from proto keep the public API clean when the wire format changes.

A Go SDK that follows these conventions means anyone building on top of OpenShell in the Kubernetes world can pick it up without learning a new paradigm. The patterns are the same ones they already use every day.

## What Would This Look Like

A Go SDK covering the full OpenShell gRPC API surface, following `k8s.io/client-go` conventions:

- **Typed sub-clients** per resource: `client.Sandboxes()`, `client.Providers()`, `client.Exec()`, `client.Services()`, similar to `clientset.CoreV1().Pods()`
- **Domain types** in a dedicated `types` package, no proto leakage into the public API
- **Watch primitives** with `ResultChan()` and `Stop()`, matching `watch.Interface` from client-go
- **Typed errors** with helpers: `IsNotFound()`, `IsAlreadyExists()`, `IsConflict()`
- **Fake client** for testing without a gRPC server, with in-memory stores and watch event broadcasting
- **Internal converter layer** handling all proto-to-domain translations

The API surface would map to something like this:

```
Client
  ├── Sandboxes()   → SandboxInterface    (create, get, list, delete, watch, wait)
  ├── Exec()        → ExecInterface       (run, stream, interactive)
  ├── Files()       → FileInterface       (upload, download)
  ├── Health()      → HealthInterface     (gateway health check)
  ├── Services()    → ServiceInterface    (expose, get, list, delete)
  └── Providers()   → ProviderInterface   (CRUD + ensure)
        ├── Profiles() → ProfileInterface (list, get, import, update, lint, delete)
        └── Refresh()  → RefreshInterface (configure, status, rotate, delete)
```

## Example Usage

```go
client, err := v1.NewClient(v1.Config{
    Address: "gateway.example.com:443",
    Auth:    v1.StaticToken("my-token"),
})
defer client.Close()

// Create a sandbox and wait for it
sandbox, _ := client.Sandboxes().Create(ctx, "my-sb", &v1.SandboxSpec{
    Template: &v1.SandboxTemplate{Image: "python:3.12"},
}, nil)
sandbox, _ = client.Sandboxes().WaitReady(ctx, sandbox.Name)

// Run a command
result, _ := client.Exec().Run(ctx, sandbox.Name,
    []string{"python3", "-c", "print('hello')"},
    v1.ExecOptions{},
)
```

Testing with a fake client (no gRPC server needed):

```go
func TestMyController(t *testing.T) {
    fc := fake.NewClient()
    defer fc.Close()

    var client v1.ClientInterface = fc

    sandbox, err := client.Sandboxes().Create(ctx, "test-sb", &v1.SandboxSpec{
        Template: &v1.SandboxTemplate{Image: "python:3.12"},
    }, nil)
    require.NoError(t, err)

    sandbox, err = client.Sandboxes().WaitReady(ctx, "test-sb")
    require.NoError(t, err)
    assert.Equal(t, "Ready", string(sandbox.Status.Phase))
}
```

## Existing Implementation

As a side note: I have already built a fully functional Go SDK following these ideas. It covers the Phase 1 and Phase 2a API surfaces (Sandbox, Provider, Exec, File, Health, Watch, Services, Profiles, Credential Refresh), includes a complete fake client package, CI pipeline, and a proto sync workflow. 32 test files, 368 test functions. It currently lives in a private repo, but I'd be happy to contribute it as a starting point if the community thinks this direction is right.

## Discussion

Go SDK support was already mentioned in the 1.0 stability discussion (SDK Interfaces listed alongside Python and TypeScript). I'd love to hear whether this direction makes sense, and where a Go SDK should live relative to the main project.
