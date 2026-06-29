# Proposal: Go SDK for OpenShell (client-go style)

> GitHub issue draft for NVIDIA/OpenShell. Agenda entry for the June 30
> contributor meeting.

## Motivation

If you want to build a Kubernetes operator for OpenShell, you need a Go SDK.
Go is the language operators are written in, and the Kubernetes ecosystem has
established patterns that Go developers expect: typed clients per resource,
domain types separated from wire formats, functional options, and
watch/streaming primitives.

This SDK follows those patterns directly. It models itself after
`k8s.io/client-go`: each API resource gets its own typed sub-client, domain
types live in a dedicated package (like `k8s.io/api`), and proto-to-Go
conversions happen behind clean interfaces. Anyone who has built a Kubernetes
controller or operator will recognize the structure immediately.

The primary use case driving this: building an OpenShell operator for
OpenShift/Kubernetes that manages sandbox lifecycles, provider configurations,
and execution policies as native Kubernetes resources.

## Scope

The SDK covers the full gRPC API surface of OpenShell. This is not a thin
wrapper around generated proto stubs. It provides:

- **Idiomatic Go domain types** separated from protobuf wire types
- **Typed error handling** with gRPC status code mapping to SDK-specific error types
- **Internal converter layer** that translates between proto messages and Go domain types
- **Watch/streaming support** for real-time resource updates

## Current Implementation

I have built an initial version that covers the Phase 1 API surface. The
foundation is solid and ready to build on:

**Sub-clients:**
- `Sandbox` — create, get, list, delete, watch sandboxes
- `Provider` — manage provider configurations
- `Exec` — execute commands in sandboxes
- `File` — file operations (read, write, list)
- `Health` — gateway health checks

**Testing infrastructure:**
- `fake` package for unit testing without a gRPC server (same concept as
  `k8s.io/client-go/fake`), with an in-memory object store and watch event
  broadcaster
- Test suite across all sub-clients and converters, running with the race
  detector enabled

**Build and CI:**
- Full GitHub Actions pipeline: golangci-lint v2, unit tests with race
  detection, build verification, proto generation checks
- Proto sync pipeline that pulls `.proto` files from upstream OpenShell and
  generates Go bindings
- `mise`-based task runner with a `Makefile` shim for convenience

## Planned Features

- **Phase 2a:** Services, Profiles, credential refresh
- **Phase 2b:** Policy, Config, SSH tunneling, TCP forwarding
- **Enhanced watch model:** Log and event streaming, server-side filtering

## Development Approach

The SDK was built using a spec-driven development workflow
([Speckit](https://github.com/speckit/speckit) /
[cc-spex](https://github.com/rhuss/cc-spex)). This is not a strict
requirement, but it proved very useful in practice: each feature starts as a
structured specification, goes through a planning phase, and then gets
implemented against well-defined tasks. The specs live alongside the code and
serve as living documentation of design decisions.

Code reviews run through multiple automated reviewers: CodeRabbit, GitHub
Copilot, and Devin. Having three independent AI reviewers catch different
classes of issues and keep the code quality high from the start. Combined with
the CI pipeline and the spec-driven workflow, this creates a tight feedback loop
that scales well even as a solo contributor.

## Current State

The SDK currently lives in a private repository. I would be happy to contribute
it to the NVIDIA org or wherever the community thinks it belongs.

Go SDK support was already mentioned in the 1.0 stability discussion (SDK
Interfaces listed alongside Python and TypeScript). This could serve as the
foundation for that effort.

I plan to present this at the next contributor meeting (June 30) and would
welcome feedback on the approach, API design, and where this fits in the
project roadmap.
