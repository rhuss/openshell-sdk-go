# Brainstorm: Reverse Port Forwarding (ssh -R)

**Date:** 2026-07-01
**Status:** active
**Issue:** https://github.com/rhuss/openshell-sdk-go/issues/18

## Problem Framing

The SDK supports client-to-sandbox forwarding: `Forward()` opens a single
TCP connection to a sandbox port, `Listen()` binds a local port and
bridges accepted connections to the sandbox, and `Tunnel()` does the same
via SSH. All traffic flows in one direction: client to sandbox.

There is no reverse direction. A process inside the sandbox cannot reach
a port on the developer's machine. This is the `ssh -R` gap: the sandbox
should be able to listen on a port, and connections to that port should
tunnel back through the gateway to a target on the client side.

**Origin:** Deferred from brainstorm #010 (ssh-tunnel-forward-opts,
Phase 2b-3). Captured in idea-inbox as `reverse-forwarding`.

**Proto constraint:** The current `ForwardTcp` RPC is client-initiated
only. The client sends a `TcpForwardInit` with a sandbox ID and target,
and the gateway relays to the supervisor. There is no mechanism for the
gateway to notify the client of incoming connections from the sandbox.
This feature requires a proto extension upstream on NVIDIA/OpenShell.

**Scope:** This brainstorm is a pure design exercise. It defines the
ideal SDK API, sketches the proto extension needed, and documents
concrete use cases. Implementation is parked until proto support lands.

## Use Cases

### 1. MCP Tool Proxy

An AI agent runs inside the sandbox but needs MCP tools hosted on the
developer's machine (file system access, browser automation, database
queries). The developer runs an MCP server locally on port 8080. Reverse
forwarding lets sandbox code call `localhost:8080` (inside the sandbox),
which tunnels back to the developer's MCP server.

```go
// Developer's machine: expose local MCP server to sandbox
err := client.TCP().RemoteListen(ctx, "my-sandbox", 8080, "localhost:8080")
// Blocks until ctx cancelled. Sandbox code can now POST to localhost:8080
// inside the sandbox, and it reaches the developer's MCP server.
```

**Why this matters:** MCP tool servers often need access to local
resources (filesystem, credentials, running applications) that cannot
be replicated inside the sandbox. Without reverse forwarding, the
developer must expose their MCP server on a public URL or set up a
separate tunnel.

### 2. Local Model Server

The developer runs Ollama, vLLM, or llama.cpp locally (potentially
with a GPU that the sandbox lacks). Sandbox code calls the model's
inference API via the reverse tunnel.

```go
// Expose local Ollama to sandbox on the same port
err := client.TCP().RemoteListen(ctx, "my-sandbox", 11434, "localhost:11434")
// Sandbox code calls http://localhost:11434/api/generate
// and it reaches the developer's local Ollama instance.
```

**Why this matters:** GPU resources are expensive. A developer iterating
on prompt engineering or agent behavior may want to run the model locally
on their own hardware while the agent code runs in the sandbox. Deploying
the model inside every sandbox is wasteful for development workflows.

### 3. Inner-Loop Dev Testing

A developer iterates on a local API server. The sandbox runs integration
tests or client code against it via the reverse tunnel. No need to
rebuild, redeploy, or push the server into the sandbox on every code
change.

```go
// Expose local dev server while iterating
err := client.TCP().RemoteListen(ctx, "test-sandbox", 3000, "localhost:3000",
    WithRemoteListenServiceID("dev-api-server"),
)
// Sandbox integration tests hit localhost:3000 inside the sandbox,
// which bridges to the developer's local server. Change code locally,
// restart server, re-run sandbox tests. No redeploy needed.
```

**Why this matters:** The inner development loop (edit, test, debug) is
the most latency-sensitive workflow. Forcing a deploy-to-sandbox step
on every change breaks flow and adds minutes per iteration.

### 4. IDE Remote Debugging

A debugger running inside the sandbox (delve for Go, debugpy for Python)
connects back to the IDE's debug adapter on the developer's machine.

```go
// Expose IDE debug adapter port to sandbox
err := client.TCP().RemoteListen(ctx, "debug-sandbox", 2345, "localhost:2345")
// delve inside the sandbox connects to localhost:2345,
// which reaches the VS Code debug adapter on the developer's machine.
```

**Why this matters:** Remote debugging is essential for diagnosing
sandbox-specific behavior (network policies, filesystem restrictions,
resource limits). Without reverse forwarding, developers resort to
printf debugging or log tailing.

