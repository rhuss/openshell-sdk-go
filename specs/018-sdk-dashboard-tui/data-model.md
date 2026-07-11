# Data Model: SDK Dashboard TUI

## Core Types

### Dashboard (root model)

The top-level Bubble Tea model. Owns all shared state and composes tab models.

- **activeTab**: int (0-4, index of currently active tab)
- **focusedPanel**: enum (Main, Log) - which panel has keyboard focus
- **conn**: ConnectionState - gateway connection lifecycle
- **tabs**: [5]TabModel - one per tab view
- **logPanel**: LogPanel - bottom log panel
- **statusBar**: StatusBar - top status bar
- **width**, **height**: int - terminal dimensions
- **quitting**: bool - shutdown signal

### ConnectionState

Tracks gateway connection and authentication lifecycle.

- **status**: enum (Connecting, Connected, Disconnected, AuthRequired, Reconnecting)
- **gatewayName**: string - name of the connected gateway
- **endpoint**: string - gateway address
- **tokenExpiry**: time.Time - when the current token expires (zero if no token)
- **reconnectAttempt**: int (0-10) - current backoff attempt count
- **reconnectDelay**: time.Duration - current backoff delay
- **lastError**: error - last connection/auth error
- **client**: ClientInterface - the SDK client (real or fake)
- **authProvider**: AuthProvider - current auth (may be RefreshableToken)

### StatusBar

Rendered from ConnectionState plus health data.

- **gatewayName**: string
- **connStatus**: string ("Connected", "Disconnected", etc.)
- **authStatus**: string ("Authenticated", "Not Authenticated", token expiry)
- **healthDot**: lipgloss.Color (green, yellow, red)

### LogPanel

Bottom panel displaying structured log entries.

- **entries**: ring buffer of LogEntry (capacity: 200)
- **viewport**: bubbles.Viewport - scrollable view
- **focused**: bool - whether Tab key has focused this panel

### LogEntry

A single structured log record.

- **time**: time.Time
- **level**: slog.Level
- **message**: string
- **attrs**: []slog.Attr - key-value pairs

## Tab Models

Each tab is an independent Bubble Tea sub-model implementing a common
TabModel interface:

```
TabModel interface {
    Init() tea.Cmd
    Update(msg tea.Msg) (TabModel, tea.Cmd)
    View() string
    Title() string
    Refresh(client) tea.Cmd
}
```

### SandboxTab

- **table**: bubbles.Table - sandbox list
- **sandboxes**: []types.Sandbox - current data
- **watcher**: v1.Watcher - live Watch subscription
- **expandedRow**: int (-1 if none) - inline detail pane
- **expandedDetail**: SandboxDetail - policy/SSH/TCP info
- **creating**: bool - create dialog active
- **createForm**: CreateForm - name and image inputs

### SandboxDetail (inline expansion)

- **policy**: types.SandboxPolicy - network rules, filesystem settings
- **networkRuleCount**: int
- **sshStatus**: string - SSH session indicator
- **tcpForwards**: []string - active TCP forward ports

### ProviderTab

- **table**: bubbles.Table - provider profiles list
- **profiles**: []types.ProviderProfile

### ServiceTab

- **table**: bubbles.Table - service endpoints list
- **endpoints**: []types.ServiceEndpoint

### HealthTab

- **status**: types.HealthResult - latest health check
- **latencies**: [30]time.Duration - sparkline data (ring buffer)
- **latencyIndex**: int - write cursor
- **lastCheck**: time.Time
- **gatewayConfig**: types.GatewayConfig - settings summary

### ExecTab

- **sandboxSelector**: list of Ready sandbox names
- **selectedSandbox**: string - currently selected sandbox name
- **commandInput**: bubbles.TextInput - command entry
- **outputViewport**: bubbles.Viewport - scrollable output
- **history**: []ExecEntry - past command results
- **running**: bool - spinner active

### ExecEntry

- **command**: string - the command that was run
- **stdout**: string
- **stderr**: string
- **exitCode**: int
- **duration**: time.Duration

## State Transitions

### ConnectionState Flow

```
Connecting -> Connected       (gRPC dial succeeds)
Connecting -> AuthRequired    (401/Unauthenticated error)
Connecting -> Disconnected    (dial fails)
Connected -> Disconnected     (stream breaks, RPC fails)
Connected -> AuthRequired     (token refresh fails with 401)
AuthRequired -> Connecting    (user presses Enter, OIDC completes)
Disconnected -> Reconnecting  (auto-backoff triggered)
Reconnecting -> Connected     (retry succeeds)
Reconnecting -> Disconnected  (max attempts reached)
Disconnected -> Connecting    (user presses 'r' for manual retry)
```

### Sandbox Phase Colors

| Phase | Color | Lipgloss |
|-------|-------|----------|
| Pending | Yellow | lipgloss.Color("3") |
| Initializing | Yellow | lipgloss.Color("3") |
| Ready | Green | lipgloss.Color("2") |
| Error | Red | lipgloss.Color("1") |
| Terminating | Gray | lipgloss.Color("8") |
| Unknown | Gray | lipgloss.Color("8") |
