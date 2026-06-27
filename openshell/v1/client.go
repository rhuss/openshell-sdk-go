// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"fmt"
	"sync"
	"time"

	internalgrpc "github.com/rhuss/openshell-sdk-go/openshell/v1/internal/grpc"
	"google.golang.org/grpc"
)

// Config holds all settings needed to create a Client.
type Config struct {
	Address     string
	TLS         *TLSConfig
	Auth        AuthProvider
	Timeout     time.Duration
	RetryPolicy *RetryPolicy
	Logger      Logger
}

// ClientInterface defines the top-level SDK surface.
type ClientInterface interface {
	Sandboxes() SandboxInterface
	Providers() ProviderInterface
	Exec() ExecInterface
	Files() FileInterface
	Health() HealthInterface
	Close() error
}

// SandboxInterface is a placeholder for sandbox operations (implemented in Phase 5).
type SandboxInterface interface{}

// ProviderInterface is defined in provider.go

// ExecInterface is a placeholder for exec operations (implemented in Phase 7).
type ExecInterface interface{}

// FileInterface is a placeholder for file operations (implemented in Phase 8).
type FileInterface interface{}


// Client implements ClientInterface. It holds a gRPC connection and provides
// sub-client accessors following the Kubernetes client-go pattern.
type Client struct {
	conn   *grpc.ClientConn
	config Config

	closeOnce sync.Once
	closeErr  error

	sandboxes SandboxInterface
	providers ProviderInterface
	exec      ExecInterface
	files     FileInterface
	health    HealthInterface
}

// NewClient creates a new SDK client connected to the given gateway.
func NewClient(cfg Config) (*Client, error) {
	if cfg.Address == "" {
		return nil, fmt.Errorf("address must not be empty")
	}

	if cfg.Auth == nil {
		cfg.Auth = NoAuth()
	}

	var tlsParams *internalgrpc.TLSParams
	if cfg.TLS != nil {
		tlsParams = &internalgrpc.TLSParams{
			CertFile: cfg.TLS.CertFile,
			KeyFile:  cfg.TLS.KeyFile,
			CAFile:   cfg.TLS.CAFile,
			Insecure: cfg.TLS.Insecure,
		}
	}

	conn, err := internalgrpc.NewConnection(cfg.Address, tlsParams, cfg.Auth)
	if err != nil {
		return nil, err
	}

	c := &Client{
		conn:   conn,
		config: cfg,
	}

	c.sandboxes = &stubSandbox{}
	c.providers = newProviderClient(conn)
	c.exec = &stubExec{}
	c.files = &stubFile{}
	c.health = newHealthClient(conn)

	return c, nil
}

// Sandboxes returns the sandbox sub-client.
func (c *Client) Sandboxes() SandboxInterface { return c.sandboxes }

// Providers returns the provider sub-client.
func (c *Client) Providers() ProviderInterface { return c.providers }

// Exec returns the exec sub-client.
func (c *Client) Exec() ExecInterface { return c.exec }

// Files returns the file sub-client.
func (c *Client) Files() FileInterface { return c.files }

// Health returns the health sub-client.
func (c *Client) Health() HealthInterface { return c.health }

// Close closes the underlying gRPC connection. Safe to call multiple times.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.closeErr = c.conn.Close()
	})
	return c.closeErr
}

type stubSandbox struct{}
type stubExec struct{}
type stubFile struct{}
