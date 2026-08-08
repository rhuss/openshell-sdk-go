// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// --- T019: fakeTCPClient stub tests ---

func TestFakeTCP_Forward_ReturnsUnimplemented(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	_, err := c.Forward(context.Background(), "default", "sandbox-1", 8080)
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeTCP_Forward_ClosedReturnsUnavailable(t *testing.T) {
	c := newFakeTCPClient(func() bool { return true })
	_, err := c.Forward(context.Background(), "default", "sandbox-1", 8080)
	require.Error(t, err)
	assert.True(t, types.IsUnavailable(err))
}

func TestFakeTCP_Forward_WithForwardOption(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	_, err := c.Forward(context.Background(), "default", "sandbox-1", 8080, v1.WithForwardServiceID("audit-svc"))
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeTCP_Forward_EmptySandboxName(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	_, err := c.Forward(context.Background(), "default", "", 8080)
	require.Error(t, err)
	assert.True(t, types.IsInvalidArgument(err))
	assert.Contains(t, err.Error(), "sandbox name")
}

func TestFakeTCP_Forward_InvalidPort(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	_, err := c.Forward(context.Background(), "default", "sandbox-1", 0)
	require.Error(t, err)
	assert.True(t, types.IsInvalidArgument(err))
}

// --- T020: fakeTCPClient.Listen tests ---

func TestFakeTCP_Listen_EmptySandboxName(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	ln, err := c.Listen(context.Background(), "default", "", 8080, 0)
	assert.Nil(t, ln)
	require.Error(t, err)
	assert.True(t, types.IsInvalidArgument(err))
}

func TestFakeTCP_Listen_InvalidRemotePort(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })

	tests := []struct {
		name string
		port uint32
	}{
		{"port zero", 0},
		{"port too high", 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ln, err := c.Listen(context.Background(), "default", "my-sandbox", tt.port, 0)
			assert.Nil(t, ln)
			require.Error(t, err)
			assert.True(t, types.IsInvalidArgument(err))
		})
	}
}

func TestFakeTCP_Listen_ValidInputsReturnUnimplemented(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	ln, err := c.Listen(context.Background(), "default", "my-sandbox", 8080, 0)
	assert.Nil(t, ln)
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeTCP_Listen_ClosedReturnsUnavailable(t *testing.T) {
	c := newFakeTCPClient(func() bool { return true })
	ln, err := c.Listen(context.Background(), "default", "my-sandbox", 8080, 0)
	assert.Nil(t, ln)
	require.Error(t, err)
	assert.True(t, types.IsUnavailable(err))
}

func TestFakeTCP_Listen_WithOptions(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	ln, err := c.Listen(context.Background(), "default", "my-sandbox", 8080, 0,
		v1.WithBindAddress("0.0.0.0"),
		v1.WithSSHTunnel(),
		v1.WithListenServiceID("svc-1"),
	)
	assert.Nil(t, ln)
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

// --- RemoteListen tests ---

func TestFakeTCP_RemoteListen_ReturnsUnimplemented(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	err := c.RemoteListen(context.Background(), "default", "my-sandbox", 8080, "localhost:8080")
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeTCP_RemoteListen_ValidationParity(t *testing.T) {
	tests := []struct {
		name        string
		closed      bool
		sandboxName string
		port        uint32
		localTarget string
		checkErr    func(error) bool
		errName     string
	}{
		{
			name:        "closed client",
			closed:      true,
			sandboxName: "my-sandbox",
			port:        8080,
			localTarget: "localhost:8080",
			checkErr:    types.IsUnavailable,
			errName:     "Unavailable",
		},
		{
			name:        "empty sandbox name",
			sandboxName: "",
			port:        8080,
			localTarget: "localhost:8080",
			checkErr:    types.IsInvalidArgument,
			errName:     "InvalidArgument",
		},
		{
			name:        "port zero",
			sandboxName: "my-sandbox",
			port:        0,
			localTarget: "localhost:8080",
			checkErr:    types.IsInvalidArgument,
			errName:     "InvalidArgument",
		},
		{
			name:        "port too high",
			sandboxName: "my-sandbox",
			port:        65536,
			localTarget: "localhost:8080",
			checkErr:    types.IsInvalidArgument,
			errName:     "InvalidArgument",
		},
		{
			name:        "malformed target missing port",
			sandboxName: "my-sandbox",
			port:        8080,
			localTarget: "localhost",
			checkErr:    types.IsInvalidArgument,
			errName:     "InvalidArgument",
		},
		{
			name:        "malformed target empty",
			sandboxName: "my-sandbox",
			port:        8080,
			localTarget: "",
			checkErr:    types.IsInvalidArgument,
			errName:     "InvalidArgument",
		},
		{
			name:        "malformed target bare IPv6",
			sandboxName: "my-sandbox",
			port:        8080,
			localTarget: "::1",
			checkErr:    types.IsInvalidArgument,
			errName:     "InvalidArgument",
		},
		{
			name:        "boundary port 1",
			sandboxName: "my-sandbox",
			port:        1,
			localTarget: "localhost:8080",
			checkErr:    types.IsUnimplemented,
			errName:     "Unimplemented",
		},
		{
			name:        "boundary port 65535",
			sandboxName: "my-sandbox",
			port:        65535,
			localTarget: "localhost:8080",
			checkErr:    types.IsUnimplemented,
			errName:     "Unimplemented",
		},
		{
			name:        "ipv6 target",
			sandboxName: "my-sandbox",
			port:        8080,
			localTarget: "[::1]:8080",
			checkErr:    types.IsUnimplemented,
			errName:     "Unimplemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newFakeTCPClient(func() bool { return tt.closed })
			err := c.RemoteListen(context.Background(), "default", tt.sandboxName, tt.port, tt.localTarget)
			require.Error(t, err)
			assert.True(t, tt.checkErr(err), "expected %s, got: %v", tt.errName, err)
		})
	}
}

func TestFakeTCP_RemoteListen_WithOptions(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	err := c.RemoteListen(context.Background(), "default", "my-sandbox", 8080, "localhost:8080",
		v1.WithRemoteBindAddress("0.0.0.0"),
		v1.WithRemoteListenServiceID("mcp-proxy"),
	)
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}
