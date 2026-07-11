# Brainstorm: SDK Dashboard TUI Example App

**Date:** 2026-07-03
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/25

## Problem Framing

The Go SDK has a rich API surface (sandboxes, exec, files, SSH/TCP,
services, providers, config, policy, health, OIDC login) but no
non-trivial example showing how these pieces compose in a real
application. The existing `example_test.go` snippets show individual
API calls in isolation. Potential adopters (and conference attendees)
need a compelling, runnable example that demonstrates the SDK's
breadth while also serving as a learning resource.

The example should also exercise the new OIDC login package being
developed in parallel (brainstorm 020), demonstrating the lazy auth
pattern (connect first, authenticate on 401) in a real application
context.

## Approaches Considered

### A: Sandbox Dashboard TUI (chosen)

A full-screen Bubble Tea dashboard called `oshell` in
`examples/oshell/`. Live-updating views for sandboxes, providers,
services, health, and exec. Tabbed interface with a log panel.

- Pros: most visually impressive, natural SDK showcase (each tab maps
  to a sub-client), Bubble Tea is the standard Go TUI library, great
  for demos and screencasts
- Cons: largest scope, Bubble Tea's Elm architecture may be unfamiliar
  to some readers

### B: Rich CLI with Lip Gloss styling

Cobra-based CLI with subcommands using Lip Gloss for styled output
and Huh for interactive prompts. No persistent TUI.

- Pros: simpler code, easier to read per-command, familiar CLI pattern
- Cons: less visually cohesive, no live updates without --watch flag,
  less impressive for demos

### C: Hybrid CLI + Dashboard mode

Cobra CLI subcommands plus a `oshell dashboard` subcommand that
launches the full TUI.

- Pros: CLI commands are easy to read individually, dashboard for demos
- Cons: largest codebase, two UX paradigms, unfocused as an "example"

## Decision

Option A: Sandbox Dashboard TUI. The dashboard is the most visually
striking for both SDK adopters and conference demos. The tabbed design
naturally partitions code into one file per sub-client, keeping each
piece readable despite the full-screen TUI. Bubble Tea is
well-established in the Go ecosystem (lazydocker, glow, soft-serve).

## Key Requirements

### Structure and Extractability

- Self-contained Go module in `examples/oshell/` with its own `go.mod`
- Imports the SDK as a regular dependency
  (`github.com/rhuss/openshell-sdk-go/openshell/v1`)
- No internal SDK package imports
- Clean directory boundary for future extraction to a standalone repo

### TUI Layout

- Top bar: gateway name, connection status, auth status (token expiry),
  health indicator
- Main area: tabbed views switchable with number keys
  - 1: Sandboxes (list with live status from Watch API, colored phase
    indicators, CRUD operations)
  - 2: Providers (profiles list, provider status)
  - 3: Services (exposed endpoints with URLs)
  - 4: Health (gateway health check, latency sparkline)
  - 5: Exec (select sandbox, run command, see output)
- Bottom: log panel (last N structured log lines, scrollable), key help

### SDK Coverage

- Sandboxes: Create, Get, List, Delete, Watch, WaitReady
- Exec: Run
- Services: List, Expose
- Providers: List profiles
- Health: Check
- Config: GetGateway
- Policy: read and display summary
- SSH/TCP: port forward status indicator
- OIDC login: lazy auth on 401 (from brainstorm 020)

### Authentication

- Lazy auth pattern: dashboard starts and connects immediately
- On 401 from gateway, prompts "Press Enter to login" inline
- Triggers OIDC browser flow, reconnects on success
- Shows token expiry countdown in the top bar
- Demonstrates RefreshableToken integration

### Structured Logging

- Custom `slog.Handler` that writes JSON to a log file AND feeds a ring
  buffer displayed in the TUI log panel
- Both SDK-level and app-level operations visible
- `--log-file` flag for offline JSON analysis
- `--gateway` flag for gateway selection

### Dependencies (Charm ecosystem)

- Bubble Tea: TUI framework (Elm architecture)
- Lip Gloss: styled rendering, colors, borders
- Bubbles: reusable components (table, spinner, viewport, textinput)

## Open Questions

- Should the example include a mock/demo mode that works without a
  real gateway (using the SDK's fake client)? Would help people run
  it immediately after cloning.
- What color theme? Match the OpenShell brand, use a popular terminal
  palette (Catppuccin, Nord), or auto-detect light/dark terminal?
- Should keyboard shortcuts follow vi conventions (j/k for navigation)
  or arrow keys only?
- How to handle the exec tab UX? Simple "type command, see output"
  textbox, or a scrollable history of past commands?
