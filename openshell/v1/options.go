// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import "time"

// CreateOptions configures resource creation.
type CreateOptions struct{}

// GetOptions configures resource retrieval.
type GetOptions struct{}

// ListOptions configures resource listing with pagination and filtering.
type ListOptions struct {
	Limit         int
	Offset        int
	LabelSelector string
}

// DeleteOptions configures resource deletion.
type DeleteOptions struct{}

// UpdateOptions configures resource updates.
type UpdateOptions struct{}

// WatchOptions configures watch behavior.
type WatchOptions struct {
	TimeoutSeconds int64
	LabelSelector  string
}

// WaitOptions configures wait behavior. Use context for timeout control.
type WaitOptions struct {
	PollInterval time.Duration
}

// ExecOptions configures command execution.
type ExecOptions struct {
	Env     map[string]string
	WorkDir string
}
