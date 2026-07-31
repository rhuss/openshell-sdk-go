# Data Model: Proto Sync from Upstream PR #2445

**Date**: 2026-07-31

## Entities

This feature does not introduce new SDK domain types. It updates proto-level message definitions only. The data model impact is limited to the proto layer.

### Proto Files (Source Artifacts)

| File | Package | Status | Description |
|------|---------|--------|-------------|
| `openshell.proto` | openshell | Updated | Core SDK service, gains auth annotations + workspace fields |
| `inference.proto` | inference | New | Inference service with 4 RPCs |
| `options.proto` | options | Updated | Gains `AuthorizationRule` message and method extension |
| `datamodel.proto` | datamodel | Sync | Data model types, may be unchanged |
| `sandbox.proto` | sandbox | Sync | Sandbox service, may be unchanged |

### Generated Packages (Output Artifacts)

| Package | Directory | Status | Contents |
|---------|-----------|--------|----------|
| `openshellv1` | `proto/openshellv1/` | Regenerated | `.pb.go`, `_grpc.pb.go` |
| `inferencev1` | `proto/inferencev1/` | New | `.pb.go`, `_grpc.pb.go` |
| `optionsv1` | `proto/optionsv1/` | Regenerated | `.pb.go` |
| `datamodelv1` | `proto/datamodelv1/` | Regenerated | `.pb.go` |
| `sandboxv1` | `proto/sandboxv1/` | Regenerated | `.pb.go` |

### Key New Proto Types (from PR #2445)

- **AuthorizationRule**: Message in `options.proto` defining permission annotations for RPCs
- **MethodOptions extension (50000)**: Custom method option carrying `AuthorizationRule`
- **Workspace fields**: New fields on request messages in `openshell.proto` for workspace scoping

### Relationships

```
inference.proto --> options.proto (imports AuthorizationRule annotations)
openshell.proto --> options.proto (imports AuthorizationRule annotations)
openshell.proto --> datamodel.proto (imports data model types)
openshell.proto --> sandbox.proto (imports sandbox types)
```

## State Transitions

N/A. No new state machines or lifecycle changes.

## Validation Rules

N/A. No new SDK-level validation. Proto-level validation is handled by protobuf serialization.
