// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"context"
	"fmt"
	"time"

	dm "github.com/rhuss/openshell-sdk-go/proto/datamodelv1"
	pb "github.com/rhuss/openshell-sdk-go/proto/openshellv1"
	"google.golang.org/grpc"
)

const defaultPollInterval = 500 * time.Millisecond

type sandboxClient struct {
	client pb.OpenShellClient
}

func newSandboxClient(conn grpc.ClientConnInterface) *sandboxClient {
	return &sandboxClient{client: pb.NewOpenShellClient(conn)}
}

func (s *sandboxClient) Create(ctx context.Context, name string, spec *SandboxSpec, labels map[string]string) (*Sandbox, error) {
	resp, err := s.client.CreateSandbox(ctx, &pb.CreateSandboxRequest{
		Name:   name,
		Spec:   sandboxSpecToProto(spec),
		Labels: labels,
	})
	if err != nil {
		return nil, fromGRPCError(err)
	}
	return sandboxFromProto(resp.GetSandbox()), nil
}

func (s *sandboxClient) Get(ctx context.Context, name string) (*Sandbox, error) {
	resp, err := s.client.GetSandbox(ctx, &pb.GetSandboxRequest{
		Name: name,
	})
	if err != nil {
		return nil, fromGRPCError(err)
	}
	return sandboxFromProto(resp.GetSandbox()), nil
}

func (s *sandboxClient) List(ctx context.Context, opts ...ListOptions) ([]*Sandbox, error) {
	req := &pb.ListSandboxesRequest{}
	if len(opts) > 0 {
		req.Limit = uint32(opts[0].Limit)
		req.Offset = uint32(opts[0].Offset)
	}

	resp, err := s.client.ListSandboxes(ctx, req)
	if err != nil {
		return nil, fromGRPCError(err)
	}

	sandboxes := make([]*Sandbox, 0, len(resp.GetSandboxes()))
	for _, proto := range resp.GetSandboxes() {
		sandboxes = append(sandboxes, sandboxFromProto(proto))
	}
	return sandboxes, nil
}

func (s *sandboxClient) Delete(ctx context.Context, name string) error {
	_, err := s.client.DeleteSandbox(ctx, &pb.DeleteSandboxRequest{
		Name: name,
	})
	if err != nil {
		return fromGRPCError(err)
	}
	return nil
}

func (s *sandboxClient) AttachProvider(ctx context.Context, sandboxName, providerName string, expectedResourceVersion uint64) (*AttachProviderResult, error) {
	resp, err := s.client.AttachSandboxProvider(ctx, &pb.AttachSandboxProviderRequest{
		SandboxName:             sandboxName,
		ProviderName:            providerName,
		ExpectedResourceVersion: expectedResourceVersion,
	})
	if err != nil {
		return nil, fromGRPCError(err)
	}
	return &AttachProviderResult{
		Sandbox:  sandboxFromProto(resp.GetSandbox()),
		Attached: resp.GetAttached(),
	}, nil
}

func (s *sandboxClient) DetachProvider(ctx context.Context, sandboxName, providerName string, expectedResourceVersion uint64) (*DetachProviderResult, error) {
	resp, err := s.client.DetachSandboxProvider(ctx, &pb.DetachSandboxProviderRequest{
		SandboxName:             sandboxName,
		ProviderName:            providerName,
		ExpectedResourceVersion: expectedResourceVersion,
	})
	if err != nil {
		return nil, fromGRPCError(err)
	}
	return &DetachProviderResult{
		Sandbox:  sandboxFromProto(resp.GetSandbox()),
		Detached: resp.GetDetached(),
	}, nil
}

func (s *sandboxClient) ListProviders(ctx context.Context, sandboxName string) ([]*Provider, error) {
	resp, err := s.client.ListSandboxProviders(ctx, &pb.ListSandboxProvidersRequest{
		SandboxName: sandboxName,
	})
	if err != nil {
		return nil, fromGRPCError(err)
	}

	providers := make([]*Provider, 0, len(resp.GetProviders()))
	for _, proto := range resp.GetProviders() {
		providers = append(providers, providerFromProto(proto))
	}
	return providers, nil
}

func (s *sandboxClient) WaitReady(ctx context.Context, name string, opts ...WaitOptions) (*Sandbox, error) {
	interval := defaultPollInterval
	if len(opts) > 0 && opts[0].PollInterval > 0 {
		interval = opts[0].PollInterval
	}

	sb, err := s.Get(ctx, name)
	if err != nil {
		return nil, err
	}

	if sb.Status.Phase == SandboxReady {
		return sb, nil
	}
	if sb.Status.Phase == SandboxError {
		return nil, fmt.Errorf("sandbox %q is in error state", name)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			sb, err = s.Get(ctx, name)
			if err != nil {
				return nil, err
			}
			if sb.Status.Phase == SandboxReady {
				return sb, nil
			}
			if sb.Status.Phase == SandboxError {
				return nil, fmt.Errorf("sandbox %q is in error state", name)
			}
		}
	}
}

// Package-level converter functions duplicated from internal/converter
// to avoid circular imports (converter imports v1).

func sandboxFromProto(s *pb.Sandbox) *Sandbox {
	if s == nil {
		return nil
	}

	result := &Sandbox{}

	if m := s.GetMetadata(); m != nil {
		result.ID = m.GetId()
		result.Name = m.GetName()
		result.CreatedAt = timeFromMillis(m.GetCreatedAtMs())
		result.Labels = m.GetLabels()
		result.ResourceVersion = m.GetResourceVersion()
	}

	if spec := s.GetSpec(); spec != nil {
		result.Spec = sandboxSpecFromProto(spec)
	}

	if status := s.GetStatus(); status != nil {
		result.Status = sandboxStatusFromProto(status)
	} else {
		result.Status.Phase = SandboxUnknown
	}

	return result
}

