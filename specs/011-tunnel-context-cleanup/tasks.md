# Tasks: SSH Tunnel Context Cancellation Cleanup

## Implementation Tasks

- [x] T001 Add cleanup goroutine to `Tunnel()` in `ssh_client.go` that watches `conn.done` and calls `sshTunnel.Close()` when it fires. Launch the goroutine after constructing the `sshTunnel` value but before returning it. The goroutine blocks on `<-conn.done`, then calls `t.Close()`. No new struct fields needed.

- [x] T002 Add test `TestSSHTunnel_ContextCancelRevokesSession` to `ssh_client_test.go`: create a tunnel with a cancellable context, cancel the context without calling Close(), wait for the done channel to close, then assert the SSH session token was revoked on the mock server. Use `require.Eventually` or a short sleep to allow the cleanup goroutine to run.

- [x] T003 Add test `TestSSHTunnel_ContextCancelThenClose` to `ssh_client_test.go`: create a tunnel, cancel the context, then explicitly call Close(). Assert that exactly one revocation occurred (token active=false) and no panic. This validates the closeOnce idempotency under the race between cleanup goroutine and explicit Close().

- [x] T004 Add test `TestSSHTunnel_CloseBeforeContextCancel` to `ssh_client_test.go`: create a tunnel with a cancellable context, call Close() explicitly (which triggers revocation), then cancel the context. Use `require.Never` to verify the cleanup goroutine does not trigger a second revocation. Assert revokeCount == 1 (session revoked exactly once).

- [x] T005 Run `make ci` to verify all existing tests pass, race detector is clean, and linter is satisfied.
