# Data Model: SSH Tunnel Context Cancellation Cleanup

No new types or entities. This feature modifies the behavior of the
existing `sshTunnel` struct by adding a cleanup goroutine. The struct
fields remain unchanged:

```
sshTunnel
├── *tcpForwardConn  (embedded, owns done channel)
├── revokeFunc       func()
├── closeOnce        sync.Once
└── closeErr         error
```

The only change is a new goroutine launched in `Tunnel()` that watches
`tcpForwardConn.done` and calls `sshTunnel.Close()`.
