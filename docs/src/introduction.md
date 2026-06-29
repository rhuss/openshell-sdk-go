# OpenShell Go SDK

The OpenShell Go SDK provides an idiomatic Go client for the OpenShell gateway API. It wraps the underlying gRPC protocol behind typed interfaces, making it straightforward to manage sandboxes, execute commands, handle providers, and more.

## Key Features

- **Sub-client pattern**: A single `Client` provides typed accessors for each API domain (Sandboxes, Exec, Providers, Files, Health, SSH, TCP, Config, Policy, Services)
- **Proto isolation**: SDK types are decoupled from protobuf wire types, so your code never imports generated proto packages
- **Fake client for testing**: An in-memory implementation of the full `ClientInterface` for testing without a live gateway
- **Watch and streaming**: First-class support for watching sandbox state changes and streaming command output
- **Typed error handling**: Functions like `IsNotFound`, `IsAlreadyExists`, and `IsConflict` for precise error classification

## Where to Start

If you are new to the SDK, the [Quick Start](getting-started.md) guide walks you through installation, connecting to a gateway, creating your first sandbox, running a command, and cleaning up.

For a deeper understanding of how the SDK is structured, see the [Architecture](architecture.md) overview.

## API Reference

Every SDK interface has a dedicated reference page. The three core interfaces (Sandboxes, Exec, Providers) include side-by-side SDK and gRPC code blocks showing how each method maps to the underlying RPC. All other interfaces include reference tables with the same mapping.

Browse the full [API Overview](api/overview.md) to see all 13 interfaces at a glance.

## Guides

- [Error Handling](error-handling.md): StatusError, typed error checks, retry patterns
- [Testing](testing.md): Fake client usage, fixture seeding, watch event testing

## Links

- [GitHub Repository](https://github.com/rhuss/openshell-sdk-go)
- [pkg.go.dev Reference](https://pkg.go.dev/github.com/rhuss/openshell-sdk-go/openshell/v1)
- [OpenShell Project](https://github.com/NVIDIA/OpenShell)
