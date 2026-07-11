# Research: SDK Dashboard TUI

## Bubble Tea Architecture

**Decision**: Use Bubble Tea v2 with the standard Model-Update-View (Elm
architecture) pattern. Each tab is an independent sub-model composed into
a root model.

**Rationale**: Bubble Tea is the de facto Go TUI framework (20k+ stars,
used by lazydocker, glow, soft-serve). The Elm architecture maps naturally
to our tabbed layout: each tab manages its own state and SDK interactions,
the root model handles tab switching and shared state (connection, auth).

**Alternatives considered**:
- tview (imperative, widget-based): more traditional but less composable,
  harder to test, no built-in Elm architecture
- tcell (low-level): too much boilerplate for this use case
- bubbletea v1: superseded, v2 is current stable

## Charm Ecosystem Libraries

**Decision**: Use Lip Gloss v2 for styling, Bubbles for reusable components
(table, spinner, viewport, textinput, help), Harmonica for sparkline
animations.

**Rationale**: These are the official companion libraries from Charm.
Bubbles provides pre-built components that match exactly what we need
(table for sandbox/provider/service lists, viewport for exec output and
logs, textinput for command entry, spinner for loading states).

## Module Structure

**Decision**: Self-contained Go module at `examples/oshell/` with its own
`go.mod`. Import SDK via `github.com/rhuss/openshell-sdk-go/openshell/v1`.

**Rationale**: Per extractability constraint. The module uses a `replace`
directive in `go.mod` to reference the local SDK during development:
`replace github.com/rhuss/openshell-sdk-go => ../..`

This allows `go run .` from the example directory without publishing
the SDK, while the `go.mod` remains valid for standalone use after
extraction (just remove the replace directive).

## OIDC Integration

**Decision**: Import the OIDC login package (`openshell/v1/oidc/`) when
available. Stub with a TODO if not yet merged. The lazy auth pattern
catches gRPC Unauthenticated errors and triggers `oidc.Login(gatewayName)`.

**Rationale**: The OIDC package (brainstorm 020) is being developed in
parallel. The example should demonstrate the integration pattern even if
the package isn't ready. The stub approach keeps the example compilable.

**Alternatives considered**:
- Build tag to conditionally compile OIDC: adds complexity, harder to read
- Skip OIDC entirely: misses a key showcase requirement

## Structured Logging Pattern

**Decision**: Implement a `teeHandler` that wraps two `slog.Handler`
instances: a `ringHandler` (feeds the TUI log panel via a ring buffer)
and an optional `slog.JSONHandler` (writes to the log file). The SDK's
`types.Logger` interface is bridged to `slog` via an adapter.

**Rationale**: Go 1.21+ includes `log/slog` in the stdlib. A tee handler
is the cleanest way to multiplex log output without custom infrastructure.
The ring buffer is a simple `[]slog.Record` with a write cursor.

## Demo Mode Architecture

**Decision**: Use the SDK's `fake.Client` directly. Pre-populate the fake
stores with 3-5 sandboxes in various phases. Use a background goroutine
to simulate phase transitions (Pending -> Initializing -> Ready) over time
via the fake's sandbox store.

**Rationale**: The fake client implements the full `ClientInterface`, so
all tabs work without code branching. The example demonstrates a real
testing pattern (fake client for integration tests).

## CLI Flag Parsing

**Decision**: Use stdlib `flag` package, not Cobra. Three flags: `--gateway`,
`--log-file`, `--demo`.

**Rationale**: This is a single-command binary, not a CLI with subcommands.
Cobra would be overkill and add a dependency. The stdlib `flag` package is
sufficient and keeps the example minimal.

**Alternatives considered**:
- Cobra: heavyweight for a single command, adds dependency
- pflag: marginal benefit over stdlib flag for 3 flags
