# Data Model: API Documentation Site

This feature produces static content (markdown files, CSS, configuration).
There is no runtime data model. The entities below describe the content
structure and relationships.

## Content Entities

### mdBook Project

- **book.toml**: Project configuration (title, authors, language, build
  output directory, preprocessors)
- **SUMMARY.md**: Defines the sidebar navigation tree. Each line maps to
  a chapter (markdown file). Nesting creates hierarchy.
- **theme/custom.css**: CSS overrides applied after default mdBook styles.

### Guide Page

A narrative markdown file addressing a specific developer concern.

| Field | Description |
|-------|-------------|
| Title | Page heading (H1) |
| Purpose | What the developer learns |
| Code examples | Compilable Go snippets |
| Prerequisites | What the reader should know first |

**Instances**: getting-started, architecture, error-handling, testing

### API Reference Page (Hero)

A detailed page for a core SDK interface with full gRPC mapping.

| Field | Description |
|-------|-------------|
| Interface name | Go interface type (e.g., SandboxInterface) |
| Sub-client accessor | How to obtain it (e.g., client.Sandboxes()) |
| Methods | Each method with signature, description, SDK code, gRPC proto |
| Related types | Input/output types used by the methods |

**Instances**: Sandboxes, Exec, Providers

### API Reference Page (Standard)

A concise page for a non-hero SDK interface with a reference table.

| Field | Description |
|-------|-------------|
| Interface name | Go interface type (e.g., ConfigInterface) |
| Sub-client accessor | How to obtain it (e.g., client.Config()) |
| Reference table | Rows: SDK method, gRPC RPC name, proto file |
| Usage example | One representative code snippet |

**Instances**: Profiles, Refresh, Services, Files, Health, SSH, TCP,
Config, Policy

### Example Test Function

A Go test function following the `Example*` naming convention.

| Field | Description |
|-------|-------------|
| Function name | `Example` + subject (e.g., `ExampleClient_Sandboxes_Create`) |
| Package | `v1_test` (external test package) |
| Dependencies | fake client package for compilation |
| Output comment | Expected output annotation for pkg.go.dev rendering |

**File locations**: `openshell/v1/example_test.go`,
`openshell/v1/example_fake_test.go`

### GitHub Actions Workflow

| Field | Description |
|-------|-------------|
| Trigger | Push to main, path filter on docs/ and example files |
| Steps | Install mdBook, install mdbook-admonish, build, deploy |
| Deployment target | GitHub Pages (gh-pages branch) |

## Relationships

```
book.toml
  └─ references SUMMARY.md (navigation structure)
       ├─ Guide Pages (getting-started, architecture, error-handling, testing)
       └─ API Reference Pages
            ├─ Hero Pages (sandboxes, exec, providers) - full gRPC mapping
            └─ Standard Pages (10 interfaces) - reference tables

Example Test Functions (separate from docs site)
  └─ compile against openshell/v1 package
  └─ use fake client for no-network execution

GitHub Actions Workflow
  └─ builds docs/ with mdBook
  └─ deploys to GitHub Pages
```
