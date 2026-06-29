# Implementation Plan: Phase 2b-2 Policy, Logs, MergeOps, ErrorConflict

**Branch**: `008-policy-logs` | **Date**: 2026-06-29 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `specs/008-policy-logs/spec.md`

## Summary

Add PolicyInterface (10 RPCs for draft-review-approve policy workflow), extend SandboxInterface with GetLogs, replace opaque MergeOperations bytes with typed structs, and add ErrorConflict error code for optimistic concurrency. Follows the established client-go sub-client pattern with interfaces, converters, typed errors, and fake stubs.

## Technical Context

**Language/Version**: Go 1.23+
**Primary Dependencies**: gRPC, protobuf (existing), testify (test only)
**Storage**: N/A (SDK is a gRPC client)
**Testing**: `go test ./... -race`, `mise run ci`
**Target Platform**: Linux/macOS/Windows (Go cross-platform library)
**Project Type**: Library (Go SDK)
**Performance Goals**: N/A (thin gRPC wrapper, no caching or batching)
**Constraints**: No new dependencies (Constitution V)
**Scale/Scope**: 10 new PolicyInterface methods, 1 GetLogs extension, 6 MergeOperation types, 1 new error code, ~20 new domain types, ~15 new files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Proto Isolation | PASS | All types in v1/types/, converters in internal/converter/ |
| II. Idiomatic Go | PASS | Functional options, context propagation, error returns |
| III. Test-First | PASS | Tests written alongside each implementation file |
| IV. Upstream Tracking | PASS | Maps directly from proto/openshell.proto RPCs |
| V. Minimal Dependencies | PASS | No new dependencies |
| VI. Secrets Never Leak | PASS | No credential fields in policy/log types |
| VII. Deep Copy at Boundaries | PASS | All converters deep-copy maps, slices, nested structs |
| VIII. Doc Examples Compile | PASS | Doc comments on all exported symbols |
| IX. Agent-Friendly Documentation | PASS | Error codes listed in interface method docs |
| X. Proto-SDK Naming Fidelity | PASS | Field names match proto semantics |
| XI. Fake-Real Parity | PASS | Fake Policy returns Unimplemented; fake Config accepts MergeOps |
| XII. Graceful Shutdown Order | N/A | No streaming/lifecycle in this spec |

## Project Structure

### Documentation (this feature)

```text
specs/008-policy-logs/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── spec.md              # Feature specification
├── checklists/
│   └── requirements.md  # Spec quality checklist
└── tasks.md             # Phase 2 output (via /speckit-tasks)
```

### Source Code (repository root)

```text
openshell/v1/
├── client.go                          # Add Policy() accessor + policy field
├── config_client.go                   # Remove MergeOperations rejection
├── config_client_test.go              # Update MergeOperations test
├── errors.go                          # Re-export ErrorConflict, IsConflict
├── policy.go                          # PolicyInterface definition (NEW)
├── policy_client.go                   # PolicyInterface implementation (NEW)
├── policy_client_test.go              # Policy client tests (NEW)
├── sandbox.go                         # Add GetLogs to SandboxInterface
├── sandbox_client.go                  # Implement GetLogs (name→id resolve)
├── sandbox_client_test.go             # GetLogs tests
├── types/
│   ├── errors.go                      # Add ErrorConflict, IsConflict
│   ├── policy.go                      # Policy domain types (NEW)
│   ├── log.go                         # Log domain types (NEW)
│   ├── network_policy.go              # NetworkPolicyRule hierarchy (NEW)
│   └── setting.go                     # Change MergeOperations type
├── internal/converter/
│   ├── errors.go                      # Map codes.Aborted → ErrorConflict
│   ├── errors_test.go                 # Test new mapping
│   ├── policy.go                      # Policy converters (NEW)
│   ├── policy_test.go                 # Policy converter tests (NEW)
│   ├── network_policy.go              # NetworkPolicyRule converters (NEW)
│   ├── network_policy_test.go         # NetworkPolicyRule converter tests (NEW)
│   ├── log.go                         # Log converters (NEW)
│   ├── log_test.go                    # Log converter tests (NEW)
│   ├── setting.go                     # Add MergeOperation conversion
│   └── setting_test.go               # Update MergeOperation tests
└── fake/
    ├── fake.go                        # Add Policy() accessor + policy field
    ├── config.go                      # Remove MergeOperations rejection
    ├── config_test.go                 # Update MergeOperations test
    ├── policy.go                      # Fake PolicyClient (NEW)
    ├── policy_test.go                 # Fake policy tests (NEW)
    ├── sandbox.go                     # Add GetLogs stub
    └── sandbox_test.go               # GetLogs stub test
```

**Structure Decision**: Follows established pattern — one file per sub-client interface definition, one `_client.go` for implementation, one `_client_test.go` for tests, matching types and converters. NetworkPolicyRule gets its own type and converter files due to the deep type hierarchy.

## Implementation Strategy

### Layer Order (Bottom-Up)

1. **Error code** (ErrorConflict) — zero dependencies, enables all subsequent tests
2. **Domain types** (types/policy.go, types/log.go, types/network_policy.go, types/setting.go update) — types only, no logic
3. **Converters** (converter/policy.go, converter/network_policy.go, converter/log.go, converter/setting.go update) — maps proto↔SDK
4. **PolicyInterface + client** (policy.go, policy_client.go) — new sub-client
5. **GetLogs extension** (sandbox.go, sandbox_client.go updates) — extends existing interface
6. **MergeOperations update** (config_client.go update) — removes rejection, wires converter
7. **Fake stubs** (fake/policy.go, fake/sandbox.go update, fake/config.go update, fake/fake.go update)
8. **ClientInterface update** (client.go) — adds Policy() accessor, wires everything

### Testing Strategy

Each layer gets tests alongside implementation:
- Converter tests: round-trip proto→SDK→proto, nil handling, zero-value semantics
- Client tests: mock gRPC server, test each method signature, error mapping, name→id resolution (GetLogs)
- Fake tests: compile-time interface check, Unimplemented returns, MergeOperations acceptance
- Error tests: IsConflict helper, codes.Aborted mapping

## Complexity Tracking

No constitution violations to justify.
