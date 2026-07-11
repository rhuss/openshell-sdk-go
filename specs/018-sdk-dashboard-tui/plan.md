# Implementation Plan: SDK Dashboard TUI

**Feature**: SDK Dashboard TUI Example App (oshell)
**Branch**: 018-sdk-dashboard-tui
**Spec**: specs/018-sdk-dashboard-tui/spec.md
**Research**: specs/018-sdk-dashboard-tui/research.md
**Data Model**: specs/018-sdk-dashboard-tui/data-model.md

## Technical Context

### Technology Stack

- **Language**: Go 1.23+ (matches SDK minimum)
- **TUI framework**: Bubble Tea v2 (Elm architecture)
- **Styling**: Lip Gloss v2
- **Components**: Bubbles (table, spinner, viewport, textinput, help)
- **CLI flags**: stdlib `flag` package
- **Logging**: stdlib `log/slog` with custom tee handler
- **SDK**: `github.com/rhuss/openshell-sdk-go/openshell/v1`
- **Gateway**: `github.com/rhuss/openshell-sdk-go/openshell/v1/gateway`
- **Fake client**: `github.com/rhuss/openshell-sdk-go/openshell/v1/fake`

### Project Location

```
examples/oshell/
  go.mod              # Self-contained module with replace directive
  go.sum
  main.go             # Entry point, flag parsing, client construction
  model.go            # Root Dashboard model (Init/Update/View)
  statusbar.go        # Top status bar rendering
  logpanel.go         # Bottom log panel + tee slog.Handler
  tab_sandbox.go      # Sandboxes tab (Watch, CRUD, detail pane)
  tab_provider.go     # Providers tab (profiles list)
  tab_service.go      # Services tab (endpoints list)
  tab_health.go       # Health tab (status, sparkline, gateway config)
  tab_exec.go         # Exec tab (sandbox selector, command input, output)
  connection.go       # ConnectionState management, lazy auth, reconnect
  demo.go             # Demo mode: fake client setup, simulated transitions
  styles.go           # Shared Lip Gloss styles and color constants
  keys.go             # Key bindings (help model)
```

### Constitution Check

| Principle | Applies | Status |
|-----------|---------|--------|
| I. Proto Isolation | Yes | OK: Example uses public SDK types only, never imports proto |
| II. Idiomatic Go | Yes | OK: stdlib flag, context propagation, error returns |
| III. Test-First | Partial | Example is a demo app, not library code. Unit tests for non-TUI logic (connection state, log handler) |
| IV. Upstream Tracking | No | Example does not modify proto or SDK |
| V. Minimal Dependencies | Yes | OK: Only Charm ecosystem + SDK. No additional deps |
| VI. Secrets Never Leak | Yes | OK: Tokens handled by SDK auth providers, never logged |
| VII. Deep Copy at Boundaries | No | Example consumes SDK types, does not cross proto boundary |
| VIII. Doc Examples Compile | Yes | OK: main.go is itself the runnable example |
| IX. Agent-Friendly Docs | Yes | OK: All exported types/functions get doc comments |
| X. Proto-SDK Naming Fidelity | No | Example uses SDK types, not proto |
| XI. Fake-Real Parity | Yes | OK: Demo mode uses the fake client unchanged |
| XII. Graceful Shutdown | Yes | OK: Close client, stop watchers before exit |
| XIII. Docs Accompany Features | Yes | OK: README section for the example |

### Gates

- **SPDX headers**: All .go files need headers (SDK convention)
- **No internal imports**: Only `openshell/v1`, `openshell/v1/gateway`,
  `openshell/v1/fake`, `openshell/v1/types`
- **Self-contained module**: Own go.mod, compilable with `go build`
- **Graceful shutdown**: Ctrl+C and 'q' both clean up properly

## Global Constraints

These apply to every task implicitly:

- **Go version**: 1.23+ (matches SDK minimum)
- **SPDX header**: Every `.go` file starts with Apache-2.0 SPDX header (NVIDIA copyright)
- **No internal imports**: Only `openshell/v1`, `openshell/v1/gateway`, `openshell/v1/fake`, `openshell/v1/types` (FR-002)
- **Self-contained module**: Own `go.mod` at `examples/oshell/`, compilable with `go build ./...` (FR-001)
- **Secrets never leak**: Tokens/credentials never appear in error messages, logs, or TUI output
- **Graceful shutdown order**: Close watchers, then SDK client, then Bubble Tea program
- **Terminal colors**: ANSI 256 colors; no true-color dependency
- **TUI framework versions**: Bubble Tea v2, Lip Gloss v2, Bubbles (current stable Charm releases)

