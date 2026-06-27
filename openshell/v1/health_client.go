// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"

	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"google.golang.org/grpc"
)

type healthClient struct {
	client pb.OpenShellClient
}

func newHealthClient(conn grpc.ClientConnInterface) *healthClient {
	return &healthClient{client: pb.NewOpenShellClient(conn)}
}

func (h *healthClient) Check(ctx context.Context) (*HealthResult, error) {
	resp, err := h.client.Health(ctx, &pb.HealthRequest{})
	if err != nil {
		return nil, fromGRPCError(err)
	}

	return &HealthResult{
		Healthy: resp.GetStatus() == pb.ServiceStatus_SERVICE_STATUS_HEALTHY,
		Version: resp.GetVersion(),
	}, nil
}
