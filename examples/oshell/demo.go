// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/fake"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/types"
)

// demoTransitionInterval is the time between phase transitions in demo mode.
const demoTransitionInterval = 5 * time.Second

// demoExecDelay is the simulated execution delay for canned exec responses.
const demoExecDelay = 500 * time.Millisecond

// setupDemoClient creates a fake client pre-populated with demo data:
// 5 sandboxes in mixed phases, provider profiles, service endpoints,
// and a healthy health result. It starts a background phase transition
// simulator and returns a cleanup function that stops the simulator.
// The caller must invoke the cleanup function before closing the client.
func setupDemoClient() (v1.ClientInterface, func()) {
	fc := fake.NewClient(
		fake.WithHealthResult(&types.HealthResult{
			Healthy: true,
			Version: "demo-v1.0.0",
		}),
	)

	// Pre-populate 5 sandboxes with mixed phases.
	sandboxes := []*types.Sandbox{
		{
			ID:   "sb-001",
			Name: "dev-workspace",
			CreatedAt: time.Now().Add(-2 * time.Hour),
			Labels:    map[string]string{"env": "dev", "team": "platform"},
			Spec: types.SandboxSpec{
				Template: &types.SandboxTemplate{
					Image: "nvidia/openshell:latest",
				},
				Providers: []string{"openai", "github"},
			},
			Status: types.SandboxStatus{
				SandboxName: "dev-workspace",
				Phase:       types.SandboxReady,
			},
		},
		{
			ID:   "sb-002",
			Name: "ml-training",
			CreatedAt: time.Now().Add(-1 * time.Hour),
			Labels:    map[string]string{"env": "staging", "team": "ml"},
			Spec: types.SandboxSpec{
				Template: &types.SandboxTemplate{
					Image: "nvidia/cuda:12.4-runtime",
				},
				Providers: []string{"anthropic"},
			},
			Status: types.SandboxStatus{
				SandboxName: "ml-training",
				Phase:       types.SandboxReady,
			},
		},
		{
			ID:   "sb-003",
			Name: "data-pipeline",
			CreatedAt: time.Now().Add(-30 * time.Minute),
			Labels:    map[string]string{"env": "prod", "team": "data"},
			Spec: types.SandboxSpec{
				Template: &types.SandboxTemplate{
					Image: "nvidia/openshell:nightly",
				},
				Providers: []string{"openai"},
			},
			Status: types.SandboxStatus{
				SandboxName: "data-pipeline",
				Phase:       types.SandboxProvisioning,
			},
		},
		{
			ID:   "sb-004",
			Name: "test-runner",
			CreatedAt: time.Now().Add(-15 * time.Minute),
			Labels:    map[string]string{"env": "test", "team": "qa"},
			Spec: types.SandboxSpec{
				Template: &types.SandboxTemplate{
					Image: "nvidia/openshell:latest",
				},
			},
			Status: types.SandboxStatus{
				SandboxName: "test-runner",
				Phase:       types.SandboxUnknown,
			},
		},
		{
			ID:   "sb-005",
			Name: "broken-sandbox",
			CreatedAt: time.Now().Add(-45 * time.Minute),
			Labels:    map[string]string{"env": "dev", "team": "platform"},
			Spec: types.SandboxSpec{
				Template: &types.SandboxTemplate{
					Image: "nvidia/openshell:experimental",
				},
			},
			Status: types.SandboxStatus{
				SandboxName: "broken-sandbox",
				Phase:       types.SandboxError,
			},
		},
	}
	for _, sb := range sandboxes {
		fc.AddSandbox(sb)
	}

	// Pre-populate providers.
	providers := []*types.Provider{
		{
			ID:        "prov-001",
			Name:      "openai",
			Type:      "openai",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			Labels:    map[string]string{"tier": "enterprise"},
		},
		{
			ID:        "prov-002",
			Name:      "anthropic",
			Type:      "anthropic",
			CreatedAt: time.Now().Add(-12 * time.Hour),
			Labels:    map[string]string{"tier": "standard"},
		},
		{
			ID:        "prov-003",
			Name:      "github",
			Type:      "github",
			CreatedAt: time.Now().Add(-6 * time.Hour),
			Labels:    map[string]string{"tier": "standard"},
		},
	}
	for _, p := range providers {
		fc.AddProvider(p)
	}

	// Start the phase transition simulator so sandboxes in Unknown or
	// Provisioning phase gradually transition to Ready.
	sim := newPhaseTransitionSimulator(fc)
	sim.Start()

	// Wrap the fake client with demo overrides for services, exec,
	// config, and provider profiles (all return Unimplemented in the
	// base fake client).
	client := &demoClient{
		ClientInterface: fc,
		sandboxes:       newDemoSandboxClient(fc.Sandboxes()),
		services:        newDemoServiceClient(),
		exec:            newDemoExecClient(),
		config:          newDemoConfigClient(),
		profiles:        newDemoProfileClient(),
		fakeClient:      fc,
	}

	cleanup := func() {
		sim.Stop()
	}

	return client, cleanup
}

