# Feature Specification: Converter Code Deduplication

**Feature Branch**: `004-converter-dedup`
**Created**: 2026-06-28
**Status**: Draft
**Input**: Extract domain types to a separate package to break the circular import cycle and deduplicate proto-to-SDK converter logic.

## User Scenarios & Testing

### User Story 1 - SDK Maintainer Modifies Conversion Logic (Priority: P1)

A developer maintaining the SDK needs to change how a proto field maps to an SDK type (e.g., adding a new field to `Sandbox`). Today, they must update both the converter package and the corresponding `*_client.go` file with identical changes. After this refactoring, they update only the converter package, and all client code uses that single implementation.

**Why this priority**: Eliminating duplicated conversion logic is the core goal. If conversion functions exist in only one place, maintenance risk drops to zero for this class of bugs.

**Independent Test**: Modify a converter function in the converter package and verify that the corresponding client method reflects the change without any edits to `*_client.go` files.

**Acceptance Scenarios**:

1. **Given** a developer adds a new field mapping in the converter package, **When** the SDK is compiled, **Then** the client methods use the updated converter without any code changes in `*_client.go` files.
2. **Given** the converter package contains all proto-to-SDK conversion functions, **When** a developer searches for conversion logic, **Then** all conversion functions are found in the converter package and none exist as unexported functions in `*_client.go` files.

---

### User Story 2 - SDK Consumer Upgrades Without Breaking Changes (Priority: P1)

An existing consumer of the SDK imports types from `openshell/v1` (e.g., `v1.Sandbox`, `v1.Config`). After the refactoring, they can continue importing from `openshell/v1` without changing their import paths or code.

**Why this priority**: Breaking existing consumers would block adoption. Re-exporting types from `v1` ensures backward compatibility.

**Independent Test**: Compile an existing test or example that imports `v1.Sandbox` and `v1.Config` and verify it builds without modifications.

**Acceptance Scenarios**:

1. **Given** a consumer imports the v1 package and uses `v1.Sandbox`, **When** they upgrade to the refactored SDK, **Then** their code compiles without any import path changes.
2. **Given** a consumer creates a `v1.Config{}` struct literal, **When** they upgrade to the refactored SDK, **Then** the struct literal compiles and works identically.

---

### User Story 3 - Converter Tests Run Independently (Priority: P2)

The converter package has its own test suite that validates proto-to-SDK mapping logic. After refactoring, the converter tests import the new types package directly instead of importing `v1`, breaking the circular dependency.

**Why this priority**: Independent converter testing was a design goal of the original converter package. The refactoring must preserve this capability.

**Independent Test**: Run the converter test suite and verify all tests pass using types from the types package.

**Acceptance Scenarios**:

1. **Given** the converter package imports types from a dedicated types package (not `v1`), **When** the converter test suite is run, **Then** all existing converter tests pass.
2. **Given** the converter package does not import `v1`, **When** the dependency graph is analyzed, **Then** no circular import exists between `v1` and the converter package.

---

### Edge Cases

- What happens when a consumer uses a type assertion on a re-exported type alias (e.g., `switch v := x.(type) { case *v1.Sandbox: }`)? Type aliases are fully compatible, so this must continue to work.
- What happens when a consumer uses struct literal initialization with field names (e.g., `v1.Sandbox{Name: "x"}`)? Type aliases preserve this capability.
- What happens when the `go doc` tool is run on the `v1` package? Re-exported type aliases should appear with documentation linking to the types package.

## Requirements

### Functional Requirements

- **FR-001**: All domain types (structs, enums, interfaces, constants) MUST be defined in a dedicated types package separate from the client implementation package.
- **FR-002**: The `v1` package MUST re-export all moved types via type aliases so that existing consumers can continue importing from `v1` without code changes.
- **FR-003**: The converter package MUST import types from the new types package instead of importing `v1`, eliminating the circular dependency.
- **FR-004**: The `v1` package MUST import and use the converter package for all proto-to-SDK conversions, removing all duplicated conversion functions from `*_client.go` files.
- **FR-005**: All existing unit tests MUST continue to pass after the refactoring.
- **FR-006**: All existing integration tests MUST continue to pass after the refactoring.
- **FR-007**: The dependency graph MUST be acyclic: `v1` imports `types` and `converter`; `converter` imports `types`; `types` imports neither `v1` nor `converter`.

### Key Entities

- **Types Package**: A new package containing all domain types (Sandbox, Provider, ExecResult, Config, etc.) that both `v1` and the converter package need.
- **Converter Package**: The existing converter package, updated to import from the types package instead of `v1`.
- **Client Package (v1)**: The existing `v1` package, updated to re-export types and delegate all conversions to the converter package.

## Success Criteria

### Measurable Outcomes

- **SC-001**: Zero duplicated conversion functions exist across the codebase. Every proto-to-SDK conversion has exactly one implementation.
- **SC-002**: All existing tests (unit and integration) pass without modification to test logic (import path changes in test files are acceptable).
- **SC-003**: The dependency graph contains no circular imports, verified by successful compilation of all packages.
- **SC-004**: Existing consumers can upgrade by changing only their dependency version, with no source code changes required.
- **SC-005**: The converter package can be tested in isolation without importing the client package.

## Assumptions

- The SDK is pre-1.0 and has limited external consumers, so the primary backward-compatibility concern is within the SDK's own test suite and examples rather than a broad ecosystem.
- The types package will live under `openshell/v1/types/` since the types are scoped to the v1 API and follow the existing `openshell/v1/` namespace convention.
- Type aliases (`type Sandbox = types.Sandbox`) in `v1/` provide full backward compatibility including struct literals, type assertions, and interface satisfaction.
- The `internal/` scoping of the converter package is preserved since it is an implementation detail not intended for external consumption.
- Proto-generated types remain in `proto/` and are not affected by this refactoring.
