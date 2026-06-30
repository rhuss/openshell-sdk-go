// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"io"
)

// forwardConfig accumulates options for the Forward method.
type forwardConfig struct {
	serviceID string
}

// ForwardOption configures a TCP forward opened via [TCPInterface.Forward].
type ForwardOption func(*forwardConfig)

// WithForwardServiceID sets an optional service identifier on the forward's
// init frame for audit and correlation purposes.
func WithForwardServiceID(id string) ForwardOption {
	return func(c *forwardConfig) {
		c.serviceID = id
	}
}

// TCPInterface defines operations for TCP port forwarding to sandboxes.
// Methods accept a sandbox name and resolve it to an ID internally.
type TCPInterface interface {
	// Forward opens a bidirectional TCP connection to the given port inside a
	// sandbox. The sandbox is identified by name; the SDK resolves it to an
	// ID internally. The returned io.ReadWriteCloser wraps the underlying
	// gRPC stream; closing it terminates the stream. Port must be in the
	// range 1-65535; out-of-range values are rejected client-side with an
	// InvalidArgument error before opening the gRPC stream.
	//
	// The connection respects context cancellation: if ctx is cancelled,
	// the stream is closed and pending Read/Write calls return a context error.
	Forward(ctx context.Context, sandboxName string, port uint32, opts ...ForwardOption) (io.ReadWriteCloser, error)
}
