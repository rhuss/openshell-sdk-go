// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package v1 provides a Go SDK for interacting with OpenShell servers.
//
// The SDK follows the Kubernetes client-go sub-client pattern: a single Client
// provides typed accessors for each resource domain (Sandboxes, Providers, Exec,
// Files, Health). All operations accept a context.Context and return idiomatic
// Go types. Proto-generated types never appear in the public API.
//
// # Quick Start
//
//	client, err := v1.NewClient(v1.Config{
//	    Address: "gateway.example.com:443",
//	    Auth:    v1.StaticToken("my-token"),
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
// # Sandbox Lifecycle
//
//	sandbox, err := client.Sandboxes().Create(ctx, v1.SandboxSpec{
//	    Image:       "python:3.12",
//	    Environment: map[string]string{"LANG": "en_US.UTF-8"},
//	}, v1.CreateOptions{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	if err := client.Sandboxes().WaitReady(ctx, sandbox.Name, v1.WaitOptions{}); err != nil {
//	    log.Fatal(err)
//	}
//
// # Command Execution
//
//	result, err := client.Exec().Run(ctx, sandbox.Name, []string{"echo", "hello"}, v1.ExecOptions{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(string(result.Stdout)) // "hello\n"
//
// # Error Handling
//
//	_, err := client.Sandboxes().Get(ctx, "missing", v1.GetOptions{})
//	if v1.IsNotFound(err) {
//	    // handle not found
//	}
//
// # Watching
//
//	watcher, err := client.Sandboxes().Watch(ctx, v1.WatchOptions{})
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer watcher.Stop()
//	for event := range watcher.ResultChan() {
//	    fmt.Printf("%s: %s\n", event.Type, event.Object.Name)
//	}
package v1
