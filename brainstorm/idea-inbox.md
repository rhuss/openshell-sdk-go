# Idea Inbox

Ideas captured from code reviews for future brainstorming.

### merge-operations-serialization

- **Source**: triage
- **Date**: 2026-06-29
- **Reference**: PR #5 (007-ssh-tcp-config)
- **Summary**: ConfigUpdate.MergeOperations field is accepted in the SDK types but not serialized into the proto request. Two independent reviewers (Devin, CodeRabbit) flagged this. Full PolicyMergeOperation type hierarchy should be implemented in Phase 2b-2.

> devin-ai-integration: "Configuration merge operations are silently discarded when updating settings" (setting.go:193)
> coderabbitai: "Serialize MergeOperations in config updates — this converter drops it entirely" (setting.go:203)

### optimistic-concurrency-error-codes

- **Source**: triage
- **Date**: 2026-06-29
- **Reference**: PR #5 (007-ssh-tcp-config)
- **Summary**: The SDK error mapper maps unmapped gRPC codes (ABORTED, FAILED_PRECONDITION) to the catch-all ErrorInternal. Two reviewers (Copilot, Devin) independently flagged that optimistic concurrency errors should have a dedicated ErrorCode (ErrorConflict or ErrorAborted) for callers to handle version mismatches programmatically.

> copilot: "The SDK currently maps unhandled gRPC codes to ErrorInternal; consider adding a dedicated ErrorCode for concurrency cases" (config_client_test.go:435)
> devin-ai-integration: "Optimistic concurrency errors misclassified as Internal due to pre-existing error mapper gap" (config_client_test.go:435)
