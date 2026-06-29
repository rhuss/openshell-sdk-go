# Data Model: SSH Tunneling and TCP Forward Options

**Feature**: 009-ssh-tunnel-forward-opts
**Date**: 2026-06-29

## Entities

### TunnelOption (new)

A functional option type for configuring SSH tunnel behavior.

| Field | Type | Description |
|-------|------|-------------|
| serviceID | string | Optional audit/correlation identifier |

**Pattern**: `type TunnelOption func(*tunnelConfig)`

**Location**: `openshell/v1/ssh.go` (public type) with unexported
`tunnelConfig` struct.

### ForwardOption (new)

A functional option type for configuring TCP forward behavior.

| Field | Type | Description |
|-------|------|-------------|
| serviceID | string | Optional audit/correlation identifier |

**Pattern**: `type ForwardOption func(*forwardConfig)`

**Location**: `openshell/v1/tcp.go` (public type) with unexported
`forwardConfig` struct.

### tunnelConfig (new, unexported)

Accumulates tunnel options.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| serviceID | string | "" | Service identifier for audit |

### forwardConfig (new, unexported)

Accumulates forward options.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| serviceID | string | "" | Service identifier for audit |

### sshTunnel (new, unexported)

Wraps a TCP forward connection with SSH session lifecycle management.
Implements `io.ReadWriteCloser`.

| Field | Type | Description |
|-------|------|-------------|
| conn | *tcpForwardConn | Underlying TCP forward connection |
| token | string | SSH session token for revocation |
| revokeFunc | func(ctx, token) (bool, error) | Session revocation callback |
| closeOnce | sync.Once | Ensures idempotent close |

**Lifecycle**:
- Created by `sshClient.Tunnel()` after successful session + forward
- `Read(p)` delegates to `conn.Read(p)`
- `Write(p)` delegates to `conn.Write(p)`
- `Close()` via `sync.Once`: close conn, then revoke session

## Relationships

```
SSHInterface.Tunnel()
    ├── calls SSHInterface.CreateSession() → SSHSession
    ├── calls TCPInterface.Forward() (internal, with SSH target)
    │       └── builds TcpForwardInit{SshRelayTarget, authorization_token, service_id}
    └── returns sshTunnel{conn, token, revokeFunc}

TCPInterface.Forward() (public)
    └── builds TcpForwardInit{TcpRelayTarget, service_id}
```

## Modified Entities

### SSHInterface (modified)

Added method:

```
Tunnel(ctx context.Context, sandboxName string, port uint32, opts ...TunnelOption) (io.ReadWriteCloser, error)
```

### TCPInterface (modified)

Signature change to accept options:

```
Forward(ctx context.Context, sandboxID string, port uint32, opts ...ForwardOption) (io.ReadWriteCloser, error)
```

### sshClient (modified)

Needs access to sandbox resolution and TCP forwarding:
- Add `sandboxes SandboxInterface` field (for name-to-ID resolution)
- Add `conn grpc.ClientConnInterface` field (for ForwardTcp stream)
- Constructor changes: `newSSHClient(conn, sandboxes)`

### tcpClient (modified)

- `Forward()` accepts variadic `ForwardOption`
- Builds `forwardConfig` from options
- Sets `ServiceId` on `TcpForwardInit` from config

### fakeSSHClient (modified)

- Add `Tunnel()` method returning Unimplemented
- Port validation before Unimplemented (Constitution XI)

### fakeTCPClient (modified)

- `Forward()` accepts variadic `ForwardOption`
- Options are accepted but ignored (returns Unimplemented as before)

## Proto Mapping

| Proto Field | Go SDK Field | Usage |
|-------------|-------------|-------|
| TcpForwardInit.service_id | forwardConfig.serviceID / tunnelConfig.serviceID | Set from WithServiceID option |
| TcpForwardInit.target.ssh | SshRelayTarget{} | Set by Tunnel() internally |
| TcpForwardInit.target.tcp | TcpRelayTarget{Host, Port} | Set by Forward() (existing) |
| TcpForwardInit.authorization_token | SSHSession.Token | Set by Tunnel() from CreateSession result |
