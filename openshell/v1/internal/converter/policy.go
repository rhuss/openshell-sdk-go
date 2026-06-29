// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"google.golang.org/protobuf/proto"

	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
)

// --- PolicyLoadStatus enum mapping ---

// PolicyLoadStatusFromProto converts a proto PolicyStatus to an SDK PolicyLoadStatus.
func PolicyLoadStatusFromProto(s pb.PolicyStatus) types.PolicyLoadStatus {
	switch s {
	case pb.PolicyStatus_POLICY_STATUS_PENDING:
		return types.PolicyLoadStatusPending
	case pb.PolicyStatus_POLICY_STATUS_LOADED:
		return types.PolicyLoadStatusLoaded
	case pb.PolicyStatus_POLICY_STATUS_FAILED:
		return types.PolicyLoadStatusFailed
	case pb.PolicyStatus_POLICY_STATUS_SUPERSEDED:
		return types.PolicyLoadStatusSuperseded
	default:
		return types.PolicyLoadStatusUnspecified
	}
}

// PolicyLoadStatusToProto converts an SDK PolicyLoadStatus to a proto PolicyStatus.
func PolicyLoadStatusToProto(s types.PolicyLoadStatus) pb.PolicyStatus {
	switch s {
	case types.PolicyLoadStatusPending:
		return pb.PolicyStatus_POLICY_STATUS_PENDING
	case types.PolicyLoadStatusLoaded:
		return pb.PolicyStatus_POLICY_STATUS_LOADED
	case types.PolicyLoadStatusFailed:
		return pb.PolicyStatus_POLICY_STATUS_FAILED
	case types.PolicyLoadStatusSuperseded:
		return pb.PolicyStatus_POLICY_STATUS_SUPERSEDED
	default:
		return pb.PolicyStatus_POLICY_STATUS_UNSPECIFIED
	}
}

// --- PolicyChunk ---

// PolicyChunkFromProto converts a proto PolicyChunk to an SDK PolicyChunk.
func PolicyChunkFromProto(c *pb.PolicyChunk) *types.PolicyChunk {
	if c == nil {
		return nil
	}
	return &types.PolicyChunk{
		ID:                c.GetId(),
		Status:            c.GetStatus(),
		RuleName:          c.GetRuleName(),
		ProposedRule:      NetworkPolicyRuleFromProto(c.GetProposedRule()),
		Rationale:         c.GetRationale(),
		SecurityNotes:     c.GetSecurityNotes(),
		Confidence:        c.GetConfidence(),
		DenialSummaryIDs:  CopyStringSlice(c.GetDenialSummaryIds()),
		CreatedAt:         TimeFromMillis(c.GetCreatedAtMs()),
		DecidedAt:         TimeFromMillis(c.GetDecidedAtMs()),
		Stage:             c.GetStage(),
		SupersedesChunkID: c.GetSupersedesChunkId(),
		HitCount:          c.GetHitCount(),
		FirstSeen:         TimeFromMillis(c.GetFirstSeenMs()),
		LastSeen:          TimeFromMillis(c.GetLastSeenMs()),
		Binary:            c.GetBinary(),
		ValidationResult:  c.GetValidationResult(),
		RejectionReason:   c.GetRejectionReason(),
	}
}

// --- DraftPolicy ---

// DraftPolicyFromProto converts a proto GetDraftPolicyResponse to an SDK DraftPolicy.
func DraftPolicyFromProto(r *pb.GetDraftPolicyResponse) *types.DraftPolicy {
	if r == nil {
		return nil
	}
	result := &types.DraftPolicy{
		RollingSummary: r.GetRollingSummary(),
		DraftVersion:   r.GetDraftVersion(),
		LastAnalyzedAt: TimeFromMillis(r.GetLastAnalyzedAtMs()),
	}
	if chunks := r.GetChunks(); len(chunks) > 0 {
		result.Chunks = make([]types.PolicyChunk, 0, len(chunks))
		for _, c := range chunks {
			if converted := PolicyChunkFromProto(c); converted != nil {
				result.Chunks = append(result.Chunks, *converted)
			}
		}
	}
	return result
}

