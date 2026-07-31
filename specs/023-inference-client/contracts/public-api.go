// Package contract defines the public API surface for the inference route client.
// This file is a design artifact, not compiled code.
package contract

import "context"

// InferenceInterface defines inference route management operations.
// Accessed via client.Inference().
type InferenceInterface interface {
	// SetRoute configures an inference route for a workspace.
	// Returns ErrorInvalidArgument if workspace, providerName, or modelID is empty.
	SetRoute(ctx context.Context, workspace string, config *InferenceRouteConfig) (*InferenceRoute, error)

	// GetRoute retrieves the inference route for a workspace by route name.
	// Returns ErrorInvalidArgument if workspace is empty.
	// Returns ErrorNotFound if no route exists for the given name.
	GetRoute(ctx context.Context, workspace, routeName string) (*InferenceRoute, error)

	// DeleteRoute removes an inference route from a workspace.
	// Returns ErrorInvalidArgument if workspace is empty.
	// Idempotent: deleting a non-existent route is not an error.
	DeleteRoute(ctx context.Context, workspace, routeName string) error
}

// InferenceRouteConfig holds parameters for setting an inference route.
type InferenceRouteConfig struct {
	ProviderName string
	ModelID      string
	RouteName    string
	NoVerify     bool
	TimeoutSecs  uint64
}

// InferenceRoute represents a configured inference route.
type InferenceRoute struct {
	ProviderName        string
	ModelID             string
	Version             uint64
	RouteName           string
	TimeoutSecs         uint64
	Workspace           string
	ValidationPerformed bool
	ValidatedEndpoints  []ValidatedEndpoint
}

// ValidatedEndpoint represents an endpoint probed during route validation.
type ValidatedEndpoint struct {
	URL      string
	Protocol string
}
