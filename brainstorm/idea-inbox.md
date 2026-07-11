# Idea Inbox

Ideas captured from code reviews for future brainstorming.

### sandbox-session

- **Source**: discussion
- **Date**: 2026-06-30
- **Reference**: PR #16 (sandbox name resolution), Python SDK pattern
- **Summary**: Add a SandboxSession convenience type that caches the resolved sandbox ID after a single Get() call. Methods like Exec, Upload, Forward hang off the session, avoiding redundant name-to-ID resolution on every call. Coexists with the existing sub-client API as a higher-level layer. Especially valuable for operator reconciler loops doing multiple operations per sandbox per cycle.

> Python SDK uses `SandboxSession` returned by `get_session(name)`. Go SDK currently resolves name-to-ID independently on every sub-client call. Session would cache the ID and provide shorthand methods: `session.Exec(cmd)`, `session.Upload(local, remote)`, etc.

### oshell-file-transfer-ui

- **Source**: brainstorm
- **Date**: 2026-07-03
- **Reference**: Brainstorm #021 (SDK dashboard TUI), out-of-scope items
- **Summary**: Add a file upload/download panel to the oshell dashboard TUI. Would exercise the SDK's Files sub-client (Upload, Download, List) with progress bars and drag-and-drop-style UX in the terminal.

> Deferred from the initial oshell dashboard scope to keep the first version focused. Natural extension once the exec tab is working since file operations pair well with command execution workflows.

### oshell-ssh-terminal

- **Source**: brainstorm
- **Date**: 2026-07-03
- **Reference**: Brainstorm #021 (SDK dashboard TUI), out-of-scope items
- **Summary**: Add an interactive SSH terminal panel inside the oshell dashboard. Would embed a terminal emulator widget in a Bubble Tea tab, using the SDK's SSH tunnel API to connect to a sandbox.

> Deferred because embedding a full terminal emulator inside Bubble Tea is a significant complexity jump. The initial dashboard shows SSH/TCP status but does not provide an interactive shell.

### oshell-provider-import-wizard

- **Source**: brainstorm
- **Date**: 2026-07-03
- **Reference**: Brainstorm #021 (SDK dashboard TUI), out-of-scope items
- **Summary**: Add a guided provider profile import wizard to the oshell dashboard. Step-by-step flow: select YAML file, preview profile fields, import with diagnostics display. Would exercise Providers().Profiles().Import() with the Huh form library.

> Deferred to keep the providers tab read-only in the first version. Import involves file selection and multi-step validation which adds significant UX complexity.

### oshell-credential-refresh-config

- **Source**: brainstorm
- **Date**: 2026-07-03
- **Reference**: Brainstorm #021 (SDK dashboard TUI), out-of-scope items
- **Summary**: Add credential refresh configuration UI to the oshell dashboard. Would let users configure gateway-owned credential refresh for providers (OAuth2 client credentials, API key rotation) through a form interface, exercising Providers().Refresh().Configure().

> Deferred because credential refresh configuration involves sensitive material input (client secrets) and the UX needs careful security consideration for a TUI context.

### oshell-policy-editor

- **Source**: brainstorm
- **Date**: 2026-07-03
- **Reference**: Brainstorm #021 (SDK dashboard TUI), out-of-scope items
- **Summary**: Add an interactive policy editor to the oshell dashboard. Visual editing of filesystem, process, and network policies with real-time validation. Would exercise Policy and Config sub-clients for read-modify-write workflows.

> Deferred because policy editing requires a structured form UI for nested objects (filesystem rules, network endpoints) which is complex to build well in a TUI.