// demoClient wraps a fake.Client and overrides sub-clients that
// return Unimplemented with demo-specific implementations.
type demoClient struct {
	v1.ClientInterface
	sandboxes  v1.SandboxInterface
	services   v1.ServiceInterface
	exec       v1.ExecInterface
	config     v1.ConfigInterface
	profiles   v1.ProfileInterface
	fakeClient *fake.Client
}

func (dc *demoClient) Sandboxes() v1.SandboxInterface {
	return dc.sandboxes
}

// Services returns the demo service client with pre-populated endpoints.
func (dc *demoClient) Services() v1.ServiceInterface {
	return dc.services
}

// Exec returns the demo exec client with canned responses.
func (dc *demoClient) Exec() v1.ExecInterface {
	return dc.exec
}

// Config returns the demo config client with pre-populated settings.
func (dc *demoClient) Config() v1.ConfigInterface {
	return dc.config
}

// Providers returns a wrapped provider interface that overrides Profiles().
func (dc *demoClient) Providers() v1.ProviderInterface {
	return &demoProviderClient{
		ProviderInterface: dc.ClientInterface.Providers(),
		profiles:          dc.profiles,
	}
}

// demoProviderClient wraps the fake provider client to override Profiles().
type demoProviderClient struct {
	v1.ProviderInterface
	profiles v1.ProfileInterface
}

// Profiles returns the demo profile client with pre-populated profiles.
func (dp *demoProviderClient) Profiles() v1.ProfileInterface {
	return dp.profiles
}

// ---------- Demo Service Client ----------

// demoServiceClient provides pre-populated service endpoints.
type demoServiceClient struct {
	endpoints []*types.ServiceEndpoint
}

func newDemoServiceClient() *demoServiceClient {
	return &demoServiceClient{
		endpoints: []*types.ServiceEndpoint{
			{
				ID:          "svc-001",
				SandboxID:   "sb-001",
				SandboxName: "dev-workspace",
				ServiceName: "jupyter",
				TargetPort:  8888,
				URL:         "https://dev-workspace-jupyter.demo.openshell.ai",
			},
			{
				ID:          "svc-002",
				SandboxID:   "sb-001",
				SandboxName: "dev-workspace",
				ServiceName: "api",
				TargetPort:  8080,
				URL:         "https://dev-workspace-api.demo.openshell.ai",
			},
			{
				ID:          "svc-003",
				SandboxID:   "sb-002",
				SandboxName: "ml-training",
				ServiceName: "tensorboard",
				TargetPort:  6006,
				URL:         "https://ml-training-tensorboard.demo.openshell.ai",
			},
			{
				ID:          "svc-004",
				SandboxID:   "sb-003",
				SandboxName: "data-pipeline",
				ServiceName: "dashboard",
				TargetPort:  3000,
				URL:         "https://data-pipeline-dashboard.demo.openshell.ai",
			},
		},
	}
}

