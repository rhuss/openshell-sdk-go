// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"testing"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1/types"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	sbv1 "github.com/rhuss/openshell-sdk-go/proto/sandboxv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- PolicyLoadStatus ---

func TestPolicyLoadStatusFromProto(t *testing.T) {
	tests := []struct {
		proto pb.PolicyStatus
		want  v1.PolicyLoadStatus
	}{
		{pb.PolicyStatus_POLICY_STATUS_UNSPECIFIED, v1.PolicyLoadStatusUnspecified},
		{pb.PolicyStatus_POLICY_STATUS_PENDING, v1.PolicyLoadStatusPending},
		{pb.PolicyStatus_POLICY_STATUS_LOADED, v1.PolicyLoadStatusLoaded},
		{pb.PolicyStatus_POLICY_STATUS_FAILED, v1.PolicyLoadStatusFailed},
		{pb.PolicyStatus_POLICY_STATUS_SUPERSEDED, v1.PolicyLoadStatusSuperseded},
	}
	for _, tt := range tests {
		t.Run(tt.proto.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, PolicyLoadStatusFromProto(tt.proto))
		})
	}
}

func TestPolicyLoadStatusToProto(t *testing.T) {
	tests := []struct {
		sdk  v1.PolicyLoadStatus
		want pb.PolicyStatus
	}{
		{v1.PolicyLoadStatusUnspecified, pb.PolicyStatus_POLICY_STATUS_UNSPECIFIED},
		{v1.PolicyLoadStatusPending, pb.PolicyStatus_POLICY_STATUS_PENDING},
		{v1.PolicyLoadStatusLoaded, pb.PolicyStatus_POLICY_STATUS_LOADED},
		{v1.PolicyLoadStatusFailed, pb.PolicyStatus_POLICY_STATUS_FAILED},
		{v1.PolicyLoadStatusSuperseded, pb.PolicyStatus_POLICY_STATUS_SUPERSEDED},
	}
	for _, tt := range tests {
		t.Run(tt.sdk.String(), func(t *testing.T) {
			assert.Equal(t, tt.want, PolicyLoadStatusToProto(tt.sdk))
		})
	}
}

func TestPolicyLoadStatusRoundTrip(t *testing.T) {
	for _, s := range []v1.PolicyLoadStatus{
		v1.PolicyLoadStatusUnspecified,
		v1.PolicyLoadStatusPending,
		v1.PolicyLoadStatusLoaded,
		v1.PolicyLoadStatusFailed,
		v1.PolicyLoadStatusSuperseded,
	} {
		assert.Equal(t, s, PolicyLoadStatusFromProto(PolicyLoadStatusToProto(s)))
	}
}

// --- PolicyChunk ---

func TestPolicyChunkFromProto(t *testing.T) {
	proto := &pb.PolicyChunk{
		Id:            "chunk-1",
		Status:        "pending",
		RuleName:      "web-api",
		Rationale:     "Observed DNS resolution",
		SecurityNotes: "No concerns",
		Confidence:    0.95,
		DenialSummaryIds: []string{"d1", "d2"},
		CreatedAtMs:   1700000000000,
		DecidedAtMs:   1700000001000,
		Stage:         "initial",
		SupersedesChunkId: "chunk-0",
		HitCount:      5,
		FirstSeenMs:   1699999999000,
		LastSeenMs:    1700000000500,
		Binary:        "/usr/bin/curl",
		ValidationResult: "valid",
		RejectionReason:  "",
		ProposedRule: &sbv1.NetworkPolicyRule{
			Name: "web-api",
			Endpoints: []*sbv1.NetworkEndpoint{
				{Host: "api.example.com", Port: 443, Protocol: "rest"},
			},
		},
	}

	chunk := PolicyChunkFromProto(proto)

	require.NotNil(t, chunk)
	assert.Equal(t, "chunk-1", chunk.ID)
	assert.Equal(t, "pending", chunk.Status)
	assert.Equal(t, "web-api", chunk.RuleName)
	assert.Equal(t, "Observed DNS resolution", chunk.Rationale)
	assert.Equal(t, "No concerns", chunk.SecurityNotes)
	assert.InDelta(t, float32(0.95), chunk.Confidence, 0.001)
	assert.Equal(t, []string{"d1", "d2"}, chunk.DenialSummaryIDs)
	assert.False(t, chunk.CreatedAt.IsZero())
	assert.False(t, chunk.DecidedAt.IsZero())
	assert.Equal(t, "initial", chunk.Stage)
	assert.Equal(t, "chunk-0", chunk.SupersedesChunkID)
	assert.Equal(t, int32(5), chunk.HitCount)
	assert.False(t, chunk.FirstSeen.IsZero())
	assert.False(t, chunk.LastSeen.IsZero())
	assert.Equal(t, "/usr/bin/curl", chunk.Binary)
	assert.Equal(t, "valid", chunk.ValidationResult)
	assert.Empty(t, chunk.RejectionReason)

	require.NotNil(t, chunk.ProposedRule)
	assert.Equal(t, "web-api", chunk.ProposedRule.Name)
	require.Len(t, chunk.ProposedRule.Endpoints, 1)
	assert.Equal(t, "api.example.com", chunk.ProposedRule.Endpoints[0].Host)
}