// --- SandboxPolicyRevision ---

// SandboxPolicyRevisionFromProto converts a proto SandboxPolicyRevision to an SDK SandboxPolicyRevision.
func SandboxPolicyRevisionFromProto(r *pb.SandboxPolicyRevision) *types.SandboxPolicyRevision {
	if r == nil {
		return nil
	}
	var policyBytes []byte
	if p := r.GetPolicy(); p != nil {
		if b, err := proto.Marshal(p); err == nil {
			policyBytes = CopyByteSlice(b)
		}
	}
	return &types.SandboxPolicyRevision{
		Version:    r.GetVersion(),
		PolicyHash: r.GetPolicyHash(),
		Status:     PolicyLoadStatusFromProto(r.GetStatus()),
		LoadError:  r.GetLoadError(),
		CreatedAt:  TimeFromMillis(r.GetCreatedAtMs()),
		LoadedAt:   TimeFromMillis(r.GetLoadedAtMs()),
		Policy:     policyBytes,
	}
}

// --- PolicyStatusResult ---

// PolicyStatusResultFromProto converts a proto GetSandboxPolicyStatusResponse to an SDK PolicyStatusResult.
func PolicyStatusResultFromProto(r *pb.GetSandboxPolicyStatusResponse) *types.PolicyStatusResult {
	if r == nil {
		return nil
	}
	result := &types.PolicyStatusResult{
		ActiveVersion: r.GetActiveVersion(),
	}
	if rev := SandboxPolicyRevisionFromProto(r.GetRevision()); rev != nil {
		result.Revision = *rev
	}
	return result
}

// --- ApproveResult ---

// ApproveResultFromProto converts a proto ApproveDraftChunkResponse to an SDK ApproveResult.
func ApproveResultFromProto(r *pb.ApproveDraftChunkResponse) *types.ApproveResult {
	if r == nil {
		return nil
	}
	return &types.ApproveResult{
		PolicyVersion: r.GetPolicyVersion(),
		PolicyHash:    r.GetPolicyHash(),
	}
}

// --- ApproveAllResult ---

// ApproveAllResultFromProto converts a proto ApproveAllDraftChunksResponse to an SDK ApproveAllResult.
func ApproveAllResultFromProto(r *pb.ApproveAllDraftChunksResponse) *types.ApproveAllResult {
	if r == nil {
		return nil
	}
	return &types.ApproveAllResult{
		PolicyVersion:  r.GetPolicyVersion(),
		PolicyHash:     r.GetPolicyHash(),
		ChunksApproved: r.GetChunksApproved(),
		ChunksSkipped:  r.GetChunksSkipped(),
	}
}

// --- UndoResult ---

// UndoResultFromProto converts a proto UndoDraftChunkResponse to an SDK UndoResult.
func UndoResultFromProto(r *pb.UndoDraftChunkResponse) *types.UndoResult {
	if r == nil {
		return nil
	}
	return &types.UndoResult{
		PolicyVersion: r.GetPolicyVersion(),
		PolicyHash:    r.GetPolicyHash(),
	}
}

// --- ClearResult ---

// ClearResultFromProto converts a proto ClearDraftChunksResponse to an SDK ClearResult.
func ClearResultFromProto(r *pb.ClearDraftChunksResponse) *types.ClearResult {
	if r == nil {
		return nil
	}
	return &types.ClearResult{
		ChunksCleared: r.GetChunksCleared(),
	}
}

// --- DraftHistoryEntry ---

// DraftHistoryEntryFromProto converts a proto DraftHistoryEntry to an SDK DraftHistoryEntry.
func DraftHistoryEntryFromProto(e *pb.DraftHistoryEntry) *types.DraftHistoryEntry {
	if e == nil {
		return nil
	}
	return &types.DraftHistoryEntry{
		Timestamp:   TimeFromMillis(e.GetTimestampMs()),
		EventType:   e.GetEventType(),
		Description: e.GetDescription(),
		ChunkID:     e.GetChunkId(),
	}
}
