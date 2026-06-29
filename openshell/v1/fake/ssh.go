// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package fake

import (
	"context"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// fakeSSHClient implements v1.SSHInterface. All methods return
// Unimplemented because SSH session management requires a real gateway.
type fakeSSHClient struct {
	closedFunc func() bool
}

// newFakeSSHClient creates a new fakeSSHClient.
func newFakeSSHClient(closedFunc func() bool) *fakeSSHClient {
	return &fakeSSHClient{closedFunc: closedFunc}
}

// CreateSession returns Unimplemented.
func (c *fakeSSHClient) CreateSession(_ context.Context, _ string) (*types.SSHSession, error) {
	if c.closedFunc() {
		return nil, &types.StatusError{Code: types.ErrorUnavailable, Message: "client is closed"}
	}
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "CreateSession is not supported by the fake client"}
}

// RevokeSession returns Unimplemented.
func (c *fakeSSHClient) RevokeSession(_ context.Context, _ string) (bool, error) {
	if c.closedFunc() {
		return false, &types.StatusError{Code: types.ErrorUnavailable, Message: "client is closed"}
	}
	return false, &types.StatusError{Code: types.ErrorUnimplemented, Message: "RevokeSession is not supported by the fake client"}
}

// Compile-time check that fakeSSHClient implements v1.SSHInterface.
var _ v1.SSHInterface = (*fakeSSHClient)(nil)
