# Policy

Accessor: `client.Policy()`

Manage network policies for sandboxes through a draft-based workflow. Policies
go through a draft, review, and approval cycle before being applied.

## Methods

| SDK Method | gRPC RPC | Proto File | Description |
|------------|----------|------------|-------------|
| `GetDraft(ctx, sandboxName string, opts ...GetDraftOption)` | `GetDraftPolicy` | `openshell.proto` | Get the current draft policy |
| `ApproveDraftChunk(ctx, sandboxName, chunkID string)` | `ApproveDraftChunk` | `openshell.proto` | Approve a single draft chunk |
| `RejectDraftChunk(ctx, sandboxName, chunkID, reason string)` | `RejectDraftChunk` | `openshell.proto` | Reject a draft chunk with a reason |
| `ApproveAllDraftChunks(ctx, sandboxName string, opts ...ApproveAllOption)` | `ApproveAllDraftChunks` | `openshell.proto` | Approve all pending draft chunks |
| `ClearDraftChunks(ctx, sandboxName string)` | `ClearDraftChunks` | `openshell.proto` | Clear all draft chunks |
| `GetDraftHistory(ctx, sandboxName string)` | `GetDraftHistory` | `openshell.proto` | Get the history of draft changes |
| `GetStatus(ctx, sandboxName string, opts ...GetStatusOption)` | `GetSandboxPolicyStatus` | `openshell.proto` | Get current policy enforcement status |
| `List(ctx, sandboxName string, opts ...ListPolicyOption)` | `ListSandboxPolicies` | `openshell.proto` | List policy revisions |
| `EditDraftChunk(ctx, sandboxName, chunkID string, proposedRule *NetworkPolicyRule)` | `EditDraftChunk` | `openshell.proto` | Edit a draft chunk's proposed rule |
| `UndoDraftChunk(ctx, sandboxName, chunkID string)` | `UndoDraftChunk` | `openshell.proto` | Undo changes to a draft chunk |

## GetDraft

Retrieve the current draft policy for a sandbox.

```go
draft, err := client.Policy().GetDraft(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
for _, chunk := range draft.Chunks {
    fmt.Printf("Chunk %s: %s -> %s (%s)\n",
        chunk.Id, chunk.Source, chunk.Destination, chunk.Status)
}
```

## ApproveAllDraftChunks

Approve all pending chunks in a single operation.

```go
result, err := client.Policy().ApproveAllDraftChunks(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Approved %d chunks, new revision: %d\n",
    result.ApprovedCount, result.Revision)
```

## GetStatus

Check the current policy enforcement status for a sandbox.

```go
status, err := client.Policy().GetStatus(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Policy mode: %s, active revision: %d\n",
    status.Mode, status.ActiveRevision)
```

## List

List all policy revisions for a sandbox.

```go
revisions, err := client.Policy().List(ctx, "my-sandbox")
if err != nil {
    log.Fatal(err)
}
for _, rev := range revisions {
    fmt.Printf("Revision %d: %s (%d rules)\n",
        rev.Revision, rev.CreatedAt, len(rev.Rules))
}
```

## RejectDraftChunk

Reject a specific draft chunk, providing a reason.

```go
err := client.Policy().RejectDraftChunk(ctx, "my-sandbox", "chunk-abc", "Too permissive")
if err != nil {
    log.Fatal(err)
}
```

## EditDraftChunk

Modify the proposed rule in a draft chunk before approval.

```go
err := client.Policy().EditDraftChunk(ctx, "my-sandbox", "chunk-abc", &v1.NetworkPolicyRule{
    Direction:   "egress",
    Protocol:    "tcp",
    Port:        443,
    Destination: "api.example.com",
})
if err != nil {
    log.Fatal(err)
}
```

See also: [Error Handling](../error-handling.md), [Testing](../testing.md)