## Approaches Considered

### A: Minimal RemoteListen (Chosen)

A single auto-bridging method on `TCPInterface`:

```go
RemoteListen(ctx context.Context, sandboxName string, remotePort uint32,
    localTarget string, opts ...RemoteListenOption) error
```

Blocks until ctx is cancelled or an unrecoverable error occurs. Each
connection to sandbox:remotePort gets auto-bridged to localTarget on
the client side.

Options:
- `WithRemoteBindAddress(addr string)` - override the bind address inside
  the sandbox (default: `127.0.0.1`). Pass `0.0.0.0` to accept from any
  interface within the sandbox.
- `WithRemoteListenServiceID(id string)` - service identifier for audit
  and correlation.

Fake: returns `Unimplemented` error (same pattern as `Listen()` and
`Tunnel()`).

- **Pros:** Simplest API. One method call = persistent reverse tunnel.
  Covers all four motivating use cases. No new types. Follows YAGNI.
  Mirrors `ssh -R` semantics directly.
- **Cons:** No escape hatch for custom per-connection handling. If
  someone needs to inspect or route connections dynamically, a separate
  method would need to be added later.

### B: RemoteListen + RemoteForward

Two methods: a low-level `RemoteForward()` that returns a single
reverse-tunneled connection as `io.ReadWriteCloser`, plus the
high-level `RemoteListen()` wrapping it in a loop.

- **Pros:** Full flexibility. Mirrors Forward/Listen duality.
- **Cons:** `RemoteForward()` has awkward semantics: unlike `Forward()`
  which connects immediately, it blocks waiting for an incoming
  connection. The caller would need a loop, effectively reimplementing
  `RemoteListen()`. The raw-connection primitive may not map cleanly to
  either proto model. Adds API surface for an unproven use case.

### C: Event-Driven RemoteListener Type

A new `RemoteListener` struct with callback-based connection handling:
`OnConnection(func(net.Conn))`.

- **Pros:** Flexible, supports both bridging and custom handling.
- **Cons:** Breaks the sub-client interface pattern the SDK is built on.
  Returns a concrete type, not expressible in `TCPInterface`. Callback
  pattern is less idiomatic Go. Harder to fake.

## Decision

**Approach A: Minimal RemoteListen.** Single auto-bridging method on
`TCPInterface`. Covers all four use cases. Raw per-connection API can be
added as a non-breaking extension if a concrete need surfaces.

## SDK API Design

### Updated TCPInterface

```go
type TCPInterface interface {
    // (existing methods)
    Forward(ctx context.Context, sandboxName string, port uint32,
        opts ...ForwardOption) (io.ReadWriteCloser, error)
    Listen(ctx context.Context, sandboxName string, remotePort uint32,
        localPort uint32, opts ...ListenOption) (net.Listener, error)

    // RemoteListen sets up reverse port forwarding from a sandbox back to
    // the client. The sandbox listens on remotePort; each accepted
    // connection is tunneled through the gateway and bridged to
    // localTarget on the client side.
    //
    // localTarget is a host:port string (e.g., "localhost:3000",
    // "127.0.0.1:8080"). The SDK dials this address for each incoming
    // connection from the sandbox.
    //
    // RemoteListen blocks until ctx is cancelled or an unrecoverable error
    // occurs. Transient per-connection errors (failed dial to localTarget,
    // broken tunnel) are logged but do not stop the listener.
    //
    // Errors:
    //   - InvalidArgument: sandboxName is empty, remotePort is 0 or
    //     > 65535, or localTarget is malformed
    //   - Unimplemented: returned by the fake client
    //   - Unavailable: client is closed
    RemoteListen(ctx context.Context, sandboxName string, remotePort uint32,
        localTarget string, opts ...RemoteListenOption) error
}
```

### Options

```go
type remoteListenConfig struct {
    bindAddress string
    serviceID   string
}

type RemoteListenOption func(*remoteListenConfig)

func WithRemoteBindAddress(addr string) RemoteListenOption
func WithRemoteListenServiceID(id string) RemoteListenOption
```

### Error Handling

- `localTarget` dial failures: logged, connection dropped, listener
  continues. The sandbox side sees a connection reset.
- Gateway/tunnel errors: transient errors retry with backoff. Permanent
  errors (sandbox deleted, auth revoked) return from `RemoteListen()`.
- Context cancellation: all active bridges torn down, method returns
  `ctx.Err()`.

## Proto Extension Sketch

