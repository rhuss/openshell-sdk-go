// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"

	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// Sandbox represents a sandbox instance.
type Sandbox = types.Sandbox

// SandboxSpec holds the desired state of a sandbox.
type SandboxSpec = types.SandboxSpec

// SandboxTemplate defines the container template for a sandbox.
type SandboxTemplate = types.SandboxTemplate

// SandboxStatus holds the observed state of a sandbox.
type SandboxStatus = types.SandboxStatus

// SandboxCondition describes an observed condition of a sandbox.
type SandboxCondition = types.SandboxCondition

// AttachProviderResult holds the result of attaching a provider to a sandbox.
type AttachProviderResult = types.AttachProviderResult

// DetachProviderResult holds the result of detaching a provider from a sandbox.
type DetachProviderResult = types.DetachProviderResult

// SandboxInterface defines lifecycle operations on sandboxes.
type SandboxInterface interface {
	Create(ctx context.Context, name string, spec *SandboxSpec, labels map[string]string) (*Sandbox, error)
	Get(ctx context.Context, name string) (*Sandbox, error)
	List(ctx context.Context, opts ...ListOptions) ([]*Sandbox, error)
	Delete(ctx context.Context, name string) error
	AttachProvider(ctx context.Context, sandboxName, providerName string, expectedResourceVersion uint64) (*AttachProviderResult, error)
	DetachProvider(ctx context.Context, sandboxName, providerName string, expectedResourceVersion uint64) (*DetachProviderResult, error)
	ListProviders(ctx context.Context, sandboxName string) ([]*Provider, error)
	WaitReady(ctx context.Context, name string, opts ...WaitOptions) (*Sandbox, error)
	Watch(ctx context.Context, name string, opts ...WatchOptions) (WatchInterface[*Sandbox], error)
}
