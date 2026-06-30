# OpenShell SDK for Go

A Go SDK for interacting with [OpenShell](https://github.com/NVIDIA/OpenShell)
servers, providing idiomatic Go bindings for shell session management, command
execution, provider configuration, and service exposure.

## Why a Go SDK?

Go is the language of the Kubernetes ecosystem. If you want to build an
operator, controller, or any automation that manages OpenShell resources as
native Kubernetes objects, you need a Go client.

This SDK is modeled after
[`k8s.io/client-go`](https://github.com/kubernetes/client-go), the standard
Kubernetes client library that every Go operator developer already knows. The
patterns will look familiar:

- **Typed sub-clients per resource**: `client.Sandboxes()`, `client.Providers()`,
  `client.Exec()`, just like `clientset.CoreV1().Pods()`
- **Domain types separated from wire formats**: clean Go structs in a `types`
  package, no proto leakage into the public API (like `k8s.io/api`)
- **Watch primitives**: channel-based watchers with `ResultChan()` and `Stop()`,
  identical to `watch.Interface` in client-go
- **Functional options**: variadic option patterns for list filtering,
  pagination, and watch configuration
- **Fake client for testing**: an in-memory implementation of the full client
  interface (like `k8s.io/client-go/kubernetes/fake`), so operators can be tested
  without a real gateway

The primary use case driving this SDK: building a Kubernetes operator for
OpenShell that manages sandbox lifecycles, provider configurations, and
execution policies as native Kubernetes resources.

## Architecture

```
Client
  ├── Sandboxes()   → SandboxInterface    (create, get, list, delete, watch, wait, logs)
  ├── Exec()        → ExecInterface       (run, stream, interactive)
  ├── Files()       → FileInterface       (upload, download)
  ├── Health()      → HealthInterface     (gateway health check)
  ├── Services()    → ServiceInterface    (expose, get, list, delete)
  ├── Providers()   → ProviderInterface   (CRUD + ensure)
  │     ├── Profiles() → ProfileInterface (list, get, import, update, lint, delete)
  │     └── Refresh()  → RefreshInterface (configure, status, rotate, delete)
  └── Policy()      → PolicyInterface     (draft review, approve, reject, merge, status)
```

All domain types live in `openshell/v1/types/`. Proto-to-Go conversions happen in
an internal converter layer. The public API surface uses type aliases so
consumers import a single package.

## Implemented Features

### Core SDK (Phase 1) ✓

| Feature | Interface | Methods |
|---------|-----------|---------|
| **Sandbox lifecycle** | `SandboxInterface` | `Create`, `Get`, `List`, `Delete`, `WaitReady`, `Watch`, `AttachProvider`, `DetachProvider`, `ListProviders` |
| **Command execution** | `ExecInterface` | `Run` (collected), `Stream` (chunked), `Interactive` (bidirectional PTY) |
| **Provider management** | `ProviderInterface` | `Create`, `Get`, `List`, `Update`, `Delete`, `Ensure` (idempotent create-or-update) |
| **File transfer** | `FileInterface` | `Upload`, `Download` |
| **Health checking** | `HealthInterface` | `Check` |
| **Real-time watch** | `WatchInterface[T]` | Generic typed event channel with `ResultChan()` and `Stop()` |
| **Typed errors** | `StatusError` | `IsNotFound`, `IsAlreadyExists`, `IsConflict`, `IsUnavailable`, `IsUnimplemented` |
| **Authentication** | `AuthProvider` | `NoAuth()`, `StaticToken()` |

### Operator API Extensions (Phase 2a) ✓

| Feature | Interface | Methods |
|---------|-----------|---------|
| **Service exposure** | `ServiceInterface` | `Expose`, `Get`, `List`, `Delete` |
| **Provider profiles** | `ProfileInterface` | `List`, `Get`, `Import`, `Update`, `Lint`, `Delete` |
| **Credential refresh** | `RefreshInterface` | `GetStatus`, `Configure`, `Rotate`, `Delete` |
| **StopOnTerminal watch** | `WatchOptions` | Auto-close watcher on terminal phase (Ready/Error) |

### Policy Management & Logs (Phase 2b-2) ✓

| Feature | Interface | Methods |
|---------|-----------|---------|
| **Draft policy review** | `PolicyInterface` | `GetDraft`, `ApproveDraftChunk`, `RejectDraftChunk`, `ApproveAllDraftChunks` |
| **Draft policy editing** | `PolicyInterface` | `ClearDraftChunks`, `EditDraftChunk`, `UndoDraftChunk`, `GetDraftHistory` |
| **Policy status & versions** | `PolicyInterface` | `GetStatus`, `List` |
| **Sandbox logs** | `SandboxInterface` | `GetLogs` (streaming log retrieval) |
| **Config merge operations** | `ConfigInterface` | `Update` with typed `PolicyMergeOperation` (6 variants) |
| **Conflict detection** | `IsConflict` | Covers optimistic concurrency and invalid state transitions |

### Fake Client for Testing ✓

The `fake` package provides a complete in-memory implementation of
`ClientInterface` that needs no gRPC server. Built for operator and
controller test suites.

| Feature | Description |
|---------|-------------|
| **Full interface compliance** | Implements `ClientInterface` with all sub-clients |
| **In-memory object stores** | Thread-safe stores with deep copy at boundaries |
| **Watch event broadcasting** | Mutations emit ADDED/MODIFIED/DELETED events to active watchers |
| **Automatic phase transitions** | `Create` → Provisioning, `WaitReady` → Ready |
| **Test fixture seeding** | `AddSandbox()`, `AddProvider()`, `AddPolicyDraft()` for pre-populating state |
| **StopOnTerminal support** | Fake watch auto-closes on terminal phases |
| **Race-detector safe** | All stores and broadcasters are `sync.Mutex`-protected |
| **Configurable health** | `WithHealthResult()` option for simulating unhealthy gateways |
| **Policy draft simulation** | In-memory draft chunks with approve/reject/undo lifecycle |

### Build and CI ✓

- Full GitHub Actions pipeline: golangci-lint v2, unit tests with race detection,
  build verification, proto generation checks
- Proto sync pipeline that pulls `.proto` files from upstream OpenShell and
  generates Go bindings
- `mise`-based task runner with a `Makefile` shim
- 40+ test files, 597 test functions

## Roadmap

- [x] **Phase 0** — Project scaffolding, CI pipeline, proto generation
- [x] **Phase 1** — Core SDK: Sandbox, Provider, Exec, File, Health, Watch
- [x] **Phase 1b** — Converter deduplication and domain type extraction
- [x] **Phase 1c** — Fake client package with in-memory stores and watch broadcaster
- [x] **Phase 2a** — Operator API: Services, Profiles, Credential Refresh, StopOnTerminal
- [x] **Phase 2b-1** — Config management: Get, Update with typed merge operations
- [x] **Phase 2b-2** — Policy management, logs, merge operations, ErrorConflict
- [ ] **Phase 2b-3** — SSH tunneling, TCP forwarding
- [ ] **Phase 3** — Enhanced watch: event streaming, server-side filtering

## Usage Examples

### Connect to a Gateway

```go
import v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"

client, err := v1.NewClient(v1.Config{
    Address: "gateway.example.com:443",
    Auth:    v1.StaticToken("my-token"),
})
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Create and Use a Sandbox

```go
// Create a sandbox with a Python image
sandbox, err := client.Sandboxes().Create(ctx, "my-sandbox", &v1.SandboxSpec{
    Template: &v1.SandboxTemplate{Image: "python:3.12"},
    Environment: map[string]string{"LANG": "en_US.UTF-8"},
}, nil)
if err != nil {
    log.Fatal(err)
}

// Wait until the sandbox is ready
sandbox, err = client.Sandboxes().WaitReady(ctx, sandbox.Name)
if err != nil {
    log.Fatal(err)
}

// Run a command
result, err := client.Exec().Run(ctx, sandbox.Name,
    []string{"python3", "-c", "print('hello from sandbox')"},
    v1.ExecOptions{},
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(result.Stdout))
```

### Stream Command Output

```go
stream, err := client.Exec().Stream(ctx, sandbox.Name,
    []string{"pip", "install", "numpy"},
    v1.ExecOptions{},
)
if err != nil {
    log.Fatal(err)
}
for {
    chunk, err := stream.Next()
    if err != nil {
        break
    }
    fmt.Print(string(chunk.Data))
}
exitCode, _ := stream.ExitCode()
fmt.Printf("Exit code: %d\n", exitCode)
```

### Interactive Session (PTY)

```go
session, err := client.Exec().Interactive(ctx, sandbox.Name,
    []string{"/bin/bash"}, 80, 24, // cols, rows
    v1.ExecOptions{},
)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// session implements io.Reader and io.Writer
session.Write([]byte("echo hello\n"))

buf := make([]byte, 1024)
n, _ := session.Read(buf)
fmt.Println(string(buf[:n]))
```

### Watch Sandbox State Changes

```go
watcher, err := client.Sandboxes().Watch(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
defer watcher.Stop()

for event := range watcher.ResultChan() {
    fmt.Printf("[%s] %s → phase: %s\n",
        event.Type,
        event.Object.Name,
        event.Object.Status.Phase,
    )
}
```

### Watch with Auto-Stop on Terminal Phase

```go
// Watcher closes automatically when sandbox reaches Ready or Error
watcher, err := client.Sandboxes().Watch(ctx, "my-sandbox",
    v1.WatchOptions{StopOnTerminal: true},
)
if err != nil {
    log.Fatal(err)
}
for event := range watcher.ResultChan() {
    fmt.Printf("phase: %s\n", event.Object.Status.Phase)
}
// channel is closed — sandbox reached a terminal state
```

### Manage Providers

```go
// Create or update a provider (idempotent)
provider, err := client.Providers().Ensure(ctx, &v1.Provider{
    Name: "openai",
    Type: "openai",
    Credentials: map[string]string{
        "api-key": "sk-...",
    },
})
if err != nil {
    log.Fatal(err)
}

// List all providers
providers, err := client.Providers().List(ctx)
for _, p := range providers {
    fmt.Printf("%s (type: %s)\n", p.Name, p.Type)
}
```

### Expose Sandbox Services

```go
// Expose an HTTP service running inside a sandbox
endpoint, err := client.Services().Expose(ctx,
    "my-sandbox", "api", 8080, true, // domain=true generates a public URL
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Service URL: %s\n", endpoint.URL)

// List all exposed services
endpoints, err := client.Services().List(ctx, "my-sandbox")
for _, ep := range endpoints {
    fmt.Printf("  %s → port %d (%s)\n", ep.ServiceName, ep.TargetPort, ep.URL)
}
```

### Provider Profiles

```go
// List available provider profiles
profiles, err := client.Providers().Profiles().List(ctx)
for _, p := range profiles {
    fmt.Printf("%s [%s]: %s\n", p.DisplayName, p.Category, p.Description)
}

// Validate a profile before importing
lintResult, err := client.Providers().Profiles().Lint(ctx, []v1.ProfileImportItem{{
    Source: "custom-llm.yaml",
    Profile: v1.ProviderProfile{
        DisplayName: "Custom LLM",
        Category:    v1.ProfileCategoryInference,
    },
}})
if !lintResult.Valid {
    for _, d := range lintResult.Diagnostics {
        fmt.Printf("[%s] %s: %s\n", d.Severity, d.Field, d.Message)
    }
}

// Import the profile
importResult, err := client.Providers().Profiles().Import(ctx, []v1.ProfileImportItem{{
    Source: "custom-llm.yaml",
    Profile: v1.ProviderProfile{
        DisplayName: "Custom LLM",
        Category:    v1.ProfileCategoryInference,
    },
}})
```

### Configure Credential Refresh

```go
// Set up OAuth2 credential refresh for a provider
status, err := client.Providers().Refresh().Configure(ctx, &v1.RefreshConfig{
    Provider:      "openai",
    CredentialKey:  "api-key",
    Strategy:      v1.RefreshStrategyOAuth2ClientCredentials,
    Material:      map[string]string{
        "client_id":     "xxx",
        "client_secret": "yyy",
        "token_url":     "https://auth.example.com/token",
    },
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Refresh status: %s (next: %s)\n", status.Status, status.NextRefreshAt)

// Trigger immediate rotation
status, err = client.Providers().Refresh().Rotate(ctx, "openai", "api-key")
```

### Review and Approve Draft Policy Chunks

```go
// Get the draft policy with pending chunks
draft, err := client.Policy().GetDraft(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Draft has %d pending chunks\n", len(draft.Chunks))

// Review and approve individual chunks
for _, chunk := range draft.Chunks {
    fmt.Printf("Chunk %s: %s\n", chunk.ID, chunk.Summary)
    result, err := client.Policy().ApproveDraftChunk(ctx, "my-sandbox", chunk.ID)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Approved, new version: %d\n", result.Version)
}

// Or approve all pending chunks at once
result, err := client.Policy().ApproveAllDraftChunks(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Approved %d chunks\n", result.ApprovedCount)
```

### Retrieve Sandbox Logs

```go
logs, err := client.Sandboxes().GetLogs(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
for _, entry := range logs {
    fmt.Printf("[%s] %s\n", entry.Timestamp.Format(time.RFC3339), entry.Message)
}
```

### Error Handling

```go
_, err := client.Sandboxes().Get(ctx, "does-not-exist")
if v1.IsNotFound(err) {
    fmt.Println("Sandbox not found")
}

_, err = client.Providers().Create(ctx, existingProvider)
if v1.IsAlreadyExists(err) {
    fmt.Println("Provider already exists, use Ensure() for idempotent upsert")
}

// Conflict covers both version conflicts and invalid state transitions
_, err = client.Policy().ApproveDraftChunk(ctx, "my-sandbox", alreadyApprovedChunkID)
if v1.IsConflict(err) {
    fmt.Println("Chunk already approved or rejected")
}
```

### File Transfer

```go
// Upload a local file into the sandbox
err := client.Files().Upload(ctx, sandbox.Name, "./script.py", "/workspace/script.py")

// Download results
err = client.Files().Download(ctx, sandbox.Name, "/workspace/output.csv", "./output.csv")
```

### Testing with the Fake Client

```go
import (
    v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
    "github.com/rhuss/openshell-sdk-go/openshell/v1/fake"
    "github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

func TestMyOperator(t *testing.T) {
    // Create a fake client — no gRPC server needed
    fc := fake.NewClient()
    defer fc.Close()

    // Pre-seed test fixtures
    fc.AddProvider(&types.Provider{
        Name: "openai",
        Type: "openai",
    })

    // Use it exactly like the real client
    var client v1.ClientInterface = fc

    providers, err := client.Providers().List(ctx)
    require.NoError(t, err)
    assert.Len(t, providers, 1)

    // Create triggers automatic phase transitions
    sandbox, err := client.Sandboxes().Create(ctx, "test-sb", &v1.SandboxSpec{
        Template: &v1.SandboxTemplate{Image: "python:3.12"},
    }, nil)
    require.NoError(t, err)
    assert.Equal(t, "Provisioning", string(sandbox.Status.Phase))

    // WaitReady transitions to Ready
    sandbox, err = client.Sandboxes().WaitReady(ctx, "test-sb")
    require.NoError(t, err)
    assert.Equal(t, "Ready", string(sandbox.Status.Phase))
}
```

### Testing with Watch Events

```go
func TestWatchEvents(t *testing.T) {
    fc := fake.NewClient()
    defer fc.Close()

    ctx := context.Background()

    // Start watching before creating
    watcher, err := fc.Sandboxes().Watch(ctx, "my-sb")
    require.NoError(t, err)
    defer watcher.Stop()

    // Create triggers an ADDED event
    _, err = fc.Sandboxes().Create(ctx, "my-sb", &v1.SandboxSpec{
        Template: &v1.SandboxTemplate{Image: "python:3.12"},
    }, nil)
    require.NoError(t, err)

    event := <-watcher.ResultChan()
    assert.Equal(t, "ADDED", string(event.Type))
    assert.Equal(t, "my-sb", event.Object.Name)
}
```

### Testing Unhealthy Gateway

```go
func TestUnhealthyGateway(t *testing.T) {
    fc := fake.NewClient(
        fake.WithHealthResult(&types.HealthResult{
            Healthy: false,
            Version: "1.2.3",
        }),
    )
    defer fc.Close()

    result, err := fc.Health().Check(ctx)
    require.NoError(t, err)
    assert.False(t, result.Healthy)
}
```

## Documentation

Full API documentation is available at the [OpenShell Go SDK Docs](https://ro14nd.de/openshell-sdk-go/) site, built with [mdBook](https://rust-lang.github.io/mdBook/).

Key sections:

- **[Getting Started](https://ro14nd.de/openshell-sdk-go/getting-started.html)** - Installation, connecting, first sandbox
- **[Error Handling](https://ro14nd.de/openshell-sdk-go/error-handling.html)** - `StatusError`, predicate functions, retry patterns
- **[Testing](https://ro14nd.de/openshell-sdk-go/testing.html)** - Fake client, fixture seeding, watch events
- **[API Reference](https://ro14nd.de/openshell-sdk-go/api/overview.html)** - Per-resource method tables and code examples

To build the docs locally:

```bash
# Install mdBook
cargo install mdbook

# Build the docs
mdbook build docs

# Serve locally with live reload
mdbook serve docs
```

## Quick Start

### Prerequisites

- Go 1.23 or later
- [mise](https://mise.jdx.dev) (recommended for reproducible builds)

### Build and Test

```bash
# Clone the repository
git clone https://github.com/rhuss/openshell-sdk-go.git
cd openshell-sdk-go

# Run tests
make test

# Run linter
make lint

# Run full CI pipeline (lint + build + test)
make ci
```

If you don't have mise installed, `make` will print installation instructions.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, build commands,
and contribution guidelines.

## License

Apache-2.0. See [LICENSE](LICENSE) for details.

Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
