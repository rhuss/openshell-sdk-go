// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

// Package main implements the oshell dashboard TUI, an example application
// demonstrating the OpenShell SDK for Go. It provides a terminal-based
// dashboard for managing sandboxes, viewing providers, services, health
// status, and executing commands.
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	tea "charm.land/bubbletea/v2"

	v1 "github.com/rhuss/openshell-sdk-go/openshell/v1"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/gateway"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command-line flags.
	gatewayName := flag.String("gateway", "", "Gateway name to connect to")
	logFile := flag.String("log-file", "", "Path to JSON log file (optional)")
	demo := flag.Bool("demo", false, "Run in demo mode with fake data")
	flag.Parse()

	// Set up the ring buffer and tee handler for logging.
	ring := newRingBuffer(ringBufferSize)

	var fileHandler slog.Handler
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			// Graceful degradation: warn to stderr and continue without file logging.
			fmt.Fprintf(os.Stderr, "Warning: cannot open log file %q: %v\n", *logFile, err)
		} else {
			defer f.Close()
			fileHandler = slog.NewJSONHandler(f, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
		}
	}

	teeH := newTeeHandler(ring, fileHandler)
	logger := slog.New(teeH)

	// Create the log panel.
	logPanel := NewLogPanel(ring)

	// Create the SDK client.
	var client v1.ClientInterface
	var connMgr *ConnectionManager
	var demoCleanup func()

	if *demo {
		// Demo mode: use fake client with pre-populated data.
		client, demoCleanup = setupDemoClient()
		connMgr = NewConnectionManager("demo")
		logger.Info("starting in demo mode")
	} else if *gatewayName != "" {
		// Gateway mode: connect to a real gateway.
		sdkLogger := &slogAdapter{logger: logger}
		gwClient, err := gateway.NewClient(*gatewayName, gateway.WithLogger(sdkLogger))
		if err != nil {
			return fmt.Errorf("failed to create gateway client: %w", err)
		}
		client = gwClient
		connMgr = NewConnectionManager(*gatewayName)
		logger.Info("connecting to gateway", "name", *gatewayName)
	} else {
		// Default to demo mode if no gateway specified.
		client, demoCleanup = setupDemoClient()
		connMgr = NewConnectionManager("demo")
		logger.Info("no gateway specified, starting in demo mode")
	}

	// Build the dashboard model.
	dashboard := NewDashboard(client, connMgr, logger, logPanel)

	// Add tabs in display order: Sandboxes, Providers, Services, Gateway.
	dashboard.AddTab(NewSandboxTab(client, logger))
	dashboard.AddTab(NewProviderTab(client, logger))
	dashboard.AddTab(NewServiceTab(client, logger))
	dashboard.AddTab(NewGatewayTab(client, logger))

	// Run the Bubble Tea program.
	p := tea.NewProgram(dashboard)

	// Wire the program into the tee handler so log entries trigger TUI updates.
	teeH.SetProgram(p)

	_, runErr := p.Run()

	// Graceful shutdown runs unconditionally, even if p.Run() errors.
	if demoCleanup != nil {
		demoCleanup()
	}
	if err := client.Close(); err != nil {
		logger.Error("error closing client", "error", err)
	}

	if runErr != nil {
		return fmt.Errorf("TUI error: %w", runErr)
	}
	return nil
}

// slogAdapter adapts slog.Logger to the SDK's types.Logger interface.
type slogAdapter struct {
	logger *slog.Logger
}

func (a *slogAdapter) Debug(msg string, keysAndValues ...any) {
	a.logger.Debug(msg, keysAndValues...)
}

func (a *slogAdapter) Info(msg string, keysAndValues ...any) {
	a.logger.Info(msg, keysAndValues...)
}

func (a *slogAdapter) Error(err error, msg string, keysAndValues ...any) {
	args := make([]any, 0, len(keysAndValues)+2)
	args = append(args, "error", err)
	args = append(args, keysAndValues...)
	a.logger.Error(msg, args...)
}
