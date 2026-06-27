// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import "context"

// ExecResult holds the collected output of a completed command execution.
type ExecResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
}

// ExecChunk represents a single chunk of output from a streaming command execution.
type ExecChunk struct {
	Stream StreamType
	Data   []byte
}

// ExecStream provides an iterator interface over streaming command output.
type ExecStream interface {
	Next() (*ExecChunk, error)
	ExitCode() (int, error)
	Close() error
}

// InteractiveSession provides bidirectional I/O for interactive command execution.
type InteractiveSession interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Resize(cols, rows uint32) error
	ExitCode() (int, error)
	Close() error
}

// ExecInterface defines command execution operations on sandboxes.
type ExecInterface interface {
	Run(ctx context.Context, sandboxID string, command []string, opts ...ExecOptions) (*ExecResult, error)
	Stream(ctx context.Context, sandboxID string, command []string, opts ...ExecOptions) (ExecStream, error)
	Interactive(ctx context.Context, sandboxID string, command []string, cols, rows uint32, opts ...ExecOptions) (InteractiveSession, error)
}
