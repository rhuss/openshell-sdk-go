# Data Model: Reverse Port Forwarding

**Feature**: 025-reverse-port-forwarding
**Date**: 2026-08-05

## Entities

### RemoteListenOption

Functional option type for configuring `RemoteListen` behavior.

```
Type: func(*remoteListenConfig)
Package: openshell/v1
Exported: yes
```

### remoteListenConfig

Internal config struct that accumulates resolved option values.

```
Fields:
  - bindAddress string   // sandbox-side bind address, default "127.0.0.1"
  - serviceID   string   // optional service identifier for audit/correlation

Package: openshell/v1
Exported: no (lowercase)
```

### Option Constructors

| Function | Sets | Default |
|----------|------|---------|
| `WithRemoteBindAddress(addr string)` | `bindAddress` | `"127.0.0.1"` |
| `WithRemoteListenServiceID(id string)` | `serviceID` | `""` |

## Interface Changes

### TCPInterface (modified)

```
Added method:
  RemoteListen(ctx context.Context, workspace, sandboxName string,
    remotePort uint32, localTarget string,
    opts ...RemoteListenOption) error
```

### fakeTCPClient (modified)

```
Added method:
  RemoteListen(_ context.Context, _, sandboxName string,
    remotePort uint32, localTarget string,
    _ ...RemoteListenOption) error

Validation order:
  1. closedFunc() → Unavailable
  2. sandboxName == "" → InvalidArgument
  3. remotePort == 0 || > 65535 → InvalidArgument
  4. net.SplitHostPort(localTarget) fails → InvalidArgument
  5. → Unimplemented
```

### tcpClient (modified)

```
Added method:
  RemoteListen(ctx context.Context, workspace, sandboxName string,
    remotePort uint32, localTarget string,
    opts ...RemoteListenOption) error

Implementation: validates inputs, returns Unimplemented (stub)
```

## Relationships

```
TCPInterface ──implements──> tcpClient (real, stub)
TCPInterface ──implements──> fakeTCPClient (fake)
RemoteListenOption ──configures──> remoteListenConfig
Client.TCP() ──returns──> TCPInterface
```

## No New State Transitions

`RemoteListen` is a blocking call with no internal state machine. It returns when context is cancelled or a permanent error occurs. No lifecycle states to track.
