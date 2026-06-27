// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package converter maps between gRPC/proto types and SDK domain types.
package converter

import (
	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var grpcToSDK = map[codes.Code]v1.ErrorCode{
	codes.NotFound:         v1.ErrorNotFound,
	codes.AlreadyExists:    v1.ErrorAlreadyExists,
	codes.Unavailable:      v1.ErrorUnavailable,
	codes.PermissionDenied: v1.ErrorPermissionDenied,
	codes.InvalidArgument:  v1.ErrorInvalidArgument,
	codes.DeadlineExceeded: v1.ErrorDeadlineExceeded,
	codes.Canceled:         v1.ErrorCancelled,
	codes.Internal:         v1.ErrorInternal,
}

// FromGRPCError converts a gRPC error to a typed StatusError.
// Returns nil for nil errors and OK status. Non-gRPC errors pass through unchanged.
func FromGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	if st.Code() == codes.OK {
		return nil
	}

	code, mapped := grpcToSDK[st.Code()]
	if !mapped {
		code = v1.ErrorInternal
	}

	return &v1.StatusError{
		Code:    code,
		Message: st.Message(),
	}
}
