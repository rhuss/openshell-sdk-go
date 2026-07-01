# Idea Inbox

Ideas captured from code reviews for future brainstorming.

### sandbox-session

- **Source**: discussion
- **Date**: 2026-06-30
- **Reference**: PR #16 (sandbox name resolution), Python SDK pattern
- **Summary**: Add a SandboxSession convenience type that caches the resolved sandbox ID after a single Get() call. Methods like Exec, Upload, Forward hang off the session, avoiding redundant name-to-ID resolution on every call. Coexists with the existing sub-client API as a higher-level layer. Especially valuable for operator reconciler loops doing multiple operations per sandbox per cycle.

> Python SDK uses `SandboxSession` returned by `get_session(name)`. Go SDK currently resolves name-to-ID independently on every sub-client call. Session would cache the ID and provide shorthand methods: `session.Exec(cmd)`, `session.Upload(local, remote)`, etc.

