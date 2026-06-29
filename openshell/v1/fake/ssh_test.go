// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// --- T018: fakeSSHClient stub tests ---

func TestFakeSSH_CreateSession_ReturnsUnimplemented(t *testing.T) {
	c := newFakeSSHClient(func() bool { return false })
	_, err := c.CreateSession(context.Background(), "sandbox-1")
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeSSH_RevokeSession_ReturnsUnimplemented(t *testing.T) {
	c := newFakeSSHClient(func() bool { return false })
	_, err := c.RevokeSession(context.Background(), "tok-abc")
	require.Error(t, err)
	assert.True(t, types.IsUnimplemented(err))
}

func TestFakeSSH_CreateSession_ClosedReturnsUnavailable(t *testing.T) {
	c := newFakeSSHClient(func() bool { return true })
	_, err := c.CreateSession(context.Background(), "sandbox-1")
	require.Error(t, err)
	assert.True(t, types.IsUnavailable(err))
}

func TestFakeSSH_RevokeSession_ClosedReturnsUnavailable(t *testing.T) {
	c := newFakeSSHClient(func() bool { return true })
	_, err := c.RevokeSession(context.Background(), "tok-abc")
	require.Error(t, err)
	assert.True(t, types.IsUnavailable(err))
}