## Implementation Phases

### Phase 1: Module Scaffold and Core Framework

Set up the Go module, root Bubble Tea model, and layout framework.
After this phase, the app launches and shows an empty tabbed layout
with a status bar and log panel.

**Files**: `go.mod`, `main.go`, `model.go`, `statusbar.go`, `logpanel.go`,
`styles.go`, `keys.go`

**Key decisions**:
- `go.mod` uses `replace` directive for local SDK reference
- Root model composes tab sub-models via the TabModel interface
- Tab key cycles focus between main and log panel
- Number keys (1-5) switch active tab
- 'q' and Ctrl+C trigger graceful shutdown
- Styles use ANSI 256 colors for broad terminal compatibility

### Phase 2: Connection Management and Logging

Implement the ConnectionState machine, gateway client construction,
the tee slog.Handler, and the lazy auth flow skeleton.

**Files**: `connection.go`, `logpanel.go` (tee handler addition)

**Key decisions**:
- `gateway.NewClient(name)` is the primary client constructor
- Connection errors are caught and classified (auth vs network)
- Exponential backoff: 1s initial, 2x multiplier, 30s cap, 10 max
- teeHandler wraps ringHandler + optional JSONHandler
- ringHandler is a []slog.Record with atomic write cursor (200 cap)
- The handler sends a Bubble Tea Cmd (tea.Msg) to push new entries
  into the TUI update loop, avoiding direct mutation

### Phase 3: Sandbox Tab (Core Tab)

Implement the sandboxes tab with live Watch updates, colored phase
indicators, create/delete operations, and inline detail expansion.

**Files**: `tab_sandbox.go`

**Key decisions**:
- Bubbles table component with custom column renderers for phase color
- Watch API via `client.Sandboxes().Watch(ctx, "")` for all sandboxes
- Watch events delivered as tea.Msg into the Update loop
- Create dialog: two textinputs (name, image) with Enter to submit
- Delete: confirm dialog before `client.Sandboxes().Delete(ctx, name)`
- Inline detail (Enter key): fetches policy via `client.Policy().List()`
  and renders below the selected row. Escape collapses.

### Phase 4: Remaining Tabs

Implement the Providers, Services, Health, and Exec tabs. Each tab
follows the same pattern established by the Sandbox tab.

**Files**: `tab_provider.go`, `tab_service.go`, `tab_health.go`,
`tab_exec.go`

**Key decisions**:
- Provider tab: `client.Providers().Profiles().List(ctx)` on tab activation
- Service tab: `client.Services().List(ctx, "")` listing all sandboxes'
  services. Enter on a row copies URL to clipboard via OSC 52 escape
- Health tab: Background ticker (10s) calling `client.Health().Check(ctx)`.
  Sparkline rendered with Unicode block characters. Gateway config
  summary from `client.Config().GetGateway(ctx)`
- Exec tab: Bubbles list for sandbox selection (Ready only), textinput
  for command, viewport for output history. `client.Exec().Run()` called
  in a tea.Cmd goroutine. Output appended to history with separator

### Phase 5: Demo Mode

Implement the fake client setup with pre-populated data and simulated
phase transitions.

**Files**: `demo.go`

**Key decisions**:
- `fake.NewClient()` with pre-populated sandbox store (3-5 sandboxes)
- Background goroutine transitions sandboxes through phases over time
  (Pending -> Initializing -> Ready, one every 5 seconds)
- Fake exec returns canned responses ("Hello from sandbox <name>!")
- Fake health always returns healthy with randomized latency (5-50ms)
- Fake providers/services pre-populated with example data
- No OIDC flow in demo mode (skip auth entirely)

### Phase 6: Polish and Documentation

Add README section, ensure SPDX headers, handle edge cases, test
the graceful shutdown path.

**Files**: `README.md` (root), all .go files (headers), edge case handling

**Key decisions**:
- README gets a new "## Example: Dashboard TUI" section with
  screenshot placeholder, usage instructions, and flag reference
- All .go files get SPDX Apache-2.0 headers
- Edge cases: empty sandbox list, terminal resize, log file write failure
- Graceful shutdown: context cancellation chain closes watcher, then
  client, then Bubble Tea program exits

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| OIDC package not ready | Medium | Low | Stub with TODO, core dashboard works without auth |
| Bubble Tea v2 API instability | Low | Medium | Pin exact version in go.mod |
| Fake client missing operations | Low | Low | Fake covers all sub-clients per codebase review |
| Terminal rendering differences | Medium | Low | Use ANSI 256 colors, test on common terminals |
| Module replace directive confusion | Low | Low | Document in README: "remove replace for standalone use" |
