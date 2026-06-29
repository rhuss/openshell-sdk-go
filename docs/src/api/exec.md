# Exec

Access the Exec API through `client.Exec()`, which returns an `ExecInterface`. It provides three modes for running commands in a sandbox:

- **Run** collects all output and returns it as a single result.
- **Stream** exposes output as an iterator for processing chunks as they arrive.
- **Interactive** opens a bidirectional session for terminal-style interaction.

## Run

Execute a command and collect all output into a single result.

```go
Run(ctx context.Context, sandboxID string, command []string, opts ...ExecOptions) (*ExecResult, error)
```

**SDK (Go):**

```go
result, err := client.Exec().Run(ctx, "sbx-123", []string{"ls", "-la"})
if err != nil {
    log.Fatal(err)
}

fmt.Println("Exit code:", result.ExitCode)
fmt.Println("Stdout:", result.Stdout)
fmt.Println("Stderr:", result.Stderr)
```

**gRPC (Proto):**

```protobuf
// Server-streaming RPC. The SDK collects all streamed ExecSandboxEvent
// messages, assembles stdout/stderr, and returns a single ExecResult.
rpc ExecSandbox(ExecSandboxRequest) returns (stream ExecSandboxEvent);
```

`ExecResult` contains the complete output after the command finishes:

| Field      | Type   | Description                  |
|------------|--------|------------------------------|
| `Stdout`   | string | Captured standard output     |
| `Stderr`   | string | Captured standard error      |
| `ExitCode` | int    | Process exit code            |

## Stream

Execute a command and process output chunks as they arrive.

```go
Stream(ctx context.Context, sandboxID string, command []string, opts ...ExecOptions) (ExecStream, error)
```

**SDK (Go):**

```go
stream, err := client.Exec().Stream(ctx, "sbx-123", []string{"tail", "-f", "/var/log/app.log"})
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

for {
    chunk, err := stream.Next()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }

    if chunk.Stream == v1.StreamStdout {
        fmt.Print(string(chunk.Data))
    }
}

exitCode, err := stream.ExitCode()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Exited with:", exitCode)
```

**gRPC (Proto):**

```protobuf
// Same server-streaming RPC as Run. The SDK exposes the raw stream
// of ExecSandboxEvent messages through the ExecStream iterator interface.
rpc ExecSandbox(ExecSandboxRequest) returns (stream ExecSandboxEvent);
```

## Interactive

Open a bidirectional terminal session with a command.

```go
Interactive(ctx context.Context, sandboxID string, command []string, cols, rows uint32, opts ...ExecOptions) (InteractiveSession, error)
```

**SDK (Go):**

```go
session, err := client.Exec().Interactive(ctx, "sbx-123", []string{"/bin/bash"}, 80, 24)
if err != nil {
    log.Fatal(err)
}
defer session.Close()

// Send input
_, err = session.Write([]byte("echo hello\n"))
if err != nil {
    log.Fatal(err)
}

// Read output
buf := make([]byte, 4096)
n, err := session.Read(buf)
if err != nil {
    log.Fatal(err)
}
fmt.Print(string(buf[:n]))

// Handle terminal resize
if err := session.Resize(120, 40); err != nil {
    log.Fatal(err)
}

// Get exit code after session ends
exitCode, err := session.ExitCode()
if err != nil {
    log.Fatal(err)
}
fmt.Println("Exited with:", exitCode)
```

**gRPC (Proto):**

```protobuf
// Bidirectional streaming RPC. The SDK wraps the two-way stream
// as an InteractiveSession with Read/Write/Resize methods.
rpc ExecSandboxInteractive(stream ExecSandboxInput)
    returns (stream ExecSandboxEvent);
```

## ExecStream

`ExecStream` provides an iterator interface over command output chunks. Call `Next()` repeatedly to receive output as it is produced. When the command finishes, `Next()` returns `io.EOF`.

```go
type ExecStream interface {
    Next() (*ExecChunk, error)
    ExitCode() (int, error)
    Close() error
}
```

| Method     | Description                                                    |
|------------|----------------------------------------------------------------|
| `Next`     | Returns the next output chunk. Returns `io.EOF` when done.    |
| `ExitCode` | Returns the process exit code. Call after `Next` returns `io.EOF`. |
| `Close`    | Releases the underlying stream resources.                      |

## InteractiveSession

`InteractiveSession` implements `io.Reader` and `io.Writer` for bidirectional communication with a running process. Use it for terminal emulation, REPL interaction, or any command that requires ongoing input.

```go
type InteractiveSession interface {
    Read(p []byte) (int, error)
    Write(p []byte) (int, error)
    Resize(cols, rows uint32) error
    ExitCode() (int, error)
    Close() error
}
```

| Method     | Description                                                         |
|------------|---------------------------------------------------------------------|
| `Read`     | Reads output from the process into the provided buffer.             |
| `Write`    | Sends input to the process.                                         |
| `Resize`   | Updates the terminal dimensions (columns and rows).                 |
| `ExitCode` | Returns the process exit code after the session ends.               |
| `Close`    | Closes the session and releases resources.                          |

## ExecChunk

Each chunk from `ExecStream.Next()` carries a segment of process output along with which stream it came from.

```go
type ExecChunk struct {
    Data   []byte
    Stream StreamType
}
```

| Field    | Type       | Description                                       |
|----------|------------|---------------------------------------------------|
| `Data`   | `[]byte`   | Raw output bytes from the process.                |
| `Stream` | StreamType | Either `StreamStdout` or `StreamStderr`.          |