func (ds *demoServiceClient) Expose(_ context.Context, _, _ string, _ uint32, _ bool) (*types.ServiceEndpoint, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Expose is not supported in demo mode"}
}

func (ds *demoServiceClient) Get(_ context.Context, sandboxName, serviceName string) (*types.ServiceEndpoint, error) {
	for _, ep := range ds.endpoints {
		if ep.SandboxName == sandboxName && ep.ServiceName == serviceName {
			return ep, nil
		}
	}
	return nil, &types.StatusError{Code: types.ErrorNotFound, Message: fmt.Sprintf("service %s/%s not found", sandboxName, serviceName)}
}

func (ds *demoServiceClient) List(_ context.Context, sandboxName string, _ ...v1.ListOptions) ([]*types.ServiceEndpoint, error) {
	if sandboxName == "" {
		// Return all endpoints.
		result := make([]*types.ServiceEndpoint, len(ds.endpoints))
		copy(result, ds.endpoints)
		return result, nil
	}
	var result []*types.ServiceEndpoint
	for _, ep := range ds.endpoints {
		if ep.SandboxName == sandboxName {
			result = append(result, ep)
		}
	}
	return result, nil
}

func (ds *demoServiceClient) Delete(_ context.Context, _, _ string) error {
	return &types.StatusError{Code: types.ErrorUnimplemented, Message: "Delete is not supported in demo mode"}
}

// ---------- Demo Exec Client ----------

// demoExecClient returns canned exec responses with a simulated delay.
type demoExecClient struct{}

func newDemoExecClient() *demoExecClient {
	return &demoExecClient{}
}

func (de *demoExecClient) Run(_ context.Context, sandboxName string, command []string, _ ...v1.ExecOptions) (*types.ExecResult, error) {
	// Simulate execution delay.
	time.Sleep(demoExecDelay)

	stdout := fmt.Sprintf("Hello from sandbox %s!\n", sandboxName)
	if len(command) > 0 {
		stdout = fmt.Sprintf("$ %s\nHello from sandbox %s!\n", joinCommand(command), sandboxName)
	}

	return &types.ExecResult{
		ExitCode: 0,
		Stdout:   []byte(stdout),
		Stderr:   []byte{},
	}, nil
}

func (de *demoExecClient) Stream(_ context.Context, _ string, _ []string, _ ...v1.ExecOptions) (v1.ExecStream, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Stream is not supported in demo mode"}
}

func (de *demoExecClient) Interactive(_ context.Context, _ string, _ []string, _, _ uint32, _ ...v1.ExecOptions) (v1.InteractiveSession, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Interactive is not supported in demo mode"}
}

// joinCommand joins command parts into a display string.
func joinCommand(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	result := parts[0]
	for _, p := range parts[1:] {
		result += " " + p
	}
	return result
}

// ---------- Demo Config Client ----------

// demoConfigClient returns pre-populated gateway configuration.
type demoConfigClient struct{}

func newDemoConfigClient() *demoConfigClient {
	return &demoConfigClient{}
}

func (dc *demoConfigClient) GetSandbox(_ context.Context, _ string) (*v1.SandboxConfig, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "GetSandbox is not supported in demo mode"}
}

func (dc *demoConfigClient) GetGateway(_ context.Context) (*v1.GatewayConfig, error) {
	return &v1.GatewayConfig{
		Settings: map[string]types.SettingValue{
			"log_level": {
				Type:      types.SettingValueString,
				StringVal: "info",
			},
			"max_sandboxes": {
				Type:   types.SettingValueInt,
				IntVal: 100,
			},
			"auto_cleanup": {
				Type:    types.SettingValueBool,
				BoolVal: true,
			},
			"default_image": {
				Type:      types.SettingValueString,
				StringVal: "nvidia/openshell:latest",
			},
			"gpu_sharing": {
				Type:    types.SettingValueBool,
				BoolVal: false,
			},
		},
		SettingsRevision: 42,
	}, nil
}

