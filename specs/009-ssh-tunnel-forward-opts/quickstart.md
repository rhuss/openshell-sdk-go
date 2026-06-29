# Quickstart: SSH Tunneling and TCP Forward Options

## SSH Tunnel (one call)

```go
client, _ := v1.NewClient(v1.Config{Address: "gateway:443"})
defer client.Close()

// Open an SSH tunnel to port 22 inside sandbox "my-sandbox"
tunnel, err := client.SSH().Tunnel(ctx, "my-sandbox", 22)
if err != nil {
    log.Fatal(err)
}
defer tunnel.Close() // auto-revokes the SSH session

// Use tunnel as io.ReadWriteCloser
_, _ = tunnel.Write([]byte("hello"))
buf := make([]byte, 1024)
n, _ := tunnel.Read(buf)
fmt.Println(string(buf[:n]))
```

## SSH Tunnel with Service ID

```go
tunnel, err := client.SSH().Tunnel(ctx, "my-sandbox", 22,
    v1.WithTunnelServiceID("my-app-v2"))
```

## TCP Forward with Service ID

```go
conn, err := client.TCP().Forward(ctx, sandboxID, 8080,
    v1.WithForwardServiceID("debug-session-42"))
```

## Notes

- `Tunnel()` takes a sandbox **name** (resolved internally to ID)
- `Forward()` takes a sandbox **ID** (unchanged from current API)
- Closing the tunnel revokes the SSH session automatically
- Context cancellation also triggers cleanup
