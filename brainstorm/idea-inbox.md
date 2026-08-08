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

### credential-handles-write-behavior

- **Source**: triage
- **Date**: 2026-08-08
- **Reference**: PR #53 (sync upstream PR #2271)
- **Summary**: CredentialHandles is documented as "not accepted as user-authored input" in the proto, but sits on the mutable Provider message used for both reads and writes. Any SDK doing read-modify-write round-trips the field back to the gateway. Needs upstream clarification (OUTPUT_ONLY annotation, separate response message, or documented reject/ignore behavior) and SDK-side stripping in Create/Update methods.

> Deferred from PR #53: Copilot flagged ProviderToProto serializing CredentialHandles; cc-review flagged the duplicated context-error wrapping in fake/sandbox.go (related circular dependency). Both point to the same upstream API design gap. See brainstorm/030-upstream-review-findings.md for full details.

### context-error-extraction

- **Source**: triage
- **Date**: 2026-08-08
- **Reference**: PR #53 (sync upstream PR #2271)
- **Summary**: The context-error wrapping logic (context.DeadlineExceeded/Canceled to StatusError) is duplicated between v1/context_errors.go and fake/sandbox.go due to a circular import constraint. Extract a shared ContextStatusError helper into the types package, which both v1 and fake already import.

> Deferred from PR #53: cc-review architecture agent flagged the duplication. The fake package can't import v1 (circular dependency), so both maintain independent switch statements. A types-level helper would eliminate the drift risk flagged by the "fake-real parity" invariant.

### pr-description-undeclared-changes

- **Source**: triage
- **Date**: 2026-08-08
- **Reference**: PR #53 (sync upstream PR #2271)
- **Summary**: Three behavioral changes in PR #53 were not declared in the PR description: (1) StatusError.Details replaced with Cause/Unwrap (breaking API change), (2) WaitReady now wraps context errors in StatusError instead of returning raw context.Err(), (3) CertFile/KeyFile mutual-dependency validation added in conn.go. Future PRs should document behavioral contract changes.

> Deferred from PR #53: cc-review goal-alignment agent flagged all three. Not code fixes, but a process improvement for PR hygiene. The changes themselves are correct.

### workspace-role-unknown-variant

- **Source**: deep-review
- **Date**: 2026-07-31
- **Reference**: 022-workspace-crud-gatewayinfo
- **Summary**: `WorkspaceRoleFromProto` defaults unrecognized proto values to `WorkspaceRoleUser`, unlike other enum converters that default to an "Unknown" variant. Consider adding `WorkspaceRoleUnknown` for forward compatibility.

> Every other enum converter in the codebase (WorkspacePhaseFromProto, ServiceStatusFromProto) defaults to an "Unknown" variant for unrecognized values. WorkspaceRole lacks an Unknown variant, forcing the default to User (least privilege). If the gateway adds new roles in the future, they will silently map to User rather than surfacing as unrecognized.

