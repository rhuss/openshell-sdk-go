# Contract: TCPInterface

## Current

```go
type TCPInterface interface {
    Forward(ctx context.Context, sandboxID string, port uint32) (io.ReadWriteCloser, error)
}
```

## Proposed

```go
type TCPInterface interface {
    Forward(ctx context.Context, sandboxID string, port uint32, opts ...ForwardOption) (io.ReadWriteCloser, error)
}
```

## New Types

```go
type ForwardOption func(*forwardConfig)

func WithForwardServiceID(id string) ForwardOption
```

## Backward Compatibility

Adding a variadic `...ForwardOption` parameter is backward compatible in Go.
Existing callers that pass `(ctx, sandboxID, port)` continue to compile
and behave identically since the variadic is empty.

## Error Codes

No change to existing error codes. The service ID option does not
introduce new failure modes.
