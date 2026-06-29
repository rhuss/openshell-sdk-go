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
	_, err := c.Forward(context.Background(), "sandbox-1", 8080)
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeTCP_Forward_ClosedReturnsUnavailable(t *testing.T) {
	c := newFakeTCPClient(func() bool { return true })
	_, err := c.Forward(context.Background(), "sandbox-1", 8080)
	require.Error(t, err)
	assert.True(t, types.IsUnavailable(err))
}

func TestFakeTCP_Forward_WithForwardOption(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	_, err := c.Forward(context.Background(), "sandbox-1", 8080, v1.WithForwardServiceID("audit-svc"))
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeTCP_Forward_InvalidPort(t *testing.T) {
	c := newFakeTCPClient(func() bool { return false })
	_, err := c.Forward(context.Background(), "sandbox-1", 0)
	require.Error(t, err)
	assert.True(t, types.IsInvalidArgument(err))
}
