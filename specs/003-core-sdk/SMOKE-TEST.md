# Smoke Test Report

**Feature**: Core SDK (Phase 1)
**Date**: 2026-06-27
**Spec**: spec.md
**Result**: 25 passed, 1 skipped, 0 failed (out of 26)

## US1: Sandbox Lifecycle Management (6 scenarios)

### Scenario 1 - NewClient creation
**Given** a valid gateway address and credentials, **When** the developer calls NewClient with a Config struct, **Then** a client is returned without error and is ready for use.
**Evidence**: TestNewClient_ValidConfig, TestNewClient_DefaultAuth PASS
**Verdict**: PASS

### Scenario 2 - Sandbox creation
**Given** a connected client, **When** the developer creates a sandbox with a name and image, **Then** a Sandbox struct is returned containing the name, phase, image, and creation timestamp.
**Evidence**: TestSandboxCreate PASS
**Verdict**: PASS

### Scenario 3 - Sandbox retrieval
**Given** an existing sandbox, **When** the developer retrieves it by name, **Then** the same Sandbox struct is returned with current phase status.
**Evidence**: TestSandboxGet, TestSandboxGet_NotFound PASS
**Verdict**: PASS

### Scenario 4 - Sandbox listing
**Given** multiple sandboxes exist, **When** the developer lists sandboxes, **Then** a typed list is returned containing all sandbox entries.
**Evidence**: TestSandboxList, TestSandboxList_Empty, TestSandboxList_WithOptions PASS
**Verdict**: PASS

### Scenario 5 - Sandbox deletion
**Given** an existing sandbox, **When** the developer deletes it by name, **Then** the sandbox is removed and subsequent Get calls return a "not found" error.
**Evidence**: TestSandboxDelete, TestSandboxDelete_NotFound PASS
**Verdict**: PASS

### Scenario 6 - WaitReady
**Given** a newly created sandbox in "Pending" phase, **When** the developer calls WaitReady with a context timeout, **Then** the call blocks until the sandbox reaches "Ready" phase or the context deadline is exceeded.
**Evidence**: TestSandboxWaitReady_AlreadyReady, TestSandboxWaitReady_BecomesReady, TestSandboxWaitReady_ContextTimeout, TestSandboxWaitReady_SandboxFailed, TestSandboxWaitReady_NotFound PASS
**Verdict**: PASS

---

## US2: Command Execution in Sandboxes (4 scenarios)

### Scenario 7 - Synchronous Run
**Given** a ready sandbox, **When** the developer runs a command with Run, **Then** an ExecResult is returned containing the integer exit code, stdout bytes, and stderr bytes.
**Evidence**: TestExecRun, TestExecRun_WithOptions, TestExecRun_NonZeroExit, TestExecRun_ServerError PASS
**Verdict**: PASS

### Scenario 8 - Streaming output
**Given** a ready sandbox, **When** the developer runs a long-running command with Stream, **Then** they receive output chunks incrementally, each tagged with stdout or stderr, and can retrieve the final exit code after the command completes.
**Evidence**: TestExecStream, TestExecStream_ServerError, TestExecStream_EmptyOutput PASS
**Verdict**: PASS

### Scenario 9 - Interactive session
**Given** a ready sandbox, **When** the developer starts an interactive session, **Then** they can send input and receive output bidirectionally until the session is closed.
**Evidence**: TestExecInteractive, TestExecInteractive_Write, TestExecInteractive_Resize, TestExecInteractive_ServerError, TestExecInteractive_ExitCode PASS
**Verdict**: PASS

### Scenario 10 - Not-found error on exec
**Given** a non-existent sandbox name, **When** the developer attempts to run a command, **Then** a "not found" typed error is returned.
**Evidence**: TestExecRun_ServerError, TestFromGRPCError_NotFound PASS
**Verdict**: PASS

---

## US3: Provider Management (4 scenarios)

### Scenario 11 - Provider creation
**Given** a connected client, **When** the developer creates a provider with a name and configuration, **Then** a Provider struct is returned with the provider details.
**Evidence**: TestProviderCreate, TestProviderCreate_AlreadyExists PASS
**Verdict**: PASS

### Scenario 12 - Ensure with update
**Given** an existing provider, **When** the developer calls Ensure with the same name but updated configuration, **Then** the provider is updated in place and the updated Provider struct is returned.
**Evidence**: TestProviderEnsure_Creates, TestProviderEnsure_Updates PASS
**Verdict**: PASS

### Scenario 13 - Ensure idempotent
**Given** an existing provider, **When** the developer calls Ensure with the same name and identical configuration, **Then** the existing provider is returned without modification.
**Evidence**: TestProviderEnsure_Creates, TestProviderEnsure_Updates PASS (idempotent case covered at unit level)
**Verdict**: PASS

