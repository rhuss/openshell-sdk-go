// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"

	"github.com/rhuss/openshell-sdk-go/openshell/v1/internal/converter"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	sbv1 "github.com/rhuss/openshell-sdk-go/proto/sandboxv1"
	"google.golang.org/grpc"
)

type configClient struct {
	client pb.OpenShellClient
}

func newConfigClient(conn grpc.ClientConnInterface) *configClient {
	return &configClient{client: pb.NewOpenShellClient(conn)}
}

func (c *configClient) GetSandbox(ctx context.Context, sandboxID string) (*SandboxConfig, error) {
	resp, err := c.client.GetSandboxConfig(ctx, &sbv1.GetSandboxConfigRequest{
		SandboxId: sandboxID,
	})
	if err != nil {
		return nil, converter.FromGRPCError(err)
	}
	return converter.SandboxConfigFromProto(resp), nil
}

func (c *configClient) GetGateway(ctx context.Context) (*GatewayConfig, error) {
	resp, err := c.client.GetGatewayConfig(ctx, &sbv1.GetGatewayConfigRequest{})
	if err != nil {
		return nil, converter.FromGRPCError(err)
	}
	return converter.GatewayConfigFromProto(resp), nil
}

func (c *configClient) Update(ctx context.Context, update *ConfigUpdate) (*ConfigUpdateResult, error) {
	if update == nil {
		return nil, &StatusError{
			Code:    ErrorInvalidArgument,
			Message: "update must not be nil",
		}
	}
	if update.MergeOperations != nil {
		return nil, &StatusError{
			Code:    ErrorInvalidArgument,
			Message: "MergeOperations is not yet supported; full policy merge support is planned for a future release",
		}
	}

	req, convErr := converter.ConfigUpdateToProto(update)
	if convErr != nil {
		return nil, &StatusError{Code: ErrorInvalidArgument, Message: convErr.Error()}
	}
	resp, err := c.client.UpdateConfig(ctx, req)
	if err != nil {
		return nil, converter.FromGRPCError(err)
	}
	return converter.ConfigUpdateResultFromProto(resp), nil
}
