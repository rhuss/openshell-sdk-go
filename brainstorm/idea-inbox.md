# Idea Inbox

Ideas captured from code reviews for future brainstorming.

### log-id-vs-name-mismatch

- **Source**: brainstorm
- **Date**: 2026-06-29
- **Reference**: brainstorm #009 (Phase 2b-2 revisit)
- **Summary**: GetSandboxLogs proto takes sandbox_id while the rest of the SDK uses sandbox name as the primary key. The SDK will resolve name→id internally for consistency, but this asymmetry in the proto may indicate a broader pattern worth auditing across all RPCs.

> brainstorm-009: "The proto GetSandboxLogs takes sandbox_id (not name). The rest of the SDK uses sandbox name everywhere. GetLogs will accept name and resolve internally."

### local-port-listener

- **Source**: brainstorm
- **Date**: 2026-06-29
- **Reference**: brainstorm #010 (Phase 2b-3), originally deferred in #008
- **Summary**: net.Listener that accepts local connections and tunnels each through Forward() or Tunnel() to a sandbox port. Like `ssh -L 8080:localhost:80`. Would let consumers bind a local port and have the SDK handle the tunneling transparently.

> brainstorm-008 open question: "TCP Forward: should there be an option for local port binding?" resolved as "ReadWriteCloser sufficient for v1." Revisit as a higher-level convenience layer.

### reverse-forwarding

- **Source**: brainstorm
- **Date**: 2026-06-29
- **Reference**: brainstorm #010 (Phase 2b-3)
- **Summary**: Remote-to-local forwarding where the sandbox connects back to the client through the gateway (like `ssh -R`). Requires proto support investigation. May not be feasible with current ForwardTcp RPC which is client-initiated only.

### context-cancellation-session-cleanup

- **Source**: triage
- **Date**: 2026-06-29
- **Reference**: PR #7 (009-ssh-tunnel-forward-opts)
- **Summary**: When a parent context is cancelled on an active SSH tunnel, the gRPC stream terminates but the SSH session is not revoked until the caller explicitly calls Close(). Both copilot and devin flagged this as a gap between spec acceptance scenario 3 ("context cancellation revokes the session") and the implementation (Go convention: caller must call Close()). A background goroutine watching streamCtx.Done() could auto-revoke, but needs careful design to avoid races with the closeOnce pattern.

> copilot: "Tunnel session cleanup isn't triggered by parent context cancellation unless the caller explicitly calls Close()"
> devin: "If the user cancels the parent context but never calls Close(), the session is NOT revoked"
