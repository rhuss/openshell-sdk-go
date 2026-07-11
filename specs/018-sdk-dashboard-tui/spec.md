# Feature Specification: SDK Dashboard TUI Example App

**Feature Branch**: `018-sdk-dashboard-tui`
**Created**: 2026-07-03
**Status**: Draft
**Input**: Brainstorm 021 - SDK Dashboard TUI Example App

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Launch Dashboard and View Sandboxes (Priority: P1)

A developer clones the SDK repository and wants to see a live dashboard of
their OpenShell gateway. They run `go run ./examples/oshell --gateway my-gw`
and see a full-screen TUI showing their sandboxes with real-time status
updates. They press number keys to switch between tabs and see different
resource views.

**Why this priority**: The sandbox tab is the core of the dashboard. Without
it, there is no meaningful example to show. It exercises the most critical
SDK operations (List, Watch, WaitReady) and demonstrates the Bubble Tea
integration pattern that all other tabs follow.

**Independent Test**: Run `oshell --gateway <name>`, verify the sandbox tab
renders a table of sandboxes with colored phase indicators that update in
real time when sandbox state changes.

**Acceptance Scenarios**:

1. **Given** a configured gateway with running sandboxes, **When** the user
   launches `oshell --gateway my-gw`, **Then** the dashboard displays a
   full-screen TUI with a top status bar, sandbox list in the main area,
   and a log panel at the bottom.
2. **Given** the sandbox tab is active, **When** a sandbox transitions from
   Pending to Ready, **Then** the phase indicator updates in real time with
   appropriate color coding (yellow for Pending, green for Ready, red for
   Error).
3. **Given** the sandbox tab is active, **When** the user presses `c`,
   **Then** a create dialog appears prompting for sandbox name and image.
4. **Given** a sandbox is selected, **When** the user presses `d`, **Then**
   the sandbox is deleted after a confirmation prompt.

---

### User Story 2 - Authenticate via OIDC on Demand (Priority: P1)

A developer launches the dashboard without cached OIDC tokens. The dashboard
connects to the gateway, receives a 401 rejection, and displays an inline
prompt asking the user to press Enter to authenticate. The browser opens for
OIDC login, and after successful authentication the dashboard reconnects and
loads normally.

**Why this priority**: Authentication is a prerequisite for all gateway
operations. The lazy auth pattern is the primary use case for the OIDC
login package and must work before any other tab can display data.

**Independent Test**: Clear cached tokens, launch `oshell`, verify the
auth prompt appears, complete the OIDC flow, and confirm the dashboard
populates with data after reconnection.

**Acceptance Scenarios**:

1. **Given** no cached OIDC tokens exist, **When** the user launches the
   dashboard, **Then** the top bar shows "Not Authenticated" and the main
   area displays "Press Enter to login via browser".
2. **Given** the auth prompt is showing, **When** the user presses Enter,
   **Then** the system browser opens to the OIDC provider's login page.
3. **Given** the user completes OIDC login in the browser, **When** the
   callback is received, **Then** the dashboard reconnects, the top bar
   updates to show the authenticated user and token expiry countdown, and
   sandbox data loads.
4. **Given** a valid cached token exists, **When** the user launches the
   dashboard, **Then** authentication succeeds silently and data loads
   immediately.
5. **Given** a token expires mid-session, **When** the RefreshableToken
   detects expiry, **Then** the token is refreshed automatically and the
   dashboard continues without interruption. If refresh fails, the auth
   prompt reappears.

---

### User Story 3 - Execute Commands in a Sandbox (Priority: P2)

A developer switches to the Exec tab, selects a running sandbox from a
dropdown, types a shell command, and sees the output rendered in the TUI.
They can run multiple commands and scroll through the output history.

**Why this priority**: Command execution is the most common SDK operation
after sandbox management. It demonstrates the Exec.Run API and shows how
to handle streaming output in a TUI context.

**Independent Test**: Switch to the Exec tab, select a sandbox, run
`echo hello`, verify the output "hello" appears in the output viewport.

**Acceptance Scenarios**:

1. **Given** the Exec tab is active and sandboxes exist, **When** the tab
   loads, **Then** a sandbox selector shows all sandboxes in Ready phase.
2. **Given** a sandbox is selected, **When** the user types a command and
   presses Enter, **Then** a spinner shows while the command runs, followed
   by stdout/stderr output rendered in a scrollable viewport.
