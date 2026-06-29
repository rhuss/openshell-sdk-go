// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"net"
	"sync"
	"testing"

	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

// --- Mock server for SSH sessions ---

type mockSSHServer struct {
	pb.UnimplementedOpenShellServer
	mu         sync.Mutex
	sessions   map[string]*pb.CreateSshSessionResponse // key: sandbox ID
	tokens     map[string]bool                         // track active tokens
	createErr  error
	revokeErr  error
	nextToken  string // override token for testing
}

func newMockSSHServer() *mockSSHServer {
	return &mockSSHServer{
		sessions: make(map[string]*pb.CreateSshSessionResponse),
		tokens:   make(map[string]bool),
	}
}

func (s *mockSSHServer) CreateSshSession(_ context.Context, req *pb.CreateSshSessionRequest) (*pb.CreateSshSessionResponse, error) { //nolint:revive // proto-generated method name
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.createErr != nil {
		return nil, s.createErr
	}

	token := "tok-" + req.GetSandboxId()
	if s.nextToken != "" {
		token = s.nextToken
	}

	resp := &pb.CreateSshSessionResponse{
		SandboxId:          req.GetSandboxId(),
		Token:              token,
		GatewayHost:        "gw.example.com",
		GatewayPort:        2222,
		GatewayScheme:      "https",
		HostKeyFingerprint: "SHA256:abc123",
		ExpiresAtMs:        1700000000000,
	}
	s.sessions[req.GetSandboxId()] = resp
	s.tokens[token] = true
	return resp, nil
}

func (s *mockSSHServer) RevokeSshSession(_ context.Context, req *pb.RevokeSshSessionRequest) (*pb.RevokeSshSessionResponse, error) { //nolint:revive // proto-generated method name
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.revokeErr != nil {
		return nil, s.revokeErr
	}

	token := req.GetToken()
	active, exists := s.tokens[token]
	if exists && active {
		s.tokens[token] = false
		return &pb.RevokeSshSessionResponse{Revoked: true}, nil
	}
	// Already revoked or not found — not an error, just revoked=false.
	return &pb.RevokeSshSessionResponse{Revoked: false}, nil
}

// --- Test setup ---

func setupSSHTest(t *testing.T, mock *mockSSHServer) (*sshClient, func()) {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterOpenShellServer(srv, mock)
	go func() { _ = srv.Serve(lis) }()

	conn, err := grpc.NewClient("passthrough:///bufconn",
		grpc.WithContextDialer(func(_ context.Context, _ string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	return newSSHClient(conn), func() {
		_ = conn.Close()
		srv.Stop()
	}
}

// --- Tests ---

func TestSSHCreateSession(t *testing.T) {
	mock := newMockSSHServer()
	client, cleanup := setupSSHTest(t, mock)
	defer cleanup()

	session, err := client.CreateSession(context.Background(), "my-sandbox")

	require.NoError(t, err)
	require.NotNil(t, session)
	assert.Equal(t, "my-sandbox", session.SandboxID)
	assert.Equal(t, "tok-my-sandbox", session.Token)
	assert.Equal(t, "gw.example.com", session.GatewayHost)
	assert.Equal(t, uint32(2222), session.GatewayPort)
	assert.Equal(t, "https", session.GatewayScheme)
	assert.Equal(t, "SHA256:abc123", session.HostKeyFingerprint)
	assert.Equal(t, int64(1700000000000), session.ExpiresAtMs)
}

func TestSSHCreateSession_Error(t *testing.T) {
	mock := newMockSSHServer()
	mock.createErr = status.Errorf(codes.NotFound, "sandbox not found")
	client, cleanup := setupSSHTest(t, mock)
	defer cleanup()

	session, err := client.CreateSession(context.Background(), "missing")

	assert.Nil(t, session)
	require.Error(t, err)
	assert.True(t, IsNotFound(err))
}

func TestSSHRevokeSession(t *testing.T) {
	mock := newMockSSHServer()
	client, cleanup := setupSSHTest(t, mock)
	defer cleanup()

	// Create a session first.
	session, err := client.CreateSession(context.Background(), "my-sandbox")
	require.NoError(t, err)

	// Revoke it — should return true.
	revoked, err := client.RevokeSession(context.Background(), session.Token)

	require.NoError(t, err)
	assert.True(t, revoked)
}

func TestSSHRevokeSession_AlreadyRevoked(t *testing.T) {
	mock := newMockSSHServer()
	client, cleanup := setupSSHTest(t, mock)
	defer cleanup()

	// Create and revoke.
	session, err := client.CreateSession(context.Background(), "my-sandbox")
	require.NoError(t, err)
	_, err = client.RevokeSession(context.Background(), session.Token)
	require.NoError(t, err)

	// Revoke again — should return false (already revoked).
	revoked, err := client.RevokeSession(context.Background(), session.Token)

	require.NoError(t, err)
	assert.False(t, revoked)
}

func TestSSHRevokeSession_Error(t *testing.T) {
	mock := newMockSSHServer()
	mock.revokeErr = status.Errorf(codes.Internal, "internal error")
	client, cleanup := setupSSHTest(t, mock)
	defer cleanup()

	revoked, err := client.RevokeSession(context.Background(), "some-token")

	assert.False(t, revoked)
	require.Error(t, err)
	var se *StatusError
	require.ErrorAs(t, err, &se)
	assert.Equal(t, ErrorInternal, se.Code)
}
