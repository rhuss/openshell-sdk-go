# Upstream Review Findings from PR #53

**Date:** 2026-08-08
**Status:** active
**Context:** Review findings from syncing upstream PR #2271 (Drop A parity)
that should be included in the next upstream SDK contribution PR.

## Summary

During code review of
[PR #53](https://github.com/rhuss/openshell-sdk-go/pull/53), bot reviewers
(Copilot, CodeRabbit) and cc-review agents identified 10 issues. Six were
fixed in the downstream repo. All findings below improve the SDK code that
will be contributed upstream.

## Must Fix (bug)

### Watch goroutine race after Stop()

**File:** `sandbox_client.go` (Watch goroutine, error delivery block)
**Severity:** Bug

When `Stop()` is called while the goroutine is blocked on `stream.Recv()`,
the context cancellation causes Recv to return a context-cancelled error.
The goroutine then enters a `select` with both `ch <-` and `<-w.done` ready,
so Go picks randomly. 50/50 chance consumers see a spurious `EventError`
after calling `Stop()`.

**Fix:** Add a `w.done` check before the error-send select:

```go
select {
case <-w.done:
    return
default:
}
select {
case ch <- Event[*Sandbox]{Type: EventError, Err: ...}:
case <-w.done:
}
```

**Applied in:** `c2fa08db` (downstream)

## Should Fix (robustness)

### TLS config silently dropped with plaintext address

**File:** `internal/grpc/conn.go` (NewConnection)
**Severity:** Robustness

When address uses `http://`, `usePlaintext=true` takes the plaintext branch
and the entire `tlsCfg` parameter is silently ignored. A caller providing
both `http://` and TLS params (CAFile, CertFile, KeyFile) gets no feedback
that their security configuration is discarded.

**Fix:** Return an error when TLS parameters conflict with a plaintext
address:

```go
if usePlaintext {
    if tlsCfg != nil && (tlsCfg.CAFile != "" || tlsCfg.CertFile != "" || tlsCfg.KeyFile != "") {
        return nil, fmt.Errorf("grpc connect: TLS parameters are ignored with plaintext (http://) address")
    }
    opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
}
```

**Applied in:** `c2fa08db` (downstream)

### WaitReady terminal-phase duplication

**File:** `sandbox_client.go` (WaitReady)
**Severity:** Maintainability

The SandboxReady/SandboxError/SandboxDeleting checks appear twice in
WaitReady: before the polling loop (fast path) and inside the ticker case.
Adding a new terminal phase requires updating both places.

**Fix:** Extract a `checkTerminalPhase(sb, name)` helper that both call
sites use.

**Applied in:** `c2fa08db` (downstream)

## Nice to Have (quality)

### Rename grpc_errors.go to context_errors.go

**File:** `grpc_errors.go`
**Severity:** Clarity

The file contains `contextError()` which converts `context.DeadlineExceeded`
and `context.Canceled` into `StatusError`. These are context errors, not gRPC
errors. The file also had a duplicate package comment conflicting with
`doc.go`.

**Fix:** Rename to `context_errors.go`, remove the misleading package
comment.

**Applied in:** `c2fa08db` (downstream)

### Deep-copy tests for CredentialHandles

**File:** `internal/converter/provider_test.go`
**Severity:** Test coverage

No tests verified deep-copy isolation for `CredentialHandles` or its nested
`Metadata` map. The project invariant requires deep copy at boundaries, but
if `CopyStringMap` were accidentally omitted, no test would catch it.

**Fix:** Add mutation tests for both `ProviderFromProto` and
`ProviderToProto` that mutate the source after conversion and assert
independence.

**Applied in:** `c2fa08db` (downstream)

### Remove WHAT comment on StopOnTerminal

**File:** `sandbox_client.go:239`
**Severity:** Style

The comment `// StopOnTerminal: close watcher after delivering a terminal
phase event` restates what the code does. The variable name and the
if-condition are self-documenting.

**Applied in:** `c2fa08db` (downstream)

## YAGNI Removal (for upstream PR)

### Remove unused option types and Config fields

**GitHub:** [#51](https://github.com/rhuss/openshell-sdk-go/issues/51)
**Severity:** API hygiene

Four empty option structs (`CreateOptions`, `GetOptions`, `DeleteOptions`,
`UpdateOptions`) are never accepted by any method signature. Three Config
fields (`Timeout`, `RetryPolicy`, `Logger`) are marked "reserved for future
use" but silently ignored by the gRPC layer. Users setting
`Config{Timeout: 5*time.Second}` get no timeout behavior.

The gateway package sets these fields via `WithTimeout`, `WithRetryPolicy`,
`WithLogger` options, but nothing downstream consumes them.

**Action:** Remove in the next upstream contribution PR:
- Delete `CreateOptions`, `GetOptions`, `DeleteOptions`, `UpdateOptions`
  from `types/options.go` and re-exports from `v1/options.go`
- Delete `WatchOptions.TimeoutSeconds` and `WatchOptions.LabelSelector`
  (keep only `StopOnTerminal`)
- Delete `Config.Timeout`, `Config.RetryPolicy`, `Config.Logger`
- Delete `RetryPolicy` struct from `types/types.go` and re-export from
  `v1/types.go`
- Delete `gateway.WithTimeout`, `gateway.WithRetryPolicy`,
  `gateway.WithLogger` options and related test code
- Add back when actually implemented

## Proto Design Feedback (for upstream discussion)

### CredentialHandles field placement

**File:** `proto/datamodel.proto:99-101`

The `credential_handles` field is documented as "internal gateway state and
is not accepted as user-authored input", but it sits on the mutable
`Provider` message used for both reads and writes. Any SDK doing a
read-modify-write cycle will round-trip `credential_handles` back to the
gateway unless it explicitly strips the field.

**Suggestions for upstream:**
- Add `google.api.field_behavior = OUTPUT_ONLY` annotation
- Or move to a response-only wrapper message
- Or document expected gateway behavior (reject vs. silently ignore)

This is not a code fix but an API design clarification to raise with the
upstream team during the next SDK PR review.

## PR #47 Findings (Reverse Port Forwarding)

From the multi-agent review of PR #47, these findings improve the SDK
code that will be contributed upstream.

### Must Fix (bug)

#### Fake Forward missing sandboxName validation

**File:** `fake/tcp.go` (Forward method)
**Severity:** Critical (flagged by 4 independent agents)

Fake `Forward` was missing `sandboxName` validation that the real client
has. Real client rejects empty sandboxName with `ErrorInvalidArgument`,
but the fake skipped from closedFunc check to port validation. This
violates the fake-real parity invariant.

**Applied in:** `bd0909b1` (downstream)

### Should Fix (robustness)

#### Real RemoteListen discards context parameter

**File:** `tcp_client.go` (RemoteListen)
**Severity:** Important

The stub used `_ context.Context`, hiding the requirement for context in
the future implementation. Forward and Listen both name their context
parameter. When RemoteListen gets a real implementation, the developer
copying the stub signature would miss wiring up context.

**Applied in:** `bd0909b1` (downstream)

#### remoteListenConfig.bindAddress default disagrees with documentation

**File:** `tcp.go` (remoteListenConfig struct)
**Severity:** Minor, security-relevant

The struct's zero-value for `bindAddress` is empty string, but
`WithRemoteBindAddress` documents the default as `"127.0.0.1"`. When the
real implementation lands, an empty bind address could default to
`0.0.0.0` server-side, exposing the forwarded port on all interfaces.

**Fix:** When implementing RemoteListen, initialize config with
`bindAddress: "127.0.0.1"`, matching the `listenConfig` pattern.

#### tcpForwardConn.Close() drops readLoop errors

**File:** `tcp_client.go` (Close method, pre-existing)
**Severity:** Notable

If CloseSend succeeds but the readLoop encountered a transport error,
that error is silently dropped. The caller of Close() gets nil even
though the connection had an error.

**Fix:** After `<-c.done`, return `c.err` if CloseSend returned nil.

### Design Considerations (for real implementation)

#### RemoteListen return type

**File:** `tcp.go` (TCPInterface)

`RemoteListen` returns bare `error`, providing no ready-signal. Unlike
`Listen` which returns `net.Listener` (caller knows the listener is
ready when the call returns), `RemoteListen` gives no way to know when
the remote side is accepting connections. Consider returning a richer
type (e.g., `RemoteListener` with `Ready() <-chan struct{}` and
`Close()`).

#### Closed-client detection in real TCP client

**File:** `tcp_client.go`

The real TCP client has no explicit closed-state check in any method
(Forward, Listen, RemoteListen). It relies on gRPC transport errors to
surface closed connections. The fake client uses `closedFunc()`. This
asymmetry means error messages differ between real and fake when the
client is closed. Consider whether the real client should check its
connection state before making gRPC calls.

## Downstream Commits

All code fixes were applied in the downstream repo and will be included
when the SDK code is contributed upstream:

| Commit | Description |
|--------|-------------|
| `6fc1eda6` | Coverage tests (contextError, WaitReady DeadlineExceeded/Deleting) |
| `e0039c78` | Deterministic deadline in fake sandbox test |
| `c2fa08db` | All six cc-review fixes (helper, race, rename, TLS, tests, comment) |
| `bd0909b1` | PR #47 fixes (fake Forward parity, context param, message assertions, IPv6 case) |