func TestPolicyChunkFromProto_Nil(t *testing.T) {
	assert.Nil(t, PolicyChunkFromProto(nil))
}

func TestPolicyChunkDeepCopy(t *testing.T) {
	proto := &pb.PolicyChunk{
		Id:               "c1",
		DenialSummaryIds: []string{"d1"},
	}

	chunk := PolicyChunkFromProto(proto)
	proto.DenialSummaryIds[0] = "changed"

	assert.Equal(t, "d1", chunk.DenialSummaryIDs[0])
}

// --- DraftPolicy ---

func TestDraftPolicyFromProto(t *testing.T) {
	proto := &pb.GetDraftPolicyResponse{
		Chunks: []*pb.PolicyChunk{
			{Id: "c1", Status: "pending", RuleName: "rule1"},
			{Id: "c2", Status: "approved", RuleName: "rule2"},
		},
		RollingSummary: "Analysis summary",
		DraftVersion:   42,
		LastAnalyzedAtMs: 1700000000000,
	}

	draft := DraftPolicyFromProto(proto)

	require.NotNil(t, draft)
	assert.Len(t, draft.Chunks, 2)
	assert.Equal(t, "c1", draft.Chunks[0].ID)
	assert.Equal(t, "c2", draft.Chunks[1].ID)
	assert.Equal(t, "Analysis summary", draft.RollingSummary)
	assert.Equal(t, uint64(42), draft.DraftVersion)
	assert.False(t, draft.LastAnalyzedAt.IsZero())
}

func TestDraftPolicyFromProto_Nil(t *testing.T) {
	assert.Nil(t, DraftPolicyFromProto(nil))
}

func TestDraftPolicyFromProto_EmptyChunks(t *testing.T) {
	proto := &pb.GetDraftPolicyResponse{
		RollingSummary: "empty",
		DraftVersion:   1,
	}

	draft := DraftPolicyFromProto(proto)
	require.NotNil(t, draft)
	assert.Empty(t, draft.Chunks)
}

// --- SandboxPolicyRevision ---

func TestSandboxPolicyRevisionFromProto(t *testing.T) {
	proto := &pb.SandboxPolicyRevision{
		Version:    3,
		PolicyHash: "sha256:abc123",
		Status:     pb.PolicyStatus_POLICY_STATUS_LOADED,
		LoadError:  "",
		CreatedAtMs: 1700000000000,
		LoadedAtMs:  1700000001000,
	}

	rev := SandboxPolicyRevisionFromProto(proto)

	require.NotNil(t, rev)
	assert.Equal(t, uint32(3), rev.Version)
	assert.Equal(t, "sha256:abc123", rev.PolicyHash)
	assert.Equal(t, v1.PolicyLoadStatusLoaded, rev.Status)
	assert.Empty(t, rev.LoadError)
	assert.False(t, rev.CreatedAt.IsZero())
	assert.False(t, rev.LoadedAt.IsZero())
}

func TestSandboxPolicyRevisionFromProto_Nil(t *testing.T) {
	assert.Nil(t, SandboxPolicyRevisionFromProto(nil))
}

func TestSandboxPolicyRevisionFromProto_WithPolicy(t *testing.T) {
	proto := &pb.SandboxPolicyRevision{
		Version:    1,
		PolicyHash: "sha256:def",
		Status:     pb.PolicyStatus_POLICY_STATUS_LOADED,
		Policy: &sbv1.SandboxPolicy{
			Version: 1,
		},
	}

	rev := SandboxPolicyRevisionFromProto(proto)
	require.NotNil(t, rev)
	assert.NotEmpty(t, rev.Policy, "policy bytes should be populated when proto policy is set")
}

