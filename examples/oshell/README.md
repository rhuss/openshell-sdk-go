# oshell - OpenShell SDK Dashboard TUI

A terminal-based dashboard for the OpenShell SDK, built with
[Bubble Tea v2](https://github.com/charmbracelet/bubbletea). It demonstrates
sandbox management, provider profiles, service endpoints, health monitoring,
and command execution through an interactive five-tab interface.

## Usage

```bash
# Demo mode (fake data, no gateway required)
go run ./examples/oshell --demo

# Connect to a real gateway
go run ./examples/oshell --gateway my-gateway

# With JSON log file output
go run ./examples/oshell --demo --log-file /tmp/oshell.log
```

## Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--gateway` | _(none)_ | Gateway name to connect to |
| `--demo` | `false` | Run with fake data (no gateway needed) |
| `--log-file` | _(none)_ | Path to JSON log file for structured logging |

If neither `--gateway` nor `--demo` is specified, the app defaults to demo mode.

## Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `1`-`5` | Switch tabs (Sandboxes, Providers, Services, Health, Exec) |
| `Tab` | Cycle focus between main panel and log panel |
| `j` / `k` | Scroll up/down in lists and tables |
| `Enter` | Expand details, submit input, or copy URL |
| `Escape` | Collapse details or cancel input |
| `c` | Create new sandbox |
| `d` | Delete selected sandbox |
| `r` | Manual refresh / retry |
| `?` | Toggle help |
| `q` / `Ctrl+C` | Quit |

## Tabs

1. **Sandboxes**: Live-updating sandbox list with colored phase indicators, create/delete, detail pane with policy and SSH/TCP status.
2. **Providers**: Provider profiles table showing name, category, and description.
3. **Services**: Exposed service endpoints with sandbox name, service name, port, and URL. Press Enter to copy URL via OSC 52.
4. **Health**: Gateway health status, latency sparkline (30 measurements), and gateway configuration summary.
5. **Exec**: Select a sandbox, type a command, view stdout/stderr output with command history.

## Demo Mode

When launched with `--demo`, the dashboard uses a fake client pre-populated with:

- 5 sandboxes in mixed phases (Ready, Provisioning, Unknown, Error)
- 3 providers (OpenAI, Anthropic, GitHub)
- 5 provider profiles across categories (Inference, Source Control, Messaging, Data)
- 4 service endpoints spread across sandboxes
- A phase transition simulator that advances sandboxes through Unknown, Provisioning, and Ready (one transition every 5 seconds)
- Canned exec responses ("Hello from sandbox!") with a 500ms simulated delay

## Screenshot

<!-- TODO: Add a screenshot or recording of the TUI -->
