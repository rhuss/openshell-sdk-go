// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/rhuss/openshell-sdk-go/openshell/v1/oidc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConnectionState represents the current state of the gateway connection.
type ConnectionState int

const (
	// StateConnecting is the initial state while establishing a connection.
	StateConnecting ConnectionState = iota
	// StateConnected means the connection is active and healthy.
	StateConnected
	// StateDisconnected means the connection has been lost.
	StateDisconnected
	// StateAuthRequired means authentication is needed (401/Unauthenticated).
	StateAuthRequired
	// StateReconnecting means a reconnection attempt is in progress.
	StateReconnecting
)

// String returns a human-readable label for the connection state.
func (s ConnectionState) String() string {
	switch s {
	case StateConnecting:
		return "Connecting (v2)"
	case StateConnected:
		return "Connected"
	case StateDisconnected:
		return "Disconnected"
	case StateAuthRequired:
		return "Auth Required"
	case StateReconnecting:
		return "Reconnecting"
	default:
		return "Unknown"
	}
}

// Backoff configuration constants for reconnection attempts.
const (
	backoffInitial    = 1 * time.Second
	backoffMultiplier = 2.0
	backoffMax        = 30 * time.Second
	backoffMaxRetries = 10
)

// connectionStateMsg is sent when the connection state changes.
type connectionStateMsg struct {
	state ConnectionState
	err   error
}

// reconnectTickMsg triggers a reconnection attempt after backoff delay.
type reconnectTickMsg struct{}

// ConnectionManager tracks connection state and manages reconnection
// with exponential backoff.
type ConnectionManager struct {
	state        ConnectionState
	retryCount   int
	lastError    error
	gatewayName  string
}

// NewConnectionManager creates a new ConnectionManager starting in
// the Connecting state.
func NewConnectionManager(gatewayName string) *ConnectionManager {
	return &ConnectionManager{
		state:       StateConnecting,
		gatewayName: gatewayName,
	}
}

// State returns the current connection state.
func (cm *ConnectionManager) State() ConnectionState {
	return cm.state
}

// LastError returns the most recent connection error, if any.
func (cm *ConnectionManager) LastError() error {
	return cm.lastError
}

// GatewayName returns the configured gateway name.
func (cm *ConnectionManager) GatewayName() string {
	return cm.gatewayName
}

// Init returns the initial command for the connection manager.
func (cm *ConnectionManager) Init() tea.Cmd {
	return cm.emitState(StateConnecting, nil)
}

// SetConnected transitions to the Connected state and resets retry count.
func (cm *ConnectionManager) SetConnected() tea.Cmd {
	cm.retryCount = 0
	return cm.emitState(StateConnected, nil)
}

// SetDisconnected transitions to the Disconnected state and schedules
// a reconnection attempt if retries remain.
func (cm *ConnectionManager) SetDisconnected(err error) tea.Cmd {
	cm.lastError = err
	if cm.retryCount >= backoffMaxRetries {
		return cm.emitState(StateDisconnected, err)
	}
	return cm.emitState(StateReconnecting, err)
}

// SetAuthRequired transitions to the AuthRequired state.
func (cm *ConnectionManager) SetAuthRequired(err error) tea.Cmd {
	cm.lastError = err
	cm.retryCount = 0
	return cm.emitState(StateAuthRequired, err)
}

// HandleStateMsg processes a connectionStateMsg and returns a follow-up
// command (such as scheduling a reconnection attempt).
func (cm *ConnectionManager) HandleStateMsg(msg connectionStateMsg) tea.Cmd {
	cm.state = msg.state
	cm.lastError = msg.err

	if msg.state == StateReconnecting {
		return cm.scheduleReconnect()
	}

	return nil
}

// HandleReconnectTick processes a reconnect tick and returns a command
// to attempt reconnection.
func (cm *ConnectionManager) HandleReconnectTick() tea.Cmd {
	cm.retryCount++
	return cm.emitState(StateConnecting, nil)
}

// scheduleReconnect returns a command that waits for the backoff duration
// before sending a reconnectTickMsg.
func (cm *ConnectionManager) scheduleReconnect() tea.Cmd {
	delay := cm.backoffDuration()
	return tea.Tick(delay, func(time.Time) tea.Msg {
		return reconnectTickMsg{}
	})
}

// backoffDuration calculates the current backoff delay based on retry count.
// Uses exponential backoff: initial * multiplier^retryCount, capped at max.
func (cm *ConnectionManager) backoffDuration() time.Duration {
	delay := float64(backoffInitial) * math.Pow(backoffMultiplier, float64(cm.retryCount))
	if delay > float64(backoffMax) {
		delay = float64(backoffMax)
	}
	return time.Duration(delay)
}

// ClassifyError inspects an error from an SDK call and transitions the
// connection state accordingly. If the error is a gRPC Unauthenticated
// error (or contains "401" / "unauthenticated"), it transitions to
// AuthRequired. Otherwise it transitions to Disconnected with backoff.
func (cm *ConnectionManager) ClassifyError(err error) tea.Cmd {
	if err == nil {
		return nil
	}
	if IsAuthError(err) {
		return cm.SetAuthRequired(err)
	}
	return cm.SetDisconnected(err)
}

// IsAuthError returns true if the error represents an authentication
// failure: gRPC Unauthenticated status code, or error text containing
// "401" or "unauthenticated" (case-insensitive).
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	// Check gRPC status code.
	if s, ok := status.FromError(err); ok {
		if s.Code() == codes.Unauthenticated {
			return true
		}
	}
	// Fallback: check error message for common auth failure indicators.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unauthenticated") || strings.Contains(msg, "401")
}

// sdkErrorMsg is sent when an SDK operation encounters an error that
// may indicate a connection or authentication failure. The Dashboard
// routes these through the ConnectionManager to detect 401s.
type sdkErrorMsg struct {
	err    error
	source string // identifies which operation failed (e.g., "sandbox.list")
}

// HandleSDKError processes an error from an SDK operation and returns
// the appropriate state transition command. If the error is an auth
// failure during an active session, it transitions to AuthRequired.
func (cm *ConnectionManager) HandleSDKError(msg sdkErrorMsg) tea.Cmd {
	if msg.err == nil {
		return nil
	}
	return cm.ClassifyError(msg.err)
}

// NewSDKErrorMsg creates a tea.Cmd that sends an sdkErrorMsg. Tab
// implementations use this to report SDK call failures back to the
// Dashboard for centralized error handling.
func NewSDKErrorMsg(err error, source string) tea.Cmd {
	return func() tea.Msg {
		return sdkErrorMsg{err: err, source: source}
	}
}

// OIDCLogin performs a browser-based OIDC login flow for the given gateway.
// The context is used to cancel the login if the dashboard shuts down
// mid-flow (avoids orphaned goroutines and lingering callback listeners).
func OIDCLogin(ctx context.Context, gatewayName string) tea.Cmd {
	return func() tea.Msg {
		loginCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
		defer cancel()
		_, err := oidc.Login(loginCtx, gatewayName)
		if err != nil {
			return authLoginMsg{err: err}
		}
		return authLoginMsg{err: nil}
	}
}

// emitState returns a command that sends a connectionStateMsg.
func (cm *ConnectionManager) emitState(state ConnectionState, err error) tea.Cmd {
	return func() tea.Msg {
		return connectionStateMsg{state: state, err: err}
	}
}
