// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"time"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	dm "github.com/rhuss/openshell-sdk-go/proto/datamodelv1"
)

// ProviderFromProto converts a proto Provider to an SDK Provider.
func ProviderFromProto(p *dm.Provider) *v1.Provider {
	if p == nil {
		return nil
	}

	result := &v1.Provider{
		Type: p.GetType(),
		Spec: v1.ProviderSpec{
			Credentials: p.GetCredentials(),
			Config:      p.GetConfig(),
		},
	}

	if m := p.GetMetadata(); m != nil {
		result.ID = m.GetId()
		result.Name = m.GetName()
		result.CreatedAt = TimeFromMillis(m.GetCreatedAtMs())
		result.Labels = m.GetLabels()
		result.ResourceVersion = m.GetResourceVersion()
	}

	if expires := p.GetCredentialExpiresAtMs(); len(expires) > 0 {
		result.Spec.CredentialExpiresAt = make(map[string]time.Time, len(expires))
		for k, ms := range expires {
			result.Spec.CredentialExpiresAt[k] = TimeFromMillis(ms)
		}
	}

	return result
}

// ProviderToProto converts an SDK Provider to a proto Provider.
func ProviderToProto(p *v1.Provider) *dm.Provider {
	if p == nil {
		return nil
	}

	result := &dm.Provider{
		Metadata: &dm.ObjectMeta{
			Id:              p.ID,
			Name:            p.Name,
			CreatedAtMs:     MillisFromTime(p.CreatedAt),
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
			result.CredentialExpiresAtMs[k] = MillisFromTime(t)
		}
	}

	return result
}
