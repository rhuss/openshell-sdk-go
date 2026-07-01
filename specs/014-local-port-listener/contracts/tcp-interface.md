# Contract: TCPInterface.Listen

**Date**: 2026-06-30
**Feature**: 014-local-port-listener

## Method Signature

```go
Listen(ctx context.Context, sandboxName string, remotePort uint32, localPort uint32, opts ...ListenOption) (net.Listener, error)
```

## Parameters

| Parameter | Type | Constraints | Description |
|-----------|------|-------------|-------------|
| ctx | context.Context | Must not be nil | Controls listener lifecycle; cancellation closes the listener |
| sandboxName | string | Must not be empty | Sandbox name, resolved to ID internally |
| remotePort | uint32 | 1-65535 | Port inside the sandbox to tunnel to |
| localPort | uint32 | 0-65535 | Local port to bind; 0 for OS-assigned |
| opts | ...ListenOption | Optional | Functional options for configuration |

## Return Values

| Value | Type | Description |
|-------|------|-------------|
| listener | net.Listener | Bound local listener; Accept returns tunneled connections |
| err | error | Non-nil on validation failure or bind failure |

## Error Codes

| Error Code | Condition |
|------------|-----------|
| InvalidArgument | sandboxName is empty |
| InvalidArgument | remotePort is 0 or > 65535 |
| InvalidArgument | localPort is > 65535 |
| InvalidArgument | WithSSHTunnel() used but SSH client is nil |
| Unimplemented | Returned by the fake client |
| Unavailable | Client is closed |
| (OS error) | Local port already in use or bind fails |

## Options

| Constructor | Effect |
|-------------|--------|
| WithBindAddress(addr string) | Override default bind address (default: "127.0.0.1") |
| WithSSHTunnel() | Use SSH tunnel transport instead of TCP forward |
| WithListenServiceID(id string) | Set service ID on each tunnel stream |

## Behavioral Contract

1. Accept() blocks until a connection arrives, establishes the tunnel, and returns a net.Conn. Tunnel setup failures are handled internally (Accept retries).
2. Close() closes the underlying listener, cancels all connections, and blocks until all bridge goroutines finish.
3. Addr() returns the bound local address (useful when localPort was 0).
4. Concurrent Accept() calls from multiple goroutines are safe.
5. Context cancellation triggers Close() behavior.

## Fake Behavior

The fake implementation validates sandboxName (non-empty) and remotePort (1-65535), then returns Unimplemented. localPort is not validated by the fake (OS binding is a runtime concern).