func sandboxSpecFromProto(spec *pb.SandboxSpec) SandboxSpec {
	result := SandboxSpec{
		LogLevel:    spec.GetLogLevel(),
		Environment: spec.GetEnvironment(),
		Providers:   spec.GetProviders(),
	}

	if tmpl := spec.GetTemplate(); tmpl != nil {
		result.Template = &SandboxTemplate{
			Image:            tmpl.GetImage(),
			RuntimeClassName: tmpl.GetRuntimeClassName(),
			AgentSocket:      tmpl.GetAgentSocket(),
			Labels:           tmpl.GetLabels(),
			Annotations:      tmpl.GetAnnotations(),
			Environment:      tmpl.GetEnvironment(),
			UserNamespaces:   tmpl.UserNamespaces,
		}
	}

	if rr := spec.GetResourceRequirements(); rr != nil {
		if gpu := rr.GetGpu(); gpu != nil && gpu.Count != nil {
			result.GPUCount = gpu.Count
		}
	}

	return result
}

func sandboxStatusFromProto(status *pb.SandboxStatus) SandboxStatus {
	result := SandboxStatus{
		SandboxName:          status.GetSandboxName(),
		AgentPod:             status.GetAgentPod(),
		AgentFd:              status.GetAgentFd(),
		SandboxFd:            status.GetSandboxFd(),
		Phase:                sandboxPhaseFromProto(status.GetPhase()),
		CurrentPolicyVersion: status.GetCurrentPolicyVersion(),
	}

	for _, c := range status.GetConditions() {
		result.Conditions = append(result.Conditions, SandboxCondition{
			Type:               c.GetType(),
			Status:             c.GetStatus(),
			Reason:             c.GetReason(),
			Message:            c.GetMessage(),
			LastTransitionTime: c.GetLastTransitionTime(),
		})
	}

	return result
}

func sandboxPhaseFromProto(phase pb.SandboxPhase) SandboxPhase {
	switch phase {
	case pb.SandboxPhase_SANDBOX_PHASE_PROVISIONING:
		return SandboxProvisioning
	case pb.SandboxPhase_SANDBOX_PHASE_READY:
		return SandboxReady
	case pb.SandboxPhase_SANDBOX_PHASE_ERROR:
		return SandboxError
	case pb.SandboxPhase_SANDBOX_PHASE_DELETING:
		return SandboxDeleting
	case pb.SandboxPhase_SANDBOX_PHASE_UNKNOWN:
		return SandboxUnknown
	default:
		return SandboxUnknown
	}
}

func sandboxSpecToProto(spec *SandboxSpec) *pb.SandboxSpec {
	if spec == nil {
		return nil
	}

	result := &pb.SandboxSpec{
		LogLevel:    spec.LogLevel,
		Environment: spec.Environment,
		Providers:   spec.Providers,
	}

	if spec.Template != nil {
		result.Template = &pb.SandboxTemplate{
			Image:            spec.Template.Image,
			RuntimeClassName: spec.Template.RuntimeClassName,
			AgentSocket:      spec.Template.AgentSocket,
			Labels:           spec.Template.Labels,
			Annotations:      spec.Template.Annotations,
			Environment:      spec.Template.Environment,
			UserNamespaces:   spec.Template.UserNamespaces,
		}
	}

	if spec.GPUCount != nil {
		result.ResourceRequirements = &pb.ResourceRequirements{
			Gpu: &pb.GpuResourceRequirements{
				Count: spec.GPUCount,
			},
		}
	}

	return result
}

func (s *sandboxClient) Watch(ctx context.Context, name string, _ ...WatchOptions) (WatchInterface[*Sandbox], error) {
	streamCtx, streamCancel := context.WithCancel(ctx)
	stream, err := s.client.WatchSandbox(streamCtx, &pb.WatchSandboxRequest{
		Id:           name,
		FollowStatus: true,
	})
	if err != nil {
		streamCancel()
		return nil, fromGRPCError(err)
	}

	first, err := stream.Recv()
	if err != nil {
		streamCancel()
		return nil, fromGRPCError(err)
	}

	ch := make(chan Event[*Sandbox], 64)
	w := newWatcher(ch, streamCancel)

	go func() {
		defer close(ch)
		ev := first
		for {
			if sbPayload, ok := ev.Payload.(*pb.SandboxStreamEvent_Sandbox); ok && sbPayload.Sandbox != nil {
				sandbox := sandboxFromProto(sbPayload.Sandbox)
				select {
				case ch <- Event[*Sandbox]{Type: EventModified, Object: sandbox}:
				case <-w.done:
					return
				}
			}
			var err error
			ev, err = stream.Recv()
			if err != nil {
				return
			}
		}
	}()

	return w, nil
}

// sandboxToProto is kept for potential future use (e.g., Update).
func sandboxToProto(s *Sandbox) *pb.Sandbox {
	if s == nil {
		return nil
	}

	return &pb.Sandbox{
		Metadata: &dm.ObjectMeta{
			Id:              s.ID,
			Name:            s.Name,
			CreatedAtMs:     millisFromTime(s.CreatedAt),
			Labels:          s.Labels,
			ResourceVersion: s.ResourceVersion,
		},
		Spec: sandboxSpecToProto(&s.Spec),
	}
}