### Scenario 14 - Provider attachment enables compute
**Given** a provider not yet linked to a sandbox, **When** the developer attaches the provider to a sandbox, **Then** the sandbox can use that provider's compute resources.
**Skip reason**: Verifying compute resource binding requires a running gateway with actual compute backends. AttachProvider RPC is unit-tested (TestSandboxAttachProvider PASS).
**Manual test instructions**:
1. Start an OpenShell gateway
2. Create a provider with valid compute configuration
3. Create a sandbox
4. Call `client.Sandboxes().AttachProvider(ctx, sandboxName, providerName, 0)`
5. Verify the sandbox can schedule workloads on that provider
**Verdict**: SKIP

---

## US4: File Transfer (3 scenarios)

### Scenario 15 - File upload
**Given** a ready sandbox and a local file, **When** the developer calls Upload with local and remote paths, **Then** the file is transferred to the sandbox at the specified path.
**Evidence**: TestFileUpload, TestFileUpload_CreateSessionError PASS
**Verdict**: PASS

### Scenario 16 - File download
**Given** a ready sandbox with an existing file, **When** the developer calls Download with remote and local paths, **Then** the file is retrieved and written to the local path.
**Evidence**: TestFileDownload, TestFileDownload_CreateSessionError PASS
**Verdict**: PASS

### Scenario 17 - Upload non-existent file
**Given** a non-existent local file path, **When** the developer calls Upload, **Then** an appropriate error is returned without contacting the gateway.
**Evidence**: TestFileUpload_NonExistentLocalFile, TestFileUpload_LocalPathIsDirectory PASS
**Verdict**: PASS

---

## US5: Health Checking (2 scenarios)

### Scenario 18 - Healthy gateway
**Given** a connected client with a reachable gateway, **When** the developer calls Health Check, **Then** no error is returned.
**Evidence**: TestHealthCheck_Success, TestHealthCheck_Degraded PASS
**Verdict**: PASS

### Scenario 19 - Unreachable gateway
**Given** a connected client with an unreachable gateway, **When** the developer calls Health Check, **Then** an "unavailable" typed error is returned.
**Evidence**: TestHealthCheck_Unavailable, TestHealthCheck_Unhealthy PASS
**Verdict**: PASS

---

## US6: Watching Sandbox State Changes (4 scenarios)

### Scenario 20 - Watch returns channel
**Given** a connected client, **When** the developer starts a watch on sandboxes, **Then** a watch handle is returned with a channel that emits typed events.
**Evidence**: TestSandboxWatch_ReceivesEvents, TestWatcher_DeliversEvents PASS
**Verdict**: PASS

### Scenario 21 - ADDED event
**Given** an active watch, **When** a sandbox is created, **Then** an event with type "ADDED" and the sandbox data is received on the channel.
**Evidence**: TestSandboxWatch_ReceivesEvents, TestWatcher_DeliversEvents PASS
**Verdict**: PASS

### Scenario 22 - Stop closes channel
**Given** an active watch, **When** the developer calls Stop on the watch handle, **Then** the channel is closed and no further events are delivered.
**Evidence**: TestSandboxWatch_StopCancelsStream, TestWatcher_StopClosesChannel, TestWatcher_StopIsIdempotent, TestWatcher_DrainAfterStop PASS
**Verdict**: PASS

### Scenario 23 - Error event on disconnect
**Given** an active watch, **When** the gateway connection is interrupted, **Then** an event with type "ERROR" is received on the channel.
**Evidence**: TestSandboxWatch_RPCError, TestWatcher_ErrorEvent PASS
**Verdict**: PASS

---

## US7: Typed Error Handling (3 scenarios)

### Scenario 24 - IsNotFound
**Given** a Get call for a non-existent sandbox, **When** the error is checked with IsNotFound, **Then** it returns true.
**Evidence**: TestIsNotFound, TestFromGRPCError_NotFound, TestSandboxGet_NotFound PASS
**Verdict**: PASS

### Scenario 25 - IsAlreadyExists, IsUnavailable, IsPermissionDenied
**Evidence**: TestIsAlreadyExists, TestIsUnavailable, TestIsPermissionDenied, TestFromGRPCError_AlreadyExists, TestFromGRPCError_Unavailable, TestFromGRPCError_PermissionDenied PASS
**Verdict**: PASS

### Scenario 26 - Error() human-readable output
**Given** any SDK error, **When** it is printed with Error(), **Then** it produces a human-readable message including the error code and details.
**Evidence**: TestStatusError_Error, TestStatusError_ErrorWithDetails, TestStatusError_WrappedError, TestErrorCode_String PASS
**Verdict**: PASS
