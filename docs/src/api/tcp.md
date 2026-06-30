# TCP

Accessor: `client.TCP()`

Forward TCP connections to sandbox ports using bidirectional gRPC streaming.

## Forward

Open a bidirectional TCP forwarding stream to a specific port inside a sandbox.
Returns an `io.ReadWriteCloser` that proxies data between the caller and the
sandbox port over gRPC streaming.

```go
conn, err := client.TCP().Forward(ctx, "sandbox-123", 5432)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Example: proxy a PostgreSQL connection through the tunnel
// Write and read data as with any net.Conn-like stream
_, err = conn.Write([]byte("SELECT 1;\n"))
if err != nil {
    log.Fatal(err)
}

buf := make([]byte, 4096)
n, err := conn.Read(buf)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Response: %s\n", buf[:n])
```

TCP forwarding is lower-level than [SSH tunneling](ssh.md). Use TCP forwarding
when you need direct port access without SSH session overhead.

See also: [SSH Tunneling](ssh.md), [Error Handling](../error-handling.md)