3. **Given** multiple commands have been executed, **When** the user scrolls
   the output viewport, **Then** all previous command outputs are visible
   with clear separation between commands.
4. **Given** a command fails with a non-zero exit code, **When** the output
   renders, **Then** the exit code is displayed in red alongside any stderr
   output.

---

### User Story 4 - View Provider Profiles and Services (Priority: P2)

A developer switches to the Providers tab to see available provider profiles
(name, category, description). They switch to the Services tab to see
exposed endpoints with their public URLs.

**Why this priority**: These tabs demonstrate read-only SDK operations with
tabular data display, covering the Providers and Services sub-clients.

**Independent Test**: Switch to each tab, verify data loads and renders
in a table with correct column headers and values.

**Acceptance Scenarios**:

1. **Given** the Providers tab is active, **When** data loads, **Then** a
   table shows provider profiles with columns: Name, Category, Description.
2. **Given** the Services tab is active, **When** data loads, **Then** a
   table shows exposed endpoints with columns: Sandbox, Service Name,
   Port, URL.
3. **Given** the Services tab is active and a service has a URL, **When**
   the user selects the service and presses Enter, **Then** the URL is
   copied to the clipboard (if supported by the terminal).

---

### User Story 5 - Monitor Health and Gateway Status (Priority: P2)

A developer switches to the Health tab to see the gateway health status
and a latency sparkline showing recent response times. The top bar always
shows a health indicator dot (green/yellow/red) regardless of which tab
is active.

**Why this priority**: Health monitoring demonstrates the Health sub-client
and provides a persistent visual indicator of gateway connectivity, which
is valuable for demos and real usage.

**Independent Test**: Switch to the Health tab, verify the health status
and latency sparkline render correctly, verify the top bar health dot
updates based on gateway state.

**Acceptance Scenarios**:

1. **Given** any tab is active, **When** the gateway is healthy, **Then**
   the top bar shows a green dot next to the gateway name.
2. **Given** the Health tab is active, **When** data loads, **Then** the
   view shows: gateway health status (healthy/degraded/unavailable), last
   check time, and a sparkline of the last 30 latency measurements.
3. **Given** the gateway becomes unreachable, **When** the health check
   fails, **Then** the top bar dot turns red and the Health tab shows
   the error details.

---

### User Story 6 - View Structured Logs in the Dashboard (Priority: P3)

A developer sees structured log lines scrolling in the bottom panel as
they interact with the dashboard. Both SDK-level operations (gRPC calls,
token refresh) and app-level operations (tab switches, data fetches)
appear. They can also specify `--log-file app.log` to write JSON logs
to a file for offline analysis.

**Why this priority**: Logging demonstrates the slog integration pattern
and provides observability into the SDK's behavior, but it is supplementary
to the core functionality.

**Independent Test**: Launch with `--log-file /tmp/test.log`, perform some
operations, verify the log panel shows entries and the file contains valid
JSON log lines.

**Acceptance Scenarios**:

1. **Given** the dashboard is running, **When** an SDK operation occurs
   (e.g., listing sandboxes), **Then** a log line appears in the bottom
   panel with timestamp, level, message, and relevant attributes.
2. **Given** the `--log-file` flag is set, **When** operations occur,
   **Then** JSON-formatted log lines are written to the specified file.
3. **Given** the log panel has more entries than visible lines, **When**
   the user focuses the log panel and scrolls, **Then** older log entries
   become visible.
4. **Given** a token refresh occurs in the background, **When** the
   refresh completes, **Then** a log entry appears showing the refresh
   result without interrupting the active tab's display.

---

### User Story 7 - Run in Demo Mode Without a Gateway (Priority: P3)

A developer clones the SDK repo and wants to try the dashboard immediately
without access to a real OpenShell gateway. They run `oshell --demo` and
the dashboard launches with simulated data from the SDK's fake client,
showing realistic sandbox lifecycles, command outputs, and status changes.

**Why this priority**: Demo mode lowers the barrier to trying the example
and makes conference demos reliable (no network dependency). However, it
is not required for the core showcase functionality.

**Independent Test**: Run `oshell --demo`, verify the dashboard renders
with simulated data, sandboxes transition through phases, and exec
commands return mock output.

**Acceptance Scenarios**:

