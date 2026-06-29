// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"

	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// SSHSession represents an SSH session created for a sandbox.
type SSHSession = types.SSHSession

// SSHInterface defines operations for managing SSH sessions.
type SSHInterface interface {
	// CreateSession creates a new SSH session for the given sandbox.
	// The returned SSHSession contains connection details including the
	// sensitive Token field that must not be logged.
	CreateSession(ctx context.Context, sandboxID string) (*SSHSession, error)
	// RevokeSession revokes an existing SSH session by its token.
	// Returns true if the session was actively revoked, false if it was
	// already expired or not found.
	RevokeSession(ctx context.Context, token string) (bool, error)
}
