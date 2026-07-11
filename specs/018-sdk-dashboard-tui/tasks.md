# Tasks: SDK Dashboard TUI (oshell)

**Feature**: SDK Dashboard TUI Example App
**Spec**: specs/018-sdk-dashboard-tui/spec.md
**Plan**: specs/018-sdk-dashboard-tui/plan.md

## Phase 1: Setup

- [x] T001 Create Go module scaffold with go.mod, go.sum, and main.go in examples/oshell/
- [x] T002 Add Bubble Tea v2, Lip Gloss v2, and Bubbles dependencies to examples/oshell/go.mod with replace directive for local SDK
- [x] T003 Create shared styles and color constants in examples/oshell/styles.go
- [x] T004 Create key bindings model with help text in examples/oshell/keys.go. Must include vi-style j/k bindings for list/table scrolling alongside arrow keys (FR-017). Define keyMap struct with Up/Down (arrow + j/k), Tab (focus cycle), number keys 1-5 (tab switch), q/Ctrl+C (quit), c (create), d (delete), Enter (expand/submit), Escape (collapse/cancel), r (manual retry)

## Phase 2: Foundational (blocking prerequisites)

- [x] T005 Implement root Dashboard Bubble Tea model (Init/Update/View) with tab switching (1-5 keys), panel focus cycling (Tab key), quit handling (q/Ctrl+C), and terminal resize handling (WindowSizeMsg reflows all panels proportionally) in examples/oshell/model.go
- [x] T006 Implement StatusBar rendering (gateway name, connection status, auth status, health dot) in examples/oshell/statusbar.go
- [x] T007 Implement tee slog.Handler with ring buffer (200 entries) and optional JSON file writer in examples/oshell/logpanel.go. Handle log file write failure gracefully: if the file cannot be opened or written, log a warning to stderr at startup and continue without file logging
- [x] T008 Implement LogPanel Bubble Tea component with scrollable viewport fed by the ring buffer in examples/oshell/logpanel.go
- [x] T009 Implement ConnectionState machine (Connecting/Connected/Disconnected/AuthRequired/Reconnecting) with exponential backoff (1s initial, 2x, 30s cap, 10 max) in examples/oshell/connection.go
- [x] T010 Implement client construction in main.go: flag parsing (--gateway, --log-file, --demo), gateway.NewClient or fake.NewClient dispatch, logger wiring

## Phase 3: User Story 1 - Sandbox Dashboard (P1)

**Goal**: Live-updating sandbox list with colored phase indicators and CRUD.
**Independent Test**: Launch oshell, verify sandbox table renders with phase colors, watch updates arrive in real time.

- [x] T011 [US1] Implement SandboxTab model with table rendering (Name, Phase, Image, Created columns) and colored phase indicators in examples/oshell/tab_sandbox.go. When no sandboxes exist, display empty state message: "No sandboxes found. Press 'c' to create one."
- [x] T012 [US1] Add Watch API integration to SandboxTab: start watcher on tab init, deliver events as tea.Msg, update table rows on Added/Modified/Deleted events in examples/oshell/tab_sandbox.go. Handle Watch stream breaks by re-establishing the watcher transparently with a log entry noting the reconnection
- [x] T013 [US1] Implement sandbox create dialog (name + image textinputs, Enter to submit, Escape to cancel) calling client.Sandboxes().Create() in examples/oshell/tab_sandbox.go
- [x] T014 [US1] Implement sandbox delete with confirmation prompt calling client.Sandboxes().Delete() in examples/oshell/tab_sandbox.go
- [x] T015 [US1] Implement inline detail pane (Enter to expand, Escape to collapse) showing policy summary via client.Policy().List() and SSH/TCP status indicators in examples/oshell/tab_sandbox.go

## Phase 4: User Story 2 - OIDC Authentication (P1)

**Goal**: Lazy auth flow: connect, detect 401, prompt login, OIDC browser flow, reconnect.
**Independent Test**: Clear cached tokens, launch oshell, verify auth prompt appears, complete OIDC flow, dashboard populates.

- [x] T016 [US2] Add 401/Unauthenticated error detection to ConnectionState: classify gRPC Unauthenticated errors as AuthRequired state in examples/oshell/connection.go
- [x] T017 [US2] Implement auth prompt UI: render "Press Enter to login via browser" when AuthRequired, trigger OIDC login on Enter in examples/oshell/model.go
- [x] T018 [US2] Integrate OIDC login package (or stub with TODO): call oidc.Login(gatewayName) in a tea.Cmd, handle success/failure, reconnect client in examples/oshell/connection.go
- [x] T019 [US2] Add token expiry countdown to StatusBar: read expiry from RefreshableToken, format as "Token: Xm Ys" in examples/oshell/statusbar.go
- [x] T020 [US2] Handle mid-session token refresh failure: detect 401 during active session, transition to AuthRequired, show re-auth prompt in examples/oshell/connection.go

## Phase 5: User Story 3 - Command Execution (P2)

**Goal**: Select sandbox, type command, see stdout/stderr output with history.
**Independent Test**: Switch to Exec tab, select sandbox, run `echo hello`, verify output appears.

- [x] T021 [P] [US3] Implement ExecTab model with sandbox selector (Ready sandboxes only), command textinput, and output viewport in examples/oshell/tab_exec.go
- [x] T022 [US3] Implement command execution: call client.Exec().Run() in a tea.Cmd goroutine, show spinner during execution, render stdout/stderr with exit code in examples/oshell/tab_exec.go
- [x] T023 [US3] Add command history: store ExecEntry records (command, stdout, stderr, exitCode, duration), render in viewport with separators between entries in examples/oshell/tab_exec.go