1. **Given** the `--demo` flag is set, **When** the dashboard launches,
   **Then** it connects to the SDK's fake client instead of a real gateway.
2. **Given** demo mode is active, **When** the sandbox tab loads, **Then**
   3-5 simulated sandboxes appear with various phases, and some transition
   through states over time.
3. **Given** demo mode is active, **When** the user runs a command in the
   Exec tab, **Then** a simulated response is returned after a brief delay.

---

### Edge Cases

- What happens when the gateway connection drops mid-session? The dashboard
  shows a "Disconnected" indicator in the top bar and attempts automatic
  reconnection with exponential backoff (1s initial, doubling, 30s cap,
  max 10 attempts). After 10 failed attempts, displays "Connection failed"
  with a manual retry option (press 'r').
- What happens when a Watch stream breaks? The watcher should be
  re-established transparently, with a log entry noting the reconnection.
- What happens when the terminal is resized? The Bubble Tea framework
  handles resize events; the layout should reflow all panels proportionally.
- What happens when no sandboxes exist? The sandbox tab shows an empty
  state message: "No sandboxes found. Press 'c' to create one."
- What happens when the log file cannot be written? The dashboard should
  log a warning to stderr at startup and continue without file logging.

## Clarifications

### Session 2026-07-03

- Q: How does the user switch focus between the main area and the log panel for scrolling? → A: Tab key cycles focus between the main area and the log panel. A visual indicator (border highlight or color change) shows which panel has focus.
- Q: What is the sandbox detail view referenced by FR-019 and FR-020? → A: Pressing Enter on a selected sandbox in the list expands an inline detail pane below the table row, showing policy summary (network rules count, filesystem settings) and SSH/TCP port forward status indicators. Press Escape to collapse.
- Q: What is the health check polling interval? → A: Every 10 seconds. The latency sparkline stores the last 30 measurements (5 minutes of history).
- Q: What are the reconnection backoff parameters when the gateway connection drops? → A: Exponential backoff starting at 1 second, doubling each attempt, capped at 30 seconds, maximum 10 attempts before displaying a persistent "Connection failed" message with a manual retry option (press 'r').
- Q: How large is the log ring buffer? → A: 200 entries. Older entries are discarded when the buffer is full. The file logger (--log-file) has no limit.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The example MUST be a self-contained Go module in
  `examples/oshell/` with its own `go.mod` that imports the SDK as a
  regular dependency (`github.com/rhuss/openshell-sdk-go/openshell/v1`).
- **FR-002**: The example MUST NOT import any internal SDK packages. All
  interactions MUST go through the public API surface.
- **FR-003**: The dashboard MUST render a full-screen TUI with three zones:
  a top status bar, a tabbed main area, and a bottom log panel.
- **FR-004**: The top status bar MUST display: gateway name, connection
  status (connected/disconnected/connecting), authentication status with
  token expiry countdown, and a health indicator dot.
- **FR-005**: The main area MUST support 4 tabbed views switchable via
  number keys (1-4): Sandboxes, Providers, Services, Gateway. Command
  execution (Exec) is integrated into the Sandboxes tab as a contextual
  popup rather than a separate tab, keeping sandbox-scoped actions
  co-located with the sandbox they operate on.
- **FR-006**: The Sandboxes tab MUST display a live-updating table using
  the Watch API with colored phase indicators (green=Ready, yellow=Pending,
  red=Error, gray=Unknown).
- **FR-007**: The Sandboxes tab MUST support creating a new sandbox (name
  and image input) and deleting a selected sandbox (with confirmation).
- **FR-008**: The Providers tab MUST display a table of provider profiles
  (display name, category, description) using the Providers sub-client.
- **FR-009**: The Services tab MUST display a table of exposed service
  endpoints (sandbox name, service name, port, URL) using the Services
  sub-client.
- **FR-010**: The Health tab MUST display the gateway health status and
  a sparkline of the last 30 latency measurements (polled every 10 seconds,
  covering 5 minutes of history) using the Health sub-client.
- **FR-011**: The Exec tab MUST allow selecting a Ready sandbox, entering
  a command, and displaying stdout/stderr output with exit code using the
  Exec sub-client.
- **FR-012**: The bottom log panel MUST display structured log entries from
  a ring buffer of 200 entries, scrollable when focused. The Tab key MUST
  cycle focus between the main area and the log panel, with a visual
  indicator (border highlight) showing which panel has focus.
