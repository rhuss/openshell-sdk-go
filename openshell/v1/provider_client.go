// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"time"

	dm "github.com/rhuss/openshell-sdk-go/proto/datamodelv1"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"google.golang.org/grpc"
)

type providerClient struct {
	client pb.OpenShellClient
}

func newProviderClient(conn grpc.ClientConnInterface) *providerClient {
	return &providerClient{client: pb.NewOpenShellClient(conn)}
}

func (p *providerClient) Create(ctx context.Context, provider *Provider) (*Provider, error) {
	resp, err := p.client.CreateProvider(ctx, &pb.CreateProviderRequest{
		Provider: providerToProto(provider),
	})
	if err != nil {
		return nil, fromGRPCError(err)
	}
	return providerFromProto(resp.GetProvider()), nil
}

func (p *providerClient) Get(ctx context.Context, name string) (*Provider, error) {
	resp, err := p.client.GetProvider(ctx, &pb.GetProviderRequest{
		Name: name,
	})
	if err != nil {
		return nil, fromGRPCError(err)
	}
	return providerFromProto(resp.GetProvider()), nil
}

func (p *providerClient) List(ctx context.Context, opts ...ListOptions) ([]*Provider, error) {
	req := &pb.ListProvidersRequest{}
	if len(opts) > 0 {
		req.Limit = uint32(opts[0].Limit)
		req.Offset = uint32(opts[0].Offset)
	}

	resp, err := p.client.ListProviders(ctx, req)
	if err != nil {
		return nil, fromGRPCError(err)
	}

	providers := make([]*Provider, 0, len(resp.GetProviders()))
	for _, proto := range resp.GetProviders() {
		providers = append(providers, providerFromProto(proto))
	}
	return providers, nil
}

func (p *providerClient) Update(ctx context.Context, provider *Provider) (*Provider, error) {
	proto := providerToProto(provider)
	req := &pb.UpdateProviderRequest{
		Provider: proto,
	}
	if proto != nil {
		req.CredentialExpiresAtMs = proto.CredentialExpiresAtMs
	}

	resp, err := p.client.UpdateProvider(ctx, req)
	if err != nil {
		return nil, fromGRPCError(err)
	}
	return providerFromProto(resp.GetProvider()), nil
}

func (p *providerClient) Delete(ctx context.Context, name string) error {
	_, err := p.client.DeleteProvider(ctx, &pb.DeleteProviderRequest{
		Name: name,
	})
	if err != nil {
		return fromGRPCError(err)
	}
	return nil
}

func (p *providerClient) Ensure(ctx context.Context, provider *Provider) (*Provider, error) {
	if provider == nil {
		return nil, &StatusError{Code: ErrorInvalidArgument, Message: "provider must not be nil"}
	}
	existing, err := p.Get(ctx, provider.Name)
	if err != nil {
		if !IsNotFound(err) {
			return nil, err
		}
		return p.Create(ctx, provider)
	}

	updated := *provider
	updated.ID = existing.ID
	updated.ResourceVersion = existing.ResourceVersion
	return p.Update(ctx, &updated)
}

func providerFromProto(p *dm.Provider) *Provider {
	if p == nil {
		return nil
	}

	result := &Provider{
		Type: p.GetType(),
		Spec: ProviderSpec{
			Config: copyStringMap(p.GetConfig()),
		},
	}

	if m := p.GetMetadata(); m != nil {
		result.ID = m.GetId()
		result.Name = m.GetName()
		result.CreatedAt = timeFromMillis(m.GetCreatedAtMs())
		result.Labels = copyStringMap(m.GetLabels())
		result.ResourceVersion = m.GetResourceVersion()
	}

	if expires := p.GetCredentialExpiresAtMs(); len(expires) > 0 {
		result.Spec.CredentialExpiresAt = make(map[string]time.Time, len(expires))
		for k, ms := range expires {
			result.Spec.CredentialExpiresAt[k] = timeFromMillis(ms)
		}
	}

	return result
}

func providerToProto(p *Provider) *dm.Provider {
	if p == nil {
		return nil
	}

	result := &dm.Provider{
		Metadata: &dm.ObjectMeta{
			Id:              p.ID,
			Name:            p.Name,
			CreatedAtMs:     millisFromTime(p.CreatedAt),
			Labels:          p.Labels,
			ResourceVersion: p.ResourceVersion,
		},
		Type:        p.Type,
		Credentials: p.Spec.Credentials,
		Config:      p.Spec.Config,
	}

	if len(p.Spec.CredentialExpiresAt) > 0 {
		result.CredentialExpiresAtMs = make(map[string]int64, len(p.Spec.CredentialExpiresAt))
		for k, t := range p.Spec.CredentialExpiresAt {
			result.CredentialExpiresAtMs[k] = millisFromTime(t)
		}
	}

	return result
}

func timeFromMillis(ms int64) time.Time {
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}

func millisFromTime(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixMilli()
}