func (dc *demoConfigClient) Update(_ context.Context, _ *v1.ConfigUpdate) (*v1.ConfigUpdateResult, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Update is not supported in demo mode"}
}

// ---------- Demo Profile Client ----------

// demoProfileClient returns pre-populated provider profiles.
type demoProfileClient struct {
	profiles []*types.ProviderProfile
}

func newDemoProfileClient() *demoProfileClient {
	return &demoProfileClient{
		profiles: []*types.ProviderProfile{
			{
				ID:          "profile-openai",
				DisplayName: "OpenAI",
				Description: "OpenAI GPT models (GPT-4, GPT-4o, o1)",
				Category:    types.ProfileCategoryInference,
				Credentials: []types.ProfileCredential{
					{Name: "api_key", Description: "OpenAI API key", Required: true, Secret: true},
				},
				InferenceCapable: true,
			},
			{
				ID:          "profile-anthropic",
				DisplayName: "Anthropic",
				Description: "Anthropic Claude models (Opus, Sonnet, Haiku)",
				Category:    types.ProfileCategoryInference,
				Credentials: []types.ProfileCredential{
					{Name: "api_key", Description: "Anthropic API key", Required: true, Secret: true},
				},
				InferenceCapable: true,
			},
			{
				ID:          "profile-github",
				DisplayName: "GitHub",
				Description: "GitHub source control integration",
				Category:    types.ProfileCategorySourceControl,
				Credentials: []types.ProfileCredential{
					{Name: "token", Description: "GitHub personal access token", Required: true, Secret: true},
				},
				InferenceCapable: false,
			},
			{
				ID:          "profile-slack",
				DisplayName: "Slack",
				Description: "Slack messaging integration",
				Category:    types.ProfileCategoryMessaging,
				Credentials: []types.ProfileCredential{
					{Name: "bot_token", Description: "Slack bot token", Required: true, Secret: true},
				},
				InferenceCapable: false,
			},
			{
				ID:          "profile-postgres",
				DisplayName: "PostgreSQL",
				Description: "PostgreSQL database connector",
				Category:    types.ProfileCategoryData,
				Credentials: []types.ProfileCredential{
					{Name: "connection_string", Description: "PostgreSQL connection URI", Required: true, Secret: true},
				},
				InferenceCapable: false,
			},
		},
	}
}

func (dp *demoProfileClient) List(_ context.Context, _ ...v1.ListOptions) ([]*types.ProviderProfile, error) {
	result := make([]*types.ProviderProfile, len(dp.profiles))
	copy(result, dp.profiles)
	return result, nil
}

func (dp *demoProfileClient) Get(_ context.Context, id string) (*types.ProviderProfile, error) {
	for _, p := range dp.profiles {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, &types.StatusError{Code: types.ErrorNotFound, Message: fmt.Sprintf("profile %q not found", id)}
}

func (dp *demoProfileClient) Import(_ context.Context, _ []types.ProfileImportItem) (*types.ImportResult, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Import is not supported in demo mode"}
}

func (dp *demoProfileClient) Update(_ context.Context, _ string, _ uint64, _ types.ProfileImportItem) (*types.UpdateResult, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Update is not supported in demo mode"}
}

func (dp *demoProfileClient) Lint(_ context.Context, _ []types.ProfileImportItem) (*types.LintResult, error) {
	return nil, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Lint is not supported in demo mode"}
}

func (dp *demoProfileClient) Delete(_ context.Context, _ string) (bool, error) {
	return false, &types.StatusError{Code: types.ErrorUnimplemented, Message: "Delete is not supported in demo mode"}
}

// ---------- Phase Transition Simulator ----------

// phaseTransitionSimulator runs a background goroutine that transitions
// fake sandboxes through Unknown -> Provisioning -> Ready, one transition
// every demoTransitionInterval.
type phaseTransitionSimulator struct {
	client *fake.Client
	mu     sync.Mutex
	stopCh chan struct{}
}

// newPhaseTransitionSimulator creates a simulator for the given fake client.
func newPhaseTransitionSimulator(client *fake.Client) *phaseTransitionSimulator {
	return &phaseTransitionSimulator{
		client: client,
		stopCh: make(chan struct{}),
	}
}

// Start begins the phase transition loop in a background goroutine.
func (pts *phaseTransitionSimulator) Start() {
	go pts.run()
}

// Stop signals the simulator to stop.
func (pts *phaseTransitionSimulator) Stop() {
	pts.mu.Lock()
	defer pts.mu.Unlock()
	select {
	case <-pts.stopCh:
		// Already stopped.
	default:
		close(pts.stopCh)
	}
}

// run is the main loop that checks for sandboxes to transition.
func (pts *phaseTransitionSimulator) run() {
	ticker := time.NewTicker(demoTransitionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-pts.stopCh:
			return
		case <-ticker.C:
			pts.transitionNext()
		}
	}
}

