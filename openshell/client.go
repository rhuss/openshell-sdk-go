// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package openshell provides a Go SDK for interacting with OpenShell servers.
package openshell

import "errors"

// Client represents a connection to an OpenShell server.
type Client struct{}

// Dial creates a new Client connected to the given address.
func Dial(address string) (*Client, error) {
	if address == "" {
		return nil, errors.New("address must not be empty")
	}
	return &Client{}, nil
}

// Close closes the client connection.
func (c *Client) Close() error { return nil }
