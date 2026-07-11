// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// StatusBar renders the top-of-screen status bar showing gateway name,
// connection state, auth status, and a health indicator dot.
type StatusBar struct {
	gatewayName string
	connState   ConnectionState
	authStatus  string
	healthDot   string
	healthColor color.Color
	tokenExpiry time.Time // zero value means no expiry known
}

// NewStatusBar creates a new StatusBar with default values.
func NewStatusBar() *StatusBar {
	return &StatusBar{
		gatewayName: "",
		connState:   StateDisconnected,
		authStatus:  "",
		healthDot:   healthDotUnhealthy,
		healthColor: colorGray,
	}
}

// SetGatewayName updates the displayed gateway name.
func (s *StatusBar) SetGatewayName(name string) {
	s.gatewayName = name
}

// SetConnectionState updates the displayed connection state.
func (s *StatusBar) SetConnectionState(state ConnectionState) {
	s.connState = state
}

// SetAuthStatus updates the auth status display string.
func (s *StatusBar) SetAuthStatus(status string) {
	s.authStatus = status
}

// SetTokenExpiry updates the token expiry time. Pass zero time to clear.
func (s *StatusBar) SetTokenExpiry(expiry time.Time) {
	s.tokenExpiry = expiry
}

// formatTokenExpiry returns a human-readable countdown string for the
// token expiry. Returns empty string if no expiry is set or the token
// has already expired.
func (s *StatusBar) formatTokenExpiry() string {
	if s.tokenExpiry.IsZero() {
		return ""
	}
	remaining := time.Until(s.tokenExpiry)
	if remaining <= 0 {
		return "expired"
	}
	minutes := int(remaining.Minutes())
	seconds := int(remaining.Seconds()) % 60
	if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// SetHealthy sets the health dot to green (healthy).
func (s *StatusBar) SetHealthy() {
	s.healthDot = healthDotHealthy
	s.healthColor = colorGreen
}

// SetDegraded sets the health dot to yellow (degraded).
func (s *StatusBar) SetDegraded() {
	s.healthDot = healthDotDegraded
	s.healthColor = colorYellow
}

// SetUnhealthy sets the health dot to red (unhealthy).
func (s *StatusBar) SetUnhealthy() {
	s.healthDot = healthDotUnhealthy
	s.healthColor = colorRed
}

// View renders the status bar for the given terminal width.
func (s *StatusBar) View(width int) string {
	dotStyle := lipgloss.NewStyle().Foreground(s.healthColor)

	var parts []string

	// Health dot.
	parts = append(parts, dotStyle.Render(s.healthDot))

	// Gateway name.
	if s.gatewayName != "" {
		parts = append(parts, fmt.Sprintf("%s %s",
			statusKeyStyle.Render("GW:"),
			statusValueStyle.Render(s.gatewayName),
		))
	}

	// Connection state.
	stateStr := s.connState.String()
	stateStyle := statusValueStyle
	switch s.connState {
	case StateConnected:
		stateStyle = lipgloss.NewStyle().Foreground(colorGreen)
	case StateConnecting, StateReconnecting:
		stateStyle = lipgloss.NewStyle().Foreground(colorYellow)
	case StateDisconnected, StateAuthRequired:
		stateStyle = lipgloss.NewStyle().Foreground(colorRed)
	}
	parts = append(parts, fmt.Sprintf("%s %s",
		statusKeyStyle.Render("Status:"),
		stateStyle.Render(stateStr),
	))

	// Auth status.
	if s.authStatus != "" {
		parts = append(parts, fmt.Sprintf("%s %s",
			statusKeyStyle.Render("Auth:"),
			statusValueStyle.Render(s.authStatus),
		))
	}

	// Token expiry countdown.
	if tokenStr := s.formatTokenExpiry(); tokenStr != "" {
		tokenStyle := statusValueStyle
		if tokenStr == "expired" {
			tokenStyle = lipgloss.NewStyle().Foreground(colorRed)
		} else {
			// Warn with yellow if less than 2 minutes remain.
			remaining := time.Until(s.tokenExpiry)
			if remaining < 2*time.Minute {
				tokenStyle = lipgloss.NewStyle().Foreground(colorYellow)
			}
		}
		parts = append(parts, fmt.Sprintf("%s %s",
			statusKeyStyle.Render("Token:"),
			tokenStyle.Render(tokenStr),
		))
	}

	content := strings.Join(parts, "  ")
	return statusBarStyle.Width(width).Render(content)
}
