# Files

Accessor: `client.Files()`

Upload and download files to and from sandboxes. Uses gRPC streaming for efficient
transfer of large files.

## Methods

| SDK Method | gRPC RPC | Proto File | Description |
|------------|----------|------------|-------------|
| `Upload(ctx, sandboxID, localPath, remotePath string)` | Client-streaming upload | `openshell.proto` | Upload a local file to a sandbox |
| `Download(ctx, sandboxID, remotePath, localPath string)` | Server-streaming download | `openshell.proto` | Download a file from a sandbox |

## Upload

Upload a file from the local filesystem to a path inside a sandbox.
The file is streamed in chunks using client-side gRPC streaming.

```go
err := client.Files().Upload(ctx, "sandbox-123", "./data/config.yaml", "/app/config.yaml")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File uploaded successfully")
```

## Download

Download a file from a sandbox to the local filesystem.
The file is received in chunks using server-side gRPC streaming.

```go
err := client.Files().Download(ctx, "sandbox-123", "/app/output.log", "./output.log")
if err != nil {
    log.Fatal(err)
}
fmt.Println("File downloaded successfully")
```

See also: [Error Handling](../error-handling.md), [Testing](../testing.md)
