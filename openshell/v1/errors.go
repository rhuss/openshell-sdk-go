// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// ErrorCode classifies SDK errors by their gRPC origin.
type ErrorCode = types.ErrorCode

// ErrorCode values for classifying gRPC errors.
const (
	ErrorNotFound         = types.ErrorNotFound
	ErrorAlreadyExists    = types.ErrorAlreadyExists
	ErrorUnavailable      = types.ErrorUnavailable
	ErrorPermissionDenied = types.ErrorPermissionDenied
	ErrorInvalidArgument  = types.ErrorInvalidArgument
	ErrorDeadlineExceeded = types.ErrorDeadlineExceeded
	ErrorCancelled        = types.ErrorCancelled
	ErrorInternal         = types.ErrorInternal
)

// StatusError is the typed error returned by all SDK operations.
type StatusError = types.StatusError

// IsNotFound returns true if the error indicates a resource was not found.
var IsNotFound = types.IsNotFound

// IsAlreadyExists returns true if the error indicates a resource already exists.
var IsAlreadyExists = types.IsAlreadyExists

// IsUnavailable returns true if the error indicates the service is unavailable.
var IsUnavailable = types.IsUnavailable

// IsPermissionDenied returns true if the error indicates insufficient permissions.
var IsPermissionDenied = types.IsPermissionDenied

// IsInvalidArgument returns true if the error indicates an invalid argument.
var IsInvalidArgument = types.IsInvalidArgument

// IsDeadlineExceeded returns true if the error indicates a deadline was exceeded.
var IsDeadlineExceeded = types.IsDeadlineExceeded

// IsCancelled returns true if the error indicates the operation was cancelled.
var IsCancelled = types.IsCancelled
