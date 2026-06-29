# Idea Inbox

Ideas captured from code reviews for future brainstorming.

### log-id-vs-name-mismatch

- **Source**: brainstorm
- **Date**: 2026-06-29
- **Reference**: brainstorm #009 (Phase 2b-2 revisit)
- **Summary**: GetSandboxLogs proto takes sandbox_id while the rest of the SDK uses sandbox name as the primary key. The SDK will resolve name→id internally for consistency, but this asymmetry in the proto may indicate a broader pattern worth auditing across all RPCs.

> brainstorm-009: "The proto GetSandboxLogs takes sandbox_id (not name). The rest of the SDK uses sandbox name everywhere. GetLogs will accept name and resolve internally."
