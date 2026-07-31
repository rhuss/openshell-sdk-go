// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package contracts documents the public API surface added by this feature.
// This file is a design artifact, not compiled code.
package contracts

import "context"

// WorkspaceInterface defines workspace and member management operations.
// Added to ClientInterface as Workspaces() accessor.
type WorkspaceInterface interface {
	// Create creates a workspace with the given name and optional labels.
	// Returns AlreadyExists if a workspace with the same name exists.
	// Returns InvalidArgument if name is empty.
	Create(ctx context.Context, name string, labels map[string]string) (*Workspace, error)

	// Get retrieves a workspace by name.
	// Returns NotFound if no workspace with the given name exists.
	// Returns InvalidArgument if name is empty.
	Get(ctx context.Context, name string) (*Workspace, error)

	// List returns workspaces visible to the caller, with optional pagination and label filtering.
	List(ctx context.Context, opts ...ListOptions) ([]*Workspace, error)

	// Delete removes a workspace by name.
	// Returns NotFound if no workspace with the given name exists.
	// Returns InvalidArgument if name is empty.
	Delete(ctx context.Context, name string) error

	// AddMember adds a principal to a workspace with the given role.
	// Returns AlreadyExists if the principal is already a member.
	// Returns InvalidArgument if workspace, principalSubject is empty, or role is invalid.
	AddMember(ctx context.Context, workspace, principalSubject string, role WorkspaceRole) (*WorkspaceMember, error)

	// RemoveMember removes a principal from a workspace.
	// Returns NotFound if the principal is not a member.
	// Returns InvalidArgument if workspace or principalSubject is empty.
	RemoveMember(ctx context.Context, workspace, principalSubject string) error

	// ListMembers returns members of a workspace with optional pagination.
	// Returns InvalidArgument if workspace is empty.
	ListMembers(ctx context.Context, workspace string, opts ...ListOptions) ([]*WorkspaceMember, error)
}

// HealthInterface extension: two new methods added to existing interface.
//
//	GetGatewayInfo(ctx context.Context) (*GatewayInfo, error)
//	GetCurrentUser(ctx context.Context) (*CurrentUser, error)

// ClientInterface extension: one new accessor.
//
//	Workspaces() WorkspaceInterface

// Workspace is a placeholder (see data-model.md).
type Workspace struct{}

// WorkspaceMember is a placeholder (see data-model.md).
type WorkspaceMember struct{}

// WorkspaceRole is a placeholder (see data-model.md).
type WorkspaceRole string

// GatewayInfo is a placeholder (see data-model.md).
type GatewayInfo struct{}

// CurrentUser is a placeholder (see data-model.md).
type CurrentUser struct{}

// ListOptions is a placeholder (see data-model.md).
type ListOptions struct{}
