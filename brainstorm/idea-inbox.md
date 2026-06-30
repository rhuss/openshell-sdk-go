# Idea Inbox

Ideas captured from code reviews for future brainstorming.

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