## Phase 6: User Story 4 - Providers and Services (P2)

**Goal**: View provider profiles and exposed service endpoints in tabular format.
**Independent Test**: Switch to each tab, verify data renders in tables with correct columns.

- [x] T024 [P] [US4] Implement ProviderTab model with table (Name, Category, Description columns) calling client.Providers().Profiles().List() in examples/oshell/tab_provider.go
- [x] T025 [P] [US4] Implement ServiceTab model with table (Sandbox, Service, Port, URL columns) calling client.Services().List() in examples/oshell/tab_service.go
- [x] T026 [US4] Add URL copy-to-clipboard on Enter key in ServiceTab using OSC 52 escape sequence in examples/oshell/tab_service.go

## Phase 7: User Story 5 - Health Monitoring (P2)

**Goal**: Gateway health status with latency sparkline and persistent health dot.
**Independent Test**: Switch to Health tab, verify status and sparkline render, verify top bar health dot.

- [x] T027 [P] [US5] Implement HealthTab model with health status display, latency sparkline (30 measurements, Unicode blocks), and gateway config summary in examples/oshell/tab_health.go
- [x] T028 [US5] Add background health check ticker (10s interval) calling client.Health().Check() and feeding latency ring buffer, plus gateway config from client.Config().GetGateway() in examples/oshell/tab_health.go
- [x] T029 [US5] Wire health dot in StatusBar: green when healthy, yellow when degraded, red when check fails, update from health ticker results in examples/oshell/statusbar.go

## Phase 8: User Story 6 - Structured Logging (P3)

**Goal**: Visible log panel with SDK and app operations, JSON file output.
**Independent Test**: Launch with --log-file, perform operations, verify log panel shows entries and file has valid JSON lines.

- [x] T030 [P] [US6] Wire slog logger into SDK client via gateway.WithLogger() and into all tab operations via context-scoped logging in examples/oshell/main.go
- [x] T031 [US6] Add log-level visual styling in LogPanel: color-code DEBUG (gray), INFO (white), WARN (yellow), ERROR (red) entries in examples/oshell/logpanel.go

## Phase 9: User Story 7 - Demo Mode (P3)

**Goal**: Launch with --demo, see simulated data from fake client.
**Independent Test**: Run `oshell --demo`, verify dashboard renders with fake sandboxes that transition through phases.

- [x] T032 [P] [US7] Implement demo mode setup: create fake.NewClient() with 5 pre-populated sandboxes (mixed phases), provider profiles, service endpoints, and healthy health result in examples/oshell/demo.go
- [x] T033 [US7] Add phase transition simulator: background goroutine that transitions fake sandboxes through Pending -> Initializing -> Ready (one every 5 seconds) in examples/oshell/demo.go
- [x] T034 [US7] Add canned exec responses in demo mode: return "Hello from sandbox <name>!" with 500ms simulated delay in examples/oshell/demo.go

## Phase 10: Polish & Cross-Cutting

- [x] T035 Add SPDX Apache-2.0 license headers to all .go files in examples/oshell/
- [x] T036 Add README.md section "## Example: Dashboard TUI" with usage instructions, flag reference, and screenshot placeholder
- [x] T037 Verify edge case handling implemented in domain tasks: empty sandbox list (T011), terminal resize reflow (T005), log file write failure (T007), Watch stream reconnection (T012). Smoke-test each scenario and fix any gaps
- [x] T038 Implement graceful shutdown: context cancellation chain closing watcher, then client, then Bubble Tea program in examples/oshell/main.go

## Dependencies

```
T001 -> T002 -> T003, T004 (setup chain)
T003, T004 -> T005 (root model needs styles and keys)
T005 -> T006, T007 (status bar and logging after root model)
T007 -> T008 (log panel needs tee handler)
T005 -> T009 (connection state after root model)
T009, T010 -> T011 (sandbox tab needs connection and client)
T011 -> T012, T013, T014, T015 (sandbox features need base tab)
T009 -> T016 -> T017 -> T018 (auth chain)
T018 -> T019, T020 (token display and refresh after auth)
T011 -> T021 (exec tab needs sandbox list pattern)
T021 -> T022 -> T023 (exec chain)
T011 -> T024, T025 (provider/service tabs parallel, follow sandbox pattern)
T025 -> T026 (clipboard needs service tab)
T005 -> T027 (health tab after root model)
T027 -> T028 -> T029 (health chain)
T007 -> T030 -> T031 (logging chain)
T010 -> T032 -> T033, T034 (demo chain)
All -> T035, T036, T037, T038 (polish after all features)
```

## Parallel Execution Opportunities

- **T003, T004**: Styles and keys are independent
- **T006, T007, T009**: Status bar, logging, and connection are independent foundations
- **T021, T024, T025, T027**: Exec, Provider, Service, and Health tabs can be built in parallel once the sandbox tab pattern is established
- **T030, T032**: Logging wiring and demo setup are independent
- **T035, T036, T037**: Polish tasks are independent

## Implementation Strategy

**MVP (Phase 1-3)**: Module scaffold + root model + sandbox tab. This alone
demonstrates the core SDK pattern (Watch API, CRUD) with a working TUI.
Roughly 15 tasks.

**Incremental delivery**: Each user story phase adds independently testable
functionality. Phases 5-7 (Exec, Providers/Services, Health) can be built
in any order.

**Total tasks**: 38
**Parallel tasks**: 12 (marked [P])
