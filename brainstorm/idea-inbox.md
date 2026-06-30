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

### sandbox-session

- **Source**: discussion
- **Date**: 2026-06-30
- **Reference**: PR #16 (sandbox name resolution), Python SDK pattern
- **Summary**: Add a SandboxSession convenience type that caches the resolved sandbox ID after a single Get() call. Methods like Exec, Upload, Forward hang off the session, avoiding redundant name-to-ID resolution on every call. Coexists with the existing sub-client API as a higher-level layer. Especially valuable for operator reconciler loops doing multiple operations per sandbox per cycle.

> Python SDK uses `SandboxSession` returned by `get_session(name)`. Go SDK currently resolves name-to-ID independently on every sub-client call. Session would cache the ID and provide shorthand methods: `session.Exec(cmd)`, `session.Upload(local, remote)`, etc.

