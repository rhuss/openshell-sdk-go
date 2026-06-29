# Data Model: Phase 2b-2 Policy, Logs, MergeOps, ErrorConflict

## Domain Types Overview

```
types/
├── errors.go           # ErrorConflict (new), IsConflict (new)
├── policy.go           # PolicyChunk, DraftPolicy, PolicyStatusResult, SandboxPolicyRevision,
│                       #   PolicyLoadStatus, ApproveResult, ApproveAllResult, UndoResult,
│                       #   ClearResult, DraftHistoryEntry
│                       #   + functional option types: GetDraftOption, ApproveAllOption,
│                       #     GetStatusOption, ListPolicyOption
├── log.go              # LogLine, LogResult, LogOption
├── network_policy.go   # NetworkPolicyRule, NetworkEndpoint, NetworkBinary, L7Rule, L7Allow,
│                       #   L7DenyRule, L7QueryMatcher, GraphqlOperation,
│                       #   PolicyMergeOperation, AddNetworkRule, RemoveNetworkEndpoint,
│                       #   RemoveNetworkRule, AddDenyRules, AddAllowRules, RemoveNetworkBinary
└── setting.go          # ConfigUpdate.MergeOperations []byte → []PolicyMergeOperation
```

## Error Types

### ErrorConflict (types/errors.go)

```
ErrorCode enum:
  + ErrorConflict (new, after ErrorUnimplemented)

grpcToSDK map:
  + codes.Aborted → types.ErrorConflict

Helper:
  + IsConflict(err error) bool
```

## Policy Types (types/policy.go)

### PolicyChunk

| Field | Type | Proto Source | Notes |
|-------|------|-------------|-------|
| ID | string | id | Unique chunk identifier |
| Status | string | status | "pending", "approved", "rejected" |
| RuleName | string | rule_name | Proposed network_policies map key |
| ProposedRule | *NetworkPolicyRule | proposed_rule | The proposed network policy rule |
| Rationale | string | rationale | Human-readable explanation |
| SecurityNotes | string | security_notes | Security concerns (empty if none) |
| Confidence | float32 | confidence | 0.0-1.0 confidence score |
| DenialSummaryIDs | []string | denial_summary_ids | Source denial IDs |
| CreatedAt | time.Time | created_at_ms | Converted from ms epoch |
| DecidedAt | time.Time | decided_at_ms | Converted from ms epoch |
| Stage | string | stage | "initial" or "refined" |
| SupersedesChunkID | string | supersedes_chunk_id | For refined chunks |
| HitCount | int32 | hit_count | Denial flush cycle count |
| FirstSeen | time.Time | first_seen_ms | Converted from ms epoch |
| LastSeen | time.Time | last_seen_ms | Converted from ms epoch |
| Binary | string | binary | Binary path that triggered denial |
| ValidationResult | string | validation_result | Prover output |
| RejectionReason | string | rejection_reason | Operator rejection text |

### DraftPolicy

| Field | Type | Proto Source |
|-------|------|-------------|
| Chunks | []PolicyChunk | chunks |
| RollingSummary | string | rolling_summary |
| DraftVersion | uint64 | draft_version |
| LastAnalyzedAt | time.Time | last_analyzed_at_ms |

### PolicyStatusResult

| Field | Type | Proto Source |
|-------|------|-------------|
| Revision | SandboxPolicyRevision | revision |
| ActiveVersion | uint32 | active_version |

### SandboxPolicyRevision

| Field | Type | Proto Source |
|-------|------|-------------|
| Version | uint32 | version |
| PolicyHash | string | policy_hash |
| Status | PolicyLoadStatus | status |
| LoadError | string | load_error |
| CreatedAt | time.Time | created_at_ms |
| LoadedAt | time.Time | loaded_at_ms |
| Policy | []byte | policy (opaque) |

### PolicyLoadStatus (enum)

| Value | Proto |
|-------|-------|
| PolicyLoadStatusUnspecified | POLICY_STATUS_UNSPECIFIED |
| PolicyLoadStatusPending | POLICY_STATUS_PENDING |
| PolicyLoadStatusLoaded | POLICY_STATUS_LOADED |
| PolicyLoadStatusFailed | POLICY_STATUS_FAILED |
| PolicyLoadStatusSuperseded | POLICY_STATUS_SUPERSEDED |

### Result Types

**ApproveResult**: PolicyVersion (uint32), PolicyHash (string)
**ApproveAllResult**: PolicyVersion (uint32), PolicyHash (string), ChunksApproved (uint32), ChunksSkipped (uint32)
**UndoResult**: PolicyVersion (uint32), PolicyHash (string)
**ClearResult**: ChunksCleared (uint32)

### DraftHistoryEntry

| Field | Type | Proto Source |
|-------|------|-------------|
| Timestamp | time.Time | timestamp_ms |
| EventType | string | event_type |
| Description | string | description |
| ChunkID | string | chunk_id |

### Functional Option Types

