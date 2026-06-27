// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"time"
)

// Provider represents an AI provider registration.
type Provider struct {
	ID              string
	Name            string
	Type            string
	CreatedAt       time.Time
	Labels          map[string]string
	ResourceVersion uint64
	Spec            ProviderSpec
}

// ProviderSpec holds provider-specific configuration and credentials.
type ProviderSpec struct {
	Credentials         map[string]string
	Config              map[string]string
	CredentialExpiresAt  map[string]time.Time
}

// ProviderInterface defines CRUD and Ensure operations on providers.
type ProviderInterface interface {
	Create(ctx context.Context, provider *Provider) (*Provider, error)
	Get(ctx context.Context, name string) (*Provider, error)
	List(ctx context.Context, opts ...ListOptions) ([]*Provider, error)
	Update(ctx context.Context, provider *Provider) (*Provider, error)
	Delete(ctx context.Context, name string) error
	Ensure(ctx context.Context, provider *Provider) (*Provider, error)
}