- **FR-013**: The dashboard MUST implement lazy OIDC authentication:
  connect first, detect 401 responses, prompt the user to authenticate
  via browser, and reconnect after successful login.
- **FR-014**: The dashboard MUST support automatic token refresh via the
  SDK's RefreshableToken, displaying the token expiry countdown in the
  top bar.
- **FR-015**: The dashboard MUST provide a custom `slog.Handler` that
  writes JSON to a log file (when `--log-file` is specified) AND feeds
  a ring buffer for the TUI log panel.
- **FR-016**: The dashboard MUST accept the following CLI flags:
  `--gateway` (gateway name), `--log-file` (JSON log output path),
  `--demo` (use fake client with simulated data).
- **FR-017**: The dashboard MUST support both arrow keys and vi-style
  (j/k) navigation for list and table scrolling.
- **FR-018**: The dashboard MUST display gateway configuration summary
  (settings revision, key settings) using the Config sub-client,
  accessible from the Health tab or top bar.
- **FR-019**: The dashboard MUST display sandbox policy summaries
  (network rules count, filesystem settings) in an inline detail pane
  that expands below the selected sandbox row when the user presses
  Enter, and collapses on Escape. Uses the Policy sub-client.
- **FR-020**: The dashboard MUST show SSH/TCP port forward status as
  indicators in the sandbox inline detail pane.
- **FR-021**: The dashboard MUST handle graceful shutdown on Ctrl+C or
  'q', closing the SDK client and any active watchers.
- **FR-022**: Demo mode MUST use the SDK's fake client to provide
  simulated data without requiring a real gateway connection.

### Key Entities

- **Dashboard**: The top-level Bubble Tea model orchestrating all views,
  connection state, and authentication state.
- **Tab View**: An independent view component per resource type (Sandbox,
  Provider, Service, Health, Exec), each encapsulating its SDK sub-client
  interactions.
- **Log Entry**: A structured log record with timestamp, level, message,
  and key-value attributes, stored in a ring buffer and optionally written
  to a file.
- **Connection State**: Tracks the gateway connection lifecycle including
  authentication status, token expiry, health, and reconnection attempts.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A developer can clone the repo, run `go run ./examples/oshell
  --demo`, and see a working dashboard within 30 seconds (no external
  dependencies required for demo mode).
- **SC-002**: The sandbox tab updates within 2 seconds of a sandbox phase
  change on the gateway (when connected to a real gateway).
- **SC-003**: All 5 tabs render their data without errors when connected
  to a gateway with at least one sandbox, one provider profile, and one
  exposed service.
- **SC-004**: The dashboard exercises at least 10 distinct SDK sub-client
  methods across all tabs: Sandboxes.List, Sandboxes.Create,
  Sandboxes.Delete, Sandboxes.Watch, Exec.Run, Services.List,
  Providers.Profiles.List, Health.Check, Config.GetGateway, Policy.List.
- **SC-005**: Structured log output written to `--log-file` contains valid
  JSON lines parseable by standard tools (jq, lnav).
- **SC-006**: The OIDC lazy auth flow completes end-to-end: launch without
  tokens, authenticate via browser, dashboard reconnects and shows data,
  all within 60 seconds of user interaction.
- **SC-007**: The example codebase is self-contained: removing the
  `examples/oshell/` directory leaves the SDK fully functional with no
  broken imports or references.
- **SC-008**: Each tab's source file is independently readable as an SDK
  usage example (a developer can understand the Providers tab code without
  reading the Sandbox tab code).

## Assumptions

- The OIDC login package (`openshell/v1/oidc/`) will be available as a
  public SDK package by the time this example is implemented. If not yet
  merged, the OIDC functionality will be stubbed with a TODO marker.
- The gateway convenience package (`openshell/v1/gateway/`) is available
  for gateway configuration loading and client construction.
- The SDK's fake client supports enough operations (List, Create, Delete,
  Watch for sandboxes; List for providers and services; Check for health)
  to power a meaningful demo mode.
- Terminal capabilities (256 colors, unicode box-drawing characters) are
  assumed. The dashboard does not need to support legacy terminals with
  limited color support.
- The example targets Go 1.23+ (matching the SDK's minimum version).
- Bubble Tea v2, Lip Gloss v2, and Bubbles are the TUI framework versions
  used. These are the current stable releases in the Charm ecosystem.