```go
type GetDraftOption func(*getDraftConfig)
type ApproveAllOption func(*approveAllConfig)
type GetStatusOption func(*getStatusConfig)
type ListPolicyOption func(*listPolicyConfig)
```

## Log Types (types/log.go)

### LogLine

| Field | Type | Proto Source |
|-------|------|-------------|
| Timestamp | time.Time | timestamp_ms |
| Level | string | level |
| Target | string | target |
| Message | string | message |
| Source | string | source |
| Fields | map[string]string | fields |

### LogResult

| Field | Type | Proto Source |
|-------|------|-------------|
| Lines | []LogLine | logs |
| BufferTotal | uint32 | buffer_total |

### LogOption

```go
type LogOption func(*logConfig)
// WithLogLines, WithLogSince, WithLogSources, WithLogMinLevel
```

## Network Policy Types (types/network_policy.go)

### NetworkPolicyRule

| Field | Type | Proto Source |
|-------|------|-------------|
| Name | string | name |
| Endpoints | []NetworkEndpoint | endpoints |
| Binaries | []NetworkBinary | binaries |

### NetworkEndpoint

| Field | Type | Proto Source |
|-------|------|-------------|
| Host | string | host |
| Port | uint32 | port |
| Ports | []uint32 | ports |
| Protocol | string | protocol |
| TLS | string | tls |
| Enforcement | string | enforcement |
| Access | string | access |
| Rules | []L7Rule | rules |
| AllowedIPs | []string | allowed_ips |
| DenyRules | []L7DenyRule | deny_rules |
| AllowEncodedSlash | bool | allow_encoded_slash |
| PersistedQueries | string | persisted_queries |
| GraphqlPersistedQueries | map[string]GraphqlOperation | graphql_persisted_queries |
| GraphqlMaxBodyBytes | uint32 | graphql_max_body_bytes |
| Path | string | path |
| WebsocketCredentialRewrite | bool | websocket_credential_rewrite |
| RequestBodyCredentialRewrite | bool | request_body_credential_rewrite |
| AdvisorProposed | bool | advisor_proposed |

### NetworkBinary

| Field | Type |
|-------|------|
| Path | string |

### L7Rule

| Field | Type |
|-------|------|
| Allow | *L7Allow |

### L7Allow / L7DenyRule (same field set)

| Field | Type |
|-------|------|
| Method | string |
| Path | string |
| Command | string |
| Query | map[string]L7QueryMatcher |
| OperationType | string |
| OperationName | string |
| Fields | []string |

### L7QueryMatcher

| Field | Type |
|-------|------|
| Glob | string |
| Any | []string |

### GraphqlOperation

| Field | Type |
|-------|------|
| OperationType | string |
| OperationName | string |
| Fields | []string |

### PolicyMergeOperation

| Field | Type | Proto Oneof |
|-------|------|-------------|
| AddRule | *AddNetworkRule | add_rule |
| RemoveEndpoint | *RemoveNetworkEndpoint | remove_endpoint |
| RemoveRule | *RemoveNetworkRule | remove_rule |
| AddDenyRules | *AddDenyRules | add_deny_rules |
| AddAllowRules | *AddAllowRules | add_allow_rules |
| RemoveBinary | *RemoveNetworkBinary | remove_binary |

### MergeOperation Variant Types

**AddNetworkRule**: RuleName (string), Rule (NetworkPolicyRule)
**RemoveNetworkEndpoint**: RuleName (string), Host (string), Port (uint32)
**RemoveNetworkRule**: RuleName (string)
**AddDenyRules**: Host (string), Port (uint32), DenyRules ([]L7DenyRule)
**AddAllowRules**: Host (string), Port (uint32), Rules ([]L7Rule)
**RemoveNetworkBinary**: RuleName (string), BinaryPath (string)

## ConfigUpdate Change (types/setting.go)

```
Before: MergeOperations []byte
After:  MergeOperations []PolicyMergeOperation
```

## Converter Mapping Summary

| Converter File | Proto → SDK | SDK → Proto |
|---------------|-------------|-------------|
| policy.go | PolicyChunkFromProto, DraftPolicyFromProto, PolicyStatusResultFromProto, SandboxPolicyRevisionFromProto, DraftHistoryEntryFromProto, ApproveResultFromProto, ApproveAllResultFromProto, UndoResultFromProto, ClearResultFromProto | (read-only types, no SDK→proto needed for chunks) |
| network_policy.go | NetworkPolicyRuleFromProto, NetworkEndpointFromProto, L7RuleFromProto, L7AllowFromProto, L7DenyRuleFromProto, L7QueryMatcherFromProto, GraphqlOperationFromProto, NetworkBinaryFromProto | NetworkPolicyRuleToProto + all nested ToProto (needed for EditDraftChunk, MergeOperations) |
| log.go | LogLineFromProto, LogResultFromProto | LogOptionsToProto (request params only) |
| setting.go | (existing) | + PolicyMergeOperationToProto (for ConfigUpdate) |
| errors.go | + codes.Aborted → ErrorConflict | (N/A) |