// transitionNext finds the first sandbox that can be transitioned and
// advances it one phase step: Unknown -> Provisioning -> Ready.
func (pts *phaseTransitionSimulator) transitionNext() {
	ctx := context.Background()

	sandboxes, err := pts.client.Sandboxes().List(ctx)
	if err != nil {
		return
	}

	for _, sb := range sandboxes {
		var nextPhase types.SandboxPhase
		switch sb.Status.Phase {
		case types.SandboxUnknown:
			nextPhase = types.SandboxProvisioning
		case types.SandboxProvisioning:
			nextPhase = types.SandboxReady
		default:
			continue
		}

		// Update via AddSandbox (store-level upsert). This updates the
		// sandbox phase in the store so List() reflects the change, but
		// does not broadcast to Watch streams because AddSandbox is a
		// direct store insert. The sandbox tab's periodic Refresh (on
		// tab switch) picks up the new phase. A broadcast-aware Update
		// method on the fake client would fix this for live Watch, but
		// that's an SDK enhancement outside this example's scope.
		sb.Status.Phase = nextPhase
		pts.client.AddSandbox(sb)
		return // One transition per tick.
	}
}

// ---------- Demo Sandbox Client (GetLogs override) ----------

type demoSandboxClient struct {
	v1.SandboxInterface
}

func newDemoSandboxClient(base v1.SandboxInterface) *demoSandboxClient {
	return &demoSandboxClient{SandboxInterface: base}
}

func (ds *demoSandboxClient) GetLogs(_ context.Context, sandboxName string, _ ...v1.LogOption) (*types.LogResult, error) {
	now := time.Now()
	lines := []types.LogLine{
		{Timestamp: now.Add(-5 * time.Minute), Level: "INFO", Target: "sandbox", Message: "Container started", Source: "sandbox", Fields: map[string]string{"sandbox": sandboxName}},
		{Timestamp: now.Add(-4 * time.Minute), Level: "INFO", Target: "runtime", Message: "Python 3.12 runtime initialized", Source: "sandbox"},
		{Timestamp: now.Add(-3 * time.Minute), Level: "DEBUG", Target: "network", Message: "Connected to gateway mesh", Source: "gateway"},
		{Timestamp: now.Add(-2 * time.Minute), Level: "INFO", Target: "provider", Message: "Provider openai attached successfully", Source: "gateway", Fields: map[string]string{"provider": "openai"}},
		{Timestamp: now.Add(-90 * time.Second), Level: "WARN", Target: "resources", Message: "GPU memory usage at 75%", Source: "sandbox"},
		{Timestamp: now.Add(-1 * time.Minute), Level: "INFO", Target: "sandbox", Message: "Ready to accept connections", Source: "sandbox"},
		{Timestamp: now.Add(-30 * time.Second), Level: "DEBUG", Target: "health", Message: "Health check passed", Source: "gateway"},
		{Timestamp: now.Add(-10 * time.Second), Level: "INFO", Target: "sandbox", Message: "Heartbeat OK", Source: "sandbox"},
	}
	return &types.LogResult{Lines: lines, BufferTotal: uint32(len(lines))}, nil
}
