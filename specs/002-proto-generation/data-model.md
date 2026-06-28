# Data Model: Proto Generation Pipeline

**Date**: 2026-06-27
**Branch**: `002-proto-generation`

## Entities

### Proto Source File

A `.proto` file copied from the upstream OpenShell repository. Not modified
by the SDK; used as input to protoc.

| Attribute | Type | Description |
|-----------|------|-------------|
| filename | string | File name (e.g., `openshell.proto`) |
| package | string | Proto package declaration (e.g., `openshell.v1`) |
| location | path | `proto/<filename>` in the SDK repo |
| upstream_path | path | Original location in OpenShell repo |

**Instances** (fixed set):

| File | Proto Package | Has Service | Imports |
|------|--------------|-------------|---------|
| openshell.proto | openshell.v1 | Yes (55 RPCs) | datamodel.proto, sandbox.proto, google/protobuf/struct.proto |
| datamodel.proto | openshell.datamodel.v1 | No | (none) |
| sandbox.proto | openshell.sandbox.v1 | No | (none) |

### Generated Go Package

A directory containing Go source files produced by protoc from a proto source
file. Each proto file maps to exactly one Go package.

| Attribute | Type | Description |
|-----------|------|-------------|
| directory | path | `proto/<package_slug>/` |
| go_package | string | Full Go import path |
| source_proto | string | Proto file this was generated from |
| files | list | Generated `.pb.go` and `_grpc.pb.go` files |

**Instances** (fixed mapping):

| Proto File | Go Package Dir | Go Import Path | Generated Files |
|------------|---------------|----------------|-----------------|
| openshell.proto | proto/openshellv1/ | github.com/rhuss/openshell-sdk-go/proto/openshellv1 | openshell.pb.go, openshell_grpc.pb.go |
| datamodel.proto | proto/datamodelv1/ | github.com/rhuss/openshell-sdk-go/proto/datamodelv1 | datamodel.pb.go |
| sandbox.proto | proto/sandboxv1/ | github.com/rhuss/openshell-sdk-go/proto/sandboxv1 | sandbox.pb.go |

### Upstream Version Marker

A metadata file tracking which upstream commit the proto files were synced from.

| Attribute | Type | Description |
|-----------|------|-------------|
| location | path | `proto/UPSTREAM_VERSION` |
| content | string | Git commit SHA (40 hex chars) or "unknown" |

## Relationships

```
Proto Source File --[generates]--> Generated Go Package  (1:1)
Proto Source File --[imports]----> Proto Source File      (openshell imports datamodel, sandbox)
Generated Go Package --[imports]--> Generated Go Package (openshellv1 imports datamodelv1, sandboxv1)
Upstream Version Marker --[tracks]--> Proto Source Files  (1:many, all share same upstream commit)
```

## Package Path Mapping

The `--go_opt=M` flags establish this mapping at generation time:

```
--go_opt=Mopenshell.proto=github.com/rhuss/openshell-sdk-go/proto/openshellv1
--go_opt=Mdatamodel.proto=github.com/rhuss/openshell-sdk-go/proto/datamodelv1
--go_opt=Msandbox.proto=github.com/rhuss/openshell-sdk-go/proto/sandboxv1
--go-grpc_opt=Mopenshell.proto=github.com/rhuss/openshell-sdk-go/proto/openshellv1
--go-grpc_opt=Mdatamodel.proto=github.com/rhuss/openshell-sdk-go/proto/datamodelv1
--go-grpc_opt=Msandbox.proto=github.com/rhuss/openshell-sdk-go/proto/sandboxv1
```