The current `ForwardTcp` RPC cannot support reverse forwarding. Two
proto-level approaches are viable:

### Model 1: Client-Polls (WaitForReverse)

The client opens a long-lived stream expressing interest in reverse
connections. The gateway notifies the client when the sandbox connects
to the remote port. The client then joins the relay.

```protobuf
// Client opens this stream to register for reverse forwarding.
// Gateway sends ReverseConnectionEvent for each incoming sandbox connection.
rpc WaitForReverse(ReverseListenInit)
    returns (stream ReverseConnectionEvent);

message ReverseListenInit {
    string sandbox_id = 1;
    uint32 port = 2;
    string bind_address = 3;  // inside sandbox, default 127.0.0.1
    string service_id = 4;
}

message ReverseConnectionEvent {
    string channel_id = 1;    // join this channel via RelayStream
    string source_addr = 2;   // who connected inside the sandbox
}
```

After receiving a `ReverseConnectionEvent`, the client opens a
`RelayStream` with the channel_id to join the relay, then bridges
to localTarget.

- **Pros:** Uses existing `RelayStream` for data. Server-streaming
  RPC is simple. Client controls when to join each connection.
- **Cons:** Two RPCs per connection (WaitForReverse + RelayStream).
  Slight latency on connection setup. Client must manage the
  WaitForReverse stream lifetime.

### Model 2: Gateway-Push (ReverseTcp Bidi Stream)

A single bidirectional stream where the gateway pushes connection
events and the client responds within the same stream. Data flows
over separate RelayStream calls.

```protobuf
// Bidi stream for reverse forwarding lifecycle.
rpc ReverseTcp(stream ReverseTcpClientMessage)
    returns (stream ReverseTcpServerMessage);

message ReverseTcpClientMessage {
    oneof payload {
        ReverseListenInit init = 1;
        ReverseAccept accept = 2;     // client accepts a connection
        ReverseReject reject = 3;     // client rejects (e.g., at capacity)
    }
}

message ReverseTcpServerMessage {
    oneof payload {
        ReverseListenReady ready = 1;         // sandbox is now listening
        ReverseConnectionEvent connection = 2; // new connection arrived
    }
}

message ReverseAccept {
    string channel_id = 1;
}

message ReverseReject {
    string channel_id = 1;
    string reason = 2;
}
```

- **Pros:** Single stream for the full lifecycle. Explicit accept/reject
  gives the client control. Ready confirmation ensures sandbox is
  actually listening before the SDK returns.
- **Cons:** More complex proto. Bidi stream is harder to implement on
  both gateway and client. Reject semantics add complexity.

### Recommendation

Model 1 (client-polls) is simpler and reuses existing patterns. The
`ReverseListenReady` confirmation from Model 2 is valuable though.
A hybrid approach (server-streaming with a ready confirmation message
before connection events) would combine the best of both.

## Key Requirements

- **Method**: `TCP().RemoteListen(ctx, sandboxName, remotePort,
  localTarget, ...RemoteListenOption) error`
- **Blocking**: Returns only on ctx cancellation or fatal error
- **Transport**: Proto-level reverse forwarding (requires upstream
  proto extension)
- **Bind address**: Defaults to `127.0.0.1` inside sandbox,
  configurable via `WithRemoteBindAddress()`
- **Error handling**: Transient per-connection errors logged, not fatal.
  Permanent errors (sandbox gone, auth revoked) return from method
- **Fake**: Returns `Unimplemented` error
- **Interface**: Added to `TCPInterface`
- **localTarget**: Standard `host:port` string, dialed per connection

## Open Questions

- Proto model: which approach does the upstream team prefer? File as
  a discussion or RFC on NVIDIA/OpenShell once this brainstorm is
  reviewed.
- Should `RemoteListen` accept a `WithOnError(func(error))` callback
  for surfacing per-connection errors to the caller? Or is logging
  sufficient?
- Sandbox-side bind semantics: does the supervisor need to allocate a
  port, or does the proto specify the exact port? What happens if the
  port is already in use inside the sandbox?
- Auth model: does reverse forwarding require a separate auth token,
  or does the client's existing credential suffice? Forward direction
  uses SSH session tokens for Tunnel(); reverse may need a different
  mechanism.
- Should there be a `WithSSHTunnel()` option like forward `Listen()`
  has, or is reverse forwarding always direct TCP?
- Connection limit: should `WithMaxConnections()` be part of v1, or
  deferred like it was for `Listen()`?
