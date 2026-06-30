# Data Model: Sandbox Name Resolution

## Affected Interfaces

### ExecInterface (exec.go)

```
Before:
  Run(ctx, sandboxID string, command []string, opts ...ExecOptions) (*ExecResult, error)
  Stream(ctx, sandboxID string, command []string, opts ...ExecOptions) (ExecStream, error)
  Interactive(ctx, sandboxID string, command []string, cols, rows uint32, opts ...ExecOptions) (InteractiveSession, error)

After:
  Run(ctx, sandboxName string, command []string, opts ...ExecOptions) (*ExecResult, error)
  Stream(ctx, sandboxName string, command []string, opts ...ExecOptions) (ExecStream, error)
  Interactive(ctx, sandboxName string, command []string, cols, rows uint32, opts ...ExecOptions) (InteractiveSession, error)
```

### TCPInterface (tcp.go)

```
Before:
  Forward(ctx, sandboxID string, port uint32, opts ...ForwardOption) (io.ReadWriteCloser, error)

After:
  Forward(ctx, sandboxName string, port uint32, opts ...ForwardOption) (io.ReadWriteCloser, error)
```

### FileInterface (file.go)

```
Before:
  Upload(ctx, sandboxID string, localPath string, remotePath string) error
  Download(ctx, sandboxID string, remotePath string, localPath string) error

After:
  Upload(ctx, sandboxName string, localPath string, remotePath string) error
  Download(ctx, sandboxName string, remotePath string, localPath string) error
```

### ConfigInterface (config.go)

```
Before:
  GetSandbox(ctx, sandboxID string) (*SandboxConfig, error)

After:
  GetSandbox(ctx, sandboxName string) (*SandboxConfig, error)
```

### SandboxInterface (sandbox.go) - Watch method only

```
No signature change (already uses "name string").
Implementation change: resolve name to ID before WatchSandboxRequest.
```

### SSHInterface (ssh.go) - NO CHANGE

```
CreateSession stays ID-based:
  CreateSession(ctx, sandboxID string) (*SSHSession, error)
Doc comment updated to recommend Tunnel() for name-based access.
```

## Affected Structs

### Sub-client structs gain SandboxInterface field

```
execClient   { client pb.OpenShellClient; sandboxes SandboxInterface }
tcpClient    { client pb.OpenShellClient; sandboxes SandboxInterface }
fileClient   { client pb.OpenShellClient; sandboxes SandboxInterface }
configClient { client pb.OpenShellClient; sandboxes SandboxInterface }
```

### Constructor signature changes

```
newExecClient(conn, sandboxes)    // was: newExecClient(conn)
newTCPClient(conn, sandboxes)     // was: newTCPClient(conn)
newFileClient(conn, sandboxes)    // was: newFileClient(conn)
newConfigClient(conn, sandboxes)  // was: newConfigClient(conn)
```

## Entity: Sandbox

The Sandbox type has two identifiers:
- `Name` (string): human-readable, set at creation, unique per gateway
- `ID` (string): opaque server-assigned, used by low-level proto RPCs

All SDK public methods will use Name. ID is an internal detail for
proto request construction.
