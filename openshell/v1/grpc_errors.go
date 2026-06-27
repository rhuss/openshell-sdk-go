// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var grpcToSDK = map[codes.Code]ErrorCode{
	codes.NotFound:         ErrorNotFound,
	codes.AlreadyExists:    ErrorAlreadyExists,
	codes.Unavailable:      ErrorUnavailable,
	codes.PermissionDenied: ErrorPermissionDenied,
	codes.InvalidArgument:  ErrorInvalidArgument,
	codes.DeadlineExceeded: ErrorDeadlineExceeded,
	codes.Canceled:         ErrorCancelled,
	codes.Internal:         ErrorInternal,
}

func fromGRPCError(err error) error {
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
		code = ErrorInternal
	}

	return &StatusError{
		Code:    code,
		Message: st.Message(),
	}
}
