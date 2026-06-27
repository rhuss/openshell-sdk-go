# OpenShell SDK for Go

A Go SDK for interacting with [OpenShell](https://github.com/NVIDIA/OpenShell)
servers, providing idiomatic Go bindings for shell session management.

> **Status**: Early development (Phase 0 scaffolding). Not yet functional.

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

If you don't have mise installed, `make` will print installation
instructions.

## Project Structure

```
openshell-sdk-go/
├── openshell/          # SDK package
│   ├── client.go       # Client type and Dial/Close
│   └── client_test.go  # Unit tests
├── mise.toml           # Tool versions and task definitions
├── Makefile            # Build shim (delegates to mise)
├── .golangci.yml       # Linter configuration
└── .github/workflows/  # CI pipeline
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, build
commands, and contribution guidelines.

## License

Apache-2.0. See [LICENSE](LICENSE) for details.

Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
