// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"fmt"
	"os"

	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"google.golang.org/grpc"
)

type fileClient struct {
	client    pb.OpenShellClient
	transport sshTransport
}

type sshTransport interface {
	upload(ctx context.Context, session *pb.CreateSshSessionResponse, localPath, remotePath string) error
	download(ctx context.Context, session *pb.CreateSshSessionResponse, remotePath, localPath string) error
}

func newFileClient(conn grpc.ClientConnInterface) *fileClient {
	return &fileClient{
		client:    pb.NewOpenShellClient(conn),
		transport: &defaultSSHTransport{},
	}
}

func (f *fileClient) Upload(ctx context.Context, sandboxID string, localPath string, remotePath string) error {
	if sandboxID == "" {
		return fmt.Errorf("sandbox ID must not be empty")
	}
	if remotePath == "" {
		return fmt.Errorf("remote path must not be empty")
	}

	info, err := os.Stat(localPath)
	if err != nil {
		return fmt.Errorf("local file error: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("local path is a directory, not a file: %s", localPath)
	}

	session, err := f.client.CreateSshSession(ctx, &pb.CreateSshSessionRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		return fromGRPCError(err)
	}

	defer func() {
		_, _ = f.client.RevokeSshSession(ctx, &pb.RevokeSshSessionRequest{
			Token: session.GetToken(),
		})
	}()

	return f.transport.upload(ctx, session, localPath, remotePath)
}

func (f *fileClient) Download(ctx context.Context, sandboxID string, remotePath string, localPath string) error {
	if sandboxID == "" {
		return fmt.Errorf("sandbox ID must not be empty")
	}
	if remotePath == "" {
		return fmt.Errorf("remote path must not be empty")
	}

	session, err := f.client.CreateSshSession(ctx, &pb.CreateSshSessionRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		return fromGRPCError(err)
	}

	defer func() {
		_, _ = f.client.RevokeSshSession(ctx, &pb.RevokeSshSessionRequest{
			Token: session.GetToken(),
		})
	}()

	return f.transport.download(ctx, session, remotePath, localPath)
}

type defaultSSHTransport struct{}

func (t *defaultSSHTransport) upload(_ context.Context, session *pb.CreateSshSessionResponse, localPath, remotePath string) error {
	return fmt.Errorf("SSH transport to %s:%d not implemented (local: %s -> remote: %s)",
		session.GetGatewayHost(), session.GetGatewayPort(), localPath, remotePath)
}

func (t *defaultSSHTransport) download(_ context.Context, session *pb.CreateSshSessionResponse, remotePath, localPath string) error {
	return fmt.Errorf("SSH transport to %s:%d not implemented (remote: %s -> local: %s)",
		session.GetGatewayHost(), session.GetGatewayPort(), remotePath, localPath)
}