// --- PolicyStatusResult ---

func TestPolicyStatusResultFromProto(t *testing.T) {
	proto := &pb.GetSandboxPolicyStatusResponse{
		Revision: &pb.SandboxPolicyRevision{
			Version:    5,
			PolicyHash: "sha256:xyz",
			Status:     pb.PolicyStatus_POLICY_STATUS_PENDING,
		},
		ActiveVersion: 4,
	}

	result := PolicyStatusResultFromProto(proto)

	require.NotNil(t, result)
	assert.Equal(t, uint32(5), result.Revision.Version)
	assert.Equal(t, "sha256:xyz", result.Revision.PolicyHash)
	assert.Equal(t, v1.PolicyLoadStatusPending, result.Revision.Status)
	assert.Equal(t, uint32(4), result.ActiveVersion)
}

func TestPolicyStatusResultFromProto_Nil(t *testing.T) {
	assert.Nil(t, PolicyStatusResultFromProto(nil))
}

// --- ApproveResult ---

func TestApproveResultFromProto(t *testing.T) {
	proto := &pb.ApproveDraftChunkResponse{
		PolicyVersion: 7,
		PolicyHash:    "sha256:merged",
	}

	result := ApproveResultFromProto(proto)

	require.NotNil(t, result)
	assert.Equal(t, uint32(7), result.PolicyVersion)
	assert.Equal(t, "sha256:merged", result.PolicyHash)
}

func TestApproveResultFromProto_Nil(t *testing.T) {
	assert.Nil(t, ApproveResultFromProto(nil))
}

// --- ApproveAllResult ---

func TestApproveAllResultFromProto(t *testing.T) {
	proto := &pb.ApproveAllDraftChunksResponse{
		PolicyVersion:  8,
		PolicyHash:     "sha256:all",
		ChunksApproved: 10,
		ChunksSkipped:  2,
	}

	result := ApproveAllResultFromProto(proto)

	require.NotNil(t, result)
	assert.Equal(t, uint32(8), result.PolicyVersion)
	assert.Equal(t, "sha256:all", result.PolicyHash)
	assert.Equal(t, uint32(10), result.ChunksApproved)
	assert.Equal(t, uint32(2), result.ChunksSkipped)
}

func TestApproveAllResultFromProto_Nil(t *testing.T) {
	assert.Nil(t, ApproveAllResultFromProto(nil))
}

// --- UndoResult ---

func TestUndoResultFromProto(t *testing.T) {
	proto := &pb.UndoDraftChunkResponse{
		PolicyVersion: 6,
		PolicyHash:    "sha256:reverted",
	}

	result := UndoResultFromProto(proto)

	require.NotNil(t, result)
	assert.Equal(t, uint32(6), result.PolicyVersion)
	assert.Equal(t, "sha256:reverted", result.PolicyHash)
}

func TestUndoResultFromProto_Nil(t *testing.T) {
	assert.Nil(t, UndoResultFromProto(nil))
}

// --- ClearResult ---

func TestClearResultFromProto(t *testing.T) {
	proto := &pb.ClearDraftChunksResponse{
		ChunksCleared: 15,
	}

	result := ClearResultFromProto(proto)

	require.NotNil(t, result)
	assert.Equal(t, uint32(15), result.ChunksCleared)
}

func TestClearResultFromProto_Nil(t *testing.T) {
	assert.Nil(t, ClearResultFromProto(nil))
}

// --- DraftHistoryEntry ---

func TestDraftHistoryEntryFromProto(t *testing.T) {
	proto := &pb.DraftHistoryEntry{
		TimestampMs: 1700000000000,
		EventType:   "approved",
		Description: "Chunk c1 approved",
		ChunkId:     "c1",
	}

	entry := DraftHistoryEntryFromProto(proto)

	require.NotNil(t, entry)
	assert.False(t, entry.Timestamp.IsZero())
	assert.Equal(t, "approved", entry.EventType)
	assert.Equal(t, "Chunk c1 approved", entry.Description)
	assert.Equal(t, "c1", entry.ChunkID)
}

func TestDraftHistoryEntryFromProto_Nil(t *testing.T) {
	assert.Nil(t, DraftHistoryEntryFromProto(nil))
}
