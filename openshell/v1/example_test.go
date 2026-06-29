// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package v1_test

import (
	"context"
	"fmt"
	"log"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/fake"
)

// ExampleClient_Sandboxes demonstrates the sandbox lifecycle: create a sandbox,
// wait for it to become ready, and then clean up.
func ExampleClient_Sandboxes() {
	client := fake.NewClient()
	defer client.Close() //nolint:errcheck

	ctx := context.Background()

	// Create a sandbox
	sb, err := client.Sandboxes().Create(ctx, "my-sandbox", &v1.SandboxSpec{}, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Phase after create:", sb.Status.Phase)

	// Wait for the sandbox to become ready
	sb, err = client.Sandboxes().WaitReady(ctx, "my-sandbox")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Phase after wait:", sb.Status.Phase)

	// Clean up
	if err := client.Sandboxes().Delete(ctx, "my-sandbox"); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deleted")
	// Output:
	// Phase after create: Provisioning
	// Phase after wait: Ready
	// Deleted
}

// ExampleClient_Providers demonstrates registering and listing providers.
func ExampleClient_Providers() {
	client := fake.NewClient()
	defer client.Close() //nolint:errcheck

	ctx := context.Background()

	// Register a provider
	_, err := client.Providers().Create(ctx, &v1.Provider{
		Name: "my-openai",
		Type: "openai",
	})
	if err != nil {
		log.Fatal(err)
	}

	// List all providers
	providers, err := client.Providers().List(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Count:", len(providers))
	fmt.Println("Name:", providers[0].Name)
	// Output:
	// Count: 1
	// Name: my-openai
}

// ExampleClient_Health demonstrates checking gateway health.
func ExampleClient_Health() {
	client := fake.NewClient()
	defer client.Close() //nolint:errcheck

	ctx := context.Background()

	result, err := client.Health().Check(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Healthy:", result.Healthy)
	// Output:
	// Healthy: true
}

// ExampleClient_Exec demonstrates running a command in a sandbox.
// The fake client returns Unimplemented for exec operations, so this
// example shows the call pattern and error handling.
func ExampleClient_Exec() {
	client := fake.NewClient()
	defer client.Close() //nolint:errcheck

	ctx := context.Background()

	_, err := client.Exec().Run(ctx, "my-sandbox", []string{"echo", "hello"})
	if v1.IsUnimplemented(err) {
		fmt.Println("Exec requires a real gateway")
	}
	// Output:
	// Exec requires a real gateway
}

// ExampleIsNotFound demonstrates handling a not-found error.
func ExampleIsNotFound() {
	client := fake.NewClient()
	defer client.Close() //nolint:errcheck

	ctx := context.Background()

	_, err := client.Sandboxes().Get(ctx, "nonexistent")
	if v1.IsNotFound(err) {
		fmt.Println("Sandbox not found")
	}
	// Output:
	// Sandbox not found
}

// ExampleIsAlreadyExists demonstrates handling a duplicate-creation error.
func ExampleIsAlreadyExists() {
	client := fake.NewClient()
	defer client.Close() //nolint:errcheck

	ctx := context.Background()

	// Create a sandbox
	_, err := client.Sandboxes().Create(ctx, "my-sandbox", &v1.SandboxSpec{}, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Try to create the same sandbox again
	_, err = client.Sandboxes().Create(ctx, "my-sandbox", &v1.SandboxSpec{}, nil)
	if v1.IsAlreadyExists(err) {
		fmt.Println("Sandbox already exists")
	}
	// Output:
	// Sandbox already exists
}

// ExampleIsUnavailable demonstrates detecting a closed client.
func ExampleIsUnavailable() {
	client := fake.NewClient()
	_ = client.Close()

	ctx := context.Background()

	_, err := client.Sandboxes().Get(ctx, "any")
	if v1.IsUnavailable(err) {
		fmt.Println("Client is closed")
	}
	// Output:
	// Client is closed
}
