# Research: Converter Code Deduplication

## R1: Duplicated Conversion Functions

**Decision**: Replace all 13 duplicated unexported conversion functions in `*_client.go` with calls to the existing exported converter package.

**Mapping of duplicates**:

| Client file (unexported) | Converter package (exported) |
|---|---|
| `sandbox_client.go: sandboxFromProto` | `converter.SandboxFromProto` |
| `sandbox_client.go: sandboxSpecFromProto` | (inlined in `converter.SandboxFromProto`) |
| `sandbox_client.go: sandboxStatusFromProto` | (inlined in `converter.SandboxFromProto`) |
| `sandbox_client.go: sandboxPhaseFromProto` | `converter.SandboxPhaseFromProto` |
| `sandbox_client.go: sandboxSpecToProto` | `converter.SandboxSpecToProto` |
| `provider_client.go: providerFromProto` | `converter.ProviderFromProto` |
| `provider_client.go: providerToProto` | `converter.ProviderToProto` |
| `provider_client.go: timeFromMillis` | `converter.TimeFromMillis` |
| `provider_client.go: millisFromTime` | `converter.MillisFromTime` |
| `exec_client.go: execChunkFromEvent` | `converter.ExecChunkFromEvent` |
| `exec_client.go: execRequestToProto` | `converter.ExecRequestToProto` |
| `exec_client.go: execInteractiveRequestToProto` | `converter.ExecInteractiveRequestToProto` |
| `exec_client.go: execResultFromEvents` | `converter.ExecResultFromEvents` |

**Helper functions** (`copyStringMap`, `copyBoolPtr`, `copyStringSlice`) in `sandbox_client.go` are not duplicated in the converter. They should move to the converter package since they support deep-copy at boundaries (Constitution VII).

**Rationale**: The converter package already has complete, tested implementations. The client duplicates were created because `v1/` cannot import `v1/internal/converter/` (circular import via `v1` types).

## R2: Types Package Location

**Decision**: `openshell/v1/types/` (Go package name: `types`)

**Rationale**: Types are scoped to the v1 API surface. Placing them under `v1/` makes the relationship clear and follows the `k8s.io/api/core/v1` pattern where types live alongside their API version.

**Alternatives considered**:
- `openshell/types/`: Version-agnostic, but these types are v1-specific (field names, enum values match v1 proto schema). A v2 API would likely define different types.

## R3: Which Types Move

**Decision**: Move domain data types only. Keep client operation interfaces in `v1/`.

**Types that move** (used by converter or are domain data):
- Structs: `Sandbox`, `SandboxSpec`, `SandboxTemplate`, `SandboxStatus`, `SandboxCondition`, `AttachProviderResult`, `DetachProviderResult`, `Provider`, `ProviderSpec`, `ExecResult`, `ExecChunk`, `HealthResult`, `Config`, `TLSConfig`, `RetryPolicy`
- Options: `CreateOptions`, `DeleteOptions`, `ExecOptions`, `GetOptions`, `ListOptions`, `UpdateOptions`, `WaitOptions`, `WatchOptions`
- Enums/constants: `ErrorCode` + error code constants, `SandboxPhase` + phase constants, `EventType` + event constants, `StreamType` + stream constants
- Error type: `StatusError` + `IsStatusError` function
- Generics: `Event[T]`
- Interfaces used in domain types: `Logger`, `AuthProvider`
- Watch: `WatchInterface[T]`

**Types that stay in `v1/`** (client logic):
- `Client`, `ClientInterface`
- `SandboxInterface`, `ExecInterface`, `ProviderInterface`, `HealthInterface`, `FileInterface`
- `ExecStream`, `InteractiveSession`
- All `*Client` implementations (`sandboxClient`, `execClient`, etc.)

**Rationale**: The converter only needs domain data types. Client operation interfaces define the API surface and reference client methods, not domain concepts. Keeping them in `v1/` avoids moving method signatures that are inherently client-specific.

## R4: Re-export Strategy

**Decision**: Type aliases in `v1/` for all moved types.

```go
type Sandbox = types.Sandbox
type Config = types.Config
const SandboxPhaseReady = types.SandboxPhaseReady
```

**Rationale**: Go type aliases (`type X = Y`) are fully transparent — struct literals, type assertions, interface satisfaction all work unchanged. This is the standard Go approach for backward-compatible package splits (used by `golang.org/x/` packages).

**Note**: Constants cannot use type aliases. They must be re-declared with the same value or re-exported as `var` (which changes mutability semantics). The cleanest approach is `const X = types.X` which works for string and integer constants.

## R5: Converter Signature Changes

**Decision**: The converter package function signatures change only in their import path for types. Function names and semantics remain identical.

Before: `func SandboxFromProto(s *pb.Sandbox) *v1.Sandbox`
After: `func SandboxFromProto(s *pb.Sandbox) *types.Sandbox`

The converter's import changes from:
```go
v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
```
to:
```go
"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
```

All converter test files need the same import update.
