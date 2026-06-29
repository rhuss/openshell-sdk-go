// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"

	"github.com/rhuss/openshell-sdk-go/openshell/v1/internal/converter"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"google.golang.org/grpc"
)

type sshClient struct {
	client pb.OpenShellClient
}

func newSSHClient(conn grpc.ClientConnInterface) *sshClient {
	return &sshClient{client: pb.NewOpenShellClient(conn)}
}

func (s *sshClient) CreateSession(ctx context.Context, sandboxID string) (*SSHSession, error) {
	resp, err := s.client.CreateSshSession(ctx, &pb.CreateSshSessionRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		return nil, converter.FromGRPCError(err)
	}
	return converter.SSHSessionFromProto(resp), nil
}

func (s *sshClient) RevokeSession(ctx context.Context, token string) (bool, error) {
	resp, err := s.client.RevokeSshSession(ctx, &pb.RevokeSshSessionRequest{
		Token: token,
	})
	if err != nil {
		return false, converter.FromGRPCError(err)
	}
	return resp.GetRevoked(), nil
}
